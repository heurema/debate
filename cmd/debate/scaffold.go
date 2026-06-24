package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/heurema/debate/internal/debate/config"
)

const (
	scaffoldModel  = "claude-haiku-4-5"
	scaffoldEffort = "medium"
)

const proposerTemplate = `---
role: debater
model: claude-haiku-4-5
effort: medium
---
You are the Proposer. Defend the proposition with clear arguments and respond to the Skeptic's objections. Be concise and direct.
`

const skepticTemplate = `---
role: debater
model: claude-haiku-4-5
effort: medium
---
You are the Skeptic. Challenge the proposition by identifying weaknesses, counter-examples, and unresolved assumptions. Be constructive and specific.
`

type initCmd struct {
	Args []string `arg:"" optional:"" name:"args" hidden:""`
}

type newCmd struct {
	Role string   `name:"role" default:"debater" help:"Persona role (debater|synthesizer)."`
	Name []string `arg:"" optional:"" name:"name" help:"Persona name."`
}

func (c *initCmd) Run(deps *cliDeps) error {
	workDir, err := deps.resolveWorkDir()
	if err != nil {
		fmt.Fprintln(deps.stderr, "error: could not get working directory:", err)
		deps.code = 1
		return nil
	}
	deps.code = runInit(c, deps.stdout, deps.stderr, workDir)
	return nil
}

func (c *newCmd) Run(deps *cliDeps) error {
	workDir, err := deps.resolveWorkDir()
	if err != nil {
		fmt.Fprintln(deps.stderr, "error: could not get working directory:", err)
		deps.code = 1
		return nil
	}
	deps.code = runNew(c, deps.stdout, deps.stderr, workDir)
	return nil
}

// cmdInit implements the "debate init" subcommand.
// It scaffolds a .heurema/debate workspace under workDir, skipping files that already exist.
func cmdInit(args []string, stdout, stderr io.Writer, workDir string) int {
	var cmd initCmd
	if code, ok := parseStandaloneCommand(&cmd, "debate init", args, stdout, stderr, workDir); !ok {
		return code
	}
	return runInit(&cmd, stdout, stderr, workDir)
}

func runInit(cmd *initCmd, stdout, stderr io.Writer, workDir string) int {
	if len(cmd.Args) > 0 {
		fmt.Fprintln(stderr, "error: debate init takes no arguments")
		printInitUsage(stderr)
		return 1
	}

	debDir := filepath.Join(workDir, ".heurema", "debate")
	personasDir := filepath.Join(debDir, "personas")
	if err := os.MkdirAll(personasDir, 0755); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	files := []struct {
		path    string
		content string
	}{
		{filepath.Join(personasDir, "proposer.md"), proposerTemplate},
		{filepath.Join(personasDir, "skeptic.md"), skepticTemplate},
	}

	for _, f := range files {
		created, err := writeIfAbsent(f.path, f.content)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		if created {
			fmt.Fprintln(stdout, "created", f.path)
		} else {
			fmt.Fprintln(stdout, "skipped", f.path, "(already exists)")
		}
	}
	return 0
}

func printInitUsage(stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: debate init")
	fmt.Fprintln(stderr, "  Scaffold a .heurema/debate workspace in the current directory.")
}

func parseStandaloneCommand(cmd any, name string, args []string, stdout, stderr io.Writer, workDir string) (int, bool) {
	parser, err := kong.New(cmd,
		kong.Name(name),
		kong.Writers(stderr, stderr),
		kong.ShortUsageOnError(),
		kong.Exit(func(code int) { panic(kongExit(code)) }),
	)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1, false
	}
	var parseErr error
	if code, ok := catchKongExit(func() {
		_, parseErr = parser.Parse(args)
	}); ok {
		return code, false
	}
	if parseErr != nil {
		_, _ = catchKongExit(func() {
			parser.FatalIfErrorf(parseErr)
		})
		return 1, false
	}
	return 0, true
}

// cmdNew implements the "debate new <name>" subcommand.
// It creates a persona file template under the discovered .heurema/debate/personas.
func cmdNew(args []string, stdout, stderr io.Writer, workDir string) int {
	cmd := newCmd{Role: "debater"}
	if code, ok := parseStandaloneCommand(&cmd, "debate new", args, stdout, stderr, workDir); !ok {
		return code
	}
	return runNew(&cmd, stdout, stderr, workDir)
}

func runNew(cmd *newCmd, stdout, stderr io.Writer, workDir string) int {
	if len(cmd.Name) == 0 {
		fmt.Fprintln(stderr, "error: debate new requires a persona name")
		printNewUsage(stderr)
		return 1
	}
	if len(cmd.Name) > 1 {
		fmt.Fprintln(stderr, "error: debate new takes exactly one positional argument")
		printNewUsage(stderr)
		return 1
	}

	name := cmd.Name[0]
	if err := validatePersonaName(name); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	if cmd.Role != "debater" && cmd.Role != "synthesizer" {
		fmt.Fprintf(stderr, "error: invalid role %q (must be debater or synthesizer)\n", cmd.Role)
		return 1
	}

	debDir, err := config.Discover(workDir)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	personasDir := filepath.Join(debDir, "personas")
	if err := os.MkdirAll(personasDir, 0755); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	personaPath := filepath.Join(personasDir, name+".md")
	created, err := writeIfAbsent(personaPath, buildPersonaTemplate(name, cmd.Role))
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	if !created {
		fmt.Fprintf(stderr, "error: persona %q already exists: %s\n", name, personaPath)
		return 1
	}
	fmt.Fprintln(stdout, "created", personaPath)
	return 0
}

func printNewUsage(stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: debate new [--role debater|synthesizer] <name>")
	fmt.Fprintln(stderr, "  Create a new persona file in the discovered .heurema/debate/personas.")
	fmt.Fprintln(stderr, "  -role string")
	fmt.Fprintln(stderr, "        persona role (debater|synthesizer) (default \"debater\")")
}

// validatePersonaName rejects names that are not simple identifiers.
// Only letters, digits, hyphens, and underscores are allowed to prevent path traversal.
func validatePersonaName(name string) error {
	if name == "" {
		return fmt.Errorf("persona name must not be empty")
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-') {
			return fmt.Errorf("persona name %q must contain only letters, digits, hyphens, and underscores", name)
		}
	}
	return nil
}

// buildPersonaTemplate returns a starter persona file for name with the given role.
func buildPersonaTemplate(name, role string) string {
	return fmt.Sprintf("---\nrole: %s\nmodel: %s\neffort: %s\n---\nYou are %s. Edit this system prompt to describe the persona's role and perspective.\n",
		role, scaffoldModel, scaffoldEffort, name)
}

// writeIfAbsent creates path with content only when path does not already exist.
// Returns (true, nil) when created, (false, nil) when the file already exists.
func writeIfAbsent(path, content string) (bool, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err == nil, err
}
