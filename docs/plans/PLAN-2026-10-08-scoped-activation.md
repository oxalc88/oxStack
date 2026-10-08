---
summary: Separate the capability catalog from explicit runtime activation and protect ownership.
read_when: Changing installation, selections, runtime adapters, ownership, or migration.
status: in-progress
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
Status: In progress

Constraints: No live Codex/agents modifications or old uninstall. No automatic migration.

Validation: engine tests and race detector pass; native Codex profile smoke confirms explicit loading. Remaining: final docs/CLI smoke, vet, format, commits and one push.
