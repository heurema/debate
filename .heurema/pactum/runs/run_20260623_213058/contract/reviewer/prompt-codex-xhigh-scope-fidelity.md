# Contract Review: Scope fidelity

You are reviewing a software change contract through the **scope-fidelity** lens.

Review the contract fields below using only your assigned lens checklist.
Do not flag issues that belong to other lenses.

## Contract

**Goal**: Slice 0: bootstrap the debate Go project skeleton — go.mod (module github.com/heurema/debate), package layout (internal/engine/{loop,transport,orchestrate}, internal/debate, cmd/debate), Makefile, and a 'debate version' command. No engine or debate logic yet.

**Scope in**:
  - Create a root Go module at go.mod with module path github.com/heurema/debate.
  - Create the Slice 0 package layout: cmd/debate, internal/debate, internal/engine/loop, internal/engine/transport, and internal/engine/orchestrate.
  - Add minimal Go source files needed for all Slice 0 packages to compile and be listed by go list ./....
  - Implement a minimal debate CLI entrypoint with a version subcommand.
  - Add a root Makefile with build, vet, test, and check targets.
  - Add at least one trivial Go test so the repository has a runnable test suite.

**Scope out**:
  - Do not implement engine loop behavior, orchestration, transport behavior, mock backend behavior, debate policy, prompt building, verdict logic, persona parsing, synthesizer behavior, or model/backend integrations.
  - Do not implement .heurema/debate discovery, config.yml parsing, context.md loading, init/new commands, or full debate task execution.
  - Do not rename the physical repository directory or migrate per-project memory/state.
  - Do not add network calls, external model calls, credentials, or nonessential third-party dependencies.

**Acceptance criteria**:
  - go.mod exists at the repository root and declares module github.com/heurema/debate.
  - go list ./... succeeds and includes github.com/heurema/debate/cmd/debate, github.com/heurema/debate/internal/debate, github.com/heurema/debate/internal/engine/loop, github.com/heurema/debate/internal/engine/transport, and github.com/heurema/debate/internal/engine/orchestrate.
  - The root Makefile defines build, vet, test, and check targets; make check runs the Slice 0 validation path successfully.
  - go run ./cmd/debate version exits 0 and prints a non-empty version string identifying the debate binary.
  - go test ./... succeeds with at least one trivial test present.
  - The implementation remains a skeleton: no engine/debate runtime logic, backend integrations, persona/config discovery, or synthesizer behavior is added.

**Validation commands**:
  - go list ./...
  - go test ./...
  - go vet ./...
  - go run ./cmd/debate version
  - make build
  - make check

**Assumptions**:
  - A default development version string is acceptable for Slice 0 when no release metadata is supplied.
  - Placeholder package files are acceptable where needed to make empty Slice 0 packages compile and appear in go list ./....
  - Slice 0 keeps the repository directory name unchanged even though the Go module path is github.com/heurema/debate.
  - The Go standard library is sufficient for this slice unless the existing project later introduces a required dependency.

## Lens: Scope fidelity

Checklist:
- Is scope.in coherent with and proportionate to the goal?
- Is scope.out coherent and not contradictory with scope.in?
- Is the scope neither over-broad nor under-broad for the stated goal?

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
