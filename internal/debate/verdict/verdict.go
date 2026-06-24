// Package verdict provides the debate Verdict implementation.
package verdict

import (
	"github.com/heurema/debate/internal/debate/signal"
	"github.com/heurema/debate/internal/engine/loop"
	"github.com/heurema/debate/internal/engine/orchestrate"
)

// Until controls when a round is considered clean.
type Until int

const (
	AllDone Until = iota // every speaker must have Done==true
	Quorum               // strictly more than half of speakers must have Done==true
)

type debateVerdict struct {
	until    Until
	prevObjs map[string]struct{}
}

// New returns an orchestrate.Verdict that judges rounds by until.
func New(until Until) orchestrate.Verdict {
	return &debateVerdict{until: until}
}

// Assess judges the current round's turns and returns the round result.
// An unparsed signal counts as not-done and contributes no objections.
func (v *debateVerdict) Assess(t *orchestrate.Transcript, rc loop.RoundContext) loop.RoundResult {
	turns := roundTurns(t.All(), rc.Round)

	doneCount := 0
	objSet := make(map[string]struct{})
	for _, turn := range turns {
		sig, ok := signal.Parse(turn.Content)
		if !ok {
			continue
		}
		if sig.Done {
			doneCount++
		}
		for _, obj := range sig.Objections {
			objSet[obj] = struct{}{}
		}
	}

	n := len(turns)
	var clean bool
	switch v.until {
	case Quorum:
		clean = n > 0 && doneCount*2 > n
	default: // AllDone
		clean = n > 0 && doneCount == n
	}

	progress := !mapsEqual(objSet, v.prevObjs)
	v.prevObjs = objSet

	return loop.RoundResult{
		Clean:    clean,
		Progress: progress,
	}
}

func roundTurns(turns []orchestrate.Turn, round int) []orchestrate.Turn {
	var out []orchestrate.Turn
	for _, t := range turns {
		if t.Round == round {
			out = append(out, t)
		}
	}
	return out
}

func mapsEqual(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}
