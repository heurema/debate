# Memory Candidate

## Run
- Run id: run_20260624_095835
- Source: deterministic

## Contract
- Goal: Slice 4: wire the cmd/debate CLI into a working debate on a deterministic offline backend (no real models yet). The command debate "<task>" plus version loads the .heurema/debate workspace via config.Load, assembles the brief (workspace context followed by the task; task from positional arg, --task @file, or stdin), builds an orchestrate.Config (participants from the panel personas via a backend resolver, a RoundRobin scheduler, prompt.NewPromptBuilder, verdict.New), runs orchestrate.Run, then runs the synthesizer once to produce the final answer. Flags: --with, --synth, --max-rounds, --json, -q, --sealed. Output contract: stdout is the answer, stderr is the live debate trace (auto-quiet off-TTY or with -q), exit 0 settled, 2 not-converged (stalemate or max), 1 error. A backend registry resolves persona.Backend to a transport; register a deterministic offline echo backend (canned reply with a valid signal block, no network) and accept an injectable resolver so tests use the engine mock backend; real acp/exec/api backends are out of scope. Fail-fast validation before opening any session. e2e tests over a fixture .heurema/debate workspace. cmd/debate uses the stdlib flag package (no third-party CLI lib), internal/debate, and internal/engine. Out of scope: real backends, debate init/new scaffolding, and the real grounded sandbox behind --sealed.
- In scope:
  - Implement the cmd/debate CLI: a default run command `debate "<task>"` plus `version`, with flags --with, --synth, --max-rounds, --json, -q/--quiet, --sealed, taking the task from a positional arg, --task @file, or stdin.
  - Implement a core debate runner that loads the workspace (config.Load), assembles the brief, builds an orchestrate.Config (participants, RoundRobin, prompt.NewPromptBuilder, verdict.New, loop.Limits), runs orchestrate.Run, then runs the synthesizer once to produce the final answer.
  - Implement a backend registry that resolves persona.Backend to a transport and register a deterministic offline echo backend (canned reply with a valid signal block) usable with no network; the runner takes an injectable resolver for tests.
  - Implement the IO and exit-code contract: stdout = final answer, stderr = live debate trace (auto-quiet off-TTY or with -q), exit 0 settled / 2 not-converged / 1 error, and a --json structured result.
  - Add fail-fast validation (workspace, personas, non-empty panel, non-empty task) before any session opens, and e2e tests driving the runner with an injected mock/echo backend over a fixture .heurema/debate workspace.
- Out of scope:
  - Real acp/exec/api backends and any real network, model, or subprocess call (later slices).
  - debate init / new scaffolding (a later slice).
  - The grounded read-only sandbox semantics behind --sealed (the flag is parsed and threaded; real grounding is the acp slice).
  - Modifying internal/engine or the Slice 2/3 signal/prompt/verdict/persona/config packages beyond what import requires.
- Acceptance criteria:
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
- Validation commands:
  - bash scripts/check-gofmt.sh
  - go test -count=1 ./...
  - go vet ./...
  - go run ./cmd/debate version
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

## Outcome
- Gate status: needs_review
- Review status: approved
- Execution exit code: 0
- Validation passed: true
- Changes need review: true

## Changes
- Changed files:
  - cmd/debate/main.go
- New files:
  - cmd/debate/e2e_test.go
  - docs/PACTUM-ISSUES-ANALYSIS.md
  - internal/debate/runner/runner.go
  - internal/debate/runner/runner_test.go
  - internal/engine/transport/echo/echo.go
- Missing files: none

## Clarifications
- None

## Review Decisions
- f_001 [medium] open internal/debate/runner/runner.go:116: The synthesizer's backend is resolved and opened only after the full debate runs, so it is never validated fail-fast. An unimplemented synthesizer backend fails only after the entire debate completes, and with the production defaultResolver the default synthesizer (model claude-haiku-4-5 -> backend claude-agent-acp) is rejected, so a valid echo-only offline workspace with no explicit synthesizer persona opens its panel, runs the debate, then errors at synthesis. The canonical offline run cannot complete out of the box with the production resolver.
- f_002 [low] open cmd/debate/main.go:113: Flags placed after the positional task are silently ignored. Go's flag package stops parsing at the first non-flag argument, and the task is taken from fs.Args(), so `debate "<task>" --json` treats --json as part of the positional task and produces prose output instead of JSON; the same applies to -q, --quiet, --sealed, --max-rounds, etc. when they follow the task.
- f_003 [medium] open internal/debate/runner/runner.go:138: The production binary cannot complete an offline end-to-end run via the default resolver: a workspace with echo-backed debaters but no explicit synthesizer persona falls back to the built-in default synthesizer (claude-haiku-4-5), whose inferred backend (claude-agent-acp) is rejected by defaultResolver. The debate converges on echo, then synthesis fails with exit 1. This undercuts the contract's 'fully wired, offline-runnable debate using a deterministic echo backend' deliverable for the default-resolver path.
- f_004 [low] open cmd/debate/e2e_test.go:17: AC11 requires the default production resolver to wire the echo backend, but no test exercises defaultResolver through a successful end-to-end debate. The only test using defaultResolver is the unimplemented-backend failure case; the settled/JSON/non-TTY happy-path tests use the test-only echoAll resolver, which returns echo for every backend (including the synthesizer's inferred acp backend) and therefore masks the default-synthesizer gap.
- f_005 [low] open docs/PACTUM-ISSUES-ANALYSIS.md:1: The change set includes docs/PACTUM-ISSUES-ANALYSIS.md, a Pactum-maintainer bug-analysis document unrelated to Slice 4 (the debate CLI). It is outside the contract scope and appears as a new file in this run's gate change summary.
- f_006 [medium] open cmd/debate/main.go:213: stdin ingestion and multi-source task composition are untested. assembleTask reads piped stdin and joins multiple task sources with a newline, but no test exercises either path: every parseAndRun call passes an empty stdin reader, and no test supplies more than one source at once. A regression in stdin handling or the newline-join composition would not be caught.
- f_007 [low] open cmd/debate/main.go:108: The unknown-flag error path and the `version` output format have no test coverage, despite both being explicit acceptance criteria for this slice. The parseAndRun parse-error branch (return exit 1 + usage on an unrecognized flag) is never exercised, and `debate version` output is asserted only by a tautological test.
- f_008 [low] open cmd/debate/main.go:187: outcomeString is a no-op wrapper: both the explicit case ("settled", "stalemate", "max") and the default branch return the input reason unchanged, so the function normalises nothing despite its doc comment claiming to. The single caller could use result.Outcome.Reason directly.
- f_009 [low] open cmd/debate/e2e_test.go:262: The runner package is imported in e2e_test.go solely to satisfy a blank assignment (var _ = runner.Config{} // ensure import is used). The import and the assignment add nothing and can both be removed; the adjacent comment 'Keep the existing TestVersion passing' is misleading since TestVersion lives in main_test.go and is unrelated to this import.
- Proposal summary: pending=0 accepted=9 rejected=0

## Reusable Project Knowledge
- scope: in scope: Implement the cmd/debate CLI: a default run command `debate "<task>"` plus `version`, with flags --with, --synth, --max-rounds, --json, -q/--quiet, --sealed, taking the task from a positional arg, --task @file, or stdin.
- scope: in scope: Implement a core debate runner that loads the workspace (config.Load), assembles the brief, builds an orchestrate.Config (participants, RoundRobin, prompt.NewPromptBuilder, verdict.New, loop.Limits), runs orchestrate.Run, then runs the synthesizer once to produce the final answer.
- scope: in scope: Implement a backend registry that resolves persona.Backend to a transport and register a deterministic offline echo backend (canned reply with a valid signal block) usable with no network; the runner takes an injectable resolver for tests.
- scope: in scope: Implement the IO and exit-code contract: stdout = final answer, stderr = live debate trace (auto-quiet off-TTY or with -q), exit 0 settled / 2 not-converged / 1 error, and a --json structured result.
- scope: in scope: Add fail-fast validation (workspace, personas, non-empty panel, non-empty task) before any session opens, and e2e tests driving the runner with an injected mock/echo backend over a fixture .heurema/debate workspace.
- scope: out of scope: Real acp/exec/api backends and any real network, model, or subprocess call (later slices).
- scope: out of scope: debate init / new scaffolding (a later slice).
- scope: out of scope: The grounded read-only sandbox semantics behind --sealed (the flag is parsed and threaded; real grounding is the acp slice).
- scope: out of scope: Modifying internal/engine or the Slice 2/3 signal/prompt/verdict/persona/config packages beyond what import requires.
- review_resolution: proposal p_001 accepted as f_001
- review_resolution: proposal p_002 accepted as f_002
- review_resolution: proposal p_003 accepted as f_003
- review_resolution: proposal p_004 accepted as f_004
- review_resolution: proposal p_005 accepted as f_005
- review_resolution: proposal p_006 accepted as f_006
- review_resolution: proposal p_007 accepted as f_007
- review_resolution: proposal p_008 accepted as f_008
- review_resolution: proposal p_009 accepted as f_009
- validation: bash scripts/check-gofmt.sh passed
- validation: go test -count=1 ./... passed
- validation: go vet ./... passed
- validation: go run ./cmd/debate version passed
- validation: bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine passed
- validation: bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3 passed

## Artifacts
- Contract: contract/contract.json
- Gate report: gate/gate-report.json
- Review: review/review.json
- Findings: review/findings.jsonl
- Resolutions: review/resolutions.jsonl
- Proposals: review/proposals.jsonl
- Proposal decisions: review/proposal-decisions.jsonl
