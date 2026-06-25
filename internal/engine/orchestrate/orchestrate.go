// Package orchestrate coordinates loop participants and turn sequencing.
package orchestrate

import (
	"context"
	"fmt"

	"github.com/heurema/debate/internal/engine/loop"
	"github.com/heurema/debate/internal/engine/transport"
)

// RenderMode controls how the transcript is rendered for a participant.
type RenderMode int

const (
	Delta RenderMode = iota // only turns since the participant last spoke
	Full                    // entire committed transcript so far
)

// Participant is one speaker in the debate.
type Participant struct {
	ID      string
	Session transport.Session
}

// Turn is one recorded utterance.
type Turn struct {
	Round   int
	Speaker string
	Content string
	Usage   transport.Usage
	Extra   any
}

// Transcript accumulates turns and tracks per-participant read cursors.
// Not safe for concurrent use.
type Transcript struct {
	turns   []Turn
	cursors map[string]int // participantID -> index after their last appended turn
}

// Append adds a turn and advances that speaker's cursor to the current end.
func (t *Transcript) Append(turn Turn) {
	if t.cursors == nil {
		t.cursors = make(map[string]int)
	}
	t.turns = append(t.turns, turn)
	t.cursors[turn.Speaker] = len(t.turns)
}

// DeltaFor returns turns added by OTHER participants since participantID last spoke.
// Read-only: does not mutate any cursor.
func (t *Transcript) DeltaFor(participantID string) []Turn {
	cursor := t.cursors[participantID] // zero if participant has never spoken
	var delta []Turn
	for _, turn := range t.turns[cursor:] {
		if turn.Speaker != participantID {
			delta = append(delta, turn)
		}
	}
	return delta
}

// All returns a copy of all turns in order.
func (t *Transcript) All() []Turn {
	out := make([]Turn, len(t.turns))
	copy(out, t.turns)
	return out
}

// Len returns the number of turns in the transcript.
func (t *Transcript) Len() int { return len(t.turns) }

// PromptBuilder builds the prompt for a participant's turn.
type PromptBuilder func(p Participant, t *Transcript, rc loop.RoundContext, m RenderMode) (string, error)

// Verdict decides whether a round was clean, made progress, or should stop.
type Verdict interface {
	Assess(t *Transcript, rc loop.RoundContext) loop.RoundResult
}

// Scheduler determines speaking order for a round.
type Scheduler interface {
	Order(rc loop.RoundContext, ps []Participant) []Participant
}

// Config is the full input to orchestrate.Run.
type Config struct {
	Participants []Participant
	Scheduler    Scheduler
	Prompt       PromptBuilder
	Verdict      Verdict
	Limits       loop.Limits
	OnTurn       func(Turn) // optional live callback
}

// Result is the output of a completed run.
type Result struct {
	Transcript *Transcript
	Outcome    loop.Outcome
}

// Run drives a full multi-participant debate loop.
// It requires at least one Participant and non-nil Scheduler, Prompt, and Verdict.
// Caller owns session lifecycle; Run neither opens nor closes sessions.
func Run(ctx context.Context, cfg Config) (Result, error) {
	if len(cfg.Participants) == 0 {
		return Result{}, fmt.Errorf("orchestrate: at least one Participant required")
	}
	if cfg.Scheduler == nil {
		return Result{}, fmt.Errorf("orchestrate: Scheduler must not be nil")
	}
	if cfg.Prompt == nil {
		return Result{}, fmt.Errorf("orchestrate: Prompt must not be nil")
	}
	if cfg.Verdict == nil {
		return Result{}, fmt.Errorf("orchestrate: Verdict must not be nil")
	}

	tr := &Transcript{}

	step := func(ctx context.Context, rc loop.RoundContext) (loop.RoundResult, error) {
		speakers := cfg.Scheduler.Order(rc, cfg.Participants)
		for _, p := range speakers {
			prompt, err := cfg.Prompt(p, tr, rc, Full)
			if err != nil {
				return loop.RoundResult{}, fmt.Errorf("orchestrate: PromptBuilder for %q: %w", p.ID, err)
			}
			res, err := p.Session.Send(ctx, prompt)
			if err != nil {
				return loop.RoundResult{}, fmt.Errorf("orchestrate: Send for %q: %w", p.ID, err)
			}
			turn := Turn{Round: rc.Round, Speaker: p.ID, Content: res.Content, Usage: res.Usage}
			tr.Append(turn)
			if cfg.OnTurn != nil {
				cfg.OnTurn(turn)
			}
		}
		return cfg.Verdict.Assess(tr, rc), nil
	}

	outcome, err := loop.Run(ctx, cfg.Limits, step)
	return Result{Transcript: tr, Outcome: outcome}, err
}

// RoundRobin returns a Scheduler that speaks participants in fixed or rotating order.
// rotate=false: same order every round (Config.Participants order).
// rotate=true: starting participant rotates by one each round; round r (1-based) starts at index (r-1)%n.
func RoundRobin(rotate bool) Scheduler {
	return &roundRobin{rotate: rotate}
}

type roundRobin struct{ rotate bool }

func (r *roundRobin) Order(rc loop.RoundContext, ps []Participant) []Participant {
	if !r.rotate || len(ps) == 0 {
		return ps
	}
	n := len(ps)
	start := (rc.Round - 1) % n
	out := make([]Participant, n)
	for i := range ps {
		out[i] = ps[(start+i)%n]
	}
	return out
}
