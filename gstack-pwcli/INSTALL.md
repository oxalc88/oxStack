# gstack + playwright-mcp

Installs gstack skills with browse/qa adapted to use `playwright-mcp` instead of gstack's compiled Bun binary. No Bun dependency. Works on Ubuntu and macOS.

## Directory structure

```
gstack-pwcli/
├── install-gstack.sh                    # Run this — handles everything
├── skills/
│   ├── browse/
│   │   └── SKILL.md                     # Adapted: playwright-mcp instead of $B binary
│   └── qa/
│       ├── SKILL.md                     # Adapted: playwright-mcp, same methodology as gstack
│       ├── references/
│       │   └── issue-taxonomy.md        # Adapted: playwright-mcp commands
│       └── templates/
│           └── qa-report-template.md    # Unchanged from gstack
```

## Option A: Add to oxStack (recommended)

Copy the `gstack-pwcli/` folder into your oxStack repo, then update your `install.sh` to call it:

```bash
# Add to oxStack's install.sh
"$HOME/projects/oxDevelop/oxStack/gstack-pwcli/install-gstack.sh"
```

## Option B: Run standalone

```bash
cd gstack-pwcli
chmod +x install-gstack.sh
./install-gstack.sh
```

## Claude Code prompt

Paste this into Claude Code after running the install script:

```
Add a "gstack + playwright-mcp" section to CLAUDE.md that says:

- Use the /browse skill with playwright-mcp for all web browsing. NEVER use mcp__claude-in-chrome__* tools.
- Available skills: /plan-ceo-review, /plan-eng-review, /review, /ship, /browse, /qa, /retro
- /browse and /qa use playwright-mcp (not gstack's Bun binary). No @ prefix on element refs — use e5 not @e5.
- /qa supports modes: diff-aware (auto on feature branches), full, quick (--quick), exhaustive (--exhaustive), regression (--regression baseline.json), report-only (--report-only).
- For authenticated page testing, log in via playwright-mcp named sessions (--session=name) instead of /setup-browser-cookies.
- If playwright-mcp is not found, run: npm install -g @playwright/mcp@latest

Then ask if I also want to add gstack to the current project so teammates get it.
```

## What gets installed

### From gstack (unchanged)

| Skill | Description |
|---|---|
| `/plan-ceo-review` | Founder/product thinking — rethink the problem, find the 10-star product |
| `/plan-eng-review` | Engineering architecture — data flow, diagrams, failure modes, test matrix |
| `/review` | Paranoid staff engineer — race conditions, N+1 queries, trust boundaries |
| `/ship` | Release engineer — sync main, run tests, push, open PR |
| `/retro` | Engineering retrospective — commit analysis, team breakdown, trends |

### Adapted to playwright-mcp

| Skill | Description |
|---|---|
| `/browse` | Browser automation via `playwright-mcp` — navigate, click, fill, screenshot, console |
| `/qa` | Systematic QA testing via `playwright-mcp` — same 11-phase methodology as gstack (test → triage → fix → verify → report), same health score rubric, same issue taxonomy |

### Skipped from gstack

| Skill | Reason |
|---|---|
| `/setup-browser-cookies` | macOS-only Keychain decryption. Replaced by playwright-mcp session persistence. |
| `/gstack-upgrade` | We manage our own updates via oxStack. |
| `/qa-only` | Merged into `/qa --report-only` flag. |

## What changed in the QA skill (vs gstack original)

The adapted `qa/SKILL.md` (580 lines) is a direct transformation of gstack's original (618 lines). Same 11 phases, same health score rubric, same fix loop with WTF-likelihood, same 50-fix hard cap.

**Removed** (38 lines): gstack preamble (update checker, session tracking), contributor mode (field reports to gstack team), AskUserQuestion format template, auto-generated comment.

**Replaced**: every `$B <command>` → `playwright-mcp <command>`. A command mapping table is included at the top of the skill for reference.

**Added**: playwright-mcp setup check, `--session=qa` on all browser commands, Astro/Starlight framework guidance, `--report-only` mode flag.

**Unchanged**: issue-taxonomy.md reference (updated command examples), qa-report-template.md (no gstack references).
