# Contract Review Fixer Prompt

You are fixing a software change contract to address blocking review findings.

Current contract version: 36ad11a5f3757dda5c79273fb70313c568665e0e5376adff531317f8f3a0e5eb

## Current Contract

**Goal**: Implement an ACP backend transport under internal/backend/acp using github.com/coder/acp-go-sdk and wire it into the cmd/debate production resolver, so debate runs real debates with Claude (claude-agent-acp) and Codex (codex-acp), while internal/engine stays dependency-free.

**Scope in**:
  - Implement internal/backend/acp: a transport.Transport over github.com/coder/acp-go-sdk that spawns the per-backend ACP adapter subprocess (claude-agent-acp, codex-acp) and runs a persistent ACP session.
  - Implement the ACP session lifecycle: Open (spawn adapter, Initialize, NewSession with Cwd), Send (Prompt, accumulate streamed agent text until end_turn), Close (terminate the subprocess/process group).
  - Implement per-backend adapter command, model/effort wiring, and grounding (cwd visibility plus always-read-only filesystem plus network; brief-only when sealed).
  - Implement error classification via a package-local classify function in internal/backend/acp (no changes to internal/engine) and recovery (reopen and replay once on a retryable session drop).
  - Add deterministic unit tests using a fake in-process ACP peer, a subprocess-lifecycle unit test using a benign subprocess to verify process-group isolation and orphan prevention without any real adapter, and a real-CLI integration test gated behind a build tag and env var that the gate compiles but does not really run.
  - Wire the cmd/debate production resolver to register the acp backend for claude-agent-acp and codex-acp, update cmd/debate tests to match the expanded backend set, add github.com/coder/acp-go-sdk to go.mod, and add a dep-guard invocation for internal/backend.

**Scope out**:
  - The exec/agy backend and the api backend (later slices).
  - Any change to internal/engine source, the internal/debate policy/config packages, or the echo/mock backends beyond what wiring the resolver requires.
  - debate init/new scaffolding.
  - Making the default test run perform any real network, subprocess adapter, or model call.

**Acceptance criteria**:
  - internal/backend/acp provides a constructor returning a transport.Transport for a given backend id (claude-agent-acp or codex-acp) implemented with github.com/coder/acp-go-sdk; it never imports internal/debate or cmd/debate.
  - Open(ctx, spec) spawns the adapter subprocess for spec's backend in its own process group (setpgid so the child and its descendants share a new pgid distinct from the parent), creates an acp client-side connection over the subprocess stdio, calls Initialize then NewSession with the resolved Cwd, and returns a transport.Session holding the subprocess, connection, and session id; a spawn or handshake failure returns a classified error, kills the process group if it was already started, and leaves no orphaned subprocess.
  - The adapter command and model/effort wiring per backend is: claude-agent-acp -> `npx -y <pkg>` where `<pkg>` defaults to `@agentclientprotocol/claude-agent-acp@latest` and is overridable by setting env var DEBATE_CLAUDE_ACP_PKG, with additional subprocess environment ANTHROPIC_MODEL=spec.Model and CLAUDE_CODE_EFFORT_LEVEL=spec.Effort; codex-acp -> `npx -y <pkg> -c model="spec.Model"` where `<pkg>` defaults to `@heurema/codex-acp@latest` and is overridable by setting env var DEBATE_CODEX_ACP_PKG, and when read-only an additional `-c sandbox_mode="read-only"` flag is appended. Each override env var, when set, replaces the entire package argument passed to npx (e.g. `@agentclientprotocol/claude-agent-acp@1.2.3` to pin a version or a local path for testing). spec.Model must be non-empty (a fail-fast error otherwise); the adapter npm package per backend is overridable via DEBATE_CLAUDE_ACP_PKG and DEBATE_CODEX_ACP_PKG.
  - Filesystem access is always read-only (the agent never writes): for codex enforced via `-c sandbox_mode="read-only"`; for claude-agent-acp, which exposes no read-only adapter flag, the transport adds no flag and documents write-prevention as governed by the adapter permission layer (best-effort for claude in this slice).
  - Grounding (project visibility) is controlled by spec.ReadOnly (set by --sealed): when false (default, grounded) NewSession Cwd is the process working directory so the agent can read project files and use the network; when true (sealed/brief-only) NewSession Cwd is a fresh empty temporary directory so the agent sees no project files via relative paths. Tests assert the chosen Cwd and the per-backend adapter command/env/flags for grounded vs sealed and for claude vs codex.
  - Session.Send(ctx, prompt) sends an acp Prompt request with the prompt text, accumulates the streamed AgentMessageChunk text through the session-update handler, and returns transport.Result with the accumulated content when the turn completes with stop reason end_turn.
  - A turn that completes with a refusal or any non-end_turn stop reason, or that errors mid-stream, returns a classified error rather than a partial success.
  - The session is persistent: the subprocess and acp session are created once in Open and reused across multiple Send calls; Close terminates the subprocess and its process group and is idempotent (safe to call twice).
  - A package-local classify function in internal/backend/acp maps acp and subprocess errors to retryable vs terminal (dropped connection / broken pipe / idle timeout -> retryable; refusal / auth / protocol error -> terminal); it does not add or change any symbol in internal/engine. On a retryable failure during Send the transport recovers exactly once: it closes the broken session, reopens a new session (re-running Initialize/NewSession), replays this session's prior prompts in order followed by the current prompt, and returns the resulting content; a second consecutive retryable failure or any terminal error is returned to the caller, and a drop after end_turn is not treated as an error.
  - Deterministic unit tests use a fake in-process ACP peer (the acp-go-sdk agent-side connection or an equivalent stub) wired to the client over in-memory pipes with no real subprocess or network, covering: the Initialize/NewSession handshake, a successful Send returning accumulated text on end_turn, persistence across multiple Sends on one session, a refusal/non-end_turn error, classify behavior, the grounded-vs-sealed Cwd choice, and recovery after a simulated retryable drop.
  - A subprocess-lifecycle unit test (no build tag; included in the default `go test ./...` run) verifies process-group isolation and orphan prevention using a benign subprocess (e.g. `/bin/cat`) with no ACP handshake: it asserts the spawned process runs in a new process group (pgid differs from the test process's pgid), and that Close kills the process group such that the subprocess is no longer alive afterward; this test requires no network, npm adapter, or model call.
  - A real-CLI integration test is gated behind a build tag (//go:build acp_integration) and an environment variable: with the tag set but the env var unset the test compiles and runs but skips itself, so it performs no network/subprocess/model call; the gate includes a command that compiles it under the tag.
  - The cmd/debate production resolver registers the acp backend for the claude-agent-acp and codex-acp backend ids and keeps the echo backend for offline use; an unknown or still-unimplemented backend (agy, api) fails fast with a clear error. The cmd/debate test suite is updated to reflect the expanded backend set: any existing test that expected claude-agent-acp or codex-acp to be unimplemented is removed or replaced with tests asserting that echo, claude-agent-acp, and codex-acp resolve to non-nil backends without invoking them, and that agy and api return a fast error; `go test -count=1 ./...` passes with no real backend invoked.
  - go.mod and go.sum gain github.com/coder/acp-go-sdk; internal/engine stays stdlib + internal/engine only and unchanged; a dep-guard invocation enforces that internal/backend depends only on the Go standard library, internal/engine, internal/backend, and github.com/coder/acp-go-sdk, and internal/debate must not import internal/backend or the acp sdk; check-gofmt, go vet ./..., go build ./..., and go test ./... pass with no real backend invoked.

**Validation commands**:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - go test -count=1 -tags acp_integration ./internal/backend/...
  - go vet ./...
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

**Assumptions**:
  - The ACP client uses github.com/coder/acp-go-sdk (the library pactum itself uses), at a version compatible with the installed adapters.
  - The ACP adapters are external npm packages launched via npx; their real availability is a runtime concern exercised only by the gated integration test.
  - Filesystem write-prevention is enforced at the adapter level (codex via sandbox_mode; claude via its permission layer, best-effort); spec.ReadOnly instead selects grounded vs brief-only by choosing the NewSession Cwd.
  - Sealed mode (spec.ReadOnly=true) limits grounding by setting the ACP session Cwd to a fresh empty temporary directory, preventing adapter processes from implicitly accessing project files via relative paths. Absolute-path access and other ambient environment state (e.g. env vars inherited by the subprocess, parent directory traversal) are not filtered at the transport layer; sealed is a best-effort grounding boundary for v1, not a security isolation guarantee. Unit tests assert the Cwd choice but do not verify what the adapter process can or cannot read at runtime.
  - The fake ACP peer implements just enough of the agent side (Initialize, NewSession, Prompt with streamed agent-message chunks and a stop reason) to drive the client deterministically, and is the default test path.
  - Recovery replays this session's full prior conversation once on a retryable drop; a second consecutive failure surfaces as an error.
  - Real model backends live under internal/backend to keep internal/engine dependency-free; the trivial echo and mock backends remain in internal/engine/transport as stdlib-only fixtures.
  - The runner stays backend-agnostic: it depends only on the transport interface and the injected Resolver, never on internal/backend.

## Blocking Findings to Address

1. [codex-xhigh/completeness] The contract does not completely specify how Codex is kept read-only in grounded mode.
   Evidence: Acceptance criteria say: "Filesystem access is always read-only" and "for codex enforced via `-c sandbox_mode=\"read-only\"`", but the adapter command criterion says the Codex read-only flag is appended only "when read-only". The assumptions also define `spec.ReadOnly=true` as sealed mode rather than filesystem-read-only mode.
2. [codex-xhigh/assumptions-surfaced] The contract assumes a POSIX runtime for process-group lifecycle behavior, but does not surface that platform assumption.
   Evidence: Acceptance requires `setpgid`, killing the process group, and a lifecycle test using a benign subprocess such as `/bin/cat`, while validation requires default `go test -count=1 ./...`.

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
  "base_version": "36ad11a5f3757dda5c79273fb70313c568665e0e5376adff531317f8f3a0e5eb",
  "contract": {
    "acceptance_criteria": ["...updated criteria..."],
    "validation": {"commands": ["...updated commands..."]}
  }
}
```

Omit any contract field you are not changing. Do not include the goal field.
