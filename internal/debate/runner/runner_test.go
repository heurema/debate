package runner_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/heurema/debate/internal/debate/progress"
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

type sessionTransport struct {
	sessions map[string]transport.Session
}

func (t sessionTransport) Open(_ context.Context, spec transport.Spec) (transport.Session, error) {
	s, ok := t.sessions[spec.ID]
	if !ok {
		return nil, fmt.Errorf("mock: no session configured for id %q", spec.ID)
	}
	return s, nil
}

type blockingSession struct {
	entered chan struct{}
	release chan struct{}
	result  transport.Result
}

func newBlockingSession(content string) *blockingSession {
	return &blockingSession{
		entered: make(chan struct{}),
		release: make(chan struct{}),
		result:  transport.Result{Content: content},
	}
}

func (s *blockingSession) Send(ctx context.Context, _ string) (transport.Result, error) {
	close(s.entered)
	select {
	case <-s.release:
		return s.result, nil
	case <-ctx.Done():
		return transport.Result{}, ctx.Err()
	}
}

func (s *blockingSession) Close() error { return nil }

type progressCapture struct {
	mu  bytes.Buffer
	ch  chan string
	mtx sync.Mutex
}

func newProgressCapture() *progressCapture {
	return &progressCapture{ch: make(chan string, 100)}
}

func (w *progressCapture) Write(p []byte) (int, error) {
	w.mtx.Lock()
	w.mu.Write(p)
	w.mtx.Unlock()
	w.ch <- string(p)
	return len(p), nil
}

func (w *progressCapture) String() string {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.mu.String()
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

func TestRun_ProgressLifecycle(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{
		"alice": echoPersonaContent,
		"bob":   echoPersonaContent,
	})
	var stderr bytes.Buffer

	_, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "track progress",
		MaxRounds: 1,
		Resolver: fixedResolver(map[string]*mock.Session{
			"alice": mock.NewSession([]mock.ScriptedResult{
				{Result: transport.Result{Content: "A"}},
			}),
			"bob": mock.NewSession([]mock.ScriptedResult{
				{Result: transport.Result{Content: "B"}},
			}),
			"synthesizer": mock.NewSession([]mock.ScriptedResult{
				{Result: transport.Result{Content: "synthesis"}},
			}),
		}),
		Progress: progress.NewEmitter(&stderr),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	events := parseProgressEvents(t, stderr.String())
	got := progressTypes(events)
	want := []string{
		"run_started",
		"workspace_loaded",
		"session_opening",
		"session_opened",
		"session_opening",
		"session_opened",
		"round_started",
		"turn_started",
		"turn_completed",
		"turn_started",
		"turn_completed",
		"round_completed",
		"synthesis_started",
		"synthesis_completed",
		"run_completed",
	}
	if !equalStrings(got, want) {
		t.Fatalf("progress event types = %v, want %v", got, want)
	}
	if events[len(events)-1].Stage != "completed" {
		t.Fatalf("final progress stage = %q, want completed", events[len(events)-1].Stage)
	}
	if events[2].Speaker != "alice" || events[4].Speaker != "bob" {
		t.Fatalf("session events use wrong speakers: %+v", events)
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

func TestRun_ProgressRunFailedUsesActiveTurnStage(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{"alice": echoPersonaContent})
	sendErr := errors.New("participant failed")
	var stderr bytes.Buffer

	_, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "fail during turn",
		MaxRounds: 1,
		Resolver: fixedResolver(map[string]*mock.Session{
			"alice":       mock.NewSession([]mock.ScriptedResult{{Err: sendErr}}),
			"synthesizer": mock.NewSession([]mock.ScriptedResult{{Result: transport.Result{Content: "unused"}}}),
		}),
		Progress: progress.NewEmitter(&stderr),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	requireRunFailedStage(t, stderr.String(), "running_turn", sendErr.Error())
}

func TestRun_ProgressRunFailedUsesFailedStageWhenNoLifecycleActive(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{"alice": echoPersonaContent})
	resolverErr := errors.New("resolver failed")
	var stderr bytes.Buffer

	_, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "fail before session opening",
		MaxRounds: 1,
		Resolver: func(_ string) (transport.Transport, error) {
			return nil, resolverErr
		},
		Progress: progress.NewEmitter(&stderr),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	requireRunFailedStage(t, stderr.String(), "failed", resolverErr.Error())
}

func TestRun_ProgressRunFailedUsesOpeningSessionStage(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{"alice": echoPersonaContent})
	var stderr bytes.Buffer

	_, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "fail opening session",
		MaxRounds: 1,
		Resolver: fixedResolver(map[string]*mock.Session{
			"synthesizer": mock.NewSession([]mock.ScriptedResult{{Result: transport.Result{Content: "unused"}}}),
		}),
		Progress: progress.NewEmitter(&stderr),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	requireRunFailedStage(t, stderr.String(), "opening_session", "no session configured")
}

func TestRun_ProgressRunFailedUsesSynthesizingStage(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{"alice": echoPersonaContent})
	synthErr := errors.New("synthesizer failed")
	var stderr bytes.Buffer

	_, err := runner.Run(context.Background(), runner.Config{
		WorkDir:   workDir,
		Task:      "fail during synthesis",
		MaxRounds: 1,
		Resolver: fixedResolver(map[string]*mock.Session{
			"alice":       mock.NewSession([]mock.ScriptedResult{{Result: transport.Result{Content: "A"}}}),
			"synthesizer": mock.NewSession([]mock.ScriptedResult{{Err: synthErr}}),
		}),
		Progress: progress.NewEmitter(&stderr),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	requireRunFailedStage(t, stderr.String(), "synthesizing", synthErr.Error())
}

func requireRunFailedStage(t *testing.T, text, wantStage, wantError string) {
	t.Helper()
	events := parseProgressEvents(t, text)
	last := events[len(events)-1]
	if last.Type != "run_failed" {
		t.Fatalf("last event type = %q, want run_failed; events=%v", last.Type, progressTypes(events))
	}
	if last.Stage != wantStage {
		t.Fatalf("run_failed stage = %q, want %s", last.Stage, wantStage)
	}
	if !strings.Contains(last.Error, wantError) {
		t.Fatalf("run_failed error = %q, want to include %q", last.Error, wantError)
	}
}

func TestRun_SynthesisHeartbeatHasSynthesizingStageOnly(t *testing.T) {
	workDir := makeWorkspace(t, map[string]string{"alice": echoPersonaContent})
	synth := newBlockingSession("synthesis")
	capture := newProgressCapture()

	resolver := func(_ string) (transport.Transport, error) {
		return sessionTransport{sessions: map[string]transport.Session{
			"alice": mock.NewSession([]mock.ScriptedResult{
				{Result: transport.Result{Content: "A"}},
			}),
			"synthesizer": synth,
		}}, nil
	}

	done := make(chan error, 1)
	go func() {
		_, err := runner.Run(context.Background(), runner.Config{
			WorkDir:           workDir,
			Task:              "synth heartbeat",
			MaxRounds:         1,
			Resolver:          resolver,
			Progress:          progress.NewEmitter(capture),
			HeartbeatInterval: time.Millisecond,
		})
		done <- err
	}()

	select {
	case <-synth.entered:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("synthesizer send did not start")
	}

	var heartbeat progressWireEvent
	for {
		select {
		case line := <-capture.ch:
			events := parseProgressEvents(t, line)
			for _, ev := range events {
				if ev.Type == "heartbeat" && ev.Stage == "synthesizing" {
					heartbeat = ev
					goto release
				}
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatal("synthesizer heartbeat was not emitted")
		}
	}

release:
	if heartbeat.SilenceMS == nil || *heartbeat.SilenceMS < 1 {
		t.Fatalf("heartbeat silence_ms = %v, want >= 1", heartbeat.SilenceMS)
	}
	if heartbeat.Round != nil || heartbeat.Speaker != "" {
		t.Fatalf("synthesizer heartbeat invented round/speaker: %+v", heartbeat)
	}
	close(synth.release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("run did not finish after releasing synthesizer")
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

type progressWireEvent struct {
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

func parseProgressEvents(t *testing.T, text string) []progressWireEvent {
	t.Helper()
	var events []progressWireEvent
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, progress.Prefix) {
			t.Fatalf("progress line missing prefix %q: %q", progress.Prefix, line)
		}
		var ev progressWireEvent
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

func progressTypes(events []progressWireEvent) []string {
	out := make([]string, len(events))
	for i, ev := range events {
		out[i] = ev.Type
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
