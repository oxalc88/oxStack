---
name: docs
description: Documentation portal pipeline — analyze a codebase into documentation-grade knowledge, generate/style portal pages, verify them, translate to Spanish, provision the AWS hosting stack, and orchestrate incremental updates. Use for any docs-portal task — generating or updating technical documentation from a codebase (serverless APIs, Angular apps), writing a docs page or HTML artifact, verifying docs before deploy, updating the /es/ mirror, or standing up the portal's AWS infra.
---

# Docs Portal Pipeline

One skill, tiered: this file is the always-loaded orchestrator (the everyday `/docs`
entry point). Each pipeline stage has its own reference file below — load only the one(s)
the task needs.

## When to use each reference

| Task | Reference |
|------|-----------|
| Code changed, portal needs to catch up — the default entry point | This file (Pipeline, below) |
| Analyze an API/app/module into documentation-grade knowledge (lambda summaries, cited flows, dead-code discrimination) | [references/code-analyze.md](references/code-analyze.md) |
| Generate/regenerate portal pages from committed knowledge JSON | [references/page-generate.md](references/page-generate.md) |
| Write any docs page, HTML artifact, or diagram — progressive disclosure, citations, SVG | [references/artifact-style.md](references/artifact-style.md) |
| Verify docs before deploy, or audit an existing portal for drift | [references/verify.md](references/verify.md) |
| Translate English pages to the `/es/` mirror | [references/translate-es.md](references/translate-es.md) |
| Stand up or adopt the docs-portal AWS stack (S3 + CloudFront) | [references/portal-provision.md](references/portal-provision.md) |

`page-generate` loads `artifact-style` alongside it. `update` (this file) drives
`code-analyze` → `page-generate` → `verify` → `translate-es` in sequence — you rarely need to
invoke a downstream reference without it.

## Pipeline (incremental orchestrator)

Everyday entry point to keep a generated docs portal in sync with its source repos. Runs the
minimal slice of the pipeline, not a full rebuild.

### 1. Detect
- Read `knowledge/state.json` (`{project: lastDocumentedCommit}`) in the portal repo.
- Per documented project: `git diff --name-only <lastDocumented>..HEAD`.
- Map changed files → affected units and stages:
  - change under `<api>/` → that API; touching `yml/functions.yml`/`serverless.yml` → also
    endpoints page and re-check of the wired/dead-code gates (a newly-unwired handler must
    drop out of the docs).
  - change under an app's `services/` or `environment*.ts` → its api-usage page + edges.
  - **No-op signals** (do not trigger regen): `requirements/`, `diagnosis/`, `*.md`, tests,
    the portal repo itself.
- Emit `changeset.json`: `{affected: [...], stages: [...], reason: [...]}`. Empty changeset →
  report "up to date", stop.

### 2. Regenerate (only the slice)
- Re-run deterministic extraction for affected units; compare `contentHash` — unchanged hash
  means the change was cosmetic, drop it from the changeset.
- Re-run analysis ([code-analyze](references/code-analyze.md)) only for units whose hash
  changed. Curation file still applies; NEW channels/integrations detected → flag for human
  curation, don't guess active/unused.
- Regenerate only the affected pages ([page-generate](references/page-generate.md)) + any page
  that renders shared data that changed (system map, endpoint tables, appendix).

### 3. Verify
- Run [verify](references/verify.md)'s deterministic layer on the whole build (cheap),
  adversarial layer only on regenerated flagship pages. Blockers stop the run BEFORE deploy.

### 4. Translate
- [translate-es](references/translate-es.md) for regenerated pages only.

### 5. Deploy & record
- `infra/deploy.sh` (build → two-pass S3 sync → CloudFront invalidation) — see
  [portal-provision](references/portal-provision.md) for the deploy script contract.
- Update `knowledge/state.json` to the documented commits.
- Write a run report to `generator/runs/<date>/`: changeset, regenerated pages, exclusion
  flags, `needs-human` list, deploy result. Surface the `needs-human` list to the user —
  never bury it.

## Rules

- Batch, don't spam: at most a few deploys per day; group pending changes.
- A verify blocker is never "fixed" by editing built output — fix knowledge/analysis and
  regenerate.
- Weekly (or on demand) run a FULL pass instead of incremental as a drift catch-all: if full
  extraction differs from committed knowledge but no changeset ever fired, change detection
  has a gap — report it.
- All git state changes in the portal repo (knowledge JSON, docs, state.json) get committed;
  delegate commits per the user's committer convention.
