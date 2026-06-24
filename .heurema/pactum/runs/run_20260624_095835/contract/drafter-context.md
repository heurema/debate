# Contract Drafter Context

## Run
- Run id: run_20260624_095835
- Run status: contract_draft

## Contract goal
Slice 4: wire the cmd/debate CLI into a working debate on a deterministic offline backend (no real models yet). The command debate "<task>" plus version loads the .heurema/debate workspace via config.Load, assembles the brief (workspace context followed by the task; task from positional arg, --task @file, or stdin), builds an orchestrate.Config (participants from the panel personas via a backend resolver, a RoundRobin scheduler, prompt.NewPromptBuilder, verdict.New), runs orchestrate.Run, then runs the synthesizer once to produce the final answer. Flags: --with, --synth, --max-rounds, --json, -q, --sealed. Output contract: stdout is the answer, stderr is the live debate trace (auto-quiet off-TTY or with -q), exit 0 settled, 2 not-converged (stalemate or max), 1 error. A backend registry resolves persona.Backend to a transport; register a deterministic offline echo backend (canned reply with a valid signal block, no network) and accept an injectable resolver so tests use the engine mock backend; real acp/exec/api backends are out of scope. Fail-fast validation before opening any session. e2e tests over a fixture .heurema/debate workspace. cmd/debate uses the stdlib flag package (no third-party CLI lib), internal/debate, and internal/engine. Out of scope: real backends, debate init/new scaffolding, and the real grounded sandbox behind --sealed.

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

Generated: 2026-06-24T09:58:35Z

Map run: map_20260624_093601
Repo map path: .heurema/pactum/map/repo-map.md
LLMS path: .heurema/pactum/map/llms.txt
Search index path: .heurema/pactum/map/search.sqlite
Accepted memory context: context/memory-context.md

Notes:
- Pactum has not yet done agentic clarification.
- This is deterministic context assembled from existing map artifacts.

## Project map

# Pactum Project Map

Generated: 2026-06-24T09:36:01Z

Repository root: `.`

## Summary

- Indexed files: 27
- Ignored files/directories: 475
- Detected languages: 5
- Code items (best-effort hints): 177

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

- Go: 17 file(s)
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
- `internal/debate/prompt/...`
- `internal/debate/signal/...`
- `internal/debate/verdict/...`
- `internal/engine/loop/...`
- `internal/engine/orchestrate/...`
- `internal/engine/transport/...`
- `scripts/dep-guard.sh`

## Code surface (best-effort code hints)

- `cmd/debate/main.go`: `go_main` `main`
- `cmd/debate/main.go`: `go_main` `main.main`
- `cmd/debate/main_test.go`: `go_func` `main.TestVersion`
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
  "query": "Slice 4: wire the cmd/debate CLI into a working debate on a deterministic offline backend (no real models yet). The command debate \"\u003ctask\u003e\" plus version loads the .heurema/debate workspace via config.Load, assembles the brief (workspace context followed by the task; task from positional arg, --task @file, or stdin), builds an orchestrate.Config (participants from the panel personas via a backend resolver, a RoundRobin scheduler, prompt.NewPromptBuilder, verdict.New), runs orchestrate.Run, then runs the synthesizer once to produce the final answer. Flags: --with, --synth, --max-rounds, --json, -q, --sealed. Output contract: stdout is the answer, stderr is the live debate trace (auto-quiet off-TTY or with -q), exit 0 settled, 2 not-converged (stalemate or max), 1 error. A backend registry resolves persona.Backend to a transport; register a deterministic offline echo backend (canned reply with a valid signal block, no network) and accept an injectable resolver so tests use the engine mock backend; real acp/exec/api backends are out of scope. Fail-fast validation before opening any session. e2e tests over a fixture .heurema/debate workspace. cmd/debate uses the stdlib flag package (no third-party CLI lib), internal/debate, and internal/engine. Out of scope: real backends, debate init/new scaffolding, and the real grounded sandbox behind --sealed.",
  "queries": [
    "cmd/debate",
    "heurema/debate",
    "config.Load",
    "verdict.New",
    "orchestrate.Run",
    "acp/exec/api",
    "internal/debate",
    "internal/engine"
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
