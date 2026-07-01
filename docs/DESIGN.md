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

`debate init` creates `.heurema/debate/personas/proposer.md`, `.heurema/debate/personas/skeptic.md`, and `.heurema/debate/tables/default.yml`. It writes into the current directory and does not overwrite existing files. It creates no `context.md`, no synthesizer persona, and no project-local Agent Skill directory — project-local debate state stays scoped to `.heurema/debate`.

Starter persona `model`/`backend` defaults are capability-aware, using the same detection precedence as the built-in synthesizer fallback (§7): an executable named `claude` on `PATH` wins, then `codex`, then `agy` or `gemini`. If none is found, `proposer.md` and `skeptic.md` are still created with the literal placeholder `unset` for both fields — syntactically valid but non-functional — and `init` writes one stderr warning naming both files and instructing the user to edit them before running a debate.

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

Panel resolution uses `--with` in the provided order when present. Repeatable flags and comma-separated selectors are equivalent ways to provide the ordered explicit panel:

```sh
debate "Pick the safest migration path" --with proposer --with skeptic
debate "Pick the safest migration path" --with proposer,skeptic
```

Otherwise panel resolution uses `--table <name>` or `tables/default.yml`. Naming a synthesizer-role persona or the same resolved persona more than once in a panel fails before any backend session opens.

## 6a. Global Agent Skill Installation

`debate init` installs or repairs one bundled [Agent Skill](https://code.claude.com/docs/en/skills) named `debate`, targeting the user's home directory rather than the project. This is global setup, separate from the `.heurema/debate` scaffold above, and only `init` performs it — `debate new`, `debate version`, and ordinary `debate` runs never write global skill state. There is no `debate skills` (or similarly named) subcommand in v1; re-running `init` is the install/repair path.

The bundled skill lives in the repository under `internal/debate/skills/bundled/debate/` (a directory named `debate`, satisfying the Agent Skills naming contract) and is embedded into the binary at build time, so installation needs no network access or external registry.

Target selection is detection-driven:

- `~/.agents/skills/debate` — the standard Agent Skills path, installed when `codex` or `gemini` is found on `PATH`, or `~/.agents` already exists.
- `~/.claude/skills/debate` — a Claude Code compatibility copy, installed when `claude` is found on `PATH`, or `~/.claude` already exists.

Both targets install independently; detecting both Codex and Gemini still yields exactly one standard target (no per-client duplicate copies), and no `~/.codex/skills` or `~/.gemini/skills` directory is created in v1. Existing `~/.agents` or `~/.claude` directories are treated as sufficient evidence for these *targets* but never choose the *runtime* (model, backend) defaults in §6 or §7 — those require an executable on `PATH`.

Each managed install carries a `.debate-skill.json` metadata file (schema marker, debate binary version, one managed content checksum, and target/source info) alongside the skill content, excluded from the checksum itself. The checksum is SHA-256 over each file's relative path and content, sorted by path, hashed once — computed identically for on-disk installed content and for the bundled content, so the two are directly comparable. Install/repair is a two-step comparison:

1. If the target is absent, create it and record the bundled checksum as managed.
2. If present with recognizable managed metadata, compare the on-disk checksum to the recorded one. A mismatch means local edits — preserve the target and warn, without inspecting the bundled content at all.
3. Only when on-disk content still matches the recorded checksum does init compare the *bundled* checksum against it: equal means current (no-op); different means the bundled skill changed since install, so init overwrites the target with the new content and refreshes the recorded checksum.

A target that exists without recognizable debate-managed metadata is treated as unmanaged and preserved with a warning, the same as a locally-modified managed target. Target paths are derived only from the detected home directory and the two fixed roots above — no user-supplied path is accepted — and the installer refuses to write through a symlinked target or intermediate directory, skipping it with a warning instead. None of this affects `debate init`'s exit code: skill-install failures, missing `HOME`, or no detected client are reported as warnings on stderr, while stdout keeps reporting created/skipped workspace files and created/updated/current/skipped skill installs.

## 7. Synthesizer

The synthesizer produces the final answer from the transcript. It does not participate in the debate panel.

Resolution order:

1. `--synth <persona>`
2. selected table `synth`
3. uniquely resolved selector `synthesizer`
4. built-in default synthesizer

The built-in default is capability-aware rather than hard-coded to one model. It resolves a (model, backend) pair with low effort and a neutral synthesis prompt as follows:

1. If every panel persona shares one supported backend family (all `claude-agent-acp`, all `codex-acp`, or all `agy`), reuse that family.
2. Otherwise, detect the first supported local runtime on `PATH`, checked in order: `claude` (`claude-haiku-4-5` / `claude-agent-acp`), then `codex` (`codex` / `codex-acp`), then `agy` or `gemini` (`gemini-pro` / `agy`).
3. If neither resolves, the run fails before opening any session with an actionable error rather than silently assuming Claude is installed.

This is the same detection precedence `debate init` uses for starter persona defaults (see §6a) — both are driven by executable-on-`PATH` presence only, never by home-directory evidence. A custom synthesizer persona is only needed when the project wants a different model, backend, or synthesis style. Synthesizer resolution rejects debater-role personas.

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
--with <persona>      Add debater persona selectors. Repeat or separate selectors with commas.
--synth <persona>     Use a synthesizer persona for the final answer.
--task <value>        Read task from inline text, @file, or - for stdin.
--max-rounds <n>      Limit debate rounds. Defaults to 10.
--json                Emit JSON final result on stdout.
-q, --quiet           Suppress stderr progress events.
--sealed              Thread read-only intent into backend transports where supported.
```

IO contract:

- stdout: final-result-only; human mode writes only the final answer, and `--json` writes only the existing result JSON object
- stderr: default-on agent-readable progress events for debate runs, plus unprefixed CLI errors and unavoidable backend/process noise
- exit `0`: settled
- exit `2`: not converged
- exit `1`: error

Progress events are a v1 line protocol on stderr. Every machine-readable line begins exactly with `@@DEBATE_PROGRESS ` followed by one JSON object. Only prefixed lines are part of the progress contract; consumers should ignore unprefixed stderr. `--json` affects only stdout result formatting and does not suppress, redirect, or alter progress. `--quiet` suppresses all progress event lines while preserving final-result stdout and normal CLI error reporting.

Every progress object has:

- `version: 1`
- `type`
- `stage`
- `elapsed_ms`

V1 `type` values are:

```text
run_started
workspace_loaded
session_opening
session_opened
round_started
turn_started
heartbeat
turn_completed
round_completed
synthesis_started
synthesis_completed
run_completed
run_failed
```

Stage mapping is:

| Event type | Stage |
|---|---|
| `run_started`, `workspace_loaded` | `loading_workspace` |
| `session_opening`, `session_opened` | `opening_session` |
| `round_started`, `round_completed` | `running_round` |
| `turn_started`, `turn_completed` | `running_turn` |
| participant-send `heartbeat` | `running_turn` |
| synthesizer-send `heartbeat` | `synthesizing` |
| `synthesis_started`, `synthesis_completed` | `synthesizing` |
| `run_completed` | `completed` |
| `run_failed` | active lifecycle stage when known, otherwise `failed` |

Event-specific required fields are:

| Event type | Required fields |
|---|---|
| `session_opening` | `speaker` |
| `session_opened` | `speaker`, `duration_ms` |
| `round_started` | `round` |
| `turn_started` | `round`, `speaker` |
| participant-send `heartbeat` | `round`, `speaker`, `silence_ms` |
| synthesizer-send `heartbeat` | `silence_ms` |
| `turn_completed` | `round`, `speaker`, `duration_ms` |
| `round_completed` | `round`, `duration_ms` |
| `synthesis_completed` | `duration_ms` |
| `run_completed` | `duration_ms` |
| `run_failed` | `error` |

`round` is 1-based. `speaker` is the resolved participant identity used by the runner/orchestrate participant list and session routing. `elapsed_ms`, `duration_ms`, and `silence_ms` are non-negative integer milliseconds when present. `run_failed.error` includes the underlying error text and does not replace the existing CLI error line.

Lifecycle event cardinality is deterministic for non-failing runs: exactly one `run_started`, one `workspace_loaded`, matching `session_opening` and `session_opened` for each participant session, matching `round_started` and `round_completed` for each completed round, matching `turn_started` and `turn_completed` for each participant turn, one `synthesis_started`, one `synthesis_completed`, and final `run_completed`. `turn_started` is emitted before the participant `Session.Send` call, and `turn_completed` is emitted after the turn is appended. On failure, at most one `run_failed` is emitted; when present, it is the final progress event and `run_completed` is not emitted.

Heartbeat cadence is fixed at 1000 milliseconds in v1. A blocking send window begins immediately before a participant `Session.Send` or synthesizer `Send` call and ends immediately after it returns. No heartbeat is required for sends that return before 1000 milliseconds. If the send is still blocked, heartbeats continue once per additional 1000 milliseconds. `silence_ms` is measured from the start of the current send window, is monotonically non-decreasing within that window, and resets for the next send. Participant heartbeats include the active `round` and `speaker`; synthesizer heartbeats use stage `synthesizing` and do not include invented round or speaker values.

Progress writes are serialized through one emitter so heartbeat and lifecycle goroutines cannot interleave prefixed JSON lines.

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
- Exactly one bundled Agent Skill (`debate`) ships in v1; no `debate skills` or similar subcommand exists, and only `debate init` installs or repairs it globally.
- Runtime model/backend defaults (starter personas and the built-in synthesizer fallback) are capability-aware via executable-on-`PATH` detection, not hard-coded to Claude.

## 12. Future Work

- Public engine module after another product validates the boundary.
- API backend with grounding tool loop.
- Allow and deny lists for project paths.
- Network allowlist for grounded mode.
- Persona selection by tags.
- Pipeline-style orchestration over multiple debate runs.
