// Package progress writes the debate CLI progress event stream.
package progress

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// Prefix starts every machine-readable progress event line on stderr.
const Prefix = "@@DEBATE_PROGRESS "

const (
	stageLoadingWorkspace = "loading_workspace"
	stageOpeningSession   = "opening_session"
	stageRunningRound     = "running_round"
	stageRunningTurn      = "running_turn"
	stageSynthesizing     = "synthesizing"
	stageCompleted        = "completed"
	stageFailed           = "failed"
)

// Emitter serializes v1 progress events to one JSON object per prefixed line.
type Emitter struct {
	mu          sync.Mutex
	w           io.Writer
	start       time.Time
	activeStage string
	started     bool
	finished    bool
}

// NewEmitter returns a concurrency-safe progress emitter.
func NewEmitter(w io.Writer) *Emitter {
	return &Emitter{w: w}
}

// Event is the v1 progress event wire model.
type Event struct {
	Version    int    `json:"version"`
	Type       string `json:"type"`
	Stage      string `json:"stage"`
	ElapsedMS  int64  `json:"elapsed_ms"`
	DurationMS *int64 `json:"duration_ms,omitempty"`
	SilenceMS  *int64 `json:"silence_ms,omitempty"`
	Round      *int   `json:"round,omitempty"`
	Speaker    string `json:"speaker,omitempty"`
	Error      string `json:"error,omitempty"`
}

func (e *Emitter) RunStarted() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.start = time.Now()
	e.started = true
	e.finished = false
	e.activeStage = stageLoadingWorkspace
	e.emitLocked(Event{Type: "run_started", Stage: stageLoadingWorkspace})
}

func (e *Emitter) WorkspaceLoaded() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeStage = ""
	e.emitLocked(Event{Type: "workspace_loaded", Stage: stageLoadingWorkspace})
}

func (e *Emitter) SessionOpening(speaker string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeStage = stageOpeningSession
	e.emitLocked(Event{Type: "session_opening", Stage: stageOpeningSession, Speaker: speaker})
}

func (e *Emitter) SessionOpened(speaker string, duration time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeStage = ""
	e.emitLocked(Event{Type: "session_opened", Stage: stageOpeningSession, Speaker: speaker, DurationMS: msPtr(duration)})
}

func (e *Emitter) RoundStarted(round int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeStage = stageRunningRound
	e.emitLocked(Event{Type: "round_started", Stage: stageRunningRound, Round: intPtr(round)})
}

func (e *Emitter) TurnStarted(round int, speaker string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeStage = stageRunningTurn
	e.emitLocked(Event{Type: "turn_started", Stage: stageRunningTurn, Round: intPtr(round), Speaker: speaker})
}

func (e *Emitter) Heartbeat(round int, speaker string, silence time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.emitLocked(Event{
		Type:      "heartbeat",
		Stage:     stageRunningTurn,
		SilenceMS: msPtr(silence),
		Round:     intPtr(round),
		Speaker:   speaker,
	})
}

func (e *Emitter) TurnCompleted(round int, speaker string, duration time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeStage = stageRunningRound
	e.emitLocked(Event{
		Type:       "turn_completed",
		Stage:      stageRunningTurn,
		Round:      intPtr(round),
		Speaker:    speaker,
		DurationMS: msPtr(duration),
	})
}

func (e *Emitter) RoundCompleted(round int, duration time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeStage = ""
	e.emitLocked(Event{
		Type:       "round_completed",
		Stage:      stageRunningRound,
		Round:      intPtr(round),
		DurationMS: msPtr(duration),
	})
}

func (e *Emitter) SynthesisStarted() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeStage = stageSynthesizing
	e.emitLocked(Event{Type: "synthesis_started", Stage: stageSynthesizing})
}

func (e *Emitter) SynthesisHeartbeat(silence time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.emitLocked(Event{Type: "heartbeat", Stage: stageSynthesizing, SilenceMS: msPtr(silence)})
}

func (e *Emitter) SynthesisCompleted(duration time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeStage = ""
	e.emitLocked(Event{Type: "synthesis_completed", Stage: stageSynthesizing, DurationMS: msPtr(duration)})
}

func (e *Emitter) RunCompleted(duration time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.activeStage = stageCompleted
	e.finished = true
	e.emitLocked(Event{Type: "run_completed", Stage: stageCompleted, DurationMS: msPtr(duration)})
}

func (e *Emitter) RunFailed(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.finished {
		return
	}
	stage := e.activeStage
	if stage == "" {
		stage = stageFailed
	}
	msg := "unknown error"
	if err != nil && err.Error() != "" {
		msg = err.Error()
	}
	e.finished = true
	e.emitLocked(Event{Type: "run_failed", Stage: stage, Error: msg})
}

func (e *Emitter) emitLocked(ev Event) {
	if e.w == nil {
		return
	}
	if !e.started {
		e.start = time.Now()
		e.started = true
	}
	ev.Version = 1
	ev.ElapsedMS = millis(time.Since(e.start))
	line, err := json.Marshal(ev)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(e.w, "%s%s\n", Prefix, line)
}

func intPtr(v int) *int {
	return &v
}

func msPtr(d time.Duration) *int64 {
	ms := millis(d)
	return &ms
}

func millis(d time.Duration) int64 {
	if d < 0 {
		return 0
	}
	return int64(d / time.Millisecond)
}
