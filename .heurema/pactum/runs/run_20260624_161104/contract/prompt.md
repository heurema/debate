# Executor Prompt

This prompt is prepared from an approved Pactum contract.
This prompt is prepared for the selected built-in agent when `pactum execute run` is used.
Pactum records execution artifacts and validates contract, map, and memory boundaries before execution.

## Contract status
- Run: run_20260624_161104
- Approval: approved
- Contract hash: 543c636e640cbf91c7ea9083b6ca50c9fdfecddf12015cbbf262935b076366a8

## Goal
Invoke agy non-interactively via --print in the exec backend so it works against the real agy CLI (which otherwise defaults to an interactive session and hangs).

## In scope
- Change the default agy argv in internal/backend/exec to include --print so agy runs a single prompt non-interactively from stdin and exits.
- Update the affected exec unit tests (argv assertions) and the gated integration test to match the new argv.

## Out of scope
- Other backends, internal/engine, the internal/debate packages, and the acp backend.
- The exec backend's stdin reconstruction / delta accumulation logic, error handling, grounding, and recovery.
- CLI flag-ordering / argument parsing (handled separately).

## Acceptance criteria
- The exec backend's default agy argv is [agy, "--print", "--model", spec.Model]; the --print flag (alias of -p) makes agy run a single prompt non-interactively, reading the prompt from stdin and printing the response before exiting. The prompt is still written to the subprocess stdin (the reconstruction logic is unchanged).
- The DEBATE_AGY_COMMAND override still replaces only the executable token (argv[0]) and preserves the --print and --model arguments in order.
- spec.Model must still be non-empty (fail-fast otherwise); the model is passed as the --model value.
- Unit tests assert the new default argv (including --print) for both the default and the DEBATE_AGY_COMMAND-overridden command.
- The gated integration test (enabled by setting DEBATE_EXEC_INTEGRATION=1) exercises the real agy --print invocation end-to-end — agy reads the prompt from stdin, prints the response, and exits without hanging — thereby serving as the runnable validation of the --print non-interactive behavior.
- check-gofmt, go build ./..., go vet ./..., and go test ./... pass; go.mod and go.sum are unchanged (no new dependency); internal/engine is unchanged and the engine/exec/backend/debate dep-guards still pass.

## Validation commands
- bash scripts/check-gofmt.sh
- go build ./...
- go test -count=1 ./...
- DEBATE_EXEC_INTEGRATION=1 go test -count=1 -tags exec_integration ./internal/backend/...
- go vet ./...
- git diff --exit-code -- go.mod go.sum
- bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
- bash scripts/dep-guard.sh ./internal/backend/exec/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend
- bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk

## Assumptions
- agy --print (alias -p) runs a single prompt non-interactively, reading the prompt from stdin and printing the model response, then exits; without it agy starts an interactive session and the subprocess hangs.
- This is a focused argv change; the stdin reconstruction, grounding, error classification, retry, and Close behavior are all unchanged.
- The verified real agy version is 1.0.11; --print/-p is its documented non-interactive single-prompt flag.

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
