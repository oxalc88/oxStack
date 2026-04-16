## Commits
When committing code, always use atomic commits (one logical change per commit). Never make a single large commit combining multiple changes. NEVER run `git commit` directly — always delegate commits to the git-committer agent (spawn a sub-agent). Group changes by: feature, refactor, cleanup/deletion, docs. Each logical change gets its own commit.

## Scope Changes
When asked to change a specific element (e.g., one image, one section), change ONLY that element. Do not apply changes to similar elements nearby. If unclear, ask before making broader changes.

## About Implementations
Always use real project assets (photos, logos, content) instead of placeholders. Never use placeholder images or generic content unless explicitly told to.

## New Languages and tools
Use mise for managing tool versions (Go, Node, etc.). Do not pick or install versions directly — check mise configuration first.

## Plan review before execute
When a plan or TODO file is referenced, ask the user to confirm the exact filename/path before proceeding. Known plan locations: check project root and .claude/ directory for plan files.

## Doc Matching
You run in an environment where `ast-grep` is available; whenever a search requires syntax-aware or structural matching, default to `ast-grep --lang rust -p '<pattern>'` (or set `--lang` appropriately) and avoid falling back to text-only tools like `rg` or `grep` unless I explicitly request a plain-text search. If you need a guide of how or when to use it check it out the /ast-grep skill.

## oxStack is the source of truth
When creating new skills, agents, MCP configs, or plugins, always place them in the oxStack repo (`~/projects/oxDeveloop/oxStack/`):
- Skills → `skills/<skill-name>/SKILL.md`
- Agents → `agents/<agent-name>.md`
- MCP servers → `mcp/claude.json`
Then run `oxstack install` to symlink/distribute them to `~/.claude/` and `~/.agents/`.
