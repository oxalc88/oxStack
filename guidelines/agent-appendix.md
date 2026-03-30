
## Agent reference (oxStack)

The following specialized agents are available in Claude Code (`~/.claude/agents/`).
They are not directly usable by Codex or OpenCode but document team conventions:

- **git-committer** — Commits staged or unstaged changes using Conventional Commits.
  Atomic, meaningful commits with no AI attribution. Follows git safety defaults above.
- **tdd-specialist** — Writes new behavior-focused tests using red-green-refactor.
  Never modifies implementation code. Produces `test-findings.md` with pass/fail status.
- **test-fixer** — Rewrites flagged tests from `test-review-findings.md` findings.
  Never writes new tests or modifies implementation code.
- **test-reviewer** — Evaluates existing test quality. Classifies tests as
  GOOD / NOISY / FRAGILE / WRONG. Produces `test-review-findings.md` with actionable
  recommendations. Does not write or rewrite tests.
