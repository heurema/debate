// Package loop manages the engine discussion loop lifecycle.
package loop

import (
	"context"
	"fmt"
)

// Limits controls when the loop terminates. All fields must be >= 1.
type Limits struct {
	Max      int // maximum rounds before "max" termination
	Settle   int // consecutive Clean rounds before "settled"
	Patience int // consecutive no-progress rounds before "stalemate"
}

// RoundContext carries per-round metadata passed to each Step call.
type RoundContext struct {
	Round int // 1-based
}

// Stop signals an immediate halt when returned inside RoundResult.
type Stop struct {
	Reason string
}

// RoundResult is returned by a Step to report round outcome.
// Precedence for terminal decisions: Stop > settled > stalemate > max.
type RoundResult struct {
	Clean    bool  // converged this round; increments clean streak
	Progress bool  // consulted only when Clean==false; true resets no-progress streak
	Stop     *Stop // non-nil forces immediate "stop" termination
}

// Outcome describes why the loop ended and how many rounds ran.
type Outcome struct {
	Reason string      // "settled" | "stalemate" | "max" | "stop" | "" (error path)
	Rounds int         // count of fully completed rounds
	Last   RoundResult // result of the last completed round; zero if none
}

// Step is called once per round. A non-nil error halts the loop immediately.
type Step func(ctx context.Context, rc RoundContext) (RoundResult, error)

// Run drives rounds until a terminal condition or error.
// It validates Limits first; any field < 1 is an error with zero rounds completed.
// ctx is checked before each round; cancellation is an error path (returns ctx.Err()).
func Run(ctx context.Context, lim Limits, step Step) (Outcome, error) {
	if lim.Max < 1 || lim.Settle < 1 || lim.Patience < 1 {
		return Outcome{}, fmt.Errorf("loop: Limits fields Max, Settle, Patience must each be >= 1")
	}

	var cleanStreak, noProgressStreak int
	var last RoundResult
	var completed int

	for round := 1; ; round++ {
		if err := ctx.Err(); err != nil {
			return Outcome{Rounds: completed, Last: last}, err
		}

		result, err := step(ctx, RoundContext{Round: round})
		if err != nil {
			return Outcome{Rounds: completed, Last: last}, err
		}

		completed = round
		last = result

		// Precedence: Stop > settled > stalemate > max
		if result.Stop != nil {
			return Outcome{Reason: "stop", Rounds: completed, Last: last}, nil
		}

		if result.Clean {
			cleanStreak++
		} else {
			cleanStreak = 0
			if result.Progress {
				noProgressStreak = 0
			} else {
				noProgressStreak++
			}
		}

		if cleanStreak >= lim.Settle {
			return Outcome{Reason: "settled", Rounds: completed, Last: last}, nil
		}
		if noProgressStreak >= lim.Patience {
			return Outcome{Reason: "stalemate", Rounds: completed, Last: last}, nil
		}
		if completed >= lim.Max {
			return Outcome{Reason: "max", Rounds: completed, Last: last}, nil
		}
	}
}
