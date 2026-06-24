# Reviewer Context

## Run
- Run id: run_20260624_103219
- Run status: contract_approved

## Contract
- Goal: Implement an ACP backend transport under internal/backend/acp using github.com/coder/acp-go-sdk and wire it into the cmd/debate production resolver, so debate runs real debates with Claude (claude-agent-acp) and Codex (codex-acp), while internal/engine stays dependency-free.
- In scope:
  - Implement internal/backend/acp: a transport.Transport over github.com/coder/acp-go-sdk that spawns the per-backend ACP adapter subprocess (claude-agent-acp, codex-acp) and runs a persistent ACP session.
  - Implement the ACP session lifecycle: Open (spawn adapter, Initialize, NewSession), Send (Prompt, accumulate streamed agent text until end_turn), Close (terminate subprocess/process group), behind an injectable process runner.
  - Implement per-backend adapter command, model/effort wiring, package overrides, always-read-only filesystem (where the adapter supports it), and grounded-vs-sealed Cwd selection plus network.
  - Map ACP/subprocess failures to retryable vs terminal consistent with transport.Classify (without changing internal/engine) and recover once by reopening and replaying on a retryable drop.
  - Add deterministic unit tests via a fake in-process ACP peer and an injectable process runner; add a real-CLI integration test gated behind a build tag and env var that the gate compiles but skips.
  - Wire the cmd/debate production resolver for claude-agent-acp and codex-acp, update the Slice-4 unimplemented-backend test, add github.com/coder/acp-go-sdk to go.mod, and add a dep-guard invocation for internal/backend.
- Out of scope:
  - The exec/agy backend and the api backend (later slices).
  - Any change to internal/engine source (including transport.Classify), the internal/debate policy/config packages, or the echo/mock backends beyond what wiring the resolver requires.
  - debate init/new scaffolding.
  - Making the default test run perform any real network, subprocess, or model call; and Windows/non-Unix process-group support.
- Acceptance criteria:
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
- Validation commands:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - env -u DEBATE_ACP_INTEGRATION go test -count=1 -tags acp_integration ./internal/backend/...
  - go vet ./...
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

## Accepted memory
- Memory context: context/memory-context.md
- Selected items: 0
- Fresh: 0
- Stale: 0
- Unknown: 0
- Stale memory may be outdated and must be verified.

## Gate report
- Gate status: needs_review
- Execution attempt id: attempt_001
- Execution exit code: 0
- Validation command results:
  - command_001: bash scripts/check-gofmt.sh (exit 0, timed out: false, result: gate/validation/command_001/result.json)
  - command_002: go build ./... (exit 0, timed out: false, result: gate/validation/command_002/result.json)
  - command_003: go test -count=1 ./... (exit 0, timed out: false, result: gate/validation/command_003/result.json)
  - command_004: env -u DEBATE_ACP_INTEGRATION go test -count=1 -tags acp_integration ./internal/backend/... (exit 0, timed out: false, result: gate/validation/command_004/result.json)
  - command_005: go vet ./... (exit 0, timed out: false, result: gate/validation/command_005/result.json)
  - command_006: bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine (exit 0, timed out: false, result: gate/validation/command_006/result.json)
  - command_007: bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk (exit 0, timed out: false, result: gate/validation/command_007/result.json)
  - command_008: bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3 (exit 0, timed out: false, result: gate/validation/command_008/result.json)
- Change summary:
  - changed files:
    - cmd/debate/e2e_test.go
    - cmd/debate/main.go
    - go.mod
    - go.sum
  - new files:
    - internal/backend/acp/acp.go
    - internal/backend/acp/acp_test.go
    - internal/backend/acp/integration_test.go
  - missing files:
    - none

## Existing manual review
- Review status: changes_requested
- Current findings summary: findings=12 open=12 resolved=0 blocking_open=1
- Existing findings:
  - f_001 severity=medium category=correctness blocking=false status=open: defaultProcessRunner starts the adapter subprocess with cmd.Start() and kills its process group on Close/handshake-failure, but cmd.Wait() is never called anywhere. A SIGKILL'd child that is never waited on becomes a defunct/zombie process that persists until the parent process exits. For the short-lived debate CLI this is reaped at process exit, but looper is intended as a reusable discussion-loop library/runtime; a long-running host opens and closes many sessions (and even failed Opens kill in openAt) and would accumulate zombies, eventually risking process-table exhaustion. This production path is exercised only by the gated (skipped) integration test, so the default suite does not catch it.
  - f_002 severity=medium category=correctness blocking=false status=open: The ACP transport silently drops spec.System. The runner populates Spec.System from each persona's system prompt (internal/debate/runner/runner.go:87 and :146), and the prompt builder (internal/debate/prompt/prompt.go) does not embed the persona system prompt into the Send content — it renders only moderator rules, the shared brief, the board, and the signal instruction. The ACP transport never reads spec.System in buildCmd, NewSession, or Send, so when a real debate runs against claude-agent-acp/codex-acp every persona receives identical context with no persona identity/role, collapsing the distinct-persona behavior the debate depends on.
  - f_003 severity=low category=correctness blocking=false status=open: Sealed-mode temporary directory is leaked. resolveCwd(sealed=true) creates a fresh temp dir via os.MkdirTemp("", "debate-sealed-*") and stores it in acpSession.cwd, but Close() only kills the subprocess and nothing ever calls os.RemoveAll on that directory. Each sealed session therefore leaves an empty debate-sealed-* directory in the temp dir that persists after the process exits, accumulating across runs.
  - f_004 severity=low category=correctness blocking=false status=open: classifyConnErr maps every JSON-RPC -32603 (Internal error) to ErrTransportDrop (retryable). Because agent-side handler errors are commonly surfaced as -32603, a terminal/permanent agent failure reported as a generic handler error is classified retryable and triggers an unnecessary recovery: close, reopen, and full prior-history replay, before the same condition re-fails and the error finally surfaces. The contract's intended mapping puts protocol errors in the terminal bucket; this widens the retryable bucket to all internal errors.
  - f_005 severity=high category=correctness blocking=true status=open: The ACP transport silently drops spec.System (the persona's system-prompt body). buildCmd consumes spec.Model and spec.Effort, openAt's NewSessionRequest sets only Cwd/McpServers, and sendOnce's PromptRequest sends only the bare per-turn prompt; spec.System is never referenced. The runner sets System=p.System for every participant and the synthesizer, and the orchestrate PromptBuilder does not include persona identity, so a real claude-agent-acp/codex-acp debate gives every participant identical moderator+brief prompts with no persona instructions, undermining the contract goal of differentiated multi-persona debate.
  - f_006 severity=low category=validation blocking=false status=open: No deterministic test asserts the process-group setting. Acceptance criterion #9 states the injectable runner exists 'so deterministic tests assert the process-group setting and that Close performs cleanup', but Setpgid:true is set only inside defaultProcessRunner, which the fake-runner tests never invoke, and the ProcessRunner signature (dir,name,args,env) does not surface SysProcAttr. Close cleanup is asserted via killCount, but the process-group setting is unasserted and unassertable through the current runner abstraction.
  - f_007 severity=medium category=validation blocking=false status=open: Open's spawn and ACP-handshake failure paths are untested. openAt handles three failure branches — spawn error -> ErrClientError, Initialize error -> kill()+classified error, NewSession error -> kill()+classified error — but the fake agent always succeeds at Initialize and NewSession and the fake runner never returns an error. The acceptance criterion 'a spawn or handshake failure returns a classified error and leaves no orphaned subprocess' (the kill() cleanup on failure) has no deterministic coverage and could regress silently.
  - f_008 severity=medium category=quality blocking=false status=open: Recovery tests do not verify that prior prompts are replayed in order or with the correct content. fakeAgent.Prompt discards the incoming p.Prompt and responds purely by positional scenario index, so TestSend_RecoveryWithHistoryReplay asserts only the final response (r3). A recovery that replayed wrong prompt text, replayed empty strings, or reordered the history would still pass. The contract explicitly requires replaying prior prompts in order.
  - f_009 severity=low category=validation blocking=false status=open: clientImpl callback methods have no deterministic test. The fake agent only calls asc.SessionUpdate and never invokes the client side, so RequestPermission's branching (no options -> cancelled vs. select-first-option) and the read-only WriteTextFile denial are never exercised in the default suite.
  - f_010 severity=low category=validation blocking=false status=open: classifyConnErr and stopReasonErr branch coverage is thin. Only the InternalError(-32603)->ErrTransportDrop path is exercised. The auth(-32000)->ErrAuth, cancelled(-32800)->ErrCanceled, io.EOF/broken-pipe->ErrTransportDrop, and default->ErrClientError branches, plus stopReasonErr's Cancelled case, are untested, even though the contract lists auth->terminal as a required classification.
  - f_011 severity=low category=quality blocking=false status=open: outcomeString is a no-op pass-through: every switch branch, including the default, returns the input `reason` unchanged. The function adds no normalization or transformation and could be replaced by using result.Outcome.Reason directly at the call site.
  - f_012 severity=medium category=quality blocking=false status=open: The contract acceptance criterion requires documenting claude-agent-acp write-prevention as best-effort governed by the adapter permission layer, but the code documents it as an absolute guarantee. The WriteTextFile comment states 'the agent must not write files' and its error says 'transport is read-only', while RequestPermission allows the first permission option for any request and the claude branch of buildCmd adds no read-only flag — so the agent can still write via its own tools. The required best-effort caveat is absent and the existing comment is misleading.
- Existing resolutions:
  - none
- Proposal summary: pending=0 accepted=12 rejected=0
- Existing proposals:
  - p_001 severity=medium category=correctness blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_001: defaultProcessRunner starts the adapter subprocess with cmd.Start() and kills its process group on Close/handshake-failure, but cmd.Wait() is never called anywhere. A SIGKILL'd child that is never waited on becomes a defunct/zombie process that persists until the parent process exits. For the short-lived debate CLI this is reaped at process exit, but looper is intended as a reusable discussion-loop library/runtime; a long-running host opens and closes many sessions (and even failed Opens kill in openAt) and would accumulate zombies, eventually risking process-table exhaustion. This production path is exercised only by the gated (skipped) integration test, so the default suite does not catch it.
    location: internal/backend/acp/acp.go:413
  - p_002 severity=medium category=correctness blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_001: The ACP transport silently drops spec.System. The runner populates Spec.System from each persona's system prompt (internal/debate/runner/runner.go:87 and :146), and the prompt builder (internal/debate/prompt/prompt.go) does not embed the persona system prompt into the Send content — it renders only moderator rules, the shared brief, the board, and the signal instruction. The ACP transport never reads spec.System in buildCmd, NewSession, or Send, so when a real debate runs against claude-agent-acp/codex-acp every persona receives identical context with no persona identity/role, collapsing the distinct-persona behavior the debate depends on.
    location: internal/backend/acp/acp.go:122
  - p_003 severity=low category=correctness blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_001: Sealed-mode temporary directory is leaked. resolveCwd(sealed=true) creates a fresh temp dir via os.MkdirTemp("", "debate-sealed-*") and stores it in acpSession.cwd, but Close() only kills the subprocess and nothing ever calls os.RemoveAll on that directory. Each sealed session therefore leaves an empty debate-sealed-* directory in the temp dir that persists after the process exits, accumulating across runs.
    location: internal/backend/acp/acp.go:154
  - p_004 severity=low category=correctness blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_001: classifyConnErr maps every JSON-RPC -32603 (Internal error) to ErrTransportDrop (retryable). Because agent-side handler errors are commonly surfaced as -32603, a terminal/permanent agent failure reported as a generic handler error is classified retryable and triggers an unnecessary recovery: close, reopen, and full prior-history replay, before the same condition re-fails and the error finally surfaces. The contract's intended mapping puts protocol errors in the terminal bucket; this widens the retryable bucket to all internal errors.
    location: internal/backend/acp/acp.go:364
  - p_005 severity=high category=correctness blocking=true status=accepted source=reviewer_attempt attempt=reviewer_attempt_004: The ACP transport silently drops spec.System (the persona's system-prompt body). buildCmd consumes spec.Model and spec.Effort, openAt's NewSessionRequest sets only Cwd/McpServers, and sendOnce's PromptRequest sends only the bare per-turn prompt; spec.System is never referenced. The runner sets System=p.System for every participant and the synthesizer, and the orchestrate PromptBuilder does not include persona identity, so a real claude-agent-acp/codex-acp debate gives every participant identical moderator+brief prompts with no persona instructions, undermining the contract goal of differentiated multi-persona debate.
    location: internal/backend/acp/acp.go:208
  - p_006 severity=low category=validation blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_004: No deterministic test asserts the process-group setting. Acceptance criterion #9 states the injectable runner exists 'so deterministic tests assert the process-group setting and that Close performs cleanup', but Setpgid:true is set only inside defaultProcessRunner, which the fake-runner tests never invoke, and the ProcessRunner signature (dir,name,args,env) does not surface SysProcAttr. Close cleanup is asserted via killCount, but the process-group setting is unasserted and unassertable through the current runner abstraction.
    location: internal/backend/acp/acp.go:402
  - p_007 severity=medium category=validation blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_002: Open's spawn and ACP-handshake failure paths are untested. openAt handles three failure branches — spawn error -> ErrClientError, Initialize error -> kill()+classified error, NewSession error -> kill()+classified error — but the fake agent always succeeds at Initialize and NewSession and the fake runner never returns an error. The acceptance criterion 'a spawn or handshake failure returns a classified error and leaves no orphaned subprocess' (the kill() cleanup on failure) has no deterministic coverage and could regress silently.
    location: internal/backend/acp/acp.go:92
  - p_008 severity=medium category=quality blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_002: Recovery tests do not verify that prior prompts are replayed in order or with the correct content. fakeAgent.Prompt discards the incoming p.Prompt and responds purely by positional scenario index, so TestSend_RecoveryWithHistoryReplay asserts only the final response (r3). A recovery that replayed wrong prompt text, replayed empty strings, or reordered the history would still pass. The contract explicitly requires replaying prior prompts in order.
    location: internal/backend/acp/acp_test.go:396
  - p_009 severity=low category=validation blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_002: clientImpl callback methods have no deterministic test. The fake agent only calls asc.SessionUpdate and never invokes the client side, so RequestPermission's branching (no options -> cancelled vs. select-first-option) and the read-only WriteTextFile denial are never exercised in the default suite.
    location: internal/backend/acp/acp.go:316
  - p_010 severity=low category=validation blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_002: classifyConnErr and stopReasonErr branch coverage is thin. Only the InternalError(-32603)->ErrTransportDrop path is exercised. The auth(-32000)->ErrAuth, cancelled(-32800)->ErrCanceled, io.EOF/broken-pipe->ErrTransportDrop, and default->ErrClientError branches, plus stopReasonErr's Cancelled case, are untested, even though the contract lists auth->terminal as a required classification.
    location: internal/backend/acp/acp.go:357
  - p_011 severity=low category=quality blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_005: outcomeString is a no-op pass-through: every switch branch, including the default, returns the input `reason` unchanged. The function adds no normalization or transformation and could be replaced by using result.Outcome.Reason directly at the call site.
    location: cmd/debate/main.go:188
  - p_012 severity=medium category=quality blocking=false status=accepted source=reviewer_attempt attempt=reviewer_attempt_003: The contract acceptance criterion requires documenting claude-agent-acp write-prevention as best-effort governed by the adapter permission layer, but the code documents it as an absolute guarantee. The WriteTextFile comment states 'the agent must not write files' and its error says 'transport is read-only', while RequestPermission allows the first permission option for any request and the claude branch of buildCmd adds no read-only flag — so the agent can still write via its own tools. The required best-effort caveat is absent and the existing comment is misleading.
    location: internal/backend/acp/acp.go:311

## Artifacts
- Contract: contract/contract.json
- Gate report: gate/gate-report.json
- Review: review/review.json
- Findings: review/findings.jsonl
- Resolutions: review/resolutions.jsonl
- Proposals: review/proposals.jsonl
- Proposal decisions: review/proposal-decisions.jsonl
- Execution result: execute/last-result.json

## Reviewer guidance
- This context is not complete semantic truth.
- Use `pactum search "<term>"` and inspect files before proposing findings.
- Do not invent changes.
- Do not approve automatically.
- Report every issue you believe is likely real: use state=candidate for uncertain findings and drop only when trigger, evidence, and fix_direction cannot be filled concretely.
