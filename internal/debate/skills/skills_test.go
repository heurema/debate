package skills_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/heurema/debate/internal/debate/capability"
	"github.com/heurema/debate/internal/debate/skills"
)

func lookup(found ...string) capability.LookPath {
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

func noneFound() capability.LookPath { return lookup() }

var bundledV1 = fstest.MapFS{
	"SKILL.md":                    &fstest.MapFile{Data: []byte("---\nname: debate\ndescription: use for debate\n---\nbody v1\n")},
	"references/cli-reference.md": &fstest.MapFile{Data: []byte("cli v1")},
}

var bundledV2 = fstest.MapFS{
	"SKILL.md":                    &fstest.MapFile{Data: []byte("---\nname: debate\ndescription: use for debate\n---\nbody v2\n")},
	"references/cli-reference.md": &fstest.MapFile{Data: []byte("cli v2")},
}

func baseOpts(home string, lookup capability.LookPath) skills.Options {
	return skills.Options{
		Home:          home,
		LookPath:      lookup,
		Bundled:       bundledV1,
		BinaryVersion: "test",
	}
}

func TestInstallOrRepair_NoClientDetected(t *testing.T) {
	home := t.TempDir()
	results := skills.InstallOrRepair(baseOpts(home, noneFound()))
	if len(results) != 1 || results[0].Action != "skipped" || results[0].Warning == "" {
		t.Fatalf("results = %+v, want one skipped result with a warning", results)
	}
	for _, sub := range []string{".agents", ".claude"} {
		if _, err := os.Stat(filepath.Join(home, sub)); !os.IsNotExist(err) {
			t.Errorf("%s should not be created when no client is detected", sub)
		}
	}
}

func TestInstallOrRepair_CodexOrGeminiYieldsStandardTargetOnly(t *testing.T) {
	for _, client := range []string{"codex", "gemini"} {
		t.Run(client, func(t *testing.T) {
			home := t.TempDir()
			results := skills.InstallOrRepair(baseOpts(home, lookup(client)))
			if len(results) != 1 {
				t.Fatalf("results = %+v, want exactly one target", results)
			}
			want := filepath.Join(home, ".agents", "skills", "debate")
			if results[0].Path != want || results[0].Action != "created" {
				t.Errorf("result = %+v, want created at %s", results[0], want)
			}
			assertNoDir(t, filepath.Join(home, ".claude"))
			assertNoDir(t, filepath.Join(home, ".codex"))
			assertNoDir(t, filepath.Join(home, ".gemini"))
		})
	}
}

func TestInstallOrRepair_ClaudeYieldsClaudeTargetOnly(t *testing.T) {
	home := t.TempDir()
	results := skills.InstallOrRepair(baseOpts(home, lookup("claude")))
	if len(results) != 1 {
		t.Fatalf("results = %+v, want exactly one target", results)
	}
	want := filepath.Join(home, ".claude", "skills", "debate")
	if results[0].Path != want || results[0].Action != "created" {
		t.Errorf("result = %+v, want created at %s", results[0], want)
	}
	assertNoDir(t, filepath.Join(home, ".agents"))
}

func TestInstallOrRepair_CodexAndClaudeYieldsBothTargets(t *testing.T) {
	home := t.TempDir()
	results := skills.InstallOrRepair(baseOpts(home, lookup("codex", "claude")))
	if len(results) != 2 {
		t.Fatalf("results = %+v, want two targets", results)
	}
	wantAgents := filepath.Join(home, ".agents", "skills", "debate")
	wantClaude := filepath.Join(home, ".claude", "skills", "debate")
	if results[0].Path != wantAgents || results[1].Path != wantClaude {
		t.Errorf("results = %+v, want [%s, %s]", results, wantAgents, wantClaude)
	}
}

func TestInstallOrRepair_CodexAndGeminiYieldsOnlyOneStandardTarget(t *testing.T) {
	home := t.TempDir()
	results := skills.InstallOrRepair(baseOpts(home, lookup("codex", "gemini")))
	if len(results) != 1 {
		t.Fatalf("results = %+v, want exactly one target (deduplicated standard)", results)
	}
}

func TestInstallOrRepair_MissingHome(t *testing.T) {
	results := skills.InstallOrRepair(baseOpts("", lookup("claude")))
	if len(results) != 1 || results[0].Action != "skipped" || results[0].Warning == "" {
		t.Fatalf("results = %+v, want one skipped result with a warning", results)
	}
}

func TestInstallOrRepair_InstalledSkillNameAndFrontmatterAreValid(t *testing.T) {
	home := t.TempDir()
	results := skills.InstallOrRepair(baseOpts(home, lookup("claude")))
	if len(results) != 1 || results[0].Action != "created" {
		t.Fatalf("results = %+v", results)
	}
	if filepath.Base(results[0].Path) != "debate" {
		t.Errorf("installed skill directory name = %q, want debate", filepath.Base(results[0].Path))
	}
	data, err := os.ReadFile(filepath.Join(results[0].Path, "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile SKILL.md: %v", err)
	}
	if !strings.Contains(string(data), "name: debate") {
		t.Errorf("installed SKILL.md missing name: debate frontmatter: %q", data)
	}
}

func TestInstallOrRepair_Idempotent(t *testing.T) {
	home := t.TempDir()
	opts := baseOpts(home, lookup("claude"))
	first := skills.InstallOrRepair(opts)
	if first[0].Action != "created" {
		t.Fatalf("first install = %+v, want created", first[0])
	}
	second := skills.InstallOrRepair(opts)
	if second[0].Action != "current" {
		t.Fatalf("second install = %+v, want current", second[0])
	}
}

func TestInstallOrRepair_UpdatesWhenBundledContentChanges(t *testing.T) {
	home := t.TempDir()
	opts := baseOpts(home, lookup("claude"))
	first := skills.InstallOrRepair(opts)
	if first[0].Action != "created" {
		t.Fatalf("first install = %+v, want created", first[0])
	}

	opts.Bundled = bundledV2
	second := skills.InstallOrRepair(opts)
	if second[0].Action != "updated" {
		t.Fatalf("second install = %+v, want updated", second[0])
	}

	data, err := os.ReadFile(filepath.Join(second[0].Path, "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile SKILL.md: %v", err)
	}
	if !strings.Contains(string(data), "body v2") {
		t.Errorf("SKILL.md not updated to bundled v2 content: %q", data)
	}

	// Idempotent again at the new content.
	third := skills.InstallOrRepair(opts)
	if third[0].Action != "current" {
		t.Fatalf("third install = %+v, want current", third[0])
	}
}

func TestInstallOrRepair_PreservesLocallyModifiedTarget(t *testing.T) {
	home := t.TempDir()
	opts := baseOpts(home, lookup("claude"))
	first := skills.InstallOrRepair(opts)
	target := first[0].Path

	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("locally edited"), 0644); err != nil {
		t.Fatal(err)
	}

	opts.Bundled = bundledV2
	second := skills.InstallOrRepair(opts)
	if second[0].Action != "skipped" || second[0].Warning == "" {
		t.Fatalf("result = %+v, want skipped with a warning for local edits", second[0])
	}
	data, err := os.ReadFile(filepath.Join(target, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "locally edited" {
		t.Errorf("locally edited content was overwritten: %q", data)
	}
}

func TestInstallOrRepair_PreservesUnmanagedTarget(t *testing.T) {
	home := t.TempDir()
	target := filepath.Join(home, ".claude", "skills", "debate")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("hand-authored"), 0644); err != nil {
		t.Fatal(err)
	}

	results := skills.InstallOrRepair(baseOpts(home, lookup("claude")))
	if results[0].Action != "skipped" || results[0].Warning == "" {
		t.Fatalf("result = %+v, want skipped with a warning for unmanaged content", results[0])
	}
	data, err := os.ReadFile(filepath.Join(target, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hand-authored" {
		t.Errorf("unmanaged content was overwritten: %q", data)
	}
}

func TestInstallOrRepair_SkipsSymlinkedTarget(t *testing.T) {
	home := t.TempDir()
	skillsDir := filepath.Join(home, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}
	elsewhere := t.TempDir()
	if err := os.Symlink(elsewhere, filepath.Join(skillsDir, "debate")); err != nil {
		t.Skipf("symlink not supported in this environment: %v", err)
	}

	results := skills.InstallOrRepair(baseOpts(home, lookup("claude")))
	if results[0].Action != "skipped" || results[0].Warning == "" {
		t.Fatalf("result = %+v, want skipped with a warning for symlinked target", results[0])
	}
	if entries, _ := os.ReadDir(elsewhere); len(entries) != 0 {
		t.Errorf("symlink target directory was written through: %v", entries)
	}
}

func TestInstallOrRepair_NoCodexOrGeminiSkillsDirsCreated(t *testing.T) {
	home := t.TempDir()
	skills.InstallOrRepair(baseOpts(home, lookup("codex", "gemini", "claude")))
	for _, sub := range []string{".codex", ".gemini"} {
		assertNoDir(t, filepath.Join(home, sub))
	}
}

func TestInstallOrRepair_ExistingAgentsDirDetectsStandardTarget(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}
	results := skills.InstallOrRepair(baseOpts(home, noneFound()))
	if len(results) != 1 || results[0].Action != "created" {
		t.Fatalf("results = %+v, want one created result from existing ~/.agents", results)
	}
}

func TestInstallOrRepair_ExistingClaudeDirDetectsClaudeTarget(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}
	results := skills.InstallOrRepair(baseOpts(home, noneFound()))
	if len(results) != 1 || results[0].Path != filepath.Join(home, ".claude", "skills", "debate") {
		t.Fatalf("results = %+v, want one created result at claude target", results)
	}
}

func TestInstallOrRepair_MetadataRecordsManagedChecksum(t *testing.T) {
	home := t.TempDir()
	results := skills.InstallOrRepair(baseOpts(home, lookup("claude")))
	data, err := os.ReadFile(filepath.Join(results[0].Path, skills.MetadataFileName))
	if err != nil {
		t.Fatalf("ReadFile metadata: %v", err)
	}
	var meta skills.Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("Unmarshal metadata: %v", err)
	}
	if meta.Checksum == "" {
		t.Error("metadata checksum is empty")
	}
	if meta.Schema == "" {
		t.Error("metadata schema is empty")
	}
}

func assertNoDir(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("%s should not exist", path)
	}
}
