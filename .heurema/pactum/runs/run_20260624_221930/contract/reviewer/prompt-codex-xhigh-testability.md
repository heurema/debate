# Contract Review: Testability

You are reviewing a software change contract through the **acceptance-testability** lens.

Review the contract fields below using only your assigned lens checklist.
Do not flag issues that belong to other lenses.

## Contract

**Goal**: Harden the ACP backend client tool contract so grounded file reads are explicitly advertised, root-scoped, and bounded, while non-read operations are denied conservatively and the existing timeout guard remains intact.

**Scope in**:
  - internal/backend/acp/acp.go: ACP InitializeRequest client capabilities, clientImpl construction/state, ReadTextFile, WriteTextFile, RequestPermission, and terminal-method error behavior.
  - internal/backend/acp/acp_test.go: focused fake ACP/client tests for advertised capabilities, root-scoped reads, line/limit bounded reads, path escape denial, permission option ordering, terminal denial, and preservation of timeout behavior.
  - internal/backend/acp/integration_test.go only if a compile-time adjustment is required by internal helper changes.

**Scope out**:
  - Changing the public cmd/debate CLI, persona/config parsing, debate orchestration, engine loop, exec backend, or README/docs.
  - Removing or weakening the existing DEBATE_ACP_OPEN_TIMEOUT and DEBATE_ACP_SEND_TIMEOUT safety boundary.
  - Implementing a real terminal, write support, shell command execution, activity-based idle watchdog, or subprocess stderr ring-buffer in this slice.
  - Adding network-dependent validation or requiring real ACP adapters in the default test suite.

**Acceptance criteria**:
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

**Validation commands**:
  - bash scripts/check-gofmt.sh
  - go test -count=1 ./internal/backend/acp
  - go test -count=1 -run '^$' -tags acp_integration ./internal/backend/acp
  - go test -count=1 ./...
  - go vet ./...
  - go build ./cmd/debate
  - bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk
  - bash -lc 'disallowed=$( { git diff --name-only -- .; git diff --cached --name-only -- .; git ls-files --others --exclude-standard -- .; } | sort -u | grep -Ev "^(internal/backend/acp/acp\.go|internal/backend/acp/acp_test\.go|internal/backend/acp/integration_test\.go|\.heurema/pactum/)" || true); test -z "$disallowed"'

**Assumptions**:
  - Pactum run/audit artifacts under .heurema/pactum/** are durable workflow records, not implementation source changes, and are exempt from the source-file allowlist gate.
  - The opened ACP session cwd is the intended and trustworthy read root for this backend slice, and it is expected to be a host-local filesystem path visible to the backend process. This is a security boundary: the implementation must capture it once, canonicalize and symlink-evaluate it before use, and fail closed if that cwd cannot be trusted. ACP adapters that expose remote, container-only, virtual, or otherwise backend-invisible cwd paths are unsupported for grounded file reads in this slice and must produce the deterministic invalid-root/root-unavailable failure rather than attempting ambiguous path resolution.
  - github.com/coder/acp-go-sdk v0.13.5 exposes InitializeRequest.ClientCapabilities, FileSystemCapabilities, ReadTextFileRequest.Line/Limit, RequestPermissionRequest.ToolCall.Kind, ToolKindRead, and PermissionOption.Kind, so no SDK upgrade is required.
  - Permission requests and options expose enough structured metadata to recognize exactly acpsdk.ToolKindRead plus acpsdk.PermissionOptionKindAllowOnce; when that structure is absent, non-read, allow-always-only, or ambiguous, the safe behavior is to return Cancelled.
  - The current timeout guard is a necessary outer safety boundary and should be preserved while this slice fixes the client tool contract.
  - A conservative permission response may reduce adapter capability, but debate currently intends read-only operation and does not need writes, terminal access, or durable permission grants.

## Lens: Testability

Checklist:
- Is each acceptance criterion backed by or expressible as a runnable validation command (not just prose)?
- Are any criteria purely prose with no machine-checkable outcome?

## Output

State your analysis in prose. If you find issues, also include a structured block:

```json
{
  "schema": "pactum.reviewer_findings.v1alpha1",
  "findings": [
    {
      "message": "Describe the contract issue clearly.",
      "severity": "medium",
      "category": "quality",
      "blocking": true,
      "evidence": "Quote or cite the contract field that shows the issue."
    }
  ]
}
```

Rules:
- Use severity: low, medium, high, critical.
- Use category: correctness, scope, quality, validation, process, other.
- Omit file and line (not applicable for contract review).
- Set blocking=true for defects that should block approval: gaps that make the contract unexecutable or ungatable.
- Set blocking=false for advisory issues.
- If no issues, say so clearly. Do not include an empty findings block.
