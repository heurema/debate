# Reviewer Context

## Run
- Run id: run_20260624_085539
- Run status: contract_approved

## Contract
- Goal: Slice 2: implement the debate policy layer in internal/debate on top of the engine, exercised only with the mock backend. (1) internal/debate/signal: parse a structured signal {position string, objections []string, done bool} from a turn text — the speaker ends its reply with a fenced signal block (triple-backtick signal ... containing JSON); the parser extracts and validates it, returns a typed result plus a parsed-ok flag, and applies the invariant that done==true with non-empty objections is treated as done=false. (2) internal/debate/prompt: a PromptBuilder matching orchestrate.PromptBuilder that renders a per-turn user message = brief (task+context) + moderator rules-of-engagement + the delta board (rendered from Transcript.DeltaFor for the speaking participant) + the signal-format instruction; support RenderMode Delta and Full. (3) internal/debate/verdict: a Verdict matching orchestrate.Verdict that parses each round turns signals and returns loop.RoundResult where Clean = all speakers done (until=all_done) or a majority done (until=quorum), and Progress = the open-objection set changed vs previous round; configurable until in {all_done, quorum}; an unparsed signal makes that speaker not-done. Unit tests use only the mock backend with scripted signal-bearing turns: convergence after the settle streak, quorum, stalemate on a frozen objection set, max rounds, and unparsed-signal handling. internal/debate imports internal/engine only (one-way). Out of scope: real acp/exec/api backends, CLI, persona files, .heurema/debate discovery/config, synthesizer, and nudge-retry orchestration (parser only).
- In scope:
  - Implement internal/debate/signal: a Signal struct {Position string, Objections []string, Done bool} and Parse(content string) (Signal, bool) that extracts and validates the trailing fenced signal block.
  - Implement internal/debate/prompt: a constructor that returns an orchestrate.PromptBuilder rendering moderator rules + the brief (task+context) + the delta board + the signal-format instruction, honoring RenderMode Delta and Full.
  - Implement internal/debate/verdict: a type implementing orchestrate.Verdict, configurable with an until mode (all_done or quorum), computing Clean and Progress into a loop.RoundResult.
  - Add unit tests, using only the mock backend and scripted signal-bearing turns, for signal parsing, prompt rendering, and verdict-driven settled/quorum/stalemate/max/unparsed behavior end to end through orchestrate.Run.
  - Generalize scripts/dep-guard.sh to take a package pattern and a list of allowed non-stdlib import prefixes, and enforce the internal/debate dependency boundary with it.
- Out of scope:
  - Real ACP, exec, API, network, subprocess, or model-backed transports.
  - CLI behavior, persona file parsing, .heurema/debate discovery, config loading, or synthesizer selection/behavior.
  - Nudge-retry orchestration (re-prompting on an unparsed signal); the debate layer here only parses and judges.
  - Transport-level system-prompt delivery, recovery/retry, telemetry, or any modification of internal/engine source.
- Acceptance criteria:
  - signal.Signal is a struct {Position string, Objections []string, Done bool}. signal.Parse(content string) (Signal, bool) scans content for fenced code blocks tagged `signal` (a line ```signal followed by JSON and a closing ```), uses the LAST such block, unmarshals its JSON object {position, objections, done} into a Signal, and returns (Signal, true) on success. The convention is that speakers place the signal block at the end of their reply, but Parse does not require this: any text appearing after the last signal block is permitted and does not prevent a successful parse.
  - signal.Parse returns (zero Signal, false) when there is no `signal` block or the block's content cannot be decoded as a JSON object into a Signal (for example, the body is not valid JSON or the top-level JSON value is not a JSON object); missing or null fields fall back to their zero values (empty string for position, null or absent objections treated as an empty slice, false for done), so a well-formed JSON object such as `{}` is a valid parse returning a zero-value Signal with Done==false; non-`signal` fenced blocks are ignored and surrounding prose — both before and after the last signal block — does not affect parsing.
  - signal.Parse enforces the invariant that a parsed signal with Done==true and len(Objections)>0 is returned with Done set to false (an agent cannot be done while it still lists blocking objections).
  - prompt provides a constructor (for example NewPromptBuilder(brief)) returning a value of type orchestrate.PromptBuilder; the brief (task and context text) is supplied at construction, not read from globals, so the engine stays policy-free.
  - The returned PromptBuilder renders, for the speaking participant p, a single string containing: the moderator rules-of-engagement, the shared brief (task + context), the rendered board, and an explicit instruction to end the reply with the signal block format.
  - In RenderMode Delta the board is built from t.DeltaFor(p.ID) (only other participants' turns since p last spoke); in RenderMode Full it is built from the whole transcript (t.All()). Each rendered turn is labelled with its Speaker and Round.
  - verdict provides a constructor taking an until mode (all_done or quorum) and returns a type implementing orchestrate.Verdict, whose Assess(t *Transcript, rc loop.RoundContext) loop.RoundResult judges the turns of the current round (those with Turn.Round == rc.Round, obtained from t.All()).
  - Assess sets RoundResult.Clean = (until==all_done: every speaker in the current round has a parsed signal with Done==true) or (until==quorum: strictly more than half of the current round's speakers have Done==true); a speaker whose signal does not parse counts as not-done. RoundResult.Stop is always nil (loop streak logic owns termination).
  - Assess sets RoundResult.Progress (consulted by the loop only when Clean==false) to true iff the set of open objections this round differs from the previous round's; the verdict tracks the previous round's open-objection set across calls. On the first call, when no previous round has been recorded, the baseline open-objection set is treated as the empty set, so Progress is true iff the current round's open-objection set is non-empty. The open-objection set is the union of Objections from the current round's parsed signals, compared as a set of strings (order and duplicates do not matter).
  - Multi-participant tests using only the mock backend and a trivial PromptBuilder assert: scripted all-done signals drive loop.Run to a `settled` Outcome after the Settle streak; quorum mode settles on a majority; a frozen non-empty objection set with nobody done drives `stalemate` after Patience; neither condition reaching its threshold drives `max`; and a turn with an unparsed signal is treated as not-done.
  - signal, prompt, and verdict tests cover: a well-formed signal block at the end of reply content, a signal block followed by trailing prose (confirming the last-block rule applies regardless of text after the block), a missing/garbled block, the Done+Objections invariant, Delta vs Full rendering, all_done vs quorum, the Progress objection-set comparison, and the unparsed-signal path.
  - internal/debate depends only on the Go standard library, internal/engine/..., and internal/debate/...; it never imports cmd/debate or any real backend, and never modifies internal/engine source.
  - scripts/dep-guard.sh is generalized to accept a package pattern argument and one or more allowed non-stdlib import-path prefixes, succeeds when every dependency of the pattern (via go list -deps -test) is either standard-library or under an allowed prefix, fails (printing offenders, non-zero exit) otherwise, and propagates a go list failure; validation invokes it in two positive modes — internal/engine (allowed: internal/engine) and internal/debate (allowed: internal/engine, internal/debate) — and in one negative mode where internal/debate is checked with only internal/debate allowed (excluding internal/engine), confirming the script exits non-zero and prints the offending import.
- Validation commands:
  - test -z "$(gofmt -l internal cmd scripts)"
  - go test -count=1 ./internal/debate/...
  - go test -count=1 ./...
  - go vet ./...
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate
  - bash -c '! bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/debate'

## Accepted memory
- Memory context: context/memory-context.md
- Selected items: 0
- Fresh: 0
- Stale: 0
- Unknown: 0
- Stale memory may be outdated and must be verified.

## Gate report
- Gate status: needs_review
- Execution attempt id: attempt_001
- Execution exit code: 0
- Validation command results:
  - command_001: test -z "$(gofmt -l internal cmd scripts)" (exit 0, timed out: false, result: gate/validation/command_001/result.json)
  - command_002: go test -count=1 ./internal/debate/... (exit 0, timed out: false, result: gate/validation/command_002/result.json)
  - command_003: go test -count=1 ./... (exit 0, timed out: false, result: gate/validation/command_003/result.json)
  - command_004: go vet ./... (exit 0, timed out: false, result: gate/validation/command_004/result.json)
  - command_005: bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine (exit 0, timed out: false, result: gate/validation/command_005/result.json)
  - command_006: bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate (exit 0, timed out: false, result: gate/validation/command_006/result.json)
  - command_007: bash -c '! bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/debate' (exit 0, timed out: false, result: gate/validation/command_007/result.json)
- Change summary:
  - changed files:
    - scripts/dep-guard.sh
  - new files:
    - internal/debate/prompt/prompt.go
    - internal/debate/prompt/prompt_test.go
    - internal/debate/signal/signal.go
    - internal/debate/signal/signal_test.go
    - internal/debate/verdict/verdict.go
    - internal/debate/verdict/verdict_test.go
  - missing files:
    - none

## Existing manual review
- Review status: pending
- Current findings summary: findings=0 open=0 resolved=0 blocking_open=0
- Existing findings:
  - none
- Existing resolutions:
  - none
- Proposal summary: pending=0 accepted=0 rejected=0
- Existing proposals:
  - none

## Artifacts
- Contract: contract/contract.json
- Gate report: gate/gate-report.json
- Review: review/review.json
- Findings: review/findings.jsonl
- Resolutions: review/resolutions.jsonl
- Proposals: review/proposals.jsonl
- Proposal decisions: review/proposal-decisions.jsonl
- Execution result: execute/last-result.json

## Reviewer guidance
- This context is not complete semantic truth.
- Use `pactum search "<term>"` and inspect files before proposing findings.
- Do not invent changes.
- Do not approve automatically.
- Report every issue you believe is likely real: use state=candidate for uncertain findings and drop only when trigger, evidence, and fix_direction cannot be filled concretely.
