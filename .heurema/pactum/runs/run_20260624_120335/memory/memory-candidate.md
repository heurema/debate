# Memory Candidate

## Run
- Run id: run_20260624_120335
- Source: deterministic

## Contract
- Goal: Implement an exec backend transport under internal/backend/exec (standard library only) that drives stateless plain-CLI agents like Gemini via agy, reconstructing full conversation context per turn from accumulated prompts and replies, and wire it into the cmd/debate production resolver while internal/engine stays unchanged.
- In scope:
  - Implement internal/backend/exec: a transport.Transport that drives a stateless CLI agent by spawning a fresh subprocess per Send via exec.CommandContext, using only the Go standard library.
  - Implement the stateless full-render Session: record each Send prompt and the agent's reply (committed only on success), and on each Send reconstruct the full stdin (system folded in, the recorded prompt/reply history, and the new prompt) for a fresh subprocess.
  - Implement per-backend command derivation (agy) with an env override, spec.Model wiring, grounded-vs-sealed working directory, ctx cancellation, and an explicit error-classification mapping with a single retry.
  - Add deterministic unit tests via a fake CLI / injectable command runner, and a real-agy integration test gated behind a build tag and env var that the gate compiles but skips.
  - Wire the cmd/debate production resolver to register the exec backend for the agy backend, update the Slice-5 unimplemented-backend test, and add targeted exec dep-guard commands.
- Out of scope:
  - The api backend (a later concern).
  - Any change to internal/engine source (including transport.Classify and the orchestrate Delta render), the internal/debate policy/config packages, or the acp/echo/mock backends beyond what wiring the resolver requires.
  - debate init/new scaffolding.
  - Making the default test run perform any real agy, model, or network call (local stub subprocesses in tests are allowed); and Windows/non-Unix support.
- Acceptance criteria:
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
- Validation commands:
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

## Outcome
- Gate status: needs_review
- Review status: approved
- Execution exit code: 0
- Validation passed: true
- Changes need review: true

## Changes
- Changed files:
  - cmd/debate/e2e_test.go
  - cmd/debate/main.go
- New files:
  - internal/backend/exec/exec.go
  - internal/backend/exec/exec_test.go
  - internal/backend/exec/integration_test.go
- Missing files: none

## Clarifications
- None

## Review Decisions
- f_001 [medium] open internal/backend/exec/exec.go:190: In sendOnce the stdin broken-pipe branch is checked before the non-zero-exit branch, so when a subprocess exits non-zero before draining all of stdin the EPIPE write error wins: Send returns a retryable transport-drop error and discards the captured stderr, instead of the terminal client error carrying the stderr diagnostic the contract requires.
- f_002 [low] open internal/backend/exec/exec.go:267: formatStderr labels stderr as truncated whenever len(data) >= maxStderrBytes, but data is read through io.LimitReader(stderr, maxStderrBytes) and is therefore exactly maxStderrBytes whenever stderr has at least that many bytes — including the case where the real stderr is exactly 4096 bytes and nothing was dropped, producing a misleading '(truncated)' suffix.
- f_003 [medium] open internal/backend/exec/exec.go:190: In sendOnce the stdin broken-pipe check runs before the non-zero-exit check, so when a subprocess exits non-zero AND closes stdin before our write completes, the error is classified as a retryable transport drop (ErrTransportDrop) instead of a terminal client error (ErrClientError), and the captured stderr is dropped from the surfaced error. The contract's error-classification criterion requires a non-zero exit to be terminal with the stderr content included. This path is realistic for this backend specifically: because each Send reconstructs the full both-sided conversation context, stdin grows monotonically each round and can exceed the OS pipe buffer (~16KB on darwin, the documented target). If a real agy run rejects a request (bad model / refusal) and exits non-zero without draining oversized stdin, our write gets EPIPE; the code then returns a retryable ErrTransportDrop, retries once (same outcome), and surfaces a retryable error with no stderr — a deterministic terminal failure presented as transient and undiagnosable.
- f_004 [low] open internal/backend/exec/exec.go:88: The acceptance criterion states filesystem read-only 'is not enforced by the exec backend (a plain CLI has no sandbox) and is documented as trusted/best-effort.' No such documentation exists in the package: the package doc comment, resolveCwd, and the ReadOnly handling describe only grounded-vs-sealed working-directory selection and never state that read-only is unenforced / trusted / best-effort. A maintainer reading the code could assume sealed mode provides a read-only/network sandbox when it only changes the working directory.
- f_005 [low] open internal/backend/exec/exec.go:267: formatStderr appends '...(truncated)' whenever len(data) >= maxStderrBytes, but data is read through io.LimitReader(stderr, maxStderrBytes) so it is capped at exactly maxStderrBytes. When the subprocess emits exactly 4096 bytes of stderr with nothing beyond it, the message is falsely labeled truncated even though no content was dropped. The drained remainder (io.Copy to io.Discard) is not consulted to decide truncation.
- f_006 [low] open cmd/debate/e2e_test.go:174: Stale, contradictory comment in the unimplemented-backend e2e test. This change moved the fail-fast fixture to `backend: api` and registered `agy` as a resolvable backend, but the comment inside TestE2E_UnimplementedBackend_Exit1 still asserts the opposite ('defaultResolver returns error for unknown backend "agy"'). agy now resolves successfully via defaultResolver (cmd/debate/main.go:56-57), so the comment is factually wrong and is a maintenance trap: a maintainer trusting it could restore `agy` as the fixture backend, which would no longer fail fast and would silently break the test's intent.
- f_007 [low] open internal/backend/exec/exec.go:111: execSession.mu mutex is unnecessary and provides only illusory thread-safety. It is locked solely in Close to guard the `closed` idempotency flag, while Send mutates s.history and reads s.spec/s.cwd unlocked. The contract assumptions explicitly state the Session is serialized by orchestrate and not required to be goroutine-safe, so a plain bool check suffices for idempotent Close and the mutex protects none of the actually-mutated state.
- f_008 [low] open internal/backend/exec/exec.go:48: New's backendID parameter is validated then discarded, creating the appearance of multi-backend support that does not exist. It is checked against BackendAgy and rejected otherwise, but never stored on execTransport, and resolveCmd hardcodes the 'agy' executable. With exactly one supported backend the parameter carries no behavior.
- f_009 [medium] resolved internal/backend/exec/exec.go:84: The exec backend does not document that --sealed / spec.ReadOnly is best-effort: filesystem read-only is NOT enforced (a plain CLI has no sandbox) and network remains available in both grounded and sealed modes. The contract acceptance criterion explicitly requires this be 'documented as trusted/best-effort', but no such comment exists. This is misleading because docs/DESIGN.md frames 'grounded read-only' as a write-rejecting sandbox (true for the ACP backend), so a user passing --sealed for a gemini/agy persona may wrongly assume filesystem writes are prevented.
  Resolution: Added a note to the resolveCwd doc comment (exec.go:88) stating explicitly that filesystem read-only is NOT enforced (a plain CLI has no sandbox), that it is trusted/best-effort, and that network is available in both grounded and sealed modes.
- f_010 [medium] resolved internal/backend/exec/exec.go:190: In sendOnce the stdin broken-pipe branch (exec.go:190) is evaluated before the success return (exec.go:199) and is not gated on the process exit status. When the subprocess exits zero with valid stdout but closes its stdin read end before Send finishes writing the full reconstructed context, the stdin Write returns EPIPE and sendOnce returns a retryable ErrTransportDrop, discarding the valid stdoutData. The single retry repeats the identical deterministic outcome, so a turn that actually succeeded is surfaced as a transient failure with empty content. This is a distinct failure mode from the accepted proposals p_001/p_003, which address only the non-zero-exit case and the lost stderr; their reorder (broken-pipe-after-non-zero-exit) fix would still leave line 190 discarding a successful zero-exit result.
  Resolution: Restructured the broken-pipe branch in sendOnce (exec.go:190). Now, when isBrokenPipe fires, the code checks whether the process also exited zero with non-empty stdout; if so it returns that result instead of discarding it (fixes the discard bug). For zero exit with empty stdout or any non-zero exit the existing ErrTransportDrop / retry behavior is preserved, keeping TestSend_BrokenPipe_Retried and TestSend_BrokenPipe_BothFail passing.
- f_011 [medium] open internal/backend/exec/exec.go:275: defaultCommandRunner — the only code that actually spawns a subprocess (StdinPipe/StdoutPipe/StderrPipe/Start/Wait wiring) — has no deterministic test coverage. Every unit test injects a fake runner; the sole caller of defaultCommandRunner is New(BackendAgy, ..., nil) inside integration_test.go, which is gated behind //go:build exec_integration AND requires DEBATE_EXEC_INTEGRATION (and AGY_MODEL). The gate runs that test with the env var cleared, so it skips and defaultCommandRunner never executes. A bug in the real pipe wiring would pass the entire default suite. The contract explicitly allows local stub subprocesses in tests ('a fake CLI stub program ... local stub subprocesses in tests are allowed'), so this seam is testable deterministically.
- f_012 [low] open internal/backend/exec/exec.go:258: The real-EPIPE branch of isBrokenPipe (strings.Contains(err.Error(), "broken pipe")) is never exercised by any runnable test. The fake runner's closeStdinEarly path uses io.Pipe + stdinR.Close(), which produces io.ErrClosedPipe ('io: read/write on closed pipe') and hits only the errors.Is(err, io.ErrClosedPipe) branch. A real subprocess that closes stdin before the write completes produces a *fs.PathError wrapping syscall.EPIPE whose message contains 'broken pipe' — the primary real-world retryable path the contract describes — and that branch has zero coverage. A regression in the string-match branch would not be caught by the gate.
- f_013 [low] open cmd/debate/main.go:190: outcomeString is a no-op wrapper: every switch branch (the explicit settled/stalemate/max case and the default) returns the input `reason` unchanged, so the function is an identity function. Its doc comment claims it 'normalises the loop reason to a user-facing string,' but no normalization happens. It is a wrapper that adds nothing / an unused extension point. This is pre-existing code (this change only added the exec import and the `case exec.BackendAgy` resolver branch to cmd/debate/main.go), so it is advisory and non-blocking.
- f_014 [medium] open docs/DESIGN.md:340: docs/DESIGN.md presents 'grounded read-only' as a guaranteed write-rejecting sandbox for all grounded agents, but the exec/agy backend wired into production by this change enforces no read-only at all (resolveCwd only selects the working directory; a plain CLI has no sandbox). The user-facing design doc is not updated to note the exec exception, so a user running a gemini/agy persona in grounded (non-sealed) mode would wrongly assume the project is protected from writes. This is the user-facing-doc counterpart of the code-comment gap tracked by f_009/p_009.
- f_015 [low] open docs/DESIGN.md:303: docs/DESIGN.md documents --sealed as 'дебат только по брифу: без чтения проекта/сети' (debate from the brief only, without reading the project OR the network), but the exec/agy backend's sealed mode only swaps the working directory to a fresh empty temp dir and does not disable network. This change's own contract states 'network is available in both modes', directly contradicting the doc. Pre-existing: the acp backend shipped the same sealed behavior in Slice 5, so the doc was already inaccurate; this change reinforces it for a second backend.
- Proposal summary: pending=0 accepted=15 rejected=0

## Reusable Project Knowledge
- scope: in scope: Implement internal/backend/exec: a transport.Transport that drives a stateless CLI agent by spawning a fresh subprocess per Send via exec.CommandContext, using only the Go standard library.
- scope: in scope: Implement the stateless full-render Session: record each Send prompt and the agent's reply (committed only on success), and on each Send reconstruct the full stdin (system folded in, the recorded prompt/reply history, and the new prompt) for a fresh subprocess.
- scope: in scope: Implement per-backend command derivation (agy) with an env override, spec.Model wiring, grounded-vs-sealed working directory, ctx cancellation, and an explicit error-classification mapping with a single retry.
- scope: in scope: Add deterministic unit tests via a fake CLI / injectable command runner, and a real-agy integration test gated behind a build tag and env var that the gate compiles but skips.
- scope: in scope: Wire the cmd/debate production resolver to register the exec backend for the agy backend, update the Slice-5 unimplemented-backend test, and add targeted exec dep-guard commands.
- scope: out of scope: The api backend (a later concern).
- scope: out of scope: Any change to internal/engine source (including transport.Classify and the orchestrate Delta render), the internal/debate policy/config packages, or the acp/echo/mock backends beyond what wiring the resolver requires.
- scope: out of scope: debate init/new scaffolding.
- scope: out of scope: Making the default test run perform any real agy, model, or network call (local stub subprocesses in tests are allowed); and Windows/non-Unix support.
- review_resolution: f_009 resolved: The exec backend does not document that --sealed / spec.ReadOnly is best-effort: filesystem read-only is NOT enforced (a plain CLI has no sandbox) and network remains available in both grounded and sealed modes. The contract acceptance criterion explicitly requires this be 'documented as trusted/best-effort', but no such comment exists. This is misleading because docs/DESIGN.md frames 'grounded read-only' as a write-rejecting sandbox (true for the ACP backend), so a user passing --sealed for a gemini/agy persona may wrongly assume filesystem writes are prevented.; resolution: Added a note to the resolveCwd doc comment (exec.go:88) stating explicitly that filesystem read-only is NOT enforced (a plain CLI has no sandbox), that it is trusted/best-effort, and that network is available in both grounded and sealed modes.
- review_resolution: f_010 resolved: In sendOnce the stdin broken-pipe branch (exec.go:190) is evaluated before the success return (exec.go:199) and is not gated on the process exit status. When the subprocess exits zero with valid stdout but closes its stdin read end before Send finishes writing the full reconstructed context, the stdin Write returns EPIPE and sendOnce returns a retryable ErrTransportDrop, discarding the valid stdoutData. The single retry repeats the identical deterministic outcome, so a turn that actually succeeded is surfaced as a transient failure with empty content. This is a distinct failure mode from the accepted proposals p_001/p_003, which address only the non-zero-exit case and the lost stderr; their reorder (broken-pipe-after-non-zero-exit) fix would still leave line 190 discarding a successful zero-exit result.; resolution: Restructured the broken-pipe branch in sendOnce (exec.go:190). Now, when isBrokenPipe fires, the code checks whether the process also exited zero with non-empty stdout; if so it returns that result instead of discarding it (fixes the discard bug). For zero exit with empty stdout or any non-zero exit the existing ErrTransportDrop / retry behavior is preserved, keeping TestSend_BrokenPipe_Retried and TestSend_BrokenPipe_BothFail passing.
- review_resolution: proposal p_001 accepted as f_001
- review_resolution: proposal p_002 accepted as f_002
- review_resolution: proposal p_003 accepted as f_003
- review_resolution: proposal p_004 accepted as f_004
- review_resolution: proposal p_005 accepted as f_005
- review_resolution: proposal p_006 accepted as f_006
- review_resolution: proposal p_007 accepted as f_007
- review_resolution: proposal p_008 accepted as f_008
- review_resolution: proposal p_009 accepted as f_009
- review_resolution: proposal p_010 accepted as f_010
- review_resolution: proposal p_011 accepted as f_011
- review_resolution: proposal p_012 accepted as f_012
- review_resolution: proposal p_013 accepted as f_013
- review_resolution: proposal p_014 accepted as f_014
- review_resolution: proposal p_015 accepted as f_015
- validation: bash scripts/check-gofmt.sh passed
- validation: go build ./... passed
- validation: go test -count=1 ./... passed
- validation: env -u DEBATE_EXEC_INTEGRATION go test -count=1 -tags exec_integration ./internal/backend/... passed
- validation: go vet ./... passed
- validation: git diff --exit-code -- internal/engine/ passed
- validation: bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine passed
- validation: bash scripts/dep-guard.sh ./internal/backend/exec/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend passed
- validation: env GOFLAGS=-tags=exec_integration bash scripts/dep-guard.sh ./internal/backend/exec/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend passed
- validation: bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk passed
- validation: bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3 passed

## Artifacts
- Contract: contract/contract.json
- Gate report: gate/gate-report.json
- Review: review/review.json
- Findings: review/findings.jsonl
- Resolutions: review/resolutions.jsonl
- Proposals: review/proposals.jsonl
- Proposal decisions: review/proposal-decisions.jsonl
