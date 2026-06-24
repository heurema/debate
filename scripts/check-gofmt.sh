#!/usr/bin/env bash
# Check that all Go files under internal, cmd, and scripts are gofmt-clean.
# Exits non-zero and prints offending files when any are unformatted.
set -uo pipefail

out=$(gofmt -l internal cmd scripts 2>&1)
if [ -n "$out" ]; then
    printf 'unformatted Go files:\n%s\n' "$out" >&2
    exit 1
fi
