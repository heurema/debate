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

- Runs a debate from a positional prompt, `--task @file`, or stdin.
- Loads debater personas from `.heurema/debate/personas`.
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
```

Each persona file is a Markdown document with front matter. The front matter controls the agent name, role, backend, model, effort, timeout, and other runtime hints. The body describes how that participant should argue.

## Personas

Create another debater:

```sh
./debate new architect
```

Create a synthesizer persona:

```sh
./debate new final-judge --role synthesizer
```

Then select explicit participants:

```sh
./debate "Pick the safest migration path" --with proposer --with skeptic --synth final-judge
```

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
debate new [--role debater|synthesizer] <name>
debate version
```

Run flags:

```text
--with <persona>      Add a debater persona. Repeat for multiple participants.
--synth <persona>     Use a synthesizer persona for the final answer.
--task <value>        Read task from inline text, @file, or - for stdin.
--max-rounds <n>      Limit debate rounds. Defaults to 10.
--json                Emit JSON.
-q, --quiet           Reduce human-readable output.
--sealed              Hide debater output from other debaters where supported.
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
