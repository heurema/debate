# Memory Candidate

## Run
- Run id: run_20260624_132440
- Source: deterministic

## Contract
- Goal: Add debate init and debate new scaffolding subcommands to cmd/debate that create a ready-to-run .heurema/debate workspace and new persona files, adding no new module dependency, without changing internal/engine, internal/debate, or internal/backend.
- In scope:
  - Implement the `debate init` subcommand: scaffold a .heurema/debate workspace under the current directory with two starter debater personas and a context.md template, safely (never overwriting existing files).
  - Implement the `debate new <name>` subcommand: create a new persona file from a template under a discovered .heurema/debate/personas, with a role flag, safely (never overwriting an existing persona).
  - Make the scaffolded workspace immediately loadable and runnable (valid personas that load via the existing config/persona packages).
  - Add deterministic unit tests using temporary directories that assert init and new behavior, including the loadability of the scaffolded workspace and the refuse-to-overwrite behavior.
- Out of scope:
  - Any change to internal/engine, internal/debate, or internal/backend source (the subcommands live in cmd/debate and reuse config/persona by import only).
  - Backends, the debate run path, the synthesizer, or convergence behavior.
  - Editing or migrating an existing workspace's content beyond adding new files; and any new third-party module dependency.
- Acceptance criteria:
  - `debate init` creates a .heurema/debate directory under the current working directory containing personas/proposer.md and personas/skeptic.md (each a valid debater persona with role debater, a concrete model and effort, and a system-prompt body) and a context.md template file; it prints the paths it created.
  - The scaffolded workspace loads successfully via the existing config.Load: discovery finds the new .heurema/debate, the panel resolves to the two starter debaters, and the built-in default synthesizer is used (init does not scaffold a synthesizer file or a config.yml, since the default panel is all debater personas).
  - The two starter personas use concrete valid values (a real model id such as claude-haiku-4-5 and a valid effort) so the workspace is immediately runnable without edits; persona.ParseFile accepts both.
  - `debate init` is safe: if a target file already exists it does not overwrite it (it skips that file with a clear message or refuses with a clear error and a documented exit code); an existing .heurema/debate is never clobbered.
  - `debate new <name>` creates <name>.md under the discovered .heurema/debate/personas from a template with YAML frontmatter (role defaulting to debater, overridable via --role debater|synthesizer; a concrete model and effort default; optional backend) and a placeholder body; it prints the created path, and the created file is accepted by persona.ParseFile.
  - `debate new` validates the name (a simple persona id, rejecting path separators) and refuses to overwrite an existing persona file with a clear error; it requires a discoverable .heurema/debate workspace (walking up parent directories like config discovery) and errors clearly when none is found, creating the personas directory within the discovered workspace if it does not yet exist.
  - init and new write only under .heurema/debate, never outside it, use the Go standard library plus the existing internal packages, and add no new module dependency (go.mod and go.sum are unchanged); an unknown flag or a missing required argument prints a clear usage error with a non-zero exit.
  - Deterministic unit tests using temporary directories assert: init creates a workspace that config.Load accepts with the two-debater panel, re-running init does not overwrite existing files, new creates a persona that persona.ParseFile accepts, and new refuses to overwrite an existing persona; check-gofmt, go vet ./..., go build ./..., and go test ./... pass.
- Validation commands:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - go vet ./...
  - go run ./cmd/debate version
  - git diff --exit-code -- go.mod go.sum
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

## Outcome
- Gate status: needs_review
- Review status: approved
- Execution exit code: 0
- Validation passed: true
- Changes need review: true

## Changes
- Changed files:
  - cmd/debate/main.go
- New files:
  - cmd/debate/scaffold.go
  - cmd/debate/scaffold_test.go
- Missing files: none

## Clarifications
- None

## Review Decisions
- f_001 [low] open cmd/debate/scaffold.go:185: writeIfAbsent does not remove the file it just created if WriteString fails. Because os.OpenFile with O_CREATE|O_EXCL has already created the (now empty or partially written) file on disk before the write error, a subsequent re-run of `debate init`/`debate new` sees the file as existing and reports 'skipped (already exists)' rather than retrying — leaving a corrupt persona that config.Load / persona.ParseFile will reject (empty body or truncated frontmatter), making the workspace unloadable until the user manually deletes the stray file.
- f_002 [medium] open cmd/debate/scaffold_test.go:169: TestCmdNew_RejectsPathSeparators does not actually verify path-separator/name validation. It runs in a bare t.TempDir() with no .heurema/debate workspace, so config.Discover fails for every input. Because validatePersonaName (scaffold.go:117) runs before config.Discover (scaffold.go:127) but the test only asserts code != 0 and never inspects stderr, the non-zero exit is guaranteed by the missing-workspace path regardless of name validation. The test would still pass if validatePersonaName were a no-op, giving false confidence in the path-traversal protection.
- f_003 [low] open cmd/debate/scaffold_test.go:44: No test asserts that `debate init` creates context.md. config.Load swallows a missing context.md (os.ErrNotExist is treated as empty context, config.go:90-93), so TestCmdInit_CreatesWorkspaceThatLoads passing does not prove context.md exists. The acceptance criterion requires init to scaffold a context.md template, but its creation is unverified.
- f_004 [low] open cmd/debate/scaffold_test.go:206: The 'too many positional arguments' error path in cmdNew (scaffold.go:110-114) has no test. cmdNew explicitly handles fs.NArg() > 1 with a distinct error and exit code, but no test passes more than one positional argument.
- f_005 [low] open cmd/debate/main.go:199: outcomeString is a no-op: every switch branch (including default) returns the input reason unchanged, so the function is an identity wrapper with a meaningless switch. It adds indirection without behavior.
- f_006 [low] open docs/DESIGN.md:306: docs/DESIGN.md §7 and docs/SLICES.md §7 state that `debate init` scaffolds a `config.yml`, but the implemented init creates only personas/proposer.md, personas/skeptic.md, and context.md (no config.yml). The user-facing CLI reference now misdescribes what `init` produces.
- f_007 [low] open docs/DESIGN.md:307: docs/DESIGN.md:307 documents `debate new <name> [--tags ...]`, but the implemented subcommand exposes `--role debater|synthesizer` and has no `--tags` flag. The actually-supported `--role` flag is undocumented in the CLI reference.
- Proposal summary: pending=0 accepted=7 rejected=0

## Reusable Project Knowledge
- scope: in scope: Implement the `debate init` subcommand: scaffold a .heurema/debate workspace under the current directory with two starter debater personas and a context.md template, safely (never overwriting existing files).
- scope: in scope: Implement the `debate new <name>` subcommand: create a new persona file from a template under a discovered .heurema/debate/personas, with a role flag, safely (never overwriting an existing persona).
- scope: in scope: Make the scaffolded workspace immediately loadable and runnable (valid personas that load via the existing config/persona packages).
- scope: in scope: Add deterministic unit tests using temporary directories that assert init and new behavior, including the loadability of the scaffolded workspace and the refuse-to-overwrite behavior.
- scope: out of scope: Any change to internal/engine, internal/debate, or internal/backend source (the subcommands live in cmd/debate and reuse config/persona by import only).
- scope: out of scope: Backends, the debate run path, the synthesizer, or convergence behavior.
- scope: out of scope: Editing or migrating an existing workspace's content beyond adding new files; and any new third-party module dependency.
- review_resolution: proposal p_001 accepted as f_001
- review_resolution: proposal p_002 accepted as f_002
- review_resolution: proposal p_003 accepted as f_003
- review_resolution: proposal p_004 accepted as f_004
- review_resolution: proposal p_005 accepted as f_005
- review_resolution: proposal p_006 accepted as f_006
- review_resolution: proposal p_007 accepted as f_007
- validation: bash scripts/check-gofmt.sh passed
- validation: go build ./... passed
- validation: go test -count=1 ./... passed
- validation: go vet ./... passed
- validation: go run ./cmd/debate version passed
- validation: git diff --exit-code -- go.mod go.sum passed
- validation: bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine passed
- validation: bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3 passed

## Artifacts
- Contract: contract/contract.json
- Gate report: gate/gate-report.json
- Review: review/review.json
- Findings: review/findings.jsonl
- Resolutions: review/resolutions.jsonl
- Proposals: review/proposals.jsonl
- Proposal decisions: review/proposal-decisions.jsonl
