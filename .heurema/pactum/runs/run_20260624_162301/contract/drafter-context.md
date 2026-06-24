# Contract Drafter Context

## Run
- Run id: run_20260624_162301
- Run status: contract_draft

## Contract goal
Replace the hand-written stdlib flag parsing in cmd/debate with the kong library (github.com/alecthomas/kong, the same CLI library pactum uses), fixing the bug where flags placed after a positional argument are ignored. Preserve all existing behavior: the bare-task run form debate "<task>", the subcommands version/init/new <name>, all run flags (--with, --synth, --max-rounds, --json, -q/--quiet, --sealed, --task @file) and the new --role flag, the task composition (positional / --task @file / stdin appended), the exit codes (0 settled, 2 not-converged, 1 error), and the stdout=answer / stderr=trace contract. The only behavioral change is that flags now parse correctly in any position (before or after the positional). Model the debate run as kong's default command so the bare task keeps working alongside the subcommands. Add kong to go.mod and update the cmd/debate tests, asserting flags work both before and after the positional. Out of scope: internal/engine, internal/debate, internal/backend, adding new subcommands (no validate), and changing the resolver/runner/synthesizer/IO contracts.

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

Generated: 2026-06-24T16:23:01Z

Map run: map_20260624_161537
Repo map path: .heurema/pactum/map/repo-map.md
LLMS path: .heurema/pactum/map/llms.txt
Search index path: .heurema/pactum/map/search.sqlite
Accepted memory context: context/memory-context.md

Notes:
- Pactum has not yet done agentic clarification.
- This is deterministic context assembled from existing map artifacts.

## Project map

# Pactum Project Map

Generated: 2026-06-24T16:15:37Z

Repository root: `.`

## Summary

- Indexed files: 45
- Ignored files/directories: 1211
- Detected languages: 5
- Code items (best-effort hints): 450

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

- Go: 33 file(s)
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
- `cmd/debate/scaffold.go`
- `cmd/debate/scaffold_test.go`
- `docs/DESIGN.md`
- `docs/SLICES.md`
- `go.mod`
- `go.sum`
- `internal/backend/acp/...`
- `internal/backend/exec/...`
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
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdInit_CreatesWorkspaceThatLoads`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdInit_DoesNotOverwriteExisting`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdInit_ExtraArgsError`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdInit_StarterPersonasParseable`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdNew_CreatesPersonaThatParses`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdNew_CreatesPersonasDirIfAbsent`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdNew_InvalidRole`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdNew_MissingName`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdNew_RefusesOverwrite`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdNew_RejectsPathSeparators`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdNew_RequiresWorkspace`
- `cmd/debate/scaffold_test.go`: `go_func` `main.TestCmdNew_SynthesizerRole`
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
- `internal/backend/exec/exec.go`: `go_func` `exec.New`
- `internal/backend/exec/exec.go`: `go_type` `exec.CommandRunner`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestBuildStdin_ContentAlreadyHasTrailingNewline`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestBuildStdin_EmptyHistory`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestBuildStdin_WithHistory`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestBuildStdin_WithSystem`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestClose_Grounded_NoTempRemoval`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestClose_Idempotent`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestClose_Sealed_RemovesTempDir`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestCmd_Default`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestCmd_ModelWired`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestCmd_Override`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestNew_InvalidBackend`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestNew_ValidBackend`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestOpen_Grounded_Cwd`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestOpen_MissingModel`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestOpen_Sealed_Cwd`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_BrokenPipe_BothFail`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_BrokenPipe_Retried`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_CancelBeforeSpawn`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_CancelDuringRun`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_FailedSendNoHistoryPollution`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_NonZeroExit_StderrInError`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_NonZeroExit_Terminal`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_PriorRepliesInHistory`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_SpawnFailure_BothFail`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_SpawnFailure_Retryable`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_StdinFormat_ContentTrailingNewline`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_StdinFormat_MultiTurn`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_StdinFormat_NoSystem`
- `internal/backend/exec/exec_test.go`: `go_func` `exec.TestSend_StdinFormat_SystemAbsent`

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
  "query": "Replace the hand-written stdlib flag parsing in cmd/debate with the kong library (github.com/alecthomas/kong, the same CLI library pactum uses), fixing the bug where flags placed after a positional argument are ignored. Preserve all existing behavior: the bare-task run form debate \"\u003ctask\u003e\", the subcommands version/init/new \u003cname\u003e, all run flags (--with, --synth, --max-rounds, --json, -q/--quiet, --sealed, --task @file) and the new --role flag, the task composition (positional / --task @file / stdin appended), the exit codes (0 settled, 2 not-converged, 1 error), and the stdout=answer / stderr=trace contract. The only behavioral change is that flags now parse correctly in any position (before or after the positional). Model the debate run as kong's default command so the bare task keeps working alongside the subcommands. Add kong to go.mod and update the cmd/debate tests, asserting flags work both before and after the positional. Out of scope: internal/engine, internal/debate, internal/backend, adding new subcommands (no validate), and changing the resolver/runner/synthesizer/IO contracts.",
  "queries": [
    "cmd/debate",
    "github.com/alecthomas/kong",
    "version/init/new",
    "-q/--quiet",
    "/",
    "go.mod",
    "internal/engine",
    "internal/debate"
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
