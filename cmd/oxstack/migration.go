package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

func planMigration(e Environment) (*Plan, error) {
	t := Target{Scope: "global", Runtime: "codex"}
	if s, err := snapshot(t.manifestPath(e)); err != nil {
		return nil, err
	} else if s.Exists {
		return nil, fmt.Errorf("new ownership already exists; use scoped install/uninstall instead")
	}
	p := &Plan{Actions: []string{"KEEP GLOBAL: personal defaults, app MCPs, plugins, trust and runtime state"}}
	path := filepath.Join(e.CodexHome, "config.toml")
	before, err := snapshot(path)
	if err != nil {
		return nil, err
	}
	if before.Link != "" {
		return nil, fmt.Errorf("legacy config is a symlink; review its owner first")
	}
	d, err := parseDocument(before.Content, true)
	if err != nil {
		return nil, err
	}
	var legacy struct {
		MCP struct {
			Servers map[string]map[string]any `toml:"servers"`
		} `toml:"mcp"`
	}
	b, err := os.ReadFile(filepath.Join(e.Root, "migration", "legacy-catalog.toml"))
	if err != nil {
		return nil, err
	}
	if err = toml.Unmarshal(b, &legacy); err != nil {
		return nil, err
	}
	after := before
	marked := strings.Contains(string(before.Content), "# --- oxStack MCP servers ---")
	for _, name := range sortedKeys(legacy.MCP.Servers) {
		key := []string{"mcp_servers", strings.ReplaceAll(name, ".", "-")}
		v, exists := lookup(d, key)
		if !exists {
			continue
		}
		expected := map[string]any{}
		for k, v := range legacy.MCP.Servers[name] {
			if k != "disabled" {
				expected[k] = v
			}
		}
		// The old generator resolves AWS_PROFILE from its repository .env. Compare
		// that exact historical input, never bake it into the new portable catalog.
		identityKnown := true
		if env, ok := expected["env"].(map[string]any); ok {
			envCopy := map[string]any{}
			for k, v := range env {
				if v == "$AWS_PROFILE" {
					value, ok := legacyAWSProfile(e.Root)
					if !ok {
						identityKnown = false
					}
					envCopy[k] = value
				} else {
					envCopy[k] = v
				}
			}
			expected["env"] = envCopy
		}
		if !marked || !identityKnown || fingerprint(v) != fingerprint(expected) {
			p.Actions = append(p.Actions, "REVIEW (not removed): "+name+" identity differs or cannot be proven")
			continue
		}
		after.Content, err = changeTOML(after.Content, key, nil, true)
		if err != nil {
			return nil, err
		}
		p.Actions = append(p.Actions, "REMOVE GLOBAL ACTIVATION: "+name)
	}
	if !sameSnapshot(before, after) {
		p.Changes = append(p.Changes, Mutation{before, after})
	}
	entries, err := os.ReadDir(filepath.Join(e.Root, "skills"))
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dest := filepath.Join(e.Home, ".agents", "skills", entry.Name())
		s, err := snapshot(dest)
		if err != nil {
			// An unowned directory is never removed by legacy migration.
			if dirExists(dest) {
				p.Actions = append(p.Actions, "REVIEW (not removed): skill "+entry.Name())
				continue
			}
			return nil, err
		}
		if !s.Exists {
			continue
		}
		expected := filepath.Join(e.Root, "skills", entry.Name())
		if s.Link != expected {
			p.Actions = append(p.Actions, "REVIEW (not removed): skill "+entry.Name()+" link differs")
			continue
		}
		gone := s
		gone.Exists = false
		gone.Link = ""
		p.Changes = append(p.Changes, Mutation{s, gone})
		p.Actions = append(p.Actions, "REMOVE GLOBAL ACTIVATION: skill "+entry.Name())
	}
	// Copied/external/legacy Codex skills cannot be identified from directory
	// names. Report them without claiming ownership or deleting their contents.
	for _, base := range []string{filepath.Join(e.Home, ".agents", "skills"), filepath.Join(e.CodexHome, "skills")} {
		list, err := os.ReadDir(base)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range list {
			if entry.Name() == ".system" {
				continue
			}
			dest := filepath.Join(base, entry.Name())
			scheduled := false
			for _, m := range p.Changes {
				if m.Before.Path == dest {
					scheduled = true
				}
			}
			if !scheduled {
				p.Actions = append(p.Actions, "REVIEW (not removed): "+dest+"; no legacy ownership proof")
			}
		}
	}
	manifest := Manifest{Format: 1, Scope: "global", Runtime: "codex", Resources: []Resource{}}
	agentPath := instructionsDestination(e, t)
	agent, err := snapshot(agentPath)
	if err != nil {
		return nil, err
	}
	if agent.Exists {
		base, err := os.ReadFile(filepath.Join(e.Root, "guidelines", "base-guidelines.md"))
		if err != nil {
			return nil, err
		}
		appendix, err := os.ReadFile(filepath.Join(e.Root, "guidelines", "agent-appendix.md"))
		if err != nil {
			return nil, err
		}
		expected := generatedHeader + "\n\n" + string(base) + "\n" + string(appendix)
		if agent.Link == "" && string(agent.Content) == expected {
			universal, err := os.ReadFile(filepath.Join(e.Root, "guidelines", "universal.md"))
			if err != nil {
				return nil, err
			}
			content := append(universal, '\n')
			next := agent
			next.Content, err = changeText(nil, content, false)
			if err != nil {
				return nil, err
			}
			p.Changes = append(p.Changes, Mutation{agent, next})
			manifest.Resources = append(manifest.Resources, Resource{Capability: "instructions", Source: "selection guidelines", Version: "working-tree", Scope: "global", Runtime: "codex", Destination: agentPath, Kind: "text", Digest: bytesDigest(content)})
			p.Actions = append(p.Actions, "KEEP GLOBAL: replace exact legacy AGENTS.md with minimal universal guidance")
		} else {
			p.Actions = append(p.Actions, "REVIEW (not rewritten): global AGENTS.md differs from known generator")
		}
	}
	p.Actions = append(p.Actions, "MOVE TO PROJECT (explicit subsequent install): AWS documentation, Powertools, Serverless, CloudFormation, Astro and domain skills", "MOVE TO PROFILE (explicit subsequent install): AWS pricing, diagrams, Bedrock costs, browser workflows")
	if len(p.Changes) == 0 {
		return p, nil
	}
	// Back up every affected input before changing it, including symlink metadata.
	// Backups and the new manifest are part of the same preflighted rollback plan.
	backupDir := filepath.Join(e.StateHome, "migration-backups", time.Now().UTC().Format("20060102T150405.000000000"))
	backups := []Mutation{}
	for i, m := range p.Changes {
		b, err := json.MarshalIndent(m.Before, "", "  ")
		if err != nil {
			return nil, err
		}
		bp := filepath.Join(backupDir, fmt.Sprintf("%03d.json", i))
		backups = append(backups, Mutation{Before: Snapshot{Path: bp, Mode: 0600}, After: Snapshot{Path: bp, Mode: 0600, Exists: true, Content: b}})
	}
	p.Changes = append(backups, p.Changes...)
	b, err = json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	mp := t.manifestPath(e)
	p.Changes = append(p.Changes, Mutation{Before: Snapshot{Path: mp, Mode: 0600}, After: Snapshot{Path: mp, Mode: 0600, Exists: true, Content: append(b, '\n')}})
	p.Actions = append(p.Actions, "BACKUP: "+backupDir)
	return p, nil
}
func legacyAWSProfile(root string) (string, bool) {
	b, err := os.ReadFile(filepath.Join(root, ".env"))
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(k) == "AWS_PROFILE" {
			return strings.TrimSpace(v), true
		}
	}
	return "", false
}
func sortedKeys[V any](m map[string]V) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func cmdMigrate() error {
	f := flag.NewFlagSet("migrate", flag.ContinueOnError)
	apply := f.Bool("apply", false, "require typed confirmation and apply exact legacy matches")
	dry := f.Bool("dry-run", false, "preview only")
	if err := f.Parse(os.Args[2:]); err != nil {
		return err
	}
	if f.NArg() > 0 || (*apply && *dry) {
		return fmt.Errorf("use migrate --dry-run or migrate --apply")
	}
	e, err := newEnvironment(repoRoot())
	if err != nil {
		return err
	}
	p, err := planMigration(e)
	if err != nil {
		return err
	}
	for _, a := range p.Actions {
		fmt.Println(a)
	}
	if !*apply {
		fmt.Println("Preview only. No files changed. Use --apply and typed MIGRATE confirmation later.")
		return nil
	}
	if !confirm(os.Stdin, os.Stdout) {
		return fmt.Errorf("migration cancelled; no files changed")
	}
	return p.Apply()
}
