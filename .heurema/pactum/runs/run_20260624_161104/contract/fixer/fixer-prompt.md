# Contract Review Fixer Prompt

You are fixing a software change contract to address blocking review findings.

Current contract version: 197f7bc7d4697c31b7194dbb7d5e36a655fef157c59b33b3cb354a50264d8728

## Current Contract

**Goal**: Invoke agy non-interactively via --print in the exec backend so it works against the real agy CLI (which otherwise defaults to an interactive session and hangs).

**Scope in**:
  - Change the default agy argv in internal/backend/exec to include --print so agy runs a single prompt non-interactively from stdin and exits.
  - Update the affected exec unit tests (argv assertions) and the gated integration test to match the new argv.

**Scope out**:
  - Other backends, internal/engine, the internal/debate packages, and the acp backend.
  - The exec backend's stdin reconstruction / delta accumulation logic, error handling, grounding, and recovery.
  - CLI flag-ordering / argument parsing (handled separately).

**Acceptance criteria**:
  - The exec backend's default agy argv is [agy, "--print", "--model", spec.Model]; the --print flag (alias of -p) makes agy run a single prompt non-interactively, reading the prompt from stdin and printing the response before exiting. The prompt is still written to the subprocess stdin (the reconstruction logic is unchanged).
  - The DEBATE_AGY_COMMAND override still replaces only the executable token (argv[0]) and preserves the --print and --model arguments in order.
  - spec.Model must still be non-empty (fail-fast otherwise); the model is passed as the --model value.
  - Unit tests assert the new default argv (including --print) for both the default and the DEBATE_AGY_COMMAND-overridden command, and the gated integration test uses the same --print invocation.
  - check-gofmt, go build ./..., go vet ./..., and go test ./... pass; go.mod and go.sum are unchanged (no new dependency); internal/engine is unchanged and the engine/exec/backend/debate dep-guards still pass.

**Validation commands**:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - env -u DEBATE_EXEC_INTEGRATION go test -count=1 -tags exec_integration ./internal/backend/...
  - go vet ./...
  - git diff --exit-code -- go.mod go.sum
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/backend/exec/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend
  - bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk

**Assumptions**:
  - agy --print (alias -p) runs a single prompt non-interactively, reading the prompt from stdin and printing the model response, then exits; without it agy starts an interactive session and the subprocess hangs.
  - This is a focused argv change; the stdin reconstruction, grounding, error classification, retry, and Close behavior are all unchanged.
  - The verified real agy version is 1.0.11; --print/-p is its documented non-interactive single-prompt flag.

## Blocking Findings to Address

1. [codex-xhigh/testability] The acceptance criterion for real `agy --print` non-interactive behavior is not backed by a runnable validation command.
   Evidence: Acceptance criteria: "the --print flag (alias of -p) makes agy run a single prompt non-interactively..." Validation command: "env -u DEBATE_EXEC_INTEGRATION go test -count=1 -tags exec_integration ./internal/backend/..."
2. [codex-xhigh/validation-soundness] The validation command for the gated exec integration test appears to unset the very environment gate needed to run it, so it likely does not validate the real agy --print invocation.
   Evidence: Acceptance criteria: "the gated integration test uses the same --print invocation." Validation command: "env -u DEBATE_EXEC_INTEGRATION go test -count=1 -tags exec_integration ./internal/backend/..."

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
  "base_version": "197f7bc7d4697c31b7194dbb7d5e36a655fef157c59b33b3cb354a50264d8728",
  "contract": {
    "acceptance_criteria": ["...updated criteria..."],
    "validation": {"commands": ["...updated commands..."]}
  }
}
```

Omit any contract field you are not changing. Do not include the goal field.
