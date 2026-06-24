# Reviewer Context

## Run
- Run id: run_20260624_161104
- Run status: contract_approved

## Contract
- Goal: Invoke agy non-interactively via --print in the exec backend so it works against the real agy CLI (which otherwise defaults to an interactive session and hangs).
- In scope:
  - Change the default agy argv in internal/backend/exec to include --print so agy runs a single prompt non-interactively from stdin and exits.
  - Update the affected exec unit tests (argv assertions) and the gated integration test to match the new argv.
- Out of scope:
  - Other backends, internal/engine, the internal/debate packages, and the acp backend.
  - The exec backend's stdin reconstruction / delta accumulation logic, error handling, grounding, and recovery.
  - CLI flag-ordering / argument parsing (handled separately).
- Acceptance criteria:
  - The exec backend's default agy argv is [agy, "--print", "--model", spec.Model]; the --print flag (alias of -p) makes agy run a single prompt non-interactively, reading the prompt from stdin and printing the response before exiting. The prompt is still written to the subprocess stdin (the reconstruction logic is unchanged).
  - The DEBATE_AGY_COMMAND override still replaces only the executable token (argv[0]) and preserves the --print and --model arguments in order.
  - spec.Model must still be non-empty (fail-fast otherwise); the model is passed as the --model value.
  - Unit tests assert the new default argv (including --print) for both the default and the DEBATE_AGY_COMMAND-overridden command.
  - The gated integration test (enabled by setting DEBATE_EXEC_INTEGRATION=1) exercises the real agy --print invocation end-to-end — agy reads the prompt from stdin, prints the response, and exits without hanging — thereby serving as the runnable validation of the --print non-interactive behavior.
  - check-gofmt, go build ./..., go vet ./..., and go test ./... pass; go.mod and go.sum are unchanged (no new dependency); internal/engine is unchanged and the engine/exec/backend/debate dep-guards still pass.
- Validation commands:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - DEBATE_EXEC_INTEGRATION=1 go test -count=1 -tags exec_integration ./internal/backend/...
  - go vet ./...
  - git diff --exit-code -- go.mod go.sum
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/backend/exec/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend
  - bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk

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
  - command_001: bash scripts/check-gofmt.sh (exit 0, timed out: false, result: gate/validation/command_001/result.json)
  - command_002: go build ./... (exit 0, timed out: false, result: gate/validation/command_002/result.json)
  - command_003: go test -count=1 ./... (exit 0, timed out: false, result: gate/validation/command_003/result.json)
  - command_004: DEBATE_EXEC_INTEGRATION=1 go test -count=1 -tags exec_integration ./internal/backend/... (exit 0, timed out: false, result: gate/validation/command_004/result.json)
  - command_005: go vet ./... (exit 0, timed out: false, result: gate/validation/command_005/result.json)
  - command_006: git diff --exit-code -- go.mod go.sum (exit 0, timed out: false, result: gate/validation/command_006/result.json)
  - command_007: bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine (exit 0, timed out: false, result: gate/validation/command_007/result.json)
  - command_008: bash scripts/dep-guard.sh ./internal/backend/exec/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend (exit 0, timed out: false, result: gate/validation/command_008/result.json)
  - command_009: bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk (exit 0, timed out: false, result: gate/validation/command_009/result.json)
- Change summary:
  - changed files:
    - internal/backend/exec/exec.go
    - internal/backend/exec/exec_test.go
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
- Report every issue you believe is likely real: use state=candidate for uncertain findings and drop only when trigger, evidence, and fix_direction cannot be filled concretely.
