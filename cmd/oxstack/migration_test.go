package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationPreviewAndApplyPreserveAppConfig(t *testing.T) {
	e := fixture(t)
	writeTest(t, filepath.Join(e.Root, "migration", "legacy-catalog.toml"), "[mcp.servers.\"awslabs.aws-documentation-mcp-server\"]\ncommand = \"uvx\"\nargs = [\"legacy-docs\"]\ndisabled = true\n")
	writeTest(t, filepath.Join(e.Root, "guidelines", "base-guidelines.md"), "legacy guidelines\n")
	writeTest(t, filepath.Join(e.Root, "guidelines", "agent-appendix.md"), "Claude-only appendix\n")
	original := `model = "personal"
# --- oxStack MCP servers ---
[mcp_servers.awslabs-aws-documentation-mcp-server]
command = "uvx"
args = ["legacy-docs"]
[mcp_servers.node_repl]
command = "app-owned"
[projects."/work"]
trust_level = "trusted"
`
	config := filepath.Join(e.CodexHome, "config.toml")
	writeTest(t, config, original)
	writeTest(t, filepath.Join(e.CodexHome, "AGENTS.md"), generatedHeader+"\n\nlegacy guidelines\n\nClaude-only appendix\n")
	link := filepath.Join(e.Home, ".agents", "skills", "bedrock-costs")
	if err := os.MkdirAll(filepath.Dir(link), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(e.Root, "skills", "bedrock-costs"), link); err != nil {
		t.Fatal(err)
	}
	copied := filepath.Join(e.Home, ".agents", "skills", "cloudflare", "SKILL.md")
	writeTest(t, copied, "unowned copy")
	p, err := planMigration(e)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(config)
	if string(before) != original {
		t.Fatal("preview wrote config")
	}
	if !strings.Contains(strings.Join(p.Actions, "\n"), "REVIEW (not removed)") {
		t.Fatal("unknown skills not reported")
	}
	if err = p.Apply(); err != nil {
		t.Fatal(err)
	}
	d := readTOML(t, config)
	if d["model"] != "personal" || d["projects"] == nil {
		t.Fatal("personal settings lost")
	}
	if _, ok := lookup(d, []string{"mcp_servers", "node_repl"}); !ok {
		t.Fatal("app config after legacy marker lost")
	}
	if _, ok := lookup(d, []string{"mcp_servers", "awslabs-aws-documentation-mcp-server"}); ok {
		t.Fatal("legacy entry retained")
	}
	if _, err := os.ReadFile(copied); err != nil {
		t.Fatal("unowned copied skill deleted")
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatal("proven legacy link retained")
	}
	files, err := filepath.Glob(filepath.Join(e.StateHome, "migration-backups", "*", "*.json"))
	if err != nil || len(files) != 3 {
		t.Fatalf("missing backups: %v %v", files, err)
	}
	applySelection(t, e, Target{Scope: "global", Runtime: "codex"}, Selection{Guidelines: []string{"universal"}})
}
func TestMigrationDoesNotClaimChangedOrUnmarkedLegacyEntries(t *testing.T) {
	for _, marked := range []bool{false, true} {
		t.Run(map[bool]string{true: "modified", false: "unmarked"}[marked], func(t *testing.T) {
			e := fixture(t)
			writeTest(t, filepath.Join(e.Root, "migration", "legacy-catalog.toml"), "[mcp.servers.\"aws.documentation\"]\ncommand = \"uvx\"\nargs = []\n")
			content := "[mcp_servers.aws-documentation]\ncommand = \"uvx\"\nargs = []\n"
			if marked {
				content = "# --- oxStack MCP servers ---\n" + strings.ReplaceAll(content, "uvx", "externally-modified")
			}
			writeTest(t, filepath.Join(e.CodexHome, "config.toml"), content)
			p, err := planMigration(e)
			if err != nil {
				t.Fatal(err)
			}
			if len(p.Changes) != 0 {
				t.Fatal("uncertain identity was claimed")
			}
		})
	}
	if confirm(strings.NewReader("yes\n"), new(strings.Builder)) {
		t.Fatal("weak migration confirmation accepted")
	}
	if !confirm(strings.NewReader("MIGRATE\n"), new(strings.Builder)) {
		t.Fatal("explicit confirmation rejected")
	}
}
