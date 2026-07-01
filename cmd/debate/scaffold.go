package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/heurema/debate/internal/debate/capability"
	"github.com/heurema/debate/internal/debate/config"
	"github.com/heurema/debate/internal/debate/skills"
	"github.com/heurema/debate/internal/debate/skills/bundled"
)

const (
	scaffoldModel  = "claude-haiku-4-5"
	scaffoldEffort = "medium"
)

// lookExecutable and userHomeDir are test seams: production wires the real
// PATH and home directory, tests override them so skill installation and
// starter-persona capability detection never touch the real machine.
var (
	lookExecutable capability.LookPath = capability.DefaultLookup
	userHomeDir                        = os.UserHomeDir
)

// unsetFamily is the placeholder written into starter personas when no
// supported local agent executable is found on PATH (AC9).
var unsetFamily = capability.Family{Model: "unset", Backend: "unset"}

const proposerBody = `You are the Proposer. Build the strongest practical solution to the task and defend it against the Skeptic's objections. When the Skeptic surfaces a real blocker, revise the proposal instead of defending a broken position; when an objection is a nice-to-have, say so plainly and hold your ground. Be concrete: cite the specific mechanism, tradeoff, or evidence behind each claim.`

const skepticBody = `You are the Skeptic. Find the blocking risks in the Proposer's solution: weak assumptions, unhandled edge cases, missing validation, and failure modes that would break it in practice. Distinguish blockers (must be fixed before this can ship) from nice-to-haves (would improve it but aren't disqualifying), and say which is which. Be constructive: point at the specific gap, not just that you disagree.`

func personaTemplate(role string, family capability.Family, body string) string {
	return fmt.Sprintf("---\nversion: 1\nrole: %s\nmodel: %s\nbackend: %s\neffort: medium\n---\n%s\n",
		role, family.Model, family.Backend, body)
}

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

	family, detected := capability.Detect(lookExecutable)
	if !detected {
		family = unsetFamily
	}

	proposerPath := filepath.Join(personasDir, "proposer.md")
	skepticPath := filepath.Join(personasDir, "skeptic.md")
	files := []struct {
		path    string
		content string
	}{
		{proposerPath, personaTemplate("debater", family, proposerBody)},
		{skepticPath, personaTemplate("debater", family, skepticBody)},
		{filepath.Join(tablesDir, "default.yml"), defaultTableTemplate},
	}

	starterCreated := make(map[string]bool, 2)
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
		starterCreated[f.path] = created
	}

	if !detected {
		var unsetPaths []string
		for _, path := range []string{proposerPath, skepticPath} {
			if starterCreated[path] {
				unsetPaths = append(unsetPaths, path)
			}
		}
		if len(unsetPaths) > 0 {
			fmt.Fprintf(stderr, "warning: %s: model and backend were set to %q because no supported executable "+
				"(claude, codex, agy, or gemini) was found on PATH; edit these fields to one of the supported "+
				"(model, backend) pairs (claude-haiku-4-5/claude-agent-acp, codex/codex-acp, gemini-pro/agy) before "+
				"running a debate with these personas.\n",
				strings.Join(unsetPaths, " and "), unsetFamily.Model)
		}
	}

	installSkill(stdout, stderr)
	return 0
}

// installSkill installs or repairs the bundled debate Agent Skill for
// detected local agent clients. It never fails init: unwritable or unsafe
// targets, a missing home directory, or no detected client are reported as
// warnings on stderr rather than errors.
func installSkill(stdout, stderr io.Writer) {
	home, err := userHomeDir()
	if err != nil {
		home = ""
	}
	results := skills.InstallOrRepair(skills.Options{
		Home:          home,
		LookPath:      lookExecutable,
		Bundled:       bundled.Skill(),
		BinaryVersion: Version,
	})
	for _, r := range results {
		if r.Path != "" {
			fmt.Fprintln(stdout, r.Action, r.Path)
		}
		if r.Warning != "" {
			fmt.Fprintln(stderr, "warning:", r.Warning)
		}
	}
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
