# Contract Drafter Context

## Run
- Run id: run_20260624_092233
- Run status: contract_draft

## Contract goal
Slice 3: implement persona loading, .heurema/debate workspace discovery, config, and panel selection in internal/debate, fixture-tested only (no engine run, no real backends, no CLI binary). (1) Persona: a persona is a markdown file with YAML frontmatter (role: debater|synthesizer defaulting to debater; model; effort; optional backend; optional tags list) plus a markdown body that is the system prompt; parse and fail-fast-validate it (reject unknown frontmatter keys; require model and effort for api/acp backends; the persona id is its filename without .md). (2) Backend inference: when backend is absent, infer it from the model name (claude-*/opus/sonnet -> claude-agent-acp; gpt-*/codex/o* -> codex-acp; gemini-* -> agy); an explicit backend overrides inference. (3) Discovery: locate .heurema/debate/ by walking up from a start directory like git does; load an optional config.yml whose only key is table (a list/selection of persona names); load an optional context.md baseline preamble; load personas from .heurema/debate/personas/*.md. (4) Selection: resolve the debater panel from config.table or an explicit list of names (when config is absent, the panel is all debater personas); personas with role synthesizer are excluded from the panel. (5) Synthesizer resolution: choose by an explicit name override, else the persona named synthesizer, else a built-in default (model claude-haiku-4-5 with a minimal prompt). (6) Fail-fast: a clear error before anything else for unknown keys, missing required fields, an unresolvable selection, or a missing .heurema/debate. YAML frontmatter and config parsing may use gopkg.in/yaml.v3. Unit tests use fixture .heurema/debate directories. Out of scope: running the engine, real acp/exec/api transports, the cmd/debate CLI wiring, actually invoking models, and synthesizer execution.

## Current contract fields
- In scope:
  - none
- Out of scope:
  - none
- Acceptance criteria:
  - none
- Validation commands:
  - none
- Assumptions:
  - none

## Answered clarifications
- None

## Repository context
# Repository Context

Generated: 2026-06-24T09:22:33Z

Map run: map_20260624_090603
Repo map path: .heurema/pactum/map/repo-map.md
LLMS path: .heurema/pactum/map/llms.txt
Search index path: .heurema/pactum/map/search.sqlite
Accepted memory context: context/memory-context.md

Notes:
- Pactum has not yet done agentic clarification.
- This is deterministic context assembled from existing map artifacts.

## Project map

# Pactum Project Map

Generated: 2026-06-24T09:06:03Z

Repository root: `.`

## Summary

- Indexed files: 21
- Ignored files/directories: 343
- Detected languages: 5
- Code items (best-effort hints): 119

## How to navigate this map

- Start with the wiki: read `wiki/overview.md` first.
- The wiki is generated from deterministic facts (file inventory and manifests).
- Code items are best-effort navigation hints, not complete semantic truth.
- Unsupported languages/framework files may have no code items.
- Imports are not treated as primary code surface.
- Source files remain the source of truth.

## Wiki pages

- `wiki/overview.md` — Project map overview
- `wiki/structure.md` — Project structure
- `wiki/commands.md` — Commands
- `wiki/entrypoints.md` — Candidate entrypoints
- `wiki/config.md` — Configuration
- `wiki/tests.md` — Tests
- Area pages:
  - `wiki/areas/.claude.md`
  - `wiki/areas/cmd.md`
  - `wiki/areas/docs.md`
  - `wiki/areas/internal.md`
  - `wiki/areas/scripts.md`

## Project map artifacts

- `files.jsonl` — deterministic per-file metadata.
- `hashes.jsonl` — per-file content hashes.
- `code-items.jsonl` — best-effort symbol hints (incomplete by design).
- `search.sqlite` — local full-text search index.
- `manifest.json` — map manifest listing all artifacts.

## Files / areas

### Detected languages

- Go: 11 file(s)
- Markdown: 6 file(s)
- Go module: 1 file(s)
- Make: 1 file(s)
- Shell: 1 file(s)

### Top-level directories

- `.claude/` (see `wiki/areas/.claude.md`)
- `cmd/` (see `wiki/areas/cmd.md`)
- `docs/` (see `wiki/areas/docs.md`)
- `internal/` (see `wiki/areas/internal.md`)
- `scripts/` (see `wiki/areas/scripts.md`)

### Important files

- `go.mod`
- `Makefile`

### File tree

- `.claude/skills/pactum/...`
- `.gitignore`
- `Makefile`
- `cmd/debate/main.go`
- `cmd/debate/main_test.go`
- `docs/DESIGN.md`
- `docs/SLICES.md`
- `go.mod`
- `internal/debate/debate.go`
- `internal/engine/loop/...`
- `internal/engine/orchestrate/...`
- `internal/engine/transport/...`
- `scripts/dep-guard.sh`

## Code surface (best-effort code hints)

- `cmd/debate/main.go`: `go_main` `main`
- `cmd/debate/main.go`: `go_main` `main.main`
- `cmd/debate/main_test.go`: `go_func` `main.TestVersion`
- `internal/engine/loop/loop.go`: `go_func` `loop.Run`
- `internal/engine/loop/loop.go`: `go_type` `loop.Limits`
- `internal/engine/loop/loop.go`: `go_type` `loop.Outcome`
- `internal/engine/loop/loop.go`: `go_type` `loop.RoundContext`
- `internal/engine/loop/loop.go`: `go_type` `loop.RoundResult`
- `internal/engine/loop/loop.go`: `go_type` `loop.Step`
- `internal/engine/loop/loop.go`: `go_type` `loop.Stop`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestImmediateStop`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestInvalidLimits`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestMax`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestPreRoundCtxAlreadyCancelled`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestPreRoundCtxCancellation`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestRoundContextNumbers`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestSettled`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestStalemate`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestStalemateResetByProgress`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestStepError`
- `internal/engine/loop/loop_test.go`: `go_func` `loop_test.TestStepErrorOnFirstRound`
- `internal/engine/orchestrate/orchestrate.go`: `go_func` `orchestrate.RoundRobin`
- `internal/engine/orchestrate/orchestrate.go`: `go_func` `orchestrate.Run`
- `internal/engine/orchestrate/orchestrate.go`: `go_method` `Transcript.All`
- `internal/engine/orchestrate/orchestrate.go`: `go_method` `Transcript.Append`
- `internal/engine/orchestrate/orchestrate.go`: `go_method` `Transcript.DeltaFor`
- `internal/engine/orchestrate/orchestrate.go`: `go_method` `Transcript.Len`
- `internal/engine/orchestrate/orchestrate.go`: `go_type` `orchestrate.Config`
- `internal/engine/orchestrate/orchestrate.go`: `go_type` `orchestrate.Participant`
- `internal/engine/orchestrate/orchestrate.go`: `go_type` `orchestrate.PromptBuilder`
- `internal/engine/orchestrate/orchestrate.go`: `go_type` `orchestrate.RenderMode`
- `internal/engine/orchestrate/orchestrate.go`: `go_type` `orchestrate.Result`
- `internal/engine/orchestrate/orchestrate.go`: `go_type` `orchestrate.Scheduler`
- `internal/engine/orchestrate/orchestrate.go`: `go_type` `orchestrate.Transcript`
- `internal/engine/orchestrate/orchestrate.go`: `go_type` `orchestrate.Turn`
- `internal/engine/orchestrate/orchestrate.go`: `go_type` `orchestrate.Verdict`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestDeltaForNextRound`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestDeltaForSameRound`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestMissingConfigFields`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestOnTurnCallback`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestOutcomeMax`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestOutcomeSettled`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestOutcomeStalemate`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestSessionSendErrorSurfaced`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestTranscriptAccumulation`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestTurnOrderFixedRoundRobin`
- `internal/engine/orchestrate/orchestrate_test.go`: `go_func` `orchestrate_test.TestTurnOrderRotatingRoundRobin`
- `internal/engine/transport/mock/mock.go`: `go_func` `mock.NewSession`
- `internal/engine/transport/mock/mock.go`: `go_func` `mock.NewTransport`
- `internal/engine/transport/mock/mock.go`: `go_method` `Session.Close`
- `internal/engine/transport/mock/mock.go`: `go_method` `Session.Closed`
- `internal/engine/transport/mock/mock.go`: `go_method` `Session.Prompts`
- `internal/engine/transport/mock/mock.go`: `go_method` `Session.Send`
- `internal/engine/transport/mock/mock.go`: `go_method` `Transport.Open`
- `internal/engine/transport/mock/mock.go`: `go_type` `mock.ScriptedResult`
- `internal/engine/transport/mock/mock.go`: `go_type` `mock.Session`
- `internal/engine/transport/mock/mock.go`: `go_type` `mock.Transport`
- `internal/engine/transport/mock/mock_test.go`: `go_func` `mock_test.TestSessionCloseIdempotent`
- `internal/engine/transport/mock/mock_test.go`: `go_func` `mock_test.TestSessionExhausted`
- `internal/engine/transport/mock/mock_test.go`: `go_func` `mock_test.TestSessionRecordsPrompts`
- `internal/engine/transport/mock/mock_test.go`: `go_func` `mock_test.TestSessionScriptedError`
- `internal/engine/transport/mock/mock_test.go`: `go_func` `mock_test.TestSessionScriptedResults`
- `internal/engine/transport/mock/mock_test.go`: `go_func` `mock_test.TestTransportOpen`
- `internal/engine/transport/mock/mock_test.go`: `go_func` `mock_test.TestTransportOpenUnknown`
- `internal/engine/transport/transport.go`: `go_func` `transport.Classify`
- `internal/engine/transport/transport.go`: `go_type` `transport.ErrorClass`
- `internal/engine/transport/transport.go`: `go_type` `transport.Result`
- `internal/engine/transport/transport.go`: `go_type` `transport.Session`
- `internal/engine/transport/transport.go`: `go_type` `transport.Spec`
- `internal/engine/transport/transport.go`: `go_type` `transport.Transport`
- `internal/engine/transport/transport.go`: `go_type` `transport.Usage`
- `internal/engine/transport/transport_test.go`: `go_func` `transport_test.TestClassifyBaresentinel`
- `internal/engine/transport/transport_test.go`: `go_func` `transport_test.TestClassifyNil`
- `internal/engine/transport/transport_test.go`: `go_func` `transport_test.TestClassifyUnknown`
- `internal/engine/transport/transport_test.go`: `go_func` `transport_test.TestClassifyWrappedSentinel`

## Language support

- File metadata is collected for common source, config, and documentation files.
- Best-effort code hints are extracted for the starter language pack: Go, Python, JavaScript, TypeScript/TSX/JSX, and C#.
- Code items are best-effort navigation hints; imports are not treated as primary code surface.
- Unsupported languages/framework files may have no code items but still appear in the wiki and file inventory.
- Pactum does not perform LSP, references, call graph, or semantic analysis in this phase.
- The map is a navigation aid, not complete semantic truth.
- Source files remain the source of truth.

## Agent guidance

- Read the wiki first (`wiki/overview.md`), then drill into the relevant area page.
- Before adding new code, search/read relevant files and code items.
- Prefer existing exported functions/types when applicable.
- If ownership is unclear, ask for clarification instead of guessing.

## Search results
{
  "query": "Slice 3: implement persona loading, .heurema/debate workspace discovery, config, and panel selection in internal/debate, fixture-tested only (no engine run, no real backends, no CLI binary). (1) Persona: a persona is a markdown file with YAML frontmatter (role: debater|synthesizer defaulting to debater; model; effort; optional backend; optional tags list) plus a markdown body that is the system prompt; parse and fail-fast-validate it (reject unknown frontmatter keys; require model and effort for api/acp backends; the persona id is its filename without .md). (2) Backend inference: when backend is absent, infer it from the model name (claude-*/opus/sonnet -\u003e claude-agent-acp; gpt-*/codex/o* -\u003e codex-acp; gemini-* -\u003e agy); an explicit backend overrides inference. (3) Discovery: locate .heurema/debate/ by walking up from a start directory like git does; load an optional config.yml whose only key is table (a list/selection of persona names); load an optional context.md baseline preamble; load personas from .heurema/debate/personas/*.md. (4) Selection: resolve the debater panel from config.table or an explicit list of names (when config is absent, the panel is all debater personas); personas with role synthesizer are excluded from the panel. (5) Synthesizer resolution: choose by an explicit name override, else the persona named synthesizer, else a built-in default (model claude-haiku-4-5 with a minimal prompt). (6) Fail-fast: a clear error before anything else for unknown keys, missing required fields, an unresolvable selection, or a missing .heurema/debate. YAML frontmatter and config parsing may use gopkg.in/yaml.v3. Unit tests use fixture .heurema/debate directories. Out of scope: running the engine, real acp/exec/api transports, the cmd/debate CLI wiring, actually invoking models, and synthesizer execution.",
  "queries": [
    "heurema/debate",
    "internal/debate",
    "api/acp",
    "claude-*/opus/sonnet",
    "gpt-*/codex/o*",
    "heurema/debate/",
    "config.yml",
    "list/selection"
  ],
  "query_source": "task",
  "results": [],
  "warnings": [
    "Search index is stale. Run: pactum map refresh."
  ]
}

## Drafter guidance
- Propose only additions to the contract fields listed in the prompt.
- Do not change or restate the contract goal.
- Do not answer clarification questions.
- Do not edit files.
- Treat repository map/search context as navigation hints, not semantic truth.
