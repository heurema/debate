# Panel guidance

The default scaffold (`debate init`) creates exactly two debater personas and one flat
table:

- **Proposer** — builds the strongest practical solution to the task and defends it
  against objections, but revises the proposal when the Skeptic surfaces a real blocker
  rather than defending a broken position.
- **Skeptic** — finds blocking risks: weak assumptions, unhandled edge cases, missing
  validation, and failure modes that would break the proposal in practice. Distinguishes
  blockers (must fix before this can ship) from nice-to-haves.

This proposer/skeptic pair is a good default for engineering decisions: implementation
plans, design reviews, "should we do X" questions, and risk assessments of a change.

## When to add more personas

Add a persona with `debate new <name>` (or `debate new <namespace>/<name>` for a
namespaced one, e.g. `reviewers/security`) when a distinct perspective would surface
different objections than the Skeptic alone — for example a security-focused reviewer,
a performance-focused reviewer, or a domain expert. Wire them into a table's `panel` list
alongside or instead of the defaults, or select them ad hoc with `--with`.

## Choosing a synthesizer

Most tasks don't need a dedicated synthesizer persona — the built-in default (matching
the panel's backend family, or the first detected local runtime) is a neutral summarizer
and is usually sufficient. Add a `synthesizer`-role persona (`debate new final-judge
--role synthesizer`) only when the project wants a different model, backend, or
synthesis style than the panel's default, and reference it via a table's `synth` field
or `--synth` at the command line.

## Choosing a table vs. `--with`

Use a table (`tables/<name>.yml`) for a panel composition that will be reused. Use
`--with persona1,persona2` for a one-off panel without editing a table file.
