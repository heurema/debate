# Contract Drafter Context

## Run
- Run id: run_20260624_120335
- Run status: contract_draft

## Contract goal
Slice 6: implement the exec backend transport in internal/backend/exec and wire it into the cmd/debate production resolver, so debate can use plain-CLI agents (Gemini via agy). The exec transport implements transport.Transport by driving a stateless CLI: because the CLI keeps no session state, the Session accumulates every prompt it receives (the per-round deltas from orchestrate) into a running transcript, and on each Send spawns a fresh subprocess, writes the full accumulated prompt (system prompt folded in, plus the brief and the conversation so far and the new delta) to stdin, reads stdout to completion as the agent reply, and returns it as transport.Result. The command is derived from the backend (agy for the agy backend) and is overridable via env; spec.Model is threaded in and spec.System is folded into the prompt; grounding applies (subprocess working directory is the project dir for grounded and a fresh empty temp dir for sealed). Open prepares the Session, Close cleans up (no persistent process). Errors map via transport.Classify consistent with its categories without changing internal/engine. Testing: deterministic unit tests use a fake CLI command (a stub program or an injectable command runner) to assert delta accumulation and full-render, stdin/stdout handling, model and command wiring with env override, grounded vs sealed working directory, and error mapping; a real-agy integration test is gated behind a build tag and env var that the gate compiles but skips. The production resolver registers exec for the agy backend; acp and echo stay. internal/backend/exec depends only on stdlib, internal/engine, and internal/backend. Out of scope: the api backend, any internal/engine or internal/debate change beyond import, and scaffolding.

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

Generated: 2026-06-24T12:03:35Z

Map run: map_20260624_112007
Repo map path: .heurema/pactum/map/repo-map.md
LLMS path: .heurema/pactum/map/llms.txt
Search index path: .heurema/pactum/map/search.sqlite
Accepted memory context: context/memory-context.md

Notes:
- Pactum has not yet done agentic clarification.
- This is deterministic context assembled from existing map artifacts.

## Project map

# Pactum Project Map

Generated: 2026-06-24T11:20:07Z

Repository root: `.`

## Summary

- Indexed files: 37
- Ignored files/directories: 717
- Detected languages: 5
- Code items (best-effort hints): 303

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

- Go: 25 file(s)
- Markdown: 6 file(s)
- Go module: 2 file(s)
- Shell: 2 file(s)
- Make: 1 file(s)

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
- `cmd/debate/e2e_test.go`
- `cmd/debate/main.go`
- `cmd/debate/main_test.go`
- `docs/DESIGN.md`
- `docs/SLICES.md`
- `go.mod`
- `go.sum`
- `internal/debate/config/...`
- `internal/debate/debate.go`
- `internal/debate/persona/...`
- `internal/debate/prompt/...`
- `internal/debate/runner/...`
- `internal/debate/signal/...`
- `internal/debate/verdict/...`
- `internal/engine/loop/...`
- `internal/engine/orchestrate/...`
- `internal/engine/transport/...`
- `scripts/check-gofmt.sh`
- `scripts/dep-guard.sh`

## Code surface (best-effort code hints)

- `cmd/debate/main.go`: `go_main` `main`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_EmptyTask_Exit1`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_JSONOutput`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_NonTTY_NoTrace`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_NotConverged_Exit2`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_SettledRun_WithTrace`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_TaskFromFile`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_UnimplementedBackend_Exit1`
- `cmd/debate/main.go`: `go_main` `main.main`
- `cmd/debate/main_test.go`: `go_func` `main.TestVersion`
- `internal/debate/config/config.go`: `go_func` `config.Discover`
- `internal/debate/config/config.go`: `go_func` `config.Load`
- `internal/debate/config/config.go`: `go_type` `config.Workspace`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestDiscover_FindsFromChildDir`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestDiscover_MissingDir`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_DefaultPanel`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_DefaultSynthesizer`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_EmptyBody`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_MissingDebateDir`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_MissingModel`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_NoContextMD`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_SelectorNamesSynthesizerRole`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_SynthOverride`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_SynthOverrideMissing`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_UninfernableModel`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_UnknownConfigKey`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_UnknownFrontmatterKey`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_UnresolvableSelectionName`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_ValidWorkspace`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_WithListOverridesConfig`
- `internal/debate/config/config_test.go`: `go_func` `config_test.TestLoad_WithListSynthesizerRole`
- `internal/debate/persona/persona.go`: `go_func` `persona.InferBackend`
- `internal/debate/persona/persona.go`: `go_func` `persona.ParseFile`
- `internal/debate/persona/persona.go`: `go_type` `persona.Persona`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestInferBackend`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_EmptyBody`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_ExplicitBackendOverridesInference`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_GeminiBackend`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_InvalidRole`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_MissingEffort`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_MissingModel`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_NoFrontmatterFails`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_RoleDefaultsToDebater`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_UninfernableModel`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_UnknownFrontmatterKey`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_ValidDebater`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_ValidSynthesizer`
- `internal/debate/persona/persona_test.go`: `go_func` `persona_test.TestParseFile_WhitespaceOnlyBody`
- `internal/debate/prompt/prompt.go`: `go_func` `prompt.NewPromptBuilder`
- `internal/debate/prompt/prompt_test.go`: `go_func` `prompt_test.TestPromptBoardLabelsTurnsByRoundAndSpeaker`
- `internal/debate/prompt/prompt_test.go`: `go_func` `prompt_test.TestPromptContainsBrief`
- `internal/debate/prompt/prompt_test.go`: `go_func` `prompt_test.TestPromptContainsSignalInstruction`
- `internal/debate/prompt/prompt_test.go`: `go_func` `prompt_test.TestPromptDeltaMode`
- `internal/debate/prompt/prompt_test.go`: `go_func` `prompt_test.TestPromptEmptyTranscript`
- `internal/debate/prompt/prompt_test.go`: `go_func` `prompt_test.TestPromptFullMode`
- `internal/debate/runner/runner.go`: `go_func` `runner.Run`
- `internal/debate/runner/runner.go`: `go_type` `runner.Config`
- `internal/debate/runner/runner.go`: `go_type` `runner.Resolver`
- `internal/debate/runner/runner.go`: `go_type` `runner.Result`
- `internal/debate/runner/runner_test.go`: `go_func` `runner_test.TestRun_EmptyTask`
- `internal/debate/runner/runner_test.go`: `go_func` `runner_test.TestRun_MaxOutcome`
- `internal/debate/runner/runner_test.go`: `go_func` `runner_test.TestRun_OnTurnFires`
- `internal/debate/runner/runner_test.go`: `go_func` `runner_test.TestRun_Settled`
- `internal/debate/signal/signal.go`: `go_func` `signal.Parse`
- `internal/debate/signal/signal.go`: `go_type` `signal.Signal`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_DoneWithObjections_Invariant`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_EmptyObject`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_GarbledJSON`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_LastBlockUsed`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_MultiLineJSON`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_NoBlock`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_NonObjectJSON`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_NonSignalFencedBlock`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_NullJSON`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_TrailingProse`
- `internal/debate/signal/signal_test.go`: `go_func` `signal_test.TestParse_WellFormed`
- `internal/debate/verdict/verdict.go`: `go_func` `verdict.New`
- `internal/debate/verdict/verdict.go`: `go_type` `verdict.Until`
- `internal/debate/verdict/verdict_test.go`: `go_func` `verdict_test.TestVerdictMax`
- `internal/debate/verdict/verdict_test.go`: `go_func` `verdict_test.TestVerdictProgressTracking`

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
  "query": "Slice 6: implement the exec backend transport in internal/backend/exec and wire it into the cmd/debate production resolver, so debate can use plain-CLI agents (Gemini via agy). The exec transport implements transport.Transport by driving a stateless CLI: because the CLI keeps no session state, the Session accumulates every prompt it receives (the per-round deltas from orchestrate) into a running transcript, and on each Send spawns a fresh subprocess, writes the full accumulated prompt (system prompt folded in, plus the brief and the conversation so far and the new delta) to stdin, reads stdout to completion as the agent reply, and returns it as transport.Result. The command is derived from the backend (agy for the agy backend) and is overridable via env; spec.Model is threaded in and spec.System is folded into the prompt; grounding applies (subprocess working directory is the project dir for grounded and a fresh empty temp dir for sealed). Open prepares the Session, Close cleans up (no persistent process). Errors map via transport.Classify consistent with its categories without changing internal/engine. Testing: deterministic unit tests use a fake CLI command (a stub program or an injectable command runner) to assert delta accumulation and full-render, stdin/stdout handling, model and command wiring with env override, grounded vs sealed working directory, and error mapping; a real-agy integration test is gated behind a build tag and env var that the gate compiles but skips. The production resolver registers exec for the agy backend; acp and echo stay. internal/backend/exec depends only on stdlib, internal/engine, and internal/backend. Out of scope: the api backend, any internal/engine or internal/debate change beyond import, and scaffolding.",
  "queries": [
    "internal/backend/exec",
    "cmd/debate",
    "spec.Model",
    "internal/engine",
    "stdin/stdout",
    "internal/backend",
    "internal/debate",
    "plain-CLI"
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
