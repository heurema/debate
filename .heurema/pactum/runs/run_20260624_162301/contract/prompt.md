# Executor Prompt

This prompt is prepared from an approved Pactum contract.
This prompt is prepared for the selected built-in agent when `pactum execute run` is used.
Pactum records execution artifacts and validates contract, map, and memory boundaries before execution.

## Contract status
- Run: run_20260624_162301
- Approval: approved
- Contract hash: 701c64fe58cc4bf817bc110cd3adfac47b3c664801c4d06b6fa2ff224335a605

## Goal
Replace the hand-written stdlib flag parsing in cmd/debate with github.com/alecthomas/kong (the CLI library pactum uses), so flags parse correctly in any position, while preserving every existing command, flag, and exit-code behavior.

## In scope
- Add github.com/alecthomas/kong and parse cmd/debate arguments with it, removing the hand-rolled flag.FlagSet / os.Args switch parsing in cmd/debate/main.go and cmd/debate/scaffold.go.
- Model the debate run as kong's default command (bare `debate "<task>"`) alongside the version, init, and new subcommands, with all flags as kong struct tags.
- Preserve the exact existing behavior (commands, flags, task composition, exit codes, stdout/stderr) and make flags parse in any position relative to positionals.
- Update the cmd/debate tests for kong, including assertions that flags work both before and after the positional argument.

## Out of scope
- Any change to internal/engine, internal/debate, or internal/backend.
- Adding new subcommands (no validate) or new flags, or changing the backend resolver, runner, synthesizer, or IO contract.
- Changing the debate algorithm, personas, or config.

## Acceptance criteria
- cmd/debate parses all arguments with github.com/alecthomas/kong (added to go.mod and go.sum); the hand-rolled flag.NewFlagSet parsing and os.Args dispatch in main.go and scaffold.go are removed.
- The CLI preserves the existing commands: a default run action invoked as `debate "<task>"` with no subcommand word, plus subcommands `version`, `init`, and `new <name>`; the run is kong's default command so the bare-task form keeps working.
- Flags parse in any position relative to the positional argument: `debate "<task>" --json`, `debate --json "<task>"`, `debate --max-rounds 2 "<task>"`, `debate "<task>" --max-rounds 2`, and `debate new <name> --role synthesizer` all apply the flag correctly. (This fixes the previous flags-after-positional bug.)
- All existing run flags are preserved with the same names and meaning: --with (panel selectors), --synth, --max-rounds, --json, -q/--quiet, --sealed, and --task (@file); the new subcommand keeps its --role flag (debater|synthesizer, default debater).
- Task sources still compose: the positional task, --task @file (file contents), and stdin (appended when piped); an empty resulting task is a fail-fast error.
- Exit codes are unchanged: 0 when settled, 2 when not converged (stalemate or max), 1 on error; an unknown flag or subcommand prints a clear kong usage/help message with a non-zero exit; the stdout=final-answer / stderr=live-trace (auto-quiet off-TTY or with -q) contract and the --json output shape are unchanged.
- cmd/debate tests are updated for kong and include assertions that a representative flag is honored both before and after the positional argument for the run command and for `new`.
- go.mod and go.sum gain github.com/alecthomas/kong; internal/engine, internal/debate, and internal/backend are unchanged; check-gofmt, go build ./..., go vet ./..., and go test ./... pass, and the engine/backend/debate dep-guards still pass.

## Validation commands
- bash scripts/check-gofmt.sh
- go build ./...
- go test -count=1 ./...
- go vet ./...
- go run ./cmd/debate version
- bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
- bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk
- bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

## Assumptions
- kong (github.com/alecthomas/kong, the library pactum uses) parses flags in any position by design and provides usage/help and exit handling; the debate run is modeled as a default command (e.g. kong's default:"withargs") so the bare `debate "<task>"` form coexists with the version/init/new subcommands.
- Only the argument-parsing layer changes; the backend resolver, runner, synthesizer, and IO contract are untouched, so behavior is identical except for the flag-position fix.
- No subcommands are added (validate was never implemented and remains out of scope); kong is a cmd/debate-only dependency and must not leak into internal/engine, internal/debate, or internal/backend.

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
- total: 0
- fresh: 0
- stale: 0
- unknown: 0

Items:
- none

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
