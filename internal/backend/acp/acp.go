// Package acp provides a transport.Transport backed by ACP adapter subprocesses.
package acp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	acpsdk "github.com/coder/acp-go-sdk"
	"github.com/heurema/debate/internal/engine/transport"
)

// Backend identifiers this package handles.
const (
	BackendClaude = "claude-agent-acp"
	BackendCodex  = "codex-acp"
)

// Env var names for overriding the npm package for each adapter.
const (
	EnvClaudePackage = "DEBATE_CLAUDE_AGENT_ACP_PACKAGE"
	EnvCodexPackage  = "DEBATE_CODEX_ACP_PACKAGE"
	EnvOpenTimeout   = "DEBATE_ACP_OPEN_TIMEOUT"
	EnvSendTimeout   = "DEBATE_ACP_SEND_TIMEOUT"
)

const (
	defaultClaudePackage = "@agentclientprotocol/claude-agent-acp@latest"
	defaultCodexPackage  = "@heurema/codex-acp@latest"
	defaultOpenTimeout   = 45 * time.Second
	defaultSendTimeout   = 5 * time.Minute
)

// ProcessRunner starts a subprocess and returns (stdin, stdout, kill, err).
// stdin is the pipe the client writes to (agent reads from), stdout is what the client reads (agent writes to).
// kill terminates the subprocess and its process group.
// dir is the working directory for the subprocess.
type ProcessRunner func(dir, name string, args, env []string) (stdin io.WriteCloser, stdout io.ReadCloser, kill func() error, err error)

// acpTransport is a transport.Transport that opens ACP adapter subprocesses.
type acpTransport struct {
	backendID string
	getEnv    func(string) string
	run       ProcessRunner
}

// New returns a transport.Transport for the given backend ID.
// getEnv is used to read environment variables (pass nil to use os.Getenv).
// run is the process runner (pass nil for the default which spawns real subprocesses).
func New(backendID string, getEnv func(string) string, run ProcessRunner) (transport.Transport, error) {
	switch backendID {
	case BackendClaude, BackendCodex:
	default:
		return nil, fmt.Errorf("acp: unknown backend %q", backendID)
	}
	if getEnv == nil {
		getEnv = os.Getenv
	}
	if run == nil {
		run = defaultProcessRunner
	}
	return &acpTransport{backendID: backendID, getEnv: getEnv, run: run}, nil
}

// Open opens an ACP session for the given spec.
// spec.Model must be non-empty.
func (t *acpTransport) Open(ctx context.Context, spec transport.Spec) (transport.Session, error) {
	if spec.Model == "" {
		return nil, fmt.Errorf("acp: spec.Model must be non-empty")
	}
	cwd, err := resolveCwd(spec.ReadOnly)
	if err != nil {
		return nil, fmt.Errorf("acp: resolve cwd: %w", transport.ErrClientError)
	}
	return t.openAt(ctx, spec, cwd)
}

// openAt opens a session pinned to a specific cwd. Used by Open and during recovery.
func (t *acpTransport) openAt(ctx context.Context, spec transport.Spec, cwd string) (*acpSession, error) {
	cmd, env := t.buildCmd(spec)

	stdinW, stdoutR, kill, err := t.run(cwd, cmd[0], cmd[1:], env)
	if err != nil {
		return nil, fmt.Errorf("acp: spawn %q: %w", t.backendID, transport.ErrClientError)
	}

	openCtx, cancel := withOptionalTimeout(ctx, timeoutFromEnv(t.getEnv, EnvOpenTimeout, defaultOpenTimeout))
	defer cancel()

	cl := &clientImpl{}
	conn := acpsdk.NewClientSideConnection(cl, stdinW, stdoutR)

	if _, err := conn.Initialize(openCtx, acpsdk.InitializeRequest{
		ProtocolVersion: acpsdk.ProtocolVersionNumber,
	}); err != nil {
		_ = kill()
		return nil, fmt.Errorf("acp: initialize %q: %w", t.backendID, classifyConnErr(err))
	}

	sessResp, err := conn.NewSession(openCtx, acpsdk.NewSessionRequest{
		Cwd:        cwd,
		McpServers: []acpsdk.McpServer{},
	})
	if err != nil {
		_ = kill()
		return nil, fmt.Errorf("acp: new session %q: %w", t.backendID, classifyConnErr(err))
	}

	return &acpSession{
		tr:        t,
		spec:      spec,
		cwd:       cwd,
		conn:      conn,
		cl:        cl,
		kill:      kill,
		sessionID: sessResp.SessionId,
	}, nil
}

// buildCmd returns the command args and environment for the given backend and spec.
// claude-agent-acp: npx -y <pkg>  with ANTHROPIC_MODEL and CLAUDE_CODE_EFFORT_LEVEL
// codex-acp:        npx -y <pkg> -c model=<model> -c sandbox_mode=read-only
func (t *acpTransport) buildCmd(spec transport.Spec) (cmd []string, env []string) {
	base := os.Environ()
	switch t.backendID {
	case BackendClaude:
		pkg := t.getEnv(EnvClaudePackage)
		if pkg == "" {
			pkg = defaultClaudePackage
		}
		cmd = []string{"npx", "-y", pkg}
		env = append(base,
			"ANTHROPIC_MODEL="+spec.Model,
			"CLAUDE_CODE_EFFORT_LEVEL="+spec.Effort,
		)
	case BackendCodex:
		pkg := t.getEnv(EnvCodexPackage)
		if pkg == "" {
			pkg = defaultCodexPackage
		}
		// Codex effort is intentionally not wired; codex-acp exposes no effort knob.
		cmd = []string{"npx", "-y", pkg, "-c", "model=" + spec.Model, "-c", "sandbox_mode=read-only"}
		env = base
	}
	return
}

// resolveCwd returns the working directory for the ACP session and subprocess.
// Grounded (readOnly=false): uses the current process working directory so the agent reads project files.
// Sealed (readOnly=true): creates a fresh empty temp directory so the agent sees no project files.
func resolveCwd(sealed bool) (string, error) {
	if !sealed {
		return os.Getwd()
	}
	return os.MkdirTemp("", "debate-sealed-*")
}

// acpSession implements transport.Session over a persistent ACP adapter subprocess.
type acpSession struct {
	tr        *acpTransport
	spec      transport.Spec
	cwd       string // resolved once in Open; reused on recovery
	conn      *acpsdk.ClientSideConnection
	cl        *clientImpl
	kill      func() error
	sessionID acpsdk.SessionId
	history   []string // prompts sent this session, kept for replay on recovery

	mu         sync.Mutex
	closed     bool
	systemSent bool // true once spec.System has been injected into this ACP session
}

// Send sends a prompt to the adapter and returns the accumulated streamed response.
// On a retryable failure it recovers exactly once: closes the broken session,
// reopens (Initialize/NewSession), replays prior prompts, then retries.
func (s *acpSession) Send(ctx context.Context, prompt string) (transport.Result, error) {
	result, err := s.sendOnce(ctx, prompt)
	if err == nil {
		s.mu.Lock()
		s.history = append(s.history, prompt)
		s.mu.Unlock()
		return result, nil
	}

	cls := transport.Classify(err)
	if cls.Kind == "idle_timeout" || cls.Kind == "deadline" {
		s.closeInternal()
		return transport.Result{}, err
	}
	if !cls.Retryable {
		return transport.Result{}, err
	}

	// Recovery: reopen and replay.
	if recErr := s.recover(ctx); recErr != nil {
		return transport.Result{}, recErr
	}

	result, err = s.sendOnce(ctx, prompt)
	if err != nil {
		return transport.Result{}, err
	}
	s.mu.Lock()
	s.history = append(s.history, prompt)
	s.mu.Unlock()
	return result, nil
}

// sendOnce sends prompt with no retry logic.
func (s *acpSession) sendOnce(ctx context.Context, prompt string) (transport.Result, error) {
	s.cl.reset()

	sendCtx, cancel := withOptionalTimeout(ctx, timeoutFromEnv(s.tr.getEnv, EnvSendTimeout, defaultSendTimeout))
	defer cancel()

	s.mu.Lock()
	systemSent := s.systemSent
	s.mu.Unlock()

	// On the first send of a session (including after recovery), prepend the persona
	// system prompt so the adapter knows the participant's role and identity.
	content := []acpsdk.ContentBlock{acpsdk.TextBlock(prompt)}
	if s.spec.System != "" && !systemSent {
		content = []acpsdk.ContentBlock{acpsdk.TextBlock(s.spec.System), acpsdk.TextBlock(prompt)}
	}

	resp, err := s.conn.Prompt(sendCtx, acpsdk.PromptRequest{
		SessionId: s.sessionID,
		Prompt:    content,
	})
	if err != nil {
		return transport.Result{}, fmt.Errorf("acp: prompt: %w", classifyConnErr(err))
	}
	if resp.StopReason != acpsdk.StopReasonEndTurn {
		return transport.Result{}, fmt.Errorf("acp: stop_reason=%q: %w", resp.StopReason, stopReasonErr(resp.StopReason))
	}

	s.mu.Lock()
	s.systemSent = true
	s.mu.Unlock()

	return transport.Result{Content: s.cl.text()}, nil
}

// recover closes the broken session, reopens, and replays prior prompts.
// Any replay failure aborts recovery with a classified error.
func (s *acpSession) recover(ctx context.Context) error {
	s.mu.Lock()
	history := make([]string, len(s.history))
	copy(history, s.history)
	s.mu.Unlock()

	s.closeInternal()

	newSess, err := s.tr.openAt(ctx, s.spec, s.cwd)
	if err != nil {
		return fmt.Errorf("acp: recovery reopen: %w", err)
	}

	s.mu.Lock()
	s.conn = newSess.conn
	s.cl = newSess.cl
	s.kill = newSess.kill
	s.sessionID = newSess.sessionID
	s.closed = false
	s.systemSent = false // re-inject system prompt on first send of the recovered session
	s.mu.Unlock()

	for _, p := range history {
		if _, err := s.sendOnce(ctx, p); err != nil {
			return fmt.Errorf("acp: recovery replay: %w", err)
		}
	}
	return nil
}

// closeInternal terminates the subprocess without setting s.closed.
func (s *acpSession) closeInternal() {
	s.mu.Lock()
	kill := s.kill
	s.mu.Unlock()
	if kill != nil {
		_ = kill()
	}
}

// Close terminates the subprocess and its process group. Idempotent.
func (s *acpSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.kill != nil {
		return s.kill()
	}
	return nil
}

// clientImpl implements acpsdk.Client for our transport.
// It accumulates agent message text chunks during a prompt turn.
type clientImpl struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (c *clientImpl) reset() {
	c.mu.Lock()
	c.buf.Reset()
	c.mu.Unlock()
}

func (c *clientImpl) text() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}

func (c *clientImpl) SessionUpdate(_ context.Context, n acpsdk.SessionNotification) error {
	if n.Update.AgentMessageChunk != nil && n.Update.AgentMessageChunk.Content.Text != nil {
		c.mu.Lock()
		c.buf.WriteString(n.Update.AgentMessageChunk.Content.Text.Text)
		c.mu.Unlock()
	}
	return nil
}

func (c *clientImpl) ReadTextFile(_ context.Context, p acpsdk.ReadTextFileRequest) (acpsdk.ReadTextFileResponse, error) {
	data, err := os.ReadFile(p.Path)
	if err != nil {
		return acpsdk.ReadTextFileResponse{}, err
	}
	return acpsdk.ReadTextFileResponse{Content: string(data)}, nil
}

// WriteTextFile denies all writes; the agent must not write files.
func (c *clientImpl) WriteTextFile(_ context.Context, _ acpsdk.WriteTextFileRequest) (acpsdk.WriteTextFileResponse, error) {
	return acpsdk.WriteTextFileResponse{}, fmt.Errorf("acp: write denied: transport is read-only")
}

func (c *clientImpl) RequestPermission(_ context.Context, p acpsdk.RequestPermissionRequest) (acpsdk.RequestPermissionResponse, error) {
	if len(p.Options) == 0 {
		return acpsdk.RequestPermissionResponse{
			Outcome: acpsdk.RequestPermissionOutcome{
				Cancelled: &acpsdk.RequestPermissionOutcomeCancelled{Outcome: "cancelled"},
			},
		}, nil
	}
	// Allow the first option (typically "allow once") to keep the adapter running.
	return acpsdk.RequestPermissionResponse{
		Outcome: acpsdk.RequestPermissionOutcome{
			Selected: &acpsdk.RequestPermissionOutcomeSelected{
				OptionId: p.Options[0].OptionId,
				Outcome:  "selected",
			},
		},
	}, nil
}

// Terminal stubs — the debate transport does not use interactive terminals.
func (c *clientImpl) CreateTerminal(_ context.Context, _ acpsdk.CreateTerminalRequest) (acpsdk.CreateTerminalResponse, error) {
	return acpsdk.CreateTerminalResponse{}, fmt.Errorf("acp: terminal not supported")
}
func (c *clientImpl) KillTerminal(_ context.Context, _ acpsdk.KillTerminalRequest) (acpsdk.KillTerminalResponse, error) {
	return acpsdk.KillTerminalResponse{}, fmt.Errorf("acp: terminal not supported")
}
func (c *clientImpl) TerminalOutput(_ context.Context, _ acpsdk.TerminalOutputRequest) (acpsdk.TerminalOutputResponse, error) {
	return acpsdk.TerminalOutputResponse{}, fmt.Errorf("acp: terminal not supported")
}
func (c *clientImpl) ReleaseTerminal(_ context.Context, _ acpsdk.ReleaseTerminalRequest) (acpsdk.ReleaseTerminalResponse, error) {
	return acpsdk.ReleaseTerminalResponse{}, fmt.Errorf("acp: terminal not supported")
}
func (c *clientImpl) WaitForTerminalExit(_ context.Context, _ acpsdk.WaitForTerminalExitRequest) (acpsdk.WaitForTerminalExitResponse, error) {
	return acpsdk.WaitForTerminalExitResponse{}, fmt.Errorf("acp: terminal not supported")
}

// classifyConnErr maps ACP connection/protocol errors to transport sentinel errors.
// Dropped connection (InternalError, code -32603) → ErrTransportDrop (retryable).
// Auth required (code -32000) → ErrAuth.
// Request cancelled (code -32800) → ErrCanceled.
// Other protocol errors → ErrClientError.
func classifyConnErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %s", transport.ErrIdleTimeout, err.Error())
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%w: %s", transport.ErrCanceled, err.Error())
	}
	var reqErr *acpsdk.RequestError
	if errors.As(err, &reqErr) {
		reqMsg := reqErr.Error()
		if errors.Is(reqErr, context.DeadlineExceeded) || strings.Contains(reqMsg, "context deadline exceeded") {
			return fmt.Errorf("%w: %s", transport.ErrIdleTimeout, reqMsg)
		}
		if errors.Is(reqErr, context.Canceled) || strings.Contains(reqMsg, "context canceled") {
			return fmt.Errorf("%w: %s", transport.ErrCanceled, reqMsg)
		}
		switch reqErr.Code {
		case -32603: // InternalError — wraps dropped connection and peer disconnect
			return fmt.Errorf("%w: %s", transport.ErrTransportDrop, reqMsg)
		case -32000: // AuthRequired
			return fmt.Errorf("%w: %s", transport.ErrAuth, reqMsg)
		case -32800: // RequestCancelled
			return fmt.Errorf("%w: %s", transport.ErrCanceled, reqMsg)
		default:
			return fmt.Errorf("%w: %s", transport.ErrClientError, reqMsg)
		}
	}
	// IO-level errors that escape the SDK wrapping.
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, io.ErrUnexpectedEOF) {
		return fmt.Errorf("%w: %s", transport.ErrTransportDrop, err.Error())
	}
	s := err.Error()
	if strings.Contains(s, "context deadline exceeded") {
		return fmt.Errorf("%w: %s", transport.ErrIdleTimeout, s)
	}
	if strings.Contains(s, "context canceled") {
		return fmt.Errorf("%w: %s", transport.ErrCanceled, s)
	}
	if strings.Contains(s, "broken pipe") || strings.Contains(s, "connection reset") {
		return fmt.Errorf("%w: %s", transport.ErrTransportDrop, s)
	}
	return fmt.Errorf("%w: %s", transport.ErrClientError, s)
}

func timeoutFromEnv(getEnv func(string) string, key string, def time.Duration) time.Duration {
	raw := strings.TrimSpace(getEnv(key))
	if raw == "" {
		return def
	}
	switch strings.ToLower(raw) {
	case "0", "off", "none", "disabled":
		return 0
	}
	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d
	}
	if secs, err := strconv.Atoi(raw); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return def
}

func withOptionalTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

// stopReasonErr maps a non-end_turn stop reason to a transport sentinel error.
func stopReasonErr(reason acpsdk.StopReason) error {
	switch reason {
	case acpsdk.StopReasonCancelled:
		return transport.ErrCanceled
	default:
		// refusal, max_tokens, max_turn_requests → terminal
		return transport.ErrClientError
	}
}

// defaultProcessRunner spawns name with args in a new process group (Unix only).
// kill sends SIGKILL to the entire process group.
func defaultProcessRunner(dir, name string, args, env []string) (io.WriteCloser, io.ReadCloser, func() error, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdinW, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stdoutR, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, nil, err
	}

	kill := func() error {
		if cmd.Process == nil {
			return nil
		}
		// Kill the entire process group.
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	return stdinW, stdoutR, kill, nil
}
