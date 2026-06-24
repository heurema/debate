# Contract Drafter Context

## Run
- Run id: run_20260624_085539
- Run status: contract_draft

## Contract goal
Slice 2: implement the debate policy layer in internal/debate on top of the engine, exercised only with the mock backend. (1) internal/debate/signal: parse a structured signal {position string, objections []string, done bool} from a turn text — the speaker ends its reply with a fenced signal block (triple-backtick signal ... containing JSON); the parser extracts and validates it, returns a typed result plus a parsed-ok flag, and applies the invariant that done==true with non-empty objections is treated as done=false. (2) internal/debate/prompt: a PromptBuilder matching orchestrate.PromptBuilder that renders a per-turn user message = brief (task+context) + moderator rules-of-engagement + the delta board (rendered from Transcript.DeltaFor for the speaking participant) + the signal-format instruction; support RenderMode Delta and Full. (3) internal/debate/verdict: a Verdict matching orchestrate.Verdict that parses each round turns signals and returns loop.RoundResult where Clean = all speakers done (until=all_done) or a majority done (until=quorum), and Progress = the open-objection set changed vs previous round; configurable until in {all_done, quorum}; an unparsed signal makes that speaker not-done. Unit tests use only the mock backend with scripted signal-bearing turns: convergence after the settle streak, quorum, stalemate on a frozen objection set, max rounds, and unparsed-signal handling. internal/debate imports internal/engine only (one-way). Out of scope: real acp/exec/api backends, CLI, persona files, .heurema/debate discovery/config, synthesizer, and nudge-retry orchestration (parser only).

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

Generated: 2026-06-24T08:55:39Z

Map run: map_20260624_081848
Repo map path: .heurema/pactum/map/repo-map.md
LLMS path: .heurema/pactum/map/llms.txt
Search index path: .heurema/pactum/map/search.sqlite
Accepted memory context: context/memory-context.md

Notes:
- Pactum has not yet done agentic clarification.
- This is deterministic context assembled from existing map artifacts.

## Project map

# Pactum Project Map

Generated: 2026-06-24T08:18:48Z

Repository root: `.`

## Summary

- Indexed files: 15
- Ignored files/directories: 236
- Detected languages: 4
- Code items (best-effort hints): 11

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

## Project map artifacts

- `files.jsonl` — deterministic per-file metadata.
- `hashes.jsonl` — per-file content hashes.
- `code-items.jsonl` — best-effort symbol hints (incomplete by design).
- `search.sqlite` — local full-text search index.
- `manifest.json` — map manifest listing all artifacts.

## Files / areas

### Detected languages

- Go: 6 file(s)
- Markdown: 6 file(s)
- Go module: 1 file(s)
- Make: 1 file(s)

### Top-level directories

- `.claude/` (see `wiki/areas/.claude.md`)
- `cmd/` (see `wiki/areas/cmd.md`)
- `docs/` (see `wiki/areas/docs.md`)
- `internal/` (see `wiki/areas/internal.md`)

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

## Code surface (best-effort code hints)

- `cmd/debate/main.go`: `go_main` `main`
- `cmd/debate/main.go`: `go_main` `main.main`
- `cmd/debate/main_test.go`: `go_func` `main.TestVersion`

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
  "query": "Slice 2: implement the debate policy layer in internal/debate on top of the engine, exercised only with the mock backend. (1) internal/debate/signal: parse a structured signal {position string, objections []string, done bool} from a turn text — the speaker ends its reply with a fenced signal block (triple-backtick signal ... containing JSON); the parser extracts and validates it, returns a typed result plus a parsed-ok flag, and applies the invariant that done==true with non-empty objections is treated as done=false. (2) internal/debate/prompt: a PromptBuilder matching orchestrate.PromptBuilder that renders a per-turn user message = brief (task+context) + moderator rules-of-engagement + the delta board (rendered from Transcript.DeltaFor for the speaking participant) + the signal-format instruction; support RenderMode Delta and Full. (3) internal/debate/verdict: a Verdict matching orchestrate.Verdict that parses each round turns signals and returns loop.RoundResult where Clean = all speakers done (until=all_done) or a majority done (until=quorum), and Progress = the open-objection set changed vs previous round; configurable until in {all_done, quorum}; an unparsed signal makes that speaker not-done. Unit tests use only the mock backend with scripted signal-bearing turns: convergence after the settle streak, quorum, stalemate on a frozen objection set, max rounds, and unparsed-signal handling. internal/debate imports internal/engine only (one-way). Out of scope: real acp/exec/api backends, CLI, persona files, .heurema/debate discovery/config, synthesizer, and nudge-retry orchestration (parser only).",
  "queries": [
    "internal/debate",
    "internal/debate/signal",
    "internal/debate/prompt",
    "internal/debate/verdict",
    "internal/engine",
    "acp/exec/api",
    "heurema/debate",
    "discovery/config"
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
