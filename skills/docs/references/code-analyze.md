# Code Analysis

Produce documentation-grade analysis of a codebase unit (one serverless API, one Angular app,
one service). Output feeds page generation — so every claim must be citable and every
documented element must be **actually in use**.

## Iron rules

1. **Never cite a `file:line` you did not open.** Every flow step, every claim, anchors to a
   file you actually read in this session. A verifier will check excerpt-vs-file match.
2. **Extract structure from config, not prose from memory.** For serverless: `serverless.yml`
   + `yml/functions.yml` (lambdas, events, routes), `yml/resources.yml` (owned tables/queues/
   buckets), `yml/iam.yml` + `yml/settings/*.yml` (touched ARNs, env vars per stage). For
   Angular: `src/**/*.service.ts` HttpClient calls + `environment*.ts` base URLs.
3. **Document only live code** — apply the dead-code discrimination below to EVERYTHING
   before analyzing it.
4. If a `generator/curation.json` exists, it is LAW: channel status (`active`/`unused`),
   version selection (`latest` only), `excludePaths`. Your findings feed it; they don't
   override it.

## Dead/old code discrimination (run BEFORE documenting anything)

A file/handler/flow is documented only if it passes ALL of these gates:

| Gate | Check | Verdict if failed |
|------|-------|-------------------|
| **Wired** | Handler file is referenced by a function entry in `serverless.yml`/`yml/functions.yml` (or, for frontends, the route/module is reachable from the app router) | Not deployed → EXCLUDE, flag |
| **Not commented out** | The function entry / route / require is not commented out in yml or code; the code block itself isn't a giant commented region | EXCLUDE, flag |
| **Naming markers** | No `old`/`Old`/`OLD` prefix/suffix, no `.bak`, `-copy`, `_deprecated`, no residence in an `old/` folder | EXCLUDE, flag |
| **Reachable** | File is transitively `require`d/imported from at least one wired handler/entry point | Orphan module → EXCLUDE, flag |
| **Latest version** | When two implementations serve the same channel/feature (e.g. `cloudkitchen` v1 vs v2 handlers, `xV2.js` alongside `x.js`), only the version the wired routes actually hit is documented | Superseded → EXCLUDE, flag |
| **Curation** | Channel/integration not marked `unused` in curation.json (e.g. Glovo, Uber in NGR orders) | EXCLUDE per curation |

Supporting (not decisive) signals: env vars only defined for non-prod stages, git
last-modified long ago while a v2 sibling is active, TODO/DEPRECATED comments.

**Flag, never silently drop:** every exclusion goes into the run report as
`{path, gate, evidence}` so a human confirms, and the aggregate feeds the portal's single
"Legacy & unused integrations" appendix page — readers who find the dead code in the repo
must be able to see it was excluded deliberately.

**Ambiguity rule:** if gates disagree (wired but named `old`, or unreferenced but recently
modified), do NOT decide — put it in the run report's `needs-human` list.

## Output shape

Per analyzed unit, produce/update its knowledge JSON:

- `lambdas[]` — name, handler `file#fn`, events, one-line summary (from reading the handler,
  not guessing from the name).
- `tables[]/queues[]` — with access-pattern descriptions grounded in the `sql/dynamodb/*.js`
  code that touches them.
- `flows[]` — the key walkthroughs: `{id, title, steps: [{file, lines, note}]}` following the
  real runtime order (handler → logic → sql → AWS). For channel-based APIs (like NGR orders),
  group flows **per channel**, each channel covering its full lifecycle (create → status →
  ready), latest version only.
- `exclusions[]` — the flagged dead/old code with gate + evidence.
- Architecture variant matters: functional pattern (handlers/logic/sql) vs hexagonal
  (application/domain/infrastructure — describe ports/adapters mapping) vs frontend
  (routes/services/API usage).

## Emphasis selection

Identify the ONE trust boundary / most dangerous file per unit (token validation, state
machine transition, payment call) — it gets the most prominent treatment downstream. Mine
incident/diagnosis docs in the repo for "gotchas": real hazards beat speculative ones.
