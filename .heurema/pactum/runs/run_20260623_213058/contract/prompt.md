# Executor Prompt

This prompt is prepared from an approved Pactum contract.
This prompt is prepared for the selected built-in agent when `pactum execute run` is used.
Pactum records execution artifacts and validates contract, map, and memory boundaries before execution.

## Contract status
- Run: run_20260623_213058
- Approval: approved
- Contract hash: 7c599a744cc52cb87d1c78d7b36df34e87f780876ce9b99e71718b476fd58196

## Goal
Slice 0: bootstrap the debate Go project skeleton — go.mod (module github.com/heurema/debate), package layout (internal/engine/{loop,transport,orchestrate}, internal/debate, cmd/debate), Makefile, and a 'debate version' command. No engine or debate logic yet.

## In scope
- Create a root Go module at go.mod with module path github.com/heurema/debate.
- Create the Slice 0 package layout: cmd/debate, internal/debate, internal/engine/loop, internal/engine/transport, and internal/engine/orchestrate.
- Add minimal Go source files needed for all Slice 0 packages to compile and be listed by go list ./....
- Implement a minimal debate CLI entrypoint with a version subcommand.
- Add a root Makefile with build, vet, test, and check targets.
- Add at least one trivial Go test so the repository has a runnable test suite.

## Out of scope
- Do not implement engine loop behavior, orchestration, transport behavior, mock backend behavior, debate policy, prompt building, verdict logic, persona parsing, synthesizer behavior, or model/backend integrations.
- Do not implement .heurema/debate discovery, config.yml parsing, context.md loading, init/new commands, or full debate task execution.
- Do not rename the physical repository directory or migrate per-project memory/state.
- Do not add network calls, external model calls, credentials, or nonessential third-party dependencies.

## Acceptance criteria
- go.mod exists at the repository root and declares module github.com/heurema/debate.
- go list ./... succeeds and includes github.com/heurema/debate/cmd/debate, github.com/heurema/debate/internal/debate, github.com/heurema/debate/internal/engine/loop, github.com/heurema/debate/internal/engine/transport, and github.com/heurema/debate/internal/engine/orchestrate.
- The root Makefile defines build, vet, test, and check targets; make check runs the Slice 0 validation path successfully.
- go run ./cmd/debate version exits 0 and prints a non-empty version string identifying the debate binary.
- go test ./... succeeds with at least one trivial test present.
- The implementation remains a skeleton: no engine/debate runtime logic, backend integrations, persona/config discovery, or synthesizer behavior is added.

## Validation commands
- go list ./...
- go test ./...
- go vet ./...
- go run ./cmd/debate version
- make build
- make check

## Assumptions
- A default development version string is acceptable for Slice 0 when no release metadata is supplied.
- Placeholder package files are acceptable where needed to make empty Slice 0 packages compile and appear in go list ./....
- Slice 0 keeps the repository directory name unchanged even though the Go module path is github.com/heurema/debate.
- The Go standard library is sufficient for this slice unless the existing project later introduces a required dependency.

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
