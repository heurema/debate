package verdict_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/heurema/debate/internal/debate/verdict"
	"github.com/heurema/debate/internal/engine/loop"
	"github.com/heurema/debate/internal/engine/orchestrate"
	"github.com/heurema/debate/internal/engine/transport"
	"github.com/heurema/debate/internal/engine/transport/mock"
)

func trivialPrompt(_ orchestrate.Participant, _ *orchestrate.Transcript, _ loop.RoundContext, _ orchestrate.RenderMode) (string, error) {
	return "speak", nil
}

func makeSession(responses []string) *mock.Session {
	script := make([]mock.ScriptedResult, len(responses))
	for i, r := range responses {
		script[i] = mock.ScriptedResult{Result: transport.Result{Content: r}}
	}
	return mock.NewSession(script)
}

func doneReply(position string) string {
	return fmt.Sprintf("My position.\n\n```signal\n{\"position\": %q, \"objections\": [], \"done\": true}\n```", position)
}

func notDoneReply(objection string) string {
	return fmt.Sprintf("Objecting.\n\n```signal\n{\"position\": \"disagree\", \"objections\": [%q], \"done\": false}\n```", objection)
}

func garbledReply() string {
	return "I have thoughts but no signal block here."
}

// TestVerdictSettledAllDone: both participants done every round → settled after Settle streak.
func TestVerdictSettledAllDone(t *testing.T) {
	sA := makeSession([]string{doneReply("A agrees"), doneReply("A agrees")})
	sB := makeSession([]string{doneReply("B agrees"), doneReply("B agrees")})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    trivialPrompt,
		Verdict:   verdict.New(verdict.AllDone),
		Limits:    loop.Limits{Max: 10, Settle: 2, Patience: 5},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Outcome.Reason != "settled" {
		t.Errorf("Outcome = %q, want settled", res.Outcome.Reason)
	}
	if res.Outcome.Rounds != 2 {
		t.Errorf("Rounds = %d, want 2", res.Outcome.Rounds)
	}
}

// TestVerdictSettledQuorum: 3 participants with 2 done → quorum → settled after Settle streak.
func TestVerdictSettledQuorum(t *testing.T) {
	sA := makeSession([]string{doneReply("A agrees"), doneReply("A agrees")})
	sB := makeSession([]string{doneReply("B agrees"), doneReply("B agrees")})
	sC := makeSession([]string{notDoneReply("still blocked"), notDoneReply("still blocked")})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
			{ID: "C", Session: sC},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    trivialPrompt,
		Verdict:   verdict.New(verdict.Quorum),
		Limits:    loop.Limits{Max: 10, Settle: 2, Patience: 5},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Outcome.Reason != "settled" {
		t.Errorf("Outcome = %q, want settled", res.Outcome.Reason)
	}
	if res.Outcome.Rounds != 2 {
		t.Errorf("Rounds = %d, want 2", res.Outcome.Rounds)
	}
}

// TestVerdictStalemate: frozen non-empty objection set with nobody done → stalemate after Patience.
func TestVerdictStalemate(t *testing.T) {
	// R1: objSet={"frozen objection"} vs prev={} → Progress=true → noPS=0
	// R2: objSet={"frozen objection"} vs prev={"frozen objection"} → Progress=false → noPS=1
	// R3: same → noPS=2 ≥ Patience=2 → stalemate
	frozen := notDoneReply("frozen objection")
	sA := makeSession([]string{frozen, frozen, frozen})
	sB := makeSession([]string{frozen, frozen, frozen})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    trivialPrompt,
		Verdict:   verdict.New(verdict.AllDone),
		Limits:    loop.Limits{Max: 10, Settle: 10, Patience: 2},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Outcome.Reason != "stalemate" {
		t.Errorf("Outcome = %q, want stalemate", res.Outcome.Reason)
	}
	if res.Outcome.Rounds != 3 {
		t.Errorf("Rounds = %d, want 3", res.Outcome.Rounds)
	}
}

// TestVerdictMax: progress every round, never converges → max rounds reached.
func TestVerdictMax(t *testing.T) {
	// Different objection each round keeps Progress=true, so no stalemate.
	// Nobody done keeps Clean=false, so no settling.
	// Max=3 terminates the loop.
	sA := makeSession([]string{notDoneReply("obj-1"), notDoneReply("obj-2"), notDoneReply("obj-3")})
	sB := makeSession([]string{notDoneReply("obj-1"), notDoneReply("obj-2"), notDoneReply("obj-3")})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    trivialPrompt,
		Verdict:   verdict.New(verdict.AllDone),
		Limits:    loop.Limits{Max: 3, Settle: 10, Patience: 10},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Outcome.Reason != "max" {
		t.Errorf("Outcome = %q, want max", res.Outcome.Reason)
	}
	if res.Outcome.Rounds != 3 {
		t.Errorf("Rounds = %d, want 3", res.Outcome.Rounds)
	}
}

// TestVerdictUnparsedSignal: a garbled turn counts as not-done and causes no error.
func TestVerdictUnparsedSignal(t *testing.T) {
	// A is done; B has no signal block (unparsed → not-done, no objections contributed).
	// AllDone: doneCount=1, n=2 → Clean=false.
	// objSet={} both rounds → Progress=false both rounds → noPS=1 after R1, noPS=2 after R2.
	// Max=2 terminates before Patience=3.
	sA := makeSession([]string{doneReply("A done"), doneReply("A done")})
	sB := makeSession([]string{garbledReply(), garbledReply()})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    trivialPrompt,
		Verdict:   verdict.New(verdict.AllDone),
		Limits:    loop.Limits{Max: 2, Settle: 10, Patience: 3},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if res.Outcome.Reason != "max" {
		t.Errorf("Outcome = %q, want max (unparsed signal should be not-done, not an error)", res.Outcome.Reason)
	}
	if res.Outcome.Rounds != 2 {
		t.Errorf("Rounds = %d, want 2", res.Outcome.Rounds)
	}
}

// TestVerdictProgressTracking: changing objection set → Progress=true; frozen → Progress=false.
func TestVerdictProgressTracking(t *testing.T) {
	// R1: objSet={"obj-a"} vs prev={} → Progress=true → noPS=0
	// R2: objSet={"obj-b"} vs prev={"obj-a"} → Progress=true → noPS=0 (reset)
	// R3: objSet={"obj-b"} vs prev={"obj-b"} → Progress=false → noPS=1 ≥ Patience=1 → stalemate
	sA := makeSession([]string{notDoneReply("obj-a"), notDoneReply("obj-b"), notDoneReply("obj-b")})
	sB := makeSession([]string{notDoneReply("obj-a"), notDoneReply("obj-b"), notDoneReply("obj-b")})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    trivialPrompt,
		Verdict:   verdict.New(verdict.AllDone),
		Limits:    loop.Limits{Max: 10, Settle: 10, Patience: 1},
	}

	res, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Outcome.Reason != "stalemate" {
		t.Errorf("Outcome = %q, want stalemate", res.Outcome.Reason)
	}
	if res.Outcome.Rounds != 3 {
		t.Errorf("Rounds = %d, want 3", res.Outcome.Rounds)
	}
}
