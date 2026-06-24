# Reviewer Context

## Run
- Run id: run_20260624_132440
- Run status: contract_approved

## Contract
- Goal: Add debate init and debate new scaffolding subcommands to cmd/debate that create a ready-to-run .heurema/debate workspace and new persona files, adding no new module dependency, without changing internal/engine, internal/debate, or internal/backend.
- In scope:
  - Implement the `debate init` subcommand: scaffold a .heurema/debate workspace under the current directory with two starter debater personas and a context.md template, safely (never overwriting existing files).
  - Implement the `debate new <name>` subcommand: create a new persona file from a template under a discovered .heurema/debate/personas, with a role flag, safely (never overwriting an existing persona).
  - Make the scaffolded workspace immediately loadable and runnable (valid personas that load via the existing config/persona packages).
  - Add deterministic unit tests using temporary directories that assert init and new behavior, including the loadability of the scaffolded workspace and the refuse-to-overwrite behavior.
- Out of scope:
  - Any change to internal/engine, internal/debate, or internal/backend source (the subcommands live in cmd/debate and reuse config/persona by import only).
  - Backends, the debate run path, the synthesizer, or convergence behavior.
  - Editing or migrating an existing workspace's content beyond adding new files; and any new third-party module dependency.
- Acceptance criteria:
  - `debate init` creates a .heurema/debate directory under the current working directory containing personas/proposer.md and personas/skeptic.md (each a valid debater persona with role debater, a concrete model and effort, and a system-prompt body) and a context.md template file; it prints the paths it created.
  - The scaffolded workspace loads successfully via the existing config.Load: discovery finds the new .heurema/debate, the panel resolves to the two starter debaters, and the built-in default synthesizer is used (init does not scaffold a synthesizer file or a config.yml, since the default panel is all debater personas).
  - The two starter personas use concrete valid values (a real model id such as claude-haiku-4-5 and a valid effort) so the workspace is immediately runnable without edits; persona.ParseFile accepts both.
  - `debate init` is safe: if a target file already exists it does not overwrite it (it skips that file with a clear message or refuses with a clear error and a documented exit code); an existing .heurema/debate is never clobbered.
  - `debate new <name>` creates <name>.md under the discovered .heurema/debate/personas from a template with YAML frontmatter (role defaulting to debater, overridable via --role debater|synthesizer; a concrete model and effort default; optional backend) and a placeholder body; it prints the created path, and the created file is accepted by persona.ParseFile.
  - `debate new` validates the name (a simple persona id, rejecting path separators) and refuses to overwrite an existing persona file with a clear error; it requires a discoverable .heurema/debate workspace (walking up parent directories like config discovery) and errors clearly when none is found, creating the personas directory within the discovered workspace if it does not yet exist.
  - init and new write only under .heurema/debate, never outside it, use the Go standard library plus the existing internal packages, and add no new module dependency (go.mod and go.sum are unchanged); an unknown flag or a missing required argument prints a clear usage error with a non-zero exit.
  - Deterministic unit tests using temporary directories assert: init creates a workspace that config.Load accepts with the two-debater panel, re-running init does not overwrite existing files, new creates a persona that persona.ParseFile accepts, and new refuses to overwrite an existing persona; check-gofmt, go vet ./..., go build ./..., and go test ./... pass.
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
    - cmd/debate/main.go
  - new files:
    - cmd/debate/scaffold.go
    - cmd/debate/scaffold_test.go
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
