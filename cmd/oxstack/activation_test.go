package main

import (
	"github.com/pelletier/go-toml/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) Environment {
	t.Helper()
	base := t.TempDir()
	base, _ = filepath.EvalSymlinks(base)
	e := Environment{Root: filepath.Join(base, "catalog"), Home: filepath.Join(base, "home"), CodexHome: filepath.Join(base, "custom-codex"), StateHome: filepath.Join(base, "state")}
	for _, name := range []string{"domain-modeling", "bedrock-costs", "astro-coding"} {
		writeTest(t, filepath.Join(e.Root, "skills", name, "SKILL.md"), "---\nname: "+name+"\ndescription: test\n---\nInstructions\n")
	}
	writeTest(t, filepath.Join(e.Root, "oxstack.toml"), "[gstack]\nforked_skills = []\n")
	writeTest(t, filepath.Join(e.Root, "mcp", "aws-documentation.toml"), "command = \"uvx\"\nargs = [\"aws-docs@1.0.0\"]\ndisabled = true\n")
	writeTest(t, filepath.Join(e.Root, "guidelines", "universal.md"), "Be clear. Verify changes.\n")
	return e
}
func writeTest(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0600); err != nil {
		t.Fatal(err)
	}
}
func applySelection(t *testing.T, e Environment, target Target, s Selection) {
	t.Helper()
	p, err := planInstall(e, target, s)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Apply(); err != nil {
		t.Fatal(err)
	}
}
func readTOML(t *testing.T, p string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var d map[string]any
	if err = toml.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	return d
}
func TestCatalogDoesNotActivateAndGlobalIsExplicit(t *testing.T) {
	e := fixture(t)
	target := Target{Scope: "global", Runtime: "codex"}
	applySelection(t, e, target, Selection{Guidelines: []string{"universal"}})
	if _, err := os.Stat(filepath.Join(e.Home, ".agents", "skills", "bedrock-costs")); !os.IsNotExist(err) {
		t.Fatalf("catalog skill activated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(e.CodexHome, "config.toml")); !os.IsNotExist(err) {
		t.Fatalf("catalog MCP activated: %v", err)
	}
	applySelection(t, e, target, Selection{Skills: []string{"domain-modeling"}, Guidelines: []string{"universal"}})
	p := filepath.Join(e.Home, ".agents", "skills", "domain-modeling")
	if dst, err := os.Readlink(p); err != nil || dst != filepath.Join(e.Root, "skills", "domain-modeling") {
		t.Fatalf("selected link: %s %v", dst, err)
	}
}
func TestAWSProjectAndNonAWSIsolation(t *testing.T) {
	e := fixture(t)
	aws := filepath.Join(e.Home, "aws")
	plain := filepath.Join(e.Home, "plain")
	yes := true
	applySelection(t, e, Target{Scope: "project", Runtime: "codex", Project: aws}, Selection{Skills: []string{"domain-modeling", "bedrock-costs"}, MCP: map[string]Activation{"aws-documentation": {Active: &yes}}})
	d := readTOML(t, filepath.Join(aws, ".codex", "config.toml"))
	if d["mcp_servers"].(map[string]any)["aws-documentation"].(map[string]any)["enabled"] != true {
		t.Fatal("AWS selection missing")
	}
	applySelection(t, e, Target{Scope: "project", Runtime: "codex", Project: plain}, Selection{Skills: []string{"domain-modeling"}})
	for _, p := range []string{filepath.Join(plain, ".codex", "config.toml"), filepath.Join(plain, ".agents", "skills", "bedrock-costs"), filepath.Join(e.CodexHome, "config.toml")} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("leaked AWS capability: %s %v", p, err)
		}
	}
	if _, err := os.Readlink(filepath.Join(aws, ".agents", "skills", "domain-modeling")); err != nil {
		t.Fatal(err)
	}
}
func TestInactiveRuntimeTranslation(t *testing.T) {
	for _, runtime := range []string{"codex", "claude", "opencode"} {
		t.Run(runtime, func(t *testing.T) {
			e := fixture(t)
			target := Target{Scope: "project", Runtime: runtime, Project: filepath.Join(e.Home, "project")}
			applySelection(t, e, target, Selection{MCP: map[string]Activation{"aws-documentation": {}}})
			if runtime == "claude" {
				if _, err := os.Stat(filepath.Join(target.Project, ".mcp.json")); !os.IsNotExist(err) {
					t.Fatal("inactive Claude MCP was registered")
				}
				return
			}
			p, path := mcpDestination(e, target, "aws-documentation")
			data, err := readDocument(p, runtime == "codex")
			if err != nil {
				t.Fatal(err)
			}
			v, _ := lookup(data, path)
			if v.(map[string]any)["enabled"] != false {
				t.Fatal("inactive MCP enabled")
			}
		})
	}
}
func TestCodexPreservesUnrelatedSettingsAndIsIdempotent(t *testing.T) {
	e := fixture(t)
	target := Target{Scope: "global", Runtime: "codex"}
	p := filepath.Join(e.CodexHome, "config.toml")
	original := `# personal defaults
model = "personal"
unknown_future = { nested = [1, 2] }
[mcp_servers.app_owned]
command = "app"
[plugins."example@app"]
enabled = true
[projects."/some/repo"]
trust_level = "trusted"
`
	writeTest(t, p, original)
	s := Selection{MCP: map[string]Activation{"aws-documentation": {}}}
	applySelection(t, e, target, s)
	before, _ := os.ReadFile(p)
	applySelection(t, e, target, s)
	after, _ := os.ReadFile(p)
	if string(before) != string(after) {
		t.Fatal("install was not idempotent")
	}
	plan, err := planUninstall(e, target)
	if err != nil {
		t.Fatal(err)
	}
	if err = plan.Apply(); err != nil {
		t.Fatal(err)
	}
	after, _ = os.ReadFile(p)
	d := readTOML(t, p)
	if d["model"] != "personal" || d["unknown_future"] == nil || d["plugins"] == nil || d["projects"] == nil {
		t.Fatal("unrelated settings removed")
	}
	if d["mcp_servers"].(map[string]any)["app_owned"] == nil {
		t.Fatal("app server removed")
	}
	if len(after) < len(original) {
		t.Fatal("unrelated source lost")
	}
}
func TestUnownedAndExternallyModifiedResourcesRefused(t *testing.T) {
	e := fixture(t)
	target := Target{Scope: "global", Runtime: "codex"}
	s := Selection{Skills: []string{"domain-modeling"}}
	link := filepath.Join(e.Home, ".agents", "skills", "domain-modeling")
	writeTest(t, filepath.Join(link, "SKILL.md"), "user-owned")
	if _, err := planInstall(e, target, s); err == nil {
		t.Fatal("overwrote unowned skill")
	}
	target = Target{Scope: "project", Runtime: "codex", Project: filepath.Join(e.Home, "project")}
	applySelection(t, e, target, Selection{MCP: map[string]Activation{"aws-documentation": {}}})
	p := filepath.Join(target.Project, ".codex", "config.toml")
	writeTest(t, p, "[mcp_servers.aws-documentation]\ncommand = \"external\"\nenabled = false\n")
	if _, err := planUninstall(e, target); err == nil {
		t.Fatal("removed externally changed server")
	}
}
func TestCodexHome(t *testing.T) {
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "alternate"))
	e, err := newEnvironment(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := canonicalLocation(os.Getenv("CODEX_HOME"))
	if e.CodexHome != expected {
		t.Fatalf("CODEX_HOME ignored: %s", e.CodexHome)
	}
}

func TestOpenCodeSkillsDoNotActivateCodex(t *testing.T) {
	e := fixture(t)
	for _, target := range []Target{{Scope: "global", Runtime: "opencode"}, {Scope: "project", Runtime: "opencode", Project: filepath.Join(e.Home, "project")}} {
		applySelection(t, e, target, Selection{Skills: []string{"domain-modeling"}})
		if _, err := os.Lstat(filepath.Join(e.Home, ".agents", "skills", "domain-modeling")); !os.IsNotExist(err) {
			t.Fatal("OpenCode activated a global Codex skill")
		}
		if target.Scope == "project" {
			if _, err := os.Lstat(filepath.Join(target.Project, ".agents", "skills", "domain-modeling")); !os.IsNotExist(err) {
				t.Fatal("OpenCode activated a project Codex skill")
			}
		}
		dest := skillDestination(e, target, "domain-modeling")
		if strings.Contains(dest, string(filepath.Separator)+".agents"+string(filepath.Separator)) {
			t.Fatal("OpenCode leaked a skill into Codex discovery")
		}
		if target.Scope == "project" && dest != filepath.Join(target.Project, ".opencode", "skills", "domain-modeling") {
			t.Fatal(dest)
		}
		if target.Scope == "global" && dest != filepath.Join(e.Home, ".config", "opencode", "skills", "domain-modeling") {
			t.Fatal(dest)
		}
	}
}

func TestGstackMaintenanceSourceIsOutsideDiscovery(t *testing.T) {
	e := fixture(t)
	t.Setenv("OXSTACK_ROOT", e.Root)
	t.Setenv("OXSTACK_GSTACK_SOURCE", "")
	if got := gstackDir(); got != filepath.Join(e.Root, "catalog", "gstack") {
		t.Fatal(got)
	}
	t.Setenv("OXSTACK_GSTACK_SOURCE", filepath.Join(e.Home, "maintained-gstack"))
	if got := gstackDir(); got != filepath.Join(e.Home, "maintained-gstack") {
		t.Fatal(got)
	}
}

func TestRuntimeProfilesCanCoexistAndUninstallIndependently(t *testing.T) {
	e := fixture(t)
	claude := Target{Scope: "profile", Runtime: "claude", Profile: "tools"}
	opencode := Target{Scope: "profile", Runtime: "opencode", Profile: "tools"}
	selection := Selection{Skills: []string{"bedrock-costs"}}
	applySelection(t, e, claude, selection)
	applySelection(t, e, opencode, selection)
	if instructionsDestination(e, claude) == instructionsDestination(e, opencode) {
		t.Fatal("runtime profiles share ownership")
	}
	p, err := planUninstall(e, claude)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Apply(); err != nil {
		t.Fatal(err)
	}
	if !fileExists(instructionsDestination(e, opencode)) {
		t.Fatal("Claude uninstall removed OpenCode profile")
	}
}
