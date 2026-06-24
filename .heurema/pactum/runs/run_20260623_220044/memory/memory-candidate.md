# Memory Candidate

## Run
- Run id: run_20260623_220044
- Source: deterministic

## Contract
- Goal: Slice 1: implement the policy-free engine on a mock backend. Package internal/engine/loop: a streak loop Run(ctx, Limits{Max,Settle,Patience}, Step) -> Outcome that drives rounds and decides settled/stalemate/max via consecutive clean/no-progress streaks. Package internal/engine/transport: Transport/Session/Spec/Result interfaces (Open->Send->Close) plus error classification, and a mock backend whose Session returns pre-scripted responses for tests. Package internal/engine/orchestrate: Participant, Turn, Transcript (with DeltaFor), a RoundRobin Scheduler, pluggable PromptBuilder and Verdict seams, a Config, and Run that wires loop+transport+transcript into round-robin rounds. Provide unit tests that drive a multi-participant debate on the mock backend with a trivial verdict, asserting turn order, transcript accumulation, delta visibility, and settled/stalemate/max outcomes. Out of scope: debate policy / signal schema, real acp/exec/api backends, CLI, personas, config discovery, synthesizer.
- In scope:
  - Implement internal/engine/loop with Limits{Max,Settle,Patience}, RoundContext, RoundResult{Clean,Progress,Stop}, Outcome, a Step func, and Run for settled/stalemate/max/stop plus step-error and ctx-cancellation paths.
  - Implement internal/engine/transport interfaces Transport, Session, Spec, Result, Usage, and a Classify error-classification function.
  - Implement a mock transport backend for tests whose sessions return pre-scripted results and errors and record sent prompts.
  - Implement internal/engine/orchestrate with Participant, Turn, Transcript including DeltaFor, a RoundRobin scheduler, pluggable PromptBuilder and Verdict seams, RenderMode, Config, and Run.
  - Add focused unit tests for loop behavior, mock transport behavior, transcript delta behavior, round-robin orchestration, and outcome propagation.
  - Add scripts/dep-guard.sh that enforces the engine dependency rule (stdlib and internal/engine only).
- Out of scope:
  - Debate-layer prompt policy, self-signal schema, signal parser, quorum/all_done verdict policy, or nudge behavior.
  - Real ACP, exec, API, network, subprocess, or model-backed transports.
  - CLI behavior, persona parsing, .heurema/debate discovery, config loading, synthesizer selection, or repository rename work.
  - Recovery, retry, degraded participant handling, live stderr streaming, telemetry, or production transport lifecycle policy.
- Acceptance criteria:
  - loop.Run(ctx, Limits{Max,Settle,Patience}, step) returns Outcome{Reason, Rounds, Last} where Reason is one of `settled`, `stalemate`, `max`, `stop`. Rounds are 1-based via RoundContext.Round. Per round, after calling step, precedence is checked in this exact order: (a) returned RoundResult.Stop != nil -> stop immediately, Reason `stop`; (b) else update streaks and if the consecutive-Clean streak reaches Settle -> Reason `settled`; (c) else if the consecutive-no-progress streak reaches Patience -> Reason `stalemate`; (d) else if Round == Max -> Reason `max`.
  - RoundResult has fields Clean bool, Progress bool, and Stop *Stop. Clean==true increments the clean streak and Clean==false resets it to 0. Progress is consulted only when Clean==false: Progress==true resets the no-progress streak, otherwise the no-progress streak increments. Stop != nil forces an immediate `stop` Outcome regardless of streaks.
  - Step is func(ctx, RoundContext) (RoundResult, error). If step returns a non-nil error, loop.Run stops immediately and returns that error together with an Outcome reflecting rounds completed so far; no further rounds run.
  - loop.Run checks ctx before each round; if ctx is already cancelled it returns ctx.Err() and an Outcome for rounds completed so far. ctx cancellation is an error path and is distinct from verdict/Stop-driven termination, so the two never conflict.
  - On any error path (step error or pre-round ctx cancellation), loop.Run returns the error and an Outcome whose Rounds equals the count of fully-completed rounds, whose Reason is the empty string (no terminal reason), and whose Last is the most recently completed RoundResult (the zero RoundResult if no round completed).
  - Limits.Max, Limits.Settle, and Limits.Patience must each be >= 1; loop.Run validates them before running any round and returns an error (running zero rounds, Outcome.Rounds == 0) if any is < 1.
  - loop unit tests cover: settled, stalemate, max, immediate Stop, step-error propagation, pre-round ctx cancellation, and invalid (< 1) limits.
  - transport exposes Transport.Open(ctx, Spec) (Session, error), Session.Send(ctx, prompt string) (Result, error), Session.Close() error, Result{Content string, Usage Usage}, a Usage struct of token counters, and a Spec carrying id, model, effort, system, read-only, and optional command. The transport package exports one sentinel error variable per named error kind: ErrRateLimit, ErrIdleTimeout, ErrTransportDrop, ErrServerError, ErrDeadline (all retryable) and ErrAuth, ErrClientError, ErrCanceled (all non-retryable). Classify(err error) ErrorClass{Retryable bool, Kind string} maps nil to a non-retryable Kind `none`; for non-nil errors it dispatches via errors.Is against the exported sentinels, so an error that is or wraps a sentinel resolves to the same Kind and Retryable as that sentinel; an error matching no sentinel resolves to non-retryable Kind `unknown`. Transport unit tests cover every named kind with both a bare sentinel and a sentinel wrapped via fmt.Errorf("%w", ErrXxx) as inputs, asserting the expected Kind and Retryable values; they also cover nil input (Kind `none`, Retryable false) and an unrecognized error (Kind `unknown`, Retryable false).
  - The mock backend implements Transport/Session for tests: each session returns pre-scripted Results (and optionally scripted errors) in order per Send, records every prompt it was sent, makes no external/network/subprocess calls, and Close is idempotent.
  - orchestrate defines Participant{ID, Session}, Turn{Round, Speaker, Content, Usage, Extra}, a Transcript (append plus ordered read), a RoundRobin(rotate bool) Scheduler, a PromptBuilder func(Participant, *Transcript, RoundContext, RenderMode) (string, error) seam, a Verdict interface { Assess(*Transcript, RoundContext) loop.RoundResult } seam, a RenderMode with at least Delta and Full, a Config, and Run(ctx, Config) (Result{Transcript, Outcome}, error).
  - orchestrate.Run drives loop.Run: each round it asks the Scheduler for the speaking order, and for each participant calls PromptBuilder (with that participant, the transcript, the RoundContext, and RenderMode Delta), sends the prompt via the participant Session, and appends a Turn to the transcript; after all turns in the round it calls Verdict.Assess(transcript, RoundContext) and returns its loop.RoundResult to the loop. Verdict-driven termination thus happens through RoundResult, never through ctx cancellation.
  - orchestrate.Run receives already-open Session values via Config.Participants and never calls Session.Open or Session.Close; the caller owns session lifecycle. If PromptBuilder or Session.Send returns an error during a round, orchestrate.Run surfaces it to loop.Run as the Step error (which stops the loop and is returned by orchestrate.Run); this slice performs no retry, recovery, or degraded-participant handling. Any turns already appended to the Transcript within the failing round before the error occurred are retained and included in the Result.Transcript returned to the caller; the partial round is not rolled back.
  - Config has fields Participants []Participant, Scheduler, Prompt (PromptBuilder), Verdict, Limits (loop.Limits), and an optional OnTurn callback; orchestrate.Run requires at least one Participant and non-nil Scheduler, Prompt, and Verdict, and returns an error before any round if a required field is missing or nil.
  - Transcript.DeltaFor(participantID) is read-only (it does not mutate cursors) and participant-relative: appending a participant own Turn advances that participant cursor to the current transcript end, so DeltaFor returns the turns appended by OTHER participants since that participant last spoke, excluding the participant own turns. Tests cover same-round visibility (earlier speakers in the current round) and next-round visibility.
  - A participant first turn is a normally-produced response appended like any other; whether a round is Clean or has Progress is decided solely by the injected Verdict (the engine defines no per-participant progress notion). Empty responses are still produced turns.
  - Multi-participant tests use only the mock backend and a trivial injected Verdict to assert round-robin turn order, transcript accumulation, DeltaFor delta visibility, and that orchestrate.Run yields settled, stalemate, and max outcomes.
  - internal/engine/... stays policy-free: it imports only the Go standard library and other internal/engine/... packages and never imports internal/debate, cmd/debate, persona/config/synthesizer code, or any real backend.
  - scripts/dep-guard.sh runs `go list -deps -test ./internal/engine/...`, succeeds when every resulting dependency is either a Go standard-library package or under github.com/heurema/debate/internal/engine, fails with a non-zero exit and prints the offending dependencies when any other dependency is present (internal/debate, cmd/debate, or any third-party module), and propagates a non-zero go list failure.
- Validation commands:
  - test -z "$(gofmt -l internal cmd scripts)"
  - go test -count=1 ./internal/engine/...
  - go test -count=1 ./...
  - go vet ./...
  - bash scripts/dep-guard.sh

## Outcome
- Gate status: needs_review
- Review status: approved
- Execution exit code: 0
- Validation passed: true
- Changes need review: true

## Changes
- Changed files:
  - internal/engine/loop/loop.go
  - internal/engine/orchestrate/orchestrate.go
  - internal/engine/transport/transport.go
- New files:
  - internal/engine/loop/loop_test.go
  - internal/engine/orchestrate/orchestrate_test.go
  - internal/engine/transport/mock/mock.go
  - internal/engine/transport/mock/mock_test.go
  - internal/engine/transport/transport_test.go
  - scripts/dep-guard.sh
- Missing files: none

## Clarifications
- None

## Review Decisions
- f_001 [low] open scripts/dep-guard.sh:5: dep-guard.sh captures go list's stderr into the dependency list on the success path via `2>&1`, so any stderr output from a successful `go list` (exit 0) is treated as a dependency line and flagged as a forbidden dependency, producing a false non-zero failure.
- f_002 [low] open scripts/dep-guard.sh:15: dep-guard.sh's allow-list grep uses an unanchored prefix `^github\.com/heurema/debate/internal/engine`, which matches more than the contract's 'under internal/engine' rule: any future package whose path merely starts with that string but is not actually under the engine subtree (e.g. `github.com/heurema/debate/internal/engineering`) would be wrongly treated as allowed and slip past the guard.
- f_003 [medium] open internal/engine/orchestrate/orchestrate.go:126: The PromptBuilder error path in orchestrate.Run is untested. The contract requires a PromptBuilder error to be surfaced to loop.Run as the Step error (alongside the Send error path), but no test injects a PromptBuilder error; only TestSessionSendErrorSurfaced covers the Session.Send error branch.
- f_004 [low] open internal/engine/orchestrate/orchestrate.go:53: Transcript.DeltaFor's documented read-only / cursor-non-mutating invariant is untested. Both delta tests call DeltaFor exactly once, immediately before that participant speaks, so a regression that advanced the cursor inside DeltaFor would be masked by the subsequent Append overwriting the cursor.
- f_005 [low] open internal/engine/orchestrate/orchestrate.go:106: The invariant that orchestrate.Run never opens or closes participant sessions (caller owns lifecycle) is untested, even though mock.Session exposes Closed() expressly to verify it.
- f_006 [low] open internal/engine/orchestrate/orchestrate_test.go:227: Test helper verdictAfter(n, fn) ignores its n parameter and merely wraps funcVerdict{fn}, adding nothing over using funcVerdict directly. Its doc comment claims it 'produces progress results until max rounds then switches', but funcVerdict.Assess unconditionally returns fn(rc.Round) with no such switching. The sole caller writes a `round` variable in the closure that is immediately discarded via `_ = round`.
- f_007 [low] open internal/engine/orchestrate/orchestrate.go:72: Transcript.Len() is an exported method with no callers anywhere in the slice (production or tests). The contract specifies a Transcript with append plus ordered read, which Append + All/DeltaFor already satisfy; Len is unused API surface.
- Proposal summary: pending=0 accepted=7 rejected=0

## Reusable Project Knowledge
- scope: in scope: Implement internal/engine/loop with Limits{Max,Settle,Patience}, RoundContext, RoundResult{Clean,Progress,Stop}, Outcome, a Step func, and Run for settled/stalemate/max/stop plus step-error and ctx-cancellation paths.
- scope: in scope: Implement internal/engine/transport interfaces Transport, Session, Spec, Result, Usage, and a Classify error-classification function.
- scope: in scope: Implement a mock transport backend for tests whose sessions return pre-scripted results and errors and record sent prompts.
- scope: in scope: Implement internal/engine/orchestrate with Participant, Turn, Transcript including DeltaFor, a RoundRobin scheduler, pluggable PromptBuilder and Verdict seams, RenderMode, Config, and Run.
- scope: in scope: Add focused unit tests for loop behavior, mock transport behavior, transcript delta behavior, round-robin orchestration, and outcome propagation.
- scope: in scope: Add scripts/dep-guard.sh that enforces the engine dependency rule (stdlib and internal/engine only).
- scope: out of scope: Debate-layer prompt policy, self-signal schema, signal parser, quorum/all_done verdict policy, or nudge behavior.
- scope: out of scope: Real ACP, exec, API, network, subprocess, or model-backed transports.
- scope: out of scope: CLI behavior, persona parsing, .heurema/debate discovery, config loading, synthesizer selection, or repository rename work.
- scope: out of scope: Recovery, retry, degraded participant handling, live stderr streaming, telemetry, or production transport lifecycle policy.
- review_resolution: proposal p_001 accepted as f_001
- review_resolution: proposal p_002 accepted as f_002
- review_resolution: proposal p_003 accepted as f_003
- review_resolution: proposal p_004 accepted as f_004
- review_resolution: proposal p_005 accepted as f_005
- review_resolution: proposal p_006 accepted as f_006
- review_resolution: proposal p_007 accepted as f_007
- validation: test -z "$(gofmt -l internal cmd scripts)" passed
- validation: go test -count=1 ./internal/engine/... passed
- validation: go test -count=1 ./... passed
- validation: go vet ./... passed
- validation: bash scripts/dep-guard.sh passed

## Artifacts
- Contract: contract/contract.json
- Gate report: gate/gate-report.json
- Review: review/review.json
- Findings: review/findings.jsonl
- Resolutions: review/resolutions.jsonl
- Proposals: review/proposals.jsonl
- Proposal decisions: review/proposal-decisions.jsonl
