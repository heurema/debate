# debate design

> Status: working design. This file records current architecture decisions and is expected to evolve with the implementation.
>
> Module: `github.com/heurema/debate`

## 1. Product

`debate` is a Go CLI for structured multi-agent deliberation. The user gives it a task, a panel of persona-driven agents discusses the task in rounds, and a synthesizer turns the resulting transcript into the final answer.

The product goal is narrow: it should feel like asking a small review panel for a decision, not like operating an agent framework. The command reads a task, loads local personas, runs the panel until convergence or a limit, then writes the final answer to stdout.

The implementation contains two layers:

- `internal/engine`: reusable orchestration machinery. It knows how to run participants in rounds, collect turns, track transcript deltas, and ask a verdict policy whether to stop.
- `internal/debate` and `cmd/debate`: product policy. It knows about `.heurema/debate`, personas, synthesizers, task assembly, CLI behavior, and backend resolution.

The dependency direction is one-way:

```text
debate product -> engine
```

The engine must not import product packages or know about personas, `.heurema`, CLI flags, or synthesizers.

## 2. Repository Shape

```text
github.com/heurema/debate
├── cmd/debate/                 # CLI entrypoint
├── internal/backend/           # concrete backend adapters
│   ├── acp/
│   └── exec/
├── internal/debate/            # product policy
│   ├── config/
│   ├── persona/
│   ├── prompt/
│   ├── runner/
│   ├── signal/
│   └── verdict/
├── internal/engine/            # reusable debate engine
│   ├── loop/
│   ├── orchestrate/
│   └── transport/
└── docs/
```

`internal/engine` is intentionally kept under `internal` until a second product needs it. If that happens, it can be extracted into a public module because the dependency boundary is already shaped for extraction.

## 3. Engine

The engine is library-only. Its input is an in-memory Go config, not YAML. Its job is to run participants until a supplied verdict says the debate is done, stalled, or out of rounds.

Core responsibilities:

- Open one session per participant.
- Schedule participants round by round.
- Render prompts through a product-supplied `PromptBuilder`.
- Append turns to a shared transcript.
- Track per-participant transcript deltas.
- Ask a product-supplied `Verdict` after each round.
- Return a transcript and outcome.

The main policy seams are:

- `PromptBuilder`: product-owned prompt rendering.
- `Verdict`: product-owned convergence logic.
- `Scheduler`: participant ordering. V1 uses round-robin.

Sketch:

```go
type Limits struct {
    Max      int
    Settle   int
    Patience int
}

type RoundResult struct {
    Clean    bool
    Progress bool
}

type Outcome struct {
    Reason string // settled | stalemate | max | stop
    Rounds int
}

type Spec struct {
    ID       string
    Model    string
    Effort   string
    System   string
    ReadOnly bool
    Command  []string
}

type Transport interface {
    Open(context.Context, Spec) (Session, error)
}

type Session interface {
    Send(context.Context, string) (Result, error)
    Close() error
}
```

The engine does not know about:

- persona files
- `.heurema/debate`
- workspace tables
- backend inference
- synthesizer selection
- signal JSON shape
- CLI output policy

## 4. Transports And Backends

The engine speaks to participants through `transport.Transport` and `transport.Session`.

Current backend adapters:

| Backend id | Adapter | Purpose |
|---|---|---|
| `echo` | `internal/engine/transport/echo` | deterministic offline smoke tests and demos |
| `claude-agent-acp` | `internal/backend/acp` | ACP-backed Claude agent process |
| `codex-acp` | `internal/backend/acp` | ACP-backed Codex agent process |
| `agy` | `internal/backend/exec` | stateless CLI subprocess backend |

Backend ids are product-level names. They resolve to engine transports during runner setup.

Persona model names infer default backends unless `backend` is set explicitly:

| Model pattern | Default backend |
|---|---|
| `claude-*`, `opus`, `sonnet`, `haiku`, `fable` | `claude-agent-acp` |
| `gpt-*`, `codex`, `o*` | `codex-acp` |
| `gemini-*` | `agy` |

ACP backends are session-oriented: one long-lived session per participant for the run. Exec backends are stateless: each turn spawns a command and renders the full prompt context.

API backends are a future extension. The engine boundary already supports them, but product-grade grounding and tool loops are deferred.

## 5. Convergence

Debate convergence uses a self-signal plus the generic streak loop.

Each participant must end each turn with:

```signal
{"position": "<current position>", "objections": ["<blocking objection>"], "done": false}
```

The parser reads the last fenced `signal` block. If `done` is true while objections are still present, the parser normalizes `done` back to false.

After every round, the debate verdict checks:

- `Clean`: the selected until policy is satisfied.
- `Progress`: the open-objection set changed since the previous round.

Current until policies:

- `all_done`: every participant in the round is done.
- `quorum`: strictly more than half of the participants are done.

The loop then applies streak rules:

- enough clean rounds -> `settled`
- enough rounds without progress -> `stalemate`
- maximum rounds reached -> `max`

The CLI currently exposes `--max-rounds`; settle and patience remain code defaults.

## 6. Workspace

The project-local workspace is discovered by walking upward from the current directory, like Git discovery. The marker is `.heurema/debate`; commands do not fall back to the repository root, home directory, or an implicit current-directory layout when that marker is absent.

```text
.heurema/debate/
├── personas/
│   ├── proposer.md
│   ├── skeptic.md
│   └── reviewers/
│       └── security.md
└── tables/
    └── default.yml
```

`debate init` creates `.heurema/debate/personas/proposer.md`, `.heurema/debate/personas/skeptic.md`, and `.heurema/debate/tables/default.yml`. It writes into the current directory and does not overwrite existing files.

Persona discovery loads Markdown files from exactly these shapes:

```text
personas/<name>.md
personas/<namespace>/<name>.md
```

Persona IDs are `name` or `namespace/name`. Segments may contain only letters, digits, hyphens, and underscores. Hidden files and non-Markdown files in persona directories are ignored. Deeper Markdown files fail fast.

Table discovery loads flat YAML files from:

```text
tables/<table>.yml
```

Table names use the same path-safe segment rule and do not contain slashes. Hidden files and non-`.yml` files in the tables directory are ignored.

A table pins a panel:

```yaml
version: 1
panel:
  - proposer
  - skeptic
# synth: synthesizer
```

Table files require `version: 1` and a non-empty `panel`. The optional `synth` field uses the same selector resolver as `--synth`. Unknown fields fail fast.

Persona files are Markdown with YAML front matter:

```markdown
---
version: 1
role: debater
model: claude-haiku-4-5
effort: medium
# backend: echo
# tags: [security]
---
You are the Skeptic. Challenge weak assumptions and identify blocking objections.
```

Required fields:

- `version: 1`
- `model`
- `effort`
- non-empty body

Optional fields:

- `role`: `debater` or `synthesizer`; defaults to `debater`
- `backend`: explicit backend override
- `tags`: reserved for selection features

Unknown front matter keys fail fast.

Participant selectors are deterministic:

- selectors containing `/` are exact full persona IDs
- selectors without `/` first resolve an exact root persona ID when present
- selectors without `/` otherwise resolve by short name only when exactly one candidate exists
- zero matches and ambiguous short names fail with actionable errors

Panel resolution uses `--with` in the provided order when present. Otherwise it uses `--table <name>` or `tables/default.yml`. Naming a synthesizer-role persona or the same resolved persona more than once in a panel fails before any backend session opens.

## 7. Synthesizer

The synthesizer produces the final answer from the transcript. It does not participate in the debate panel.

Resolution order:

1. `--synth <persona>`
2. selected table `synth`
3. uniquely resolved selector `synthesizer`
4. built-in default synthesizer

The built-in default uses `claude-haiku-4-5` with low effort and a neutral synthesis prompt. A custom synthesizer persona is only needed when the project wants a different model, backend, or synthesis style. Synthesizer resolution rejects debater-role personas.

## 8. Task Input

The task is assembled from:

- positional arguments
- `--task <text>`
- `--task @path/to/file`
- `--task -`, which reads stdin and does not add a literal `-`
- piped stdin when `--task -` has not already consumed stdin

Sources compose. For example, a user can pipe a diff and add an instruction:

```sh
git diff | debate "Find blocking risks in this change"
```

There is no separate `context.md` contract. Context belongs in the task, or in files that grounded agents can read.

## 9. Agent Access

The intended execution mode is grounded read-only access:

- Agents may inspect the project directory.
- Agents may use read-only commands and web access where the backend supports it.
- Agents must not mutate the project filesystem.

`--sealed` threads read-only intent into transport specs for runs that should rely only on the brief and not on project exploration. Backend support may vary.

Known risk: read access plus network access can expose secrets if a repository contains them. Run grounded agents only in repositories that are appropriate for model inspection.

## 10. CLI

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
--with <persona>      Add a debater persona. Repeat for multiple participants.
--synth <persona>     Use a synthesizer persona for the final answer.
--task <value>        Read task from inline text, @file, or - for stdin.
--max-rounds <n>      Limit debate rounds. Defaults to 10.
--json                Emit JSON.
-q, --quiet           Reduce human-readable output.
--sealed              Thread read-only intent into backend transports where supported.
```

IO contract:

- stdout: final answer, or structured JSON with `--json`
- stderr: live debate trace when enabled
- exit `0`: settled
- exit `2`: not converged
- exit `1`: error

The CLI validates workspace shape and persona config before the first model call where possible.

## 11. Fixed Decisions

- Product name and binary: `debate`.
- Module path: `github.com/heurema/debate`.
- Engine is embedded under `internal/engine` until a second product needs extraction.
- Product depends on engine; engine never depends on product.
- V1 scheduler: round-robin.
- V1 convergence: self-signal plus streak.
- Persona format is backend-invariant.
- Backend inference comes from model name, with `backend` as an escape hatch.
- Synthesizer is product policy, not engine policy.
- `context.md` is not part of the workspace contract.

## 12. Future Work

- Public engine module after another product validates the boundary.
- API backend with grounding tool loop.
- Allow and deny lists for project paths.
- Network allowlist for grounded mode.
- Persona selection by tags.
- Pipeline-style orchestration over multiple debate runs.
