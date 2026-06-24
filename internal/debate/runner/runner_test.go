package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heurema/debate/internal/debate/runner"
	"github.com/heurema/debate/internal/engine/orchestrate"
	"github.com/heurema/debate/internal/engine/transport"
	"github.com/heurema/debate/internal/engine/transport/echo"
	"github.com/heurema/debate/internal/engine/transport/mock"
)

// echoResolver returns the echo transport for any backend name.
func echoResolver(_ string) (transport.Transport, error) {
	return echo.New(), nil
}

// fixedResolver returns a transport that opens sessions keyed by persona ID.
func fixedResolver(sessions map[string]*mock.Session) runner.Resolver {
	return func(_ string) (transport.Transport, error) {
		return mock.NewTransport(sessions), nil
	}
}

// makeWorkspace creates a minimal .heurema/debate workspace under a temp dir
// and returns the temp dir root. Each key in personas maps to a persona filename
// (without .md); the value is the full file content.
func makeWorkspace(t *testing.T, personas map[string]string) string {
	t.Helper()
	root := t.TempDir()
	debDir := filepath.Join(root, ".heurema", "debate")
	personasDir := filepath.Join(debDir, "personas")
	if err := os.MkdirAll(personasDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range personas {
		path := filepath.Join(personasDir, name+".md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const echoPersonaContent = `---
model: echo-local
effort: low
backend: echo
---
You are a debate participant. State your position clearly.
`

func TestRun_Settled(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{
		"alice": echoPersonaContent,
		"bob":   echoPersonaContent,
	})

	result, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "should we use tabs or spaces?",
		MaxRounds: 5,
		Resolver:  echoResolver,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Outcome.Reason != "settled" {
		t.Errorf("outcome = %q, want settled", result.Outcome.Reason)
	}
	if result.Answer == "" {
		t.Error("answer is empty")
	}
	if len(result.Turns) == 0 {
		t.Error("no turns recorded")
	}
}

func TestRun_EmptyTask(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{"alice": echoPersonaContent})

	_, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "   ",
		MaxRounds: 5,
		Resolver:  echoResolver,
	})
	if err == nil {
		t.Fatal("expected error for whitespace-only task, got nil")
	}
}

func TestRun_BriefEqualsTask(t *testing.T) {
	const task = "should we use tabs or spaces?"
	workDir := makeWorkspace(t, map[string]string{"alice": echoPersonaContent})

	const agreedContent = "I agree.\n\n```signal\n{\"position\": \"agree\", \"objections\": [], \"done\": true}\n```"
	aliceSess := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: agreedContent}},
		{Result: transport.Result{Content: agreedContent}},
	})
	synthSess := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "synthesis"}},
	})

	_, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      task,
		MaxRounds: 5,
		Resolver: fixedResolver(map[string]*mock.Session{
			"alice":       aliceSess,
			"synthesizer": synthSess,
		}),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	prompts := aliceSess.Prompts()
	if len(prompts) == 0 {
		t.Fatal("alice session received no prompts")
	}
	// The prompt builder embeds the brief under "## Brief\n\n"; verify the brief is the task alone.
	const briefMarker = "## Brief\n\n"
	idx := strings.Index(prompts[0], briefMarker)
	if idx < 0 {
		t.Fatalf("prompt does not contain %q", briefMarker)
	}
	afterMarker := prompts[0][idx+len(briefMarker):]
	// Brief runs until the next section or end.
	endIdx := strings.Index(afterMarker, "\n\n")
	var brief string
	if endIdx < 0 {
		brief = afterMarker
	} else {
		brief = afterMarker[:endIdx]
	}
	if brief != task {
		t.Errorf("assembled brief = %q, want %q", brief, task)
	}
}

func TestRun_MaxOutcome(t *testing.T) {
	// notDoneContent produces a valid non-done signal so the debate makes "progress"
	// in round 1 (objection set changes from empty to non-empty) but never converges.
	const notDoneContent = "I disagree.\n\n```signal\n" +
		"{\"position\": \"disagree\", \"objections\": [\"needs work\"], \"done\": false}\n```"

	workDir := makeWorkspace(t, map[string]string{
		"alice": echoPersonaContent,
		"bob":   echoPersonaContent,
	})

	// With maxRounds=1: one round runs, outcome is "max" (Max=1 reached).
	result, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "controversial topic",
		MaxRounds: 1,
		Resolver: fixedResolver(map[string]*mock.Session{
			"alice":       mock.NewSession(notDoneScripts(5, notDoneContent)),
			"bob":         mock.NewSession(notDoneScripts(5, notDoneContent)),
			"synthesizer": mock.NewSession([]mock.ScriptedResult{{Result: transport.Result{Content: "synthesis"}}}),
		}),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Outcome.Reason != "max" {
		t.Errorf("outcome = %q, want max", result.Outcome.Reason)
	}
	if result.Answer == "" {
		t.Error("expected non-empty answer even on non-converged run")
	}
}

func TestRun_OnTurnFires(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{"alice": echoPersonaContent})

	var fired []orchestrate.Turn
	_, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "callback test",
		MaxRounds: 5,
		OnTurn:    func(t orchestrate.Turn) { fired = append(fired, t) },
		Resolver:  echoResolver,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(fired) == 0 {
		t.Error("OnTurn was never called")
	}
}

func notDoneScripts(n int, content string) []mock.ScriptedResult {
	out := make([]mock.ScriptedResult, n)
	for i := range out {
		out[i] = mock.ScriptedResult{Result: transport.Result{Content: content}}
	}
	return out
}
