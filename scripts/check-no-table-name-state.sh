#!/usr/bin/env sh
set -eu

file="internal/debate/config/config.go"

if matches=$(
  awk '
    /^type table struct[[:space:]]*\{/ { in_table_struct = 1 }
    in_table_struct && /^[[:space:]]*Name[[:space:]]+string([[:space:]]|$)/ {
      print FILENAME ":" FNR ":" $0
    }
    in_table_struct && /^\}/ { in_table_struct = 0 }

    /(^|[^[:alnum:]_])table[[:space:]]*\{/ { in_table_literal = 1 }
    in_table_literal && /(^|[[:space:]])Name[[:space:]]*:/ {
      print FILENAME ":" FNR ":" $0
    }
    in_table_literal && /\}/ { in_table_literal = 0 }
  ' "$file"
); then
  if [ -n "$matches" ]; then
    printf '%s\n' "$matches"
    echo "table name state must not be stored in internal/debate/config" >&2
    exit 1
  fi
else
  exit "$?"
fi
