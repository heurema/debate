// Package exec provides a transport.Transport that drives stateless CLI agents.
// Each Send spawns a fresh subprocess and reconstructs the full conversation
// context (system, prior prompts and replies, current prompt) as stdin.
package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/heurema/debate/internal/engine/transport"
)

// BackendAgy is the backend identifier for agy.
const BackendAgy = "agy"

// EnvAgyCommand overrides the agy executable (argv[0]); the model flag is preserved.
const EnvAgyCommand = "DEBATE_AGY_COMMAND"

// maxStderrBytes is the stderr capture limit for non-zero exit diagnostics.
const maxStderrBytes = 4 * 1024

// CommandRunner spawns a subprocess and returns I/O handles.
// wait blocks until the subprocess exits and returns its exit error (nil on zero exit).
// Injected in tests to avoid invoking a real external program.
type CommandRunner func(ctx context.Context, name string, args []string, dir string) (
	stdin io.WriteCloser,
	stdout io.ReadCloser,
	stderr io.ReadCloser,
	wait func() error,
	err error,
)

type execTransport struct {
	getEnv func(string) string
	run    CommandRunner
}

// New returns a transport.Transport for the agy backend.
// getEnv is used to read environment variables (pass nil to use os.Getenv).
// run is the command runner (pass nil for the default real-subprocess runner).
func New(backendID string, getEnv func(string) string, run CommandRunner) (transport.Transport, error) {
	if backendID != BackendAgy {
		return nil, fmt.Errorf("exec: unknown backend %q", backendID)
	}
	if getEnv == nil {
		getEnv = os.Getenv
	}
	if run == nil {
		run = defaultCommandRunner
	}
	return &execTransport{getEnv: getEnv, run: run}, nil
}

// Open opens a session for the given spec. No subprocess is spawned here; it is
// spawned per Send because the CLI is stateless.
func (t *execTransport) Open(_ context.Context, spec transport.Spec) (transport.Session, error) {
	if spec.Model == "" {
		return nil, fmt.Errorf("exec: spec.Model must be non-empty")
	}
	cwd, sealedDir, err := resolveCwd(spec.ReadOnly)
	if err != nil {
		return nil, fmt.Errorf("exec: resolve cwd: %w", err)
	}
	return &execSession{tr: t, spec: spec, cwd: cwd, sealedDir: sealedDir}, nil
}

// resolveCmd returns the argv for the subprocess.
// Default: [agy, --print, --model, spec.Model]. EnvAgyCommand overrides argv[0] only.
// --print makes agy run a single prompt non-interactively (reads from stdin, prints response, exits).
func (t *execTransport) resolveCmd(spec transport.Spec) []string {
	name := "agy"
	if override := t.getEnv(EnvAgyCommand); override != "" {
		name = override
	}
	return []string{name, "--print", "--model", spec.Model}
}

// resolveCwd returns the working directory for the subprocess.
// Grounded (sealed=false): uses the process cwd.
// Sealed (sealed=true): creates a fresh empty temp dir; sealedDir is non-empty
// and the caller is responsible for removing it on Close.
//
// Note: filesystem read-only is NOT enforced in either mode. A plain CLI has no
// sandbox, so read-only is trusted/best-effort — the subprocess may write anywhere
// it has OS permission. Network is available in both grounded and sealed modes.
func resolveCwd(sealed bool) (cwd, sealedDir string, err error) {
	if !sealed {
		cwd, err = os.Getwd()
		return
	}
	sealedDir, err = os.MkdirTemp("", "debate-exec-sealed-*")
	cwd = sealedDir
	return
}

// turn is one committed prompt/reply pair from a successful Send.
type turn struct {
	prompt string
	reply  string
}

type execSession struct {
	tr        *execTransport
	spec      transport.Spec
	cwd       string
	sealedDir string // non-empty if sealed; removed on Close
	history   []turn // committed after each successful Send

	mu     sync.Mutex
	closed bool
}

// Send spawns a fresh subprocess, writes the reconstructed stdin (full context),
// and returns the response. On a retryable error, retries exactly once without
// committing to history, so a failed Send does not pollute subsequent Sends.
func (s *execSession) Send(ctx context.Context, prompt string) (transport.Result, error) {
	result, err := s.sendOnce(ctx, prompt)
	if err == nil {
		s.history = append(s.history, turn{prompt: prompt, reply: result.Content})
		return result, nil
	}

	if !transport.Classify(err).Retryable {
		return transport.Result{}, err
	}

	result, err = s.sendOnce(ctx, prompt)
	if err != nil {
		return transport.Result{}, err
	}
	s.history = append(s.history, turn{prompt: prompt, reply: result.Content})
	return result, nil
}

func (s *execSession) sendOnce(ctx context.Context, prompt string) (transport.Result, error) {
	if ctx.Err() != nil {
		return transport.Result{}, fmt.Errorf("%w: %s", transport.ErrCanceled, ctx.Err().Error())
	}

	argv := s.tr.resolveCmd(s.spec)
	stdinData := buildStdin(s.spec.System, s.history, prompt)

	stdin, stdout, stderr, wait, err := s.tr.run(ctx, argv[0], argv[1:], s.cwd)
	if err != nil {
		return transport.Result{}, fmt.Errorf("%w: spawn: %s", transport.ErrTransportDrop, err.Error())
	}

	// Write stdin in its own goroutine so it runs concurrently with stdout/stderr reads.
	stdinErrCh := make(chan error, 1)
	go func() {
		_, werr := stdin.Write(stdinData)
		cerr := stdin.Close()
		if werr != nil {
			stdinErrCh <- werr
		} else {
			stdinErrCh <- cerr
		}
	}()

	var (
		stdoutData []byte
		stderrBuf  []byte
	)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		stdoutData, _ = io.ReadAll(stdout)
	}()
	go func() {
		defer wg.Done()
		stderrBuf, _ = io.ReadAll(io.LimitReader(stderr, maxStderrBytes))
		// Drain the remainder so the subprocess is never blocked on stderr writes.
		_, _ = io.Copy(io.Discard, stderr)
	}()

	stdinErr := <-stdinErrCh
	wg.Wait()
	waitErr := wait()

	// Context cancellation takes precedence: a kill also causes a broken-pipe on stdin,
	// so we check ctx first to avoid misclassifying that as retryable.
	if ctx.Err() != nil {
		return transport.Result{}, fmt.Errorf("%w: %s", transport.ErrCanceled, ctx.Err().Error())
	}

	// Broken pipe on stdin write means the subprocess closed stdin early.
	// If it also exited zero with content, the process simply didn't need more
	// input — trust the result rather than discarding it. Otherwise treat it as
	// retryable so a fresh attempt can write the full context.
	if stdinErr != nil && isBrokenPipe(stdinErr) {
		if waitErr == nil && len(stdoutData) > 0 {
			return transport.Result{Content: string(stdoutData)}, nil
		}
		return transport.Result{}, fmt.Errorf("%w: stdin: %s", transport.ErrTransportDrop, stdinErr.Error())
	}

	// Non-zero exit is terminal; include captured stderr for diagnosis.
	if waitErr != nil {
		return transport.Result{}, fmt.Errorf("%w: %s", transport.ErrClientError, formatStderr(stderrBuf))
	}

	return transport.Result{Content: string(stdoutData)}, nil
}

// Close releases session resources. Idempotent.
// Removes the sealed temp directory if one was created in Open.
func (s *execSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.sealedDir != "" {
		return os.RemoveAll(s.sealedDir)
	}
	return nil
}

// buildStdin constructs the full stdin input for the subprocess.
//
// Format: system block (when non-empty), alternating [prompt]/[reply] pairs for
// each prior turn, then the current [prompt]. Each block is "[$label]\n$content\n";
// a trailing "\n" is appended to content if not already present. Consecutive blocks
// are separated by exactly one blank line ("\n"). The last block ends without a
// trailing blank line.
func buildStdin(system string, history []turn, currentPrompt string) []byte {
	type blk struct{ label, content string }
	var blocks []blk

	if system != "" {
		blocks = append(blocks, blk{"system", system})
	}
	for _, t := range history {
		blocks = append(blocks, blk{"prompt", t.prompt})
		blocks = append(blocks, blk{"reply", t.reply})
	}
	blocks = append(blocks, blk{"prompt", currentPrompt})

	var buf bytes.Buffer
	for i, b := range blocks {
		buf.WriteString("[" + b.label + "]\n")
		buf.WriteString(b.content)
		if len(b.content) == 0 || b.content[len(b.content)-1] != '\n' {
			buf.WriteByte('\n')
		}
		if i < len(blocks)-1 {
			buf.WriteByte('\n') // blank-line separator between blocks
		}
	}
	return buf.Bytes()
}

func isBrokenPipe(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.ErrClosedPipe) {
		return true
	}
	return strings.Contains(err.Error(), "broken pipe")
}

// formatStderr formats captured stderr for inclusion in a terminal error message.
func formatStderr(data []byte) string {
	if len(data) == 0 {
		return "(no stderr output)"
	}
	s := string(data)
	if len(data) >= maxStderrBytes {
		s += "...(truncated)"
	}
	return s
}

// defaultCommandRunner spawns name with args using exec.CommandContext.
// The subprocess's working directory is set to dir.
func defaultCommandRunner(ctx context.Context, name string, args []string, dir string) (
	io.WriteCloser, io.ReadCloser, io.ReadCloser, func() error, error,
) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	stdinW, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	stdoutR, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	stderrR, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, nil, nil, err
	}

	return stdinW, stdoutR, stderrR, cmd.Wait, nil
}
