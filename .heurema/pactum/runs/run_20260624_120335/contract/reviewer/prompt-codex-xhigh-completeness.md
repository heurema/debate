# Contract Review: Completeness

You are reviewing a software change contract through the **contract-completeness** lens.

Review the contract fields below using only your assigned lens checklist.
Do not flag issues that belong to other lenses.

## Contract

**Goal**: Implement an exec backend transport under internal/backend/exec (standard library only) that drives stateless plain-CLI agents like Gemini via agy, reconstructing full conversation context per turn from accumulated prompts and replies, and wire it into the cmd/debate production resolver while internal/engine stays unchanged.

**Scope in**:
  - Implement internal/backend/exec: a transport.Transport that drives a stateless CLI agent by spawning a fresh subprocess per Send via exec.CommandContext, using only the Go standard library.
  - Implement the stateless full-render Session: record each Send prompt and the agent's reply (committed only on success), and on each Send reconstruct the full stdin (system folded in, the recorded prompt/reply history, and the new prompt) for a fresh subprocess.
  - Implement per-backend command derivation (agy) with an env override, spec.Model wiring, grounded-vs-sealed working directory, ctx cancellation, and an explicit error-classification mapping with a single retry.
  - Add deterministic unit tests via a fake CLI / injectable command runner, and a real-agy integration test gated behind a build tag and env var that the gate compiles but skips.
  - Wire the cmd/debate production resolver to register the exec backend for the agy backend, update the Slice-5 unimplemented-backend test, and add targeted exec dep-guard commands.

**Scope out**:
  - The api backend (a later concern).
  - Any change to internal/engine source (including transport.Classify and the orchestrate Delta render), the internal/debate policy/config packages, or the acp/echo/mock backends beyond what wiring the resolver requires.
  - debate init/new scaffolding.
  - Making the default test run perform any real agy, model, or network call (local stub subprocesses in tests are allowed); and Windows/non-Unix support.

**Acceptance criteria**:
  - internal/backend/exec provides a constructor returning a transport.Transport for the agy backend; it never imports internal/debate or cmd/debate and uses only the Go standard library plus internal/engine and internal/backend.
  - Open(ctx, spec) returns a transport.Session holding the spec and an empty recorded history and resolves the working directory (the process working directory for grounded, a fresh empty temporary directory for sealed); it spawns no process because the CLI is invoked per Send.
  - The command default argv is [agy, "--model", spec.Model] reading the prompt from stdin; the executable token (argv[0]) is overridable via env var DEBATE_AGY_COMMAND while the model argument is preserved; spec.Model must be non-empty (fail-fast otherwise). Tests assert the default and overridden argv and the model wiring. The exact real-agy flags are exercised only by the gated integration test.
  - Session.Send(ctx, prompt) spawns a fresh subprocess via exec.CommandContext(ctx, ...) of the resolved command with its working directory set to the resolved Cwd, writes the reconstructed stdin, closes stdin, and concurrently reads stdout to completion while draining stderr to a buffer; on a zero exit it returns the stdout content as transport.Result and discards the stderr buffer; on a non-zero exit it returns a terminal error whose message includes the captured stderr content (truncated to a reasonable fixed limit, e.g. 4 KiB) for diagnosis.
  - The reconstructed stdin is exact and tested: spec.System first (once, omitted entirely when empty), then the recorded history as alternating fixed-labeled blocks (each prior Send prompt and the agent's prior reply), then the current prompt. The canonical byte-level format is: each block opens with its label on its own line—`[system]` for the system block, `[prompt]` for every Send prompt (prior and current), `[reply]` for every prior agent reply—followed by a newline and then the block content verbatim; if the content does not already end with `\n` one is appended so the block always ends on a clean newline boundary; consecutive blocks are separated by exactly one blank line, i.e. a single additional `\n` after the block's trailing `\n` (yielding the byte sequence `\n\n` between a block's last content byte and the next label line); the entire stdin ends with the trailing `\n` of the last block's content with no additional blank line or separator after it. Prior Send prompts and agent replies appear in history order with each prompt under `[prompt]` and each reply under `[reply]`; the current Send prompt is always last. Tests assert the exact reconstructed byte sequence across a multi-Send sequence (at least three turns), verify the system block is present when spec.System is non-empty and absent entirely when it is empty, and confirm prior agent replies appear verbatim under their `[reply]` labels.
  - The prompt and reply are committed to the recorded history only after a zero-exit success; a failed or canceled Send (including its single retry) does not mutate the committed history, so a retry reconstructs identical stdin and later Sends are not polluted by the failure.
  - Because the CLI is stateless, the Session is the sole conversation memory and reconstructs full both-sided context every Send; there is no persistent subprocess.
  - spec.System is always folded into the stdin input and never dropped; spec.ReadOnly (from --sealed) selects the subprocess working directory (process working directory vs fresh empty temp dir); network is available in both modes; filesystem read-only is not enforced by the exec backend (a plain CLI has no sandbox) and is documented as trusted/best-effort.
  - Error classification is explicit and consistent with transport.Classify without modifying internal/engine: a spawn failure or a stdin broken-pipe error is retryable; a non-zero process exit is terminal (a usage/refusal result) with the error message including the captured stderr content (read concurrently and truncated at a reasonable fixed limit) for diagnosis; ctx cancellation returns a terminal canceled error. A retryable failure is retried exactly once (a fresh subprocess with identical reconstructed stdin) before the error is surfaced.
  - ctx cancellation during a Send terminates the subprocess (via exec.CommandContext) and returns promptly with the cancellation error; tests cover cancellation before spawn and during the subprocess run.
  - Close releases session resources (including removing a sealed temporary directory) and is idempotent; there is no persistent process to terminate.
  - The subprocess spawn is behind an injectable command runner so deterministic tests assert the command, arguments, working directory, and stdin content without invoking a real external program.
  - Deterministic unit tests (a fake CLI stub program and/or the injectable runner, no real agy) cover: accumulation and exact full-render reconstruction across multiple Sends (including prior replies), system folding, stdin/stdout handling, default and env-overridden command, model wiring, grounded vs sealed working directory, the error-classification mapping, the single retry on a stdin broken pipe with unchanged history, and ctx cancellation.
  - A real-agy integration test is gated behind build tag //go:build exec_integration AND env var DEBATE_EXEC_INTEGRATION: with the tag set but DEBATE_EXEC_INTEGRATION unset it compiles and skips; the gate includes a command that compiles it under the tag with the env var cleared.
  - The cmd/debate production resolver registers the exec backend for the agy backend id and keeps acp and echo; the Slice-5 unimplemented-backend e2e fixture is updated to use a still-unimplemented backend (api/unknown) for its fail-fast assertion, and a test asserts that the agy backend resolves to a transport without running a real program in the default suite.
  - internal/engine is not modified — enforced by `git diff --exit-code -- internal/engine/` (which must produce no output, i.e. exit 0 with no diff) and the engine dep-guard (which keeps internal/engine free of new dependencies); targeted dep-guards enforce that internal/backend/exec depends only on the Go standard library, internal/engine, and internal/backend (no third-party dependency) in both the default and the exec_integration build, while the broader internal/backend dep-guard (which allows the acp sdk for the acp backend) still passes; check-gofmt, go vet ./..., go build ./..., and go test ./... pass with no real backend invoked.

**Validation commands**:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - env -u DEBATE_EXEC_INTEGRATION go test -count=1 -tags exec_integration ./internal/backend/...
  - go vet ./...
  - git diff --exit-code -- internal/engine/
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/backend/exec/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend
  - env GOFLAGS=-tags=exec_integration bash scripts/dep-guard.sh ./internal/backend/exec/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend
  - bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

**Assumptions**:
  - The agy CLI reads its prompt from stdin and writes the reply to stdout; its exact flags are configurable and only exercised by the gated integration test; the default argv convention is [agy, --model, spec.Model].
  - Each Send prompt is an opaque, already-rendered block produced by the debate PromptBuilder; the exec backend never re-renders it, only records and concatenates prompts and replies to reconstruct full context.
  - The exec backend is stateless at the process level; conversation state (incoming prompts and the agent's own prior replies) lives in the Session, so the orchestrate Delta render is unchanged; Send calls are serialized per Session by orchestrate (one turn at a time), so the Session is not required to be goroutine-safe.
  - spec.System is folded into the stdin input and spec.Model is passed per the command convention; filesystem read-only is best-effort/trusted for a plain CLI (no sandbox), with grounding controlled by the subprocess working directory; sealed mode uses a fresh empty temp working directory and does not disable network.
  - ctx cancellation is handled via exec.CommandContext and surfaces as a terminal canceled error; recovery for a stateless backend is a single retry of the Send (there is no session to reopen), and a failed/canceled Send does not commit to history.
  - internal/backend/exec is standard-library only; real backends live under internal/backend to keep internal/engine dependency-free; the runner stays backend-agnostic via the injected Resolver.
  - Subprocess behavior targets Unix-like platforms (darwin/linux); other GOOS support is out of scope.

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
