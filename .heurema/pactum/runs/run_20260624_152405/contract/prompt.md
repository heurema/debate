# Executor Prompt

This prompt is prepared from an approved Pactum contract.
This prompt is prepared for the selected built-in agent when `pactum execute run` is used.
Pactum records execution artifacts and validates contract, map, and memory boundaries before execution.

## Contract status
- Run: run_20260624_152405
- Approval: approved
- Contract hash: 40967fcfe0552c0737ab4ecdba4488ffd1cf3b90c0a1abc9252460c90fb1c80f

## Goal
Remove the context.md baseline-preamble feature so debate context lives only in the task (plus the grounded sandbox): drop config.Workspace.Context and context.md loading, assemble the brief from the task alone, and stop scaffolding context.md in debate init.

## In scope
- Drop the Context field from config.Workspace and stop reading .heurema/debate/context.md in internal/debate/config.
- Make internal/debate/runner assemble the debate brief from the task alone and remove every use of Workspace.Context in the runner and synthesizer.
- Make debate init in cmd/debate scaffold only the two starter debater personas and no longer create a context.md file.
- Update all affected tests so none assert context.md creation or a Workspace.Context field.

## Out of scope
- Any change to grounding, the backends, personas, or the engine.
- Docs updates (DESIGN.md / SLICES.md) — handled separately.
- Introducing any new dependency or new feature.

## Acceptance criteria
- config.Workspace no longer has a Context field, and config.Load no longer opens or reads .heurema/debate/context.md; a context.md present in a workspace is ignored and does not affect loading or cause an error.
- internal/debate/runner assembles the debate brief from the task text alone (no context preamble), and neither the runner nor the synthesizer references Workspace.Context.
- `debate init` scaffolds only personas/proposer.md and personas/skeptic.md under .heurema/debate and does not create a context.md file; it prints only the persona paths it created.
- The scaffolded workspace still loads via config.Load with the two-debater panel and the built-in default synthesizer.
- All affected tests are updated: no test asserts context.md creation or a Workspace.Context field; the init test asserts the two persona files are created and the workspace loads (two-debater panel) with no context.md; the runner test asserts the assembled brief equals the task.
- check-gofmt, go build ./..., go vet ./..., and go test ./... pass; go.mod and go.sum are unchanged (no new dependency); the engine and debate dep-guards still pass.

## Validation commands
- bash scripts/check-gofmt.sh
- go build ./...
- go test -count=1 ./...
- go vet ./...
- go run ./cmd/debate version
- git diff --exit-code -- go.mod go.sum
- bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
- bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

## Assumptions
- Debate context is supplied in the task (optionally via --task @file) and deeper project context is read by the grounded agents themselves; the static context.md preamble is therefore redundant and removed.
- Removing Workspace.Context is an internal-only change; only internal callers in runner and cmd reference it.
- A pre-existing context.md in a user workspace becomes a no-op (ignored), not an error.

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
