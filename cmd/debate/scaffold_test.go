package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heurema/debate/internal/debate/config"
	"github.com/heurema/debate/internal/debate/persona"
)

func TestCmdInit_CreatesWorkspaceThatLoads(t *testing.T) {
	workDir := t.TempDir()
	var out, errout bytes.Buffer
	code := cmdInit(nil, &out, &errout, workDir)
	if code != 0 {
		t.Fatalf("cmdInit exit %d: stderr=%q", code, errout.String())
	}

	ws, err := config.Load(workDir, nil, "")
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if len(ws.Panel) != 2 {
		t.Fatalf("Panel len = %d, want 2", len(ws.Panel))
	}
	// Lexicographic: proposer < skeptic
	if ws.Panel[0].ID != "proposer" || ws.Panel[1].ID != "skeptic" {
		t.Errorf("Panel IDs = [%s, %s], want [proposer, skeptic]", ws.Panel[0].ID, ws.Panel[1].ID)
	}
	if ws.Synthesizer.Role != "synthesizer" {
		t.Errorf("Synthesizer role = %q, want synthesizer", ws.Synthesizer.Role)
	}

	contextPath := filepath.Join(workDir, ".heurema", "debate", "context.md")
	if _, err := os.Stat(contextPath); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("context.md should not be created by init; stat err: %v", err)
	}
}

func TestCmdInit_StarterPersonasParseable(t *testing.T) {
	workDir := t.TempDir()
	if code := cmdInit(nil, &bytes.Buffer{}, &bytes.Buffer{}, workDir); code != 0 {
		t.Fatal("cmdInit failed")
	}

	personasDir := filepath.Join(workDir, ".heurema", "debate", "personas")
	for _, name := range []string{"proposer", "skeptic"} {
		p, err := persona.ParseFile(filepath.Join(personasDir, name+".md"))
		if err != nil {
			t.Errorf("persona.ParseFile(%s): %v", name, err)
			continue
		}
		if p.ID != name {
			t.Errorf("%s: ID = %q, want %q", name, p.ID, name)
		}
		if p.Role != "debater" {
			t.Errorf("%s: Role = %q, want debater", name, p.Role)
		}
	}
}

func TestCmdInit_DoesNotOverwriteExisting(t *testing.T) {
	workDir := t.TempDir()
	if code := cmdInit(nil, &bytes.Buffer{}, &bytes.Buffer{}, workDir); code != 0 {
		t.Fatal("first cmdInit failed")
	}

	// Replace proposer.md with sentinel content.
	proposerPath := filepath.Join(workDir, ".heurema", "debate", "personas", "proposer.md")
	sentinel := "sentinel content\n"
	if err := os.WriteFile(proposerPath, []byte(sentinel), 0644); err != nil {
		t.Fatal(err)
	}

	// Second init must not overwrite.
	var out bytes.Buffer
	if code := cmdInit(nil, &out, &bytes.Buffer{}, workDir); code != 0 {
		t.Fatalf("second cmdInit exit non-zero")
	}

	got, err := os.ReadFile(proposerPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != sentinel {
		t.Errorf("proposer.md overwritten; got %q, want %q", string(got), sentinel)
	}
	if !strings.Contains(out.String(), "skipped") {
		t.Errorf("expected 'skipped' in output, got: %q", out.String())
	}
}

func TestCmdInit_ExtraArgsError(t *testing.T) {
	workDir := t.TempDir()
	var errout bytes.Buffer
	code := cmdInit([]string{"unexpected"}, &bytes.Buffer{}, &errout, workDir)
	if code == 0 {
		t.Error("expected non-zero exit for extra args to init")
	}
}

func TestCmdNew_CreatesPersonaThatParses(t *testing.T) {
	workDir := t.TempDir()
	if code := cmdInit(nil, &bytes.Buffer{}, &bytes.Buffer{}, workDir); code != 0 {
		t.Fatal("cmdInit failed")
	}

	var out, errout bytes.Buffer
	code := cmdNew([]string{"analyst"}, &out, &errout, workDir)
	if code != 0 {
		t.Fatalf("cmdNew exit %d: stderr=%q", code, errout.String())
	}

	personaPath := filepath.Join(workDir, ".heurema", "debate", "personas", "analyst.md")
	p, err := persona.ParseFile(personaPath)
	if err != nil {
		t.Fatalf("persona.ParseFile: %v", err)
	}
	if p.ID != "analyst" {
		t.Errorf("ID = %q, want analyst", p.ID)
	}
	if p.Role != "debater" {
		t.Errorf("Role = %q, want debater", p.Role)
	}
	if strings.Contains(out.String(), personaPath) == false {
		t.Errorf("expected output to mention created path, got: %q", out.String())
	}
}

func TestCmdNew_SynthesizerRole(t *testing.T) {
	workDir := t.TempDir()
	if code := cmdInit(nil, &bytes.Buffer{}, &bytes.Buffer{}, workDir); code != 0 {
		t.Fatal("cmdInit failed")
	}

	var errout bytes.Buffer
	code := cmdNew([]string{"--role", "synthesizer", "mysynth"}, &bytes.Buffer{}, &errout, workDir)
	if code != 0 {
		t.Fatalf("cmdNew exit %d: stderr=%q", code, errout.String())
	}

	personaPath := filepath.Join(workDir, ".heurema", "debate", "personas", "mysynth.md")
	p, err := persona.ParseFile(personaPath)
	if err != nil {
		t.Fatalf("persona.ParseFile: %v", err)
	}
	if p.Role != "synthesizer" {
		t.Errorf("Role = %q, want synthesizer", p.Role)
	}
}

func TestCmdNew_RoleFlagBeforeAndAfterName(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		file string
	}{
		{name: "before", args: []string{"--role", "synthesizer", "before"}, file: "before.md"},
		{name: "after", args: []string{"after", "--role", "synthesizer"}, file: "after.md"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			workDir := t.TempDir()
			if code := cmdInit(nil, &bytes.Buffer{}, &bytes.Buffer{}, workDir); code != 0 {
				t.Fatal("cmdInit failed")
			}

			args := append([]string{"new"}, tc.args...)
			var errout bytes.Buffer
			code := parseCLI(args, &bytes.Buffer{}, &errout, strings.NewReader(""), false, noEnv, echoAll, workDir)
			if code != 0 {
				t.Fatalf("cmdNew exit %d: stderr=%q", code, errout.String())
			}

			personaPath := filepath.Join(workDir, ".heurema", "debate", "personas", tc.file)
			p, err := persona.ParseFile(personaPath)
			if err != nil {
				t.Fatalf("persona.ParseFile: %v", err)
			}
			if p.Role != "synthesizer" {
				t.Errorf("Role = %q, want synthesizer", p.Role)
			}
		})
	}
}

func TestCmdNew_RefusesOverwrite(t *testing.T) {
	workDir := t.TempDir()
	if code := cmdInit(nil, &bytes.Buffer{}, &bytes.Buffer{}, workDir); code != 0 {
		t.Fatal("cmdInit failed")
	}
	if code := cmdNew([]string{"analyst"}, &bytes.Buffer{}, &bytes.Buffer{}, workDir); code != 0 {
		t.Fatal("first cmdNew failed")
	}

	var errout bytes.Buffer
	code := cmdNew([]string{"analyst"}, &bytes.Buffer{}, &errout, workDir)
	if code == 0 {
		t.Error("expected non-zero exit when persona already exists")
	}
	if !strings.Contains(errout.String(), "already exists") {
		t.Errorf("expected 'already exists' in stderr, got: %q", errout.String())
	}
}

func TestCmdNew_RejectsPathSeparators(t *testing.T) {
	workDir := t.TempDir()
	cases := []string{"foo/bar", `foo\bar`, "../evil", "foo.bar"}
	for _, name := range cases {
		var errout bytes.Buffer
		code := cmdNew([]string{name}, &bytes.Buffer{}, &errout, workDir)
		if code == 0 {
			t.Errorf("expected non-zero exit for name %q", name)
		}
	}
}

func TestCmdNew_RequiresWorkspace(t *testing.T) {
	workDir := t.TempDir() // no .heurema/debate
	var errout bytes.Buffer
	code := cmdNew([]string{"analyst"}, &bytes.Buffer{}, &errout, workDir)
	if code == 0 {
		t.Error("expected non-zero exit when no .heurema/debate found")
	}
	if !strings.Contains(errout.String(), ".heurema/debate") {
		t.Errorf("expected error about .heurema/debate, got: %q", errout.String())
	}
}

func TestCmdNew_InvalidRole(t *testing.T) {
	workDir := t.TempDir()
	if code := cmdInit(nil, &bytes.Buffer{}, &bytes.Buffer{}, workDir); code != 0 {
		t.Fatal("cmdInit failed")
	}

	var errout bytes.Buffer
	code := cmdNew([]string{"--role", "moderator", "foo"}, &bytes.Buffer{}, &errout, workDir)
	if code == 0 {
		t.Error("expected non-zero exit for invalid role")
	}
}

func TestCmdNew_MissingName(t *testing.T) {
	workDir := t.TempDir()
	var errout bytes.Buffer
	code := cmdNew(nil, &bytes.Buffer{}, &errout, workDir)
	if code == 0 {
		t.Error("expected non-zero exit when name omitted")
	}
}

func TestCmdNew_CreatesPersonasDirIfAbsent(t *testing.T) {
	workDir := t.TempDir()
	// Create workspace directory but no personas subdir.
	debDir := filepath.Join(workDir, ".heurema", "debate")
	if err := os.MkdirAll(debDir, 0755); err != nil {
		t.Fatal(err)
	}

	var errout bytes.Buffer
	code := cmdNew([]string{"analyst"}, &bytes.Buffer{}, &errout, workDir)
	if code != 0 {
		t.Fatalf("cmdNew exit %d: stderr=%q", code, errout.String())
	}

	personaPath := filepath.Join(debDir, "personas", "analyst.md")
	if _, err := os.Stat(personaPath); err != nil {
		t.Errorf("expected persona file to exist: %v", err)
	}
}
