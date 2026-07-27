# Translation (en → es)

English is the source of truth; `/es/` is a generated mirror. Translation is the LAST
pipeline stage — never translate inline during page generation.

## Invariants (violating any of these breaks the build or the reader's trust)

- **Never translate**: identifiers, function/lambda names, file paths, code blocks, CLI
  commands, JSON keys, env var names, AWS resource names, URLs.
- **Tree isomorphism**: same file paths under `/es/`, same heading structure. Heading anchor
  slugs must keep working — if the framework derives anchors from heading text, keep explicit
  anchor IDs (`{#original-anchor}`) so en/es anchors match.
- **Frontmatter**: translate `title`/`description` values only; keep keys and everything else.
- Component props/attributes (`<CallstackStep file=… lines=…>`) untouched; only translate
  human-facing slot text.
- Footer and generated tables: translate column headers and labels, never the data cells that
  contain identifiers.

## Glossary discipline

Use the project glossary file (convention: `generator/glossary-es.json`) and apply it
consistently. Seed for the NGR domain:

| en | es |
|----|----|
| order | pedido |
| store | tienda / local (keep consistent per project — NGR uses "tienda") |
| channel | canal |
| flow | flujo |
| gotcha | advertencia |
| trust boundary | frontera de confianza |
| state machine | máquina de estados |
| walkthrough | recorrido |

Terms the team uses in Spanish already (pedido listo, carta, torre de control) stay in their
native Spanish form — do not round-trip them through English.

New recurring term without a glossary entry → add it to the glossary file in the same change,
so the next run stays consistent.

## Process

1. Diff-driven: translate only pages whose English source changed since the last run
   (content hash / changeset), plus any page missing under `/es/`.
2. Translate meaning, not words — natural technical Spanish (Latin American, "tú" avoided:
   use impersonal or "usted"-neutral constructions typical of tech docs).
3. After translating, run the locale-parity check from [verify](verify.md) (isomorphic trees,
   anchors resolve) before considering the stage done.
