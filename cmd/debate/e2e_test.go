package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heurema/debate/internal/debate/runner"
	"github.com/heurema/debate/internal/engine/transport"
	"github.com/heurema/debate/internal/engine/transport/echo"
	"github.com/heurema/debate/internal/engine/transport/mock"
)

// echoAll returns the echo transport regardless of backend name.
func echoAll(_ string) (transport.Transport, error) {
	return echo.New(), nil
}

// noEnv is a getEnv stub that returns "" for all keys.
func noEnv(_ string) string { return "" }

// forceTraceEnv enables DEBATE_FORCE_TRACE.
func forceTraceEnv(key string) string {
	if key == "DEBATE_FORCE_TRACE" {
		return "1"
	}
	return ""
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
	ctxPath := filepath.Join(debDir, "context.md")
	if err := os.WriteFile(ctxPath, []byte("Fixture workspace context.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"alice", "bob"} {
		path := filepath.Join(personasDir, name+".md")
		content := "---\nmodel: echo-local\neffort: low\nbackend: echo\n---\n" +
			"You are a debate participant. State your position clearly.\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// makeUnimplementedWorkspace creates a workspace whose personas use a real (unimplemented) backend.
func makeUnimplementedWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	debDir := filepath.Join(root, ".heurema", "debate")
	personasDir := filepath.Join(debDir, "personas")
	if err := os.MkdirAll(personasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// model: claude-sonnet-4-6 infers backend claude-agent-acp, which is not implemented.
	path := filepath.Join(personasDir, "agent.md")
	content := "---\nmodel: claude-sonnet-4-6\neffort: low\n---\nYou are a debate participant.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// (a) settled run with DEBATE_FORCE_TRACE=1: stdout has the answer, stderr has trace, exit 0.
func TestE2E_SettledRun_WithTrace(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseAndRun(
		[]string{"should we use tabs or spaces?"},
		&stdout, &stderr, strings.NewReader(""),
		false,         // not a TTY
		forceTraceEnv, // DEBATE_FORCE_TRACE=1
		echoAll,
		workDir,
	)

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Error("stdout is empty — expected the synthesized answer")
	}
	if stderr.Len() == 0 {
		t.Error("stderr is empty — expected debate trace with DEBATE_FORCE_TRACE=1")
	}
	if !strings.Contains(stderr.String(), "[Round") {
		t.Errorf("trace does not contain [Round ...]; stderr: %s", stderr.String())
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
		false, noEnv, resolver, workDir,
	)

	if code != 2 {
		t.Errorf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

// (c) non-TTY environment without DEBATE_FORCE_TRACE produces empty stderr.
func TestE2E_NonTTY_NoTrace(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseAndRun(
		[]string{"should we do X?"},
		&stdout, &stderr, strings.NewReader(""),
		false, noEnv, echoAll, workDir,
	)

	if code != 0 {
		t.Errorf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr must be empty in non-TTY path without DEBATE_FORCE_TRACE; got: %s", stderr.String())
	}
}

// (d) empty task fails fast with exit 1.
func TestE2E_EmptyTask_Exit1(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseAndRun(
		[]string{},
		&stdout, &stderr, strings.NewReader(""),
		false, noEnv, echoAll, workDir,
	)

	if code != 1 {
		t.Errorf("exit code = %d, want 1; stderr: %s", code, stderr.String())
	}
	// No runner activity: stdout should be empty.
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty for empty-task error; got: %s", stdout.String())
	}
}

// (e) unimplemented backend fails fast with exit 1.
func TestE2E_UnimplementedBackend_Exit1(t *testing.T) {
	workDir := makeUnimplementedWorkspace(t)
	var stdout, stderr bytes.Buffer

	// defaultResolver returns error for claude-agent-acp.
	code := parseAndRun(
		[]string{"any task"},
		&stdout, &stderr, strings.NewReader(""),
		false, noEnv, defaultResolver, workDir,
	)

	if code != 1 {
		t.Errorf("exit code = %d, want 1; stderr: %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty on backend error; got: %s", stdout.String())
	}
}

// Additional: --json output has the required structure.
func TestE2E_JSONOutput(t *testing.T) {
	workDir := makeE2EWorkspace(t)
	var stdout, stderr bytes.Buffer

	code := parseAndRun(
		[]string{"--json", "a task"},
		&stdout, &stderr, strings.NewReader(""),
		false, noEnv, echoAll, workDir,
	)

	if code != 0 {
		t.Fatalf("exit code = %d; stderr: %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("--json must suppress stderr trace; got: %s", stderr.String())
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

func keysOf(m map[string]json.RawMessage) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
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
		false, noEnv, echoAll, workDir,
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
