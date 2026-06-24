# Contract Review: Testability

You are reviewing a software change contract through the **acceptance-testability** lens.

Review the contract fields below using only your assigned lens checklist.
Do not flag issues that belong to other lenses.

## Contract

**Goal**: Slice 2: implement the debate policy layer in internal/debate on top of the engine, exercised only with the mock backend. (1) internal/debate/signal: parse a structured signal {position string, objections []string, done bool} from a turn text — the speaker ends its reply with a fenced signal block (triple-backtick signal ... containing JSON); the parser extracts and validates it, returns a typed result plus a parsed-ok flag, and applies the invariant that done==true with non-empty objections is treated as done=false. (2) internal/debate/prompt: a PromptBuilder matching orchestrate.PromptBuilder that renders a per-turn user message = brief (task+context) + moderator rules-of-engagement + the delta board (rendered from Transcript.DeltaFor for the speaking participant) + the signal-format instruction; support RenderMode Delta and Full. (3) internal/debate/verdict: a Verdict matching orchestrate.Verdict that parses each round turns signals and returns loop.RoundResult where Clean = all speakers done (until=all_done) or a majority done (until=quorum), and Progress = the open-objection set changed vs previous round; configurable until in {all_done, quorum}; an unparsed signal makes that speaker not-done. Unit tests use only the mock backend with scripted signal-bearing turns: convergence after the settle streak, quorum, stalemate on a frozen objection set, max rounds, and unparsed-signal handling. internal/debate imports internal/engine only (one-way). Out of scope: real acp/exec/api backends, CLI, persona files, .heurema/debate discovery/config, synthesizer, and nudge-retry orchestration (parser only).

**Scope in**:
  - Implement internal/debate/signal: a Signal struct {Position string, Objections []string, Done bool} and Parse(content string) (Signal, bool) that extracts and validates the trailing fenced signal block.
  - Implement internal/debate/prompt: a constructor that returns an orchestrate.PromptBuilder rendering moderator rules + the brief (task+context) + the delta board + the signal-format instruction, honoring RenderMode Delta and Full.
  - Implement internal/debate/verdict: a type implementing orchestrate.Verdict, configurable with an until mode (all_done or quorum), computing Clean and Progress into a loop.RoundResult.
  - Add unit tests, using only the mock backend and scripted signal-bearing turns, for signal parsing, prompt rendering, and verdict-driven settled/quorum/stalemate/max/unparsed behavior end to end through orchestrate.Run.
  - Generalize scripts/dep-guard.sh to take a package pattern and a list of allowed non-stdlib import prefixes, and enforce the internal/debate dependency boundary with it.

**Scope out**:
  - Real ACP, exec, API, network, subprocess, or model-backed transports.
  - CLI behavior, persona file parsing, .heurema/debate discovery, config loading, or synthesizer selection/behavior.
  - Nudge-retry orchestration (re-prompting on an unparsed signal); the debate layer here only parses and judges.
  - Transport-level system-prompt delivery, recovery/retry, telemetry, or any modification of internal/engine source.

**Acceptance criteria**:
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

**Validation commands**:
  - test -z "$(gofmt -l internal cmd scripts)"
  - go test -count=1 ./internal/debate/...
  - go test -count=1 ./...
  - go vet ./...
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate
  - bash -c '! bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/debate'

**Assumptions**:
  - The signal block format is a fenced code block tagged `signal` whose body is a JSON object {position, objections, done}; Parse uses the last such block in the content.
  - The brief (task + context) is injected into the PromptBuilder at construction; the debate layer owns the moderator-rules text and the signal-format instruction text; their exact wording is an implementation detail.
  - quorum means strictly more than half of the current round's speakers have Done==true.
  - The open-objection set used for Progress is the set-union of Objections strings from the current round's parsed signals, compared by string-set equality (order and duplicates do not matter).
  - settle and stall_after live in loop.Limits and are set by the caller; the verdict only produces Clean/Progress and never sets Stop.
  - Nudge-retry on an unparsed signal is orchestration and is out of scope; the verdict treats an unparsed signal as a not-done speaker contributing no objections.
  - Tests inject a trivial or debate PromptBuilder and scripted mock sessions; no real model, network, or subprocess is used.
  - internal/debate may import internal/engine (the one-way dependency debate -> engine); it must not be imported by internal/engine.

## Lens: Testability

Checklist:
- Is each acceptance criterion backed by or expressible as a runnable validation command (not just prose)?
- Are any criteria purely prose with no machine-checkable outcome?

## Output

Report likely-real defects (recall-first), then gate on precision before marking blocking.
Use state=candidate with explicit uncertainty when you believe a finding is real but have not fully confirmed it.

State your analysis in prose. If you find issues, also include a structured block:

```json
{
  "schema": "pactum.contract_reviewer_result.v1alpha1",
  "findings": [
    {
      "message": "Describe the contract issue clearly.",
      "severity": "medium",
      "category": "quality",
      "blocking": true,
      "evidence": "Quote or cite the contract field that shows the issue.",
      "material_impact": "Concrete way this spec defect would make the implementation wrong, ambiguous, or stuck.",
      "fix_direction": "What the contract author should change to resolve this.",
      "uncertainty": "Any doubt about this finding — omit if confident.",
      "state": "candidate"
    }
  ]
}
```

Rules:
- Use severity: low, medium, high, critical.
- Use category: correctness, scope, quality, validation, process, other.
- Omit file and line (not applicable for contract review).
- Set state=candidate when likely real but not fully confirmed; set state=confirmed when certain.
- HARD RULE: blocking=true is allowed ONLY for a material spec defect that would make the implementation wrong, ambiguous, or stuck.
- Wording, style, naming, redundancy, and completeness/thoroughness preferences MUST be blocking=false (advisory).
- Every blocking finding MUST include a concrete material_impact explaining the implementation consequence.
- If you cannot state a concrete material_impact, mark the finding blocking=false (advisory).
- Set blocking=false for advisory issues.
- If no issues, say so clearly. Do not include an empty findings block.
