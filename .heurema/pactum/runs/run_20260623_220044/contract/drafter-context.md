# Contract Drafter Context

## Run
- Run id: run_20260623_220044
- Run status: contract_draft

## Contract goal
Slice 1: implement the policy-free engine on a mock backend. Package internal/engine/loop: a streak loop Run(ctx, Limits{Max,Settle,Patience}, Step) -> Outcome that drives rounds and decides settled/stalemate/max via consecutive clean/no-progress streaks. Package internal/engine/transport: Transport/Session/Spec/Result interfaces (Open->Send->Close) plus error classification, and a mock backend whose Session returns pre-scripted responses for tests. Package internal/engine/orchestrate: Participant, Turn, Transcript (with DeltaFor), a RoundRobin Scheduler, pluggable PromptBuilder and Verdict seams, a Config, and Run that wires loop+transport+transcript into round-robin rounds. Provide unit tests that drive a multi-participant debate on the mock backend with a trivial verdict, asserting turn order, transcript accumulation, delta visibility, and settled/stalemate/max outcomes. Out of scope: debate policy / signal schema, real acp/exec/api backends, CLI, personas, config discovery, synthesizer.

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

Generated: 2026-06-23T22:00:44Z

Map run: map_20260623_215017
Repo map path: .heurema/pactum/map/repo-map.md
LLMS path: .heurema/pactum/map/llms.txt
Search index path: .heurema/pactum/map/search.sqlite
Accepted memory context: context/memory-context.md

Notes:
- Pactum has not yet done agentic clarification.
- This is deterministic context assembled from existing map artifacts.

## Project map

# Pactum Project Map

Generated: 2026-06-23T21:50:17Z

Repository root: `.`

## Summary

- Indexed files: 6
- Ignored files/directories: 43
- Detected languages: 1
- Code items (best-effort hints): 0

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
  - `wiki/areas/docs.md`

## Project map artifacts

- `files.jsonl` — deterministic per-file metadata.
- `hashes.jsonl` — per-file content hashes.
- `code-items.jsonl` — best-effort symbol hints (incomplete by design).
- `search.sqlite` — local full-text search index.
- `manifest.json` — map manifest listing all artifacts.

## Files / areas

### Detected languages

- Markdown: 6 file(s)

### Top-level directories

- `.claude/` (see `wiki/areas/.claude.md`)
- `docs/` (see `wiki/areas/docs.md`)

### Important files

- None detected

### File tree

- `.claude/skills/pactum/...`
- `docs/DESIGN.md`
- `docs/SLICES.md`

## Code surface (best-effort code hints)

- None detected

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
  "query": "Slice 1: implement the policy-free engine on a mock backend. Package internal/engine/loop: a streak loop Run(ctx, Limits{Max,Settle,Patience}, Step) -\u003e Outcome that drives rounds and decides settled/stalemate/max via consecutive clean/no-progress streaks. Package internal/engine/transport: Transport/Session/Spec/Result interfaces (Open-\u003eSend-\u003eClose) plus error classification, and a mock backend whose Session returns pre-scripted responses for tests. Package internal/engine/orchestrate: Participant, Turn, Transcript (with DeltaFor), a RoundRobin Scheduler, pluggable PromptBuilder and Verdict seams, a Config, and Run that wires loop+transport+transcript into round-robin rounds. Provide unit tests that drive a multi-participant debate on the mock backend with a trivial verdict, asserting turn order, transcript accumulation, delta visibility, and settled/stalemate/max outcomes. Out of scope: debate policy / signal schema, real acp/exec/api backends, CLI, personas, config discovery, synthesizer.",
  "queries": [
    "internal/engine/loop",
    "settled/stalemate/max",
    "clean/no-progress",
    "internal/engine/transport",
    "Transport/Session/Spec/Result",
    "internal/engine/orchestrate",
    "/",
    "acp/exec/api"
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
