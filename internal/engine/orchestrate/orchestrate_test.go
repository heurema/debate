package orchestrate_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/heurema/debate/internal/engine/loop"
	"github.com/heurema/debate/internal/engine/orchestrate"
	"github.com/heurema/debate/internal/engine/transport"
	"github.com/heurema/debate/internal/engine/transport/mock"
)

// trivialVerdict returns a fixed RoundResult every round.
type trivialVerdict struct{ result loop.RoundResult }

func (v trivialVerdict) Assess(_ *orchestrate.Transcript, _ loop.RoundContext) loop.RoundResult {
	return v.result
}

// echoPrompt builds a prompt that identifies the participant and their delta.
func echoPrompt(p orchestrate.Participant, t *orchestrate.Transcript, rc loop.RoundContext, m orchestrate.RenderMode) (string, error) {
	return fmt.Sprintf("round=%d speaker=%s", rc.Round, p.ID), nil
}

func makeSession(responses []string) *mock.Session {
	script := make([]mock.ScriptedResult, len(responses))
	for i, r := range responses {
		script[i] = mock.ScriptedResult{Result: transport.Result{Content: r}}
	}
	return mock.NewSession(script)
}

// TestTurnOrderFixedRoundRobin verifies that with rotate=false, order is stable every round.
func TestTurnOrderFixedRoundRobin(t *testing.T) {
	sA := makeSession([]string{"a1", "a2"})
	sB := makeSession([]string{"b1", "b2"})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    echoPrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: false, Progress: true}},
		Limits:    loop.Limits{Max: 2, Settle: 5, Patience: 5},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	turns := res.Transcript.All()
	// 2 rounds × 2 participants = 4 turns
	if len(turns) != 4 {
		t.Fatalf("turn count = %d, want 4", len(turns))
	}
	wantSpeakers := []string{"A", "B", "A", "B"}
	for i, turn := range turns {
		if turn.Speaker != wantSpeakers[i] {
			t.Errorf("turn[%d].Speaker = %q, want %q", i, turn.Speaker, wantSpeakers[i])
		}
	}
}

// TestTurnOrderRotatingRoundRobin verifies rotation: round r starts at (r-1)%n.
func TestTurnOrderRotatingRoundRobin(t *testing.T) {
	sA := makeSession([]string{"a1", "a2"})
	sB := makeSession([]string{"b1", "b2"})
	sC := makeSession([]string{"c1", "c2"})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
			{ID: "C", Session: sC},
		},
		Scheduler: orchestrate.RoundRobin(true),
		Prompt:    echoPrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: false, Progress: true}},
		Limits:    loop.Limits{Max: 2, Settle: 5, Patience: 5},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	turns := res.Transcript.All()
	// round 1: start at (1-1)%3=0 → A,B,C
	// round 2: start at (2-1)%3=1 → B,C,A
	wantSpeakers := []string{"A", "B", "C", "B", "C", "A"}
	if len(turns) != len(wantSpeakers) {
		t.Fatalf("turn count = %d, want %d", len(turns), len(wantSpeakers))
	}
	for i, turn := range turns {
		if turn.Speaker != wantSpeakers[i] {
			t.Errorf("turn[%d].Speaker = %q, want %q", i, turn.Speaker, wantSpeakers[i])
		}
	}
}

// TestTranscriptAccumulation checks turn fields and content.
func TestTranscriptAccumulation(t *testing.T) {
	sA := makeSession([]string{"hello from A"})
	sB := makeSession([]string{"hello from B"})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    echoPrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: true}},
		Limits:    loop.Limits{Max: 5, Settle: 1, Patience: 5},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	turns := res.Transcript.All()
	if len(turns) != 2 {
		t.Fatalf("turn count = %d, want 2", len(turns))
	}
	if turns[0].Round != 1 || turns[0].Speaker != "A" || turns[0].Content != "hello from A" {
		t.Errorf("turn[0] = %+v", turns[0])
	}
	if turns[1].Round != 1 || turns[1].Speaker != "B" || turns[1].Content != "hello from B" {
		t.Errorf("turn[1] = %+v", turns[1])
	}
}

// TestDeltaForSameRound verifies that B sees A's turn when building its prompt in the same round.
func TestDeltaForSameRound(t *testing.T) {
	var deltaSeenByB []orchestrate.Turn

	capturePrompt := func(p orchestrate.Participant, t *orchestrate.Transcript, rc loop.RoundContext, m orchestrate.RenderMode) (string, error) {
		if p.ID == "B" {
			deltaSeenByB = t.DeltaFor("B")
		}
		return fmt.Sprintf("round=%d speaker=%s", rc.Round, p.ID), nil
	}

	sA := makeSession([]string{"A speaks"})
	sB := makeSession([]string{"B speaks"})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    capturePrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: true}},
		Limits:    loop.Limits{Max: 5, Settle: 1, Patience: 5},
	}

	_, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// B should see A's turn (same round, A spoke first)
	if len(deltaSeenByB) != 1 || deltaSeenByB[0].Speaker != "A" {
		t.Errorf("B's delta = %+v, want [A's turn]", deltaSeenByB)
	}
}

// TestDeltaForNextRound verifies per-participant delta across rounds.
func TestDeltaForNextRound(t *testing.T) {
	var deltaSeenByARound2 []orchestrate.Turn

	round := 0
	capturePrompt := func(p orchestrate.Participant, t *orchestrate.Transcript, rc loop.RoundContext, m orchestrate.RenderMode) (string, error) {
		if p.ID == "A" && rc.Round == 2 {
			deltaSeenByARound2 = t.DeltaFor("A")
		}
		return fmt.Sprintf("round=%d speaker=%s", rc.Round, p.ID), nil
	}

	sA := makeSession([]string{"A r1", "A r2"})
	sB := makeSession([]string{"B r1", "B r2"})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    capturePrompt,
		Verdict: verdictAfter(2, func(r int) loop.RoundResult {
			round = r
			return loop.RoundResult{Clean: false, Progress: true}
		}),
		Limits: loop.Limits{Max: 2, Settle: 5, Patience: 5},
	}

	_, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	_ = round

	// At the start of round 2, A's cursor is after their round-1 turn (index 1).
	// The transcript contains [A r1, B r1]. A's cursor=1 (after their own r1 turn).
	// DeltaFor("A") = turns from index 1 = [B r1] (B spoke since A last spoke).
	if len(deltaSeenByARound2) != 1 || deltaSeenByARound2[0].Speaker != "B" {
		t.Errorf("A's delta in round 2 = %+v, want [B r1]", deltaSeenByARound2)
	}
}

// verdictAfter returns a Verdict that produces progress results until max rounds then switches.
type funcVerdict struct {
	fn func(round int) loop.RoundResult
}

func (v funcVerdict) Assess(_ *orchestrate.Transcript, rc loop.RoundContext) loop.RoundResult {
	return v.fn(rc.Round)
}

func verdictAfter(n int, fn func(int) loop.RoundResult) orchestrate.Verdict {
	return funcVerdict{fn: fn}
}

// TestOutcomeSettled verifies settled outcome propagation.
func TestOutcomeSettled(t *testing.T) {
	sA := makeSession([]string{"a", "a"})
	sB := makeSession([]string{"b", "b"})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    echoPrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: true}},
		Limits:    loop.Limits{Max: 10, Settle: 2, Patience: 5},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Outcome.Reason != "settled" {
		t.Errorf("Outcome.Reason = %q, want settled", res.Outcome.Reason)
	}
	if res.Outcome.Rounds != 2 {
		t.Errorf("Outcome.Rounds = %d, want 2", res.Outcome.Rounds)
	}
}

// TestOutcomeStalemate verifies stalemate outcome propagation.
func TestOutcomeStalemate(t *testing.T) {
	sA := makeSession([]string{"a", "a", "a"})
	sB := makeSession([]string{"b", "b", "b"})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    echoPrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: false, Progress: false}},
		Limits:    loop.Limits{Max: 10, Settle: 5, Patience: 3},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Outcome.Reason != "stalemate" {
		t.Errorf("Outcome.Reason = %q, want stalemate", res.Outcome.Reason)
	}
}

// TestOutcomeMax verifies max outcome propagation.
func TestOutcomeMax(t *testing.T) {
	sA := makeSession([]string{"a", "a"})
	sB := makeSession([]string{"b", "b"})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    echoPrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: false, Progress: true}},
		Limits:    loop.Limits{Max: 2, Settle: 5, Patience: 5},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Outcome.Reason != "max" {
		t.Errorf("Outcome.Reason = %q, want max", res.Outcome.Reason)
	}
	if res.Outcome.Rounds != 2 {
		t.Errorf("Outcome.Rounds = %d, want 2", res.Outcome.Rounds)
	}
}

// TestMissingConfigFields verifies that missing required fields are rejected before any round.
func TestMissingConfigFields(t *testing.T) {
	sA := makeSession([]string{"a"})
	base := orchestrate.Config{
		Participants: []orchestrate.Participant{{ID: "A", Session: sA}},
		Scheduler:    orchestrate.RoundRobin(false),
		Prompt:       echoPrompt,
		Verdict:      trivialVerdict{loop.RoundResult{Clean: true}},
		Limits:       loop.Limits{Max: 1, Settle: 1, Patience: 1},
	}

	cases := []struct {
		name string
		cfg  orchestrate.Config
	}{
		{"no participants", orchestrate.Config{Scheduler: base.Scheduler, Prompt: base.Prompt, Verdict: base.Verdict, Limits: base.Limits}},
		{"nil scheduler", orchestrate.Config{Participants: base.Participants, Prompt: base.Prompt, Verdict: base.Verdict, Limits: base.Limits}},
		{"nil prompt", orchestrate.Config{Participants: base.Participants, Scheduler: base.Scheduler, Verdict: base.Verdict, Limits: base.Limits}},
		{"nil verdict", orchestrate.Config{Participants: base.Participants, Scheduler: base.Scheduler, Prompt: base.Prompt, Limits: base.Limits}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := orchestrate.Run(context.Background(), tc.cfg)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

// TestOnTurnCallback verifies that OnTurn is called for each turn.
func TestOnTurnCallback(t *testing.T) {
	sA := makeSession([]string{"a"})
	sB := makeSession([]string{"b"})

	var recorded []orchestrate.Turn
	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    echoPrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: true}},
		Limits:    loop.Limits{Max: 5, Settle: 1, Patience: 5},
		OnTurn:    func(turn orchestrate.Turn) { recorded = append(recorded, turn) },
	}

	_, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(recorded) != 2 {
		t.Errorf("OnTurn called %d times, want 2", len(recorded))
	}
}

// TestSessionSendErrorSurfaced verifies that a Send error is returned and partial turns are kept.
func TestSessionSendErrorSurfaced(t *testing.T) {
	sendErr := errors.New("send failed")
	sA := makeSession([]string{"a1"})
	sB := mock.NewSession([]mock.ScriptedResult{
		{Err: sendErr},
	})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    echoPrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: true}},
		Limits:    loop.Limits{Max: 5, Settle: 1, Patience: 5},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sendErr) {
		t.Errorf("err = %v, want to wrap sendErr", err)
	}
	// A's turn was appended before B's error; partial turn is retained
	turns := res.Transcript.All()
	if len(turns) != 1 || turns[0].Speaker != "A" {
		t.Errorf("transcript = %+v, want [A's turn]", turns)
	}
}
