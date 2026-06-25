package prompt_test

import (
	"strings"
	"testing"

	"github.com/heurema/debate/internal/debate/prompt"
	"github.com/heurema/debate/internal/engine/loop"
	"github.com/heurema/debate/internal/engine/orchestrate"
)

func build(t *testing.T, brief string, p orchestrate.Participant, tr *orchestrate.Transcript, m orchestrate.RenderMode) string {
	t.Helper()
	builder := prompt.NewPromptBuilder(brief)
	text, err := builder(p, tr, loop.RoundContext{Round: 1}, m)
	if err != nil {
		t.Fatalf("PromptBuilder error: %v", err)
	}
	return text
}

func TestPromptContainsBrief(t *testing.T) {
	brief := "Unique task description abc123"
	tr := &orchestrate.Transcript{}
	p := orchestrate.Participant{ID: "A"}
	text := build(t, brief, p, tr, orchestrate.Delta)
	if !strings.Contains(text, brief) {
		t.Errorf("prompt does not contain brief %q", brief)
	}
}

func TestPromptContainsSignalInstruction(t *testing.T) {
	tr := &orchestrate.Transcript{}
	p := orchestrate.Participant{ID: "A"}
	text := build(t, "task", p, tr, orchestrate.Delta)
	if !strings.Contains(text, "```signal") {
		t.Error("prompt does not contain signal block instruction")
	}
	if !strings.Contains(text, "done") {
		t.Error("prompt does not reference done field")
	}
}

func TestPromptDeltaMode(t *testing.T) {
	tr := &orchestrate.Transcript{}
	tr.Append(orchestrate.Turn{Round: 1, Speaker: "A", Content: "A's turn content unique"})
	tr.Append(orchestrate.Turn{Round: 1, Speaker: "B", Content: "B's unique content xyz987"})

	p := orchestrate.Participant{ID: "A"}
	text := build(t, "task", p, tr, orchestrate.Delta)

	if !strings.Contains(text, "B's unique content xyz987") {
		t.Error("Delta mode: expected B's content in prompt")
	}
	if strings.Contains(text, "A's turn content unique") {
		t.Error("Delta mode: A's own content should not appear in prompt")
	}
}

func TestPromptFullMode(t *testing.T) {
	tr := &orchestrate.Transcript{}
	tr.Append(orchestrate.Turn{Round: 1, Speaker: "A", Content: "A's full content"})
	tr.Append(orchestrate.Turn{Round: 1, Speaker: "B", Content: "B's full content"})

	p := orchestrate.Participant{ID: "A"}
	text := build(t, "task", p, tr, orchestrate.Full)

	if !strings.Contains(text, "A's full content") {
		t.Error("Full mode: expected A's content in prompt")
	}
	if !strings.Contains(text, "B's full content") {
		t.Error("Full mode: expected B's content in prompt")
	}
}

func TestPromptBoardLabelsTurnsByRoundAndSpeaker(t *testing.T) {
	tr := &orchestrate.Transcript{}
	tr.Append(orchestrate.Turn{Round: 2, Speaker: "C", Content: "C speaks in round 2"})

	p := orchestrate.Participant{ID: "A"}
	builder := prompt.NewPromptBuilder("task")
	text, err := builder(p, tr, loop.RoundContext{Round: 2}, orchestrate.Full)
	if err != nil {
		t.Fatalf("PromptBuilder error: %v", err)
	}
	if !strings.Contains(text, "Round 2") {
		t.Error("board entry should include Round number")
	}
	if !strings.Contains(text, "C") {
		t.Error("board entry should include speaker name")
	}
}

func TestPromptFullModeRendersRequiredSectionsAndInstructions(t *testing.T) {
	const brief = "Decide whether to keep the release blocker open."
	tr := &orchestrate.Transcript{}
	tr.Append(orchestrate.Turn{Round: 1, Speaker: "Alice", Content: "Alice previous position"})
	tr.Append(orchestrate.Turn{Round: 2, Speaker: "Bob", Content: "Bob current objection"})

	text := build(t, brief, orchestrate.Participant{ID: "Alice"}, tr, orchestrate.Full)

	required := []string{
		"You are a participant in a structured deliberation. Each turn:",
		"1. State your current position clearly.",
		"2. List any blocking objections that remain unresolved.",
		"3. Set done=true only when you have no remaining objections.",
		"## Brief\n\n" + brief,
		"## Discussion",
		"[Round 1 — Alice]\nAlice previous position",
		"[Round 2 — Bob]\nBob current objection",
		"End your reply with a fenced signal block:",
		"```signal",
		`{"position": "<your position>", "objections": ["<obj1>", "..."], "done": <true|false>}`,
		"Set done=true only if you have no objections. If you list any objections, done must be false.",
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("prompt missing required text %q\nprompt:\n%s", want, text)
		}
	}
}

func TestPromptEmptyTranscript(t *testing.T) {
	tr := &orchestrate.Transcript{}
	p := orchestrate.Participant{ID: "A"}
	brief := "empty transcript task"

	for _, m := range []orchestrate.RenderMode{orchestrate.Delta, orchestrate.Full} {
		text := build(t, brief, p, tr, m)
		if !strings.Contains(text, brief) {
			t.Errorf("mode %v: brief missing from prompt", m)
		}
		if !strings.Contains(text, "```signal") {
			t.Errorf("mode %v: signal instruction missing from prompt", m)
		}
	}
}
