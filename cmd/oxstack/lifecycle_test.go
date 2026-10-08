package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestActiveNativeAdaptersAndInheritedCredentials(t *testing.T) {
	yes, no := true, false
	for _, runtime := range []string{"codex", "claude", "opencode"} {
		t.Run(runtime, func(t *testing.T) {
			e := fixture(t)
			writeTest(t, filepath.Join(e.Root, "mcp", "cloudformation.toml"), "command = \"cfn\"\ninherit_env = [\"AWS_PROFILE\"]\n")
			target := Target{Scope: "project", Runtime: runtime, Project: filepath.Join(e.Home, "project")}
			selection := Selection{MCP: map[string]Activation{"cloudformation": {Active: &yes}}}
			applySelection(t, e, target, selection)
			p, key := mcpDestination(e, target, "cloudformation")
			d, err := readDocument(p, runtime == "codex")
			if err != nil {
				t.Fatal(err)
			}
			v, _ := lookup(d, key)
			entry := v.(map[string]any)
			switch runtime {
			case "codex":
				if entry["enabled"] != true || entry["env_vars"] == nil || entry["env"] != nil {
					t.Fatal(entry)
				}
			case "claude":
				if entry["type"] != "stdio" || entry["disabled"] != nil || entry["env"].(map[string]any)["AWS_PROFILE"] != "${AWS_PROFILE}" {
					t.Fatal(entry)
				}
			case "opencode":
				if entry["enabled"] != true || entry["environment"].(map[string]any)["AWS_PROFILE"] != "{env:AWS_PROFILE}" {
					t.Fatal(entry)
				}
			}
			selection.MCP["cloudformation"] = Activation{Active: &no}
			applySelection(t, e, target, selection)
			if runtime == "claude" {
				d, _ = readDocument(p, false)
				if _, ok := lookup(d, key); ok {
					t.Fatal("inactive Claude entry remains registered")
				}
			}
		})
	}
	if _, err := activeState(&yes, &no, true); err == nil {
		t.Fatal("contradictory activation accepted")
	}
}
func TestProfilesAreExplicitAndSkillsAreReferences(t *testing.T) {
	for _, runtime := range []string{"codex", "claude", "opencode"} {
		t.Run(runtime, func(t *testing.T) {
			e := fixture(t)
			target := Target{Scope: "profile", Profile: "aws-tools", Runtime: runtime}
			yes := true
			applySelection(t, e, target, Selection{Skills: []string{"bedrock-costs"}, MCP: map[string]Activation{"aws-documentation": {Active: &yes}}})
			if _, err := os.Stat(filepath.Join(e.Home, ".agents", "skills", "bedrock-costs")); !os.IsNotExist(err) {
				t.Fatal("profile globally installed a skill")
			}
			if _, err := os.Stat(filepath.Join(e.CodexHome, "config.toml")); !os.IsNotExist(err) {
				t.Fatal("profile modified base Codex config")
			}
			if runtime == "codex" {
				d := readTOML(t, filepath.Join(e.CodexHome, "aws-tools.config.toml"))
				if !strings.Contains(d["developer_instructions"].(string), "bedrock-costs/SKILL.md") {
					t.Fatal(d)
				}
			}
			p, err := planUninstall(e, target)
			if err != nil {
				t.Fatal(err)
			}
			if err = p.Apply(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestProjectInstructionsPreserveUserText(t *testing.T) {
	e := fixture(t)
	target := Target{Scope: "project", Runtime: "codex", Project: filepath.Join(e.Home, "project")}
	path := instructionsDestination(e, target)
	original := "# User rules\nDo the user's work.\n"
	writeTest(t, path, original)
	applySelection(t, e, target, Selection{Guidelines: []string{"universal"}})
	p, err := planUninstall(e, target)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Apply(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if string(b) != original {
		t.Fatalf("user instructions changed: %q", b)
	}
}
func TestTOMLParserPreservesUnrelatedSyntax(t *testing.T) {
	original := `# personal comment
model = "own"
future = { "quoted.key" = "value" }
[mcp_servers."owned.name"]
command = "old"
args = [
 "[plugins.not_a_table]",
]
[mcp_servers."owned.name".env]
TEXT = """line one
[projects.not_a_table]
line three"""
[plugins."other.plugin"] # keep this exact header
future = [ { x = 1 }, { x = 2 } ]
[[future_array]]
name = "one"
`
	out, err := changeTOML([]byte(original), []string{"mcp_servers", "owned.name"}, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	retained := original[strings.Index(original, `[plugins."other.plugin"]`):]
	if !bytes.Contains(out, []byte(retained)) {
		t.Fatalf("unrelated suffix changed: %s", out)
	}
	if _, err = changeTOML([]byte(`mcp_servers = { own = { command = "old" }, user = { command = "keep" } }`), []string{"mcp_servers", "own"}, nil, true); err == nil {
		t.Fatal("shared inline ancestor edited unsafely")
	}
}
func TestUninstallRejectsTamperedManifestAndLinks(t *testing.T) {
	e := fixture(t)
	target := Target{Scope: "project", Runtime: "codex", Project: filepath.Join(e.Home, "project")}
	applySelection(t, e, target, Selection{Skills: []string{"domain-modeling"}})
	m, err := loadManifest(e, target)
	if err != nil {
		t.Fatal(err)
	}
	m.Resources[0].Destination = filepath.Join(e.Home, "unrelated")
	b, _ := json.Marshal(m)
	writeTest(t, target.manifestPath(e), string(b))
	if _, err = planUninstall(e, target); err == nil {
		t.Fatal("tampered destination accepted")
	}
}
func TestPreflightConflictDoesNotPartiallyInstall(t *testing.T) {
	e := fixture(t)
	target := Target{Scope: "project", Runtime: "codex", Project: filepath.Join(e.Home, "project")}
	p, err := planInstall(e, target, Selection{Skills: []string{"domain-modeling"}, MCP: map[string]Activation{"aws-documentation": {}}})
	if err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(target.Project, ".codex", "config.toml")
	writeTest(t, cfg, "model = \"external\"\n")
	if err = p.Apply(); err == nil {
		t.Fatal("changed destination overwritten")
	}
	if _, err = os.Lstat(filepath.Join(target.Project, ".agents", "skills", "domain-modeling")); !os.IsNotExist(err) {
		t.Fatal("partial skill installation")
	}
}
func TestChangedSelectionRemovesOnlyPreviousOwnedEntries(t *testing.T) {
	e := fixture(t)
	target := Target{Scope: "project", Runtime: "codex", Project: filepath.Join(e.Home, "project")}
	applySelection(t, e, target, Selection{Skills: []string{"bedrock-costs"}, MCP: map[string]Activation{"aws-documentation": {}}})
	applySelection(t, e, target, Selection{Skills: []string{"domain-modeling"}})
	if _, err := os.Lstat(filepath.Join(target.Project, ".agents", "skills", "bedrock-costs")); !os.IsNotExist(err) {
		t.Fatal("deselected skill retained")
	}
	d := readTOML(t, filepath.Join(target.Project, ".codex", "config.toml"))
	if _, ok := lookup(d, []string{"mcp_servers", "aws-documentation"}); ok {
		t.Fatal("deselected MCP retained")
	}
}
func gitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if b, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("git %v: %s %v", args, b, e)
	}
}
func TestProjectSelectionWorksInGitWorktree(t *testing.T) {
	e := fixture(t)
	repo := filepath.Join(e.Home, "repo")
	writeTest(t, filepath.Join(repo, "README.md"), "project")
	gitTest(t, repo, "init")
	gitTest(t, repo, "add", "README.md")
	gitTest(t, repo, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "initial")
	worktree := filepath.Join(e.Home, "worktree")
	gitTest(t, repo, "worktree", "add", "-b", "feature", worktree)
	target := Target{Scope: "project", Runtime: "codex", Project: worktree}
	applySelection(t, e, target, Selection{Skills: []string{"domain-modeling"}, MCP: map[string]Activation{"aws-documentation": {}}})
	if _, err := os.Stat(filepath.Join(worktree, ".codex", "config.toml")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".codex", "config.toml")); !os.IsNotExist(err) {
		t.Fatal("main checkout modified")
	}
	info, err := os.Stat(filepath.Join(worktree, ".git"))
	if err != nil || info.IsDir() {
		t.Fatal("test did not use a linked worktree")
	}
}
func TestCLIRejectsAmbiguousScopeAndCatalogUnknownKeys(t *testing.T) {
	for _, args := range [][]string{nil, {"--global", "--project", "."}, {"--profile", "../escape"}, {"--project", "."}} {
		if _, err := parseOptions("install", args, t.TempDir()); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	e := fixture(t)
	writeTest(t, filepath.Join(e.Root, "selection.toml"), "skils = [\"domain-modeling\"]\n")
	if _, err := readSelection(filepath.Join(e.Root, "selection.toml")); err == nil {
		t.Fatal("misspelled selection accepted")
	}
}

func TestWriteFailureRollsBackEarlierChanges(t *testing.T) {
	base := t.TempDir()
	base, _ = filepath.EvalSymlinks(base)
	first := filepath.Join(base, "first")
	second := filepath.Join(base, "second")
	writeTest(t, first, "before")
	writeTest(t, second, "existing")
	a, _ := snapshot(first)
	b, _ := snapshot(second)
	nextA := a
	nextA.Content = []byte("after")
	nextB := b
	nextB.Link = filepath.Join(base, "target")
	p := Plan{Changes: []Mutation{{Before: a, After: nextA}, {Before: b, After: nextB}}}
	if err := p.Apply(); err == nil {
		t.Fatal("expected second write to fail")
	}
	now, _ := snapshot(first)
	if !sameSnapshot(now, a) {
		t.Fatal("first write was not rolled back")
	}
	now, _ = snapshot(second)
	if !sameSnapshot(now, b) {
		t.Fatal("existing second file changed")
	}
}
func TestSymlinkedConfigAndSkillParentsRefused(t *testing.T) {
	e := fixture(t)
	project := filepath.Join(e.Home, "project")
	outside := filepath.Join(e.Home, "outside")
	if err := os.MkdirAll(project, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(project, ".agents")); err != nil {
		t.Fatal(err)
	}
	if _, err := planInstall(e, Target{Scope: "project", Runtime: "codex", Project: project}, Selection{Skills: []string{"domain-modeling"}}); err == nil {
		t.Fatal("symlinked parent followed")
	}
}
