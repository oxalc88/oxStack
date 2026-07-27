# Page Generation

Turn knowledge JSON (produced by [code-analyze](code-analyze.md) / the extraction scripts)
into portal pages. Style rules come from [artifact-style](artifact-style.md) — load it
alongside this one.

## Iron rule: generate from the graph, not the repo

At this stage you do NOT re-read source repos. Pages are filled from the knowledge JSON so
they stay consistent with the graph (and with each other). The only mechanical exception:
source excerpts for `<CallstackStep>` blocks are injected by the generator script from the
`file:line` citations — guaranteed-real code, never retyped.

If the knowledge JSON lacks something a template needs, STOP and send the gap back to the
analysis stage — do not improvise content.

## Page templates

### Per-API section (`/api/<name>/`)
| Page | Content | Source fields |
|------|---------|---------------|
| `index` | What it does, business domain, lambda count, owned tables/queues/buckets, who consumes it | summary, lambdas.length, tables, queues, edges (incoming) |
| `architecture` | Layering walkthrough (handlers → logic → sql, or hexagonal ports/adapters) + anatomy SVG | pattern, require-graph |
| `flows/…` | Walkthrough pages. Channel-based APIs: one page per channel (`channels/<channel>/`) covering its FULL lifecycle (create → status updates → ready), latest version only | flows[], curation |
| `endpoints` | Generated table: name, method, path, handler file | lambdas[].events |
| `data` | Tables + GSIs + access patterns; queues | tables[], queues[] |

Hexagonal modules (e.g. NGR `/overflow`) get an extra `hexagonal` page mapping real folders
onto the ports/adapters hexagon.

### Per-app section (`/apps/<name>/`)
`index` (purpose, users, stack) · `api-usage` (which API endpoints it calls — table from
edges.json + arrows on the system map) · `flows` (1–2 user flows as flowcharts).

### Ecosystem section (`/ecosystem/`)
`system-map` (all projects + edges, narrative) · flagship cross-project lifecycle page
(e.g. an order end-to-end: channel → API → state machine → POS → frontends) · `auth` ·
`environments` (stages/regions/deploy model).

### Appendix (mandatory when curation excludes anything)
One "Legacy & unused integrations" page listing everything present in code but deliberately
undocumented (`{what, where, why excluded}`) — prevents readers from thinking the docs are
stale when they find dead handlers.

## Consistency requirements

- Sidebar/nav is generated from the same knowledge JSON — never hand-edited per page.
- Every internal link target must exist in the same build (verifier checks).
- en is the source of truth; `/es/` tree must stay isomorphic (same pages, same anchors) —
  translation is a separate stage ([translate-es](translate-es.md)), do not translate inline
  here.
- Every page footer: "Generated from `<project>@<short-sha>` on `<date>`".
