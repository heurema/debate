# Project Memory

## Accepted memory items

### mem_001 - Replace the hand-written stdlib flag parsing in cmd/debate with github.com/al...
- Run: run_20260624_162301
- Freshness: stale
- Files: cmd/debate/e2e_test.go, cmd/debate/main.go, cmd/debate/scaffold.go, cmd/debate/scaffold_test.go, docs/DESIGN.md, go.mod, go.sum
- Summary: Reviewed run run_20260624_162301 with gate status needs_review and review status approved. Goal: Replace the hand-written stdlib flag parsing in cmd/debate with github.com/alecthomas/kong (the CLI library pactum uses), so flags parse cor...
- Candidate: runs/run_20260624_162301/memory/memory-candidate.json

### mem_002 - Slice 0: bootstrap the debate Go project skeleton — go.mod (module github.com...
- Run: run_20260623_213058
- Freshness: fresh
- Files: Makefile, cmd/debate/main.go, cmd/debate/main_test.go, go.mod, internal/debate/debate.go, internal/engine/loop/loop.go, internal/engine/orchestrate/orchestrate.go, internal/engine/transport/transport.go
- Summary: Reviewed run run_20260623_213058 with gate status needs_review and review status approved. Goal: Slice 0: bootstrap the debate Go project skeleton — go.mod (module github.com/heurema/debate), package layout (internal/engine/{loop,transpo...
- Candidate: runs/run_20260623_213058/memory/memory-candidate.json

### mem_003 - Slice 1: implement the policy-free engine on a mock backend. Package internal...
- Run: run_20260623_220044
- Freshness: fresh
- Files: internal/engine/loop/loop.go, internal/engine/loop/loop_test.go, internal/engine/orchestrate/orchestrate.go, internal/engine/orchestrate/orchestrate_test.go, internal/engine/transport/mock/mock.go, internal/engine/transport/mock/mock_test.go, internal/engine/transport/transport.go, internal/engine/transport/transport_test.go, scripts/dep-guard.sh
- Summary: Reviewed run run_20260623_220044 with gate status needs_review and review status approved. Goal: Slice 1: implement the policy-free engine on a mock backend. Package internal/engine/loop: a streak loop Run(ctx, Limits{Max,Settle,Patience...
- Candidate: runs/run_20260623_220044/memory/memory-candidate.json

### mem_004 - Slice 2: implement the debate policy layer in internal/debate on top of the e...
- Run: run_20260624_085539
- Freshness: fresh
- Files: internal/debate/prompt/prompt.go, internal/debate/prompt/prompt_test.go, internal/debate/signal/signal.go, internal/debate/signal/signal_test.go, internal/debate/verdict/verdict.go, internal/debate/verdict/verdict_test.go, scripts/dep-guard.sh
- Summary: Reviewed run run_20260624_085539 with gate status needs_review and review status approved. Goal: Slice 2: implement the debate policy layer in internal/debate on top of the engine, exercised only with the mock backend. (1) internal/debat...
- Candidate: runs/run_20260624_085539/memory/memory-candidate.json

### mem_005 - Slice 3: implement persona loading, .heurema/debate workspace discovery, conf...
- Run: run_20260624_092233
- Freshness: fresh
- Files: go.mod, go.sum, internal/debate/config/config.go, internal/debate/config/config_test.go, internal/debate/persona/persona.go, internal/debate/persona/persona_test.go, scripts/check-gofmt.sh
- Summary: Reviewed run run_20260624_092233 with gate status needs_review and review status approved. Goal: Slice 3: implement persona loading, .heurema/debate workspace discovery, config, and panel selection in internal/debate, fixture-tested only...
- Candidate: runs/run_20260624_092233/memory/memory-candidate.json

### mem_006 - Slice 4: wire the cmd/debate CLI into a working debate on a deterministic off...
- Run: run_20260624_095835
- Freshness: fresh
- Files: cmd/debate/e2e_test.go, cmd/debate/main.go, docs/PACTUM-ISSUES-ANALYSIS.md, internal/debate/runner/runner.go, internal/debate/runner/runner_test.go, internal/engine/transport/echo/echo.go
- Summary: Reviewed run run_20260624_095835 with gate status needs_review and review status approved. Goal: Slice 4: wire the cmd/debate CLI into a working debate on a deterministic offline backend (no real models yet). The command debate "<task>" ...
- Candidate: runs/run_20260624_095835/memory/memory-candidate.json

### mem_007 - Implement an ACP backend transport under internal/backend/acp using github.co...
- Run: run_20260624_103219
- Freshness: fresh
- Files: cmd/debate/e2e_test.go, cmd/debate/main.go, docs/DESIGN.md, go.mod, go.sum, internal/backend/acp/acp.go, internal/backend/acp/acp_test.go, internal/backend/acp/integration_test.go
- Summary: Reviewed run run_20260624_103219 with gate status needs_review and review status approved. Goal: Implement an ACP backend transport under internal/backend/acp using github.com/coder/acp-go-sdk and wire it into the cmd/debate production r...
- Candidate: runs/run_20260624_103219/memory/memory-candidate.json

### mem_008 - Implement an exec backend transport under internal/backend/exec (standard lib...
- Run: run_20260624_120335
- Freshness: fresh
- Files: cmd/debate/e2e_test.go, cmd/debate/main.go, docs/DESIGN.md, internal/backend/exec/exec.go, internal/backend/exec/exec_test.go, internal/backend/exec/integration_test.go
- Summary: Reviewed run run_20260624_120335 with gate status needs_review and review status approved. Goal: Implement an exec backend transport under internal/backend/exec (standard library only) that drives stateless plain-CLI agents like Gemini v...
- Candidate: runs/run_20260624_120335/memory/memory-candidate.json

### mem_009 - Add debate init and debate new scaffolding subcommands to cmd/debate that cre...
- Run: run_20260624_132440
- Freshness: fresh
- Files: cmd/debate/main.go, cmd/debate/scaffold.go, cmd/debate/scaffold_test.go, docs/DESIGN.md
- Summary: Reviewed run run_20260624_132440 with gate status needs_review and review status approved. Goal: Add debate init and debate new scaffolding subcommands to cmd/debate that create a ready-to-run .heurema/debate workspace and new persona fi...
- Candidate: runs/run_20260624_132440/memory/memory-candidate.json

### mem_010 - Remove the context.md baseline-preamble feature so debate context lives only ...
- Run: run_20260624_152405
- Freshness: fresh
- Files: cmd/debate/e2e_test.go, cmd/debate/scaffold.go, cmd/debate/scaffold_test.go, internal/debate/config/config.go, internal/debate/config/config_test.go, internal/debate/runner/runner.go, internal/debate/runner/runner_test.go
- Summary: Reviewed run run_20260624_152405 with gate status needs_review and review status approved. Goal: Remove the context.md baseline-preamble feature so debate context lives only in the task (plus the grounded sandbox): drop config.Workspace....
- Candidate: runs/run_20260624_152405/memory/memory-candidate.json

### mem_011 - Invoke agy non-interactively via --print in the exec backend so it works agai...
- Run: run_20260624_161104
- Freshness: fresh
- Files: internal/backend/exec/exec.go, internal/backend/exec/exec_test.go, internal/backend/exec/integration_test.go
- Summary: Reviewed run run_20260624_161104 with gate status needs_review and review status approved. Goal: Invoke agy non-interactively via --print in the exec backend so it works against the real agy CLI (which otherwise defaults to an interactiv...
- Candidate: runs/run_20260624_161104/memory/memory-candidate.json
