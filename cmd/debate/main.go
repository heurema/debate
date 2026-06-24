package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/heurema/debate/internal/debate/runner"
	"github.com/heurema/debate/internal/engine/orchestrate"
	"github.com/heurema/debate/internal/engine/transport"
	"github.com/heurema/debate/internal/engine/transport/echo"
)

// Version is set at build time via -ldflags; defaults to dev build.
var Version = "dev"

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "version" {
		fmt.Println("debate", Version)
		os.Exit(0)
	}

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: could not get working directory:", err)
		os.Exit(1)
	}

	isTerminal := stderrIsTerminal()
	code := parseAndRun(os.Args[1:], os.Stdout, os.Stderr, os.Stdin, isTerminal, os.Getenv, defaultResolver, workDir)
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
// Only echo is implemented in this slice; real backends are pending the acp slice.
func defaultResolver(backend string) (transport.Transport, error) {
	switch backend {
	case "echo":
		return echo.New(), nil
	case "claude-agent-acp", "codex-acp", "agy":
		return nil, fmt.Errorf("backend %q is not yet implemented (pending acp slice)", backend)
	default:
		return nil, fmt.Errorf("unknown backend %q", backend)
	}
}

// stringSlice implements flag.Value for repeatable --with flags.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
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
	var (
		with          stringSlice
		synthOverride string
		taskFlag      string
		maxRounds     int
		jsonOut       bool
		quiet         bool
		sealed        bool
	)

	fs := flag.NewFlagSet("debate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: debate [flags] <task>")
		fmt.Fprintln(stderr, "       debate version")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "flags:")
		fs.PrintDefaults()
	}
	fs.Var(&with, "with", "persona id to include in panel (repeatable)")
	fs.StringVar(&synthOverride, "synth", "", "override synthesizer persona id")
	fs.StringVar(&taskFlag, "task", "", "task text or @file (reads file content when prefixed with @)")
	fs.IntVar(&maxRounds, "max-rounds", 10, "maximum debate rounds")
	fs.BoolVar(&jsonOut, "json", false, "write JSON result to stdout, suppress stderr trace")
	fs.BoolVar(&quiet, "q", false, "suppress stderr debate trace")
	fs.BoolVar(&quiet, "quiet", false, "suppress stderr debate trace")
	fs.BoolVar(&sealed, "sealed", false, "read-only intent threaded into transport specs")

	if err := fs.Parse(args); err != nil {
		// flag.ContinueOnError already printed the error and usage.
		return 1
	}

	task, err := assembleTask(taskFlag, fs.Args(), stdin)
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
	traceEnabled := !quiet && !jsonOut && (isTerminal || forceTrace)

	var onTurn func(orchestrate.Turn)
	if traceEnabled {
		onTurn = func(t orchestrate.Turn) {
			fmt.Fprintf(stderr, "[Round %d — %s]\n%s\n\n", t.Round, t.Speaker, t.Content)
		}
	}

	cfg := runner.Config{
		WorkDir:       workDir,
		WithList:      []string(with),
		SynthOverride: synthOverride,
		Task:          task,
		MaxRounds:     maxRounds,
		Sealed:        sealed,
		OnTurn:        onTurn,
		Resolver:      resolver,
	}

	result, err := runner.Run(context.Background(), cfg)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	if jsonOut {
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
