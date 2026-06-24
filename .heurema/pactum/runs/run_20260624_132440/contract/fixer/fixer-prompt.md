# Contract Review Fixer Prompt

You are fixing a software change contract to address blocking review findings.

Current contract version: 2da31eaf7c88f4d8836a11de2f2f88db22d9543adf25d2f31400d50bd57f82ec

## Current Contract

**Goal**: Add debate init and debate new scaffolding subcommands to cmd/debate that create a ready-to-run .heurema/debate workspace and new persona files, using the standard library only, without changing internal/engine, internal/debate, or internal/backend.

**Scope in**:
  - Implement the `debate init` subcommand: scaffold a .heurema/debate workspace under the current directory with two starter debater personas and a context.md template, safely (never overwriting existing files).
  - Implement the `debate new <name>` subcommand: create a new persona file from a template under .heurema/debate/personas, with a role flag, safely (never overwriting an existing persona).
  - Make the scaffolded workspace immediately loadable and runnable (valid personas that load via the existing config/persona packages).
  - Add deterministic unit tests using temporary directories that assert init and new behavior, including the loadability of the scaffolded workspace and the refuse-to-overwrite behavior.

**Scope out**:
  - Any change to internal/engine, internal/debate, or internal/backend source beyond importing config/persona for validation in tests.
  - Backends, the debate run path, the synthesizer, or convergence behavior.
  - Editing or migrating an existing workspace's content beyond adding new files.

**Acceptance criteria**:
  - `debate init` creates a .heurema/debate directory under the current working directory containing personas/proposer.md and personas/skeptic.md (each a valid debater persona with role debater, a concrete model and effort, and a system-prompt body) and a context.md template file; it prints the paths it created.
  - The scaffolded workspace loads successfully via the existing config.Load: discovery finds the new .heurema/debate, the panel resolves to the two starter debaters, and the built-in default synthesizer is used (init does not scaffold a synthesizer file).
  - The two starter personas use concrete valid values (a real model id such as claude-haiku-4-5 and a valid effort) so the workspace is immediately runnable without edits; persona.ParseFile accepts both.
  - `debate init` is safe and idempotent-friendly: if a target file already exists it does not overwrite it (it skips with a clear message or refuses with a clear error and documented exit code); an existing .heurema/debate is never clobbered.
  - `debate new <name>` creates .heurema/debate/personas/<name>.md from a template with YAML frontmatter (role defaulting to debater, overridable via --role debater|synthesizer; a model and effort placeholder or default; optional backend) and a placeholder body; it prints the created path.
  - `debate new` validates the name (a simple persona id, no path separators or .md suffix needed) and refuses to overwrite an existing persona file with a clear error; if no .heurema/debate workspace is found, `debate new` exits with a non-zero code and a clear error directing the user to run `debate init` first; if the workspace exists but the personas subdirectory does not, `debate new` creates it before writing the new file.
  - init and new write only under .heurema/debate, never outside it, and use only the Go standard library (no new dependency); unknown flags or a missing required argument print a clear usage error with non-zero exit.
  - The file-writing scaffolding logic for init and new must reside in a dedicated sub-package (e.g., cmd/debate/scaffold) that imports only Go standard library packages — no imports from github.com/heurema/debate/, github.com/coder/acp-go-sdk, gopkg.in/yaml.v3, or any other external module. The dep-guard for that sub-package must pass with no external dependencies allowed. The top-level cmd/debate wiring (cobra command registration) may reference the scaffold sub-package and the existing allowed external deps.
  - Deterministic unit tests using temporary directories assert: init creates a workspace that config.Load accepts with the two-debater panel, re-running init does not overwrite existing files, new creates a persona that persona.ParseFile accepts, new refuses to overwrite an existing persona, and new errors with a non-zero exit when no workspace exists; check-gofmt, go vet ./..., go build ./..., and go test ./... pass.

**Validation commands**:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - go vet ./...
  - go run ./cmd/debate version
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3
  - bash scripts/dep-guard.sh ./cmd/debate/scaffold/...
  - bash scripts/dep-guard.sh ./cmd/debate/... github.com/heurema/debate/ github.com/coder/acp-go-sdk gopkg.in/yaml.v3

**Assumptions**:
  - init scaffolds exactly two debater personas (proposer and skeptic) and relies on the built-in default synthesizer, matching the product design; it does not scaffold a config.yml because the default panel is all debater personas.
  - Starter personas use concrete model/effort defaults so the workspace runs out of the box; the bodies are short starter prompts the user is expected to edit.
  - debate new defaults the role to debater and writes a template persona the user edits before running; the model/effort may be concrete defaults or clearly-marked placeholders that still parse.
  - Scaffolding writes plain files under .heurema/debate and does not touch internal/engine, internal/debate, or internal/backend source; the subcommands live in cmd/debate and use the standard library only.
  - The context.md template is an optional baseline preamble example (a short comment/placeholder) and is valid to load (an empty or comment-only context is acceptable to config.Load).

## Blocking Findings to Address

1. [codex-xhigh/validation-soundness] The scaffold dep-guard validation command is not gate-runnable as listed because it passes only the package pattern and no allowed prefix.
   Evidence: Validation commands include: `bash scripts/dep-guard.sh ./cmd/debate/scaffold/...`; the contract also says the dep-guard for that sub-package must pass with no external dependencies allowed.

## Fixer Instructions

- Address each blocking finding by updating the relevant contract field.
- Do NOT change the goal field — it is out of scope for the fixer.
- Only include the contract fields you are changing in the output.
- base_version must exactly match the version shown above.

## Output

Output your reasoning, then a single JSON block with the revise payload:

```json
{
  "schema": "pactum.contract_revise.v1alpha1",
  "base_version": "2da31eaf7c88f4d8836a11de2f2f88db22d9543adf25d2f31400d50bd57f82ec",
  "contract": {
    "acceptance_criteria": ["...updated criteria..."],
    "validation": {"commands": ["...updated commands..."]}
  }
}
```

Omit any contract field you are not changing. Do not include the goal field.
