package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SkillProvenance struct {
	Kind       string `json:"kind,omitempty"`
	SourcePath string `json:"source_path,omitempty"`
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	Skill      string `json:"skill"`
	TreeDigest string `json:"tree_digest"`
}

func treeDigest(root string) (string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ".git" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("catalog import refuses symlinks: %s", path)
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if rel == ".oxstack-source.json" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = bytesDigest(b)
		return nil
	})
	return fingerprint(files), err
}
func importExternal(root, name, source string) error {
	return importExternalMode(root, name, source, false)
}

func importExternalMode(root, name, source string, snapshot bool) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid skill name")
	}
	var cfg Config
	if err := decodeTOML(filepath.Join(root, "oxstack.toml"), &cfg); err != nil {
		return err
	}
	skill, ok := cfg.Skills.External[name]
	if !ok {
		return fmt.Errorf("unknown external skill %s", name)
	}
	source, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	if !fileExists(filepath.Join(source, "SKILL.md")) {
		return fmt.Errorf("source must be a skill directory containing SKILL.md")
	}
	revision := ""
	kind := ""
	if snapshot {
		if skill.Revision != "" {
			return fmt.Errorf("local snapshot cannot satisfy a Git revision pin")
		}
		kind = "local-snapshot"
	} else {
		revision, err = runSilentDir(source, "git", "rev-parse", "HEAD")
		if err != nil {
			return fmt.Errorf("external source must be a Git checkout: %w", err)
		}
		revision = strings.TrimSpace(revision)
		if skill.Revision != "" && skill.Revision != revision {
			return fmt.Errorf("source revision differs from catalog pin")
		}
		status, err := runSilentDir(source, "git", "status", "--porcelain", "--", source)
		if err != nil {
			return err
		}
		if strings.TrimSpace(status) != "" {
			return fmt.Errorf("source skill has uncommitted changes")
		}
	}
	digest, err := treeDigest(source)
	if err != nil {
		return err
	}
	dest := filepath.Join(root, "catalog", "external", name)
	if _, err := os.Lstat(dest); !os.IsNotExist(err) {
		return fmt.Errorf("catalog destination exists; review/remove it explicitly before importing")
	}
	if err := safeParents(dest); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(dest), ".oxstack-import-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	err = filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ".git" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(source, path)
		p := filepath.Join(stage, rel)
		if d.IsDir() {
			return os.MkdirAll(p, 0700)
		}
		return copyFile(path, p)
	})
	if err != nil {
		return err
	}
	p := SkillProvenance{Kind: kind, SourcePath: source, Repository: skill.Repo, Revision: revision, Skill: skill.Skill, TreeDigest: digest}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(stage, ".oxstack-source.json"), b, 0600); err != nil {
		return err
	}
	if check, err := treeDigest(stage); err != nil || check != digest {
		return fmt.Errorf("source changed during import")
	}
	return os.Rename(stage, dest)
}
func cmdCatalog() error {
	args := os.Args[2:]
	if len(args) < 2 || args[0] != "import" {
		return fmt.Errorf("usage: oxstack catalog import NAME --source /reviewed/checkout/skill")
	}
	f := flag.NewFlagSet("catalog import", flag.ContinueOnError)
	source := f.String("source", "", "reviewed local skill directory")
	snapshot := f.Bool("snapshot", false, "preserve a local copy with unknown upstream revision; never bypasses Git pins")
	if err := f.Parse(args[2:]); err != nil {
		return err
	}
	if *source == "" || f.NArg() > 0 {
		return fmt.Errorf("--source is required")
	}
	if err := importExternalMode(repoRoot(), args[1], *source, *snapshot); err != nil {
		return err
	}
	fmt.Println("Imported catalog only; no runtime activated.")
	return nil
}
