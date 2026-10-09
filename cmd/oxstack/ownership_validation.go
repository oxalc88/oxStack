package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func validateOwnedResource(e Environment, t Target, r Resource) error {
	if r.Runtime != t.Runtime || r.Scope != t.Scope || !filepath.IsAbs(r.Destination) || r.Digest == "" {
		return fmt.Errorf("invalid ownership identity")
	}
	if strings.HasPrefix(r.Capability, "mcp:") {
		name := strings.TrimPrefix(r.Capability, "mcp:")
		if !validName.MatchString(name) {
			return fmt.Errorf("unsafe owned MCP name")
		}
		dest, key := mcpDestination(e, t, name)
		if r.Destination != dest || fmt.Sprint(r.Key) != fmt.Sprint(key) || r.Kind != map[bool]string{true: "toml", false: "json"}[t.Runtime == "codex"] {
			return fmt.Errorf("MCP ownership points outside its scope")
		}
		return nil
	}
	if strings.HasPrefix(r.Capability, "skill:") {
		name := strings.TrimPrefix(r.Capability, "skill:")
		if !validName.MatchString(name) || t.Scope == "profile" || r.Kind != "link" || r.Destination != skillDestination(e, t, name) || r.Link == "" || r.Digest != fingerprint(r.Link) {
			return fmt.Errorf("skill ownership points outside its scope")
		}
		return nil
	}
	if r.Capability == "instructions" || r.Capability == "instructions-reference" {
		dest := instructionsDestination(e, t)
		kind := "text"
		var key []string
		if t.Scope == "profile" {
			kind = "file"
			if t.Runtime == "codex" {
				dest, _ = mcpDestination(e, t, "")
				kind = "toml"
				key = []string{"developer_instructions"}
			}
			if r.Capability == "instructions-reference" {
				dest, _ = mcpDestination(e, t, "")
				kind = "json"
				key = []string{"instructions"}
			}
		}
		// Recognize the exact previous profile filename so unchanged owned files
		// can be retired during an upgrade. Fingerprint checks still apply.
		if t.Scope == "profile" && t.Runtime != "codex" && r.Capability == "instructions" && r.Destination == filepath.Join(profileDir(e, t), "instructions.md") {
			dest = r.Destination
		}
		if r.Destination != dest || r.Kind != kind || fmt.Sprint(r.Key) != fmt.Sprint(key) {
			return fmt.Errorf("instruction ownership points outside its scope")
		}
		return nil
	}
	return fmt.Errorf("unknown owned capability %q", r.Capability)
}
