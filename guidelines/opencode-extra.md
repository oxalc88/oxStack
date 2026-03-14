## Multi-model delivery flow 

1) Plan pass: Claude Opus (read/analysis only), produce plan artifact.
2) Execute pass: OpenCode with Kimi, implement only approved plan scope.
3) Review pass: Codex in read-only mode first, optional autofix only after human approval.
4) Store all run artifacts under .agent-artifacts/<timestamp>/.
