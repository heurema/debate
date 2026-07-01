package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heurema/debate/internal/debate/capability"
	"github.com/heurema/debate/internal/debate/config"
	"github.com/heurema/debate/internal/debate/persona"
)

// fakeEnvironment overrides the lookExecutable and userHomeDir test seams so
// that runInit's capability detection and skill installation never touch the
// real PATH or the real user home directory. It restores both after the test.
// found lists which of claude/codex/agy/gemini executables are simulated on
// PATH; home is used as the fake HOME for global skill installation.
func fakeEnvironment(t *testing.T, home string, found ...string) {
	t.Helper()
	fakeEnvironmentWithHome(t, func() (string, error) { return home, nil }, found...)
}

func fakeEnvironmentWithHome(t *testing.T, homeFn func() (string, error), found ...string) {
	t.Helper()
	origLook, origHome := lookExecutable, userHomeDir
	set := make(map[string]bool, len(found))
	for _, f := range found {
		set[f] = true
	}
	lookExecutable = func(name string) (string, error) {
		if set[name] {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found")
	}
	userHomeDir = homeFn
	t.Cleanup(func() {
		lookExecutable = origLook
		userHomeDir = origHome
	})
}

func TestCmdInit_CreatesWorkspaceThatLoads(t *testing.T) {
	workDir := t.TempDir()
	fakeEnvironment(t, t.TempDir(), "claude")
	var out, errout bytes.Buffer
	code := cmdInit(nil, &out, &errout, workDir)
	if code != 0 {
		t.Fatalf("cmdInit exit %d: stderr=%q", code, errout.String())
	}

	ws, err := config.Load(workDir, "", nil, "")
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

	tablePath := filepath.Join(workDir, ".heurema", "debate", "tables", "default.yml")
	table, err := os.ReadFile(tablePath)
	if err != nil {
		t.Fatalf("default table missing: %v", err)
	}
	if !strings.Contains(string(table), "version: 1") ||
		!strings.Contains(string(table), "proposer") ||
		!strings.Contains(string(table), "skeptic") {
		t.Errorf("default table content is not v1 panel: %q", string(table))
	}
}

func TestCmdInit_StarterPersonasParseable(t *testing.T) {
	workDir := t.TempDir()
	fakeEnvironment(t, t.TempDir(), "claude")
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
	fakeEnvironment(t, t.TempDir(), "claude")
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
	fakeEnvironment(t, t.TempDir(), "claude")
	var errout bytes.Buffer
	code := cmdInit([]string{"unexpected"}, &bytes.Buffer{}, &errout, workDir)
	if code == 0 {
		t.Error("expected non-zero exit for extra args to init")
	}
}

func TestCmdNew_CreatesPersonaThatParses(t *testing.T) {
	workDir := t.TempDir()
	fakeEnvironment(t, t.TempDir(), "claude")
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
	fakeEnvironment(t, t.TempDir(), "claude")
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

func TestCmdNew_CreatesNamespacedPersona(t *testing.T) {
	workDir := t.TempDir()
	fakeEnvironment(t, t.TempDir(), "claude")
	if code := cmdInit(nil, &bytes.Buffer{}, &bytes.Buffer{}, workDir); code != 0 {
		t.Fatal("cmdInit failed")
	}

	var out, errout bytes.Buffer
	code := cmdNew([]string{"team/analyst"}, &out, &errout, workDir)
	if code != 0 {
		t.Fatalf("cmdNew exit %d: stderr=%q", code, errout.String())
	}

	personaPath := filepath.Join(workDir, ".heurema", "debate", "personas", "team", "analyst.md")
	p, err := persona.ParseFileWithID(personaPath, "team/analyst")
	if err != nil {
		t.Fatalf("persona.ParseFileWithID: %v", err)
	}
	if p.ID != "team/analyst" {
		t.Errorf("ID = %q, want team/analyst", p.ID)
	}
	if !strings.Contains(out.String(), personaPath) {
		t.Errorf("expected output to mention created path, got: %q", out.String())
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
			fakeEnvironment(t, t.TempDir(), "claude")
			if code := cmdInit(nil, &bytes.Buffer{}, &bytes.Buffer{}, workDir); code != 0 {
				t.Fatal("cmdInit failed")
			}

			args := append([]string{"new"}, tc.args...)
			var errout bytes.Buffer
			code := parseCLI(args, &bytes.Buffer{}, &errout, strings.NewReader(""), echoAll, workDir)
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
	fakeEnvironment(t, t.TempDir(), "claude")
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
	cases := []string{`foo\bar`, "../evil", "foo.bar", "/absolute", "too/deep/name", "foo/", "/foo", "foo//bar"}
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
	fakeEnvironment(t, t.TempDir(), "claude")
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

func TestRunInit_StarterPersonaDefaults_CapabilityAware(t *testing.T) {
	cases := []struct {
		name    string
		found   []string
		model   string
		backend string
	}{
		{"claude", []string{"claude"}, capability.Claude.Model, capability.Claude.Backend},
		{"codex", []string{"codex"}, capability.Codex.Model, capability.Codex.Backend},
		{"agy", []string{"agy"}, capability.Gemini.Model, capability.Gemini.Backend},
		{"gemini", []string{"gemini"}, capability.Gemini.Model, capability.Gemini.Backend},
		{"claude and codex prefers claude", []string{"claude", "codex"}, capability.Claude.Model, capability.Claude.Backend},
		{"codex and agy prefers codex", []string{"codex", "agy"}, capability.Codex.Model, capability.Codex.Backend},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			workDir := t.TempDir()
			fakeEnvironment(t, t.TempDir(), tc.found...)
			var out, errout bytes.Buffer
			if code := cmdInit(nil, &out, &errout, workDir); code != 0 {
				t.Fatalf("cmdInit exit non-zero: stderr=%q", errout.String())
			}
			personasDir := filepath.Join(workDir, ".heurema", "debate", "personas")
			for _, name := range []string{"proposer", "skeptic"} {
				p, err := persona.ParseFile(filepath.Join(personasDir, name+".md"))
				if err != nil {
					t.Fatalf("persona.ParseFile(%s): %v", name, err)
				}
				if p.Model != tc.model || p.Backend != tc.backend {
					t.Errorf("%s: (Model, Backend) = (%q, %q), want (%q, %q)", name, p.Model, p.Backend, tc.model, tc.backend)
				}
			}
			if strings.Contains(errout.String(), "unset") {
				t.Errorf("unexpected unset-placeholder warning when a supported executable is detected: %q", errout.String())
			}
		})
	}
}

func TestRunInit_NoSupportedExecutable_UsesUnsetPlaceholderAndWarns(t *testing.T) {
	workDir := t.TempDir()
	fakeEnvironment(t, t.TempDir()) // no executables found
	var out, errout bytes.Buffer
	if code := cmdInit(nil, &out, &errout, workDir); code != 0 {
		t.Fatalf("cmdInit exit non-zero: stderr=%q", errout.String())
	}

	personasDir := filepath.Join(workDir, ".heurema", "debate", "personas")
	proposerPath := filepath.Join(personasDir, "proposer.md")
	skepticPath := filepath.Join(personasDir, "skeptic.md")
	for _, path := range []string{proposerPath, skepticPath} {
		p, err := persona.ParseFile(path)
		if err != nil {
			t.Fatalf("persona.ParseFile(%s): %v", path, err)
		}
		if p.Model != "unset" || p.Backend != "unset" {
			t.Errorf("%s: (Model, Backend) = (%q, %q), want (unset, unset)", path, p.Model, p.Backend)
		}
	}

	// Exactly one warning names the starter personas (a separate warning about
	// no detected skill-install client is also expected and is not counted here).
	occurrences := strings.Count(errout.String(), proposerPath)
	if occurrences != 1 {
		t.Fatalf("expected exactly one warning naming %s, got %d: %q", proposerPath, occurrences, errout.String())
	}
	for _, want := range []string{proposerPath, skepticPath, "unset", "claude, codex, agy, or gemini"} {
		if !strings.Contains(errout.String(), want) {
			t.Errorf("warning missing %q: %q", want, errout.String())
		}
	}
}

func TestRunInit_NoSupportedExecutable_WarnsOnlyForCreatedUnsetPersonas(t *testing.T) {
	workDir := t.TempDir()
	fakeEnvironment(t, t.TempDir()) // no executables found
	personasDir := filepath.Join(workDir, ".heurema", "debate", "personas")
	if err := os.MkdirAll(personasDir, 0755); err != nil {
		t.Fatal(err)
	}
	proposerPath := filepath.Join(personasDir, "proposer.md")
	if err := os.WriteFile(proposerPath, []byte(personaTemplate("debater", capability.Claude, proposerBody)), 0644); err != nil {
		t.Fatal(err)
	}

	var out, errout bytes.Buffer
	if code := cmdInit(nil, &out, &errout, workDir); code != 0 {
		t.Fatalf("cmdInit exit non-zero: stderr=%q", errout.String())
	}

	skepticPath := filepath.Join(personasDir, "skeptic.md")
	stderr := errout.String()
	if strings.Contains(stderr, proposerPath) {
		t.Fatalf("warning should not claim the preserved proposer was set to unset: %q", stderr)
	}
	for _, want := range []string{skepticPath, "unset", "claude, codex, agy, or gemini"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("warning missing %q: %q", want, stderr)
		}
	}
}

func TestRunInit_ExistingHomeDirsDoNotAffectPersonaDefaults(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}
	workDir := t.TempDir()
	fakeEnvironment(t, home) // no executables found, despite existing home dirs
	var out, errout bytes.Buffer
	if code := cmdInit(nil, &out, &errout, workDir); code != 0 {
		t.Fatalf("cmdInit exit non-zero: stderr=%q", errout.String())
	}

	p, err := persona.ParseFile(filepath.Join(workDir, ".heurema", "debate", "personas", "proposer.md"))
	if err != nil {
		t.Fatalf("persona.ParseFile: %v", err)
	}
	if p.Model != "unset" || p.Backend != "unset" {
		t.Errorf("(Model, Backend) = (%q, %q), want (unset, unset); existing home dirs must not choose runtime defaults", p.Model, p.Backend)
	}
}

func TestRunInit_SkillInstall_MissingHomeStillSucceeds(t *testing.T) {
	workDir := t.TempDir()
	fakeEnvironmentWithHome(t, func() (string, error) {
		return "", errors.New("home unavailable")
	}, "claude")
	var out, errout bytes.Buffer
	if code := cmdInit(nil, &out, &errout, workDir); code != 0 {
		t.Fatalf("cmdInit exit non-zero: stderr=%q", errout.String())
	}

	if !strings.Contains(errout.String(), "HOME is not set") {
		t.Fatalf("expected missing-HOME warning, got stderr=%q", errout.String())
	}
	if strings.Contains(out.String(), ".agents/skills/debate") || strings.Contains(out.String(), ".claude/skills/debate") {
		t.Fatalf("missing HOME should not report a global skill target, got stdout=%q", out.String())
	}
	if _, err := os.Stat(filepath.Join(workDir, ".heurema", "debate", "tables", "default.yml")); err != nil {
		t.Fatalf("init should still create the workspace when HOME is missing: %v", err)
	}
}

func TestRunInit_SkillInstall_ClientDetectionMatrix(t *testing.T) {
	cases := []struct {
		name       string
		found      []string
		wantAgents bool
		wantClaude bool
	}{
		{"none", nil, false, false},
		{"codex", []string{"codex"}, true, false},
		{"gemini", []string{"gemini"}, true, false},
		{"claude", []string{"claude"}, false, true},
		{"codex and claude", []string{"codex", "claude"}, true, true},
		{"codex and gemini", []string{"codex", "gemini"}, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			workDir := t.TempDir()
			fakeEnvironment(t, home, tc.found...)
			var out, errout bytes.Buffer
			if code := cmdInit(nil, &out, &errout, workDir); code != 0 {
				t.Fatalf("cmdInit exit non-zero: stderr=%q", errout.String())
			}

			agentsPath := filepath.Join(home, ".agents", "skills", "debate")
			claudePath := filepath.Join(home, ".claude", "skills", "debate")
			if _, err := os.Stat(agentsPath); (err == nil) != tc.wantAgents {
				t.Errorf("agents skill target exists = %v, want %v", err == nil, tc.wantAgents)
			}
			if _, err := os.Stat(claudePath); (err == nil) != tc.wantClaude {
				t.Errorf("claude skill target exists = %v, want %v", err == nil, tc.wantClaude)
			}
			for _, sub := range []string{".codex", ".gemini"} {
				if _, err := os.Stat(filepath.Join(home, sub)); !os.IsNotExist(err) {
					t.Errorf("%s should never be created by skill installation", sub)
				}
			}
			if !tc.wantAgents && !tc.wantClaude {
				if !strings.Contains(errout.String(), "warning:") {
					t.Errorf("expected an explanatory warning when no client is detected, got stderr=%q", errout.String())
				}
			}
		})
	}
}

func TestRunInit_SkillInstall_IdempotentAndPreservesLocalEdits(t *testing.T) {
	home := t.TempDir()
	workDir := t.TempDir()
	fakeEnvironment(t, home, "claude")

	var out1 bytes.Buffer
	if code := cmdInit(nil, &out1, &bytes.Buffer{}, workDir); code != 0 {
		t.Fatal("first cmdInit failed")
	}
	target := filepath.Join(home, ".claude", "skills", "debate")
	if !strings.Contains(out1.String(), "created "+target) {
		t.Fatalf("expected first init to report created %s, got: %q", target, out1.String())
	}

	// Re-running init on unmodified content is idempotent.
	workDir2 := t.TempDir()
	var out2 bytes.Buffer
	if code := cmdInit(nil, &out2, &bytes.Buffer{}, workDir2); code != 0 {
		t.Fatal("second cmdInit failed")
	}
	if !strings.Contains(out2.String(), target) || strings.Contains(out2.String(), "created "+target) {
		t.Errorf("expected second init to report the target as current/skipped, not created, got: %q", out2.String())
	}

	// Locally modifying the installed skill preserves it on the next init.
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("locally edited"), 0644); err != nil {
		t.Fatal(err)
	}
	workDir3 := t.TempDir()
	var out3, errout3 bytes.Buffer
	if code := cmdInit(nil, &out3, &errout3, workDir3); code != 0 {
		t.Fatal("third cmdInit failed")
	}
	if !strings.Contains(errout3.String(), "warning:") {
		t.Errorf("expected a warning for the locally modified target, got stderr=%q", errout3.String())
	}
	data, err := os.ReadFile(filepath.Join(target, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "locally edited" {
		t.Errorf("locally edited content was overwritten: %q", data)
	}
}

func TestNoDebateSkillsCommand(t *testing.T) {
	if isNamedCommand("skills") {
		t.Fatal("\"skills\" must not be registered as a named command")
	}

	// With no workspace present, "debate skills" is treated as a run task and
	// fails through the normal CLI error path rather than any dedicated
	// skill-management command.
	workDir := t.TempDir()
	var errout bytes.Buffer
	code := parseCLI([]string{"skills"}, &bytes.Buffer{}, &errout, strings.NewReader(""), echoAll, workDir)
	if code == 0 {
		t.Error("expected non-zero exit for \"debate skills\" with no workspace present")
	}
}
