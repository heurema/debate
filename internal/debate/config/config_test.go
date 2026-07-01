package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heurema/debate/internal/debate/capability"
	"github.com/heurema/debate/internal/debate/config"
	"github.com/heurema/debate/internal/debate/persona"
)

// fakeLookPath simulates which executables are on PATH without touching the
// real environment. Used to make capability.Detect deterministic in tests.
func fakeLookPath(found ...string) capability.LookPath {
	set := make(map[string]bool, len(found))
	for _, f := range found {
		set[f] = true
	}
	return func(name string) (string, error) {
		if set[name] {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found")
	}
}

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
version: 1
model: claude-sonnet-4-6
effort: high
---
You are Alice, a careful logical reasoner.
`

const bobPersona = `---
version: 1
model: claude-opus-4-8
effort: medium
---
You are Bob, a pragmatic problem solver.
`

const synthPersona = `---
version: 1
role: synthesizer
model: claude-haiku-4-5
effort: low
---
You are the synthesizer. Summarize the discussion.
`

const defaultTable = `version: 1
panel:
  - alice
  - bob
`

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

func TestDiscover_MissingDir(t *testing.T) {
	tmp := t.TempDir()
	_, err := config.Discover(tmp)
	if err == nil {
		t.Fatal("expected error when .heurema/debate absent, got nil")
	}
}

func TestLoad_DefaultTablePreservesPanelOrder(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":       alicePersona,
		"personas/bob.md":         bobPersona,
		"personas/synthesizer.md": synthPersona,
		"tables/default.yml":      defaultTable,
	})
	ws, err := config.Load(root, "", nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ws.Panel) != 2 {
		t.Fatalf("Panel len = %d, want 2", len(ws.Panel))
	}
	if ws.Panel[0].ID != "alice" || ws.Panel[1].ID != "bob" {
		t.Errorf("Panel IDs = [%s, %s], want [alice, bob]", ws.Panel[0].ID, ws.Panel[1].ID)
	}
	if ws.Synthesizer.ID != "synthesizer" {
		t.Errorf("Synthesizer ID = %q, want synthesizer", ws.Synthesizer.ID)
	}
}

func TestLoad_SelectedTableAndSynth(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":     alicePersona,
		"personas/bob.md":       bobPersona,
		"personas/custom.md":    synthPersona,
		"tables/default.yml":    defaultTable,
		"tables/alternate.yml":  "version: 1\npanel:\n  - bob\n  - alice\nsynth: custom\n",
		"tables/not-a-table.md": "ignored\n",
		"tables/.hidden.yml":    "not parsed\n",
	})
	ws, err := config.Load(root, "alternate", nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Panel[0].ID != "bob" || ws.Panel[1].ID != "alice" {
		t.Errorf("Panel IDs = [%s, %s], want [bob, alice]", ws.Panel[0].ID, ws.Panel[1].ID)
	}
	if ws.Synthesizer.ID != "custom" {
		t.Errorf("Synthesizer ID = %q, want custom", ws.Synthesizer.ID)
	}
}

func TestLoad_WithListOverridesOnlyPanel(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":    alicePersona,
		"personas/bob.md":      bobPersona,
		"personas/custom.md":   synthPersona,
		"tables/default.yml":   defaultTable,
		"tables/selected.yml":  "version: 1\npanel:\n  - alice\nsynth: custom\n",
		"personas/notes.txt":   "ignored\n",
		"personas/.hidden.md":  "ignored\n",
		"personas/team/readme": "ignored\n",
	})
	ws, err := config.Load(root, "selected", []string{"bob", "alice"}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Panel[0].ID != "bob" || ws.Panel[1].ID != "alice" {
		t.Errorf("Panel IDs = [%s, %s], want [bob, alice]", ws.Panel[0].ID, ws.Panel[1].ID)
	}
	if ws.Synthesizer.ID != "custom" {
		t.Errorf("Synthesizer ID = %q, want custom", ws.Synthesizer.ID)
	}
}

func TestLoad_WithListCanRunWithoutTables(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
	})
	ws, err := config.Load(root, "", []string{"alice"}, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Panel[0].ID != "alice" {
		t.Errorf("Panel ID = %q, want alice", ws.Panel[0].ID)
	}
	if ws.Synthesizer.ID != "synthesizer" {
		t.Errorf("Synthesizer ID = %q, want built-in synthesizer", ws.Synthesizer.ID)
	}
}

func TestLoad_NamespacedPersonasUseFullIDs(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/team/alice.md": alicePersona,
		"personas/bob.md":        bobPersona,
		"tables/default.yml":     "version: 1\npanel:\n  - team/alice\n  - bob\n",
	})
	ws, err := config.Load(root, "", nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Panel[0].ID != "team/alice" || ws.Panel[1].ID != "bob" {
		t.Errorf("Panel IDs = [%s, %s], want [team/alice, bob]", ws.Panel[0].ID, ws.Panel[1].ID)
	}
}

func TestLoad_ShortSelectorPrefersExactRootPersona(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":      alicePersona,
		"personas/red/alice.md":  bobPersona,
		"personas/blue/alice.md": bobPersona,
		"personas/bob.md":        bobPersona,
		"tables/default.yml":     "version: 1\npanel:\n  - alice\n  - red/alice\n  - blue/alice\n  - bob\n",
	})
	ws, err := config.Load(root, "", nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := []string{"alice", "red/alice", "blue/alice", "bob"}
	if got := personaIDs(ws.Panel); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("Panel IDs = %v, want %v", got, want)
	}
}

func TestLoad_ShortSelectorResolvesSingleNamespacedCandidate(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/team/alice.md": alicePersona,
		"tables/default.yml":     "version: 1\npanel:\n  - alice\n",
	})
	ws, err := config.Load(root, "", nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Panel[0].ID != "team/alice" {
		t.Fatalf("Panel ID = %q, want team/alice", ws.Panel[0].ID)
	}
}

func TestLoad_ShortSelectorAmbiguityListsQualifiedIDs(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/red/alice.md":  alicePersona,
		"personas/blue/alice.md": bobPersona,
		"tables/default.yml":     "version: 1\npanel:\n  - alice\n",
	})
	_, err := config.Load(root, "", nil, "")
	if err == nil {
		t.Fatal("expected ambiguous selector error, got nil")
	}
	want := `panel: selector "alice" is ambiguous; use a qualified persona ID: blue/alice, red/alice`
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestLoad_PanelRejectsDuplicateResolvedPersonas(t *testing.T) {
	for _, tc := range []struct {
		name     string
		files    map[string]string
		withList []string
	}{
		{
			name: "table panel",
			files: map[string]string{
				"personas/alice.md":  alicePersona,
				"tables/default.yml": "version: 1\npanel:\n  - alice\n  - alice\n",
			},
		},
		{
			name: "table panel equivalent selectors",
			files: map[string]string{
				"personas/team/alice.md": alicePersona,
				"tables/default.yml":     "version: 1\npanel:\n  - alice\n  - team/alice\n",
			},
		},
		{
			name: "with override equivalent selectors",
			files: map[string]string{
				"personas/alice.md": alicePersona,
			},
			withList: []string{"alice", "alice"},
		},
		{
			name: "with override root short and qualified duplicate",
			files: map[string]string{
				"personas/team/alice.md": alicePersona,
			},
			withList: []string{"alice", "team/alice"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := makeDebateDir(t, tc.files)
			_, err := config.Load(root, "", tc.withList, "")
			if err == nil {
				t.Fatal("expected duplicate panel error, got nil")
			}
			if !strings.Contains(err.Error(), "duplicate persona") {
				t.Fatalf("error should mention duplicate persona: %v", err)
			}
		})
	}
}

func TestLoad_MissingSelectorsNameSelector(t *testing.T) {
	for _, tc := range []struct {
		name          string
		files         map[string]string
		withList      []string
		synthOverride string
		wantSelector  string
	}{
		{
			name: "table panel",
			files: map[string]string{
				"personas/alice.md":  alicePersona,
				"tables/default.yml": "version: 1\npanel:\n  - missing\n",
			},
			wantSelector: "missing",
		},
		{
			name: "with override",
			files: map[string]string{
				"personas/alice.md": alicePersona,
			},
			withList:     []string{"missing"},
			wantSelector: "missing",
		},
		{
			name: "table synth",
			files: map[string]string{
				"personas/alice.md":  alicePersona,
				"tables/default.yml": "version: 1\npanel:\n  - alice\nsynth: missing\n",
			},
			wantSelector: "missing",
		},
		{
			name: "synth override",
			files: map[string]string{
				"personas/alice.md":  alicePersona,
				"tables/default.yml": "version: 1\npanel:\n  - alice\n",
			},
			synthOverride: "missing",
			wantSelector:  "missing",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := makeDebateDir(t, tc.files)
			_, err := config.Load(root, "", tc.withList, tc.synthOverride)
			if err == nil {
				t.Fatal("expected missing selector error, got nil")
			}
			if !strings.Contains(err.Error(), `selector "`+tc.wantSelector+`" did not match any persona`) {
				t.Fatalf("error should name missing selector %q: %v", tc.wantSelector, err)
			}
		})
	}
}

func TestLoad_DefaultSynthesizerAmbiguityFails(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":               alicePersona,
		"personas/red/synthesizer.md":     synthPersona,
		"personas/blue/synthesizer.md":    synthPersona,
		"tables/default.yml":              "version: 1\npanel:\n  - alice\n",
		"personas/blue/ordinary-note.txt": "ignored\n",
	})
	_, err := config.Load(root, "", nil, "")
	if err == nil {
		t.Fatal("expected ambiguous default synthesizer error, got nil")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("error should mention ambiguity: %v", err)
	}
}

func TestLoad_PanelRejectsSynthesizerRole(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":       alicePersona,
		"personas/synthesizer.md": synthPersona,
		"tables/default.yml":      "version: 1\npanel:\n  - alice\n  - synthesizer\n",
	})
	_, err := config.Load(root, "", nil, "")
	if err == nil {
		t.Fatal("expected error when panel names synthesizer-role persona, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be in the debater panel") {
		t.Errorf("error should mention debater panel role rejection: %v", err)
	}
}

func TestLoad_SynthRejectsDebaterRole(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":  alicePersona,
		"personas/bob.md":    bobPersona,
		"tables/default.yml": "version: 1\npanel:\n  - alice\nsynth: bob\n",
	})
	_, err := config.Load(root, "", nil, "")
	if err == nil {
		t.Fatal("expected error when synth names debater-role persona, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be used as synthesizer") {
		t.Errorf("error should mention synthesizer role rejection: %v", err)
	}
}

func TestLoad_DefaultSynthesizerRejectsRootDebaterPersona(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":       alicePersona,
		"personas/synthesizer.md": alicePersona,
		"tables/default.yml":      "version: 1\npanel:\n  - alice\n",
	})
	_, err := config.Load(root, "", nil, "")
	if err == nil {
		t.Fatal("expected root debater synthesizer error, got nil")
	}
	if !strings.Contains(err.Error(), `persona "synthesizer" has role debater`) {
		t.Fatalf("error should reject root debater synthesizer before built-in fallback: %v", err)
	}
}

func TestLoad_DefaultSynthesizerPrefersRootOverNamespaced(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":            alicePersona,
		"personas/synthesizer.md":      synthPersona,
		"personas/team/synthesizer.md": synthPersona,
		"tables/default.yml":           "version: 1\npanel:\n  - alice\n",
	})
	ws, err := config.Load(root, "", nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Synthesizer.ID != "synthesizer" {
		t.Fatalf("Synthesizer ID = %q, want synthesizer", ws.Synthesizer.ID)
	}
}

func TestLoad_SynthOverrideWinsOverTable(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":    alicePersona,
		"personas/table.md":    synthPersona,
		"personas/override.md": synthPersona,
		"tables/default.yml":   "version: 1\npanel:\n  - alice\nsynth: table\n",
	})
	ws, err := config.Load(root, "", nil, "override")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Synthesizer.ID != "override" {
		t.Errorf("Synthesizer ID = %q, want override", ws.Synthesizer.ID)
	}
}

func TestLoad_MissingTablesWithoutWithFails(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
	})
	_, err := config.Load(root, "", nil, "")
	if err == nil {
		t.Fatal("expected missing tables error, got nil")
	}
	if !strings.Contains(err.Error(), "tables") {
		t.Errorf("error should mention tables: %v", err)
	}
}

func TestLoad_ConfigYMLIsIgnoredAndDoesNotProvideTable(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
		"config.yml":        "table:\n  - alice\n",
	})
	_, err := config.Load(root, "", nil, "")
	if err == nil {
		t.Fatal("expected missing tables error despite config.yml, got nil")
	}
	if !strings.Contains(err.Error(), "tables") {
		t.Errorf("error should mention tables: %v", err)
	}
}

func TestLoad_NoTableFilesWithoutWithFails(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md": alicePersona,
		"tables/readme.md":  "ignored\n",
	})
	_, err := config.Load(root, "", nil, "")
	if err == nil {
		t.Fatal("expected no table files error, got nil")
	}
	if !strings.Contains(err.Error(), "no table files") {
		t.Errorf("error should mention no table files: %v", err)
	}
}

func TestLoad_MissingSelectedTableFails(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":  alicePersona,
		"tables/default.yml": "version: 1\npanel:\n  - alice\n",
	})
	_, err := config.Load(root, "missing", nil, "")
	if err == nil {
		t.Fatal("expected missing table error, got nil")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("error should name missing table: %v", err)
	}
}

func TestLoad_TableValidation(t *testing.T) {
	cases := []struct {
		name  string
		table string
		want  string
	}{
		{name: "missing version", table: "panel:\n  - alice\n", want: "version"},
		{name: "unsupported version", table: "version: 2\npanel:\n  - alice\n", want: "unsupported"},
		{name: "empty panel", table: "version: 1\npanel: []\n", want: "empty"},
		{name: "unknown field", table: "version: 1\npanel:\n  - alice\nextra: true\n", want: "field"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := makeDebateDir(t, map[string]string{
				"personas/alice.md":  alicePersona,
				"tables/default.yml": tc.table,
			})
			_, err := config.Load(root, "", nil, "")
			if err == nil {
				t.Fatal("expected table validation error, got nil")
			}
			if !strings.Contains(strings.ToLower(err.Error()), tc.want) {
				t.Errorf("error %q should contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestLoad_InvalidPersonaPathsFail(t *testing.T) {
	cases := []struct {
		name string
		file string
	}{
		{name: "invalid root segment", file: "personas/alice.bad.md"},
		{name: "invalid namespace", file: "personas/bad.ns/alice.md"},
		{name: "deep markdown", file: "personas/team/deep/alice.md"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := makeDebateDir(t, map[string]string{
				tc.file:              alicePersona,
				"tables/default.yml": "version: 1\npanel:\n  - alice\n",
			})
			_, err := config.Load(root, "", nil, "")
			if err == nil {
				t.Fatal("expected invalid persona path error, got nil")
			}
		})
	}
}

func TestLoad_MissingDebateDir(t *testing.T) {
	tmp := t.TempDir()
	_, err := config.Load(tmp, "", nil, "")
	if err == nil {
		t.Fatal("expected error for missing .heurema/debate, got nil")
	}
	if !strings.Contains(err.Error(), ".heurema/debate") {
		t.Errorf("error should mention .heurema/debate: %v", err)
	}
}

const codexPersona = `---
version: 1
model: codex
effort: medium
---
You are Codex, a pragmatic implementer.
`

const geminiPersona = `---
version: 1
model: gemini-pro
effort: medium
---
You are Gemini, an alternative perspective.
`

func TestLoad_DefaultSynthesizerHomogeneousNonClaudePanel(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":  codexPersona,
		"personas/bob.md":    codexPersona,
		"tables/default.yml": defaultTable,
	})
	// No supported executable on PATH; homogeneous codex panel must still
	// resolve without consulting capability.Detect.
	orig := config.LookPath
	config.LookPath = fakeLookPath()
	defer func() { config.LookPath = orig }()

	ws, err := config.Load(root, "", nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Synthesizer.Model != "codex" || ws.Synthesizer.Backend != "codex-acp" {
		t.Errorf("Synthesizer = %+v, want model codex backend codex-acp", ws.Synthesizer)
	}
}

func TestLoad_DefaultSynthesizerMixedPanelFallsBackToDetection(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":  alicePersona, // claude-agent-acp
		"personas/bob.md":    codexPersona, // codex-acp: mixed panel
		"tables/default.yml": defaultTable,
	})
	orig := config.LookPath
	config.LookPath = fakeLookPath("codex")
	defer func() { config.LookPath = orig }()

	ws, err := config.Load(root, "", nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Synthesizer.Model != "codex" || ws.Synthesizer.Backend != "codex-acp" {
		t.Errorf("Synthesizer = %+v, want detection fallback to codex", ws.Synthesizer)
	}
}

func TestLoad_DefaultSynthesizerNoSupportedBackendFailsActionably(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":  alicePersona,
		"personas/bob.md":    codexPersona, // mixed panel forces detection fallback
		"tables/default.yml": defaultTable,
	})
	orig := config.LookPath
	config.LookPath = fakeLookPath() // nothing detected
	defer func() { config.LookPath = orig }()

	_, err := config.Load(root, "", nil, "")
	if err == nil {
		t.Fatal("expected error when no supported backend is detected and panel is mixed")
	}
	for _, want := range []string{"--synth", "table synth", "synthesizer persona", "backend/model"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing actionable hint %q", err.Error(), want)
		}
	}
}

func TestLoad_DefaultSynthesizerGeminiHomogeneousPanel(t *testing.T) {
	root := makeDebateDir(t, map[string]string{
		"personas/alice.md":  geminiPersona,
		"tables/default.yml": "version: 1\npanel:\n  - alice\n",
	})
	orig := config.LookPath
	config.LookPath = fakeLookPath()
	defer func() { config.LookPath = orig }()

	ws, err := config.Load(root, "", nil, "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if ws.Synthesizer.Model != "gemini-pro" || ws.Synthesizer.Backend != "agy" {
		t.Errorf("Synthesizer = %+v, want model gemini-pro backend agy", ws.Synthesizer)
	}
}

func personaIDs(panel []persona.Persona) []string {
	ids := make([]string, len(panel))
	for i, p := range panel {
		ids[i] = p.ID
	}
	return ids
}
