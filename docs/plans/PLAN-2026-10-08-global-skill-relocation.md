---
summary: Preserve third-party global skills as inactive local catalog snapshots.
read_when: Importing installed skills with unknown revisions or retiring global discovery copies.
status: done
---

# Global skill relocation

1. Add explicit snapshot import with honest unknown-revision provenance and content validation.
2. Preserve all installed skill content in the catalog before changing discovery paths.
3. Archive original installations and record hashes, modes, destinations and dependent link targets.
4. Preserve Claude/Devin behavior by retargeting their existing links.
5. Verify fresh Codex prompt discovery, clean-project MCPs and the Go checks.

Completed: 20 unique catalog snapshots; 31 original installations archived;
10 existing dependent links retargeted. Local private inventories retain installer
metadata and file hashes/modes. No project activation was added. Built-in skills,
app plugins, MCP configuration and installer lock metadata were left in place.

Validation: Go tests, race detector, vet and diff whitespace check pass. Native
Codex 0.161.0 prompt rendering in a temporary clean directory, with tool MCPs and
plugins disabled only for that invocation, contains no relocated global skill
paths. The unmodified native MCP list contains zero AWS-family definitions.

Local snapshots are ignored by Git, not a portable upstream pin. Keep their
private archives and verify upstream provenance separately before replacing them.
