# Review Fix Prompt

This prompt is prepared for a write-enabled executor agent subprocess.
Pactum captures the fix attempt artifacts and may parse the required structured outcome block.

## Objective
Address the current run's review findings against the approved Pactum contract.

## Inputs
- Fixer context: .heurema/pactum/runs/run_20260624_162301/review/fix/fixer-context.md
- Contract: .heurema/pactum/runs/run_20260624_162301/contract/contract.json
- Review artifacts: .heurema/pactum/runs/run_20260624_162301/review/review.json, .heurema/pactum/runs/run_20260624_162301/review/findings.jsonl, .heurema/pactum/runs/run_20260624_162301/review/resolutions.jsonl

## Approved contract
- Goal: Replace the hand-written stdlib flag parsing in cmd/debate with github.com/alecthomas/kong (the CLI library pactum uses), so flags parse correctly in any position, while preserving every existing command, flag, and exit-code behavior.
- In scope:
  - Add github.com/alecthomas/kong and parse cmd/debate arguments with it, removing the hand-rolled flag.FlagSet / os.Args switch parsing in cmd/debate/main.go and cmd/debate/scaffold.go.
  - Model the debate run as kong's default command (bare `debate "<task>"`) alongside the version, init, and new subcommands, with all flags as kong struct tags.
  - Preserve the exact existing behavior (commands, flags, task composition, exit codes, stdout/stderr) and make flags parse in any position relative to positionals.
  - Update the cmd/debate tests for kong, including assertions that flags work both before and after the positional argument.
- Out of scope:
  - Any change to internal/engine, internal/debate, or internal/backend.
  - Adding new subcommands (no validate) or new flags, or changing the backend resolver, runner, synthesizer, or IO contract.
  - Changing the debate algorithm, personas, or config.
- Acceptance criteria:
  - cmd/debate parses all arguments with github.com/alecthomas/kong (added to go.mod and go.sum); the hand-rolled flag.NewFlagSet parsing and os.Args dispatch in main.go and scaffold.go are removed.
  - The CLI preserves the existing commands: a default run action invoked as `debate "<task>"` with no subcommand word, plus subcommands `version`, `init`, and `new <name>`; the run is kong's default command so the bare-task form keeps working.
  - Flags parse in any position relative to the positional argument: `debate "<task>" --json`, `debate --json "<task>"`, `debate --max-rounds 2 "<task>"`, `debate "<task>" --max-rounds 2`, and `debate new <name> --role synthesizer` all apply the flag correctly. (This fixes the previous flags-after-positional bug.)
  - All existing run flags are preserved with the same names and meaning: --with (panel selectors), --synth, --max-rounds, --json, -q/--quiet, --sealed, and --task (@file); the new subcommand keeps its --role flag (debater|synthesizer, default debater).
  - Task sources still compose: the positional task, --task @file (file contents), and stdin (appended when piped); an empty resulting task is a fail-fast error.
  - Exit codes are unchanged: 0 when settled, 2 when not converged (stalemate or max), 1 on error; an unknown flag or subcommand prints a clear kong usage/help message with a non-zero exit; the stdout=final-answer / stderr=live-trace (auto-quiet off-TTY or with -q) contract and the --json output shape are unchanged.
  - cmd/debate tests are updated for kong and include assertions that a representative flag is honored both before and after the positional argument for the run command and for `new`.
  - go.mod and go.sum gain github.com/alecthomas/kong; internal/engine, internal/debate, and internal/backend are unchanged; check-gofmt, go build ./..., go vet ./..., and go test ./... pass, and the engine/backend/debate dep-guards still pass.
- Validation commands:
  - bash scripts/check-gofmt.sh
  - go build ./...
  - go test -count=1 ./...
  - go vet ./...
  - go run ./cmd/debate version
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/backend/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/backend github.com/coder/acp-go-sdk
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

## Current review findings
- Summary: findings=6 open=6 resolved=0 blocking_open=3
- Blocking findings (fix or rebut these — emit exactly one fix-outcome for each):
  - f_001 severity=medium category=correctness blocking=true status=open: The kong default run is still exposed as an explicit `run` subcommand, adding a command outside the approved CLI surface and changing task composition for tasks whose first unquoted word is `run`.
    location: cmd/debate/main.go:53
  - f_002 severity=medium category=quality blocking=true status=open: The new max-rounds before/after test can pass even if --max-rounds is ignored.
    location: cmd/debate/e2e_test.go:126
  - f_003 severity=medium category=quality blocking=true status=open: The tests do not exercise the top-level Kong CLI path that main actually uses.
    location: cmd/debate/main.go:101
- Advisory (non-blocking) findings (context only — do NOT edit code and do NOT emit outcomes for them):
  - f_004 severity=low category=quality blocking=false status=open: Kong parse-error paths have no targeted regression tests.
    location: cmd/debate/main.go:145
  - f_005 severity=low category=quality blocking=false status=open: The change introduces duplicate Kong parsing entrypoints outside the production CLI path.
    location: cmd/debate/main.go:190
  - f_006 severity=low category=quality blocking=false status=open: The CLI docs still say `debate new` flags must come before the name, so the newly supported after-name form is undocumented.
    location: docs/DESIGN.md:305

## Fix boundaries
- Trace each finding to the relevant code before acting.
- Fix valid findings in place.
- For false positives, explain a concrete rebuttal instead of changing code.
- Keep changes inside the approved contract and review-finding scope.
- Do not edit `.heurema` artifacts.
- Do not run `pactum review approve`, `pactum review finding resolve`, or `pactum review run`.

## House style
- Match the surrounding code: idiom, naming, comment density.
- Comment only where the code is not self-explanatory; do not narrate the obvious.
- Search for and reuse existing helpers before writing new ones.
- Keep the diff small and focused: change only what the contract requires.
- Simplicity first: no enterprise patterns for simple problems, question every new abstraction, no premature generalization or optimization.
- Over-engineering DON'Ts: wrappers that add nothing, factories or abstractions for a single case, unused extension points, dual implementations where the old path has no callers, silent fallbacks that hide failures.
- No dead code, no commented-out code, no unused parameters.
- Handle errors per the project's existing convention; no silent failures.
- Tests verify behavior, not implementation details, and cover error paths.
- Fake-test DON'Ts: always-pass tests, hardcoded-value checks, assertions on mock behavior instead of the code under test, ignored errors, commented-out cases.

The reviewer will re-check your fixes against the discipline rules above.

## Output shape
Your final output MUST include exactly one fenced `json` block with this shape:

```json
{
  "schema": "pactum.review_fix_outcomes.v1alpha1",
  "outcomes": [
    {
      "finding_id": "f_001",
      "outcome": "fixed",
      "note": "What changed and where, or the concrete rebuttal/blocker."
    }
  ]
}
```

Rules:
- Include exactly one outcome entry for every blocking finding listed above with status open.
- Do NOT edit code for advisory (non-blocking) findings, and do NOT emit outcomes for them; they are context only.
- Use outcome fixed when you changed code to address a valid blocking finding.
- Use outcome rebutted when the blocking finding is a false positive; note must contain the concrete rebuttal.
- Use outcome blocked when concrete missing information or state prevents a fix.
- Do not include advisory or resolved findings in the outcomes list.
