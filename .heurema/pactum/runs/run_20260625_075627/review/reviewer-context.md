# Reviewer Context

## Run
- Run id: run_20260625_075627
- Run status: contract_approved

## Contract
- Goal: Make participant turn prompts chat-like by giving every debater the full transcript accumulated so far on every turn, while keeping synthesis as a single post-loop step.
- In scope:
  - internal/engine/orchestrate: change participant-turn prompt rendering from Delta mode to Full transcript mode in the debate loop
  - internal/engine/orchestrate tests: prove participant prompts receive Full mode and see all prior turns, including their own earlier turns
  - internal/debate/prompt tests/comments as needed: preserve and document Full-mode rendering of all transcript turns
  - internal/debate/runner or CLI tests as needed: prove the synthesizer is still called once after the debate loop and receives the complete transcript
- Out of scope:
  - Do not add a CLI flag, config option, or persona option for context mode in this slice
  - Do not change RoundRobin scheduling, fixed participant order, verdict semantics, settle/patience defaults, or max-round behavior
  - Do not invoke the synthesizer during each round or add per-round synthesis
  - Do not implement transcript persistence, run storage, README expansion, backend changes, or model/persona changes
- Acceptance criteria:
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
- Validation commands:
  - bash scripts/check-gofmt.sh
  - go test -count=1 ./internal/engine/orchestrate ./internal/debate/prompt ./internal/debate/runner ./cmd/debate
  - bash -c 'set -euo pipefail; rg --version >/dev/null; forbidden="(ContextMode|TranscriptMode|PromptMode|HistoryMode|DeltaMode|FullMode|contextMode|transcriptMode|promptMode|historyMode|deltaMode|fullMode|context[[:space:]_-]?mode|transcript[[:space:]_-]?mode|prompt[[:space:]_-]?mode|history[[:space:]_-]?mode|delta[[:space:]_-]?mode|full[[:space:]_-]?mode)"; set +e; rg -n --glob "!*_test.go" "\"[^\"]*$forbidden[^\"]*\"" ./cmd/debate; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -gt 1 ]; then exit "$status"; fi; set +e; rg -n --glob "!*_test.go" "(yaml|json|toml|mapstructure):\"[^\"]*$forbidden[^\"]*\"|frontmatter[^\n]*$forbidden|$forbidden[^\n]*frontmatter" ./cmd/debate ./internal/debate/config ./internal/debate/persona ./internal/debate/runner ./internal/engine/orchestrate; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -gt 1 ]; then exit "$status"; fi; set +e; rg -n --glob "!*_test.go" "^(type|func|const|var)[[:space:]]+[A-Z][A-Za-z0-9_]*.*$forbidden|^[[:space:]]+[A-Z][A-Za-z0-9_]*[[:space:]].*$forbidden" ./internal/debate/runner ./internal/engine/orchestrate ./internal/debate/persona; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -eq 1 ]; then exit 0; else exit "$status"; fi'
  - go test -count=1 ./...
  - go vet ./...
  - go build ./cmd/debate

## Accepted memory
- Memory context: context/memory-context.md
- Selected items: 5
- Fresh: 5
- Stale: 0
- Unknown: 0
- Stale memory may be outdated and must be verified.

## Gate report
- Gate status: needs_review
- Execution attempt id: attempt_001
- Execution exit code: 0
- Validation command results:
  - command_001: bash scripts/check-gofmt.sh (exit 0, timed out: false, result: gate/validation/command_001/result.json)
  - command_002: go test -count=1 ./internal/engine/orchestrate ./internal/debate/prompt ./internal/debate/runner ./cmd/debate (exit 0, timed out: false, result: gate/validation/command_002/result.json)
  - command_003: bash -c 'set -euo pipefail; rg --version >/dev/null; forbidden="(ContextMode|TranscriptMode|PromptMode|HistoryMode|DeltaMode|FullMode|contextMode|transcriptMode|promptMode|historyMode|deltaMode|fullMode|context[[:space:]_-]?mode|transcript[[:space:]_-]?mode|prompt[[:space:]_-]?mode|history[[:space:]_-]?mode|delta[[:space:]_-]?mode|full[[:space:]_-]?mode)"; set +e; rg -n --glob "!*_test.go" "\"[^\"]*$forbidden[^\"]*\"" ./cmd/debate; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -gt 1 ]; then exit "$status"; fi; set +e; rg -n --glob "!*_test.go" "(yaml|json|toml|mapstructure):\"[^\"]*$forbidden[^\"]*\"|frontmatter[^\n]*$forbidden|$forbidden[^\n]*frontmatter" ./cmd/debate ./internal/debate/config ./internal/debate/persona ./internal/debate/runner ./internal/engine/orchestrate; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -gt 1 ]; then exit "$status"; fi; set +e; rg -n --glob "!*_test.go" "^(type|func|const|var)[[:space:]]+[A-Z][A-Za-z0-9_]*.*$forbidden|^[[:space:]]+[A-Z][A-Za-z0-9_]*[[:space:]].*$forbidden" ./internal/debate/runner ./internal/engine/orchestrate ./internal/debate/persona; status=$?; set -e; if [ "$status" -eq 0 ]; then exit 1; elif [ "$status" -eq 1 ]; then exit 0; else exit "$status"; fi' (exit 0, timed out: false, result: gate/validation/command_003/result.json)
  - command_004: go test -count=1 ./... (exit 0, timed out: false, result: gate/validation/command_004/result.json)
  - command_005: go vet ./... (exit 0, timed out: false, result: gate/validation/command_005/result.json)
  - command_006: go build ./cmd/debate (exit 0, timed out: false, result: gate/validation/command_006/result.json)
- Change summary:
  - changed files:
    - internal/debate/prompt/prompt_test.go
    - internal/debate/runner/runner_test.go
    - internal/engine/orchestrate/orchestrate.go
    - internal/engine/orchestrate/orchestrate_test.go
  - new files:
    - none
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
- If you are not certain an issue is real after verification, do not flag it.
