# Contract Drafter Context

## Run
- Run id: run_20260623_213058
- Run status: contract_draft

## Contract goal
Slice 0: bootstrap the debate Go project skeleton — go.mod (module github.com/heurema/debate), package layout (internal/engine/{loop,transport,orchestrate}, internal/debate, cmd/debate), Makefile, and a 'debate version' command. No engine or debate logic yet.

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

Generated: 2026-06-23T21:30:58Z

Map run: map_20260623_212315
Repo map path: .heurema/pactum/map/repo-map.md
LLMS path: .heurema/pactum/map/llms.txt
Search index path: .heurema/pactum/map/search.sqlite
Accepted memory context: context/memory-context.md

Notes:
- Pactum has not yet done agentic clarification.
- This is deterministic context assembled from existing map artifacts.

## Project map

# Pactum Project Map

Generated: 2026-06-23T21:23:15Z

Repository root: `.`

## Summary

- Indexed files: 2
- Ignored files/directories: 9
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
  - `wiki/areas/docs.md`

## Project map artifacts

- `files.jsonl` — deterministic per-file metadata.
- `hashes.jsonl` — per-file content hashes.
- `code-items.jsonl` — best-effort symbol hints (incomplete by design).
- `search.sqlite` — local full-text search index.
- `manifest.json` — map manifest listing all artifacts.

## Files / areas

### Detected languages

- Markdown: 2 file(s)

### Top-level directories

- `docs/` (see `wiki/areas/docs.md`)

### Important files

- None detected

### File tree

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
  "query": "Slice 0: bootstrap the debate Go project skeleton — go.mod (module github.com/heurema/debate), package layout (internal/engine/{loop,transport,orchestrate}, internal/debate, cmd/debate), Makefile, and a 'debate version' command. No engine or debate logic yet.",
  "queries": [
    "go.mod",
    "github.com/heurema/debate",
    "internal/engine/{loop,transport,orchestrate",
    "internal/debate",
    "cmd/debate",
    "slice",
    "bootstrap",
    "project"
  ],
  "query_source": "task",
  "results": [
    {
      "rank": 1,
      "id": "wiki:commands.md",
      "kind": "wiki",
      "path": "map/wiki/commands.md",
      "title": "Commands",
      "language": "",
      "code_kind": "",
      "score": -3.918286203385769,
      "snippet": "...No package.json scripts detected.\n\n## Go commands\n\n- No go.mod detected.\n\n## Python commands\n\n- No Python project...",
      "source_query": "go.mod"
    },
    {
      "rank": 2,
      "id": "wiki:structure.md",
      "kind": "wiki",
      "path": "map/wiki/structure.md",
      "title": "Project structure",
      "language": "",
      "code_kind": "",
      "score": -0.2520706352942152,
      "snippet": "# Project structure\n\nGenerated: 2026-06-23T21:23:15Z\n\nThis page is part of the deterministic map...",
      "source_query": "project"
    },
    {
      "rank": 3,
      "id": "repo-map.md",
      "kind": "repo_map",
      "path": "map/repo-map.md",
      "title": "Repository map",
      "language": "",
      "code_kind": "",
      "score": -0.20111578541040045,
      "snippet": "# Pactum Project Map\n\nGenerated: 2026-06-23T21:23:15Z\n\nRepository root: `.`\n\n## Summary\n\n- Indexed files: 2\n- Ignored...",
      "source_query": "project"
    },
    {
      "rank": 4,
      "id": "llms.txt",
      "kind": "llms",
      "path": "map/llms.txt",
      "title": "LLM map pointer",
      "language": "",
      "code_kind": "",
      "score": -0.1968289343717683,
      "snippet": "# Pactum project map — agent router\n\nThis is a generated, deterministic Pactum project map. Read the map...",
      "source_query": "project"
    },
    {
      "rank": 5,
      "id": "wiki:overview.md",
      "kind": "wiki",
      "path": "map/wiki/overview.md",
      "title": "Project map overview",
      "language": "",
      "code_kind": "",
      "score": -0.19097847966065565,
      "snippet": "# Project map overview\n\nGenerated: 2026-06-23T21:23:15Z\n\nThis page is part of the deterministic...",
      "source_query": "project"
    }
  ]
}

## Drafter guidance
- Propose only additions to the contract fields listed in the prompt.
- Do not change or restate the contract goal.
- Do not answer clarification questions.
- Do not edit files.
- Treat repository map/search context as navigation hints, not semantic truth.
