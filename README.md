# oxStack

Portable AI tools configuration: custom skills, agents, and development guidelines for Claude Code, Codex CLI, and OpenCode.

## Quick Start

```bash
git clone <repo-url> ~/projects/oxDevelop/oxStack
cd ~/projects/oxDevelop/oxStack
./install.sh
```

The script creates symlinks from this repo to the appropriate config directories. Edits made here are immediately reflected everywhere.

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
| **tdd-specialist** | sonnet | Test-writing specialist with red-green-refactor workflow |
| **test-reviewer** | sonnet | Test quality reviewer (GOOD/NOISY/FRAGILE/WRONG classification) |

### Development Guidelines

Generated to `~/.codex/AGENTS.md` and `~/.opencode/AGENTS.md`:

Shared development philosophy, process, git safety defaults, and quality gates. The OpenCode version includes an additional multi-model delivery flow section.

## External Skills

These third-party skills are installed separately and are **not managed by this repo**. They are documented here for reproducibility on new machines.

### Install All

```bash
npx skills add vercel-labs/skills --skill find-skills
npx skills add addyosmani/web-quality-skills --skill accessibility
npx skills add addyosmani/web-quality-skills --skill best-practices
npx skills add addyosmani/web-quality-skills --skill core-web-vitals
npx skills add addyosmani/web-quality-skills --skill performance
npx skills add addyosmani/web-quality-skills --skill seo
npx skills add addyosmani/web-quality-skills --skill web-quality-audit
npx skills add mattpocock/skills/tdd
```

### Verify

After installing, external skills should appear in:
- `~/.agents/skills/` (canonical location)
- `~/.claude/skills/` (symlinked automatically by the skills CLI)

## Uninstall

```bash
./install.sh --uninstall
```

Removes all symlinks and generated files created by the install script. External skills and backups (`.bak` files) are not affected.

## Updating

Edit files directly in this repo. Changes to skills and agents are reflected immediately through symlinks.

For generated files (Codex/OpenCode AGENTS.md), re-run:

```bash
./install.sh
```
