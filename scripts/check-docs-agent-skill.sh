#!/usr/bin/env sh
set -eu

docs="README.md docs/DESIGN.md"
skill_dir="internal/debate/skills/bundled/debate"
skill_md="$skill_dir/SKILL.md"

if [ "$(basename "$skill_dir")" != "debate" ]; then
  echo "bundled skill parent directory must be named debate: $skill_dir" >&2
  exit 1
fi

if [ ! -f "$skill_md" ]; then
  echo "bundled skill is missing SKILL.md: $skill_md" >&2
  exit 1
fi

if [ "$(head -n 1 "$skill_md")" != "---" ]; then
  echo "$skill_md must start with YAML frontmatter" >&2
  exit 1
fi

if ! grep -q '^name: debate$' "$skill_md"; then
  echo "$skill_md frontmatter must set name: debate" >&2
  exit 1
fi

if ! grep -q '^description: .' "$skill_md"; then
  echo "$skill_md frontmatter must have a non-empty description" >&2
  exit 1
fi

for needle in \
  '.agents/skills/debate' \
  '.claude/skills/debate' \
  '.debate-skill.json' \
  'checksum' \
  'idempotent' \
  'local edit' \
  'debate init'
do
  if ! grep -R -- "$needle" $docs >/dev/null; then
    echo "docs are missing agent skill install marker: $needle" >&2
    exit 1
  fi
done

# These must never appear as a documented command (start of a usage line),
# though prose may still explain that no such command or target exists.
for forbidden in \
  '^debate skills' \
  '^debate install-skills' \
  '^debate status' \
  '^debate doctor' \
  '^debate clean' \
  '^debate upgrade' \
  '^- `~/\.codex/skills' \
  '^- `~/\.gemini/skills'
do
  if matches=$(grep -R -n -E -- "$forbidden" $docs); then
    printf '%s\n' "$matches"
    echo "docs must not document a v1 feature that does not exist: $forbidden" >&2
    exit 1
  else
    status=$?
    if [ "$status" -gt 1 ]; then
      exit "$status"
    fi
  fi
done
