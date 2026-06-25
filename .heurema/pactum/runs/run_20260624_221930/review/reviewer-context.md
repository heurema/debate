# Reviewer Context

## Run
- Run id: run_20260624_221930
- Run status: contract_approved

## Contract
- Goal: Harden the ACP backend client tool contract so grounded file reads are explicitly advertised, root-scoped, and bounded, while non-read operations are denied conservatively and the existing timeout guard remains intact.
- In scope:
  - internal/backend/acp/acp.go: ACP InitializeRequest client capabilities, clientImpl construction/state, ReadTextFile, WriteTextFile, RequestPermission, and terminal-method error behavior.
  - internal/backend/acp/acp_test.go: focused fake ACP/client tests for advertised capabilities, root-scoped reads, line/limit bounded reads, path escape denial, permission option ordering, terminal denial, and preservation of timeout behavior.
  - internal/backend/acp/integration_test.go only if a compile-time adjustment is required by internal helper changes.
- Out of scope:
  - Changing the public cmd/debate CLI, persona/config parsing, debate orchestration, engine loop, exec backend, or README/docs.
  - Removing or weakening the existing DEBATE_ACP_OPEN_TIMEOUT and DEBATE_ACP_SEND_TIMEOUT safety boundary.
  - Implementing a real terminal, write support, shell command execution, activity-based idle watchdog, or subprocess stderr ring-buffer in this slice.
  - Adding network-dependent validation or requiring real ACP adapters in the default test suite.
- Acceptance criteria:
  - Initialize sends explicit ClientCapabilities with fs.readTextFile=true, fs.writeTextFile=false, and terminal=false; tests assert the fake ACP agent observes these values.
  - clientImpl is session-root aware: it captures a canonical, absolute, symlink-evaluated read root from the opened ACP session cwd once during session/client initialization and uses that fixed root for all later reads; if the cwd is missing, empty, relative, cannot be made absolute, cannot be evaluated, or otherwise cannot be trusted, initialization or reads fail closed with a deterministic invalid-root/root-unavailable error; later cwd changes during the session do not expand or change the captured read root.
  - ReadTextFile resolves requested absolute or relative paths against the captured session read root and rejects path traversal, direct outside-root paths, and symlink escapes with deterministic errors whose text includes a stable outside-root/path-escape marker asserted by tests; sealed temp-directory behavior remains isolated.
  - Tests explicitly cover a symlink inside the session root pointing outside the root and assert that reading through that symlink is denied with the stable outside-root/path-escape marker.
  - ReadTextFile honors ACP Line and Limit fields with explicit boundary semantics: Line is 1-based and inclusive; absent or zero Line starts at line 1; positive Limit returns at most that many logical lines; absent or zero Limit reads through EOF subject to the byte cap; Line beyond EOF returns empty content without reading outside the file; negative Line or Limit values, if representable by the SDK types, are rejected with deterministic validation errors. Logical lines are delimited by LF bytes (`\n`); CRLF is counted as one line break because the LF terminates the line, and returned content preserves the original selected bytes including CR bytes and original line terminators. A final non-empty unterminated segment counts as a line, while a trailing LF does not create an additional selectable empty line.
  - ReadTextFile applies a deterministic max byte cap of 1 MiB (1048576 bytes) to returned content after root validation and line selection; if the selected content would exceed the cap, ReadTextFile returns a deterministic size-limit error instead of partial content. Tests cover full read, line/limit read with exact newline-preservation expectations including CRLF and trailing-newline behavior, zero/omitted boundary behavior, line beyond EOF, oversized selected content, cap-safe limited content, symlink escape denial, and outside-root denial.
  - WriteTextFile remains denied with a deterministic unsupported/write-denied error asserted by tests, and every terminal client method exposed by github.com/coder/acp-go-sdk v0.13.5 remains denied with a deterministic terminal-unsupported error asserted by tests: CreateTerminal (terminal/create), KillTerminal (terminal/kill), TerminalOutput (terminal/output), ReleaseTerminal (terminal/release), and WaitForTerminalExit (terminal/wait_for_exit).
  - RequestPermission does not blindly select Options[0]. It selects an option only when the request metadata and option metadata exactly match safe read-only semantics: RequestPermissionRequest.ToolCall.Kind is non-nil and equals acpsdk.ToolKindRead, the selected PermissionOption.Kind equals acpsdk.PermissionOptionKindAllowOnce, and the selected PermissionOption.OptionId is non-empty. The implementation scans options in caller-provided order and selects the first option meeting those exact criteria, so a reject option before a later allow-once read option is skipped. It must not infer safety from option name/title text, raw input, locations, or _meta. If ToolCall.Kind is missing or is any value other than acpsdk.ToolKindRead, if no allow-once option is present, if only allow-always/reject/unknown options are present, or if the semantics are otherwise missing or ambiguous, it returns RequestPermissionOutcome.Cancelled. Tests cover reject-first ordering, missing kind, non-read tool kinds, allow-always-only, empty options, empty option id, and ambiguous metadata cancellation.
  - Existing send/open timeout behavior remains unchanged: prompt/open timeouts classify as idle_timeout, do not retry prompt timeouts, and kill the spawned ACP session; existing timeout tests still pass.
  - The implementation change stays inside the ACP backend package; no public CLI, docs, engine, debate policy, or exec backend behavior changes. The validation allowlist command ignores Pactum run/audit artifacts under .heurema/pactum/**, then fails if any other unstaged, staged, or untracked file outside the approved ACP backend paths is present.
- Validation commands:
  - bash scripts/check-gofmt.sh
  - go test -count=1 ./internal/backend/acp
  - go test -count=1 -run '^$' -tags acp_integration ./internal/backend/acp
  - go test -count=1 ./...
  - go vet ./...
  - go build ./cmd/debate
  - bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk
  - bash -lc 'disallowed=$( { git diff --name-only -- .; git diff --cached --name-only -- .; git ls-files --others --exclude-standard -- .; } | sort -u | grep -Ev "^(internal/backend/acp/acp\.go|internal/backend/acp/acp_test\.go|internal/backend/acp/integration_test\.go|\.heurema/pactum/)" || true); test -z "$disallowed"'

## Accepted memory
- Memory context: context/memory-context.md
- Selected items: 5
- Fresh: 4
- Stale: 1
- Unknown: 0
- Stale memory may be outdated and must be verified.

## Gate report
- Gate status: needs_review
- Execution attempt id: attempt_001
- Execution exit code: 0
- Validation command results:
  - command_001: bash scripts/check-gofmt.sh (exit 0, timed out: false, result: gate/validation/command_001/result.json)
  - command_002: go test -count=1 ./internal/backend/acp (exit 0, timed out: false, result: gate/validation/command_002/result.json)
  - command_003: go test -count=1 -run '^$' -tags acp_integration ./internal/backend/acp (exit 0, timed out: false, result: gate/validation/command_003/result.json)
  - command_004: go test -count=1 ./... (exit 0, timed out: false, result: gate/validation/command_004/result.json)
  - command_005: go vet ./... (exit 0, timed out: false, result: gate/validation/command_005/result.json)
  - command_006: go build ./cmd/debate (exit 0, timed out: false, result: gate/validation/command_006/result.json)
  - command_007: bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk (exit 0, timed out: false, result: gate/validation/command_007/result.json)
  - command_008: bash -lc 'disallowed=$( { git diff --name-only -- .; git diff --cached --name-only -- .; git ls-files --others --exclude-standard -- .; } | sort -u | grep -Ev "^(internal/backend/acp/acp\.go|internal/backend/acp/acp_test\.go|internal/backend/acp/integration_test\.go|\.heurema/pactum/)" || true); test -z "$disallowed"' (exit 0, timed out: false, result: gate/validation/command_008/result.json)
- Change summary:
  - changed files:
    - internal/backend/acp/acp.go
    - internal/backend/acp/acp_test.go
  - new files:
    - none
  - missing files:
    - none

## Existing manual review
- Review status: pending
- Current findings summary: findings=0 open=0 resolved=0 blocking_open=0
- Existing findings:
  - none
- Existing resolutions:
  - none
- Proposal summary: pending=0 accepted=0 rejected=0
- Existing proposals:
  - none

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
- If you are not certain an issue is real after verification, do not flag it.
