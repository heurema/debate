# Contract Review Fixer Prompt

You are fixing a software change contract to address blocking review findings.

Current contract version: eeaeddff49a63c4986a2d0c3be479ccbfab7ee6a97d850228f45bc73a1091a8f

## Current Contract

**Goal**: Slice 3: implement persona loading, .heurema/debate workspace discovery, config, and panel selection in internal/debate, fixture-tested only (no engine run, no real backends, no CLI binary). (1) Persona: a persona is a markdown file with YAML frontmatter (role: debater|synthesizer defaulting to debater; model; effort; optional backend; optional tags list) plus a markdown body that is the system prompt; parse and fail-fast-validate it (reject unknown frontmatter keys; require model and effort for api/acp backends; the persona id is its filename without .md). (2) Backend inference: when backend is absent, infer it from the model name (claude-*/opus/sonnet -> claude-agent-acp; gpt-*/codex/o* -> codex-acp; gemini-* -> agy); an explicit backend overrides inference. (3) Discovery: locate .heurema/debate/ by walking up from a start directory like git does; load an optional config.yml whose only key is table (a list/selection of persona names); load an optional context.md baseline preamble; load personas from .heurema/debate/personas/*.md. (4) Selection: resolve the debater panel from config.table or an explicit list of names (when config is absent, the panel is all debater personas); personas with role synthesizer are excluded from the panel. (5) Synthesizer resolution: choose by an explicit name override, else the persona named synthesizer, else a built-in default (model claude-haiku-4-5 with a minimal prompt). (6) Fail-fast: a clear error before anything else for unknown keys, missing required fields, an unresolvable selection, or a missing .heurema/debate. YAML frontmatter and config parsing may use gopkg.in/yaml.v3. Unit tests use fixture .heurema/debate directories. Out of scope: running the engine, real acp/exec/api transports, the cmd/debate CLI wiring, actually invoking models, and synthesizer execution.

**Scope in**:
  - Implement internal/debate/persona: a Persona type and a parser/validator for a markdown persona file with YAML frontmatter and a markdown body, including fail-fast validation and backend inference.
  - Implement internal/debate/config: .heurema/debate discovery by walking up parent directories, loading of optional config.yml (key table) and optional context.md, and loading of personas from .heurema/debate/personas/*.md.
  - Implement debater-panel selection and synthesizer resolution (including the built-in default synthesizer) with deterministic ordering and fail-fast errors.
  - Add unit tests using fixture .heurema/debate directories created under a temp dir.
  - Add gopkg.in/yaml.v3 to go.mod and extend scripts/dep-guard.sh invocation to allow it for internal/debate.
  - Add scripts/check-gofmt.sh that runs gofmt -l over internal, cmd, and scripts and exits non-zero (printing offenders) when any Go file is unformatted.

**Scope out**:
  - Running the engine or building an orchestrate.Config; that is the CLI slice.
  - Real ACP, exec, API, network, subprocess, or model-backed transports, and actually invoking any model.
  - The cmd/debate CLI wiring, flag parsing, stdout/stderr contract, and synthesizer execution.
  - Modifying internal/engine source or the Slice 2 signal/prompt/verdict packages beyond what import requires.

**Acceptance criteria**:
  - persona.Persona is a struct {ID string, Role string, Model string, Effort string, Backend string, Tags []string, System string}; Role is `debater` or `synthesizer`, defaulting to `debater` when absent; System is the markdown body; ID is the file name without the .md extension.
  - Persona parsing reads a markdown file whose leading ---fenced YAML frontmatter provides role/model/effort/backend/tags and whose remaining body becomes System; it fail-fast rejects unknown frontmatter keys and a role other than debater/synthesizer.
  - Persona validation fail-fast requires a non-empty Model, a non-empty Effort, and a System body that is non-empty after trimming surrounding whitespace; an empty body is rejected with a clear error.
  - Backend inference: when Backend is empty it is inferred from Model — claude-*/opus/sonnet/haiku/fable -> claude-agent-acp; gpt-*/codex/o<digit>* -> codex-acp; gemini-* -> agy. An explicit Backend overrides inference. A model from which no backend can be inferred, with no explicit Backend, is a fail-fast error.
  - config.Discover(startDir) walks up parent directories (like git finding .git) and returns the first .heurema/debate directory found, or a fail-fast error if none exists up to the filesystem root.
  - config loads an optional .heurema/debate/config.yml whose only recognized key is `table` (a list of persona-name selectors); a missing config.yml is allowed, and an unknown key in config.yml is fail-fast rejected.
  - config loads an optional .heurema/debate/context.md as the baseline preamble text (empty when the file is absent) and loads every .heurema/debate/personas/*.md as a Persona.
  - Panel selection returns the debater panel in a deterministic order: from an explicit with-list (preserving the given order) if provided, else from config table (preserving table order), else all personas whose Role is debater ordered lexicographically by persona ID. Personas with Role synthesizer are never in the panel; an unresolvable selector name is a fail-fast error; the resolved panel must be non-empty.
  - Synthesizer resolution returns the synthesizer Persona: the persona named by an explicit override if given, else the persona whose ID is `synthesizer`, else a built-in default Persona with exactly Role synthesizer, Model claude-haiku-4-5, Effort `low`, Backend inferred by the normal rules (so claude-agent-acp), and a fixed non-empty built-in system prompt. Tests assert these exact default values. An override naming a missing persona is a fail-fast error.
  - Loading a workspace performs these steps in this exact order and stops at the FIRST error: (1) discover .heurema/debate, (2) parse config.yml, (3) read context.md, (4) parse persona files in lexicographic filename order, (5) resolve the panel, (6) resolve the synthesizer. Fixture tests cover a valid workspace, an unknown frontmatter key, a missing model, an empty body, an un-inferable model, an unknown config key, an unresolvable selection name, and a missing .heurema/debate.
  - internal/debate/persona and internal/debate/config depend only on the Go standard library, gopkg.in/yaml.v3, and other internal/debate/... and internal/engine/... packages; they never import cmd/debate or any real backend transport and never invoke a model.
  - All tests use fixture .heurema/debate directories created under a temp directory and never read the real repository workspace; go test ./... and go vet ./... pass.
  - scripts/check-gofmt.sh runs gofmt -l over internal, cmd, and scripts and exits non-zero (printing the offending files) when any Go file is unformatted, and exits zero otherwise.

**Validation commands**:
  - bash scripts/check-gofmt.sh
  - go test -count=1 ./internal/debate/...
  - go test -count=1 ./...
  - go vet ./...
  - bash scripts/dep-guard.sh ./internal/engine/... github.com/heurema/debate/internal/engine
  - bash scripts/dep-guard.sh ./internal/debate/... github.com/heurema/debate/internal/engine github.com/heurema/debate/internal/debate gopkg.in/yaml.v3

**Assumptions**:
  - The persona schema is invariant: role, model, effort, optional backend, optional tags, and the markdown body as the system prompt; model, effort, and a non-empty body are required for every persona including the synthesizer.
  - Backend strings are the product-level backends (claude-agent-acp, codex-acp, agy, and future api variants); this slice only records the backend string and resolves it from the model — it never opens a transport.
  - .heurema/debate discovery walks up to the filesystem root and stops at the first .heurema/debate directory, mirroring git .git discovery.
  - config.yml has exactly one recognized key, table; its absence means the panel defaults to all debater personas ordered lexicographically by ID; table values are persona names (tag-based selection is a later slice).
  - Default panel ordering is lexicographic by persona ID; an explicit with-list and config.table each preserve their given order.
  - The built-in default synthesizer uses Model claude-haiku-4-5, Effort low, and a fixed minimal built-in prompt; a user synthesizer.md overrides it and an explicit override name overrides that.
  - gopkg.in/yaml.v3 is the YAML parser for frontmatter and config.yml; adding it to go.mod and go.sum is expected and the only third-party dependency this slice introduces.
  - This slice produces in-memory loaded and validated data (personas, panel, synthesizer persona, baseline preamble); it does not build an orchestrate.Config or run the engine.

## Blocking Findings to Address

1. [codex-xhigh/assumptions-surfaced] The contract does not explicitly state how panel selection should handle a selector that names an existing synthesizer persona.
   Evidence: Acceptance criteria: "Personas with Role synthesizer are never in the panel; an unresolvable selector name is a fail-fast error" and assumptions: "table values are persona names".

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
  "base_version": "eeaeddff49a63c4986a2d0c3be479ccbfab7ee6a97d850228f45bc73a1091a8f",
  "contract": {
    "acceptance_criteria": ["...updated criteria..."],
    "validation": {"commands": ["...updated commands..."]}
  }
}
```

Omit any contract field you are not changing. Do not include the goal field.
