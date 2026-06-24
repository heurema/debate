// Package prompt provides the debate PromptBuilder.
package prompt

import (
	"fmt"
	"strings"

	"github.com/heurema/debate/internal/engine/loop"
	"github.com/heurema/debate/internal/engine/orchestrate"
)

const moderatorRules = `You are a participant in a structured deliberation. Each turn:
1. State your current position clearly.
2. List any blocking objections that remain unresolved.
3. Set done=true only when you have no remaining objections.`

const signalInstruction = "End your reply with a fenced signal block:\n\n" +
	"```signal\n" +
	`{"position": "<your position>", "objections": ["<obj1>", "..."], "done": <true|false>}` + "\n" +
	"```\n\n" +
	"Set done=true only if you have no objections. If you list any objections, done must be false."

// NewPromptBuilder returns an orchestrate.PromptBuilder that renders the moderator
// rules, the shared brief, the board of relevant turns, and the signal-format instruction.
func NewPromptBuilder(brief string) orchestrate.PromptBuilder {
	return func(p orchestrate.Participant, t *orchestrate.Transcript, rc loop.RoundContext, m orchestrate.RenderMode) (string, error) {
		var turns []orchestrate.Turn
		if m == orchestrate.Full {
			turns = t.All()
		} else {
			turns = t.DeltaFor(p.ID)
		}

		var sb strings.Builder
		sb.WriteString(moderatorRules)
		sb.WriteString("\n\n## Brief\n\n")
		sb.WriteString(brief)

		if board := renderBoard(turns); board != "" {
			sb.WriteString("\n\n## Discussion\n\n")
			sb.WriteString(board)
		}

		sb.WriteString("\n\n")
		sb.WriteString(signalInstruction)
		return sb.String(), nil
	}
}

func renderBoard(turns []orchestrate.Turn) string {
	if len(turns) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, t := range turns {
		fmt.Fprintf(&sb, "[Round %d — %s]\n%s\n\n", t.Round, t.Speaker, t.Content)
	}
	return strings.TrimRight(sb.String(), "\n")
}
