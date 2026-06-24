# Contract Draft

## Goal
Slice 4: wire the cmd/debate CLI into a working debate on a deterministic offline backend (no real models yet). The command debate "<task>" plus version loads the .heurema/debate workspace via config.Load, assembles the brief (workspace context followed by the task; task from positional arg, --task @file, or stdin), builds an orchestrate.Config (participants from the panel personas via a backend resolver, a RoundRobin scheduler, prompt.NewPromptBuilder, verdict.New), runs orchestrate.Run, then runs the synthesizer once to produce the final answer. Flags: --with, --synth, --max-rounds, --json, -q, --sealed. Output contract: stdout is the answer, stderr is the live debate trace (auto-quiet off-TTY or with -q), exit 0 settled, 2 not-converged (stalemate or max), 1 error. A backend registry resolves persona.Backend to a transport; register a deterministic offline echo backend (canned reply with a valid signal block, no network) and accept an injectable resolver so tests use the engine mock backend; real acp/exec/api backends are out of scope. Fail-fast validation before opening any session. e2e tests over a fixture .heurema/debate workspace. cmd/debate uses the stdlib flag package (no third-party CLI lib), internal/debate, and internal/engine. Out of scope: real backends, debate init/new scaffolding, and the real grounded sandbox behind --sealed.

## Current status
Contract status: approved
Manual clarification, contract approval, prompt build, and agent execution are available through staged Pactum commands.

## Relevant repository context
- Map run: map_20260624_093601
- Repo map: .heurema/pactum/map/repo-map.md
- Search results: context/search-results.json (0 result(s))

## Clarifications
- None

## In scope
- Implement the cmd/debate CLI: a default run command `debate "<task>"` plus `version`, with flags --with, --synth, --max-rounds, --json, -q/--quiet, --sealed, taking the task from a positional arg, --task @file, or stdin.
- Implement a core debate runner that loads the workspace (config.Load), assembles the brief, builds an orchestrate.Config (participants, RoundRobin, prompt.NewPromptBuilder, verdict.New, loop.Limits), runs orchestrate.Run, then runs the synthesizer once to produce the final answer.
- Implement a backend registry that resolves persona.Backend to a transport and register a deterministic offline echo backend (canned reply with a valid signal block) usable with no network; the runner takes an injectable resolver for tests.
- Implement the IO and exit-code contract: stdout = final answer, stderr = live debate trace (auto-quiet off-TTY or with -q), exit 0 settled / 2 not-converged / 1 error, and a --json structured result.
- Add fail-fast validation (workspace, personas, non-empty panel, non-empty task) before any session opens, and e2e tests driving the runner with an injected mock/echo backend over a fixture .heurema/debate workspace.

## Out of scope
- Real acp/exec/api backends and any real network, model, or subprocess call (later slices).
- debate init / new scaffolding (a later slice).
- The grounded read-only sandbox semantics behind --sealed (the flag is parsed and threaded; real grounding is the acp slice).
- Modifying internal/engine or the Slice 2/3 signal/prompt/verdict/persona/config packages beyond what import requires.

## Acceptance criteria
- cmd/debate runs as `debate "<task>"`: the task is read from the positional argument, or --task @file (the file contents), or stdin when piped; when more than one source is present they compose (stdin is appended); an empty resulting task is a fail-fast error before any session opens.
- `debate version` prints the binary version and exits 0; an unknown flag or subcommand prints a clear usage message and exits non-zero.
- The runner calls config.Load(startDir, withList, synthOverride) with withList from --with and synthOverride from --synth, and reports any workspace/persona/selection error fail-fast (exit 1) before opening any session.
- The brief given to prompt.NewPromptBuilder is the assembled text of Workspace.Context (baseline) followed by the task; --sealed sets a brief-only/read-only intent that is threaded into the transport.Spec (ReadOnly) for later grounding.
- For each persona in Workspace.Panel (in order) the runner builds a transport.Spec from the persona (ID, Model, Effort, System, ReadOnly) and opens a Session via the backend resolver keyed by persona.Backend, then builds orchestrate.Participant{ID, Session} preserving panel order.
- The runner builds orchestrate.Config with those participants, an orchestrate RoundRobin scheduler, prompt.NewPromptBuilder(brief), verdict.New(until) (until defaulting to all_done), and loop.Limits whose Max comes from --max-rounds (default 10) and whose Settle/Patience are built-in code defaults; it then calls orchestrate.Run.
- After orchestrate.Run the runner invokes the synthesizer exactly once: it opens a Session for Workspace.Synthesizer via the resolver, sends a synthesis prompt built from the task and the final transcript, and uses the returned content as the final answer. The synthesizer never takes part in the debate rounds.
- stdout receives only the final answer; the live debate trace is written to stderr; stderr tracing auto-quiets when stderr is not a TTY or when -q/--quiet is set, unless DEBATE_FORCE_TRACE=1 is set in the environment (which overrides TTY detection and forces trace output regardless); --json implies machine-readable output on stdout and suppresses the human stderr trace; with --json the command writes to stdout a JSON object with exactly these top-level keys: `answer` (string, the synthesizer reply), `outcome` (string, one of "settled", "stalemate", or "max"), `rounds` (integer, count of completed debate rounds), and `turns` (array of objects each containing `round` (1-indexed integer), `speaker` (string, persona ID), and `content` (string, turn reply)) — no other top-level keys are present.
- The process exit code is 0 when the Outcome reason is settled, 2 when it is stalemate or max (did not converge), and 1 on any error; this mapping is documented in code and covered by tests.
- A backend registry resolves persona.Backend to a transport; a deterministic offline `echo` backend is registered that returns, with no network/model/subprocess call, a canned reply containing a valid signal block so a debate can converge; an unimplemented backend (claude-agent-acp, codex-acp, agy) is a clear fail-fast error in this slice.
- The runner accepts an injectable backend resolver so tests supply a scripted mock backend (internal/engine/transport/mock); the default production resolver wires the echo backend.
- e2e tests over a fixture .heurema/debate workspace assert: (a) a full run with DEBATE_FORCE_TRACE=1 prints a synthesized answer on stdout and the debate trace on stderr, and returns exit 0 for a settled run; (b) a non-converged run returns exit 2; (c) a run without DEBATE_FORCE_TRACE in a non-TTY environment produces empty stderr (auto-quiet path); (d) an empty task fails fast with exit 1; (e) an unimplemented backend fails fast with exit 1.
- cmd/debate uses only the Go standard library (the flag package for parsing, no third-party CLI framework), internal/debate/..., and internal/engine/...; internal/engine and the Slice 2/3 packages are not modified; check-gofmt, go vet ./..., and go test ./... pass.

## Validation commands
- bash scripts/check-gofmt.sh
- go test -count=1 ./...
- go vet ./...
- go run ./cmd/debate version
- bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
- bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

## Assumptions
- The first real model-backed run is the acp slice; this slice's deliverable is a fully wired, offline-runnable debate using a deterministic echo backend plus an injectable resolver for tests.
- The synthesizer is invoked once after the debate through its own session and never participates in the debate rounds.
- Settle and Patience are built-in code defaults; only Max is exposed (via --max-rounds) with a default of 10; until defaults to all_done.
- --sealed toggles a brief-only/read-only intent threaded into transport.Spec.ReadOnly; the grounded read-only sandbox itself is realized in the acp slice.
- Auto-quiet uses TTY detection on stderr; -q forces quiet; --json implies machine-readable output on stdout and suppresses the human stderr trace; DEBATE_FORCE_TRACE=1 in the environment overrides TTY detection and forces trace output to stderr — e2e tests set this variable to capture and assert on the trace, and a separate e2e assertion verifies stderr is empty in the non-TTY/non-forced path.
- cmd/debate parses arguments with the standard library flag package; no third-party CLI dependency is introduced.
- The echo backend exists for offline runs and demos; it produces deterministic content and a valid signal and is not a real model.

## Open questions
- None
