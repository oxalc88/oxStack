# Docs Verification

Gate between generation and deploy. Two layers: deterministic checks (scriptable, run first,
cheap) and adversarial LLM review (expensive, targeted). A page ships only when both pass.

## Layer 1 — Deterministic (run/write scripts; fail the build on any miss)

- **Citation existence**: every `file:line` cited on any page resolves to a real file, and
  the line range exists at the referenced commit.
- **Excerpt match**: every embedded source excerpt is byte-identical to the cited lines
  (excerpts are injected mechanically, so a mismatch means stale knowledge JSON → re-extract).
- **Link integrity**: every internal link and anchor resolves within the same build; the
  static build itself (`vitepress build`) exits 0.
- **Locale parity**: en and es page trees are isomorphic — same paths, same heading anchors.
- **Dead-code leak**: nothing listed in `curation.json` as `unused`/`excludePaths` appears in
  any flow, endpoint table, or diagram (grep the built output for excluded handler names and
  channel identifiers, e.g. glovo/uber handlers in NGR).
- **Coverage**: every unit the inventory says is active has its section; the "Legacy & unused"
  appendix lists every exclusion flagged during analysis.
- **Freshness footer** present on every generated page.

## Layer 2 — Adversarial review (LLM, per flagship page)

For each flagship page, run a skeptic pass whose explicit goal is to **refute** the page
against the actual code (this is the one verification stage allowed to read the repos):

- For each claim: open the cited file. Does the code actually do what the prose says? Check
  branch conditions, error paths, and stage-specific behavior (a claim true in dev but not
  prod is WRONG — e.g. flags applied only for prod).
- For each walkthrough: is the step order the real runtime order? Would a request actually
  traverse these files? Is a middleware/interceptor step missing?
- For each diagram: does every arrow correspond to an edge with evidence? Any invented
  integration?
- For gotchas: is each one grounded (incident doc, code comment, config) or speculative?

Verdict per claim: CONFIRMED / WRONG (with the contradicting `file:line`) / UNVERIFIABLE.
WRONG blocks deploy; UNVERIFIABLE goes to the run report's `needs-human` list. Default to
skepticism: a claim you cannot ground is not CONFIRMED.

## Output

A run report: `{page, checks: {...}, claims: [{text, verdict, evidence}], blockers[]}`.
Deploy proceeds only with zero blockers. Never fix a WRONG claim by editing the page prose
directly — fix the knowledge JSON (or the analysis) and regenerate, or the next regeneration
reintroduces the error.
