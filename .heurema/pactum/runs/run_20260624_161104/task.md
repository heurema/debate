# Task

Fix the agy exec backend to invoke agy non-interactively. The real agy CLI defaults to an interactive session and only runs a single prompt non-interactively under --print (alias -p), so the current default argv [agy, --model, spec.Model] hangs against real agy. Change internal/backend/exec/exec.go so the default argv becomes [agy, --print, --model, spec.Model], keeping the prompt on stdin so agy prints the response and exits. The DEBATE_AGY_COMMAND override still replaces only argv[0] and preserves --print and --model. Update the affected unit tests (exec_test.go argv assertions) and the gated integration test to match the new argv. Out of scope: other backends, internal/engine, the stdin reconstruction/accumulation logic, and the acp backend.

Generated: 2026-06-24T16:11:04Z
