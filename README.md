---
summary: Maintain a capability catalog and explicitly activate it by scope and runtime.
read_when:
  - Installing or uninstalling oxStack capabilities.
  - Adding skills, MCP definitions, projects, or profiles.
  - Migrating a legacy global installation.
---

# oxStack

oxStack maintains reusable capabilities for Codex, Claude Code and OpenCode.
**Catalog != activation.** Adding a skill or MCP definition never installs it
into a runtime. Every installation needs an explicit scope and selection.

```text
skills/                         Maintained local skill catalog
catalog/external/               Reviewed, imported external snapshots (ignored)
mcp/*.toml                      Runtime-neutral server definitions
guidelines/*.md                 Reusable instruction templates
selections/global.toml          Explicit universal baseline
selections/projects/*.toml      Example project selections
selections/profiles/*.toml      Explicit on-demand selections
migration/legacy-catalog.toml   Frozen old signatures, not an active catalog
```

The default global selection contains only minimal universal instructions. It
selects **no skills and no MCP servers**. AWS, Astro, Cloudflare, web audits,
browser automation and domain workflows are never globally installed merely
because oxStack maintains them. You can explicitly select a capability globally,
but that deliberately exposes it across projects.

## Build and checks

Use the Go version declared in `.mise.toml`:

```bash
go test ./...
go test -race ./...
go vet ./...
go build -o /tmp/oxstack ./cmd/oxstack
```

Run from this repository, or set `OXSTACK_ROOT` to its absolute path. The CLI
resolves its repository through the executable location, `OXSTACK_ROOT`, or a
repository working directory. The old ignored `./oxstack` executable is not
rebuilt automatically: use `go run ./cmd/oxstack` until you rebuild your launcher.

`install` without a scope fails. It does not clone gstack, install npm/browser
packages, install agents, modify every runtime, or read `.env` for new activation.
`--runtime` defaults to `codex`; invoke each runtime deliberately.

## Global capabilities

```bash
go run ./cmd/oxstack install --global --dry-run
go run ./cmd/oxstack install --global
```

This manages a bounded universal instruction block in `$CODEX_HOME/AGENTS.md`.
Existing user text is preserved. Legacy generated instructions require migration
first. `CODEX_HOME` defaults to `~/.codex` and must be absolute.

A custom global selection can explicitly opt into skills or tools:

```bash
go run ./cmd/oxstack install --global --selection /path/to/global.toml
```

This selection replaces the **previous oxStack-owned selection** for that
scope/runtime, not the user's unrelated configuration.

## Project capabilities

```bash
go run ./cmd/oxstack install --project /path/to/lambda-repo \
  --selection selections/projects/aws-lambda.toml --dry-run
go run ./cmd/oxstack install --project /path/to/lambda-repo \
  --selection selections/projects/aws-lambda.toml
```

Codex receives selected skill links under `.agents/skills/` and selected MCPs in
`.codex/config.toml`. Optional guidelines are added as an owned block in the
project's `AGENTS.md`. Paths use the supplied checkout/worktree, not the common
Git directory. Each worktree has independent ownership.

Trust the repository through Codex before relying on its project config. An
empty project config does **not** suppress global servers. A genuinely clean
non-AWS project requires the global installation to be migrated too.

Example selection:

```toml
skills = ["domain-modeling"]
guidelines = []

[mcp.aws-documentation]
active = true

[mcp.powertools]
active = true
```

Add `powertools` to the skill catalog separately if you maintain a Powertools
skill: the MCP definition and skill are distinct capabilities. The shipped
example does not invent a missing skill.

## Profiles / on-demand

The installed Codex 0.161.0 and current documentation use profile **files**:
`$CODEX_HOME/<name>.config.toml`, selected with `--profile <name>`.

```bash
go run ./cmd/oxstack install --profile aws-tools
codex --profile aws-tools
```

The shipped profile selects AWS pricing, diagrams and the Bedrock cost workflow.
Installing it does not change base `config.toml` or global skill discovery.
Profile skills are explicit `SKILL.md` references in `developer_instructions`,
not globally installed skills and not additions to `/skills`. Read them only
when the selected workflow is requested. Project scopes use real skill links.

Create another file under `selections/profiles/`, or pass `--selection FILE`.
Use the browser example only after importing its external skill; installing a
workflow does not install its browser executable or dependencies.

## Add capabilities to the catalog

For a local skill, create `skills/<name>/SKILL.md` with valid skill frontmatter.
Nothing activates until `<name>` appears in a selection's `skills` list.
The two `gstack-ab` variants are also catalog capabilities: `qa-byOx` and
`design-review-byOx`. `sync`/`update` maintain those source files; they no longer
copy changes into global skill directories. Provision upstream gstack separately under `catalog/gstack`, or set
`OXSTACK_GSTACK_SOURCE` to a maintained checkout outside runtime skill discovery.

For an external skill, add catalog metadata to `oxstack.toml`:

```toml
[skills.external.example]
repo = "owner/repository"
skill = "example"
revision = "FULL_REVIEWED_GIT_COMMIT"
```

Import from a reviewed, clean local Git skill directory:

```bash
go run ./cmd/oxstack catalog import example --source /checkout/path/to/example
```

This copies only into `catalog/external/example`, records the resolved Git
commit and content digest, and activates nothing. If the catalog revision is
empty, the resolved commit still pins the imported snapshot; prefer an explicit
reviewed pin for reproducibility. Import refuses dirty content, symlinks, wrong
pins and existing catalog destinations. Review and replace snapshots explicitly
when updating; source repository metadata is operator-declared.

Existing installed copies without verified upstream revisions can be preserved explicitly:

```bash
go run ./cmd/oxstack catalog import example --source /installed/example --snapshot
```

Snapshot mode records `kind = local-snapshot`, the historical source path and a
content hash; it makes no Git revision claim. It refuses a declared Git pin and
activates nothing. Unknown upstream sources must be declared as `repo = "UNKNOWN"`.
Local catalog contents are ignored by Git: back them up separately. To relocate
global skills, first import and verify every copy, preserve the originals outside
skill discovery, and retarget dependent links. Record old paths and link targets
in a private rollback inventory. Do not delete originals or automatically activate
the imported catalog.

For an MCP, create `mcp/<name>.toml`:

```toml
source = "upstream/package"
version = "1.2.3"
command = "uvx"
args = ["upstream-package@1.2.3"]
inherit_env = ["AWS_PROFILE", "AWS_REGION"]
```

Or use `url = "https://example.com/mcp"` instead of command/args for an HTTP MCP.
Select it under `[mcp.<name>]`. Keep AWS server names identifiable (for example
`aws-cloudformation`) so capability checks can recognize them. Optional selection `env` values are scope-specific;
never commit secrets. Runtime credential inheritance uses Codex `env_vars`,
Claude variable references and OpenCode environment references. No particular
AWS account/profile is baked into the portable catalog.

`active` is canonical. Legacy `disabled = true` input maps to inactive; specifying
both is an error. Codex/OpenCode emit `enabled = false`; inactive Claude servers
are **not registered**, since an invented per-server disabled field would not
reliably disable them. The shipped legacy packages retain unversioned/latest
commands; review and pin these before production use. Import/install never runs
them, and the tests do not validate upstream server health.

## Native runtime adapters

| Scope | Codex | Claude Code | OpenCode |
|---|---|---|---|
| Global MCP | `$CODEX_HOME/config.toml` | `~/.claude.json` | `$XDG_CONFIG_HOME/opencode/opencode.json` |
| Project MCP | `.codex/config.toml` | `.mcp.json` | `opencode.json` |
| Global skill | `~/.agents/skills` | `~/.claude/skills` | `$XDG_CONFIG_HOME/opencode/skills` |
| Project skill | `.agents/skills` | `.claude/skills` | `.opencode/skills` |
| Global instructions | `$CODEX_HOME/AGENTS.md` | `~/.claude/CLAUDE.md` | `$XDG_CONFIG_HOME/opencode/AGENTS.md` |

Other runtimes have no fabricated Codex-style profile syntax. Explicit profile
artifacts live under `${XDG_STATE_HOME:-~/.local/state}/oxstack/profiles/<name>`:

```bash
go run ./cmd/oxstack install --profile aws-tools --runtime claude
claude --mcp-config ~/.local/state/oxstack/profiles/aws-tools/claude.json \
  --append-system-prompt-file ~/.local/state/oxstack/profiles/aws-tools/instructions-claude.md

go run ./cmd/oxstack install --profile aws-tools --runtime opencode
OPENCODE_CONFIG=~/.local/state/oxstack/profiles/aws-tools/opencode.json opencode
```

Claude may require native project-server approval. OpenCode's explicit config
merges with its normal configuration. These invocations do not erase existing
user/account tools. Adjust paths if `XDG_STATE_HOME` is set. Existing JSONC
configuration is refused instead of being destructively converted to JSON.

## Ownership and uninstall

Ownership records include capability, source, version/provenance, scope,
destination, runtime and fingerprint. Global/profile manifests live under
`$XDG_STATE_HOME/oxstack`; project manifests under `.oxstack/ownership-<runtime>.json`.
Keep project ownership local and ignore `.oxstack/` in the consumer repository;
absolute skill links also require the oxStack checkout at their target location.

```bash
go run ./cmd/oxstack uninstall --project /path/to/repo --dry-run
go run ./cmd/oxstack uninstall --project /path/to/repo
```

Uninstall uses the manifest, never all catalog names or marker-to-EOF truncation.
It removes only unchanged owned links, entries and instruction blocks. It leaves
unrelated settings, app MCPs, user instructions, directories and runtime state.
Missing/modified owned entries, unsafe symlink parents, ambiguous shared inline
TOML, corrupt manifests and unowned collisions fail safely. Existing user entries
are not adopted even when their content matches. Resolve ownership deliberately.

TOML edits preserve unrelated source bytes and future unknown settings; JSON
preserves unrelated values but may reformat the file. Installs preflight the
whole plan, recheck before writes, write files atomically and roll back normal
write errors. Do not run simultaneous installers or edit destinations during
apply. Power loss/process termination can leave a partial multi-file operation;
stop and review the manifest and files rather than forcing deletion.

## Migration and verification

Migration is **not** automatic and is currently scoped to the old Codex install.
See [the migration runbook](docs/guides/scoped-activation-migration.md).
The old uninstall is gone. `pull-config` is retired: edit selections rather than
copying one runtime's global activation back into the catalog.

After migration:

```bash
codex -C /path/to/non-aws-project mcp list --json |
jq -e '[.[] | select(.name | test("aws|bedrock|powertools|serverless"; "i"))] | length == 0'

python3 scripts/verify-clean-project.py /path/to/non-aws-project
```

On the tested Codex 0.161.0, `mcp list` discovers project configuration from its
actual working directory even when `-C` names another directory. Also run the
MCP check after `cd /path/to/non-aws-project`. The verification script sets both
working directory and `-C` to prevent a false pass from missing project config.

Inspect `/skills` and `/mcp` in fresh CLI and desktop sessions. The verification
script checks known AWS skill paths in the actual prompt; app/account plugins,
unknown naming conventions and other independent installers need separate review.

For a safe native CLI smoke check before migrating your machine:

```bash
python3 scripts/smoke-codex.py
```

It uses temporary Codex/state/project paths and only lists configured MCPs;
it does not launch AWS servers or modify live configuration. Skill deployment
is covered by the temporary-home Go tests; the unchanged live user skill catalog
is intentionally not cleaned by this smoke check.

References: [Codex configuration](https://learn.chatgpt.com/docs/config-file/config-basic),
[Claude MCP scopes](https://code.claude.com/docs/en/mcp),
[OpenCode MCP](https://opencode.ai/docs/mcp-servers/).

OpenCode skills use its native `.opencode/skills/` project directory or
`$XDG_CONFIG_HOME/opencode/skills/` globally. They do not install into Codex's
shared `.agents/skills/` directory. Claude uses `.claude/skills/`.

Personal app-plugin selections and user MCP definitions can also live in native
Codex profile files without adding them to oxStack's portable catalog. Installed
plugins can remain inactive in the base config and be enabled in an explicit
profile. Keep app-managed definitions and local paths out of the shared catalog.

Profile instruction filenames include the runtime (`instructions-claude.md`,
`instructions-opencode.md`) so different runtimes can share a profile name safely.
