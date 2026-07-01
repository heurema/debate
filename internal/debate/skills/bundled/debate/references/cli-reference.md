# CLI reference

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
--json                Emit JSON final result on stdout.
-q, --quiet           Suppress stderr progress events.
--sealed              Thread read-only intent into backend transports where supported.
```

The task is assembled from `--task <text>`, `--task @path/to/file`, or `--task -`
(reads stdin, no literal `-`), then positional arguments, then piped stdin when
`--task -` hasn't already consumed it. Sources are joined with newlines — for
example, piping a diff and adding an instruction:

```sh
git diff | debate "Find blocking risks in this change"
```

`debate init` scaffolds `.heurema/debate/personas/proposer.md`, `.heurema/debate/personas/skeptic.md`,
and `.heurema/debate/tables/default.yml` in the current directory, and installs or repairs
this skill globally for detected local agent clients. It never overwrites existing
workspace files.

`debate new [--role debater|synthesizer] <name|namespace/name>` creates a persona file
under the discovered workspace's `personas/` directory and fails if the persona already
exists.

`debate version` prints the binary version.

Exit codes: `0` settled, `2` did not converge (stalemate or max rounds), `1` error.
