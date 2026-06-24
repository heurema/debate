# Reviewer Context

## Run
- Run id: run_20260624_162301
- Run status: contract_approved

## Contract
- Goal: Replace the hand-written stdlib flag parsing in cmd/debate with github.com/alecthomas/kong (the CLI library pactum uses), so flags parse correctly in any position, while preserving every existing command, flag, and exit-code behavior.
- In scope:
  - Add github.com/alecthomas/kong and parse cmd/debate arguments with it, removing the hand-rolled flag.FlagSet / os.Args switch parsing in cmd/debate/main.go and cmd/debate/scaffold.go.
  - Model the debate run as kong's default command (bare `debate "<task>"`) alongside the version, init, and new subcommands, with all flags as kong struct tags.
  - Preserve the exact existing behavior (commands, flags, task composition, exit codes, stdout/stderr) and make flags parse in any position relative to positionals.
  - Update the cmd/debate tests for kong, including assertions that flags work both before and after the positional argument.
- Out of scope:
  - Any change to internal/engine, internal/debate, or internal/backend.
  - Adding new subcommands (no validate) or new flags, or changing the backend resolver, runner, synthesizer, or IO contract.
  - Changing the debate algorithm, personas, or config.
- Acceptance criteria:
  - cmd/debate parses all arguments with github.com/alecthomas/kong (added to go.mod and go.sum); the hand-rolled flag.NewFlagSet parsing and os.Args dispatch in main.go and scaffold.go are removed.
  - The CLI preserves the existing commands: a default run action invoked as `debate "<task>"` with no subcommand word, plus subcommands `version`, `init`, and `new <name>`; the run is kong's default command so the bare-task form keeps working.
  - Flags parse in any position relative to the positional argument: `debate "<task>" --json`, `debate --json "<task>"`, `debate --max-rounds 2 "<task>"`, `debate "<task>" --max-rounds 2`, and `debate new <name> --role synthesizer` all apply the flag correctly. (This fixes the previous flags-after-positional bug.)
  - All existing run flags are preserved with the same names and meaning: --with (panel selectors), --synth, --max-rounds, --json, -q/--quiet, --sealed, and --task (@file); the new subcommand keeps its --role flag (debater|synthesizer, default debater).
  - Task sources still compose: the positional task, --task @file (file contents), and stdin (appended when piped); an empty resulting task is a fail-fast error.
  - Exit codes are unchanged: 0 when settled, 2 when not converged (stalemate or max), 1 on error; an unknown flag or subcommand prints a clear kong usage/help message with a non-zero exit; the stdout=final-answer / stderr=live-trace (auto-quiet off-TTY or with -q) contract and the --json output shape are unchanged.
  - cmd/debate tests are updated for kong and include assertions that a representative flag is honored both before and after the positional argument for the run command and for `new`.
  - go.mod and go.sum gain github.com/alecthomas/kong; internal/engine, internal/debate, and internal/backend are unchanged; check-gofmt, go build ./..., go vet ./..., and go test ./... pass, and the engine/backend/debate dep-guards still pass.
- Validation commands:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - go vet ./...
  - go run ./cmd/debate version
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk
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
- Execution attempt id: attempt_002
- Execution exit code: 0
- Validation command results:
  - command_001: bash scripts/check-gofmt.sh (exit 0, timed out: false, result: gate/validation/command_001/result.json)
  - command_002: go build ./... (exit 0, timed out: false, result: gate/validation/command_002/result.json)
  - command_003: go test -count=1 ./... (exit 0, timed out: false, result: gate/validation/command_003/result.json)
  - command_004: go vet ./... (exit 0, timed out: false, result: gate/validation/command_004/result.json)
  - command_005: go run ./cmd/debate version (exit 0, timed out: false, result: gate/validation/command_005/result.json)
  - command_006: bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine (exit 0, timed out: false, result: gate/validation/command_006/result.json)
  - command_007: bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk (exit 0, timed out: false, result: gate/validation/command_007/result.json)
  - command_008: bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3 (exit 0, timed out: false, result: gate/validation/command_008/result.json)
- Change summary:
  - changed files:
    - cmd/debate/e2e_test.go
    - cmd/debate/main.go
    - cmd/debate/scaffold.go
    - cmd/debate/scaffold_test.go
    - go.mod
    - go.sum
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
