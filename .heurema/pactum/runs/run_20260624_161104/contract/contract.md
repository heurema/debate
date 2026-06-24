# Contract Draft

## Goal
Invoke agy non-interactively via --print in the exec backend so it works against the real agy CLI (which otherwise defaults to an interactive session and hangs).

## Current status
Contract status: approved
Manual clarification, contract approval, prompt build, and agent execution are available through staged Pactum commands.

## Relevant repository context
- Map run: map_20260624_152728
- Repo map: .heurema/pactum/map/repo-map.md
- Search results: context/search-results.json (0 result(s))

## Clarifications
- None

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

## Open questions
- None
