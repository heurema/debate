#!/usr/bin/env sh
set -eu

docs="README.md docs/DESIGN.md docs/SLICES.md"
docs_and_comments="$docs cmd/debate/e2e_test.go cmd/debate/main.go"

if grep -n 'config\.yml' $docs >/tmp/debate-docs-config-yml.txt; then
  cat /tmp/debate-docs-config-yml.txt
  echo "docs must not document config.yml for the v1 workspace" >&2
  exit 1
fi

for needle in \
  '.heurema/debate' \
  'personas' \
  'tables/default.yml' \
  'version: 1' \
  '--table' \
  '--with' \
  '--synth' \
  'namespace/name'
do
  if ! grep -R -- "$needle" README.md docs/DESIGN.md docs/SLICES.md >/dev/null; then
    echo "docs are missing v1 workspace marker: $needle" >&2
    exit 1
  fi
done

for forbidden in \
  'Hide debater output from other debaters' \
  'Hide project/web access' \
  'hidden cross-talk' \
  'defaultResolver returns error for unknown backend "agy"'
do
  if matches=$(grep -R -n -- "$forbidden" $docs_and_comments); then
    printf '%s\n' "$matches"
    echo "docs/comments contain stale v1 workspace wording: $forbidden" >&2
    exit 1
  else
    status=$?
    if [ "$status" -gt 1 ]; then
      exit "$status"
    fi
  fi
done
