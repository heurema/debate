# Contract Drafter Context

## Run
- Run id: run_20260624_132440
- Run status: contract_draft

## Contract goal
Slice 7: add debate init and debate new scaffolding subcommands to cmd/debate, using the standard library only, without changing internal/engine, internal/debate, or internal/backend. debate init scaffolds a .heurema/debate workspace under the current directory with two starter debater personas (proposer, skeptic) and a context.md template, never overwriting existing files; the scaffolded workspace loads via config.Load with the two-debater panel and the built-in default synthesizer. debate new <name> creates a new persona file under .heurema/debate/personas from a template with frontmatter (role defaulting to debater, overridable via --role debater|synthesizer; concrete model and effort defaults; optional backend) and a placeholder body, refusing to overwrite an existing persona. Deterministic tests use temporary directories to assert init creates a config.Load-able workspace, re-running init does not overwrite, new creates a persona.ParseFile-able file, and new refuses to overwrite. Out of scope: backends, the run and synthesizer path, and any internal/engine, internal/debate, or internal/backend source change.

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

Generated: 2026-06-24T13:24:40Z

Map run: map_20260624_125018
Repo map path: .heurema/pactum/map/repo-map.md
LLMS path: .heurema/pactum/map/llms.txt
Search index path: .heurema/pactum/map/search.sqlite
Accepted memory context: context/memory-context.md

Notes:
- Pactum has not yet done agentic clarification.
- This is deterministic context assembled from existing map artifacts.

## Project map

# Pactum Project Map

Generated: 2026-06-24T12:50:18Z

Repository root: `.`

## Summary

- Indexed files: 40
- Ignored files/directories: 915
- Detected languages: 5
- Code items (best-effort hints): 357

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

- Go: 28 file(s)
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
- `internal/backend/acp/...`
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
- `cmd/debate/e2e_test.go`: `go_func` `main.TestDefaultResolver_ACPBackendsResolve`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_EmptyTask_Exit1`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_JSONOutput`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_NonTTY_NoTrace`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_NotConverged_Exit2`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_SettledRun_WithTrace`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_TaskFromFile`
- `cmd/debate/e2e_test.go`: `go_func` `main.TestE2E_UnimplementedBackend_Exit1`
- `cmd/debate/main.go`: `go_main` `main.main`
- `cmd/debate/main_test.go`: `go_func` `main.TestVersion`
- `internal/backend/acp/acp.go`: `go_func` `acp.New`
- `internal/backend/acp/acp.go`: `go_type` `acp.ProcessRunner`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestBuildCmd_ClaudeDefault`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestBuildCmd_ClaudeOverride`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestBuildCmd_CodexAlwaysSandboxReadOnly`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestBuildCmd_CodexDefault`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestBuildCmd_CodexIgnoresEffort`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestBuildCmd_CodexOverride`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestClose_Idempotent`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestGrounded_Cwd`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestNew_UnknownBackend`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestNew_ValidBackends`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestOpen_Handshake`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestOpen_MissingModel`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestSealed_Cwd`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestSend_EndTurn`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestSend_MultipleTurns`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestSend_NonEndTurnStop`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestSend_RecoveryReplayFailure`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestSend_RecoveryWithHistoryReplay`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestSend_Refusal`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestSend_RetryableDropRecovery`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestSend_SecondDropIsTerminal`
- `internal/backend/acp/acp_test.go`: `go_func` `acp.TestSend_SystemPromptInjected`
- `internal/backend/acp/integration_test.go`: `go_func` `acp.TestIntegration_ClaudeAgentACP`
- `internal/backend/acp/integration_test.go`: `go_func` `acp.TestIntegration_CodexACP`
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
  "query": "Slice 7: add debate init and debate new scaffolding subcommands to cmd/debate, using the standard library only, without changing internal/engine, internal/debate, or internal/backend. debate init scaffolds a .heurema/debate workspace under the current directory with two starter debater personas (proposer, skeptic) and a context.md template, never overwriting existing files; the scaffolded workspace loads via config.Load with the two-debater panel and the built-in default synthesizer. debate new \u003cname\u003e creates a new persona file under .heurema/debate/personas from a template with frontmatter (role defaulting to debater, overridable via --role debater|synthesizer; concrete model and effort defaults; optional backend) and a placeholder body, refusing to overwrite an existing persona. Deterministic tests use temporary directories to assert init creates a config.Load-able workspace, re-running init does not overwrite, new creates a persona.ParseFile-able file, and new refuses to overwrite. Out of scope: backends, the run and synthesizer path, and any internal/engine, internal/debate, or internal/backend source change.",
  "queries": [
    "cmd/debate",
    "internal/engine",
    "internal/debate",
    "internal/backend",
    "heurema/debate",
    "context.md",
    "config.Load",
    "heurema/debate/personas"
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
