#!/usr/bin/env bash
# Enforce that a package pattern's non-stdlib dependencies are under allowed import prefixes.
# Usage: dep-guard.sh <package-pattern> <allowed-prefix>...
set -uo pipefail

if [ "$#" -lt 2 ]; then
    printf 'usage: dep-guard.sh <package-pattern> <allowed-prefix>...\n' >&2
    exit 1
fi

pattern=$1
shift

deps=$(go list -deps -test -f '{{if not .Standard}}{{.ImportPath}}{{end}}' "$pattern" 2>&1)
rc=$?
if [ "$rc" -ne 0 ]; then
    printf '%s\n' "$deps" >&2
    exit "$rc"
fi

offending=()
while IFS= read -r dep; do
    [ -z "$dep" ] && continue
    ok=false
    for prefix in "$@"; do
        case "$dep" in
            "${prefix}"*) ok=true; break ;;
        esac
    done
    [ "$ok" = "false" ] && offending+=("$dep")
done <<< "$deps"

if [ "${#offending[@]}" -gt 0 ]; then
    printf 'dep-guard: forbidden dependencies in %s:\n' "$pattern" >&2
    printf '%s\n' "${offending[@]}" >&2
    exit 1
fi
