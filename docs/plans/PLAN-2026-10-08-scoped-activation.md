---
summary: Separate the capability catalog from explicit runtime activation and protect ownership.
read_when: Changing installation, selections, runtime adapters, ownership, or migration.
status: done
---

# Scoped activation implementation

## Stage 1: Catalog and native scopes
Goal: Explicit selections and runtime adapters; no implicit installation.
Success: Global, project and on-demand selection behavior is tested, including inactive MCPs.
Status: Complete

## Stage 2: Ownership and lifecycle
Goal: Fingerprinted entries, safe install/update/uninstall, and migration preview.
Success: Unrelated settings survive; modified or unowned resources cannot be deleted.
Status: Complete

## Stage 3: Documentation and verification
Goal: Minimal global guidance, examples, migration runbook and effective-surface verification.
Success: Full tests, race detector, vet, formatting, isolated CLI smoke checks and atomic commits.
Status: Complete

Constraints: No live Codex/agents modifications or old uninstall. No automatic migration.

Validation: all behavior tests, race detector, vet, formatting and native Codex smoke pass. The smoke validates trusted AWS/non-AWS project configuration, explicit profiles, inactive pricing/CloudFormation, and zero base leakage. Python scripts parse cleanly. Implementation and migration are committed separately; final docs/verification commit completes the plan. One push follows verification.

Compatibility findings: Codex 0.161.0 mcp list uses actual cwd for project discovery despite -C; verification sets both. Inactive Claude servers remain unregistered. Profile skills are explicit references, not global discovery. Legacy resources without ownership proof remain REVIEW items.
