# Agent progress stream

Run stdout is final-result-only. Progress is emitted on stderr by default, one event per
line. Every machine-readable progress line begins exactly with:

```text
@@DEBATE_PROGRESS
```

followed by one JSON object with `version: 1`, `type`, `stage`, and `elapsed_ms`.
Consumers should ignore additional fields and ignore stderr lines without the prefix.
`--json` changes only stdout formatting; it does not suppress progress. `--quiet`
suppresses all progress event lines.

V1 event types: `run_started`, `workspace_loaded`, `session_opening`, `session_opened`,
`round_started`, `turn_started`, `heartbeat`, `turn_completed`, `round_completed`,
`synthesis_started`, `synthesis_completed`, `run_completed`, `run_failed`.

Stage mapping:

| Event type | Stage |
|---|---|
| `run_started`, `workspace_loaded` | `loading_workspace` |
| `session_opening`, `session_opened` | `opening_session` |
| `round_started`, `round_completed` | `running_round` |
| `turn_started`, `turn_completed` | `running_turn` |
| participant `heartbeat` | `running_turn` |
| synthesizer `heartbeat` | `synthesizing` |
| `synthesis_started`, `synthesis_completed` | `synthesizing` |
| `run_completed` | `completed` |
| `run_failed` | active lifecycle stage when known, otherwise `failed` |

Event-specific required fields:

| Event type | Required fields |
|---|---|
| `session_opening` | `speaker` |
| `session_opened` | `speaker`, `duration_ms` |
| `round_started` | `round` |
| `turn_started` | `round`, `speaker` |
| participant `heartbeat` | `round`, `speaker`, `silence_ms` |
| synthesizer `heartbeat` | `silence_ms` |
| `turn_completed` | `round`, `speaker`, `duration_ms` |
| `round_completed` | `round`, `duration_ms` |
| `synthesis_completed` | `duration_ms` |
| `run_completed` | `duration_ms` |
| `run_failed` | `error` |

Lifecycle ordering is deterministic for successful runs: one `run_started`, one
`workspace_loaded`, matching session open events per participant, matching round/turn
start/completion events, one synthesis start/completion pair, and a final
`run_completed`. On failure, at most one `run_failed` is emitted and it is the final
progress event.
