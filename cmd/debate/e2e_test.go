package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heurema/debate/internal/debate/progress"
	"github.com/heurema/debate/internal/debate/runner"
	"github.com/heurema/debate/internal/engine/transport"
	"github.com/heurema/debate/internal/engine/transport/echo"
	"github.com/heurema/debate/internal/engine/transport/mock"
)

// echoAll returns the echo transport regardless of backend name.
func echoAll(_ string) (transport.Transport, error) {
	return echo.New(), nil
}

// makeE2EWorkspace creates a fixture workspace and returns the root dir.
func makeE2EWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	debDir := filepath.Join(root, ".heurema", "debate")
	personasDir := filepath.Join(debDir, "personas")
	if err := os.MkdirAll(personasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"alice", "bob"} {
		path := filepath.Join(personasDir, name+".md")
		content := "---\nversion: 1\nmodel: echo-local\neffort: low\nbackend: echo\n---\n" +
			"You are a debate participant. State your position clearly.\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tablesDir := filepath.Join(debDir, "tables")
	if err := os.MkdirAll(tablesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	table := "version: 1\npanel:\n  - alice\n  - bob\n"
	if err := os.WriteFile(filepath.Join(tablesDir, "default.yml"), []byte(table), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// makeUnimplementedWorkspace creates a workspace whose personas use an unknown backend.
func makeUnimplementedWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	debDir := filepath.Join(root, ".heurema", "debate")
	personasDir := filepath.Join(debDir, "personas")
	if err := os.MkdirAll(personasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// backend: api is not implemented and is unknown to defaultResolver.
	path := filepath.Join(personasDir, "agent.md")
	content := "---\nversion: 1\nmodel: api-model-1\neffort: low\nbackend: api\n---\nYou are a debate participant.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	tablesDir := filepath.Join(debDir, "tables")
	if err := os.MkdirAll(tablesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	table := "version: 1\npanel:\n  - agent\n"
	if err := os.WriteFile(filepath.Join(tablesDir, "default.yml"), []byte(table), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// (a) settled run: stdout has the answer, stderr has default progress events, exit 0.
func TestE2E_SettledRun_DefaultProgress(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseAndRun(
		[]string{"should we use tabs or spaces?"},
		&stdout, &stderr, strings.NewReader(""),
		echoAll,
		workDir,
	)

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Error("stdout is empty — expected the synthesized answer")
	}
	if strings.Contains(stdout.String(), "@@DEBATE_PROGRESS ") {
		t.Fatalf("stdout contains progress prefix: %s", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Error("stderr is empty — expected default progress events")
	}
	events := parseOnlyProgressLines(t, stderr.String())
	if events[0].Type != "run_started" || events[len(events)-1].Type != "run_completed" {
		t.Fatalf("progress endpoints = %s/%s, want run_started/run_completed", events[0].Type, events[len(events)-1].Type)
	}
	if strings.Contains(stderr.String(), "[Round") {
		t.Fatalf("legacy human turn trace was emitted: %s", stderr.String())
	}
}

// (b) non-converged run returns exit 2.
func TestE2E_NotConverged_Exit2(t *testing.T) {
	workDir := makeE2EWorkspace(t)

	const notDoneContent = "I disagree.\n\n```signal\n" +
		"{\"position\": \"disagree\", \"objections\": [\"needs work\"], \"done\": false}\n```"

	resolver := func(_ string) (transport.Transport, error) {
		return mock.NewTransport(map[string]*mock.Session{
			"alice":       mock.NewSession(notDoneScripts(5, notDoneContent)),
			"bob":         mock.NewSession(notDoneScripts(5, notDoneContent)),
			"synthesizer": mock.NewSession([]mock.ScriptedResult{{Result: transport.Result{Content: "synthesis"}}}),
		}), nil
	}

	var stdout, stderr bytes.Buffer
	code := parseAndRun(
		[]string{"--max-rounds", "1", "controversial topic"},
		&stdout, &stderr, strings.NewReader(""),
		resolver, workDir,
	)

	if code != 2 {
		t.Errorf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestE2E_MaxRoundsFlagBeforeAndAfterTask(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	const notDoneContent = "I disagree.\n\n```signal\n" +
		"{\"position\": \"disagree\", \"objections\": [\"needs work\"], \"done\": false}\n```"

	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "before", args: []string{"--max-rounds", "1", "controversial topic"}},
		{name: "after", args: []string{"controversial topic", "--max-rounds", "1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolver := func(_ string) (transport.Transport, error) {
				return mock.NewTransport(map[string]*mock.Session{
					"alice":       mock.NewSession(notDoneScripts(5, notDoneContent)),
					"bob":         mock.NewSession(notDoneScripts(5, notDoneContent)),
					"synthesizer": mock.NewSession([]mock.ScriptedResult{{Result: transport.Result{Content: "synthesis"}}}),
				}), nil
			}

			var stdout, stderr bytes.Buffer
			code := parseCLI(
				tc.args,
				&stdout, &stderr, strings.NewReader(""),
				resolver, workDir,
			)
			if code != 2 {
				t.Errorf("exit code = %d, want 2; stderr: %s", code, stderr.String())
			}
			events := parseOnlyProgressLines(t, stderr.String())
			if turns := countProgressType(events, "turn_completed"); turns != 2 {
				t.Errorf("turn_completed events = %d, want 2; stderr: %s", turns, stderr.String())
			}
		})
	}
}

func TestE2E_RunWordIsTaskNotCommand(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	const notDoneContent = "I disagree.\n\n```signal\n" +
		"{\"position\": \"disagree\", \"objections\": [\"needs work\"], \"done\": false}\n```"
	synth := mock.NewSession([]mock.ScriptedResult{{Result: transport.Result{Content: "synthesis"}}})
	resolver := func(_ string) (transport.Transport, error) {
		return mock.NewTransport(map[string]*mock.Session{
			"alice":       mock.NewSession(notDoneScripts(1, notDoneContent)),
			"bob":         mock.NewSession(notDoneScripts(1, notDoneContent)),
			"synthesizer": synth,
		}), nil
	}

	var stdout, stderr bytes.Buffer
	code := parseCLI(
		[]string{"run", "--max-rounds", "1"},
		&stdout, &stderr, strings.NewReader(""),
		resolver, workDir,
	)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	prompts := synth.Prompts()
	if len(prompts) != 1 {
		t.Fatalf("synthesizer prompts = %d, want 1", len(prompts))
	}
	if !strings.Contains(prompts[0], "Task: run\n\n") {
		t.Fatalf("synthesizer prompt does not contain task %q: %s", "run", prompts[0])
	}
}

func TestE2E_TableFlagSelectsPanel(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	tablePath := filepath.Join(workDir, ".heurema", "debate", "tables", "solo.yml")
	if err := os.WriteFile(tablePath, []byte("version: 1\npanel:\n  - bob\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	const content = "I agree.\n\n```signal\n{\"position\": \"agree\", \"objections\": [], \"done\": true}\n```"
	bob := mock.NewSession(notDoneScripts(5, content))
	synth := mock.NewSession([]mock.ScriptedResult{{Result: transport.Result{Content: "synthesis"}}})
	resolver := func(_ string) (transport.Transport, error) {
		return mock.NewTransport(map[string]*mock.Session{
			"bob":         bob,
			"synthesizer": synth,
		}), nil
	}

	var stdout, stderr bytes.Buffer
	code := parseAndRun(
		[]string{"--table", "solo", "use solo table"},
		&stdout, &stderr, strings.NewReader(""),
		resolver, workDir,
	)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if len(bob.Prompts()) == 0 {
		t.Fatal("bob was not run")
	}
}

func TestCLI_WithSelectorsSupportRepeatableAndCommaSeparatedForms(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		order []string
	}{
		{
			name:  "repeatable",
			args:  []string{"--with", "proposer", "--with", "skeptic"},
			order: []string{"proposer", "skeptic"},
		},
		{
			name:  "comma separated",
			args:  []string{"--with", "proposer,skeptic"},
			order: []string{"proposer", "skeptic"},
		},
		{
			name:  "equals comma separated",
			args:  []string{"--with=proposer,skeptic"},
			order: []string{"proposer", "skeptic"},
		},
		{
			name:  "mixed repeatable and comma separated",
			args:  []string{"--with", "proposer,skeptic", "--with", "reviewers/security"},
			order: []string{"proposer", "skeptic", "reviewers/security"},
		},
		{
			name:  "whitespace around comma tokens",
			args:  []string{"--with", "proposer, skeptic"},
			order: []string{"proposer", "skeptic"},
		},
		{
			name:  "whitespace around single token",
			args:  []string{"--with", " proposer "},
			order: []string{"proposer"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			workDir := makeSelectorWorkspace(t)
			args := append([]string{"--json", "--quiet"}, tc.args...)
			args = append(args, "task")

			var stdout, stderr bytes.Buffer
			code := parseAndRun(
				args,
				&stdout, &stderr, strings.NewReader(""),
				echoAll, workDir,
			)

			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			got := speakerOrderFromJSON(t, stdout.Bytes())
			want := append(append([]string{}, tc.order...), tc.order...)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("speaker order = %v, want %v", got, want)
			}
		})
	}
}

func TestCLI_UnknownFlagReportsUsage(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseCLI(
		[]string{"--definitely-not-a-flag", "task"},
		&stdout, &stderr, strings.NewReader(""),
		echoAll, workDir,
	)

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	errText := stderr.String()
	if !strings.Contains(errText, "--definitely-not-a-flag") {
		t.Fatalf("stderr missing unknown flag: %q", errText)
	}
	if !strings.Contains(errText, "Usage:") {
		t.Fatalf("stderr missing usage text: %q", errText)
	}
}

// (c) non-TTY environment still gets default progress on stderr.
func TestE2E_NonTTY_DefaultProgress(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseAndRun(
		[]string{"should we do X?"},
		&stdout, &stderr, strings.NewReader(""),
		echoAll, workDir,
	)

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if len(parseOnlyProgressLines(t, stderr.String())) == 0 {
		t.Fatal("expected progress events")
	}
}

// (d) empty task fails fast with exit 1.
func TestE2E_EmptyTask_Exit1(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseAndRun(
		[]string{},
		&stdout, &stderr, strings.NewReader(""),
		echoAll, workDir,
	)

	if code != 1 {
		t.Errorf("exit code = %d, want 1; stderr: %s", code, stderr.String())
	}
	// No runner activity: stdout should be empty.
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty for empty-task error; got: %s", stdout.String())
	}
}

// (e) unknown backend fails fast with exit 1.
func TestE2E_UnimplementedBackend_Exit1(t *testing.T) {
	workDir := makeUnimplementedWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseAndRun(
		[]string{"any task"},
		&stdout, &stderr, strings.NewReader(""),
		defaultResolver, workDir,
	)

	if code != 1 {
		t.Errorf("exit code = %d, want 1; stderr: %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty on backend error; got: %s", stdout.String())
	}
}

// TestDefaultResolver_ACPBackendsResolve asserts that claude-agent-acp, codex-acp, and agy
// resolve to a transport.Transport in the default resolver without opening a real session.
func TestDefaultResolver_ACPBackendsResolve(t *testing.T) {
	for _, backend := range []string{"claude-agent-acp", "codex-acp", "agy"} {
		tr, err := defaultResolver(backend)
		if err != nil {
			t.Errorf("defaultResolver(%q): %v", backend, err)
			continue
		}
		if tr == nil {
			t.Errorf("defaultResolver(%q): returned nil transport", backend)
		}
	}
}

// Additional: --json output has the required structure.
func TestE2E_JSONOutput(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseAndRun(
		[]string{"--json", "a task"},
		&stdout, &stderr, strings.NewReader(""),
		echoAll, workDir,
	)

	if code != 0 {
		t.Fatalf("exit code = %d; stderr: %s", code, stderr.String())
	}
	parseOnlyProgressLines(t, stderr.String())
	if strings.Contains(stdout.String(), "@@DEBATE_PROGRESS ") {
		t.Fatalf("stdout contains progress prefix: %s", stdout.String())
	}

	var out map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON on stdout: %v\nraw: %s", err, stdout.String())
	}
	for _, key := range []string{"answer", "outcome", "rounds", "turns"} {
		if _, ok := out[key]; !ok {
			t.Errorf("JSON missing key %q", key)
		}
	}
	// No extra keys.
	if len(out) != 4 {
		t.Errorf("JSON has %d top-level keys, want 4; keys: %v", len(out), keysOf(out))
	}
}

func TestE2E_JSONFlagBeforeAndAfterTask(t *testing.T) {
	workDir := makeE2EWorkspace(t)

	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "before", args: []string{"--json", "a task"}},
		{name: "after", args: []string{"a task", "--json"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := parseCLI(
				tc.args,
				&stdout, &stderr, strings.NewReader(""),
				echoAll, workDir,
			)

			if code != 0 {
				t.Fatalf("exit code = %d; stderr: %s", code, stderr.String())
			}
			var out map[string]json.RawMessage
			if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
				t.Fatalf("invalid JSON on stdout: %v\nraw: %s", err, stdout.String())
			}
			parseOnlyProgressLines(t, stderr.String())
		})
	}
}

func TestE2E_QuietSuppressesProgress(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "human", args: []string{"--quiet", "a task"}},
		{name: "json", args: []string{"--json", "--quiet", "a task"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := parseAndRun(
				tc.args,
				&stdout, &stderr, strings.NewReader(""),
				echoAll, workDir,
			)

			if code != 0 {
				t.Fatalf("exit code = %d; stderr: %s", code, stderr.String())
			}
			if stdout.Len() == 0 {
				t.Fatal("stdout is empty")
			}
			if strings.Contains(stderr.String(), progress.Prefix) {
				t.Fatalf("--quiet stderr contains progress events: %s", stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("--quiet stderr = %q, want empty for successful echo run", stderr.String())
			}
		})
	}
}

func TestE2E_ErrorIncludesRunFailedBeforeErrorLine(t *testing.T) {
	workDir := makeUnimplementedWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseAndRun(
		[]string{"any task"},
		&stdout, &stderr, strings.NewReader(""),
		defaultResolver, workDir,
	)

	if code != 1 {
		t.Fatalf("exit code = %d, want 1; stderr: %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	lines := nonEmptyLines(stderr.String())
	if len(lines) < 2 {
		t.Fatalf("stderr lines = %v, want progress and error line", lines)
	}
	events := parseProgressLines(t, stderr.String())
	lastProgress := events[len(events)-1]
	if lastProgress.Type != "run_failed" {
		t.Fatalf("last progress event = %q, want run_failed", lastProgress.Type)
	}
	if !strings.HasPrefix(lines[len(lines)-1], "error:") {
		t.Fatalf("final stderr line = %q, want existing error line", lines[len(lines)-1])
	}
	if strings.HasPrefix(lines[len(lines)-1], progress.Prefix) {
		t.Fatalf("error line unexpectedly uses progress prefix: %q", lines[len(lines)-1])
	}
}

func keysOf(m map[string]json.RawMessage) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func makeSelectorWorkspace(t *testing.T) string {
	t.Helper()
	workDir := makeE2EWorkspace(t)
	for _, id := range []string{"proposer", "skeptic", "reviewers/security"} {
		path := filepath.Join(workDir, ".heurema", "debate", "personas", id+".md")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		content := "---\nversion: 1\nmodel: echo-local\neffort: low\nbackend: echo\n---\n" +
			"You are " + id + ". State your position clearly.\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return workDir
}

func speakerOrderFromJSON(t *testing.T, data []byte) []string {
	t.Helper()
	var out struct {
		Turns []struct {
			Speaker string `json:"speaker"`
		} `json:"turns"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("invalid JSON on stdout: %v\nraw: %s", err, string(data))
	}
	speakers := make([]string, len(out.Turns))
	for i, turn := range out.Turns {
		speakers[i] = turn.Speaker
	}
	return speakers
}

type cliProgressEvent struct {
	Version    int    `json:"version"`
	Type       string `json:"type"`
	Stage      string `json:"stage"`
	ElapsedMS  int64  `json:"elapsed_ms"`
	DurationMS *int64 `json:"duration_ms,omitempty"`
	SilenceMS  *int64 `json:"silence_ms,omitempty"`
	Round      *int   `json:"round,omitempty"`
	Speaker    string `json:"speaker,omitempty"`
	Error      string `json:"error,omitempty"`
}

func parseOnlyProgressLines(t *testing.T, text string) []cliProgressEvent {
	t.Helper()
	lines := nonEmptyLines(text)
	for _, line := range lines {
		if !strings.HasPrefix(line, progress.Prefix) {
			t.Fatalf("stderr line missing progress prefix %q: %q", progress.Prefix, line)
		}
	}
	return parseProgressLines(t, text)
}

func parseProgressLines(t *testing.T, text string) []cliProgressEvent {
	t.Helper()
	var events []cliProgressEvent
	for _, line := range nonEmptyLines(text) {
		if !strings.HasPrefix(line, progress.Prefix) {
			continue
		}
		var ev cliProgressEvent
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, progress.Prefix)), &ev); err != nil {
			t.Fatalf("invalid progress JSON %q: %v", line, err)
		}
		if ev.Version != 1 {
			t.Fatalf("progress version = %d, want 1 in %q", ev.Version, line)
		}
		if ev.Type == "" || ev.Stage == "" {
			t.Fatalf("progress event missing type/stage: %+v", ev)
		}
		if ev.ElapsedMS < 0 {
			t.Fatalf("progress elapsed_ms is negative: %+v", ev)
		}
		events = append(events, ev)
	}
	if len(events) == 0 {
		t.Fatal("no progress events parsed")
	}
	return events
}

func nonEmptyLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func countProgressType(events []cliProgressEvent, typ string) int {
	var count int
	for _, ev := range events {
		if ev.Type == typ {
			count++
		}
	}
	return count
}

// Additional: --task @file reads the file as task.
func TestE2E_TaskFromFile(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	taskFile := filepath.Join(t.TempDir(), "task.txt")
	if err := os.WriteFile(taskFile, []byte("file task content"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := parseAndRun(
		[]string{"--task", "@" + taskFile},
		&stdout, &stderr, strings.NewReader(""),
		echoAll, workDir,
	)

	if code != 0 {
		t.Errorf("exit code = %d; stderr: %s", code, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Error("expected non-empty stdout")
	}
}

func notDoneScripts(n int, content string) []mock.ScriptedResult {
	out := make([]mock.ScriptedResult, n)
	for i := range out {
		out[i] = mock.ScriptedResult{Result: transport.Result{Content: content}}
	}
	return out
}

// Keep the existing TestVersion passing.
var _ = runner.Config{} // ensure import is used
