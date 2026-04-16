# oxStack

Portable AI tools configuration: custom skills, agents, and development guidelines for Claude Code, Codex CLI, and OpenCode.

## Prerequisites

- [mise](https://mise.jdx.dev/) — manages the Go toolchain automatically
- `npm` — required for agent-browser and third-party skills

## Quick Start

```bash
git clone <repo-url> ~/projects/oxDevelop/oxStack
cd ~/projects/oxDevelop/oxStack
cp .env.example .env   # Edit with your values (AWS_PROFILE, etc.)

# Install Go via mise (version pinned in .mise.toml)
mise install

# Build and install the CLI
go install ./cmd/oxstack

# Run the full installer
oxstack install
```

## CLI Usage

The `oxstack` CLI is a single Go binary that works on macOS, Linux, and Windows.

### `oxstack install`

Runs the full installer — symlinks skills and agents, generates guideline files, merges MCP server configs, clones gstack, installs agent-browser, installs third-party skills, and deploys the -byOx skill variants.

```bash
oxstack install
```

### `oxstack sync`

Checks gstack for upstream changes to forked skills (qa, design-review). Pulls latest gstack, shows commits since your last sync, and filters out boilerplate — only methodology-relevant changes are shown. Prompts to mark the current gstack commit as synced.

```bash
oxstack sync
```

### `oxstack pull-config`

Syncs MCP `disabled` flags from your live `~/.claude/settings.json` back into `oxStack/mcp/claude.json` so the state is preserved for new machines. Shows a per-server diff and prompts before writing.

```bash
oxstack pull-config
```

Refuses to run if `mcp/claude.json` has uncommitted changes. Only touches the `disabled` field — never writes env-substituted values or permissions back to the repo.

### `oxstack update`

Regenerates -byOx skills from gstack's latest SKILL.md. For each forked skill it:

1. Extracts the methodology from gstack (strips frontmatter and boilerplate sections)
2. Transforms `$B` browse commands to `agent-browser` equivalents
3. Preserves your -byOx header (frontmatter, intro, command mapping table)
4. Shows a colored, section-aware diff summary and asks for confirmation before writing

```bash
oxstack update          # colored summary + condensed diff
oxstack update --full   # full unified diff (raw, verbose)
```

### `oxstack help`

```bash
oxstack help
```

## Building from Source

```bash
# Build in current directory
go build -o oxstack ./cmd/oxstack

# Install to $GOPATH/bin (add to PATH)
go install ./cmd/oxstack

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o oxstack-linux ./cmd/oxstack
GOOS=windows GOARCH=amd64 go build -o oxstack.exe ./cmd/oxstack
```

If running outside the repo, set `OXSTACK_ROOT` to the repo path so the CLI can find its config files:

```bash
export OXSTACK_ROOT=~/projects/oxDevelop/oxStack
```

## What Gets Installed

### Custom Skills

Deployed to both `~/.claude/skills/` and `~/.agents/skills/`:

| Skill | Description |
|-------|-------------|
| **astro-coding** | Astro/Starlight implementation patterns, critical rules, and error catalog |
| **data-modeling** | Translate domain models into persistence designs (ER diagrams, data dictionaries, index strategies) |
| **domain-modeling** | DDD-based domain analysis (bounded contexts, entities, aggregates, invariants, state machines) |

### Custom Agents

Deployed to `~/.claude/agents/`:

| Agent | Model | Description |
|-------|-------|-------------|
| **git-committer** | haiku | Commits changes with Conventional Commits format |
| **tdd-specialist** | sonnet | Writes new behavior-focused tests via red-green-refactor. Produces `test-findings.md`. |
| **test-fixer** | sonnet | Rewrites flagged tests from `test-review-findings.md`. Never writes new tests. |
| **test-reviewer** | sonnet | Evaluates existing test quality (GOOD/NOISY/FRAGILE/WRONG). Produces `test-review-findings.md`. |

### gstack + agent-browser Skills

Installed automatically by `oxstack install`. Clones [gstack](https://github.com/garrytan/gstack), installs [agent-browser](https://github.com/vercel-labs/agent-browser), and deploys -byOx skill variants that use `agent-browser` instead of gstack's compiled Bun binary. No Bun or Playwright dependency — works on macOS, Linux, and Windows.

| Skill | Description |
|-------|-------------|
| **qa-byOx** | Systematic QA testing via agent-browser — test, fix, verify loop |
| **design-review-byOx** | Designer's eye audit via agent-browser — find and fix visual issues |

### Third-party Skills

Installed automatically by `oxstack install`. Defined in `oxstack.toml` under `[skills.external]` — add a one-line entry to track a new skill; no code changes or rebuilds needed.

| Skill | Source |
|-------|--------|
| **agent-browser** | Vercel — browser automation |
| **find-skills** | Vercel — discover installable skills |
| **accessibility** | Addy Osmani — WCAG 2.2 audit |
| **best-practices** | Addy Osmani — security & code quality |
| **core-web-vitals** | Addy Osmani — LCP/INP/CLS optimization |
| **performance** | Addy Osmani — page speed optimization |
| **seo** | Addy Osmani — search engine optimization |
| **web-quality-audit** | Addy Osmani — Lighthouse audit |
| **tdd** | Matt Pocock — test-driven development |
| **ast-grep** | ast-grep — AST-based structural code search |

### MCP Servers

Defined in `oxstack.toml` under `[mcp.servers.*]`. Merged into `~/.claude/settings.json` and appended to `~/.codex/config.toml` by `oxstack install`.

| Server | Description |
|--------|-------------|
| **awslabs.aws-diagram-mcp-server** | Generate AWS architecture diagrams |
| **awslabs.aws-documentation-mcp-server** | Query AWS documentation |
| **awslabs.aws-pricing-mcp-server** | Look up AWS service pricing |
| **awslabs.cfn-mcp-server** | CloudFormation template analysis (readonly) |
| **serverless** | Serverless Framework MCP integration |
| **ultracite** | Ultracite remote MCP |
| **powertools** | AWS Lambda Powertools MCP |

Servers needing env vars use values from `.env` (see `.env.example`). Set `disabled = true` in `oxstack.toml` for servers you want off by default on new machines. Use `oxstack pull-config` to capture toggled state back from your live settings.

### Development Guidelines

Generated to `~/.codex/AGENTS.md` and `~/.opencode/AGENTS.md`. Shared development philosophy, process, git safety defaults, and quality gates. The OpenCode version includes an additional multi-model delivery flow section.

## Workflow

```
oxstack install      # First time setup (or after adding new skills/agents)
oxstack sync         # After gstack updates — review methodology changes
oxstack update       # Apply gstack methodology changes to -byOx skills
oxstack pull-config  # Capture MCP disabled toggles back into mcp/claude.json
oxstack uninstall    # Remove all symlinks, generated files, and MCP servers
```

Edit skills and agents directly in this repo. Changes are reflected immediately through symlinks — no reinstall needed. Only re-run `oxstack install` when adding new skills/agents or regenerating guideline files.

## Uninstall

```bash
oxstack uninstall
```

Removes all skill/agent symlinks, generated AGENTS.md files, MCP server entries, and -byOx skills. External skills and backups (`.bak` files) are not affected.
