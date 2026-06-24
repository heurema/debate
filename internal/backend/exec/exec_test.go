package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/heurema/debate/internal/engine/transport"
)

// noEnv returns "" for all keys.
func noEnv(_ string) string { return "" }

// --- fake command runner ---

// callConfig pre-configures one fake subprocess invocation.
type callConfig struct {
	stdout   string
	stderr   string
	exitErr  error
	spawnErr error // if set, the runner returns this instead of spawning
	// closeStdinEarly closes the stdinR pipe immediately, causing a broken-pipe
	// error when sendOnce tries to write stdin.
	closeStdinEarly bool
	// blockUntilCtx makes wait() block on ctx.Done(), simulating a long-running process.
	blockUntilCtx bool
}

// callRecord captures what was actually passed to one runner invocation.
type callRecord struct {
	name  string
	args  []string
	dir   string
	stdin []byte
}

type fakeRunner struct {
	mu      sync.Mutex
	configs []callConfig
	records []callRecord
	idx     int
}

func (f *fakeRunner) run(ctx context.Context, name string, args []string, dir string) (
	io.WriteCloser, io.ReadCloser, io.ReadCloser, func() error, error,
) {
	f.mu.Lock()
	idx := f.idx
	f.idx++
	var cfg callConfig
	if idx < len(f.configs) {
		cfg = f.configs[idx]
	}
	f.mu.Unlock()

	if cfg.spawnErr != nil {
		return nil, nil, nil, nil, cfg.spawnErr
	}

	stdinR, stdinW := io.Pipe()

	if cfg.closeStdinEarly {
		// Close the read end now so writes from sendOnce fail with ErrClosedPipe.
		stdinR.Close()
	}

	// Drain stdin into the record asynchronously.
	stdinDone := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(stdinR)
		stdinDone <- data
	}()

	stdout := io.NopCloser(strings.NewReader(cfg.stdout))
	stderr := io.NopCloser(strings.NewReader(cfg.stderr))

	wait := func() error {
		stdinData := <-stdinDone
		f.mu.Lock()
		f.records = append(f.records, callRecord{
			name:  name,
			args:  args,
			dir:   dir,
			stdin: stdinData,
		})
		f.mu.Unlock()

		if cfg.blockUntilCtx {
			<-ctx.Done()
			return ctx.Err()
		}
		return cfg.exitErr
	}

	return stdinW, stdout, stderr, wait, nil
}

func (f *fakeRunner) record(i int) callRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.records[i]
}

func (f *fakeRunner) recordCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.records)
}

// openSession opens a session and registers cleanup.
func openSession(t *testing.T, tr transport.Transport, spec transport.Spec) transport.Session {
	t.Helper()
	sess, err := tr.Open(context.Background(), spec)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = sess.Close() })
	return sess
}

// --- New / Open ---

func TestNew_InvalidBackend(t *testing.T) {
	_, err := New("unknown-backend", noEnv, nil)
	if err == nil {
		t.Fatal("want error for unknown backend")
	}
}

func TestNew_ValidBackend(t *testing.T) {
	tr, err := New(BackendAgy, noEnv, (&fakeRunner{}).run)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if tr == nil {
		t.Fatal("New returned nil transport")
	}
}

func TestOpen_MissingModel(t *testing.T) {
	tr, _ := New(BackendAgy, noEnv, (&fakeRunner{}).run)
	_, err := tr.Open(context.Background(), transport.Spec{ID: "p", Model: ""})
	if err == nil {
		t.Fatal("want error for empty model")
	}
}

// --- Working directory ---

func TestOpen_Grounded_Cwd(t *testing.T) {
	fr := &fakeRunner{configs: []callConfig{{stdout: "reply"}}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{ID: "p", Model: "m", ReadOnly: false})

	_, err := sess.Send(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	wd, _ := os.Getwd()
	rec := fr.record(0)
	if rec.dir != wd {
		t.Errorf("grounded dir = %q, want %q", rec.dir, wd)
	}
}

func TestOpen_Sealed_Cwd(t *testing.T) {
	fr := &fakeRunner{configs: []callConfig{{stdout: "reply"}}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{ID: "p", Model: "m", ReadOnly: true})

	_, err := sess.Send(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	wd, _ := os.Getwd()
	rec := fr.record(0)
	if rec.dir == wd {
		t.Error("sealed dir must not be the process cwd")
	}
	if rec.dir == "" {
		t.Error("sealed dir must not be empty")
	}
	// Temp dir should be removed on Close.
	_ = sess.Close()
	if _, err := os.Stat(rec.dir); !os.IsNotExist(err) {
		t.Errorf("sealed temp dir %q still exists after Close", rec.dir)
	}
}

// --- Command resolution ---

func TestCmd_Default(t *testing.T) {
	fr := &fakeRunner{configs: []callConfig{{stdout: "r"}}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "gemini-pro"})

	if _, err := sess.Send(context.Background(), "p"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	rec := fr.record(0)
	if rec.name != "agy" {
		t.Errorf("name = %q, want %q", rec.name, "agy")
	}
	if len(rec.args) < 2 || rec.args[0] != "--model" || rec.args[1] != "gemini-pro" {
		t.Errorf("args = %v, want [--model gemini-pro ...]", rec.args)
	}
}

func TestCmd_Override(t *testing.T) {
	getEnv := func(k string) string {
		if k == EnvAgyCommand {
			return "/usr/local/bin/custom-agy"
		}
		return ""
	}
	fr := &fakeRunner{configs: []callConfig{{stdout: "r"}}}
	tr, _ := New(BackendAgy, getEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "gemini-flash"})

	if _, err := sess.Send(context.Background(), "p"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	rec := fr.record(0)
	if rec.name != "/usr/local/bin/custom-agy" {
		t.Errorf("name = %q, want %q", rec.name, "/usr/local/bin/custom-agy")
	}
	// Model arg must still be wired.
	if len(rec.args) < 2 || rec.args[0] != "--model" || rec.args[1] != "gemini-flash" {
		t.Errorf("args = %v, want [--model gemini-flash ...]", rec.args)
	}
}

func TestCmd_ModelWired(t *testing.T) {
	fr := &fakeRunner{configs: []callConfig{{stdout: "r"}}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "gemini-ultra"})
	if _, err := sess.Send(context.Background(), "p"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	rec := fr.record(0)
	found := false
	for _, a := range rec.args {
		if a == "gemini-ultra" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("model not found in args: %v", rec.args)
	}
}

// --- stdin format ---

func TestSend_StdinFormat_NoSystem(t *testing.T) {
	// Single send, no system prompt: stdin is just [prompt]\n<content>\n
	fr := &fakeRunner{configs: []callConfig{{stdout: "r1"}}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	if _, err := sess.Send(context.Background(), "hello"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	got := string(fr.record(0).stdin)
	want := "[prompt]\nhello\n"
	if got != want {
		t.Errorf("stdin = %q, want %q", got, want)
	}
}

func TestSend_StdinFormat_WithSystem(t *testing.T) {
	// Single send, with system prompt: [system]\n<sys>\n\n[prompt]\n<prompt>\n
	fr := &fakeRunner{configs: []callConfig{{stdout: "r1"}}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m", System: "you are alice"})

	if _, err := sess.Send(context.Background(), "what is your name?"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	got := string(fr.record(0).stdin)
	want := "[system]\nyou are alice\n\n[prompt]\nwhat is your name?\n"
	if got != want {
		t.Errorf("stdin = %q, want %q", got, want)
	}
}

func TestSend_StdinFormat_SystemAbsent(t *testing.T) {
	// No system prompt: the [system] block must be entirely absent.
	fr := &fakeRunner{configs: []callConfig{{stdout: "r1"}}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m", System: ""})

	if _, err := sess.Send(context.Background(), "ping"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	got := string(fr.record(0).stdin)
	if strings.Contains(got, "[system]") {
		t.Errorf("stdin must not contain [system] when System is empty; got %q", got)
	}
}

func TestSend_StdinFormat_MultiTurn(t *testing.T) {
	// Three sends with a system prompt. Verify the exact byte sequence.
	fr := &fakeRunner{configs: []callConfig{
		{stdout: "reply one"},
		{stdout: "reply two"},
		{stdout: "reply three"},
	}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m", System: "sys"})

	ctx := context.Background()
	if _, err := sess.Send(ctx, "prompt one"); err != nil {
		t.Fatalf("send 1: %v", err)
	}
	if _, err := sess.Send(ctx, "prompt two"); err != nil {
		t.Fatalf("send 2: %v", err)
	}
	if _, err := sess.Send(ctx, "prompt three"); err != nil {
		t.Fatalf("send 3: %v", err)
	}

	// First send: system + prompt.
	want1 := "[system]\nsys\n\n[prompt]\nprompt one\n"
	got1 := string(fr.record(0).stdin)
	if got1 != want1 {
		t.Errorf("send 1 stdin = %q\nwant %q", got1, want1)
	}

	// Second send: system + turn1 prompt + turn1 reply + current prompt.
	want2 := "[system]\nsys\n\n[prompt]\nprompt one\n\n[reply]\nreply one\n\n[prompt]\nprompt two\n"
	got2 := string(fr.record(1).stdin)
	if got2 != want2 {
		t.Errorf("send 2 stdin = %q\nwant %q", got2, want2)
	}

	// Third send: system + turn1 + turn2 + current prompt.
	want3 := "[system]\nsys\n\n" +
		"[prompt]\nprompt one\n\n[reply]\nreply one\n\n" +
		"[prompt]\nprompt two\n\n[reply]\nreply two\n\n" +
		"[prompt]\nprompt three\n"
	got3 := string(fr.record(2).stdin)
	if got3 != want3 {
		t.Errorf("send 3 stdin = %q\nwant %q", got3, want3)
	}
}

func TestSend_StdinFormat_ContentTrailingNewline(t *testing.T) {
	// Content that already ends with \n should not get an extra one.
	fr := &fakeRunner{configs: []callConfig{{stdout: "reply\n"}}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	if _, err := sess.Send(context.Background(), "prompt with newline\n"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	got := string(fr.record(0).stdin)
	// Should not have double newline at content end.
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("unexpected triple newline in stdin: %q", got)
	}
	want := "[prompt]\nprompt with newline\n"
	if got != want {
		t.Errorf("stdin = %q, want %q", got, want)
	}
}

func TestSend_PriorRepliesInHistory(t *testing.T) {
	// Prior agent replies must appear verbatim under [reply] labels.
	fr := &fakeRunner{configs: []callConfig{
		{stdout: "agent reply verbatim"},
		{stdout: "r2"},
	}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	ctx := context.Background()
	r1, err := sess.Send(ctx, "first prompt")
	if err != nil {
		t.Fatalf("send 1: %v", err)
	}
	if r1.Content != "agent reply verbatim" {
		t.Fatalf("unexpected reply: %q", r1.Content)
	}

	if _, err := sess.Send(ctx, "second prompt"); err != nil {
		t.Fatalf("send 2: %v", err)
	}

	stdin2 := string(fr.record(1).stdin)
	if !strings.Contains(stdin2, "[reply]\nagent reply verbatim\n") {
		t.Errorf("second send stdin must contain prior reply verbatim; got %q", stdin2)
	}
}

// --- history commitment ---

func TestSend_FailedSendNoHistoryPollution(t *testing.T) {
	// A failed Send (non-retryable error) must not commit to history.
	// The subsequent Send must reconstruct stdin without the failed turn.
	fr := &fakeRunner{configs: []callConfig{
		{exitErr: errors.New("process error")}, // non-zero exit → terminal
		{stdout: "ok"},
	}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	ctx := context.Background()
	_, err := sess.Send(ctx, "bad prompt")
	if err == nil {
		t.Fatal("want error on non-zero exit")
	}

	// Subsequent send: history must be empty (only [prompt]\n<current prompt>\n).
	if _, err := sess.Send(ctx, "good prompt"); err != nil {
		t.Fatalf("send 2: %v", err)
	}
	stdin2 := string(fr.record(1).stdin)
	want := "[prompt]\ngood prompt\n"
	if stdin2 != want {
		t.Errorf("stdin after failed send = %q, want %q (history must be clean)", stdin2, want)
	}
}

// --- error classification ---

func TestSend_SpawnFailure_Retryable(t *testing.T) {
	fr := &fakeRunner{configs: []callConfig{
		{spawnErr: errors.New("exec: not found")},
		{stdout: "ok"},
	}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	result, err := sess.Send(context.Background(), "p")
	if err != nil {
		t.Fatalf("Send after retry: %v", err)
	}
	if result.Content != "ok" {
		t.Errorf("content = %q, want %q", result.Content, "ok")
	}
	if fr.recordCount() != 1 {
		// Only the second (successful) call records; the first returns spawnErr before recording.
		t.Errorf("want 1 recorded call (successful), got %d", fr.recordCount())
	}
}

func TestSend_SpawnFailure_BothFail(t *testing.T) {
	fr := &fakeRunner{configs: []callConfig{
		{spawnErr: errors.New("spawn1")},
		{spawnErr: errors.New("spawn2")},
	}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	_, err := sess.Send(context.Background(), "p")
	if err == nil {
		t.Fatal("want error when both spawns fail")
	}
	cls := transport.Classify(err)
	if !cls.Retryable {
		t.Errorf("spawn failure must be retryable, got %+v", cls)
	}
}

func TestSend_NonZeroExit_Terminal(t *testing.T) {
	fr := &fakeRunner{configs: []callConfig{
		{stderr: "bad model name", exitErr: errors.New("exit 1")},
	}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	_, err := sess.Send(context.Background(), "p")
	if err == nil {
		t.Fatal("want error on non-zero exit")
	}
	cls := transport.Classify(err)
	if cls.Retryable {
		t.Errorf("non-zero exit must be terminal (non-retryable), got %+v", cls)
	}
	if !strings.Contains(err.Error(), "bad model name") {
		t.Errorf("error must include captured stderr; got %q", err.Error())
	}
}

func TestSend_NonZeroExit_StderrInError(t *testing.T) {
	stderrContent := strings.Repeat("x", maxStderrBytes+10)
	fr := &fakeRunner{configs: []callConfig{
		{stderr: stderrContent, exitErr: fmt.Errorf("exit 2")},
	}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	_, err := sess.Send(context.Background(), "p")
	if err == nil {
		t.Fatal("want error on non-zero exit")
	}
	if !strings.Contains(err.Error(), "truncated") {
		t.Errorf("oversized stderr must be truncated in error message; got %q", err.Error())
	}
}

func TestSend_BrokenPipe_Retried(t *testing.T) {
	// First call: stdinR is closed immediately, causing broken-pipe when sendOnce writes.
	// Second call (retry): succeeds.
	// The broken-pipe call still reaches wait() and appends record(0) with empty stdin;
	// the retry appends record(1) with the actual prompt.
	fr := &fakeRunner{configs: []callConfig{
		{closeStdinEarly: true},
		{stdout: "success reply"},
	}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	result, err := sess.Send(context.Background(), "p")
	if err != nil {
		t.Fatalf("Send after retry: %v", err)
	}
	if result.Content != "success reply" {
		t.Errorf("content = %q, want %q", result.Content, "success reply")
	}
	// Both calls reach wait() so we expect 2 records.
	if fr.recordCount() < 2 {
		t.Fatalf("expected 2 runner calls (initial + retry), got %d", fr.recordCount())
	}
	// Retry stdin must be the prompt with no history (broken-pipe turn is not committed).
	retryStdin := string(fr.record(1).stdin)
	want := "[prompt]\np\n"
	if retryStdin != want {
		t.Errorf("retry stdin = %q, want %q (history must be clean)", retryStdin, want)
	}
}

func TestSend_BrokenPipe_BothFail(t *testing.T) {
	fr := &fakeRunner{configs: []callConfig{
		{closeStdinEarly: true},
		{closeStdinEarly: true},
	}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	_, err := sess.Send(context.Background(), "p")
	if err == nil {
		t.Fatal("want error when both attempts fail with broken pipe")
	}
	cls := transport.Classify(err)
	if !cls.Retryable {
		t.Errorf("broken pipe must be retryable, got %+v", cls)
	}
}

// --- cancellation ---

func TestSend_CancelBeforeSpawn(t *testing.T) {
	fr := &fakeRunner{configs: []callConfig{{stdout: "r"}}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess := openSession(t, tr, transport.Spec{Model: "m"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before Send

	_, err := sess.Send(ctx, "p")
	if err == nil {
		t.Fatal("want error on pre-canceled context")
	}
	cls := transport.Classify(err)
	if cls.Retryable {
		t.Errorf("cancellation must not be retryable, got %+v", cls)
	}
	// Runner must not have been called (cancel before spawn).
	if fr.recordCount() != 0 {
		t.Errorf("runner must not be called on pre-canceled ctx, got %d calls", fr.recordCount())
	}
}

func TestSend_CancelDuringRun(t *testing.T) {
	runnerCalled := make(chan struct{})

	customRun := func(ctx context.Context, name string, args []string, dir string) (
		io.WriteCloser, io.ReadCloser, io.ReadCloser, func() error, error,
	) {
		// Signal the test that the runner has been invoked.
		select {
		case <-runnerCalled:
		default:
			close(runnerCalled)
		}

		stdinR, stdinW := io.Pipe()
		// Drain stdin eagerly so the write goroutine is not blocked.
		go func() { _, _ = io.ReadAll(stdinR) }()

		stdout := io.NopCloser(bytes.NewReader(nil))
		stderr := io.NopCloser(bytes.NewReader(nil))

		wait := func() error {
			<-ctx.Done()
			return ctx.Err()
		}

		return stdinW, stdout, stderr, wait, nil
	}

	tr, err := New(BackendAgy, noEnv, customRun)
	if err != nil {
		t.Fatal(err)
	}
	sess, err := tr.Open(context.Background(), transport.Spec{Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		_, sendErr := sess.Send(ctx, "hello")
		errCh <- sendErr
	}()

	<-runnerCalled
	cancel()

	sendErr := <-errCh
	if sendErr == nil {
		t.Fatal("want error on ctx cancellation during run")
	}
	cls := transport.Classify(sendErr)
	if cls.Retryable {
		t.Errorf("cancellation must not be retryable, got %+v", cls)
	}
}

// --- Close ---

func TestClose_Idempotent(t *testing.T) {
	tr, _ := New(BackendAgy, noEnv, (&fakeRunner{}).run)
	sess, err := tr.Open(context.Background(), transport.Spec{Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := sess.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestClose_Sealed_RemovesTempDir(t *testing.T) {
	fr := &fakeRunner{configs: []callConfig{{stdout: "r"}}}
	tr, _ := New(BackendAgy, noEnv, fr.run)
	sess, err := tr.Open(context.Background(), transport.Spec{Model: "m", ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	// Get the sealed dir before closing.
	esess := sess.(*execSession)
	sealedDir := esess.sealedDir
	if sealedDir == "" {
		t.Fatal("sealed session must have a sealedDir")
	}
	if _, err := os.Stat(sealedDir); err != nil {
		t.Fatalf("sealedDir must exist before Close: %v", err)
	}

	if err := sess.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := os.Stat(sealedDir); !os.IsNotExist(err) {
		t.Errorf("sealedDir %q must be removed after Close", sealedDir)
	}
}

func TestClose_Grounded_NoTempRemoval(t *testing.T) {
	tr, _ := New(BackendAgy, noEnv, (&fakeRunner{}).run)
	sess, err := tr.Open(context.Background(), transport.Spec{Model: "m", ReadOnly: false})
	if err != nil {
		t.Fatal(err)
	}
	esess := sess.(*execSession)
	if esess.sealedDir != "" {
		t.Errorf("grounded session must not have a sealedDir, got %q", esess.sealedDir)
	}
	if err := sess.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// --- buildStdin unit tests ---

func TestBuildStdin_EmptyHistory(t *testing.T) {
	got := buildStdin("", nil, "hello")
	want := "[prompt]\nhello\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildStdin_WithSystem(t *testing.T) {
	got := buildStdin("you are alice", nil, "hi")
	want := "[system]\nyou are alice\n\n[prompt]\nhi\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildStdin_WithHistory(t *testing.T) {
	history := []turn{
		{prompt: "p1", reply: "r1"},
		{prompt: "p2", reply: "r2"},
	}
	got := buildStdin("sys", history, "p3")
	want := "[system]\nsys\n\n" +
		"[prompt]\np1\n\n[reply]\nr1\n\n" +
		"[prompt]\np2\n\n[reply]\nr2\n\n" +
		"[prompt]\np3\n"
	if string(got) != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestBuildStdin_ContentAlreadyHasTrailingNewline(t *testing.T) {
	got := buildStdin("sys\n", []turn{{prompt: "p1\n", reply: "r1\n"}}, "p2\n")
	// Each block should not double the trailing newline.
	if bytes.Count(got, []byte("\n\n\n")) > 0 {
		t.Errorf("triple newline found in output: %q", got)
	}
	want := "[system]\nsys\n\n[prompt]\np1\n\n[reply]\nr1\n\n[prompt]\np2\n"
	if string(got) != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}
