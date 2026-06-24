# Memory Candidate

## Run
- Run id: run_20260624_162301
- Source: deterministic

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

## Outcome
- Gate status: needs_review
- Review status: approved
- Execution exit code: 0
- Validation passed: true
- Changes need review: true

## Changes
- Changed files:
  - cmd/debate/e2e_test.go
  - cmd/debate/main.go
  - cmd/debate/scaffold.go
  - cmd/debate/scaffold_test.go
  - go.mod
  - go.sum
- New files: none
- Missing files: none

## Clarifications
- None

## Review Decisions
- f_001 [medium] resolved cmd/debate/main.go:53: The kong default run is still exposed as an explicit `run` subcommand, adding a command outside the approved CLI surface and changing task composition for tasks whose first unquoted word is `run`.
  Resolution: Removed the top-level Kong `run` child command in cmd/debate/main.go. Bare run arguments are now parsed through a root runCmd parser, so `run` is task text rather than an explicit subcommand; added TestE2E_RunWordIsTaskNotCommand.
- f_002 [medium] resolved cmd/debate/e2e_test.go:126: The new max-rounds before/after test can pass even if --max-rounds is ignored.
  Resolution: Strengthened TestE2E_MaxRoundsFlagBeforeAndAfterTask in cmd/debate/e2e_test.go to use parseCLI and assert the forced trace contains exactly one full round, so the test fails if --max-rounds is ignored.
- f_003 [medium] resolved cmd/debate/main.go:101: The tests do not exercise the top-level Kong CLI path that main actually uses.
  Resolution: Updated representative run and new-command before/after flag tests to exercise parseCLI, the same top-level path used by main.
- f_004 [low] open cmd/debate/main.go:145: Kong parse-error paths have no targeted regression tests.
- f_005 [low] open cmd/debate/main.go:190: The change introduces duplicate Kong parsing entrypoints outside the production CLI path.
- f_006 [low] open docs/DESIGN.md:305: The CLI docs still say `debate new` flags must come before the name, so the newly supported after-name form is undocumented.
- Proposal summary: pending=0 accepted=6 rejected=2

## Reusable Project Knowledge
- scope: in scope: Add github.com/alecthomas/kong and parse cmd/debate arguments with it, removing the hand-rolled flag.FlagSet / os.Args switch parsing in cmd/debate/main.go and cmd/debate/scaffold.go.
- scope: in scope: Model the debate run as kong's default command (bare `debate "<task>"`) alongside the version, init, and new subcommands, with all flags as kong struct tags.
- scope: in scope: Preserve the exact existing behavior (commands, flags, task composition, exit codes, stdout/stderr) and make flags parse in any position relative to positionals.
- scope: in scope: Update the cmd/debate tests for kong, including assertions that flags work both before and after the positional argument.
- scope: out of scope: Any change to internal/engine, internal/debate, or internal/backend.
- scope: out of scope: Adding new subcommands (no validate) or new flags, or changing the backend resolver, runner, synthesizer, or IO contract.
- scope: out of scope: Changing the debate algorithm, personas, or config.
- review_resolution: f_001 resolved: The kong default run is still exposed as an explicit `run` subcommand, adding a command outside the approved CLI surface and changing task composition for tasks whose first unquoted word is `run`.; resolution: Removed the top-level Kong `run` child command in cmd/debate/main.go. Bare run arguments are now parsed through a root runCmd parser, so `run` is task text rather than an explicit subcommand; added TestE2E_RunWordIsTaskNotCommand.
- review_resolution: f_002 resolved: The new max-rounds before/after test can pass even if --max-rounds is ignored.; resolution: Strengthened TestE2E_MaxRoundsFlagBeforeAndAfterTask in cmd/debate/e2e_test.go to use parseCLI and assert the forced trace contains exactly one full round, so the test fails if --max-rounds is ignored.
- review_resolution: f_003 resolved: The tests do not exercise the top-level Kong CLI path that main actually uses.; resolution: Updated representative run and new-command before/after flag tests to exercise parseCLI, the same top-level path used by main.
- review_resolution: proposal p_001 accepted as f_001
- review_resolution: proposal p_002 accepted as f_002
- review_resolution: proposal p_003 accepted as f_003
- review_resolution: proposal p_004 accepted as f_004
- review_resolution: proposal p_005 accepted as f_005
- review_resolution: proposal p_006 accepted as f_006
- review_resolution: proposal p_007 rejected: stale pre-existing docs for -t alias; contract preserves --task only
- review_resolution: proposal p_008 rejected: validate subcommand explicitly out of scope for this contract
- validation: bash scripts/check-gofmt.sh passed
- validation: go build ./... passed
- validation: go test -count=1 ./... passed
- validation: go vet ./... passed
- validation: go run ./cmd/debate version passed
- validation: bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine passed
- validation: bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk passed
- validation: bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3 passed

## Artifacts
- Contract: contract/contract.json
- Gate report: gate/gate-report.json
- Review: review/review.json
- Findings: review/findings.jsonl
- Resolutions: review/resolutions.jsonl
- Proposals: review/proposals.jsonl
- Proposal decisions: review/proposal-decisions.jsonl
