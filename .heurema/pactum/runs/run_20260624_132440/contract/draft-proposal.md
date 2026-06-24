# Contract Draft Proposal

## Status
- Run id: run_20260624_132440
- Status: accepted
- Source: drafter_attempt
- Drafter attempt: drafter_attempt_001
- Drafter: codex
- Accepted by: manual
- Accepted at: 2026-06-24T13:26:11Z

## In scope
- Add `debate init` handling in `cmd/debate` that creates `.heurema/debate`, `.heurema/debate/personas`, `context.md`, `config.yml`, and starter `proposer.md` and `skeptic.md` debater persona files.
- Add `debate new <name>` handling in `cmd/debate` that creates `.heurema/debate/personas/<name>.md` from a parseable persona template.
- Support `debate new <name> --role debater|synthesizer`, defaulting to `debater` when `--role` is omitted.
- Add deterministic Go tests under `cmd/debate` using temporary directories for init/new scaffold behavior.

## Out of scope
- Changing source under `internal/engine`, `internal/debate`, or `internal/backend`.
- Changing debate run behavior, synthesizer execution behavior, backend resolution, or transport implementations.
- Adding new third-party module dependencies or modifying `go.mod` or `go.sum`.

## Acceptance criteria
- `debate init` creates a workspace in the current working directory that `config.Load(root, nil, "")` loads successfully with exactly `proposer` and `skeptic` in the debater panel and the built-in default synthesizer.
- `debate init` creates non-empty, parseable `proposer.md` and `skeptic.md` files whose roles are `debater`, and creates a non-empty `context.md` template.
- Re-running `debate init` in an already scaffolded directory succeeds or reports the existing workspace without overwriting existing `config.yml`, `context.md`, `proposer.md`, or `skeptic.md` contents.
- `debate new <name>` creates `.heurema/debate/personas/<name>.md` with YAML frontmatter accepted by `persona.ParseFile`, including concrete `model` and `effort` values and a non-empty placeholder body.
- `debate new <name>` defaults the generated persona role to `debater`; `--role synthesizer` generates a parseable synthesizer persona; invalid role values are rejected with exit code 1 and no persona file.
- `debate new <name>` refuses to overwrite an existing persona file and leaves the original file contents unchanged.
- Existing `debate version` and normal debate run flag behavior continue to work.

## Validation commands
- go test ./cmd/debate
- go test ./internal/debate/config ./internal/debate/persona
- go test ./...
- ./scripts/check-gofmt.sh
- go vet ./...
- git diff --exit-code -- go.mod go.sum

## Assumptions
- The starter personas may use existing inferable Claude model defaults, such as `claude-sonnet-4-6` or another model already accepted by `persona.InferBackend`, with explicit `effort` values.
- `debate init` should use `config.yml` table entries `proposer` and `skeptic` so the scaffolded panel order is deterministic.
- Persona names for `debate new <name>` are intended to map directly to Markdown basenames under `.heurema/debate/personas`; unsafe path components such as slashes or `..` should be rejected.

