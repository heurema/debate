# Contract Review Fixer Prompt

You are fixing a software change contract to address blocking review findings.

Current contract version: 45eb82e67dabbfeb7d150e12dadffff4dabc46287d2019dbf3ce8a098edc8657

## Current Contract

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
  - Session.Send(ctx, prompt) spawns a fresh subprocess via exec.CommandContext(ctx, ...) of the resolved command with its working directory set to the resolved Cwd, writes the reconstructed stdin, closes stdin, reads stdout to completion, and on a zero exit returns the reply as transport.Result.
  - The reconstructed stdin is exact and tested: spec.System first (once, omitted entirely when empty), then the recorded history as alternating fixed-labeled blocks (each prior Send prompt and the agent's prior reply), then the current prompt, with a single blank line between blocks and a trailing newline at the end of the input. Each Send prompt is treated as an opaque already-rendered block (the backend concatenates, it does not re-render). Tests assert the exact reconstructed stdin bytes across a multi-Send sequence, including that the agent's own prior replies are included.
  - The prompt and reply are committed to the recorded history only after a zero-exit success; a failed or canceled Send (including its single retry) does not mutate the committed history, so a retry reconstructs identical stdin and later Sends are not polluted by the failure.
  - Because the CLI is stateless, the Session is the sole conversation memory and reconstructs full both-sided context every Send; there is no persistent subprocess.
  - spec.System is always folded into the stdin input and never dropped; spec.ReadOnly (from --sealed) selects the subprocess working directory (process working directory vs fresh empty temp dir); network is available in both modes; filesystem read-only is not enforced by the exec backend (a plain CLI has no sandbox) and is documented as trusted/best-effort.
  - Error classification is explicit and consistent with transport.Classify without modifying internal/engine: a spawn failure or a stdin broken-pipe error is retryable; a non-zero process exit is terminal (a usage/refusal result); ctx cancellation returns a terminal canceled error. A retryable failure is retried exactly once (a fresh subprocess with identical reconstructed stdin) before the error is surfaced.
  - ctx cancellation during a Send terminates the subprocess (via exec.CommandContext) and returns promptly with the cancellation error; tests cover cancellation before spawn and during the subprocess run.
  - Close releases session resources (including removing a sealed temporary directory) and is idempotent; there is no persistent process to terminate.
  - The subprocess spawn is behind an injectable command runner so deterministic tests assert the command, arguments, working directory, and stdin content without invoking a real external program.
  - Deterministic unit tests (a fake CLI stub program and/or the injectable runner, no real agy) cover: accumulation and exact full-render reconstruction across multiple Sends (including prior replies), system folding, stdin/stdout handling, default and env-overridden command, model wiring, grounded vs sealed working directory, the error-classification mapping, the single retry on a stdin broken pipe with unchanged history, and ctx cancellation.
  - A real-agy integration test is gated behind build tag //go:build exec_integration AND env var DEBATE_EXEC_INTEGRATION: with the tag set but DEBATE_EXEC_INTEGRATION unset it compiles and skips; the gate includes a command that compiles it under the tag with the env var cleared.
  - The cmd/debate production resolver registers the exec backend for the agy backend id and keeps acp and echo; the Slice-5 unimplemented-backend e2e fixture is updated to use a still-unimplemented backend (api/unknown) for its fail-fast assertion, and a test asserts that the agy backend resolves to a transport without running a real program in the default suite.
  - internal/engine is not modified — enforced by human review and the engine dep-guard (which keeps internal/engine free of new dependencies), not by a git-diff command; targeted dep-guards enforce that internal/backend/exec depends only on the Go standard library, internal/engine, and internal/backend (no third-party dependency) in both the default and the exec_integration build, while the broader internal/backend dep-guard (which allows the acp sdk for the acp backend) still passes; check-gofmt, go vet ./..., go build ./..., and go test ./... pass with no real backend invoked.

**Validation commands**:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - env -u DEBATE_EXEC_INTEGRATION go test -count=1 -tags exec_integration ./internal/backend/...
  - go vet ./...
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

## Blocking Findings to Address

1. [codex-xhigh/completeness] The contract requires exact stdin reconstruction but does not define the canonical labels or byte-level newline policy for system, prompt, and reply blocks.
   Evidence: “recorded history as alternating fixed-labeled blocks ... with a single blank line between blocks and a trailing newline at the end of the input”
2. [codex-xhigh/completeness] Non-zero process exit behavior and stderr handling are not fully specified.
   Evidence: “reads stdout to completion, and on a zero exit returns the reply as transport.Result” and “a non-zero process exit is terminal (a usage/refusal result)”
3. [codex-xhigh/testability] The acceptance criterion that internal/engine is not modified is explicitly left to human review rather than a runnable validation command.
   Evidence: “internal/engine is not modified — enforced by human review and the engine dep-guard ... not by a git-diff command”

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
  "base_version": "45eb82e67dabbfeb7d150e12dadffff4dabc46287d2019dbf3ce8a098edc8657",
  "contract": {
    "acceptance_criteria": ["...updated criteria..."],
    "validation": {"commands": ["...updated commands..."]}
  }
}
```

Omit any contract field you are not changing. Do not include the goal field.
