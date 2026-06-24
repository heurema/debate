package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

const contextTemplate = `<!-- Add context relevant to your debate topic here. -->
<!-- This preamble is prepended to every turn; keep it concise. -->
`

// cmdInit implements the "debate init" subcommand.
// It scaffolds a .heurema/debate workspace under workDir, skipping files that already exist.
func cmdInit(args []string, stdout, stderr io.Writer, workDir string) int {
	fs := flag.NewFlagSet("debate init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: debate init")
		fmt.Fprintln(stderr, "  Scaffold a .heurema/debate workspace in the current directory.")
	}
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() > 0 {
		fmt.Fprintln(stderr, "error: debate init takes no arguments")
		fs.Usage()
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
		{filepath.Join(debDir, "context.md"), contextTemplate},
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

// cmdNew implements the "debate new <name>" subcommand.
// It creates a persona file template under the discovered .heurema/debate/personas.
func cmdNew(args []string, stdout, stderr io.Writer, workDir string) int {
	var role string

	fs := flag.NewFlagSet("debate new", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&role, "role", "debater", "persona role (debater|synthesizer)")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: debate new [--role debater|synthesizer] <name>")
		fmt.Fprintln(stderr, "  Create a new persona file in the discovered .heurema/debate/personas.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() == 0 {
		fmt.Fprintln(stderr, "error: debate new requires a persona name")
		fs.Usage()
		return 1
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "error: debate new takes exactly one positional argument")
		fs.Usage()
		return 1
	}

	name := fs.Arg(0)
	if err := validatePersonaName(name); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	if role != "debater" && role != "synthesizer" {
		fmt.Fprintf(stderr, "error: invalid role %q (must be debater or synthesizer)\n", role)
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
	created, err := writeIfAbsent(personaPath, buildPersonaTemplate(name, role))
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
