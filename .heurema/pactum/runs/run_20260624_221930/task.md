# Task

Implement the ACP tool-contract hardening that follows the timeout guard: keep the existing open/prompt timeout safety boundary, but fix the root-cause risk in internal/backend/acp by making file-read behavior session-root scoped and bounded, making permission decisions conservative instead of blindly selecting the first option, making unsupported write/terminal operations fail clearly, and adding focused tests for grounded file reads, path escape denial, permission ordering, terminal requests, and timeout behavior. Do not change the public CLI or debate orchestration outside the ACP backend except for tests and necessary internal helpers.

Generated: 2026-06-24T22:19:30Z
