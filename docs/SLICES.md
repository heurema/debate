# debate implementation slices

> Status: v1 implementation map. Slices are ordered by dependency and designed to stay vertically testable.
>
> Full design: [`docs/DESIGN.md`](DESIGN.md)

## 0. Project Skeleton

Create the Go module `github.com/heurema/debate`, package layout, build/test wiring, and a minimal `debate version` command.

Expected result:

- repository builds
- tests run
- package boundaries exist
- CLI binary has a version surface

## 1. Engine On A Mock Backend

Build the reusable engine without any real model calls.

Scope:

- `internal/engine/loop`: streak loop with max, settle, and patience limits
- `internal/engine/transport`: transport/session/spec/result interfaces
- `internal/engine/transport/mock`: scripted test backend
- `internal/engine/orchestrate`: participants, turns, transcript, deltas, scheduler, prompt/verdict seams

Expected result:

- N mock participants run round-robin
- transcript is accumulated
- `DeltaFor` exposes only unseen turns
- a trivial verdict can stop the loop
- loop behavior is tested independently

## 2. Debate Policy

Add the product policy that turns the generic engine into a debate.

Scope:

- prompt builder that renders moderator rules, task brief, discussion board, and signal instructions
- `signal` parser for fenced JSON convergence blocks
- verdict implementation for `all_done` and `quorum`
- progress detection through open-objection set changes

Expected result:

- scripted participants can converge
- malformed or missing signals count as not done
- `done=true` with objections is normalized to not done
- max-round and stalemate behavior is covered

## 3. Personas And Workspace Loading

Load `.heurema/debate` from the project tree and resolve the active panel.

Scope:

- walk-up discovery for `.heurema/debate`
- persona parser for Markdown plus YAML front matter
- strict front matter validation
- optional `config.yml` with `table`
- default panel from all `role: debater` personas
- `--with` override
- synthesizer resolution
- backend inference from model names

Expected result:

- valid fixture workspaces load
- broken personas fail fast
- unknown config keys fail fast
- missing workspace returns a clear error
- synthesizer-role personas cannot be placed in the debater panel

## 4. CLI And Synthesizer On Test Backends

Connect the CLI to the product runner and engine while still allowing deterministic local execution.

Scope:

- task assembly from positional args, `--task`, files, stdin, and pipes
- runner config construction
- live trace to stderr when appropriate
- stdout final answer
- JSON output mode
- exit-code contract
- built-in default synthesizer
- `--synth`, `--max-rounds`, `--json`, `-q`, `--sealed`

Expected result:

- `debate "task"` works against deterministic backends
- `--json` is automation-friendly
- errors happen before model calls when config is invalid
- stdout/stderr behavior is stable enough for tests

## 5. ACP Backends

Add persistent ACP-backed model runners.

Scope:

- `claude-agent-acp`
- `codex-acp`
- one session per participant per run
- model, effort, system prompt, and read-only intent mapped into transport specs
- response classification and retry surface
- tests with fake ACP process plumbing

Expected result:

- Claude and Codex backend ids resolve through the default resolver
- ACP process setup is covered by unit tests
- runtime failures are classified as retryable or terminal where possible

## 6. Exec Backend

Add stateless CLI subprocess execution for non-ACP tools.

Scope:

- `agy` backend id
- prompt passed to stdin
- response read from stdout
- command override through environment where supported
- full-context rendering for every turn
- fake CLI tests

Expected result:

- Gemini-style persona configs can resolve to `agy`
- exec failures produce clear errors
- subprocess behavior is covered without real model calls

## 7. Scaffolding

Make project setup ergonomic.

Scope:

- `debate init`
- `debate new <name>`
- `debate new --role synthesizer <name>`
- persona-name validation
- no clobbering existing files

Expected result:

- `debate init` creates `.heurema/debate/personas/proposer.md` and `skeptic.md`
- generated personas parse and validate
- repeated init skips existing files
- `debate new` discovers the workspace and creates exactly one persona file

## Deferred

Not in the current v1 slice set:

- API backend
- API grounding tool loop
- allow and deny lists for filesystem grounding
- network allowlist
- persona selection by tags
- default tag behavior for newly generated personas
- multi-run pipeline orchestration
