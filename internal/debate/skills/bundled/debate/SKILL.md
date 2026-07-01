---
name: debate
description: Use when running the debate CLI, running or maintaining a .heurema/debate workspace, initializing a new workspace with debate init, choosing a debate panel or table, creating personas with debate new, or reading the @@DEBATE_PROGRESS stderr stream emitted during a debate run.
---

# debate

`debate` runs a structured multi-agent debate over a task: a panel of persona-driven
debaters argues in rounds, and a synthesizer turns the resulting transcript into a
final answer.

## Quick start

```sh
debate init                          # scaffold .heurema/debate in the current directory
debate "Should we ship this now?"     # run a debate; prints the final answer to stdout
debate "..." --json                   # machine-readable result on stdout
debate new architect                  # add a debater persona
debate new final-judge --role synthesizer
```

## Where workspace state lives

- `.heurema/debate/personas/*.md` — persona files (YAML front matter + system prompt body)
- `.heurema/debate/tables/*.yml` — named panels (`version`, `panel`, optional `synth`)
- Workspace discovery walks upward from the current directory, like `git`. `debate init`
  is the exception: it scaffolds into the current directory and never overwrites
  existing personas or tables.

`debate init` also installs or repairs a global copy of this skill for detected local
agent clients (`~/.agents/skills/debate`, and `~/.claude/skills/debate` for Claude Code
compatibility). It preserves any local edits to an installed copy — re-run `debate init`
to repair or update an unmodified one.

See `references/cli-reference.md` for the full command and flag set, `references/workspace-format.md`
for persona and table file formats, `references/progress-stream.md` for the
`@@DEBATE_PROGRESS` stderr protocol, and `references/panel-guidance.md` for choosing or
designing a panel.

## Key behavior to know before running a debate

- stdout is final-result-only: human mode prints only the final answer, `--json` prints
  only the result JSON object. Progress is on stderr as `@@DEBATE_PROGRESS <json>` lines;
  pass `--quiet` to suppress it.
- Exit code `0` means the debate settled, `2` means it did not converge, `1` is an error.
- `--sealed` threads read-only intent into backend transports where supported; it does not
  guarantee enforcement by every backend.
