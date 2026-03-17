#!/usr/bin/env bash
set -euo pipefail

# gstack + playwright-cli installer
# Installs gstack skills, replaces browse/qa with playwright-cli versions
# Works on Ubuntu and macOS — no Bun required

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

# --- Step 2: Install playwright-cli ---
if command -v playwright-cli &>/dev/null; then
  info "playwright-cli already installed: $(which playwright-cli)"
else
  info "Installing playwright-cli..."
  npm install -g @playwright/mcp@latest
fi

# Install chromium browser if needed
if ! npx playwright-cli open --help &>/dev/null 2>&1; then
  info "Installing Chromium for Playwright..."
  npx playwright install chromium
  # Ubuntu may need system deps
  if [ "$(uname)" = "Linux" ]; then
    npx playwright install-deps chromium 2>/dev/null || warn "Could not install system deps — run 'npx playwright install-deps chromium' manually if browse fails"
  fi
fi

# --- Step 3: Replace browse skill with playwright-cli version ---
info "Replacing browse skill with playwright-cli version..."
rm -rf "$GSTACK_DIR/browse/SKILL.md"
cp "$SCRIPT_DIR/skills/browse/SKILL.md" "$GSTACK_DIR/browse/SKILL.md"

# --- Step 4: Replace qa skill with playwright-cli version ---
info "Replacing qa skill with playwright-cli version..."
rm -rf "$GSTACK_DIR/qa/SKILL.md"
cp "$SCRIPT_DIR/skills/qa/SKILL.md" "$GSTACK_DIR/qa/SKILL.md"
cp "$SCRIPT_DIR/skills/qa/references/issue-taxonomy.md" "$GSTACK_DIR/qa/references/issue-taxonomy.md"
# Keep gstack's qa-report-template.md as-is (no gstack-specific content)

# --- Step 5: Create skill symlinks ---
info "Creating skill symlinks..."
mkdir -p "$SKILLS_DIR"

# Skills we keep from gstack (unchanged)
GSTACK_SKILLS="plan-ceo-review plan-eng-review review ship retro"

# Skills we replaced with playwright-cli versions
REPLACED_SKILLS="browse qa"

# Skills we skip entirely
# setup-browser-cookies (macOS-only, replaced by playwright-cli sessions)
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
for skip in setup-browser-cookies gstack-upgrade qa-only; do
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
echo "    /browse            — Browser automation via playwright-cli"
echo "    /qa                — Systematic QA testing via playwright-cli"
echo "    /retro             — Weekly engineering retrospective"
echo ""
echo "  Add this to your CLAUDE.md:"
echo ""
cat << 'CLAUDE_SECTION'
## gstack + playwright-cli

Use the /browse skill with playwright-cli for all web browsing. NEVER use mcp__claude-in-chrome__* tools.

Available skills: /plan-ceo-review, /plan-eng-review, /review, /ship, /browse, /qa, /retro

- /browse and /qa use playwright-cli (not gstack's Bun binary). No @ prefix on element refs — use e5 not @e5.
- /qa supports modes: diff-aware (auto on feature branches), full, quick, regression, report-only.
- For authenticated page testing, log in via playwright-cli sessions instead of /setup-browser-cookies.
- If playwright-cli is not found, run: npm install -g @playwright/mcp@latest
CLAUDE_SECTION