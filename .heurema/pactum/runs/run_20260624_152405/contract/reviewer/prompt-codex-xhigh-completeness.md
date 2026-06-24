# Contract Review: Completeness

You are reviewing a software change contract through the **contract-completeness** lens.

Review the contract fields below using only your assigned lens checklist.
Do not flag issues that belong to other lenses.

## Contract

**Goal**: Remove the context.md baseline-preamble feature so debate context lives only in the task (plus the grounded sandbox): drop config.Workspace.Context and context.md loading, assemble the brief from the task alone, and stop scaffolding context.md in debate init.

**Scope in**:
  - Drop the Context field from config.Workspace and stop reading .heurema/debate/context.md in internal/debate/config.
  - Make internal/debate/runner assemble the debate brief from the task alone and remove every use of Workspace.Context in the runner and synthesizer.
  - Make debate init in cmd/debate scaffold only the two starter debater personas and no longer create a context.md file.
  - Update all affected tests so none assert context.md creation or a Workspace.Context field.

**Scope out**:
  - Any change to grounding, the backends, personas, or the engine.
  - Docs updates (DESIGN.md / SLICES.md) — handled separately.
  - Introducing any new dependency or new feature.

**Acceptance criteria**:
  - config.Workspace no longer has a Context field, and config.Load no longer opens or reads .heurema/debate/context.md; a context.md present in a workspace is ignored and does not affect loading or cause an error.
  - internal/debate/runner assembles the debate brief from the task text alone (no context preamble), and neither the runner nor the synthesizer references Workspace.Context.
  - `debate init` scaffolds only personas/proposer.md and personas/skeptic.md under .heurema/debate and does not create a context.md file; it prints only the persona paths it created.
  - The scaffolded workspace still loads via config.Load with the two-debater panel and the built-in default synthesizer.
  - All affected tests are updated: no test asserts context.md creation or a Workspace.Context field; the init test asserts the two persona files are created and the workspace loads (two-debater panel) with no context.md; the runner test asserts the assembled brief equals the task.
  - check-gofmt, go build ./..., go vet ./..., and go test ./... pass; go.mod and go.sum are unchanged (no new dependency); the engine and debate dep-guards still pass.

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
  - Debate context is supplied in the task (optionally via --task @file) and deeper project context is read by the grounded agents themselves; the static context.md preamble is therefore redundant and removed.
  - Removing Workspace.Context is an internal-only change; only internal callers in runner and cmd reference it.
  - A pre-existing context.md in a user workspace becomes a no-op (ignored), not an error.

## Lens: Completeness

Checklist:
- Does the contract fully cover its goal? Are there gaps in scope or acceptance_criteria?
- Is every acceptance criterion specific and observable enough to verify?

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
