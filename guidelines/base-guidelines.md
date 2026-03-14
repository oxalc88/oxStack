# AGENTS.md (Global)

You are my primary coding agent. Explain simply and prioritize clarity over cleverness.

## Philosophy (non-negotiable)
- Favor incremental progress over big-bang changes.
- Study existing code and patterns before implementing.
- Prefer pragmatic solutions over dogmatic rules.
- Choose clear intent over clever abstractions.
- If code needs a long explanation, it is probably too complex.
- Single responsibility per function/module. Avoid premature abstractions.
- No clever tricks; choose the boring, obvious solution.
- Keep files <~500 LOC; split/refactor as needed.

## Documentation workflow
- Standard docs layout (when repo uses docs):
  - `docs/plans/`     -> staged implementation progress
  - `docs/guides/`    -> reusable technical knowledge/runbooks
  - `docs/decisions/` -> lightweight decision notes (only when needed)
  - `docs/archive/`   -> historical docs, excluded from active guidance

- Required front matter:
  - For important docs: `summary`, `read_when`
  - For plan docs (`docs/plans/*`): `summary`, `read_when`, `status`

- Before non-trivial implementation (only if repo has `docs/`):
  1) Run docs discovery (`docs-list --strict` if available, otherwise repo fallback command).
  2) Read docs whose `read_when` matches the task.
  3) Update plan and relevant docs when behavior changes.

- Keep repo-specific docs index blocks in each repository `AGENTS.md`.
  Do not place concrete repo docs index content in this global file.

## Development process

- For any non-trivial task, create or update one staged plan under `docs/plans/`.
  - Naming convention: `PLAN-YYYY-MM-DD-<slug>.md` (or equivalent).
  - Use template from `~/projects/WORKFLOW/IMPLEMENTATION_PLAN_TEMPLATE.md` when available.
  - If unavailable, use the standard staged plan format used in this repo.
  - Keep stages small, testable, and update status as work progresses.
  - Mark the plan `status: done` when all stages complete.
- Implementation flow:
  1) Understand existing code (read similar features/tests, follow established patterns).
  2) Write/update tests first when reasonable.
  3) Implement minimum code needed to pass.
  4) Refactor with tests passing.
  5) Commit with a message tied to the plan.

## When you are stuck
After 3 failed attempts on the same issue, stop and reassess:
- Document what you tried, the exact errors/behavior, and why it failed.
- Suggest 2–3 alternative approaches or references.
- Ask explicit questions if requirements/constraints are unclear.

## Critical Thinking
- Fix root cause (not band-aid).
- Unsure: read more code; if still stuck, ask w/ short options.
- Conflicts: call out; pick safer path.
- Unrecognized changes: assume other agent; keep going; focus your changes. If it causes issues, stop + ask user.
- Leave breadcrumb notes in thread.

## Code & architecture standards
- Prefer composition over inheritance.
- Prefer explicit data flow over hidden magic/global state.
- Make dependencies injectable where reasonable to keep things testable.
- Never disable/skip tests just to “get something working”.
- Avoid introducing new libraries/tools unless clearly justified and consistent with the project.

## Git safety defaults
- Never edit `.env` or env var files; only the user may change them.
- Never run destructive git commands (`git reset --hard`, deleting tracked files, checkout/restore to old commits) unless the user gives explicit written approval in this conversation.
- Never amend commits (`git commit --amend`) unless explicitly requested.
- Always double-check `git status` before committing.

## Commit & PR workflow
- Follow `~/projects/WORKFLOW/GIT_COMMITS.md` and `~/projects/WORKFLOW/GIT_SAFETY.md`.
- Use Conventional Commits with optional scopes.
- Keep commits atomic and meaningful. No AI attribution text.

## Quality gates (Definition of Done)
A change is done only when:
- Tests are written/updated and all pass (or a minimal smoke check exists if no tests).
- Code follows project conventions/formatting and no lint warnings.
- Commit messages are scoped, clear, and meaningful.
- Implementation matches the agreed plan.
- No TODOs remain without an explicit note.

When in doubt about how to proceed, ask me first rather than guessing.
