# Memory Candidate

## Run
- Run id: run_20260624_161104
- Source: deterministic

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

## Outcome
- Gate status: needs_review
- Review status: approved
- Execution exit code: 0
- Validation passed: true
- Changes need review: true

## Changes
- Changed files:
  - internal/backend/exec/exec.go
  - internal/backend/exec/exec_test.go
- New files: none
- Missing files: none

## Clarifications
- None

## Review Decisions
- f_001 [medium] open internal/backend/exec/integration_test.go:22: The gated integration test TestIntegration_Agy skips unless AGY_MODEL is set, but the contract's validation command_004 (DEBATE_EXEC_INTEGRATION=1 go test -count=1 -tags exec_integration ./internal/backend/...) does not set AGY_MODEL. The test therefore skips at integration_test.go:21-24 and invokes no agy subprocess; the gate reports 'ok' only because a skipped test passes. Acceptance criterion 5 claims this command 'exercises the real agy --print invocation end-to-end', but no subprocess ran, so the non-interactive --print behavior (the whole point of the change) is never validated by the gate. The unit tests use a fake runner and assert only argv strings, so they cannot prove --print actually prevents the interactive hang.
- f_002 [medium] open internal/backend/exec/integration_test.go:21: Acceptance criterion 5 claims the gated integration test 'exercises the real agy --print invocation end-to-end ... serving as the runnable validation of the --print non-interactive behavior', but the gate never actually ran it. TestIntegration_Agy requires AGY_MODEL to be set (integration_test.go:21-24), while the contract's validation command (command_004) sets only DEBATE_EXEC_INTEGRATION=1. The test therefore skipped: command_004 stdout shows 'ok github.com/heurema/debate/internal/backend/exec 0.520s' with duration_ms=699 — far too fast for a real LLM CLI call. The --print no-hang behavior, which is the entire purpose of this change, has no automated validation; correctness rests solely on the contract's stated assumption about agy 1.0.11.
- f_003 [medium] open internal/backend/exec/integration_test.go:23: Acceptance criterion #5 claims the gated integration test (TestIntegration_Agy) is the runnable validation of the agy --print non-interactive (no-hang) behavior, but the contract's validation command (command_004) only sets DEBATE_EXEC_INTEGRATION=1 and not AGY_MODEL. TestIntegration_Agy skips at integration_test.go:23 when AGY_MODEL is empty, so the real --print invocation is never exercised by the gate. The gate's command_004 finished in 0.52s with 'ok' and no real-agy response — i.e. it skipped. The only automated coverage of the change is the unit assertion that '--print' appears in argv (TestCmd_Default/TestCmd_Override); the actual 'does not hang against real agy' behavior is unverified by any command the gate runs.
- Proposal summary: pending=0 accepted=3 rejected=0

## Reusable Project Knowledge
- scope: in scope: Change the default agy argv in internal/backend/exec to include --print so agy runs a single prompt non-interactively from stdin and exits.
- scope: in scope: Update the affected exec unit tests (argv assertions) and the gated integration test to match the new argv.
- scope: out of scope: Other backends, internal/engine, the internal/debate packages, and the acp backend.
- scope: out of scope: The exec backend's stdin reconstruction / delta accumulation logic, error handling, grounding, and recovery.
- scope: out of scope: CLI flag-ordering / argument parsing (handled separately).
- review_resolution: proposal p_001 accepted as f_001
- review_resolution: proposal p_002 accepted as f_002
- review_resolution: proposal p_003 accepted as f_003
- validation: bash scripts/check-gofmt.sh passed
- validation: go build ./... passed
- validation: go test -count=1 ./... passed
- validation: DEBATE_EXEC_INTEGRATION=1 go test -count=1 -tags exec_integration ./internal/backend/... passed
- validation: go vet ./... passed
- validation: git diff --exit-code -- go.mod go.sum passed
- validation: bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine passed
- validation: bash scripts/dep-guard.sh ./internal/backend/exec/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend passed
- validation: bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk passed

## Artifacts
- Contract: contract/contract.json
- Gate report: gate/gate-report.json
- Review: review/review.json
- Findings: review/findings.jsonl
- Resolutions: review/resolutions.jsonl
- Proposals: review/proposals.jsonl
- Proposal decisions: review/proposal-decisions.jsonl
