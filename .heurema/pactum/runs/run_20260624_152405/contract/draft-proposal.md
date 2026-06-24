# Contract Draft Proposal

## Status
- Run id: run_20260624_152405
- Status: accepted
- Source: drafter_attempt
- Drafter attempt: drafter_attempt_001
- Drafter: codex
- Accepted by: manual
- Accepted at: 2026-06-24T15:26:36Z

## In scope
- Remove all context.md loading and storage from internal/debate/config, including the Workspace.Context field and context.md-specific read/error handling.
- Make internal/debate/runner build debater prompts from cfg.Task alone, with no workspace context preamble and no Workspace.Context dependency in runner or synthesis paths.
- Change debate init scaffolding in cmd/debate to create only proposer.md and skeptic.md starter debater personas, without creating .heurema/debate/context.md.
- Update config, runner, scaffold, and command/e2e tests and fixtures to reflect that context.md is ignored and not scaffolded.

## Out of scope
- Documentation updates.
- Grounding, read-only sandbox behavior, transport backends, persona parsing semantics, engine orchestration, verdict logic, and signal handling changes.
- Adding dependencies or changing go.mod/go.sum.
- Deleting, migrating, or modifying existing user-created .heurema/debate/context.md files.
- Changing task assembly from CLI flags, positional args, files, or stdin except for removing the workspace context preamble from debate prompts.

## Acceptance criteria
- config.Workspace no longer has a Context field, and config.Load succeeds without reading or exposing .heurema/debate/context.md when such a file exists.
- No production code path in internal/debate/config, internal/debate/runner, or cmd/debate depends on Workspace.Context, contextFileName, contextTemplate, or an assembleBrief function that accepts workspace context.
- A runner test captures the debater prompt and verifies the Brief content is exactly the supplied task, with no context.md content prepended or included.
- The synthesizer prompt is built from the task and transcript only; context.md content is not included in synthesis input.
- debate init creates a loadable workspace whose panel contains exactly proposer and skeptic in lexicographic order and whose .heurema/debate/context.md file does not exist.
- Existing debate behaviors unrelated to context.md remain covered and passing, including panel selection, synthesizer default/override behavior, trace/json output, max-round handling, backend resolution, and sealed/read-only transport specs.

## Validation commands
- ./scripts/check-gofmt.sh
- go test ./internal/debate/config ./internal/debate/runner ./cmd/debate
- make check
- bash -c '! rg -n "Context[[:space:]]+string|Workspace\\.Context|ws\\.Context|contextFileName|contextTemplate|Read context\\.md|assembleBrief\\(" internal/debate cmd/debate'
- git diff --exit-code -- go.mod go.sum

## Assumptions
- Breaking internal Go references to config.Workspace.Context is acceptable because the package is under internal/ and affected repository tests will be updated.
- Existing context.md files should be left on disk and ignored rather than removed.
- Mentions of context.md in documentation may remain because docs updates are explicitly handled separately.

