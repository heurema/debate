# Corrective Reviewer Prompt

This prompt is prepared for a corrective reviewer attempt.
Your previous response did not include a valid `pactum.reviewer_findings.v1alpha1` JSON block.
Findings expressed only in prose in the previous attempt are not recoverable; re-review the task using the inputs below.

## Inputs
- Reviewer context: .heurema/pactum/runs/run_20260624_162301/review/reviewer-context.md
- Contract: .heurema/pactum/runs/run_20260624_162301/contract/contract.json
- Gate report: .heurema/pactum/runs/run_20260624_162301/gate/gate-report.json
- Review artifacts: .heurema/pactum/runs/run_20260624_162301/review/review.json, .heurema/pactum/runs/run_20260624_162301/review/findings.jsonl, .heurema/pactum/runs/run_20260624_162301/review/resolutions.jsonl, .heurema/pactum/runs/run_20260624_162301/review/proposals.jsonl, .heurema/pactum/runs/run_20260624_162301/review/proposal-decisions.jsonl

## Review lens: Implementation vs contract

You are the implementation reviewer. Review the task against the approved contract and gate report, focusing on your lens:
- Does the diff achieve the contract goal?
- Is every in-scope item and acceptance criterion covered?
- Is wiring and integration complete (components registered, configs updated)?
- Are there missing pieces that prevent the change from working end to end?

## Required structured output

You MUST emit exactly one fenced JSON block. If you have no findings, emit `"findings": []`.

```json
{
  "schema": "pactum.reviewer_findings.v1alpha1",
  "findings": []
}
```
