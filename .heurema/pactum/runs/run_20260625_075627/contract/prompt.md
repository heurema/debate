# Executor Prompt

This prompt is prepared from an approved Pactum contract.
This prompt is prepared for the selected built-in agent when `pactum execute run` is used.
Pactum records execution artifacts and validates contract, map, and memory boundaries before execution.

## Contract status
- Run: run_20260625_075627
- Approval: approved
- Contract hash: 7fe856523097008e83726679d2a916a09ba9d8cf1d893a34d572bd5958d2b743

## Goal
Make participant turn prompts chat-like by giving every debater the full transcript accumulated so far on every turn, while keeping synthesis as a single post-loop step.

## In scope
- internal/engine/orchestrate: change participant-turn prompt rendering from Delta mode to Full transcript mode in the debate loop
- internal/engine/orchestrate tests: prove participant prompts receive Full mode and see all prior turns, including their own earlier turns
- internal/debate/prompt tests/comments as needed: preserve and document Full-mode rendering of all transcript turns
- internal/debate/runner or CLI tests as needed: prove the synthesizer is still called once after the debate loop and receives the complete transcript

## Out of scope
- Do not add a CLI flag, config option, or persona option for context mode in this slice
- Do not change RoundRobin scheduling, fixed participant order, verdict semantics, settle/patience defaults, or max-round behavior
- Do not invoke the synthesizer during each round or add per-round synthesis
- Do not implement transcript persistence, run storage, README expansion, backend changes, or model/persona changes

## Acceptance criteria
- Runtime debate behavior is a sequential shared-chat transcript: before each participant responds, their prompt includes the complete committed debate transcript available at that moment, then their response is appended to that same transcript for subsequent participants.
- Each participant turn prompt is built in Full transcript mode during orchestrate.Run, not Delta mode.
- Delta and DeltaFor may remain as internal helpers or test utilities, but the normal debate runtime must not use Delta mode for participant turn prompts.
- On a later turn, a participant receives all prior transcript turns available at that moment, including that participant's own earlier turns and other participants' turns from previous and current rounds.
- Participants still cannot see future turns that have not happened yet; the transcript is full only up to the current turn construction point.
- A runnable unit test explicitly proves future-turn exclusion by asserting that a participant prompt does not include transcript turns generated after that prompt was constructed.
- A runnable prompt-rendering unit test or golden snapshot explicitly asserts that a participant prompt still contains the existing moderator rules text, debate brief text, discussion board/transcript block, round and speaker labels on transcript entries, and signal instruction text; the test must fail if any of those sections, labels, or instructions are omitted or renamed.
- Synthesizer execution remains outside the debate loop: it is opened/sent exactly once only after orchestrate.Run completes successfully, using the final completed transcript.
- If orchestrate.Run returns an error or aborts before completing the debate loop, including from failed participant calls or retry exhaustion, the synthesizer must not be opened or sent with a partial transcript.
- Round ordering remains fixed RoundRobin(false), and a runnable orchestrate unit test must fail if participant order accidentally rotates during the debate loop. Existing rotation-helper tests alone are not sufficient for this acceptance criterion.
- No public CLI/API/config surface is added for context mode in this slice: no user-facing flag, config YAML/frontmatter field or tag, exported runner/orchestrate option, or persona setting named context mode, transcript mode, prompt mode, history mode, delta mode, or full mode is introduced in common spellings including PascalCase, lowerCamel, space-separated, kebab-case, or snake_case. This prohibition does not ban internal implementation identifiers or test/golden text needed to prove Full transcript rendering, as long as they are not exposed as CLI/API/config/persona surface.
- Relevant unit tests are added or updated so the old delta-only participant runtime behavior would fail the suite.

## Validation commands
- bash scripts/check-gofmt.sh
- go test -count=1 ./internal/engine/orchestrate ./internal/debate/prompt ./internal/debate/runner ./cmd/debate
- bash -c 'set -euo pipefail; rg --version >/dev/null; forbidden="(ContextMode|TranscriptMode|PromptMode|HistoryMode|DeltaMode|FullMode|contextMode|transcriptMode|promptMode|historyMode|deltaMode|fullMode|context[[:space:]_-]?mode|transcript[[:space:]_-]?mode|prompt[[:space:]_-]?mode|history[[:space:]_-]?mode|delta[[:space:]_-]?mode|full[[:space:]_-]?mode)"; set +e; rg -n --glob "!*_test.go" "\"[^\"]*$forbidden[^\"]*\"" ./cmd/debate; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -gt 1 ]; then exit "$status"; fi; set +e; rg -n --glob "!*_test.go" "(yaml|json|toml|mapstructure):\"[^\"]*$forbidden[^\"]*\"|frontmatter[^\n]*$forbidden|$forbidden[^\n]*frontmatter" ./cmd/debate ./internal/debate/config ./internal/debate/persona ./internal/debate/runner ./internal/engine/orchestrate; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -gt 1 ]; then exit "$status"; fi; set +e; rg -n --glob "!*_test.go" "^(type|func|const|var)[[:space:]]+[A-Z][A-Za-z0-9_]*.*$forbidden|^[[:space:]]+[A-Z][A-Za-z0-9_]*[[:space:]].*$forbidden" ./internal/debate/runner ./internal/engine/orchestrate ./internal/debate/persona; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -eq 1 ]; then exit 0; else exit "$status"; fi'
- go test -count=1 ./...
- go vet ./...
- go build ./cmd/debate

## Assumptions
- The requested chat-like behavior means a sequential shared-chat transcript with full committed transcript so far, not a simultaneous shared-room state and not future turns in the same round.
- For non-happy-path participant behavior, committed transcript turns are defined by append semantics: only the participant response actually appended to the transcript is committed. Failed participant calls that abort, skipped turns that do not append, and failed retry attempts are not committed and must not appear in later full prompts. If a retry eventually succeeds, only the final successful appended response is committed. An empty response is committed only if existing runtime behavior accepts and appends it as a transcript turn; otherwise it is not committed.
- The higher token cost from resending full transcript context is acceptable for this slice and is intentionally not made configurable yet.
- Existing backend session history behavior remains unchanged; correctness must not depend on a backend remembering prior turns because the prompt now carries the full transcript so far.
- Delta and DeltaFor are not required to be removed by this contract; only the normal debate runtime participant prompt mode changes.
- The synthesizer is a final summarizer, not a per-round moderator.

## Clarifications
- None

## Project context
- Executor context: context/executor-context.md
- Repo map: .heurema/pactum/map/repo-map.md
- Search results: context/search-results.json
- Accepted memory context: context/memory-context.md

## Accepted memory

Memory context:
- context/memory-context.md

Selected memory:
- total: 5
- fresh: 5
- stale: 0
- unknown: 0

Items:
- mem_005 [fresh] score=44 — Slice 3: implement persona loading, .heurema/debate workspace discovery, conf...
- mem_006 [fresh] score=43 — Slice 4: wire the cmd/debate CLI into a working debate on a deterministic off...
- mem_004 [fresh] score=39 — Slice 2: implement the debate policy layer in internal/debate on top of the e...
- mem_002 [fresh] score=36 — Slice 0: bootstrap the debate Go project skeleton — go.mod (module github.com...
- mem_003 [fresh] score=36 — Slice 1: implement the policy-free engine on a mock backend. Package internal...

Rules:
- Accepted memory is context, not semantic truth.
- Stale memory may be outdated; verify before using.
- Use `pactum search "<term>"` and inspect current source files before relying on memory.
- Do not implement from memory alone.

## Instructions for future executor
- Follow the approved contract.
- Do not implement out-of-scope work.
- Search before creating new code.
- Prefer existing code items when applicable.
- If the contract is ambiguous, stop and request clarification.
- Use the listed validation commands as expected checks.
- Pactum gate can run approved validation commands after execution.

## House style
- Match the surrounding code: idiom, naming, comment density.
- Comment only where the code is not self-explanatory; do not narrate the obvious.
- Search for and reuse existing helpers before writing new ones.
- Keep the diff small and focused: change only what the contract requires.
- Simplicity first: no enterprise patterns for simple problems, question every new abstraction, no premature generalization or optimization.
- Over-engineering DON'Ts: wrappers that add nothing, factories or abstractions for a single case, unused extension points, dual implementations where the old path has no callers, silent fallbacks that hide failures.
- No dead code, no commented-out code, no unused parameters.
- Handle errors per the project's existing convention; no silent failures.
- Tests verify behavior, not implementation details, and cover error paths.
- Fake-test DON'Ts: always-pass tests, hardcoded-value checks, assertions on mock behavior instead of the code under test, ignored errors, commented-out cases.
