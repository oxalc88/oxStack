# Docs Artifact Style

Quality bar defined by three reference pages (open them before writing if unfamiliar):
- `thariqs.github.io/html-effectiveness/04-code-understanding.html` — walkthroughs
- `thariqs.github.io/html-effectiveness/10-svg-illustrations.html` — inline SVG
- `thariqs.github.io/html-effectiveness/13-flowchart-diagram.html` — flowcharts

## Page structure — progressive disclosure

1. Open with a **small conceptual diagram** of the flow/system. Never start with a wall of
   code or an exhaustive table.
2. Then the **numbered walkthrough** (steps mirror actual runtime order: client → handler →
   logic → data store).
3. Then key files / reference tables.
4. End with a **Gotchas** section: real hazards (throttling, TTL quirks, prod-only flags),
   sourced from incidents or code comments — never speculative filler.

One page = one question a developer actually asks ("how does an order reach the POS?").

## Code walkthroughs

- Each step: 2–4 sentences of prose + a collapsible "show source" block. Reader can skim
  prose without expanding anything.
- **Precise citations always**: `repo/path/file.js:22-48` on every step and claim. Citation
  without having read the file = forbidden.
- **Selective emphasis**: the trust boundary / most critical file gets the most prominent
  treatment; everything else stays brief. Uniform depth is a smell.
- Source excerpts are embedded at generation time (mechanically, from the citation), never
  paraphrased/retyped by hand.

## Diagrams

- **Inline SVG, self-contained** — no raster, no external CDN, no fetch.
- **Theme-aware**: `currentColor` + theme CSS variables (in VitePress:
  `var(--vp-c-text-1)`, `var(--vp-c-brand-1)`, `var(--vp-c-bg-soft)`) so dark mode works.
- Label directly on the diagram; avoid separate legends.
- ≤ ~15 nodes per SVG; split rather than cram.
- **Flowcharts** for branching logic (decision diamonds); **swimlanes** when a flow crosses
  systems. Every node representing code cites its file.
- Tiering rule: Mermaid is acceptable for bulk-generated per-module flowcharts; **flagship
  diagrams** (system map, main lifecycle, auth, architecture anatomy) must be deliberately
  laid-out authored SVG.

## Writing rules

- Complete sentences; explain in prose, enumerate in tables (endpoints, env vars, tables) —
  never the reverse.
- No implementation claim without a citation.
- Translations (e.g. `/es/`) translate prose only — NEVER identifiers, file paths, or code.
- Footer every generated page: "Generated from `<project>@<short-sha>` on `<date>`".

## Component vocabulary (when the target is a VitePress portal)

`<CallstackStep n file lines>` (numbered step + collapsible source) · `<SystemMap highlight>`
(ecosystem SVG from edges data) · `<FlowChart>` (inline SVG wrapper) · `<Gotcha level>` ·
`<EndpointTable api>` (rendered from knowledge JSON, never hand-written). Build once in
`.vitepress/theme/components/`, reuse everywhere.
