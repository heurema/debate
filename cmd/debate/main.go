package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/heurema/debate/internal/backend/acp"
	"github.com/heurema/debate/internal/backend/exec"
	"github.com/heurema/debate/internal/debate/runner"
	"github.com/heurema/debate/internal/engine/orchestrate"
	"github.com/heurema/debate/internal/engine/transport"
	"github.com/heurema/debate/internal/engine/transport/echo"
)

// Version is set at build time via -ldflags; defaults to dev build.
var Version = "dev"

func main() {
	isTerminal := stderrIsTerminal()
	code := parseCLI(os.Args[1:], os.Stdout, os.Stderr, os.Stdin, isTerminal, os.Getenv, defaultResolver, "")
	os.Exit(code)
}

// stderrIsTerminal reports whether os.Stderr is a character device (i.e., a TTY).
func stderrIsTerminal() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// defaultResolver resolves backend identifiers to transports.
func defaultResolver(backend string) (transport.Transport, error) {
	switch backend {
	case "echo":
		return echo.New(), nil
	case acp.BackendClaude, acp.BackendCodex:
		return acp.New(backend, os.Getenv, nil)
	case exec.BackendAgy:
		return exec.New(backend, os.Getenv, nil)
	default:
		return nil, fmt.Errorf("unknown backend %q", backend)
	}
}

type cli struct {
	Version versionCmd `cmd:"" help:"Print the debate version."`
	Init    initCmd    `cmd:"" help:"Scaffold a .heurema/debate workspace in the current directory."`
	New     newCmd     `cmd:"" help:"Create a new persona file in the discovered .heurema/debate/personas."`
}

type cliDeps struct {
	stdout     io.Writer
	stderr     io.Writer
	stdin      io.Reader
	isTerminal bool
	getEnv     func(string) string
	resolver   runner.Resolver
	workDir    string
	getWorkDir func() (string, error)
	code       int
}

type runCmd struct {
	With          []string `name:"with" help:"Persona id to include in panel. Repeatable."`
	SynthOverride string   `name:"synth" help:"Override synthesizer persona id."`
	TaskFlag      string   `name:"task" help:"Task text or @file (reads file content when prefixed with @)."`
	MaxRounds     int      `name:"max-rounds" default:"10" help:"Maximum debate rounds."`
	JSONOut       bool     `name:"json" help:"Write JSON result to stdout, suppress stderr trace."`
	Quiet         bool     `name:"quiet" short:"q" help:"Suppress stderr debate trace."`
	Sealed        bool     `name:"sealed" help:"Read-only intent threaded into transport specs."`
	Task          []string `arg:"" optional:"" name:"task" help:"Task text."`
}

type versionCmd struct{}

func (c *runCmd) Run(deps *cliDeps) error {
	workDir, err := deps.resolveWorkDir()
	if err != nil {
		fmt.Fprintln(deps.stderr, "error: could not get working directory:", err)
		deps.code = 1
		return nil
	}
	deps.code = runDebate(c, deps.stdout, deps.stderr, deps.stdin, deps.isTerminal, deps.getEnv, deps.resolver, workDir)
	return nil
}

func (versionCmd) Run(deps *cliDeps) error {
	fmt.Fprintln(deps.stdout, "debate", Version)
	deps.code = 0
	return nil
}

func parseCLI(
	args []string,
	stdout, stderr io.Writer,
	stdin io.Reader,
	isTerminal bool,
	getEnv func(string) string,
	resolver runner.Resolver,
	workDir string,
) int {
	if len(args) > 0 && isNamedCommand(args[0]) {
		var app cli
		return parseAndDispatch(&app, args, stdout, stderr, stdin, isTerminal, getEnv, resolver, workDir)
	}
	return parseAndRun(args, stdout, stderr, stdin, isTerminal, getEnv, resolver, workDir)
}

func isNamedCommand(arg string) bool {
	switch arg {
	case "version", "init", "new":
		return true
	default:
		return false
	}
}

func parseAndDispatch(
	app any,
	args []string,
	stdout, stderr io.Writer,
	stdin io.Reader,
	isTerminal bool,
	getEnv func(string) string,
	resolver runner.Resolver,
	workDir string,
) int {
	deps := &cliDeps{
		stdout:     stdout,
		stderr:     stderr,
		stdin:      stdin,
		isTerminal: isTerminal,
		getEnv:     getEnv,
		resolver:   resolver,
		workDir:    workDir,
		getWorkDir: os.Getwd,
		code:       0,
	}
	parser, err := kong.New(app,
		kong.Name("debate"),
		kong.Writers(stderr, stderr),
		kong.ShortUsageOnError(),
		kong.Exit(func(code int) { panic(kongExit(code)) }),
	)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	var ctx *kong.Context
	if code, ok := catchKongExit(func() {
		ctx, err = parser.Parse(args)
	}); ok {
		return code
	}
	if err != nil {
		_, _ = catchKongExit(func() {
			parser.FatalIfErrorf(err)
		})
		return 1
	}
	if err := ctx.Run(deps); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return deps.code
}

type kongExit int

func catchKongExit(fn func()) (code int, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			exit, isExit := r.(kongExit)
			if !isExit {
				panic(r)
			}
			code = int(exit)
			ok = true
		}
	}()
	fn()
	return 0, false
}

func (d *cliDeps) resolveWorkDir() (string, error) {
	if d.workDir != "" {
		return d.workDir, nil
	}
	return d.getWorkDir()
}

// parseAndRun parses debate flags, assembles the task, runs the debate, and writes output.
// Returns the process exit code: 0 settled, 2 not-converged, 1 error.
func parseAndRun(
	args []string,
	stdout, stderr io.Writer,
	stdin io.Reader,
	isTerminal bool,
	getEnv func(string) string,
	resolver runner.Resolver,
	workDir string,
) int {
	var cmd runCmd
	return parseAndDispatch(&cmd, args, stdout, stderr, stdin, isTerminal, getEnv, resolver, workDir)
}

func runDebate(
	cmd *runCmd,
	stdout, stderr io.Writer,
	stdin io.Reader,
	isTerminal bool,
	getEnv func(string) string,
	resolver runner.Resolver,
	workDir string,
) int {
	task, err := assembleTask(cmd.TaskFlag, cmd.Task, stdin)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	if strings.TrimSpace(task) == "" {
		fmt.Fprintln(stderr, "error: task is empty; provide a task as a positional argument, --task value, or via stdin")
		return 1
	}

	// Trace to stderr when: not quiet, not JSON, and (it's a TTY or DEBATE_FORCE_TRACE=1).
	forceTrace := getEnv("DEBATE_FORCE_TRACE") == "1"
	traceEnabled := !cmd.Quiet && !cmd.JSONOut && (isTerminal || forceTrace)

	var onTurn func(orchestrate.Turn)
	if traceEnabled {
		onTurn = func(t orchestrate.Turn) {
			fmt.Fprintf(stderr, "[Round %d — %s]\n%s\n\n", t.Round, t.Speaker, t.Content)
		}
	}

	cfg := runner.Config{
		WorkDir:       workDir,
		WithList:      cmd.With,
		SynthOverride: cmd.SynthOverride,
		Task:          task,
		MaxRounds:     cmd.MaxRounds,
		Sealed:        cmd.Sealed,
		OnTurn:        onTurn,
		Resolver:      resolver,
	}

	result, err := runner.Run(context.Background(), cfg)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	if cmd.JSONOut {
		type jsonTurn struct {
			Round   int    `json:"round"`
			Speaker string `json:"speaker"`
			Content string `json:"content"`
		}
		type jsonResult struct {
			Answer  string     `json:"answer"`
			Outcome string     `json:"outcome"`
			Rounds  int        `json:"rounds"`
			Turns   []jsonTurn `json:"turns"`
		}

		turns := make([]jsonTurn, len(result.Turns))
		for i, t := range result.Turns {
			turns[i] = jsonTurn{Round: t.Round, Speaker: t.Speaker, Content: t.Content}
		}
		out := jsonResult{
			Answer:  result.Answer,
			Outcome: outcomeString(result.Outcome.Reason),
			Rounds:  result.Outcome.Rounds,
			Turns:   turns,
		}
		enc := json.NewEncoder(stdout)
		if err := enc.Encode(out); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
	} else {
		fmt.Fprintln(stdout, result.Answer)
	}

	return exitCode(result.Outcome.Reason)
}

// outcomeString normalises the loop reason to a user-facing string.
func outcomeString(reason string) string {
	switch reason {
	case "settled", "stalemate", "max":
		return reason
	default:
		return reason
	}
}

// exitCode maps loop outcome reasons to process exit codes.
// settled → 0 (converged), stalemate/max → 2 (did not converge), anything else → 1 (error).
func exitCode(reason string) int {
	switch reason {
	case "settled":
		return 0
	case "stalemate", "max":
		return 2
	default:
		return 1
	}
}

// assembleTask builds the task string from --task, positional args, and piped stdin.
// Stdin is always read when it is not an *os.File (e.g. in tests with a bytes.Buffer).
// When stdin is an *os.File, it is read only if it is not a character device (i.e. it is piped).
// Non-empty parts are joined with a newline.
func assembleTask(taskFlag string, positional []string, stdin io.Reader) (string, error) {
	var parts []string

	if taskFlag != "" {
		if strings.HasPrefix(taskFlag, "@") {
			path := taskFlag[1:]
			data, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("--task @%s: %w", path, err)
			}
			parts = append(parts, strings.TrimRight(string(data), "\n"))
		} else {
			parts = append(parts, taskFlag)
		}
	}

	if len(positional) > 0 {
		parts = append(parts, strings.Join(positional, " "))
	}

	if stdinText, err := readIfPiped(stdin); err != nil {
		return "", fmt.Errorf("stdin: %w", err)
	} else if stdinText != "" {
		parts = append(parts, stdinText)
	}

	return strings.Join(parts, "\n"), nil
}

// readIfPiped reads r when it is not a TTY.
// For *os.File, it checks ModeCharDevice; other readers are always read.
func readIfPiped(r io.Reader) (string, error) {
	if f, ok := r.(*os.File); ok {
		fi, err := f.Stat()
		if err != nil {
			return "", err
		}
		if fi.Mode()&os.ModeCharDevice != 0 {
			return "", nil // interactive TTY — do not read
		}
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\n"), nil
}
