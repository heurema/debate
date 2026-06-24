# Contract Drafter Context

## Run
- Run id: run_20260624_103219
- Run status: contract_draft

## Contract goal
Slice 5: implement the ACP backend transport in internal/engine/transport/acp and wire it into the cmd/debate production resolver, so debate can run a real debate with Claude (backend claude-agent-acp) and Codex (backend codex-acp). The acp transport implements transport.Transport over the Agent Client Protocol: Open spawns the backend's ACP CLI subprocess, performs the ACP initialize and session/new handshake, and returns a Session holding the persistent subprocess and session id. Session.Send sends the prompt as a session/prompt request and returns the agent's final text as transport.Result; the session is persistent and reused across rounds. Grounding: the session runs with cwd as a read-only sandbox plus network (agent reads project files and the internet but must not mutate the filesystem); transport.Spec.ReadOnly (from --sealed) tightens to brief-only. Session.Close shuts down the subprocess. Errors map via transport.Classify and a dropped session recovers by reopening and replaying. Testing: a fake in-process ACP peer (speaking the protocol over pipes / an injected stream) drives deterministic unit tests of handshake, send, persistence, error classification, and recovery; a real-CLI integration test is gated behind a build tag and env var so the default test run needs no network or API key. The cmd/debate production resolver registers the acp transport for claude-agent-acp and codex-acp; the echo backend stays for offline use. Out of scope: the exec/agy backend, the api backend, and any change to engine/policy/config packages beyond import.

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

Generated: 2026-06-24T10:32:19Z

Map run: map_20260624_100801
Repo map path: .heurema/pactum/map/repo-map.md
LLMS path: .heurema/pactum/map/llms.txt
Search index path: .heurema/pactum/map/search.sqlite
Accepted memory context: context/memory-context.md

Notes:
- Pactum has not yet done agentic clarification.
- This is deterministic context assembled from existing map artifacts.

## Project map

# Pactum Project Map

Generated: 2026-06-24T10:08:01Z

Repository root: `.`

## Summary

- Indexed files: 33
- Ignored files/directories: 574
- Detected languages: 5
- Code items (best-effort hints): 244

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

- Go: 21 file(s)
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
- `internal/debate/signal/...`
- `internal/debate/verdict/...`
- `internal/engine/loop/...`
- `internal/engine/orchestrate/...`
- `internal/engine/transport/...`
- `scripts/check-gofmt.sh`
- `scripts/dep-guard.sh`

## Code surface (best-effort code hints)

- `cmd/debate/main.go`: `go_main` `main`
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
- `internal/debate/verdict/verdict_test.go`: `go_func` `verdict_test.TestVerdictSettledAllDone`
- `internal/debate/verdict/verdict_test.go`: `go_func` `verdict_test.TestVerdictSettledQuorum`
- `internal/debate/verdict/verdict_test.go`: `go_func` `verdict_test.TestVerdictStalemate`
- `internal/debate/verdict/verdict_test.go`: `go_func` `verdict_test.TestVerdictUnparsedSignal`
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
  "query": "Slice 5: implement the ACP backend transport in internal/engine/transport/acp and wire it into the cmd/debate production resolver, so debate can run a real debate with Claude (backend claude-agent-acp) and Codex (backend codex-acp). The acp transport implements transport.Transport over the Agent Client Protocol: Open spawns the backend's ACP CLI subprocess, performs the ACP initialize and session/new handshake, and returns a Session holding the persistent subprocess and session id. Session.Send sends the prompt as a session/prompt request and returns the agent's final text as transport.Result; the session is persistent and reused across rounds. Grounding: the session runs with cwd as a read-only sandbox plus network (agent reads project files and the internet but must not mutate the filesystem); transport.Spec.ReadOnly (from --sealed) tightens to brief-only. Session.Close shuts down the subprocess. Errors map via transport.Classify and a dropped session recovers by reopening and replaying. Testing: a fake in-process ACP peer (speaking the protocol over pipes / an injected stream) drives deterministic unit tests of handshake, send, persistence, error classification, and recovery; a real-CLI integration test is gated behind a build tag and env var so the default test run needs no network or API key. The cmd/debate production resolver registers the acp transport for claude-agent-acp and codex-acp; the echo backend stays for offline use. Out of scope: the exec/agy backend, the api backend, and any change to engine/policy/config packages beyond import.",
  "queries": [
    "internal/engine/transport/acp",
    "cmd/debate",
    "session/new",
    "Session.Send",
    "session/prompt",
    "Session.Close",
    "/",
    "exec/agy"
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
