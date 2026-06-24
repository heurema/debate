package acp

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	acpsdk "github.com/coder/acp-go-sdk"
	"github.com/heurema/debate/internal/engine/transport"
)

// --- fake ACP peer ---

// promptScenario configures one Prompt() response from the fake agent.
type promptScenario struct {
	chunks []string          // agent_message_chunk text pieces to stream
	stop   acpsdk.StopReason // stop reason to return
	err    error             // if non-nil, returned as error (becomes InternalError on wire)
}

// fakeAgent implements acpsdk.Agent for deterministic unit tests.
// It uses an injectable queue of per-call scenarios.
type fakeAgent struct {
	mu              sync.Mutex
	asc             *acpsdk.AgentSideConnection
	scenarios       []promptScenario
	callIdx         int
	newSessCwds     []string                // Cwd from each NewSession call, in order
	receivedBlocks  [][]acpsdk.ContentBlock // prompt content from each Prompt call, in order
	newSessionDelay time.Duration
}

func (f *fakeAgent) setScenarios(ss ...promptScenario) {
	f.mu.Lock()
	f.scenarios = ss
	f.callIdx = 0
	f.mu.Unlock()
}

func (f *fakeAgent) Initialize(_ context.Context, _ acpsdk.InitializeRequest) (acpsdk.InitializeResponse, error) {
	return acpsdk.InitializeResponse{ProtocolVersion: acpsdk.ProtocolVersionNumber}, nil
}
func (f *fakeAgent) Authenticate(_ context.Context, _ acpsdk.AuthenticateRequest) (acpsdk.AuthenticateResponse, error) {
	return acpsdk.AuthenticateResponse{}, nil
}
func (f *fakeAgent) Logout(_ context.Context, _ acpsdk.LogoutRequest) (acpsdk.LogoutResponse, error) {
	return acpsdk.LogoutResponse{}, nil
}
func (f *fakeAgent) Cancel(_ context.Context, _ acpsdk.CancelNotification) error { return nil }
func (f *fakeAgent) CloseSession(_ context.Context, _ acpsdk.CloseSessionRequest) (acpsdk.CloseSessionResponse, error) {
	return acpsdk.CloseSessionResponse{}, nil
}
func (f *fakeAgent) ListSessions(_ context.Context, _ acpsdk.ListSessionsRequest) (acpsdk.ListSessionsResponse, error) {
	return acpsdk.ListSessionsResponse{}, nil
}
func (f *fakeAgent) ResumeSession(_ context.Context, _ acpsdk.ResumeSessionRequest) (acpsdk.ResumeSessionResponse, error) {
	return acpsdk.ResumeSessionResponse{}, nil
}
func (f *fakeAgent) SetSessionConfigOption(_ context.Context, _ acpsdk.SetSessionConfigOptionRequest) (acpsdk.SetSessionConfigOptionResponse, error) {
	return acpsdk.SetSessionConfigOptionResponse{}, nil
}
func (f *fakeAgent) SetSessionMode(_ context.Context, _ acpsdk.SetSessionModeRequest) (acpsdk.SetSessionModeResponse, error) {
	return acpsdk.SetSessionModeResponse{}, nil
}

func (f *fakeAgent) NewSession(ctx context.Context, r acpsdk.NewSessionRequest) (acpsdk.NewSessionResponse, error) {
	if f.newSessionDelay > 0 {
		select {
		case <-time.After(f.newSessionDelay):
		case <-ctx.Done():
			return acpsdk.NewSessionResponse{}, ctx.Err()
		}
	}
	f.mu.Lock()
	f.newSessCwds = append(f.newSessCwds, r.Cwd)
	f.mu.Unlock()
	return acpsdk.NewSessionResponse{SessionId: "sess-test"}, nil
}

func (f *fakeAgent) Prompt(ctx context.Context, p acpsdk.PromptRequest) (acpsdk.PromptResponse, error) {
	f.mu.Lock()
	idx := f.callIdx
	f.callIdx++
	f.receivedBlocks = append(f.receivedBlocks, p.Prompt)
	var sc promptScenario
	if idx < len(f.scenarios) {
		sc = f.scenarios[idx]
	} else {
		sc = promptScenario{chunks: []string{"default"}, stop: acpsdk.StopReasonEndTurn}
	}
	asc := f.asc
	f.mu.Unlock()

	if sc.err != nil {
		return acpsdk.PromptResponse{}, sc.err
	}
	if len(sc.chunks) == 0 && sc.stop == "" {
		select {
		case <-ctx.Done():
			return acpsdk.PromptResponse{}, ctx.Err()
		}
	}
	for _, chunk := range sc.chunks {
		_ = asc.SessionUpdate(ctx, acpsdk.SessionNotification{
			SessionId: p.SessionId,
			Update:    acpsdk.UpdateAgentMessageText(chunk),
		})
	}
	return acpsdk.PromptResponse{StopReason: sc.stop}, nil
}

// --- fake runner infrastructure ---

// spawnRecord records one ProcessRunner call.
type spawnRecord struct {
	dir  string
	name string
	args []string
	env  []string
}

// fakeRunnerState tracks spawn and kill calls.
type fakeRunnerState struct {
	mu        sync.Mutex
	spawns    []spawnRecord
	killCount int
}

// newFakeRunner returns a ProcessRunner backed by in-process ACP pipes.
// The same fakeAgent is reused across multiple spawn calls (for recovery tests).
func newFakeRunner(t *testing.T, agent *fakeAgent) (ProcessRunner, *fakeRunnerState) {
	t.Helper()
	state := &fakeRunnerState{}
	var cleanups []func()

	run := func(dir, name string, args, env []string) (io.WriteCloser, io.ReadCloser, func() error, error) {
		state.mu.Lock()
		state.spawns = append(state.spawns, spawnRecord{dir, name, args, env})
		state.mu.Unlock()

		// c2a: client→agent; a2c: agent→client
		c2aR, c2aW := io.Pipe()
		a2cR, a2cW := io.Pipe()

		asc := acpsdk.NewAgentSideConnection(agent, a2cW, c2aR)
		agent.mu.Lock()
		agent.asc = asc
		agent.mu.Unlock()

		var once sync.Once
		kill := func() error {
			once.Do(func() {
				state.mu.Lock()
				state.killCount++
				state.mu.Unlock()
				c2aW.Close()
				c2aR.Close()
				a2cW.Close()
				a2cR.Close()
			})
			return nil
		}
		cleanups = append(cleanups, func() { _ = kill() })
		return c2aW, a2cR, kill, nil
	}

	t.Cleanup(func() {
		for _, fn := range cleanups {
			fn()
		}
	})

	return run, state
}

// openSession opens a session using the transport and fails the test on error.
func openSession(t *testing.T, tr transport.Transport, spec transport.Spec) transport.Session {
	t.Helper()
	sess, err := tr.Open(context.Background(), spec)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = sess.Close() })
	return sess
}

// noEnv always returns "".
func noEnv(_ string) string { return "" }

func envWith(k, v string) func(string) string {
	return func(got string) string {
		if got == k {
			return v
		}
		return ""
	}
}

// --- tests ---

func TestNew_UnknownBackend(t *testing.T) {
	_, err := New("unknown-backend", noEnv, nil)
	if err == nil {
		t.Fatal("want error for unknown backend")
	}
}

func TestNew_ValidBackends(t *testing.T) {
	for _, backend := range []string{BackendClaude, BackendCodex} {
		_, err := New(backend, noEnv, func(_, _ string, _, _ []string) (io.WriteCloser, io.ReadCloser, func() error, error) {
			return nil, nil, nil, fmt.Errorf("unused")
		})
		if err != nil {
			t.Errorf("New(%q): %v", backend, err)
		}
	}
}

func TestBuildCmd_ClaudeDefault(t *testing.T) {
	tr := &acpTransport{backendID: BackendClaude, getEnv: noEnv}
	cmd, env := tr.buildCmd(transport.Spec{Model: "claude-opus", Effort: "high"})
	if len(cmd) < 3 || cmd[0] != "npx" || cmd[1] != "-y" || cmd[2] != defaultClaudePackage {
		t.Errorf("unexpected cmd: %v", cmd)
	}
	if !containsEnv(env, "ANTHROPIC_MODEL=claude-opus") {
		t.Errorf("env missing ANTHROPIC_MODEL: %v", env)
	}
	if !containsEnv(env, "CLAUDE_CODE_EFFORT_LEVEL=high") {
		t.Errorf("env missing CLAUDE_CODE_EFFORT_LEVEL: %v", env)
	}
}

func TestBuildCmd_ClaudeOverride(t *testing.T) {
	getEnv := func(k string) string {
		if k == EnvClaudePackage {
			return "@custom/claude-pkg@1.0.0"
		}
		return ""
	}
	tr := &acpTransport{backendID: BackendClaude, getEnv: getEnv}
	cmd, _ := tr.buildCmd(transport.Spec{Model: "m", Effort: "low"})
	if len(cmd) < 3 || cmd[2] != "@custom/claude-pkg@1.0.0" {
		t.Errorf("expected override package in cmd: %v", cmd)
	}
}

func TestBuildCmd_CodexDefault(t *testing.T) {
	tr := &acpTransport{backendID: BackendCodex, getEnv: noEnv}
	cmd, _ := tr.buildCmd(transport.Spec{Model: "codex-mini", Effort: "low"})
	if len(cmd) < 3 || cmd[2] != defaultCodexPackage {
		t.Errorf("unexpected cmd: %v", cmd)
	}
	if !contains(cmd, "model=codex-mini") {
		t.Errorf("expected model flag in cmd: %v", cmd)
	}
	if !contains(cmd, "sandbox_mode=read-only") {
		t.Errorf("expected sandbox_mode flag in cmd: %v", cmd)
	}
}

func TestBuildCmd_CodexOverride(t *testing.T) {
	getEnv := func(k string) string {
		if k == EnvCodexPackage {
			return "@custom/codex-pkg@2.0.0"
		}
		return ""
	}
	tr := &acpTransport{backendID: BackendCodex, getEnv: getEnv}
	cmd, _ := tr.buildCmd(transport.Spec{Model: "m", Effort: "low"})
	if len(cmd) < 3 || cmd[2] != "@custom/codex-pkg@2.0.0" {
		t.Errorf("expected override package in cmd: %v", cmd)
	}
}

func TestBuildCmd_CodexIgnoresEffort(t *testing.T) {
	tr := &acpTransport{backendID: BackendCodex, getEnv: noEnv}
	// Use a unique sentinel effort value that would only appear if buildCmd explicitly added it.
	_, env := tr.buildCmd(transport.Spec{Model: "codex-mini", Effort: "SENTINEL_EFFORT_VALUE"})
	// Codex effort is intentionally not wired; buildCmd must not add CLAUDE_CODE_EFFORT_LEVEL.
	if containsEnv(env, "CLAUDE_CODE_EFFORT_LEVEL=SENTINEL_EFFORT_VALUE") {
		t.Error("codex buildCmd must not wire CLAUDE_CODE_EFFORT_LEVEL (effort is ignored for codex)")
	}
}

func TestBuildCmd_CodexAlwaysSandboxReadOnly(t *testing.T) {
	tr := &acpTransport{backendID: BackendCodex, getEnv: noEnv}
	// Both grounded and sealed specs should produce sandbox_mode=read-only.
	for _, readOnly := range []bool{false, true} {
		cmd, _ := tr.buildCmd(transport.Spec{Model: "m", ReadOnly: readOnly})
		if !contains(cmd, "sandbox_mode=read-only") {
			t.Errorf("ReadOnly=%v: expected sandbox_mode=read-only in cmd: %v", readOnly, cmd)
		}
	}
}

func TestOpen_MissingModel(t *testing.T) {
	agent := &fakeAgent{}
	run, _ := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	_, err := tr.Open(context.Background(), transport.Spec{ID: "p1", Model: ""})
	if err == nil {
		t.Fatal("want error for empty model")
	}
}

func TestOpen_Handshake(t *testing.T) {
	agent := &fakeAgent{}
	run, state := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	_ = openSession(t, tr, transport.Spec{ID: "p1", Model: "m", Effort: "low"})
	if len(state.spawns) != 1 {
		t.Errorf("want 1 spawn, got %d", len(state.spawns))
	}
}

func TestOpen_NewSessionTimeout(t *testing.T) {
	agent := &fakeAgent{newSessionDelay: 50 * time.Millisecond}
	run, state := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, envWith(EnvOpenTimeout, "1ms"), run)

	_, err := tr.Open(context.Background(), transport.Spec{ID: "p1", Model: "m"})
	if err == nil {
		t.Fatal("want timeout error")
	}
	if cls := transport.Classify(err); cls.Kind != "idle_timeout" {
		t.Fatalf("Classify = %+v, err=%v; want idle_timeout", cls, err)
	}
	if state.killCount == 0 {
		t.Fatal("want spawned ACP process killed on open timeout")
	}
}

func TestSend_EndTurn(t *testing.T) {
	agent := &fakeAgent{}
	agent.setScenarios(promptScenario{
		chunks: []string{"hello", " world"},
		stop:   acpsdk.StopReasonEndTurn,
	})
	run, _ := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	sess := openSession(t, tr, transport.Spec{ID: "p1", Model: "m"})

	result, err := sess.Send(context.Background(), "ping")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if result.Content != "hello world" {
		t.Errorf("content = %q, want %q", result.Content, "hello world")
	}
}

func TestSend_PromptTimeout(t *testing.T) {
	agent := &fakeAgent{}
	agent.setScenarios(
		promptScenario{}, // blocks until context deadline
		promptScenario{}, // retry after recovery also blocks
	)
	run, state := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, envWith(EnvSendTimeout, "1ms"), run)
	sess := openSession(t, tr, transport.Spec{ID: "p1", Model: "m"})

	_, err := sess.Send(context.Background(), "ping")
	if err == nil {
		t.Fatal("want timeout error")
	}
	if cls := transport.Classify(err); cls.Kind != "idle_timeout" {
		t.Fatalf("Classify = %+v, err=%v; want idle_timeout", cls, err)
	}
	if len(state.spawns) != 1 {
		t.Fatalf("timeout must not retry; got %d spawns", len(state.spawns))
	}
	if state.killCount == 0 {
		t.Fatal("want timed-out ACP session killed")
	}
}

func TestSend_MultipleTurns(t *testing.T) {
	agent := &fakeAgent{}
	agent.setScenarios(
		promptScenario{chunks: []string{"turn1"}, stop: acpsdk.StopReasonEndTurn},
		promptScenario{chunks: []string{"turn2"}, stop: acpsdk.StopReasonEndTurn},
	)
	run, _ := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	sess := openSession(t, tr, transport.Spec{ID: "p1", Model: "m"})

	r1, err := sess.Send(context.Background(), "first")
	if err != nil || r1.Content != "turn1" {
		t.Fatalf("first Send: content=%q err=%v", r1.Content, err)
	}
	r2, err := sess.Send(context.Background(), "second")
	if err != nil || r2.Content != "turn2" {
		t.Fatalf("second Send: content=%q err=%v", r2.Content, err)
	}
}

func TestSend_SystemPromptInjected(t *testing.T) {
	// spec.System must be prepended to the first Prompt call and omitted from subsequent ones.
	agent := &fakeAgent{}
	agent.setScenarios(
		promptScenario{chunks: []string{"r1"}, stop: acpsdk.StopReasonEndTurn},
		promptScenario{chunks: []string{"r2"}, stop: acpsdk.StopReasonEndTurn},
	)
	run, _ := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	sess := openSession(t, tr, transport.Spec{ID: "p1", Model: "m", System: "you are alice"})

	ctx := context.Background()
	if _, err := sess.Send(ctx, "turn1"); err != nil {
		t.Fatalf("turn1: %v", err)
	}
	if _, err := sess.Send(ctx, "turn2"); err != nil {
		t.Fatalf("turn2: %v", err)
	}

	agent.mu.Lock()
	blocks := agent.receivedBlocks
	agent.mu.Unlock()

	if len(blocks) != 2 {
		t.Fatalf("want 2 Prompt calls, got %d", len(blocks))
	}
	// First call: system block + turn block.
	if len(blocks[0]) != 2 {
		t.Errorf("first Prompt: want 2 blocks (system+turn), got %d", len(blocks[0]))
	}
	// Second call: turn block only; system must not be repeated.
	if len(blocks[1]) != 1 {
		t.Errorf("second Prompt: want 1 block (turn only), got %d", len(blocks[1]))
	}
}

func TestSend_Refusal(t *testing.T) {
	agent := &fakeAgent{}
	agent.setScenarios(promptScenario{stop: acpsdk.StopReasonRefusal})
	run, _ := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	sess := openSession(t, tr, transport.Spec{ID: "p1", Model: "m"})

	_, err := sess.Send(context.Background(), "prompt")
	if err == nil {
		t.Fatal("want error on refusal")
	}
	cls := transport.Classify(err)
	if cls.Retryable {
		t.Errorf("refusal must be non-retryable, got %+v", cls)
	}
}

func TestSend_NonEndTurnStop(t *testing.T) {
	agent := &fakeAgent{}
	agent.setScenarios(promptScenario{stop: acpsdk.StopReasonMaxTokens})
	run, _ := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	sess := openSession(t, tr, transport.Spec{ID: "p1", Model: "m"})

	_, err := sess.Send(context.Background(), "prompt")
	if err == nil {
		t.Fatal("want error on max_tokens stop")
	}
	cls := transport.Classify(err)
	if cls.Retryable {
		t.Errorf("max_tokens must be non-retryable, got %+v", cls)
	}
}

func TestSend_RetryableDropRecovery(t *testing.T) {
	// Scenario: first Prompt returns an error (simulates transport drop via InternalError).
	// Recovery reopens, replays prior history (none here), retries → succeeds.
	agent := &fakeAgent{}
	agent.setScenarios(
		promptScenario{err: fmt.Errorf("simulated drop")},                      // → InternalError → retryable
		promptScenario{chunks: []string{"ok"}, stop: acpsdk.StopReasonEndTurn}, // retry succeeds
	)
	run, state := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	sess := openSession(t, tr, transport.Spec{ID: "p1", Model: "m"})

	result, err := sess.Send(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Send after recovery: %v", err)
	}
	if result.Content != "ok" {
		t.Errorf("content = %q, want %q", result.Content, "ok")
	}
	// Two spawns: original + recovery reopen.
	if len(state.spawns) != 2 {
		t.Errorf("want 2 spawns, got %d", len(state.spawns))
	}
	// Original session killed during recovery.
	if state.killCount < 1 {
		t.Errorf("want at least 1 kill, got %d", state.killCount)
	}
}

func TestSend_RecoveryWithHistoryReplay(t *testing.T) {
	// Scenario: send two prompts successfully, then drop on third.
	// Recovery replays first two, then retries third.
	agent := &fakeAgent{}
	agent.setScenarios(
		promptScenario{chunks: []string{"r1"}, stop: acpsdk.StopReasonEndTurn},
		promptScenario{chunks: []string{"r2"}, stop: acpsdk.StopReasonEndTurn},
		promptScenario{err: fmt.Errorf("drop")},                                     // third call drops
		promptScenario{chunks: []string{"replay1"}, stop: acpsdk.StopReasonEndTurn}, // replay of p1
		promptScenario{chunks: []string{"replay2"}, stop: acpsdk.StopReasonEndTurn}, // replay of p2
		promptScenario{chunks: []string{"r3"}, stop: acpsdk.StopReasonEndTurn},      // retry of p3
	)
	run, _ := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	sess := openSession(t, tr, transport.Spec{ID: "p1", Model: "m"})

	ctx := context.Background()
	if _, err := sess.Send(ctx, "p1"); err != nil {
		t.Fatalf("p1: %v", err)
	}
	if _, err := sess.Send(ctx, "p2"); err != nil {
		t.Fatalf("p2: %v", err)
	}
	result, err := sess.Send(ctx, "p3")
	if err != nil {
		t.Fatalf("p3 after recovery: %v", err)
	}
	if result.Content != "r3" {
		t.Errorf("p3 content = %q, want %q", result.Content, "r3")
	}
}

func TestSend_RecoveryReplayFailure(t *testing.T) {
	// Scenario: drop on first call, recovery opens ok, but replay fails.
	agent := &fakeAgent{}
	agent.setScenarios(
		promptScenario{chunks: []string{"hist"}, stop: acpsdk.StopReasonEndTurn}, // first call ok
		promptScenario{err: fmt.Errorf("drop")},                                  // second call drops
		promptScenario{err: fmt.Errorf("replay also fails")},                     // replay of hist fails
	)
	run, _ := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	sess := openSession(t, tr, transport.Spec{ID: "p1", Model: "m"})

	ctx := context.Background()
	if _, err := sess.Send(ctx, "hist"); err != nil {
		t.Fatalf("hist: %v", err)
	}
	_, err := sess.Send(ctx, "trigger-drop")
	if err == nil {
		t.Fatal("want error when replay fails")
	}
}

func TestSend_SecondDropIsTerminal(t *testing.T) {
	// Scenario: both the original call and the retry after recovery drop.
	// The second drop should be returned as an error without further recovery.
	agent := &fakeAgent{}
	agent.setScenarios(
		promptScenario{err: fmt.Errorf("drop1")}, // first attempt drops
		promptScenario{err: fmt.Errorf("drop2")}, // retry after recovery also drops
	)
	run, state := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	sess := openSession(t, tr, transport.Spec{ID: "p1", Model: "m"})

	_, err := sess.Send(context.Background(), "hello")
	if err == nil {
		t.Fatal("want error on second drop")
	}
	// Two spawns: original + recovery.
	if len(state.spawns) != 2 {
		t.Errorf("want 2 spawns, got %d", len(state.spawns))
	}
}

func TestClose_Idempotent(t *testing.T) {
	agent := &fakeAgent{}
	run, state := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	sess, err := tr.Open(context.Background(), transport.Spec{ID: "p1", Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := sess.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	// Kill should be called at most once.
	if state.killCount > 1 {
		t.Errorf("Close should only kill once, got %d kills", state.killCount)
	}
}

func TestGrounded_Cwd(t *testing.T) {
	agent := &fakeAgent{}
	run, state := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	_ = openSession(t, tr, transport.Spec{ID: "p1", Model: "m", ReadOnly: false})

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.spawns) == 0 {
		t.Fatal("no spawn recorded")
	}
	if state.spawns[0].dir != wd {
		t.Errorf("grounded spawn dir = %q, want %q", state.spawns[0].dir, wd)
	}
	agent.mu.Lock()
	if len(agent.newSessCwds) == 0 || agent.newSessCwds[0] != wd {
		t.Errorf("NewSession cwd = %v, want %q", agent.newSessCwds, wd)
	}
	agent.mu.Unlock()
}

func TestSealed_Cwd(t *testing.T) {
	agent := &fakeAgent{}
	run, state := newFakeRunner(t, agent)
	tr, _ := New(BackendClaude, noEnv, run)
	_ = openSession(t, tr, transport.Spec{ID: "p1", Model: "m", ReadOnly: true})

	wd, _ := os.Getwd()
	if len(state.spawns) == 0 {
		t.Fatal("no spawn recorded")
	}
	spawnDir := state.spawns[0].dir
	if spawnDir == wd {
		t.Error("sealed spawn dir must not be the process cwd")
	}
	if spawnDir == "" {
		t.Error("sealed spawn dir must not be empty")
	}
	// The NewSession cwd must match the spawn dir.
	agent.mu.Lock()
	defer agent.mu.Unlock()
	if len(agent.newSessCwds) == 0 {
		t.Fatal("no NewSession call recorded")
	}
	if agent.newSessCwds[0] != spawnDir {
		t.Errorf("NewSession cwd = %q, want spawn dir %q", agent.newSessCwds[0], spawnDir)
	}
}

// --- helpers ---

func containsEnv(env []string, kv string) bool {
	for _, e := range env {
		if e == kv {
			return true
		}
	}
	return false
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
