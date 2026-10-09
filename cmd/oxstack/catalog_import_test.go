package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExternalImportIsPinnedCatalogOnly(t *testing.T) {
	e := fixture(t)
	source := filepath.Join(e.Home, "upstream")
	writeTest(t, filepath.Join(source, "SKILL.md"), "---\nname: browser\ndescription: browser\n---\nReviewed workflow\n")
	gitTest(t, source, "init")
	gitTest(t, source, "add", "SKILL.md")
	gitTest(t, source, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "initial")
	revision, err := runSilentDir(source, "git", "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	writeTest(t, filepath.Join(e.Root, "oxstack.toml"), "[skills.external.browser]\nrepo = \"example/browser\"\nskill = \"browser\"\nrevision = \""+strings.TrimSpace(revision)+"\"\n")
	if err = importExternal(e.Root, "browser", source); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(e.Home, ".agents", "skills", "browser")); !os.IsNotExist(err) {
		t.Fatal("import activated global skill")
	}
	if _, err = os.Stat(filepath.Join(e.Root, "catalog", "external", "browser", ".git")); !os.IsNotExist(err) {
		t.Fatal("Git metadata copied into skill catalog")
	}
	c, err := loadCatalog(e.Root)
	if err != nil {
		t.Fatal(err)
	}
	if c.Skills["browser"].Version != strings.TrimSpace(revision) {
		t.Fatal("resolved version not recorded")
	}
	target := Target{Scope: "project", Runtime: "codex", Project: filepath.Join(e.Home, "project")}
	applySelection(t, e, target, Selection{Skills: []string{"browser"}})
	writeTest(t, filepath.Join(e.Root, "catalog", "external", "browser", "SKILL.md"), "external mutation")
	if _, err = loadCatalog(e.Root); err == nil {
		t.Fatal("tampered imported snapshot accepted")
	}
}
func TestExternalImportRejectsDirtyOrWrongRevision(t *testing.T) {
	e := fixture(t)
	source := filepath.Join(e.Home, "upstream")
	writeTest(t, filepath.Join(source, "SKILL.md"), "workflow")
	gitTest(t, source, "init")
	gitTest(t, source, "add", "SKILL.md")
	gitTest(t, source, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-m", "initial")
	writeTest(t, filepath.Join(e.Root, "oxstack.toml"), "[skills.external.browser]\nrepo = \"example/browser\"\nskill = \"browser\"\nrevision = \"wrong-revision\"\n")
	if err := importExternal(e.Root, "browser", source); err == nil {
		t.Fatal("wrong revision accepted")
	}
	writeTest(t, filepath.Join(e.Root, "oxstack.toml"), "[skills.external.browser]\nrepo = \"example/browser\"\nskill = \"browser\"\n")
	writeTest(t, filepath.Join(source, "SKILL.md"), "changed")
	if err := importExternal(e.Root, "browser", source); err == nil {
		t.Fatal("dirty source accepted")
	}
}
func TestCLIPropagatesPlanningErrors(t *testing.T) {
	e := fixture(t)
	t.Setenv("CODEX_HOME", e.CodexHome)
	t.Setenv("XDG_STATE_HOME", e.StateHome)
	t.Setenv("OXSTACK_ROOT", e.Root)
	selection := filepath.Join(e.Root, "selection.toml")
	writeTest(t, selection, "skills = [\"does-not-exist\"]\n")
	err := runActivation("install", []string{"--project", filepath.Join(e.Home, "project"), "--selection", selection})
	if err == nil || !strings.Contains(err.Error(), "unknown skill") {
		t.Fatalf("lost planning error: %v", err)
	}
}

func TestLocalSnapshotPreservesUnknownRevisionWithoutActivation(t *testing.T) {
	e := fixture(t)
	source := filepath.Join(e.Home, "installed")
	writeTest(t, filepath.Join(source, "SKILL.md"), "local workflow")
	writeTest(t, filepath.Join(e.Root, "oxstack.toml"), "[skills.external.browser]\nrepo = \"UNKNOWN\"\nskill = \"browser\"\n")
	if err := importExternalMode(e.Root, "browser", source, true); err != nil {
		t.Fatal(err)
	}
	c, err := loadCatalog(e.Root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(c.Skills["browser"].Version, "local-sha256:") {
		t.Fatal("snapshot impersonates Git revision")
	}
	if _, err := os.Stat(filepath.Join(e.Home, ".agents", "skills", "browser")); !os.IsNotExist(err) {
		t.Fatal("snapshot activated globally")
	}
	applySelection(t, e, Target{Scope: "project", Runtime: "codex", Project: filepath.Join(e.Home, "project")}, Selection{Skills: []string{"browser"}})
	writeTest(t, filepath.Join(e.Root, "catalog", "external", "browser", "SKILL.md"), "tampered")
	if _, err := loadCatalog(e.Root); err == nil {
		t.Fatal("tampered snapshot accepted")
	}
}

func TestLocalSnapshotCannotBypassGitPin(t *testing.T) {
	e := fixture(t)
	source := filepath.Join(e.Home, "installed")
	writeTest(t, filepath.Join(source, "SKILL.md"), "workflow")
	writeTest(t, filepath.Join(e.Root, "oxstack.toml"), "[skills.external.browser]\nrepo = \"example/browser\"\nskill = \"browser\"\nrevision = \"reviewed-commit\"\n")
	if err := importExternalMode(e.Root, "browser", source, true); err == nil {
		t.Fatal("Git pin bypassed")
	}
}
