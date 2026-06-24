# Contract Review: Validation soundness

You are reviewing a software change contract through the **validation-soundness** lens.

Review the contract fields below using only your assigned lens checklist.
Do not flag issues that belong to other lenses.

## Contract

**Goal**: Add debate init and debate new scaffolding subcommands to cmd/debate that create a ready-to-run .heurema/debate workspace and new persona files, adding no new module dependency, without changing internal/engine, internal/debate, or internal/backend.

**Scope in**:
  - Implement the `debate init` subcommand: scaffold a .heurema/debate workspace under the current directory with two starter debater personas and a context.md template, safely (never overwriting existing files).
  - Implement the `debate new <name>` subcommand: create a new persona file from a template under a discovered .heurema/debate/personas, with a role flag, safely (never overwriting an existing persona).
  - Make the scaffolded workspace immediately loadable and runnable (valid personas that load via the existing config/persona packages).
  - Add deterministic unit tests using temporary directories that assert init and new behavior, including the loadability of the scaffolded workspace and the refuse-to-overwrite behavior.

**Scope out**:
  - Any change to internal/engine, internal/debate, or internal/backend source (the subcommands live in cmd/debate and reuse config/persona by import only).
  - Backends, the debate run path, the synthesizer, or convergence behavior.
  - Editing or migrating an existing workspace's content beyond adding new files; and any new third-party module dependency.

**Acceptance criteria**:
  - `debate init` creates a .heurema/debate directory under the current working directory containing personas/proposer.md and personas/skeptic.md (each a valid debater persona with role debater, a concrete model and effort, and a system-prompt body) and a context.md template file; it prints the paths it created.
  - The scaffolded workspace loads successfully via the existing config.Load: discovery finds the new .heurema/debate, the panel resolves to the two starter debaters, and the built-in default synthesizer is used (init does not scaffold a synthesizer file or a config.yml, since the default panel is all debater personas).
  - The two starter personas use concrete valid values (a real model id such as claude-haiku-4-5 and a valid effort) so the workspace is immediately runnable without edits; persona.ParseFile accepts both.
  - `debate init` is safe: if a target file already exists it does not overwrite it (it skips that file with a clear message or refuses with a clear error and a documented exit code); an existing .heurema/debate is never clobbered.
  - `debate new <name>` creates <name>.md under the discovered .heurema/debate/personas from a template with YAML frontmatter (role defaulting to debater, overridable via --role debater|synthesizer; a concrete model and effort default; optional backend) and a placeholder body; it prints the created path, and the created file is accepted by persona.ParseFile.
  - `debate new` validates the name (a simple persona id, rejecting path separators) and refuses to overwrite an existing persona file with a clear error; it requires a discoverable .heurema/debate workspace (walking up parent directories like config discovery) and errors clearly when none is found, creating the personas directory within the discovered workspace if it does not yet exist.
  - init and new write only under .heurema/debate, never outside it, use the Go standard library plus the existing internal packages, and add no new module dependency (go.mod and go.sum are unchanged); an unknown flag or a missing required argument prints a clear usage error with a non-zero exit.
  - Deterministic unit tests using temporary directories assert: init creates a workspace that config.Load accepts with the two-debater panel, re-running init does not overwrite existing files, new creates a persona that persona.ParseFile accepts, and new refuses to overwrite an existing persona; check-gofmt, go vet ./..., go build ./..., and go test ./... pass.

**Validation commands**:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - go vet ./...
  - go run ./cmd/debate version
  - git diff --exit-code -- go.mod go.sum
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

**Assumptions**:
  - init scaffolds exactly two debater personas (proposer and skeptic) and relies on the built-in default synthesizer, matching the product design; it does not scaffold a config.yml because the default panel is all debater personas.
  - Starter personas use concrete model/effort defaults so the workspace runs out of the box; the bodies are short starter prompts the user is expected to edit.
  - debate new defaults the role to debater and writes a template persona the user edits before running; the model/effort are concrete defaults that already parse.
  - Scaffolding writes plain files under .heurema/debate and does not modify internal/engine, internal/debate, or internal/backend source; it reuses config/persona only by import (in code and tests) and introduces no new third-party dependency.
  - The context.md template is an optional baseline preamble example (a short comment/placeholder) that config.Load accepts (an empty or comment-only context is fine).

## Lens: Validation soundness

Checklist:
- Are validation.commands gate-runnable (no shell forms the gate cannot execute)?
- Are they non-vacuous: would they fail on wrong output?
- Are they self-consistent and not contradictory with the tests?

## Output

Report likely-real defects (recall-first), then gate on precision before marking blocking.
Use state=candidate with explicit uncertainty when you believe a finding is real but have not fully confirmed it.

State your analysis in prose. If you find issues, also include a structured block:

```json
{
  "schema": "pactum.contract_reviewer_result.v1alpha1",
  "findings": [
    {
      "message": "Describe the contract issue clearly.",
      "severity": "medium",
      "category": "quality",
      "blocking": true,
      "evidence": "Quote or cite the contract field that shows the issue.",
      "material_impact": "Concrete way this spec defect would make the implementation wrong, ambiguous, or stuck.",
      "fix_direction": "What the contract author should change to resolve this.",
      "uncertainty": "Any doubt about this finding — omit if confident.",
      "state": "candidate"
    }
  ]
}
```

Rules:
- Use severity: low, medium, high, critical.
- Use category: correctness, scope, quality, validation, process, other.
- Omit file and line (not applicable for contract review).
- Set state=candidate when likely real but not fully confirmed; set state=confirmed when certain.
- HARD RULE: blocking=true is allowed ONLY for a material spec defect that would make the implementation wrong, ambiguous, or stuck.
- Wording, style, naming, redundancy, and completeness/thoroughness preferences MUST be blocking=false (advisory).
- Every blocking finding MUST include a concrete material_impact explaining the implementation consequence.
- If you cannot state a concrete material_impact, mark the finding blocking=false (advisory).
- Set blocking=false for advisory issues.
- If no issues, say so clearly. Do not include an empty findings block.
