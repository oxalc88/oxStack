#!/usr/bin/env bash
set -euo pipefail

# gstack + agent-browser installer
# Installs gstack skills, replaces qa with agent-browser version
# Works on Ubuntu and macOS — no Bun or Playwright required

GSTACK_DIR="$HOME/.claude/skills/gstack"
SKILLS_DIR="$HOME/.claude/skills"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[✗]${NC} $1"; }

# --- Step 1: Clone gstack ---
if [ -d "$GSTACK_DIR" ]; then
  warn "gstack already exists at $GSTACK_DIR — pulling latest"
  cd "$GSTACK_DIR" && git pull --ff-only 2>/dev/null || true
else
  info "Cloning gstack..."
  git clone --depth 1 https://github.com/garrytan/gstack.git "$GSTACK_DIR"
fi

# --- Step 2: Install agent-browser ---
if command -v agent-browser &>/dev/null; then
  info "agent-browser already installed: $(which agent-browser)"
else
  info "Installing agent-browser..."
  npm install -g agent-browser
fi

# Install browser + system deps if needed
info "Setting up browser for agent-browser..."
agent-browser install --with-deps

# --- Step 3: Install official agent-browser skill for /browse ---
info "Installing official agent-browser skill..."
npx skills add vercel-labs/agent-browser --skill agent-browser

# --- Step 4: Replace qa skill with agent-browser version ---
info "Replacing qa skill with agent-browser version..."
rm -rf "$GSTACK_DIR/qa/SKILL.md"
cp "$SCRIPT_DIR/skills/qa/SKILL.md" "$GSTACK_DIR/qa/SKILL.md"
cp "$SCRIPT_DIR/skills/qa/references/issue-taxonomy.md" "$GSTACK_DIR/qa/references/issue-taxonomy.md"
# Keep gstack's qa-report-template.md as-is (no gstack-specific content)

# --- Step 5: Create skill symlinks ---
info "Creating skill symlinks..."
mkdir -p "$SKILLS_DIR"

# Skills we keep from gstack (unchanged)
GSTACK_SKILLS="plan-ceo-review plan-eng-review review ship retro"

# Skills we replaced with agent-browser versions
REPLACED_SKILLS="qa"

# Skills we skip entirely
# browse (now handled by official agent-browser skill)
# setup-browser-cookies (macOS-only, replaced by agent-browser sessions)
# gstack-upgrade (we manage our own updates)
# qa-only (merged into qa --report-only)

for skill in $GSTACK_SKILLS $REPLACED_SKILLS; do
  if [ -d "$GSTACK_DIR/$skill" ] && [ -f "$GSTACK_DIR/$skill/SKILL.md" ]; then
    ln -snf "gstack/$skill" "$SKILLS_DIR/$skill"
    info "  Linked: /$(basename "$skill")"
  else
    warn "  Skipped: $skill (not found in gstack)"
  fi
done

# --- Step 6: Remove symlinks for skipped skills ---
for skip in setup-browser-cookies gstack-upgrade qa-only browse; do
  if [ -L "$SKILLS_DIR/$skip" ]; then
    rm -f "$SKILLS_DIR/$skip"
    info "  Removed symlink: $skip"
  fi
done

# --- Done ---
echo ""
info "Installation complete!"
echo ""
echo "  Available skills:"
echo "    /plan-ceo-review   — Founder/product thinking mode"
echo "    /plan-eng-review   — Engineering architecture mode"
echo "    /review            — Paranoid staff engineer code review"
echo "    /ship              — Release engineer (sync, test, push, PR)"
echo "    /browse            — Browser automation via agent-browser (official skill)"
echo "    /qa                — Systematic QA testing via agent-browser"
echo "    /retro             — Weekly engineering retrospective"
echo ""
echo "  Add this to your CLAUDE.md:"
echo ""
cat << 'CLAUDE_SECTION'
## gstack + agent-browser

Use the /browse skill with agent-browser for all web browsing. NEVER use mcp__claude-in-chrome__* tools.

Available skills: /plan-ceo-review, /plan-eng-review, /review, /ship, /browse, /qa, /retro

- /browse uses the official agent-browser skill. /qa uses agent-browser (not gstack's Bun binary). No @ prefix on element refs — use e5 not @e5.
- /qa supports modes: diff-aware (auto on feature branches), full, quick, regression, report-only.
- For authenticated page testing, log in via agent-browser sessions (--session name) instead of /setup-browser-cookies.
- If agent-browser is not found, run: npm install -g agent-browser && agent-browser install --with-deps
CLAUDE_SECTION
