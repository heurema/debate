// Package runner wires the debate workspace into an orchestrated run and synthesizes the result.
package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/heurema/debate/internal/debate/config"
	"github.com/heurema/debate/internal/debate/prompt"
	"github.com/heurema/debate/internal/debate/verdict"
	"github.com/heurema/debate/internal/engine/loop"
	"github.com/heurema/debate/internal/engine/orchestrate"
	"github.com/heurema/debate/internal/engine/transport"
)

// Resolver maps a backend identifier to a transport.Transport.
// Return an error for unimplemented or unknown backends.
type Resolver func(backend string) (transport.Transport, error)

// Config is the full input to Run.
type Config struct {
	WorkDir       string                 // start dir for .heurema/debate discovery
	TableName     string                 // --table: selected table name
	WithList      []string               // --with: panel selector override
	SynthOverride string                 // --synth: synthesizer override
	Task          string                 // assembled task text (must be non-empty)
	MaxRounds     int                    // max debate rounds; < 1 defaults to 10
	Sealed        bool                   // sets ReadOnly on all transport.Spec values
	OnTurn        func(orchestrate.Turn) // optional live callback per turn
	Resolver      Resolver
}

// Result is the output of a successful Run.
type Result struct {
	Answer  string
	Outcome loop.Outcome
	Turns   []orchestrate.Turn
}

// settleDefault and patienceDefault are built-in loop tuning values not exposed as flags.
const (
	settleDefault   = 2
	patienceDefault = 3
)

// Run validates inputs fail-fast, loads the workspace, runs the debate loop,
// then invokes the synthesizer once to produce the final answer.
// Sessions are opened after validation; all opened sessions are closed before return.
func Run(ctx context.Context, cfg Config) (Result, error) {
	if strings.TrimSpace(cfg.Task) == "" {
		return Result{}, fmt.Errorf("task is empty")
	}

	maxRounds := cfg.MaxRounds
	if maxRounds < 1 {
		maxRounds = 10
	}

	ws, err := config.Load(cfg.WorkDir, cfg.TableName, cfg.WithList, cfg.SynthOverride)
	if err != nil {
		return Result{}, err
	}
	if len(ws.Panel) == 0 {
		return Result{}, fmt.Errorf("panel is empty")
	}

	brief := cfg.Task

	// Open one session per panel participant.
	var sessions []transport.Session
	defer func() {
		for _, s := range sessions {
			s.Close()
		}
	}()

	participants := make([]orchestrate.Participant, 0, len(ws.Panel))
	for _, p := range ws.Panel {
		tr, err := cfg.Resolver(p.Backend)
		if err != nil {
			return Result{}, fmt.Errorf("backend %q: %w", p.Backend, err)
		}
		spec := transport.Spec{
			ID:       p.ID,
			Model:    p.Model,
			Effort:   p.Effort,
			System:   p.System,
			ReadOnly: cfg.Sealed,
		}
		sess, err := tr.Open(ctx, spec)
		if err != nil {
			return Result{}, fmt.Errorf("open session for %q: %w", p.ID, err)
		}
		sessions = append(sessions, sess)
		participants = append(participants, orchestrate.Participant{ID: p.ID, Session: sess})
	}

	orcCfg := orchestrate.Config{
		Participants: participants,
		Scheduler:    orchestrate.RoundRobin(false),
		Prompt:       prompt.NewPromptBuilder(brief),
		Verdict:      verdict.New(verdict.AllDone),
		Limits: loop.Limits{
			Max:      maxRounds,
			Settle:   settleDefault,
			Patience: patienceDefault,
		},
		OnTurn: cfg.OnTurn,
	}

	res, err := orchestrate.Run(ctx, orcCfg)
	if err != nil {
		return Result{}, fmt.Errorf("debate: %w", err)
	}

	answer, err := synthesize(ctx, cfg, ws, res.Transcript)
	if err != nil {
		return Result{}, fmt.Errorf("synthesize: %w", err)
	}

	return Result{
		Answer:  answer,
		Outcome: res.Outcome,
		Turns:   res.Transcript.All(),
	}, nil
}

func synthesize(ctx context.Context, cfg Config, ws config.Workspace, tr *orchestrate.Transcript) (string, error) {
	synth := ws.Synthesizer
	backend, err := cfg.Resolver(synth.Backend)
	if err != nil {
		return "", fmt.Errorf("synthesizer backend %q: %w", synth.Backend, err)
	}
	spec := transport.Spec{
		ID:       synth.ID,
		Model:    synth.Model,
		Effort:   synth.Effort,
		System:   synth.System,
		ReadOnly: cfg.Sealed,
	}
	sess, err := backend.Open(ctx, spec)
	if err != nil {
		return "", fmt.Errorf("open synthesizer session: %w", err)
	}
	defer sess.Close()

	result, err := sess.Send(ctx, buildSynthPrompt(cfg.Task, tr))
	if err != nil {
		return "", fmt.Errorf("synthesizer send: %w", err)
	}
	return result.Content, nil
}

func buildSynthPrompt(task string, tr *orchestrate.Transcript) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Task: %s\n\nDebate transcript:", task)
	for _, t := range tr.All() {
		fmt.Fprintf(&sb, "\n\n[Round %d — %s]\n%s", t.Round, t.Speaker, t.Content)
	}
	sb.WriteString("\n\nSynthesize the debate: summarize areas of agreement, unresolved objections, and provide a proposed resolution.")
	return sb.String()
}
