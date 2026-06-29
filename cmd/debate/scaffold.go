package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/heurema/debate/internal/debate/config"
)

const (
	scaffoldModel  = "claude-haiku-4-5"
	scaffoldEffort = "medium"
)

const proposerTemplate = `---
version: 1
role: debater
model: claude-haiku-4-5
effort: medium
---
You are the Proposer. Defend the proposition with clear arguments and respond to the Skeptic's objections. Be concise and direct.
`

const skepticTemplate = `---
version: 1
role: debater
model: claude-haiku-4-5
effort: medium
---
You are the Skeptic. Challenge the proposition by identifying weaknesses, counter-examples, and unresolved assumptions. Be constructive and specific.
`

const defaultTableTemplate = `version: 1
panel:
  - proposer
  - skeptic
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
	cliArgs := append([]string{"init"}, args...)
	return parseCLI(cliArgs, stdout, stderr, strings.NewReader(""), defaultResolver, workDir)
}

func runInit(cmd *initCmd, stdout, stderr io.Writer, workDir string) int {
	if len(cmd.Args) > 0 {
		fmt.Fprintln(stderr, "error: debate init takes no arguments")
		printInitUsage(stderr)
		return 1
	}

	debDir := filepath.Join(workDir, ".heurema", "debate")
	personasDir := filepath.Join(debDir, "personas")
	tablesDir := filepath.Join(debDir, "tables")
	if err := os.MkdirAll(personasDir, 0755); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	if err := os.MkdirAll(tablesDir, 0755); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	files := []struct {
		path    string
		content string
	}{
		{filepath.Join(personasDir, "proposer.md"), proposerTemplate},
		{filepath.Join(personasDir, "skeptic.md"), skepticTemplate},
		{filepath.Join(tablesDir, "default.yml"), defaultTableTemplate},
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

// cmdNew implements the "debate new <name>" subcommand.
// It creates a persona file template under the discovered .heurema/debate/personas.
func cmdNew(args []string, stdout, stderr io.Writer, workDir string) int {
	cliArgs := append([]string{"new"}, args...)
	return parseCLI(cliArgs, stdout, stderr, strings.NewReader(""), defaultResolver, workDir)
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

	personaID := cmd.Name[0]
	segments, err := validatePersonaID(personaID)
	if err != nil {
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
	targetDir := personasDir
	if len(segments) == 2 {
		targetDir = filepath.Join(personasDir, segments[0])
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	name := segments[len(segments)-1]
	personaPath := filepath.Join(targetDir, name+".md")
	created, err := writeIfAbsent(personaPath, buildPersonaTemplate(personaID, cmd.Role))
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	if !created {
		fmt.Fprintf(stderr, "error: persona %q already exists: %s\n", personaID, personaPath)
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

// validatePersonaID rejects IDs outside the v1 root-or-one-namespace form.
func validatePersonaID(id string) ([]string, error) {
	if id == "" {
		return nil, fmt.Errorf("persona name must not be empty")
	}
	if filepath.IsAbs(id) {
		return nil, fmt.Errorf("persona name %q must be relative", id)
	}
	if strings.Contains(id, `\`) {
		return nil, fmt.Errorf("persona name %q must use '/' as the namespace separator", id)
	}
	segments := strings.Split(id, "/")
	if len(segments) > 2 {
		return nil, fmt.Errorf("persona name %q is too deep; use name or namespace/name", id)
	}
	for _, segment := range segments {
		if !config.ValidSegment(segment) {
			return nil, fmt.Errorf("persona name %q must use non-empty segments containing only letters, digits, hyphens, and underscores", id)
		}
	}
	return segments, nil
}

// buildPersonaTemplate returns a starter persona file for name with the given role.
func buildPersonaTemplate(name, role string) string {
	return fmt.Sprintf("---\nversion: 1\nrole: %s\nmodel: %s\neffort: %s\n---\nYou are %s. Edit this system prompt to describe the persona's role and perspective.\n",
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
