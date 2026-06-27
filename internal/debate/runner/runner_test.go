package runner_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

type countingTransport struct {
	sessions map[string]*mock.Session
	opens    map[string]int
}

func newCountingTransport(sessions map[string]*mock.Session) *countingTransport {
	return &countingTransport{sessions: sessions, opens: make(map[string]int)}
}

func (t *countingTransport) Open(_ context.Context, spec transport.Spec) (transport.Session, error) {
	t.opens[spec.ID]++
	s, ok := t.sessions[spec.ID]
	if !ok {
		return nil, fmt.Errorf("mock: no session configured for id %q", spec.ID)
	}
	return s, nil
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
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var names []string
	for name := range personas {
		if name != "synthesizer" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	var table strings.Builder
	table.WriteString("version: 1\npanel:\n")
	for _, name := range names {
		fmt.Fprintf(&table, "  - %s\n", name)
	}
	tablesDir := filepath.Join(debDir, "tables")
	if err := os.MkdirAll(tablesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tablesDir, "default.yml"), []byte(table.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

const echoPersonaContent = `---
version: 1
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

func TestRun_SynthesizerCalledOnceWithFinalTranscript(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{
		"alice": echoPersonaContent,
		"bob":   echoPersonaContent,
	})

	aliceSess := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "A r1"}},
		{Result: transport.Result{Content: "A r2"}},
	})
	bobSess := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "B r1"}},
		{Result: transport.Result{Content: "B r2"}},
	})
	synthSess := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "synthesis"}},
	})
	tr := newCountingTransport(map[string]*mock.Session{
		"alice":       aliceSess,
		"bob":         bobSess,
		"synthesizer": synthSess,
	})

	result, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "collect all turns",
		MaxRounds: 2,
		Resolver:  func(_ string) (transport.Transport, error) { return tr, nil },
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Turns) != 4 {
		t.Fatalf("turns = %d, want 4", len(result.Turns))
	}
	if tr.opens["synthesizer"] != 1 {
		t.Fatalf("synthesizer opens = %d, want 1", tr.opens["synthesizer"])
	}

	prompts := synthSess.Prompts()
	if len(prompts) != 1 {
		t.Fatalf("synthesizer prompts = %d, want 1", len(prompts))
	}
	for _, want := range []string{
		"Task: collect all turns",
		"[Round 1 — alice]\nA r1",
		"[Round 1 — bob]\nB r1",
		"[Round 2 — alice]\nA r2",
		"[Round 2 — bob]\nB r2",
		"Synthesize the debate:",
	} {
		if !strings.Contains(prompts[0], want) {
			t.Fatalf("synthesizer prompt missing %q\nprompt:\n%s", want, prompts[0])
		}
	}
}

func TestRun_UsesFullPersonaIDsForNamespacedParticipants(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{
		"team/alice": echoPersonaContent,
	})

	aliceSess := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "namespaced turn"}},
	})
	synthSess := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "synthesis"}},
	})
	tr := newCountingTransport(map[string]*mock.Session{
		"team/alice":  aliceSess,
		"synthesizer": synthSess,
	})

	result, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "keep IDs qualified",
		MaxRounds: 1,
		Resolver:  func(_ string) (transport.Transport, error) { return tr, nil },
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if tr.opens["team/alice"] != 1 {
		t.Fatalf("team/alice opens = %d, want 1", tr.opens["team/alice"])
	}
	if len(result.Turns) != 1 || result.Turns[0].Speaker != "team/alice" {
		t.Fatalf("turn speakers = %+v, want team/alice", result.Turns)
	}
	prompts := synthSess.Prompts()
	if len(prompts) != 1 || !strings.Contains(prompts[0], "[Round 1 — team/alice]") {
		t.Fatalf("synthesizer prompt missing full persona ID: %v", prompts)
	}
}

func TestRun_DoesNotSynthesizeAfterDebateError(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{
		"alice": echoPersonaContent,
		"bob":   echoPersonaContent,
	})

	sendErr := errors.New("participant failed")
	aliceSess := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "A partial"}},
	})
	bobSess := mock.NewSession([]mock.ScriptedResult{
		{Err: sendErr},
	})
	synthSess := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "must not be used"}},
	})

	_, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "abort before synthesis",
		MaxRounds: 2,
		Resolver: fixedResolver(map[string]*mock.Session{
			"alice":       aliceSess,
			"bob":         bobSess,
			"synthesizer": synthSess,
		}),
	})
	if err == nil {
		t.Fatal("expected participant error, got nil")
	}
	if !strings.Contains(err.Error(), sendErr.Error()) {
		t.Fatalf("error = %v, want to include %q", err, sendErr)
	}
	if prompts := synthSess.Prompts(); len(prompts) != 0 {
		t.Fatalf("synthesizer prompts = %d, want 0: %v", len(prompts), prompts)
	}
	if synthSess.Closed() {
		t.Fatal("synthesizer session was opened on debate error")
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
