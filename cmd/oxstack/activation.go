package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func planInstall(e Environment, t Target, s Selection) (*Plan, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	if t.Scope == "global" && t.Runtime == "codex" {
		legacy, err := os.ReadFile(instructionsDestination(e, t))
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if strings.Contains(string(legacy), generatedHeader) {
			return nil, fmt.Errorf("legacy global instructions detected; preview migrate before installation")
		}
	}
	c, err := loadCatalog(e.Root)
	if err != nil {
		return nil, err
	}
	resources := []Resource{}
	seen := map[string]bool{}
	for _, name := range s.Skills {
		if !validName.MatchString(name) || seen[name] {
			return nil, fmt.Errorf("invalid/duplicate skill %q", name)
		}
		seen[name] = true
		skill, ok := c.Skills[name]
		if !ok {
			return nil, fmt.Errorf("unknown skill %s", name)
		}
		if !fileExists(filepath.Join(skill.Path, "SKILL.md")) {
			return nil, fmt.Errorf("skill %s not imported; use catalog import first", name)
		}
		if t.Scope != "profile" {
			dest := skillDestination(e, t, name)
			resources = append(resources, Resource{Capability: "skill:" + name, Source: skill.Source, Version: skill.Version, Scope: t.Scope, Destination: dest, Runtime: t.Runtime, Kind: "link", Digest: fingerprint(skill.Path), Link: skill.Path})
		}
	}
	names := []string{}
	for name := range s.MCP {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		server, ok := c.MCP[name]
		if !ok {
			return nil, fmt.Errorf("unknown MCP %s", name)
		}
		server, active, err := selectServer(server, s.MCP[name])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if server.Args == nil {
			server.Args = []string{}
		}
		value, register, err := serverValue(t.Runtime, server, active)
		if err != nil {
			return nil, err
		}
		if !register {
			continue
		}
		dest, key := mcpDestination(e, t, name)
		source := server.Source
		if source == "" {
			source = "oxStack/mcp/" + name + ".toml"
		}
		resources = append(resources, makeEntry(t, "mcp:"+name, source, server.Version, dest, key, value))
	}
	var guidance strings.Builder
	for _, name := range s.Guidelines {
		if !validName.MatchString(name) {
			return nil, fmt.Errorf("invalid guideline %q", name)
		}
		b, err := os.ReadFile(filepath.Join(e.Root, "guidelines", name+".md"))
		if err != nil {
			return nil, err
		}
		guidance.Write(b)
		guidance.WriteByte('\n')
	}
	if t.Scope == "profile" && len(s.Skills) > 0 {
		guidance.WriteString(profileInstructions(c, s))
	}
	if guidance.Len() > 0 {
		text := guidance.String()
		if t.Scope == "profile" && t.Runtime == "codex" {
			dest, _ := mcpDestination(e, t, "")
			resources = append(resources, makeEntry(t, "instructions", "selection guidelines and skill references", "working-tree", dest, []string{"developer_instructions"}, text))
		} else if t.Scope == "profile" && t.Runtime == "opencode" {
			path := instructionsDestination(e, t)
			resources = append(resources, Resource{Capability: "instructions", Source: "selection", Version: "working-tree", Scope: t.Scope, Destination: path, Runtime: t.Runtime, Kind: "file", Digest: bytesDigest([]byte(text)), Content: []byte(text)})
			dest, _ := mcpDestination(e, t, "")
			resources = append(resources, makeEntry(t, "instructions-reference", "selection", "working-tree", dest, []string{"instructions"}, []string{path}))
		} else {
			kind := "text"
			if t.Scope == "profile" {
				kind = "file"
			}
			resources = append(resources, Resource{Capability: "instructions", Source: "selection guidelines", Version: "working-tree", Scope: t.Scope, Destination: instructionsDestination(e, t), Runtime: t.Runtime, Kind: kind, Digest: bytesDigest([]byte(text)), Content: []byte(text)})
		}
	}
	if t.Scope == "profile" {
		refs := []CapabilityReference{}
		for _, name := range s.Skills {
			skill := c.Skills[name]
			refs = append(refs, CapabilityReference{Capability: "skill:" + name, Source: skill.Source, Version: skill.Version})
		}
		for i := range resources {
			if resources[i].Capability == "instructions" {
				resources[i].References = refs
			}
		}
	}
	sort.Slice(resources, func(i, j int) bool { return resourceID(resources[i]) < resourceID(resources[j]) })
	return buildPlan(e, t, resources, false)
}
