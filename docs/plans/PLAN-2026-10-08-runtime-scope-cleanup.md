---
summary: Retire remaining global domain activation and isolate runtime-native destinations.
read_when: Reviewing global cleanup, runtime skill deployment or on-demand profile ownership.
status: done
---

# Runtime scope cleanup

- Preserved local global tool definitions and original skill installations in private recovery inventories.
- Moved optional Codex app plugins and Xcode tooling into native, explicitly selected profile files.
- Retired reviewed legacy Claude/OpenCode MCP activation without discarding personal settings.
- Preserved customized MCP versions in private on-demand configurations.
- Archived global skill copies and moved the upstream gstack checkout into a catalog-only maintenance location.
- Kept universal monitoring hooks; retired project-specific global allow rules.
- Replaced mandatory global workflow ceremonies with universal agreements.

Implementation findings and corrections:
- OpenCode skill deployment used Codex's shared `.agents/skills` directory. Native OpenCode destinations now prevent cross-runtime global activation.
- Claude/OpenCode profiles shared an instruction filename. Each runtime now has its own file; unchanged old owned files upgrade safely and modified files are refused.
- gstack maintenance assumed a globally activated Claude checkout. It now uses `catalog/gstack` or explicit `OXSTACK_GSTACK_SOURCE`.

Validation: Go tests, race detector, vet, isolated native Codex smoke, fresh default MCP/prompt checks and isolated Stow guard checks passed. Optional profiles remain unselected by default. Recovery inventories and machine-specific profiles remain local, outside Git.

Account/workspace-managed plugins and app processes can have independent activation state. Start fresh sessions and verify after reopening apps. Original archives remain available for recovery. Upstream gstack workflows may contain assumptions about global paths; activating those workflows needs a separate compatibility review, while catalog maintenance uses the preserved checkout.
