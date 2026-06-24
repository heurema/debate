# Contract Draft

## Goal
Implement an ACP backend transport under internal/backend/acp using github.com/coder/acp-go-sdk and wire it into the cmd/debate production resolver, so debate runs real debates with Claude (claude-agent-acp) and Codex (codex-acp), while internal/engine stays dependency-free.

## Current status
Contract status: approved
Manual clarification, contract approval, prompt build, and agent execution are available through staged Pactum commands.

## Relevant repository context
- Map run: map_20260624_100801
- Repo map: .heurema/pactum/map/repo-map.md
- Search results: context/search-results.json (0 result(s))

## Clarifications
- None

## In scope
- Implement internal/backend/acp: a transport.Transport over github.com/coder/acp-go-sdk that spawns the per-backend ACP adapter subprocess (claude-agent-acp, codex-acp) and runs a persistent ACP session.
- Implement the ACP session lifecycle: Open (spawn adapter, Initialize, NewSession), Send (Prompt, accumulate streamed agent text until end_turn), Close (terminate subprocess/process group), behind an injectable process runner.
- Implement per-backend adapter command, model/effort wiring, package overrides, always-read-only filesystem (where the adapter supports it), and grounded-vs-sealed Cwd selection plus network.
- Map ACP/subprocess failures to retryable vs terminal consistent with transport.Classify (without changing internal/engine) and recover once by reopening and replaying on a retryable drop.
- Add deterministic unit tests via a fake in-process ACP peer and an injectable process runner; add a real-CLI integration test gated behind a build tag and env var that the gate compiles but skips.
- Wire the cmd/debate production resolver for claude-agent-acp and codex-acp, update the Slice-4 unimplemented-backend test, add github.com/coder/acp-go-sdk to go.mod, and add a dep-guard invocation for internal/backend.

## Out of scope
- The exec/agy backend and the api backend (later slices).
- Any change to internal/engine source (including transport.Classify), the internal/debate policy/config packages, or the echo/mock backends beyond what wiring the resolver requires.
- debate init/new scaffolding.
- Making the default test run perform any real network, subprocess, or model call; and Windows/non-Unix process-group support.

## Acceptance criteria
- internal/backend/acp provides a constructor returning a transport.Transport for a given backend id (claude-agent-acp or codex-acp) implemented with github.com/coder/acp-go-sdk; it never imports internal/debate or cmd/debate.
- Open(ctx, spec) spawns the adapter subprocess (via an injectable process runner) in its own process group with the subprocess working directory set to the resolved Cwd, creates an acp client-side connection over the subprocess stdio, calls Initialize then NewSession with the same resolved Cwd, and returns a transport.Session holding the subprocess, connection, and session id; a spawn or handshake failure returns a classified error and leaves no orphaned subprocess.
- Adapter command and model/effort wiring per backend: claude-agent-acp -> `npx -y <pkg>` (default pkg @agentclientprotocol/claude-agent-acp@latest, overridable via env DEBATE_CLAUDE_AGENT_ACP_PACKAGE) with environment ANTHROPIC_MODEL=spec.Model and CLAUDE_CODE_EFFORT_LEVEL=spec.Effort; codex-acp -> `npx -y <pkg> -c model="spec.Model" -c sandbox_mode="read-only"` (default pkg @heurema/codex-acp@latest, overridable via env DEBATE_CODEX_ACP_PACKAGE), where codex effort is intentionally not wired (codex-acp exposes no effort knob) and spec.Effort is ignored for codex. spec.Model must be non-empty (fail-fast otherwise). The override env var replaces only the npm package token and preserves `npx -y` and the backend flags; tests assert default and overridden commands for both backends.
- Filesystem access is always read-only where the adapter supports it (the agent never writes), independent of grounding: codex always receives `-c sandbox_mode="read-only"` in both grounded and sealed modes; claude-agent-acp exposes no read-only adapter flag, so the transport adds none and documents claude write-prevention as best-effort governed by the adapter permission layer.
- Grounding (project visibility) is selected by spec.ReadOnly (set by --sealed): grounded (false, default) sets both NewSession Cwd and the subprocess working directory to the process working directory so the agent reads project files; sealed/brief-only (true) sets both to a fresh empty temporary directory so the agent sees no project files. Network access is available in both modes (sealed does not disable network). Tests assert the Cwd, the subprocess working directory, and the per-backend command/env/flags for grounded vs sealed and claude vs codex.
- Session.Send(ctx, prompt) sends an acp Prompt request, accumulates the streamed AgentMessageChunk text via the session-update handler, and returns transport.Result with the accumulated content when the turn completes with stop reason end_turn.
- A turn that completes with a refusal or any non-end_turn stop reason, or that errors mid-stream, returns a classified error rather than a partial success.
- The session is persistent: the subprocess and acp session are created once in Open and reused across multiple Send calls; Close terminates the subprocess and its process group and is idempotent (safe to call twice).
- The acp backend determines retryable vs terminal consistent with transport.Classify's category definitions without modifying internal/engine (dropped connection / broken pipe / idle timeout -> retryable; refusal / auth / protocol error -> terminal). On a retryable failure during Send the transport recovers exactly once: it closes the broken session, reopens (Initialize/NewSession), replays this session's prior prompts in order (each must complete with end_turn and its returned text is discarded; any replay failure aborts recovery with a classified error), then re-sends the current prompt and returns its result. A second consecutive retryable failure or any terminal error is returned to the caller, and a drop after end_turn is not treated as an error.
- The subprocess spawn is behind an injectable process/command runner so deterministic tests assert the process-group setting and that Close performs cleanup, without launching a real subprocess.
- Deterministic unit tests use a fake in-process ACP peer (the acp-go-sdk agent-side connection or an equivalent stub) over in-memory pipes with no real subprocess or network, covering: the handshake, a successful Send accumulating text on end_turn, persistence across multiple Sends, a refusal/non-end_turn error, retryable-vs-terminal classification, the grounded-vs-sealed Cwd and subprocess working directory, the per-backend command/env/flags (default and overridden), and recovery after a simulated retryable drop including a replay-failure abort.
- A real-CLI integration test is gated behind build tag //go:build acp_integration AND env var DEBATE_ACP_INTEGRATION: with the tag set but DEBATE_ACP_INTEGRATION unset it compiles and skips (no real network/subprocess/model call); the gate includes a command that compiles it under the tag with the env var cleared.
- The cmd/debate production resolver registers the acp backend for claude-agent-acp and codex-acp and keeps echo for offline use; the Slice-4 unimplemented-backend e2e fixture is updated to use a still-unimplemented backend (agy/api/unknown) for its fail-fast assertion, and a test asserts that claude-agent-acp and codex-acp resolve to a transport without opening a real session in the default suite.
- go.mod and go.sum gain github.com/coder/acp-go-sdk; internal/engine stays stdlib + internal/engine only and unchanged; a dep-guard invocation enforces that internal/backend depends only on the Go standard library, internal/engine, internal/backend, and github.com/coder/acp-go-sdk, and internal/debate must not import internal/backend or the acp sdk; check-gofmt, go vet ./..., go build ./..., and go test ./... pass with no real backend invoked.

## Validation commands
- bash scripts/check-gofmt.sh
- go build ./...
- go test -count=1 ./...
- env -u DEBATE_ACP_INTEGRATION go test -count=1 -tags acp_integration ./internal/backend/...
- go vet ./...
- bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
- bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk
- bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

## Assumptions
- The ACP client uses github.com/coder/acp-go-sdk (the library pactum itself uses), at a version compatible with the installed adapters.
- The ACP adapters are external npm packages launched via npx; their real availability is a runtime concern exercised only by the gated integration test.
- Filesystem write-prevention is enforced at the adapter level (codex via sandbox_mode always; claude via its permission layer, best-effort); spec.ReadOnly instead selects grounded vs brief-only by choosing the NewSession Cwd and the subprocess working directory.
- Both adapters honor NewSession Cwd (and the matching subprocess working directory) for relative file/tool access; preventing absolute-path or project-root leakage beyond Cwd is out of scope and assumed adapter-governed.
- Sealed mode removes project-file visibility (empty Cwd and subprocess working directory) but does not disable network access.
- Recovery replays this session's full prior conversation once on a retryable drop; replayed prompts must complete with end_turn and a second consecutive failure surfaces as an error.
- Process-group lifecycle targets Unix-like platforms (darwin/linux); other GOOS support is out of scope.
- Real model backends live under internal/backend to keep internal/engine dependency-free; the trivial echo and mock backends remain in internal/engine/transport as stdlib-only fixtures; the runner stays backend-agnostic via the injected Resolver.
- The fake ACP peer implements just enough of the agent side (Initialize, NewSession, Prompt with streamed chunks and a stop reason) to drive the client deterministically, and is the default test path.

## Open questions
- None
