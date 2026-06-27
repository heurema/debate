```text
     _      _           _
  __| | ___| |__   __ _| |_ ___
 / _` |/ _ \ '_ \ / _` | __/ _ \
| (_| |  __/ |_) | (_| | ||  __/
 \__,_|\___|_.__/ \__,_|\__\___|

        debate agents, not opinions
```

# debate

`debate` is a small contract-first CLI for running structured multi-agent debates over a task, design choice, review target, or implementation plan.

It gives each participant a persona, collects their arguments round by round, and can hand the transcript to a synthesizer for a final decision. The repository is intentionally compact: the core is a Go CLI, while `.heurema/` holds local personas, prompts, and project-local state.

## What It Does

- Runs a debate from a positional prompt, `--task @file`, `--task -`, or piped stdin.
- Loads Markdown personas from `.heurema/debate/personas` with optional one-level namespaces.
- Selects panels from a default table, a named table, or an ad-hoc `--with` list.
- Supports proposer/skeptic style review loops and optional synthesis.
- Can emit human-readable output or JSON for automation.
- Keeps backend selection in persona files instead of hard-coding one model.

## Quick Start

Build the CLI:

```sh
go build -o debate ./cmd/debate
```

Check the binary:

```sh
./debate version
```

Create the local debate workspace:

```sh
./debate init
```

Run a first debate:

```sh
./debate "Should this feature be implemented now or deferred?"
```

Ask for machine-readable output:

```sh
./debate "Review the current implementation plan" --json
```

## Workspace Layout

After `debate init`, the project-local state lives under `.heurema/debate`:

```text
.heurema/
  debate/
    personas/
      proposer.md
      skeptic.md
      reviewers/
        security.md
    tables/
      default.yml
      architecture.yml
```

Commands that read an existing workspace discover the nearest ancestor containing `.heurema/debate`, starting with the current directory. `debate init` is the exception: it creates the scaffold in the current directory.

Each persona file is a Markdown document with YAML front matter:

```markdown
---
version: 1
role: debater
model: claude-haiku-4-5
effort: medium
# backend: echo
# tags: [architecture]
---
You are the Skeptic. Challenge weak assumptions and identify blocking objections.
```

Required fields are `version: 1`, `model`, `effort`, and a non-empty body. `role` defaults to `debater` and may be `debater` or `synthesizer`. `backend` overrides model-based backend inference, and `tags` are preserved for future selection features. Unknown fields fail fast.

Persona files may live at `personas/<name>.md` or `personas/<namespace>/<name>.md`. The qualified persona ID is the relative path without `.md`, such as `skeptic` or `reviewers/security`. Hidden files and non-Markdown files are ignored. Markdown files nested deeper than one namespace level fail fast.

Tables are explicit and flat under `.heurema/debate/tables`:

```yaml
version: 1
panel:
  - proposer
  - skeptic
# synth: final-judge
```

A run without `--with` uses `--table <name>` or `tables/default.yml`. Table files require `version: 1` and a non-empty `panel`. The optional `synth` field chooses the synthesizer for that table. Tables are flat YAML files, table names do not contain `/`, and panel order is preserved.

## Personas

Create another debater:

```sh
./debate new architect
```

Create a namespaced persona:

```sh
./debate new reviewers/security
```

Create a synthesizer persona:

```sh
./debate new final-judge --role synthesizer
```

Then select explicit participants. Repeatable `--with` values and comma-separated selectors are equivalent conveniences for providing the ordered explicit panel:

```sh
./debate "Pick the safest migration path" --with proposer --with skeptic --synth final-judge
./debate "Pick the safest migration path" --with proposer,skeptic --synth final-judge
```

Selectors are deterministic. `namespace/name` is an exact persona ID. A short selector such as `skeptic` first resolves an exact root persona ID when present; otherwise it works only when exactly one loaded persona has that short name. Ambiguous short names fail and list the qualified IDs.

Empty comma entries are rejected, so `--with proposer,` and `--with proposer,,skeptic` fail before any backend session opens.

Selection precedence is:

- panel: `--with` in the provided order, otherwise the selected table panel
- synthesizer: `--synth`, then selected table `synth`, then a uniquely resolved `synthesizer` persona, then the built-in default

For deterministic local smoke tests, point a persona at the echo backend:

```yaml
backend: echo
model: echo-local
```

Real model execution depends on your local adapters and credentials. The current backend registry includes echo, ACP-backed Claude/Codex runners, and an exec runner for `agy`.

## CLI Reference

```text
debate "<task>" [flags]
debate --task @path/to/task.md [flags]
debate --task - [flags]
debate init
debate new [--role debater|synthesizer] <name|namespace/name>
debate version
```

Run flags:

```text
--table <name>        Select a flat table from .heurema/debate/tables.
--with <persona>      Add debater persona selectors. Repeat or separate selectors with commas.
--synth <persona>     Use a synthesizer persona for the final answer.
--task <value>        Read task from inline text, @file, or - for stdin.
--max-rounds <n>      Limit debate rounds. Defaults to 10.
--json                Emit JSON.
-q, --quiet           Reduce human-readable output.
--sealed              Thread read-only intent into backend transports where supported.
```

## Development

Run the test suite:

```sh
go test ./...
```

Build the binary:

```sh
go build -o debate ./cmd/debate
```

## Status

This is an active personal project. The implemented surface is intentionally narrow: initialize personas, create personas, parse debate runs, execute configured agents, and serialize results.
