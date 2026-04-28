---
name: build-test-commit
description: Implement a list of features one at a time through a Build → Test → Simplify → Commit pipeline. Use when the user provides multiple features/fixes to ship and wants atomic commits per item with no mid-pipeline check-ins. Delegates commits to the git-committer agent.
---

# Build → Test → Simplify → Commit

Implement features one at a time through a disciplined pipeline. Never bundle multiple features in one commit. Do not ask the user between features — keep going until all are done.

## Input

The user provides a list of features, fixes, or changes to implement.

## Pipeline (per feature)

### 1. Build
- Implement the feature or fix
- Keep changes focused — only touch what's needed for this one item

### 2. Test
- Run the project's test suite (`go test ./...`, `npm test`, or whatever applies)
- If tests fail, fix the code and re-run until all tests pass
- Also run the relevant linter (`go vet ./...`, `npx tsc --noEmit`, etc.)

### 3. Simplify
- Review the changed code for:
  - Duplicated logic that should be extracted
  - Overcomplicated implementations that can be simplified
  - Efficiency issues or unnecessary abstractions
- Fix any issues found
- Re-run tests to confirm simplification didn't break anything

### 4. Commit
- Delegate to the git-committer agent — NEVER run `git commit` directly
- One atomic commit per feature
- Group by: feature, refactor, cleanup, docs
- Commit message should explain the "why", not just the "what"

## Rules

- Process each feature sequentially through all four steps before starting the next
- Never combine unrelated changes in a single commit
- If a feature requires both implementation and test changes, those go in the same commit
- If simplification produces significant refactoring, that gets its own commit
- Report final status with a summary of all commits made
