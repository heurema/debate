#!/usr/bin/env bash
# Enforce that internal/engine/... imports only stdlib and other internal/engine packages.
set -uo pipefail

deps=$(go list -deps -test -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/engine/... 2>&1)
rc=$?
if [ "$rc" -ne 0 ]; then
    printf '%s\n' "$deps" >&2
    exit "$rc"
fi

# Non-standard packages that are NOT under our engine path are forbidden.
offending=$(printf '%s\n' "$deps" \
    | grep -v '^$' \
    | grep -v '^github\.com/heurema/debate/internal/engine' \
    || true)

if [ -n "$offending" ]; then
    printf 'dep-guard: forbidden dependencies in internal/engine/...:\n' >&2
    printf '%s\n' "$offending" >&2
    exit 1
fi
