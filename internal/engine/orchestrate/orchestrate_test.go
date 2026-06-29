package orchestrate_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/heurema/debate/internal/engine/loop"
	"github.com/heurema/debate/internal/engine/orchestrate"
	"github.com/heurema/debate/internal/engine/transport"
	"github.com/heurema/debate/internal/engine/transport/mock"
)

type progressEvent struct {
	typ        string
	round      int
	speaker    string
	durationMS int64
	silenceMS  int64
}

type recordingProgress struct {
	mu     sync.Mutex
	events []progressEvent
	ch     chan progressEvent
}

func newRecordingProgress() *recordingProgress {
	return &recordingProgress{ch: make(chan progressEvent, 20)}
}

func (p *recordingProgress) add(ev progressEvent) {
	p.mu.Lock()
	p.events = append(p.events, ev)
	p.mu.Unlock()
	p.ch <- ev
}

func (p *recordingProgress) RoundStarted(round int) {
	p.add(progressEvent{typ: "round_started", round: round})
}

func (p *recordingProgress) TurnStarted(round int, speaker string) {
	p.add(progressEvent{typ: "turn_started", round: round, speaker: speaker})
}

func (p *recordingProgress) Heartbeat(round int, speaker string, silence time.Duration) {
	p.add(progressEvent{typ: "heartbeat", round: round, speaker: speaker, silenceMS: int64(silence / time.Millisecond)})
}

func (p *recordingProgress) TurnCompleted(round int, speaker string, duration time.Duration) {
	p.add(progressEvent{typ: "turn_completed", round: round, speaker: speaker, durationMS: int64(duration / time.Millisecond)})
}

func (p *recordingProgress) RoundCompleted(round int, duration time.Duration) {
	p.add(progressEvent{typ: "round_completed", round: round, durationMS: int64(duration / time.Millisecond)})
}

func (p *recordingProgress) snapshot() []progressEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]progressEvent, len(p.events))
	copy(out, p.events)
	return out
}

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

func TestRunBuildsParticipantPromptsWithFullTranscriptSoFar(t *testing.T) {
	type promptCall struct {
		speaker string
		round   int
		mode    orchestrate.RenderMode
		turns   []orchestrate.Turn
	}
	var calls []promptCall

	capturePrompt := func(p orchestrate.Participant, t *orchestrate.Transcript, rc loop.RoundContext, m orchestrate.RenderMode) (string, error) {
		calls = append(calls, promptCall{
			speaker: p.ID,
			round:   rc.Round,
			mode:    m,
			turns:   t.All(),
		})
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
		Verdict:   trivialVerdict{loop.RoundResult{Clean: false, Progress: true}},
		Limits:    loop.Limits{Max: 2, Settle: 5, Patience: 5},
	}

	_, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(calls) != 4 {
		t.Fatalf("prompt calls = %d, want 4", len(calls))
	}
	for i, call := range calls {
		if call.mode != orchestrate.Full {
			t.Fatalf("call[%d] mode = %v, want Full", i, call.mode)
		}
	}

	want := []struct {
		speaker  string
		round    int
		contents []string
	}{
		{speaker: "A", round: 1, contents: nil},
		{speaker: "B", round: 1, contents: []string{"A r1"}},
		{speaker: "A", round: 2, contents: []string{"A r1", "B r1"}},
		{speaker: "B", round: 2, contents: []string{"A r1", "B r1", "A r2"}},
	}
	for i, wantCall := range want {
		call := calls[i]
		if call.speaker != wantCall.speaker || call.round != wantCall.round {
			t.Fatalf("call[%d] = speaker %q round %d, want speaker %q round %d", i, call.speaker, call.round, wantCall.speaker, wantCall.round)
		}
		if got := turnContents(call.turns); !equalStrings(got, wantCall.contents) {
			t.Fatalf("call[%d] transcript contents = %v, want %v", i, got, wantCall.contents)
		}
	}
}

func TestRunPromptTranscriptExcludesFutureTurns(t *testing.T) {
	var aliceRound2 []orchestrate.Turn

	capturePrompt := func(p orchestrate.Participant, t *orchestrate.Transcript, rc loop.RoundContext, m orchestrate.RenderMode) (string, error) {
		if p.ID == "A" && rc.Round == 2 {
			aliceRound2 = t.All()
		}
		return fmt.Sprintf("round=%d speaker=%s", rc.Round, p.ID), nil
	}

	sA := makeSession([]string{"A already committed", "A future response"})
	sB := makeSession([]string{"B already committed", "B future response"})

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    capturePrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: false, Progress: true}},
		Limits:    loop.Limits{Max: 2, Settle: 5, Patience: 5},
	}

	_, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := turnContents(aliceRound2)
	if !equalStrings(got, []string{"A already committed", "B already committed"}) {
		t.Fatalf("A round 2 prompt transcript = %v, want only committed prior turns", got)
	}
	for _, future := range []string{"A future response", "B future response"} {
		if containsString(got, future) {
			t.Fatalf("A round 2 prompt transcript contains future turn %q: %v", future, got)
		}
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

func TestProgressLifecycleOrdering(t *testing.T) {
	sA := makeSession([]string{"a"})
	sB := makeSession([]string{"b"})
	progress := newRecordingProgress()

	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{
			{ID: "A", Session: sA},
			{ID: "B", Session: sB},
		},
		Scheduler: orchestrate.RoundRobin(false),
		Prompt:    echoPrompt,
		Verdict:   trivialVerdict{loop.RoundResult{Clean: true}},
		Limits:    loop.Limits{Max: 5, Settle: 1, Patience: 5},
		Progress:  progress,
	}

	_, err := orchestrate.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	events := progress.snapshot()
	got := eventTypes(events)
	want := []string{
		"round_started",
		"turn_started",
		"turn_completed",
		"turn_started",
		"turn_completed",
		"round_completed",
	}
	if !equalStrings(got, want) {
		t.Fatalf("progress event order = %v, want %v", got, want)
	}
	if events[1].round != 1 || events[1].speaker != "A" {
		t.Fatalf("first turn_started = %+v, want round 1 speaker A", events[1])
	}
	if events[2].round != 1 || events[2].speaker != "A" {
		t.Fatalf("first turn_completed = %+v, want round 1 speaker A", events[2])
	}
	if events[3].round != 1 || events[3].speaker != "B" {
		t.Fatalf("second turn_started = %+v, want round 1 speaker B", events[3])
	}
}

type blockingSession struct {
	entered chan struct{}
	release chan struct{}
	result  transport.Result
}

func newBlockingSession(content string) *blockingSession {
	return &blockingSession{
		entered: make(chan struct{}),
		release: make(chan struct{}),
		result:  transport.Result{Content: content},
	}
}

func (s *blockingSession) Send(ctx context.Context, _ string) (transport.Result, error) {
	close(s.entered)
	select {
	case <-s.release:
		return s.result, nil
	case <-ctx.Done():
		return transport.Result{}, ctx.Err()
	}
}

func (s *blockingSession) Close() error { return nil }

func TestProgressHeartbeatDuringBlockedSend(t *testing.T) {
	session := newBlockingSession("a")
	progress := newRecordingProgress()
	cfg := orchestrate.Config{
		Participants: []orchestrate.Participant{{ID: "A", Session: session}},
		Scheduler:    orchestrate.RoundRobin(false),
		Prompt:       echoPrompt,
		Verdict:      trivialVerdict{loop.RoundResult{Clean: true}},
		Limits:       loop.Limits{Max: 5, Settle: 1, Patience: 5},
		Progress:     progress,

		HeartbeatInterval: time.Millisecond,
	}

	done := make(chan error, 1)
	go func() {
		_, err := orchestrate.Run(context.Background(), cfg)
		done <- err
	}()

	select {
	case <-session.entered:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("session send did not start")
	}

	var heartbeat progressEvent
	for {
		select {
		case ev := <-progress.ch:
			if ev.typ == "heartbeat" {
				heartbeat = ev
				goto release
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatal("heartbeat was not emitted while send was blocked")
		}
	}

release:
	if heartbeat.round != 1 || heartbeat.speaker != "A" || heartbeat.silenceMS < 1 {
		t.Fatalf("heartbeat = %+v, want round 1 speaker A with silence >= 1ms", heartbeat)
	}
	close(session.release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("run did not finish after releasing send")
	}

	events := progress.snapshot()
	last := events[len(events)-1]
	if last.typ != "round_completed" {
		t.Fatalf("last progress event = %+v, want round_completed", last)
	}
	for i, ev := range events {
		if ev.typ == "heartbeat" && i > indexOfEvent(events, "turn_completed") {
			t.Fatalf("heartbeat emitted after turn_completed: events=%v", events)
		}
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

func turnContents(turns []orchestrate.Turn) []string {
	out := make([]string, len(turns))
	for i, turn := range turns {
		out[i] = turn.Content
	}
	return out
}

func eventTypes(events []progressEvent) []string {
	out := make([]string, len(events))
	for i, ev := range events {
		out[i] = ev.typ
	}
	return out
}

func indexOfEvent(events []progressEvent, typ string) int {
	for i, ev := range events {
		if ev.typ == typ {
			return i
		}
	}
	return len(events)
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
