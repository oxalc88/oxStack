package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadCatalog(root string) (Catalog, error) {
	revision := catalogRevision(root)
	c := Catalog{Skills: map[string]Skill{}, MCP: map[string]Server{}}
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		return c, err
	}
	for _, entry := range entries {
		p := filepath.Join(root, "skills", entry.Name())
		if entry.IsDir() && fileExists(filepath.Join(p, "SKILL.md")) {
			c.Skills[entry.Name()] = Skill{Path: p, Source: "oxStack/skills/" + entry.Name(), Version: revision}
		}
	}
	var cfg Config
	if err := decodeTOML(filepath.Join(root, "oxstack.toml"), &cfg); err != nil {
		return c, err
	}
	for name, s := range cfg.Skills.External {
		if !validName.MatchString(name) {
			return c, fmt.Errorf("invalid skill name %q", name)
		}
		if _, ok := c.Skills[name]; ok {
			return c, fmt.Errorf("duplicate skill %s", name)
		}
		path := filepath.Join(root, "catalog", "external", name)
		version := s.Revision
		if fileExists(filepath.Join(path, "SKILL.md")) {
			b, err := os.ReadFile(filepath.Join(path, ".oxstack-source.json"))
			if err != nil {
				return c, err
			}
			var provenance SkillProvenance
			if err = json.Unmarshal(b, &provenance); err != nil {
				return c, err
			}
			digest, err := treeDigest(path)
			if err != nil {
				return c, err
			}
			if provenance.Repository != s.Repo || provenance.Skill != s.Skill || provenance.Revision == "" || digest != provenance.TreeDigest || (s.Revision != "" && s.Revision != provenance.Revision) {
				return c, fmt.Errorf("external skill provenance changed: %s", name)
			}
			version = provenance.Revision
		}
		c.Skills[name] = Skill{Path: path, Source: s.Repo + "/" + s.Skill, Version: version}
	}
	for _, name := range []string{"qa", "design-review"} {
		p := filepath.Join(root, "gstack-ab", "skills", name)
		if fileExists(filepath.Join(p, "SKILL.md")) {
			c.Skills[name+"-byOx"] = Skill{Path: p, Source: "oxStack/gstack-ab/" + name, Version: revision}
		}
	}
	entries, err = os.ReadDir(filepath.Join(root, "mcp"))
	if err != nil {
		return c, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".toml")
		if !validName.MatchString(name) {
			return c, fmt.Errorf("invalid MCP name %q", name)
		}
		var s Server
		if err := decodeTOML(filepath.Join(root, "mcp", entry.Name()), &s); err != nil {
			return c, err
		}
		if _, _, err := selectServer(s, Activation{}); err != nil {
			return c, fmt.Errorf("%s: %w", name, err)
		}
		c.MCP[name] = s
	}
	return c, nil
}

func catalogRevision(root string) string {
	out, err := runSilentDir(root, "git", "--no-optional-locks", "rev-parse", "HEAD")
	if err != nil {
		return "working-tree"
	}
	revision := strings.TrimSpace(out)
	status, err := runSilentDir(root, "git", "--no-optional-locks", "status", "--porcelain")
	if err == nil && strings.TrimSpace(status) != "" {
		revision += "+working-tree"
	}
	return revision
}
