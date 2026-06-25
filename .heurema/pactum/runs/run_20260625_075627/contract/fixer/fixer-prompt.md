# Contract Review Fixer Prompt

You are fixing a software change contract to address blocking review findings.

Current contract version: 5219b5b9f616e2ed6445155a81a810c7a822f7f79d33cf5a06638f2279f35717

## Current Contract

**Goal**: Make participant turn prompts chat-like by giving every debater the full transcript accumulated so far on every turn, while keeping synthesis as a single post-loop step.

**Scope in**:
  - internal/engine/orchestrate: change participant-turn prompt rendering from Delta mode to Full transcript mode in the debate loop
  - internal/engine/orchestrate tests: prove participant prompts receive Full mode and see all prior turns, including their own earlier turns
  - internal/debate/prompt tests/comments as needed: preserve and document Full-mode rendering of all transcript turns
  - internal/debate/runner or CLI tests as needed: prove the synthesizer is still called once after the debate loop and receives the complete transcript

**Scope out**:
  - Do not add a CLI flag, config option, or persona option for context mode in this slice
  - Do not change RoundRobin scheduling, fixed participant order, verdict semantics, settle/patience defaults, or max-round behavior
  - Do not invoke the synthesizer during each round or add per-round synthesis
  - Do not implement transcript persistence, run storage, README expansion, backend changes, or model/persona changes

**Acceptance criteria**:
  - Runtime debate behavior is a sequential shared-chat transcript: before each participant responds, their prompt includes the complete committed debate transcript available at that moment, then their response is appended to that same transcript for subsequent participants.
  - Each participant turn prompt is built in Full transcript mode during orchestrate.Run, not Delta mode.
  - Delta and DeltaFor may remain as internal helpers or test utilities, but the normal debate runtime must not use Delta mode for participant turn prompts.
  - On a later turn, a participant receives all prior transcript turns available at that moment, including that participant's own earlier turns and other participants' turns from previous and current rounds.
  - Participants still cannot see future turns that have not happened yet; the transcript is full only up to the current turn construction point.
  - A runnable unit test explicitly proves future-turn exclusion by asserting that a participant prompt does not include transcript turns generated after that prompt was constructed.
  - A runnable prompt-rendering unit test or golden snapshot explicitly asserts that a participant prompt still contains the existing moderator rules text, debate brief text, discussion board/transcript block, round and speaker labels on transcript entries, and signal instruction text; the test must fail if any of those sections, labels, or instructions are omitted or renamed.
  - Synthesizer execution remains outside the debate loop: it is opened/sent exactly once after orchestrate.Run returns, using the final transcript.
  - Round ordering remains fixed RoundRobin(false), and a runnable orchestrate unit test must fail if participant order accidentally rotates during the debate loop. Existing rotation-helper tests alone are not sufficient for this acceptance criterion.
  - No public CLI/API/config surface is added for context mode in this slice: no user-facing flag, config YAML/frontmatter field or tag, exported runner/orchestrate option, or persona setting named context mode, transcript mode, prompt mode, history mode, delta mode, or full mode is introduced in common spellings including PascalCase, lowerCamel, kebab-case, or snake_case. This prohibition does not ban internal implementation identifiers or test/golden text needed to prove Full transcript rendering, as long as they are not exposed as CLI/API/config/persona surface.
  - Relevant unit tests are added or updated so the old delta-only participant runtime behavior would fail the suite.

**Validation commands**:
  - bash scripts/check-gofmt.sh
  - go test -count=1 ./internal/engine/orchestrate ./internal/debate/prompt ./internal/debate/runner ./cmd/debate
  - bash -c 'set -euo pipefail; rg --version >/dev/null; forbidden="(ContextMode|TranscriptMode|PromptMode|HistoryMode|DeltaMode|FullMode|contextMode|transcriptMode|promptMode|historyMode|deltaMode|fullMode|context[-_]?mode|transcript[-_]?mode|prompt[-_]?mode|history[-_]?mode|delta[-_]?mode|full[-_]?mode)"; set +e; rg -n --glob "!*_test.go" "\"[^\"]*$forbidden[^\"]*\"" ./cmd/debate; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -gt 1 ]; then exit "$status"; fi; set +e; rg -n --glob "!*_test.go" "(yaml|json|toml|mapstructure):\"[^\"]*$forbidden[^\"]*\"|frontmatter[^\n]*$forbidden|$forbidden[^\n]*frontmatter" ./cmd/debate ./internal/debate/config ./internal/debate/persona ./internal/debate/runner ./internal/engine/orchestrate; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -gt 1 ]; then exit "$status"; fi; set +e; rg -n --glob "!*_test.go" "^(type|func|const|var)[[:space:]]+[A-Z][A-Za-z0-9_]*.*$forbidden|^[[:space:]]+[A-Z][A-Za-z0-9_]*[[:space:]].*$forbidden" ./internal/debate/runner ./internal/engine/orchestrate ./internal/debate/persona; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -eq 1 ]; then exit 0; else exit "$status"; fi'
  - go test -count=1 ./...
  - go vet ./...
  - go build ./cmd/debate

**Assumptions**:
  - The requested chat-like behavior means a sequential shared-chat transcript with full committed transcript so far, not a simultaneous shared-room state and not future turns in the same round.
  - For non-happy-path participant behavior, committed transcript turns are defined by append semantics: only the participant response actually appended to the transcript is committed. Failed participant calls that abort, skipped turns that do not append, and failed retry attempts are not committed and must not appear in later full prompts. If a retry eventually succeeds, only the final successful appended response is committed. An empty response is committed only if existing runtime behavior accepts and appends it as a transcript turn; otherwise it is not committed.
  - The higher token cost from resending full transcript context is acceptable for this slice and is intentionally not made configurable yet.
  - Existing backend session history behavior remains unchanged; correctness must not depend on a backend remembering prior turns because the prompt now carries the full transcript so far.
  - Delta and DeltaFor are not required to be removed by this contract; only the normal debate runtime participant prompt mode changes.
  - The synthesizer is a final summarizer, not a per-round moderator.

## Blocking Findings to Address

1. [codex-xhigh/validation-soundness] The public-surface validation grep misses the exact space-separated names prohibited by the acceptance criteria, such as "context mode", "delta mode", and "full mode".
   Evidence: Acceptance criteria prohibit names including "context mode, transcript mode, prompt mode, history mode, delta mode, or full mode", but the validation command uses alternatives like `context[-_]?mode`, `delta[-_]?mode`, and `full[-_]?mode`, which do not match a space.
2. [codex-xhigh/assumptions-surfaced] The contract should explicitly state whether final synthesis is expected only after a successful orchestrate.Run, or also after runs that return an error or partial transcript. This affects implementation and test behavior for failed participant calls, skipped turns, and retry exhaustion.
   Evidence: Acceptance criteria: "Synthesizer execution remains outside the debate loop: it is opened/sent exactly once after orchestrate.Run returns, using the final transcript." Assumptions define failed participant append semantics but do not define whether synthesis runs after an aborted/error return.

## Fixer Instructions

- Address each blocking finding by updating the relevant contract field.
- Do NOT change the goal field — it is out of scope for the fixer.
- Only include the contract fields you are changing in the output.
- base_version must exactly match the version shown above.

## Output

Output your reasoning, then a single JSON block with the revise payload:

```json
{
  "schema": "pactum.contract_revise.v1alpha1",
  "base_version": "5219b5b9f616e2ed6445155a81a810c7a822f7f79d33cf5a06638f2279f35717",
  "contract": {
    "acceptance_criteria": ["...updated criteria..."],
    "validation": {"commands": ["...updated commands..."]}
  }
}
```

Omit any contract field you are not changing. Do not include the goal field.
