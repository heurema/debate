# Contract Review: Completeness

You are reviewing a software change contract through the **contract-completeness** lens.

Review the contract fields below using only your assigned lens checklist.
Do not flag issues that belong to other lenses.

## Contract

**Goal**: Replace the hand-written stdlib flag parsing in cmd/debate with github.com/alecthomas/kong (the CLI library pactum uses), so flags parse correctly in any position, while preserving every existing command, flag, and exit-code behavior.

**Scope in**:
  - Add github.com/alecthomas/kong and parse cmd/debate arguments with it, removing the hand-rolled flag.FlagSet / os.Args switch parsing in cmd/debate/main.go and cmd/debate/scaffold.go.
  - Model the debate run as kong's default command (bare `debate "<task>"`) alongside the version, init, and new subcommands, with all flags as kong struct tags.
  - Preserve the exact existing behavior (commands, flags, task composition, exit codes, stdout/stderr) and make flags parse in any position relative to positionals.
  - Update the cmd/debate tests for kong, including assertions that flags work both before and after the positional argument.

**Scope out**:
  - Any change to internal/engine, internal/debate, or internal/backend.
  - Adding new subcommands (no validate) or new flags, or changing the backend resolver, runner, synthesizer, or IO contract.
  - Changing the debate algorithm, personas, or config.

**Acceptance criteria**:
  - cmd/debate parses all arguments with github.com/alecthomas/kong (added to go.mod and go.sum); the hand-rolled flag.NewFlagSet parsing and os.Args dispatch in main.go and scaffold.go are removed.
  - The CLI preserves the existing commands: a default run action invoked as `debate "<task>"` with no subcommand word, plus subcommands `version`, `init`, and `new <name>`; the run is kong's default command so the bare-task form keeps working.
  - Flags parse in any position relative to the positional argument: `debate "<task>" --json`, `debate --json "<task>"`, `debate --max-rounds 2 "<task>"`, `debate "<task>" --max-rounds 2`, and `debate new <name> --role synthesizer` all apply the flag correctly. (This fixes the previous flags-after-positional bug.)
  - All existing run flags are preserved with the same names and meaning: --with (panel selectors), --synth, --max-rounds, --json, -q/--quiet, --sealed, and --task (@file); the new subcommand keeps its --role flag (debater|synthesizer, default debater).
  - Task sources still compose: the positional task, --task @file (file contents), and stdin (appended when piped); an empty resulting task is a fail-fast error.
  - Exit codes are unchanged: 0 when settled, 2 when not converged (stalemate or max), 1 on error; an unknown flag or subcommand prints a clear kong usage/help message with a non-zero exit; the stdout=final-answer / stderr=live-trace (auto-quiet off-TTY or with -q) contract and the --json output shape are unchanged.
  - cmd/debate tests are updated for kong and include assertions that a representative flag is honored both before and after the positional argument for the run command and for `new`.
  - go.mod and go.sum gain github.com/alecthomas/kong; internal/engine, internal/debate, and internal/backend are unchanged; check-gofmt, go build ./..., go vet ./..., and go test ./... pass, and the engine/backend/debate dep-guards still pass.

**Validation commands**:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - go vet ./...
  - go run ./cmd/debate version
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

**Assumptions**:
  - kong (github.com/alecthomas/kong, the library pactum uses) parses flags in any position by design and provides usage/help and exit handling; the debate run is modeled as a default command (e.g. kong's default:"withargs") so the bare `debate "<task>"` form coexists with the version/init/new subcommands.
  - Only the argument-parsing layer changes; the backend resolver, runner, synthesizer, and IO contract are untouched, so behavior is identical except for the flag-position fix.
  - No subcommands are added (validate was never implemented and remains out of scope); kong is a cmd/debate-only dependency and must not leak into internal/engine, internal/debate, or internal/backend.

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
