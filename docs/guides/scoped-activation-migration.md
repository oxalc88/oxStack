---
summary: Preview and safely retire identifiable legacy Codex global activation without claiming unknown files.
read_when: Migrating an old oxStack installation or reviewing remaining global context.
---

# Legacy Codex migration

Run the updated source, not the old ignored repository binary:

```bash
cd /Users/lannister/projects/oxDeveloop/oxStack
go run ./cmd/oxstack migrate --dry-run
```

Preview reports REMOVE GLOBAL ACTIVATION, KEEP GLOBAL, MOVE TO PROJECT,
MOVE TO PROFILE and REVIEW items. No backup, manifest, config or skill is written
in preview. Project/profile moves are suggestions: the command cannot infer
which projects should receive capabilities.

When you approve the displayed changes, run:

```bash
go run ./cmd/oxstack migrate --apply
```

Type `MIGRATE` at the prompt. There is no unattended `--yes` bypass. The plan
is rechecked against current files before writing. The command:

- Removes only MCP entries matching the frozen old generator signatures, with
  the historical oxStack marker present. Other entries anywhere after the marker
  remain intact. Changed/unknown entries are reported for review, not removed.
- Compares the old generated global AGENTS file to the exact old guideline
  concatenation. Only an exact match is replaced with minimal universal guidance.
- Removes only old global skill links pointing exactly at this oxStack checkout's
  corresponding `skills/<name>` directory.
- Preserves copied Cloudflare skills, external skill installations, legacy
  `$CODEX_HOME/skills` content, user files and other uncertain resources.
- Stores private backups of original file bytes and symlink metadata under
  `$XDG_STATE_HOME/oxstack/migration-backups/<timestamp>`, then records new ownership.

The old CloudFormation signature used an AWS profile substituted from oxStack's
`.env`. Migration reads that historical value solely to prove identity. If it is
missing or differs, the server is reported for review. No new portable config
receives that machine-specific value. No environment file is modified.

## Remaining review items

A REVIEW line is deliberately unresolved. Do not interpret a successful partial
migration as proof that the effective global catalog is clean. Establish source
and ownership before retiring these resources individually. In particular, the
copied Cloudflare duplicates and third-party global skills from other installers
cannot safely be deleted from their names alone. Their removal needs a separate
reviewed change; this migration neither adopts nor weakens checks to delete them.

The command does not migrate Claude/OpenCode config, account-installed plugins,
Orca/herdr hooks, accumulated approval rules or desktop app defaults. Coordinate
those owners separately. It refuses an existing new global ownership manifest;
subsequent changes use scoped install/uninstall instead of repeating legacy adoption.

## Activate chosen replacements

```bash
go run ./cmd/oxstack install --global
go run ./cmd/oxstack install --project /path/to/lambda-repo \
  --selection selections/projects/aws-lambda.toml
go run ./cmd/oxstack install --profile aws-tools
```

Trust the AWS repository in Codex. Do not add project trust entries to unrelated
projects automatically. Leave optional profiles unselected during non-AWS work.

Restart relevant apps/daemons normally and use fresh threads. Existing sessions
may retain old context. Account/plugin tools remain separate from local MCP TOML.

## Verify

```bash
codex -C /path/to/non-aws-project mcp list --json |
jq -e '[.[] | select(.name | test("aws|bedrock|powertools|serverless"; "i"))] | length == 0'

python3 scripts/verify-clean-project.py /path/to/non-aws-project
codex --profile aws-tools -C /path/to/non-aws-project mcp list --json
```

Codex 0.161.0 `mcp list` project discovery uses the actual working directory;
repeat the shell check after `cd /path/to/non-aws-project`. The verification
script sets that directory explicitly. This matters when checking local servers.

The first two checks should pass only after all relevant global REVIEW items
are resolved. The explicitly selected profile should expose pricing and diagrams.
Check `/skills` and `/mcp` in a fresh desktop thread as well.

## Recovery

For a normal write failure, the operation attempts to restore only destinations
that still match its own writes. If another process changed a destination,
rollback stops claiming that file and reports the conflict. Keep the original
error and inspect each path before retrying.

For legacy migration, each backup JSON includes the original absolute path,
bytes (base64 encoded), mode, existence and symlink target. Restore only reviewed
paths from those snapshots, with current-state checks. Do not run the old installer
or uninstall to roll back. Multi-file operations are not crash-atomic: a killed
process or power loss requires manual inspection before a retry.

## Preserve reviewed third-party installations

After individually reviewing the remaining copied skills, use `catalog import
NAME --source DIR --snapshot` to retain an installed copy whose Git revision is
unknown. Declare the capability in `oxstack.toml` first. Known installer repository
metadata is evidence of origin, not proof of a Git revision; use `UNKNOWN` when
origin is unverified. The imported copy remains inactive until explicitly selected.

Before retiring a global copy, compare every file's bytes and permissions against
the preserved catalog. Archive the original outside all discovery directories;
keep a private inventory containing original/destination paths, file hashes,
permissions and old/new dependent symlink targets. Retarget existing other-runtime
links only after verification. Leave installer lock metadata in place as historical
evidence. This reviewed relocation is separate from `migrate`, which continues to
refuse ownership of unknown copies.

To recover, first verify archived files against that inventory and require that
original destinations are absent. Restore the original directories, then restore
only dependent symlinks that still match the recorded replacement targets. Stop
on any conflict; never overwrite later user changes. Catalog copies and archives
are local data and require backups independently of Git.
