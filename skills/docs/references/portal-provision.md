# AWS Provisioning

Provision or adopt the standard static docs-portal stack: **private S3 bucket, served
exclusively through CloudFront via Origin Access Control, with a viewer-request CloudFront
Function** for directory-index rewrites. Reference implementation: the NGR portal
(`d1ab5hbiwe071c.cloudfront.net`, account 779053260276, profile `ngr-dev`).

## Non-negotiables

- All resource IDs live in `infra/deploy.config.json` — NEVER hardcoded in script bodies.
  Orphaned resources happen when IDs live only in a chat session.
- Every provisioning step is **check → skip-or-create** (idempotent). Re-running against a
  live stack must be a no-op.
- The bucket is private: block ALL public access, no S3 website configuration. Access only
  via the OAC-conditioned bucket policy.
- Verify identity first: `aws sts get-caller-identity` must match the config's account,
  abort otherwise.

## Stack shape (create in this order — later steps need earlier ARNs)

1. **S3 bucket** — private, public access blocked, no website config.
2. **CloudFront Function** (viewer-request, publish after create) — appends `index.html` to
   URIs ending in `/`. Canonical code (works for VitePress with `cleanUrls: false`):

```javascript
function handler(event) {
  var request = event.request;
  if (request.uri.endsWith('/')) {
    request.uri += 'index.html';
  }
  return request;
}
```

3. **OAC** — `sigv4` / `always` / origin type `s3`.
4. **Distribution** — from a JSON template with placeholders (`__BUCKET__`, `__OAC_ID__`,
   `__FUNCTION_ARN__`). Standard values: PriceClass_100, http2, DefaultRootObject
   `index.html`, redirect-to-https, Compress true, managed CachingOptimized policy
   `658327ea-f89d-4fab-a63d-7e88639e58f6`, CustomErrorResponses **403→/404.html (code 404,
   TTL 10s)** and **404→/404.html (code 404, TTL 10s)**. The 403 mapping is required — a
   private S3 REST origin returns 403 for missing keys, not 404.
5. **Bucket policy** — LAST, because it conditions on the distribution ARN:
   `Principal: cloudfront.amazonaws.com`, `Action: s3:GetObject`,
   `Condition StringEquals AWS:SourceArn = <distribution ARN>`.
6. Write all created IDs back into `deploy.config.json`; print a summary table.

## Deploy script contract (`infra/deploy.sh`)

1. Identity guard (same as above).
2. Build the site; abort if `dist/index.html` missing.
3. Two-pass sync — order matters so cached HTML never references deleted assets:
   - Pass A: hashed immutable assets → `--cache-control "public,max-age=31536000,immutable"`
   - Pass B: everything else, `--delete`, → `--cache-control "public,max-age=300,must-revalidate"`
4. `aws cloudfront create-invalidation --paths "/*"` (one billable path; don't micro-optimize).
5. Print live URL.

Also provide `teardown.sh` (disable dist → wait → delete dist → function → OAC → empty+delete
bucket) gated behind `--yes-really` plus typed bucket-name confirmation.

## Adoption mode (existing orphaned stack)

When the stack already exists: discover it (`aws cloudfront list-distributions` filtered by
domain → origin bucket → `get-distribution-config`, `get-function`, `get-bucket-policy`,
`get-origin-access-control`), record everything in `deploy.config.json`, and verify
`provision.sh` no-ops. Before ANY first redeploy, `aws s3 sync` the live site down as a
committed rollback copy and get explicit user approval.

## Acceptance checklist

- [ ] `provision.sh` exits 0 with no changes against an already-provisioned account
- [ ] `deploy.sh --dry-run` prints exact aws commands without executing
- [ ] Full deploy visible at the CloudFront URL within ~1 min
- [ ] No resource ID in any script body
- [ ] Scripts pass `shellcheck`
