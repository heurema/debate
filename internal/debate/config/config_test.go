package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heurema/debate/internal/debate/config"
)

// makeDebateDir creates a .heurema/debate directory under a temp root and
// writes the given files relative to .heurema/debate. Returns the temp root.
func makeDebateDir(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	debDir := filepath.Join(root, ".heurema", "debate")
	for rel, content := range files {
		full := filepath.Join(debDir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const alicePersona = `---
model: claude-sonnet-4-6
effort: high
---
You are Alice, a careful logical reasoner.
`

const bobPersona = `---
model: claude-opus-4-8
effort: medium
---
You are Bob, a pragmatic problem solver.
`

const synthPersona = `---
role: synthesizer
model: claude-haiku-4-5
effort: low
---
You are the synthesizer. Summarize the discussion.
`

// TestDiscover_FindsFromChildDir verifies that Discover walks up to find .heurema/debate.
func TestDiscover_FindsFromChildDir(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
	})
	child := filepath.Join(root, "some", "nested", "dir")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}
	found, err := config.Discover(child)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	want := filepath.Join(root, ".heurema", "debate")
	if found != want {
		t.Errorf("Discover = %q, want %q", found, want)
	}
}

// TestDiscover_MissingDir verifies Discover returns an error when no .heurema/debate exists.
func TestDiscover_MissingDir(t *testing.T) {
	tmp := t.TempDir()
	_, err := config.Discover(tmp)
	if err == nil {
		t.Fatal("expected error when .heurema/debate absent, got nil")
	}
}

// TestLoad_ValidWorkspace loads a workspace with two debaters and a synthesizer persona.
func TestLoad_ValidWorkspace(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":       alicePersona,
		"personas/bob.md":         bobPersona,
		"personas/synthesizer.md": synthPersona,
		"context.md":              "# Discussion context\n\nThis is the baseline.",
		"config.yml":              "table:\n  - alice\n  - bob\n",
	})
	ws, err := config.Load(root, nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ws.Panel) != 2 {
		t.Errorf("Panel len = %d, want 2", len(ws.Panel))
	}
	if ws.Panel[0].ID != "alice" || ws.Panel[1].ID != "bob" {
		t.Errorf("Panel IDs = [%s, %s], want [alice, bob]", ws.Panel[0].ID, ws.Panel[1].ID)
	}
	if ws.Synthesizer.ID != "synthesizer" {
		t.Errorf("Synthesizer ID = %q, want synthesizer", ws.Synthesizer.ID)
	}
	if ws.Synthesizer.Role != "synthesizer" {
		t.Errorf("Synthesizer Role = %q, want synthesizer", ws.Synthesizer.Role)
	}
	if !strings.Contains(ws.Context, "baseline") {
		t.Errorf("Context missing expected content: %q", ws.Context)
	}
}

// TestLoad_DefaultPanel verifies that with no config table and no withList, all debaters are used.
func TestLoad_DefaultPanel(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/bob.md":   bobPersona,
		"personas/alice.md": alicePersona,
	})
	ws, err := config.Load(root, nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Lexicographic order: alice < bob
	if len(ws.Panel) != 2 {
		t.Fatalf("Panel len = %d, want 2", len(ws.Panel))
	}
	if ws.Panel[0].ID != "alice" || ws.Panel[1].ID != "bob" {
		t.Errorf("Panel = [%s, %s], want [alice, bob]", ws.Panel[0].ID, ws.Panel[1].ID)
	}
}

// TestLoad_WithListOverridesConfig verifies explicit withList takes precedence and preserves order.
func TestLoad_WithListOverridesConfig(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
		"personas/bob.md":   bobPersona,
		"config.yml":        "table:\n  - alice\n",
	})
	ws, err := config.Load(root, []string{"bob", "alice"}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ws.Panel) != 2 {
		t.Fatalf("Panel len = %d, want 2", len(ws.Panel))
	}
	if ws.Panel[0].ID != "bob" || ws.Panel[1].ID != "alice" {
		t.Errorf("Panel = [%s, %s], want [bob, alice]", ws.Panel[0].ID, ws.Panel[1].ID)
	}
}

// TestLoad_DefaultSynthesizer verifies the built-in default synthesizer values.
func TestLoad_DefaultSynthesizer(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
	})
	ws, err := config.Load(root, nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	s := ws.Synthesizer
	if s.Role != "synthesizer" {
		t.Errorf("default synth Role = %q, want synthesizer", s.Role)
	}
	if s.Model != "claude-haiku-4-5" {
		t.Errorf("default synth Model = %q, want claude-haiku-4-5", s.Model)
	}
	if s.Effort != "low" {
		t.Errorf("default synth Effort = %q, want low", s.Effort)
	}
	if s.Backend != "claude-agent-acp" {
		t.Errorf("default synth Backend = %q, want claude-agent-acp", s.Backend)
	}
	if s.System == "" {
		t.Error("default synth System must be non-empty")
	}
}

// TestLoad_SynthOverride verifies an explicit synthesizer override.
func TestLoad_SynthOverride(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":       alicePersona,
		"personas/synthesizer.md": synthPersona,
	})
	ws, err := config.Load(root, nil, "synthesizer")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Synthesizer.ID != "synthesizer" {
		t.Errorf("Synthesizer ID = %q, want synthesizer", ws.Synthesizer.ID)
	}
}

// TestLoad_SynthOverrideMissing verifies a missing synth override is a fail-fast error.
func TestLoad_SynthOverrideMissing(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
	})
	_, err := config.Load(root, nil, "nosuchpersona")
	if err == nil {
		t.Fatal("expected error for missing synthesizer override, got nil")
	}
	if !strings.Contains(err.Error(), "nosuchpersona") {
		t.Errorf("error should name the missing persona: %v", err)
	}
}

// TestLoad_NoContextMD verifies that a missing context.md is allowed.
func TestLoad_NoContextMD(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
	})
	ws, err := config.Load(root, nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Context != "" {
		t.Errorf("Context = %q, want empty string when context.md absent", ws.Context)
	}
}

// TestLoad_UnknownConfigKey verifies an unknown key in config.yml is a fail-fast error.
func TestLoad_UnknownConfigKey(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
		"config.yml":        "table:\n  - alice\nunknown_key: oops\n",
	})
	_, err := config.Load(root, nil, "")
	if err == nil {
		t.Fatal("expected error for unknown config key, got nil")
	}
}

// TestLoad_UnknownFrontmatterKey verifies a persona with an unknown frontmatter key fails.
func TestLoad_UnknownFrontmatterKey(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": `---
model: claude-sonnet-4-6
effort: high
mystery: field
---
Alice's system prompt.
`,
	})
	_, err := config.Load(root, nil, "")
	if err == nil {
		t.Fatal("expected error for unknown frontmatter key, got nil")
	}
}

// TestLoad_MissingModel verifies a persona with no model fails.
func TestLoad_MissingModel(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": `---
effort: high
---
Alice's system prompt.
`,
	})
	_, err := config.Load(root, nil, "")
	if err == nil {
		t.Fatal("expected error for missing model, got nil")
	}
	if !strings.Contains(err.Error(), "model") {
		t.Errorf("error should mention model: %v", err)
	}
}

// TestLoad_EmptyBody verifies a persona with an empty body fails.
func TestLoad_EmptyBody(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": "---\nmodel: claude-sonnet-4-6\neffort: high\n---\n",
	})
	_, err := config.Load(root, nil, "")
	if err == nil {
		t.Fatal("expected error for empty persona body, got nil")
	}
}

// TestLoad_UninfernableModel verifies a persona whose model backend cannot be inferred fails.
func TestLoad_UninfernableModel(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": `---
model: some-unknown-llm
effort: high
---
Alice's system prompt.
`,
	})
	_, err := config.Load(root, nil, "")
	if err == nil {
		t.Fatal("expected error for uninferrable model, got nil")
	}
	if !strings.Contains(err.Error(), "cannot infer backend") {
		t.Errorf("error should mention backend inference: %v", err)
	}
}

// TestLoad_UnresolvableSelectionName verifies a selector naming a nonexistent persona fails.
func TestLoad_UnresolvableSelectionName(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
		"config.yml":        "table:\n  - alice\n  - carol\n",
	})
	_, err := config.Load(root, nil, "")
	if err == nil {
		t.Fatal("expected error for unresolvable selector, got nil")
	}
	if !strings.Contains(err.Error(), "carol") {
		t.Errorf("error should name the missing persona: %v", err)
	}
}

// TestLoad_SelectorNamesSynthesizerRole verifies that naming a synthesizer-role persona
// in a selector is a fail-fast error (distinct from a nonexistent persona).
func TestLoad_SelectorNamesSynthesizerRole(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":       alicePersona,
		"personas/synthesizer.md": synthPersona,
		"config.yml":              "table:\n  - alice\n  - synthesizer\n",
	})
	_, err := config.Load(root, nil, "")
	if err == nil {
		t.Fatal("expected error when selector names a synthesizer-role persona, got nil")
	}
	// Error must mention synthesizer role, not just "not found"
	if !strings.Contains(strings.ToLower(err.Error()), "synthesizer") {
		t.Errorf("error should mention synthesizer: %v", err)
	}
}

// TestLoad_WithListSynthesizerRole verifies that withList naming a synthesizer-role persona fails.
func TestLoad_WithListSynthesizerRole(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":       alicePersona,
		"personas/synthesizer.md": synthPersona,
	})
	_, err := config.Load(root, []string{"alice", "synthesizer"}, "")
	if err == nil {
		t.Fatal("expected error when withList names a synthesizer-role persona, got nil")
	}
}

// TestLoad_MissingDebateDir verifies that missing .heurema/debate returns a clear error.
func TestLoad_MissingDebateDir(t *testing.T) {
	tmp := t.TempDir()
	_, err := config.Load(tmp, nil, "")
	if err == nil {
		t.Fatal("expected error for missing .heurema/debate, got nil")
	}
	if !strings.Contains(err.Error(), ".heurema/debate") {
		t.Errorf("error should mention .heurema/debate: %v", err)
	}
}
