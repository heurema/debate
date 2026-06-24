# Contract Draft Proposal

## Status
- Run id: run_20260624_103219
- Status: accepted
- Source: drafter_attempt
- Drafter attempt: drafter_attempt_001
- Drafter: codex
- Accepted by: manual
- Accepted at: 2026-06-24T10:35:48Z

## In scope
- Add internal/engine/transport/acp implementing transport.Transport over ACP subprocess stdio or injected streams.
- Implement ACP Open handshake: spawn backend command, initialize, session/new, and retain the persistent subprocess plus session id.
- Implement Session.Send as session/prompt and return the agent final text as transport.Result.
- Implement Session.Close to terminate the ACP subprocess/session without leaking processes.
- Implement grounded mode policy for ACP sessions: normal mode runs from the debate work directory with read/web access and no filesystem mutation; Spec.ReadOnly sealed mode disables project/web tools and uses brief-only prompting.
- Implement retryable dropped-session recovery by reopening the ACP subprocess and replaying prior successful prompts before retrying the current prompt once.
- Add deterministic fake ACP peer tests for handshake, persistent send reuse, close behavior, error classification, sealed policy, and recovery replay.
- Wire cmd/debate defaultResolver so claude-agent-acp and codex-acp return ACP transports while echo remains available offline.

## Out of scope
- Do not implement the exec/agy backend.
- Do not implement API-backed transports.
- Do not change the .heurema/debate config or persona schema.
- Do not add filesystem secret filtering, path allow/deny lists, or network allowlists beyond the ACP read-only/sealed policy required for this slice.
- Do not make default unit tests require network, API credentials, or installed real ACP CLIs.

## Acceptance criteria
- defaultResolver("echo") still returns the echo transport; defaultResolver("claude-agent-acp") and defaultResolver("codex-acp") return usable ACP transports; defaultResolver("agy") remains an unimplemented-backend error.
- A fake ACP peer observes Open sending initialize followed by session/new exactly once per session, with the configured model, effort, system prompt, cwd, and sandbox/tool policy.
- Multiple Session.Send calls on one session use the same ACP session id and do not repeat initialize or session/new.
- Session.Send returns the final ACP response text in transport.Result.Content and preserves zero-value usage when the peer provides no usage counters.
- Session.Close shuts down the subprocess/session and repeated Close calls do not panic or leave a running child process.
- ACP protocol errors, subprocess exits, context cancellation, deadlines, auth/rate-limit/server errors, and malformed responses map to transport sentinel errors so transport.Classify reports the expected retryable/non-retryable class.
- When a retryable transport drop occurs after prior successful sends, the ACP transport opens a new session, replays prior prompts, retries the current prompt once, and returns the retried final text.
- When Spec.ReadOnly is true, tests verify the ACP launch/session policy disables project and web tool access; when false, tests verify read/web access is enabled while write/mutation access is disabled.
- The default go test ./... path uses only fake ACP peers and does not require network access, API keys, or real Claude/Codex ACP binaries.
- A real-CLI ACP integration test exists behind an acp_integration build tag and an explicit environment variable gate, and skips with a clear message when prerequisites are absent.

## Validation commands
- bash scripts/check-gofmt.sh
- go vet ./...
- go test ./...
- go test ./internal/engine/transport/acp ./cmd/debate
- go test -tags=acp_integration ./internal/engine/transport/acp
- go build ./cmd/debate

## Assumptions
- Real ACP CLIs are invoked as claude-agent-acp and codex-acp unless transport.Spec.Command supplies an override.
- Spec.ReadOnly is interpreted for ACP as sealed brief-only mode; normal ACP mode still uses a read-only sandbox with network/web access.
- If no suitable ACP Go SDK is already available, a minimal in-repository ACP client covering initialize, session/new, session/prompt, and shutdown is acceptable.
- Recovery replay may repeat prior prompts against the backend after a retryable drop.
- Real integration validation requires installed and authenticated ACP CLIs and is optional unless the integration environment variable is set.

