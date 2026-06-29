#!/usr/bin/env sh
set -eu

docs="README.md docs/DESIGN.md"

for needle in \
  '@@DEBATE_PROGRESS ' \
  'version: 1' \
  'run_started' \
  'workspace_loaded' \
  'session_opening' \
  'session_opened' \
  'round_started' \
  'turn_started' \
  'heartbeat' \
  'turn_completed' \
  'round_completed' \
  'synthesis_started' \
  'synthesis_completed' \
  'run_completed' \
  'run_failed' \
  'loading_workspace' \
  'opening_session' \
  'running_round' \
  'running_turn' \
  'synthesizing' \
  'completed' \
  'failed' \
  'duration_ms' \
  'silence_ms' \
  'elapsed_ms' \
  '1000 milliseconds' \
  '--quiet' \
  'stdout' \
  'stderr'
do
  if ! grep -R -- "$needle" $docs >/dev/null; then
    echo "docs are missing progress stream marker: $needle" >&2
    exit 1
  fi
done

for doc in $docs; do
  if ! grep -q -- 'final-result-only' "$doc"; then
    echo "$doc must document stdout as final-result-only" >&2
    exit 1
  fi
  if ! grep -q -- 'Stage mapping' "$doc"; then
    echo "$doc must document progress stage mapping" >&2
    exit 1
  fi
  if ! grep -q -- 'Event-specific required fields' "$doc"; then
    echo "$doc must document event-specific required fields" >&2
    exit 1
  fi
done
