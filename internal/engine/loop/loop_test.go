package loop_test

import (
	"context"
	"errors"
	"testing"

	"github.com/heurema/debate/internal/engine/loop"
)

var validLimits = loop.Limits{Max: 10, Settle: 2, Patience: 3}

func makeStep(results []loop.RoundResult) loop.Step {
	i := 0
	return func(ctx context.Context, rc loop.RoundContext) (loop.RoundResult, error) {
		if i >= len(results) {
			return loop.RoundResult{}, nil
		}
		r := results[i]
		i++
		return r, nil
	}
}

func TestSettled(t *testing.T) {
	step := makeStep([]loop.RoundResult{
		{Clean: true},
		{Clean: true},
	})
	out, err := loop.Run(context.Background(), loop.Limits{Max: 10, Settle: 2, Patience: 5}, step)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Reason != "settled" {
		t.Errorf("reason = %q, want settled", out.Reason)
	}
	if out.Rounds != 2 {
		t.Errorf("rounds = %d, want 2", out.Rounds)
	}
	if !out.Last.Clean {
		t.Error("Last.Clean should be true")
	}
}

func TestStalemate(t *testing.T) {
	// 3 consecutive no-progress rounds -> stalemate with Patience=3
	step := makeStep([]loop.RoundResult{
		{Clean: false, Progress: false},
		{Clean: false, Progress: false},
		{Clean: false, Progress: false},
	})
	out, err := loop.Run(context.Background(), loop.Limits{Max: 10, Settle: 5, Patience: 3}, step)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Reason != "stalemate" {
		t.Errorf("reason = %q, want stalemate", out.Reason)
	}
	if out.Rounds != 3 {
		t.Errorf("rounds = %d, want 3", out.Rounds)
	}
}

func TestStalemateResetByProgress(t *testing.T) {
	// Progress resets the no-progress streak; stalemate should come later
	step := makeStep([]loop.RoundResult{
		{Clean: false, Progress: false}, // streak=1
		{Clean: false, Progress: true},  // streak reset to 0
		{Clean: false, Progress: false}, // streak=1
		{Clean: false, Progress: false}, // streak=2
		{Clean: false, Progress: false}, // streak=3 -> stalemate
	})
	out, err := loop.Run(context.Background(), loop.Limits{Max: 10, Settle: 5, Patience: 3}, step)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Reason != "stalemate" {
		t.Errorf("reason = %q, want stalemate", out.Reason)
	}
	if out.Rounds != 5 {
		t.Errorf("rounds = %d, want 5", out.Rounds)
	}
}

func TestMax(t *testing.T) {
	step := makeStep([]loop.RoundResult{
		{Clean: false, Progress: true},
		{Clean: false, Progress: true},
		{Clean: false, Progress: true},
	})
	out, err := loop.Run(context.Background(), loop.Limits{Max: 3, Settle: 5, Patience: 5}, step)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Reason != "max" {
		t.Errorf("reason = %q, want max", out.Reason)
	}
	if out.Rounds != 3 {
		t.Errorf("rounds = %d, want 3", out.Rounds)
	}
}

func TestImmediateStop(t *testing.T) {
	step := makeStep([]loop.RoundResult{
		{Stop: &loop.Stop{Reason: "done"}},
	})
	out, err := loop.Run(context.Background(), validLimits, step)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Reason != "stop" {
		t.Errorf("reason = %q, want stop", out.Reason)
	}
	if out.Rounds != 1 {
		t.Errorf("rounds = %d, want 1", out.Rounds)
	}
	if out.Last.Stop == nil {
		t.Error("Last.Stop should be non-nil")
	}
}

func TestStepError(t *testing.T) {
	stepErr := errors.New("step failed")
	calls := 0
	step := func(ctx context.Context, rc loop.RoundContext) (loop.RoundResult, error) {
		calls++
		if calls == 2 {
			return loop.RoundResult{}, stepErr
		}
		return loop.RoundResult{Clean: false, Progress: true}, nil
	}
	out, err := loop.Run(context.Background(), validLimits, step)
	if !errors.Is(err, stepErr) {
		t.Fatalf("err = %v, want stepErr", err)
	}
	// 1 fully completed round before the error
	if out.Rounds != 1 {
		t.Errorf("rounds = %d, want 1", out.Rounds)
	}
	if out.Reason != "" {
		t.Errorf("reason = %q, want empty on error", out.Reason)
	}
	if calls != 2 {
		t.Errorf("step called %d times, want 2", calls)
	}
}

func TestStepErrorOnFirstRound(t *testing.T) {
	stepErr := errors.New("fail")
	step := func(ctx context.Context, rc loop.RoundContext) (loop.RoundResult, error) {
		return loop.RoundResult{}, stepErr
	}
	out, err := loop.Run(context.Background(), validLimits, step)
	if !errors.Is(err, stepErr) {
		t.Fatalf("err = %v, want stepErr", err)
	}
	if out.Rounds != 0 {
		t.Errorf("rounds = %d, want 0 (no round completed)", out.Rounds)
	}
	// Last should be zero RoundResult
	if out.Last.Clean || out.Last.Progress || out.Last.Stop != nil {
		t.Error("Last should be zero RoundResult when no round completed")
	}
}

func TestPreRoundCtxCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	step := func(ctx context.Context, rc loop.RoundContext) (loop.RoundResult, error) {
		calls++
		if calls == 1 {
			cancel() // cancel after first round succeeds
		}
		return loop.RoundResult{Clean: false, Progress: true}, nil
	}

	out, err := loop.Run(ctx, validLimits, step)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	// First round completed, second round pre-check detects cancellation
	if out.Rounds != 1 {
		t.Errorf("rounds = %d, want 1", out.Rounds)
	}
	if out.Reason != "" {
		t.Errorf("reason = %q, want empty on ctx error", out.Reason)
	}
}

func TestPreRoundCtxAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before Run

	calls := 0
	step := func(ctx context.Context, rc loop.RoundContext) (loop.RoundResult, error) {
		calls++
		return loop.RoundResult{}, nil
	}

	out, err := loop.Run(ctx, validLimits, step)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if calls != 0 {
		t.Errorf("step called %d times, want 0", calls)
	}
	if out.Rounds != 0 {
		t.Errorf("rounds = %d, want 0", out.Rounds)
	}
}

func TestInvalidLimits(t *testing.T) {
	cases := []loop.Limits{
		{Max: 0, Settle: 1, Patience: 1},
		{Max: 1, Settle: 0, Patience: 1},
		{Max: 1, Settle: 1, Patience: 0},
		{Max: -1, Settle: 1, Patience: 1},
	}
	for _, lim := range cases {
		out, err := loop.Run(context.Background(), lim, func(ctx context.Context, rc loop.RoundContext) (loop.RoundResult, error) {
			return loop.RoundResult{}, nil
		})
		if err == nil {
			t.Errorf("limits %+v: expected error, got nil", lim)
		}
		if out.Rounds != 0 {
			t.Errorf("limits %+v: rounds = %d, want 0", lim, out.Rounds)
		}
	}
}

func TestRoundContextNumbers(t *testing.T) {
	var rounds []int
	step := func(ctx context.Context, rc loop.RoundContext) (loop.RoundResult, error) {
		rounds = append(rounds, rc.Round)
		return loop.RoundResult{Clean: false, Progress: true}, nil
	}
	loop.Run(context.Background(), loop.Limits{Max: 3, Settle: 5, Patience: 5}, step)
	if len(rounds) != 3 || rounds[0] != 1 || rounds[1] != 2 || rounds[2] != 3 {
		t.Errorf("round numbers = %v, want [1 2 3]", rounds)
	}
}
