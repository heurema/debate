# Reviewer Context

## Run
- Run id: run_20260624_152405
- Run status: contract_approved

## Contract
- Goal: Remove the context.md baseline-preamble feature so debate context lives only in the task (plus the grounded sandbox): drop config.Workspace.Context and context.md loading, assemble the brief from the task alone, and stop scaffolding context.md in debate init.
- In scope:
  - Drop the Context field from config.Workspace and stop reading .heurema/debate/context.md in internal/debate/config.
  - Make internal/debate/runner assemble the debate brief from the task alone and remove every use of Workspace.Context in the runner and synthesizer.
  - Make debate init in cmd/debate scaffold only the two starter debater personas and no longer create a context.md file.
  - Update all affected tests so none assert context.md creation or a Workspace.Context field.
- Out of scope:
  - Any change to grounding, the backends, personas, or the engine.
  - Docs updates (DESIGN.md / SLICES.md) — handled separately.
  - Introducing any new dependency or new feature.
- Acceptance criteria:
  - config.Workspace no longer has a Context field, and config.Load no longer opens or reads .heurema/debate/context.md; a context.md present in a workspace is ignored and does not affect loading or cause an error.
  - internal/debate/runner assembles the debate brief from the task text alone (no context preamble), and neither the runner nor the synthesizer references Workspace.Context.
  - `debate init` scaffolds only personas/proposer.md and personas/skeptic.md under .heurema/debate and does not create a context.md file; it prints only the persona paths it created.
  - The scaffolded workspace still loads via config.Load with the two-debater panel and the built-in default synthesizer.
  - All affected tests are updated: no test asserts context.md creation or a Workspace.Context field; the init test asserts the two persona files are created and the workspace loads (two-debater panel) with no context.md; the runner test asserts the assembled brief equals the task.
  - check-gofmt, go build ./..., go vet ./..., and go test ./... pass; go.mod and go.sum are unchanged (no new dependency); the engine and debate dep-guards still pass.
- Validation commands:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - go vet ./...
  - go run ./cmd/debate version
  - git diff --exit-code -- go.mod go.sum
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

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
  - command_004: go vet ./... (exit 0, timed out: false, result: gate/validation/command_004/result.json)
  - command_005: go run ./cmd/debate version (exit 0, timed out: false, result: gate/validation/command_005/result.json)
  - command_006: git diff --exit-code -- go.mod go.sum (exit 0, timed out: false, result: gate/validation/command_006/result.json)
  - command_007: bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine (exit 0, timed out: false, result: gate/validation/command_007/result.json)
  - command_008: bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3 (exit 0, timed out: false, result: gate/validation/command_008/result.json)
- Change summary:
  - changed files:
    - cmd/debate/e2e_test.go
    - cmd/debate/scaffold.go
    - cmd/debate/scaffold_test.go
    - internal/debate/config/config.go
    - internal/debate/config/config_test.go
    - internal/debate/runner/runner.go
    - internal/debate/runner/runner_test.go
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
