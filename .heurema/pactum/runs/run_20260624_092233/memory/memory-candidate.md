# Memory Candidate

## Run
- Run id: run_20260624_092233
- Source: deterministic

## Contract
- Goal: Slice 3: implement persona loading, .heurema/debate workspace discovery, config, and panel selection in internal/debate, fixture-tested only (no engine run, no real backends, no CLI binary). (1) Persona: a persona is a markdown file with YAML frontmatter (role: debater|synthesizer defaulting to debater; model; effort; optional backend; optional tags list) plus a markdown body that is the system prompt; parse and fail-fast-validate it (reject unknown frontmatter keys; require model and effort for api/acp backends; the persona id is its filename without .md). (2) Backend inference: when backend is absent, infer it from the model name (claude-*/opus/sonnet -> claude-agent-acp; gpt-*/codex/o* -> codex-acp; gemini-* -> agy); an explicit backend overrides inference. (3) Discovery: locate .heurema/debate/ by walking up from a start directory like git does; load an optional config.yml whose only key is table (a list/selection of persona names); load an optional context.md baseline preamble; load personas from .heurema/debate/personas/*.md. (4) Selection: resolve the debater panel from config.table or an explicit list of names (when config is absent, the panel is all debater personas); personas with role synthesizer are excluded from the panel. (5) Synthesizer resolution: choose by an explicit name override, else the persona named synthesizer, else a built-in default (model claude-haiku-4-5 with a minimal prompt). (6) Fail-fast: a clear error before anything else for unknown keys, missing required fields, an unresolvable selection, or a missing .heurema/debate. YAML frontmatter and config parsing may use gopkg.in/yaml.v3. Unit tests use fixture .heurema/debate directories. Out of scope: running the engine, real acp/exec/api transports, the cmd/debate CLI wiring, actually invoking models, and synthesizer execution.
- In scope:
  - Implement internal/debate/persona: a Persona type and a parser/validator for a markdown persona file with YAML frontmatter and a markdown body, including fail-fast validation and backend inference.
  - Implement internal/debate/config: .heurema/debate discovery by walking up parent directories, loading of optional config.yml (key table) and optional context.md, and loading of personas from .heurema/debate/personas/*.md.
  - Implement debater-panel selection and synthesizer resolution (including the built-in default synthesizer) with deterministic ordering and fail-fast errors.
  - Add unit tests using fixture .heurema/debate directories created under a temp dir.
  - Add gopkg.in/yaml.v3 to go.mod and extend scripts/dep-guard.sh invocation to allow it for internal/debate.
  - Add scripts/check-gofmt.sh that runs gofmt -l over internal, cmd, and scripts and exits non-zero (printing offenders) when any Go file is unformatted.
- Out of scope:
  - Running the engine or building an orchestrate.Config; that is the CLI slice.
  - Real ACP, exec, API, network, subprocess, or model-backed transports, and actually invoking any model.
  - The cmd/debate CLI wiring, flag parsing, stdout/stderr contract, and synthesizer execution.
  - Modifying internal/engine source or the Slice 2 signal/prompt/verdict packages beyond what import requires.
- Acceptance criteria:
  - persona.Persona is a struct {ID string, Role string, Model string, Effort string, Backend string, Tags []string, System string}; Role is `debater` or `synthesizer`, defaulting to `debater` when absent; System is the markdown body; ID is the file name without the .md extension.
  - Persona parsing reads a markdown file whose leading ---fenced YAML frontmatter provides role/model/effort/backend/tags and whose remaining body becomes System; it fail-fast rejects unknown frontmatter keys and a role other than debater/synthesizer.
  - Persona validation fail-fast requires a non-empty Model, a non-empty Effort, and a System body that is non-empty after trimming surrounding whitespace; an empty body is rejected with a clear error.
  - Backend inference: when Backend is empty it is inferred from Model — claude-*/opus/sonnet/haiku/fable -> claude-agent-acp; gpt-*/codex/o<digit>* -> codex-acp; gemini-* -> agy. An explicit Backend overrides inference. A model from which no backend can be inferred, with no explicit Backend, is a fail-fast error.
  - config.Discover(startDir) walks up parent directories (like git finding .git) and returns the first .heurema/debate directory found, or a fail-fast error if none exists up to the filesystem root.
  - config loads an optional .heurema/debate/config.yml whose only recognized key is `table` (a list of persona-name selectors); a missing config.yml is allowed, and an unknown key in config.yml is fail-fast rejected.
  - config loads an optional .heurema/debate/context.md as the baseline preamble text (empty when the file is absent) and loads every .heurema/debate/personas/*.md as a Persona.
  - Panel selection returns the debater panel in a deterministic order: from an explicit with-list (preserving the given order) if provided, else from config table (preserving table order), else all personas whose Role is debater ordered lexicographically by persona ID. Personas with Role synthesizer are never in the panel. A selector that names a persona whose Role is synthesizer is a fail-fast error, distinct from a selector that names a nonexistent persona (which is also a fail-fast error); silent skipping of synthesizer-role personas is not permitted when they are explicitly named in a selector. The resolved panel must be non-empty.
  - Synthesizer resolution returns the synthesizer Persona: the persona named by an explicit override if given, else the persona whose ID is `synthesizer`, else a built-in default Persona with exactly Role synthesizer, Model claude-haiku-4-5, Effort `low`, Backend inferred by the normal rules (so claude-agent-acp), and a fixed non-empty built-in system prompt. Tests assert these exact default values. An override naming a missing persona is a fail-fast error.
  - Loading a workspace performs these steps in this exact order and stops at the FIRST error: (1) discover .heurema/debate, (2) parse config.yml, (3) read context.md, (4) parse persona files in lexicographic filename order, (5) resolve the panel, (6) resolve the synthesizer. Fixture tests cover a valid workspace, an unknown frontmatter key, a missing model, an empty body, an un-inferable model, an unknown config key, an unresolvable selection name, a selector that names an existing synthesizer-role persona (fail-fast error), and a missing .heurema/debate.
  - internal/debate/persona and internal/debate/config depend only on the Go standard library, gopkg.in/yaml.v3, and other internal/debate/... and internal/engine/... packages; they never import cmd/debate or any real backend transport and never invoke a model.
  - All tests use fixture .heurema/debate directories created under a temp directory and never read the real repository workspace; go test ./... and go vet ./... pass.
  - scripts/check-gofmt.sh runs gofmt -l over internal, cmd, and scripts and exits non-zero (printing the offending files) when any Go file is unformatted, and exits zero otherwise.
- Validation commands:
  - bash scripts/check-gofmt.sh
  - go test -count=1 ./internal/debate/...
  - go test -count=1 ./...
  - go vet ./...
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
  - go.mod
- New files:
  - go.sum
  - internal/debate/config/config.go
  - internal/debate/config/config_test.go
  - internal/debate/persona/persona.go
  - internal/debate/persona/persona_test.go
  - scripts/check-gofmt.sh
- Missing files: none

## Clarifications
- None

## Review Decisions
- f_001 [medium] resolved internal/debate/config/config.go:166: Default debater panel is ordered by filename (including the .md suffix) rather than by persona ID, contradicting the acceptance criterion 'all personas whose Role is debater ordered lexicographically by persona ID'. config.go:98-102 sorts full glob paths with sort.Strings, and resolvePanel (config.go:165-170) reuses that order for the default panel. Filename order and ID order diverge when one ID is a prefix of another extended by a character that sorts before '.' (0x2E), e.g. '-' (0x2D): files alice.md and alice-pro.md sort to [alice-pro, alice] by filename but must be [alice, alice-pro] by ID.
  Resolution: Added sort.Slice(panel, func(i, j int) bool { return panel[i].ID < panel[j].ID }) at the end of the default-panel branch in resolvePanel (config.go:172). The glob-path sort at line 102 includes the .md suffix, so alice-pro.md sorts before alice.md because '-' (0x2D) < '.' (0x2E), inverting the required ID order. The explicit sort by p.ID corrects this without touching the filename sort (which the lexicographic persona-file loading order depends on). Also updated the stale comment from 'sorted by filename=ID' to 'sorted by persona ID'.
- f_002 [low] open internal/debate/config/config.go:79: An empty or comment-only config.yml at .heurema/debate/config.yml causes Load to fail-fast with a confusing "config.yml: EOF" error instead of behaving like an absent config (default panel = all debaters). The code only treats os.ErrNotExist as 'missing'; a present-but-empty file is read successfully, then dec.Decode returns io.EOF (yaml.v3 returns io.EOF when there is no YAML document), which is not ErrNotExist and is surfaced as an error.
- f_003 [low] open go.mod:5: go.mod declares gopkg.in/yaml.v3 as `// indirect`, but it is a direct dependency: persona.go and config.go both import "gopkg.in/yaml.v3" in non-test code. `go mod tidy` would remove the // indirect marker, so the module graph metadata is inaccurate.
- f_004 [medium] open internal/debate/config/config.go:167: The default-panel branch that excludes synthesizer-role personas is exercised but never asserted. The contract requires synthesizer-role personas to never appear in the panel, including the default (no-selector) branch where resolvePanel filters on `p.Role == "debater"`. The only default-branch test containing a synthesizer.md (TestLoad_SynthOverride) asserts only ws.Synthesizer.ID and never inspects ws.Panel; TestLoad_DefaultPanel has no synthesizer persona; TestLoad_ValidWorkspace uses a table selector. A regression that wrongly included a synthesizer in the default panel would pass all current tests.
- f_005 [low] open internal/debate/persona/persona.go:137: The unclosed-frontmatter error path is untested. splitFrontmatter returns errors.New("unclosed YAML frontmatter") when a persona file starts with a `---` line but never has a closing `---`, but no test feeds such input. All persona fixtures either close their frontmatter or have no leading `---`. This new error path introduced by this change has no coverage, so a regression that mis-handles an unterminated frontmatter fence would not be caught.
- f_006 [medium] open internal/debate/config/config_test.go:323: TestLoad_SelectorNamesSynthesizerRole is supposed to verify acceptance criterion #8 — that naming a synthesizer-role persona in a selector is a fail-fast error 'distinct from a selector that names a nonexistent persona' — but its assertion `strings.Contains(strings.ToLower(err.Error()), "synthesizer")` is satisfied for the wrong reason. The fixture names the persona literally `synthesizer`, so the substring 'synthesizer' appears in BOTH the intended role error (config.go:154 'has role synthesizer and cannot be in the debater panel') AND the wrong-behavior not-found error (config.go:151 'persona "synthesizer" not found'). A regression that removed the role check or omitted synthesizer-role personas from byID, producing the 'not found' error, would still pass this test. The assertion's own comment claims it checks 'not just not found', but it does not enforce that. The companion TestLoad_WithListSynthesizerRole (config_test.go:329) only asserts err != nil with no message check, so it is even weaker. This is distinct from finding f_004 (default no-selector branch has no panel assertion); this concerns the selector branch's error being mis-pinned.
- f_007 [low] open internal/debate/config/config.go:192: buildDefaultSynthesizer carries an error return that can never fire: it calls persona.InferBackend(defaultSynthModel) where defaultSynthModel is the compile-time constant "claude-haiku-4-5", which InferBackend always resolves to "claude-agent-acp" with a nil error. The error branch at config.go:194-196 is structurally unreachable dead code, and the (persona.Persona, error) signature propagates an error that the default-synthesizer path can never produce.
- Proposal summary: pending=0 accepted=7 rejected=0

## Reusable Project Knowledge
- scope: in scope: Implement internal/debate/persona: a Persona type and a parser/validator for a markdown persona file with YAML frontmatter and a markdown body, including fail-fast validation and backend inference.
- scope: in scope: Implement internal/debate/config: .heurema/debate discovery by walking up parent directories, loading of optional config.yml (key table) and optional context.md, and loading of personas from .heurema/debate/personas/*.md.
- scope: in scope: Implement debater-panel selection and synthesizer resolution (including the built-in default synthesizer) with deterministic ordering and fail-fast errors.
- scope: in scope: Add unit tests using fixture .heurema/debate directories created under a temp dir.
- scope: in scope: Add gopkg.in/yaml.v3 to go.mod and extend scripts/dep-guard.sh invocation to allow it for internal/debate.
- scope: in scope: Add scripts/check-gofmt.sh that runs gofmt -l over internal, cmd, and scripts and exits non-zero (printing offenders) when any Go file is unformatted.
- scope: out of scope: Running the engine or building an orchestrate.Config; that is the CLI slice.
- scope: out of scope: Real ACP, exec, API, network, subprocess, or model-backed transports, and actually invoking any model.
- scope: out of scope: The cmd/debate CLI wiring, flag parsing, stdout/stderr contract, and synthesizer execution.
- scope: out of scope: Modifying internal/engine source or the Slice 2 signal/prompt/verdict packages beyond what import requires.
- review_resolution: f_001 resolved: Default debater panel is ordered by filename (including the .md suffix) rather than by persona ID, contradicting the acceptance criterion 'all personas whose Role is debater ordered lexicographically by persona ID'. config.go:98-102 sorts full glob paths with sort.Strings, and resolvePanel (config.go:165-170) reuses that order for the default panel. Filename order and ID order diverge when one ID is a prefix of another extended by a character that sorts before '.' (0x2E), e.g. '-' (0x2D): files alice.md and alice-pro.md sort to [alice-pro, alice] by filename but must be [alice, alice-pro] by ID.; resolution: Added sort.Slice(panel, func(i, j int) bool { return panel[i].ID < panel[j].ID }) at the end of the default-panel branch in resolvePanel (config.go:172). The glob-path sort at line 102 includes the .md suffix, so alice-pro.md sorts before alice.md because '-' (0x2D) < '.' (0x2E), inverting the required ID order. The explicit sort by p.ID corrects this without touching the filename sort (which the lexicographic persona-file loading order depends on). Also updated the stale comment from 'sorted by filename=ID' to 'sorted by persona ID'.
- review_resolution: proposal p_001 accepted as f_001
- review_resolution: proposal p_002 accepted as f_002
- review_resolution: proposal p_003 accepted as f_003
- review_resolution: proposal p_004 accepted as f_004
- review_resolution: proposal p_005 accepted as f_005
- review_resolution: proposal p_006 accepted as f_006
- review_resolution: proposal p_007 accepted as f_007
- validation: bash scripts/check-gofmt.sh passed
- validation: go test -count=1 ./internal/debate/... passed
- validation: go test -count=1 ./... passed
- validation: go vet ./... passed
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
