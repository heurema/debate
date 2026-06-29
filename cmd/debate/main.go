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
	"github.com/heurema/debate/internal/debate/progress"
	"github.com/heurema/debate/internal/debate/runner"
	"github.com/heurema/debate/internal/engine/transport"
	"github.com/heurema/debate/internal/engine/transport/echo"
)

// Version is set at build time via -ldflags; defaults to dev build.
var Version = "dev"

func main() {
	code := parseCLI(os.Args[1:], os.Stdout, os.Stderr, os.Stdin, defaultResolver, "")
	os.Exit(code)
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
	resolver   runner.Resolver
	workDir    string
	getWorkDir func() (string, error)
	code       int
}

type runCmd struct {
	Table         string   `name:"table" help:"Table name to run (defaults to default)."`
	With          []string `name:"with" sep:"none" help:"Add debater persona selectors. Repeat or separate selectors with commas."`
	SynthOverride string   `name:"synth" help:"Override synthesizer persona id."`
	TaskFlag      string   `name:"task" help:"Task text, @file, or - for stdin."`
	MaxRounds     int      `name:"max-rounds" default:"10" help:"Maximum debate rounds."`
	JSONOut       bool     `name:"json" help:"Write JSON result to stdout."`
	Quiet         bool     `name:"quiet" short:"q" help:"Suppress stderr progress events."`
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
	deps.code = runDebate(c, deps.stdout, deps.stderr, deps.stdin, deps.resolver, workDir)
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
	resolver runner.Resolver,
	workDir string,
) int {
	if len(args) > 0 && isNamedCommand(args[0]) {
		var app cli
		return parseAndDispatch(&app, args, stdout, stderr, stdin, resolver, workDir)
	}
	return parseAndRun(args, stdout, stderr, stdin, resolver, workDir)
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
	resolver runner.Resolver,
	workDir string,
) int {
	deps := &cliDeps{
		stdout:     stdout,
		stderr:     stderr,
		stdin:      stdin,
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
	resolver runner.Resolver,
	workDir string,
) int {
	var cmd runCmd
	return parseAndDispatch(&cmd, args, stdout, stderr, stdin, resolver, workDir)
}

func validateWithSelectorValue(value string) error {
	for _, part := range strings.Split(value, ",") {
		if strings.TrimSpace(part) == "" {
			return fmt.Errorf("--with contains an empty persona selector; remove empty entries or pass one persona per --with")
		}
	}
	return nil
}

func runDebate(
	cmd *runCmd,
	stdout, stderr io.Writer,
	stdin io.Reader,
	resolver runner.Resolver,
	workDir string,
) int {
	withList, err := normalizeWithSelectors(cmd.With)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	task, err := assembleTask(cmd.TaskFlag, cmd.Task, stdin)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	if strings.TrimSpace(task) == "" {
		fmt.Fprintln(stderr, "error: task is empty; provide a task as a positional argument, --task value, or via stdin")
		return 1
	}

	var progressSink runner.Progress
	if !cmd.Quiet {
		progressSink = progress.NewEmitter(stderr)
	}

	cfg := runner.Config{
		WorkDir:       workDir,
		TableName:     cmd.Table,
		WithList:      withList,
		SynthOverride: cmd.SynthOverride,
		Task:          task,
		MaxRounds:     cmd.MaxRounds,
		Sealed:        cmd.Sealed,
		Resolver:      resolver,
		Progress:      progressSink,
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
			Outcome: result.Outcome.Reason,
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

func normalizeWithSelectors(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	selectors := make([]string, 0, len(values))
	for _, value := range values {
		if err := validateWithSelectorValue(value); err != nil {
			return nil, err
		}
		for _, part := range strings.Split(value, ",") {
			selector := strings.TrimSpace(part)
			selectors = append(selectors, selector)
		}
	}
	return selectors, nil
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
	stdinConsumed := false

	if taskFlag != "" {
		switch {
		case taskFlag == "-":
			stdinText, err := readIfPiped(stdin)
			if err != nil {
				return "", fmt.Errorf("stdin: %w", err)
			}
			stdinConsumed = true
			if stdinText != "" {
				parts = append(parts, stdinText)
			}
		case strings.HasPrefix(taskFlag, "@"):
			path := taskFlag[1:]
			data, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("--task @%s: %w", path, err)
			}
			parts = append(parts, strings.TrimRight(string(data), "\n"))
		default:
			parts = append(parts, taskFlag)
		}
	}

	if len(positional) > 0 {
		parts = append(parts, strings.Join(positional, " "))
	}

	if !stdinConsumed {
		if stdinText, err := readIfPiped(stdin); err != nil {
			return "", fmt.Errorf("stdin: %w", err)
		} else if stdinText != "" {
			parts = append(parts, stdinText)
		}
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
