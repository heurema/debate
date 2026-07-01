# Workspace format

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

## Personas

Markdown files with YAML front matter:

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

Required: `version: 1`, `model`, `effort`, and a non-empty body. Optional: `role`
(`debater` or `synthesizer`, defaults to `debater`), `backend` (overrides model-based
backend inference), `tags` (reserved for selection features). Unknown fields fail fast.

Persona files live at `personas/<name>.md` or `personas/<namespace>/<name>.md` (one
namespace level only). The persona ID is the relative path without `.md`, e.g. `skeptic`
or `reviewers/security`. Hidden and non-Markdown files are ignored.

Model name infers a default backend unless `backend` is set explicitly:

| Model pattern | Default backend |
|---|---|
| `claude-*`, `opus`, `sonnet`, `haiku`, `fable` | `claude-agent-acp` |
| `gpt-*`, `codex`, `o*` | `codex-acp` |
| `gemini-*` | `agy` |

## Tables

Flat YAML files under `tables/`:

```yaml
version: 1
panel:
  - proposer
  - skeptic
# synth: final-judge
```

Required: `version: 1` and a non-empty `panel`. Optional `synth` names the synthesizer
for that table using the same selector rules as `--synth`.

## Selectors

`namespace/name` is an exact persona ID. A short selector (no `/`) first resolves an
exact root persona ID; otherwise it resolves only when exactly one loaded persona has
that short name. Ambiguous short names fail and list the qualified candidate IDs.

## Synthesizer resolution

1. `--synth <persona>`
2. selected table's `synth`
3. a uniquely resolved `synthesizer`-role persona
4. the built-in default synthesizer

The built-in default matches the debate panel's backend family when every panel persona
shares one (all Claude, all Codex, or all Gemini/agy); otherwise it detects the first
supported local runtime on PATH (claude, then codex, then agy/gemini). If none is
available and the panel isn't homogeneous, the run fails before opening any session with
an actionable error — add `--synth`, a table `synth`, a synthesizer persona, or set
persona backend/model explicitly.
