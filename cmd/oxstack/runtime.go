package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func profileDir(e Environment, t Target) string {
	return filepath.Join(e.StateHome, "profiles", t.Profile)
}
func mcpDestination(e Environment, t Target, name string) (string, []string) {
	switch t.Runtime {
	case "codex":
		p := filepath.Join(e.CodexHome, "config.toml")
		if t.Scope == "project" {
			p = filepath.Join(t.Project, ".codex", "config.toml")
		}
		if t.Scope == "profile" {
			p = filepath.Join(e.CodexHome, t.Profile+".config.toml")
		}
		return p, []string{"mcp_servers", name}
	case "claude":
		p := filepath.Join(e.Home, ".claude.json")
		if t.Scope == "project" {
			p = filepath.Join(t.Project, ".mcp.json")
		}
		if t.Scope == "profile" {
			p = filepath.Join(profileDir(e, t), "claude.json")
		}
		return p, []string{"mcpServers", name}
	default:
		p := filepath.Join(e.Home, ".config", "opencode", "opencode.json")
		if x := e.ConfigHome; x != "" {
			p = filepath.Join(x, "opencode", "opencode.json")
		}
		if t.Scope == "project" {
			p = filepath.Join(t.Project, "opencode.json")
		}
		if t.Scope == "profile" {
			p = filepath.Join(profileDir(e, t), "opencode.json")
		}
		return p, []string{"mcp", name}
	}
}
func serverValue(runtime string, s Server, active bool) (map[string]any, bool, error) {
	// Claude has no portable per-definition disabled flag. An inactive selection
	// is absent from registration instead of using an ignored custom field.
	if runtime == "claude" && !active {
		return nil, false, nil
	}
	if len(s.InheritEnv) > 0 && s.URL != "" {
		return nil, false, fmt.Errorf("inherited process environment requires a stdio MCP")
	}
	v := map[string]any{}
	inherited := []string{}
	for _, key := range s.InheritEnv {
		if !validEnvName.MatchString(key) {
			return nil, false, fmt.Errorf("invalid inherited environment variable %q", key)
		}
		if _, explicit := s.Env[key]; !explicit {
			inherited = append(inherited, key)
		}
	}
	if s.Env == nil {
		s.Env = map[string]string{}
	}
	switch runtime {
	case "claude":
		for _, key := range inherited {
			s.Env[key] = "${" + key + "}"
		}
	case "opencode":
		for _, key := range inherited {
			s.Env[key] = "{env:" + key + "}"
		}
	}

	switch runtime {
	case "codex":
		v["enabled"] = active
		if len(inherited) > 0 {
			v["env_vars"] = inherited
		}
		if s.URL != "" {
			v["url"] = s.URL
			if len(s.Env) > 0 {
				return nil, false, fmt.Errorf("Codex HTTP MCP cannot carry process env")
			}
		} else {
			v["command"] = s.Command
			v["args"] = s.Args
			if len(s.Env) > 0 {
				v["env"] = s.Env
			}
		}
	case "claude":
		if s.URL != "" {
			v["type"] = "http"
			v["url"] = s.URL
			if len(s.Env) > 0 {
				return nil, false, fmt.Errorf("Claude HTTP MCP cannot carry process env")
			}
		} else {
			v["type"] = "stdio"
			v["command"] = s.Command
			v["args"] = s.Args
			if len(s.Env) > 0 {
				v["env"] = s.Env
			}
		}
	case "opencode":
		v["enabled"] = active
		if s.URL != "" {
			v["type"] = "remote"
			v["url"] = s.URL
			if len(s.Env) > 0 {
				return nil, false, fmt.Errorf("OpenCode remote MCP cannot carry process env")
			}
		} else {
			v["type"] = "local"
			v["command"] = append([]string{s.Command}, s.Args...)
			if len(s.Env) > 0 {
				v["environment"] = s.Env
			}
		}
	default:
		return nil, false, fmt.Errorf("unknown runtime %s", runtime)
	}
	return v, true, nil
}
func skillDestination(e Environment, t Target, name string) string {
	base := e.Home
	if t.Scope == "project" {
		base = t.Project
	}
	if t.Runtime == "opencode" {
		if t.Scope == "project" {
			return filepath.Join(t.Project, ".opencode", "skills", name)
		}
		config := e.ConfigHome
		if config == "" {
			config = filepath.Join(e.Home, ".config")
		}
		return filepath.Join(config, "opencode", "skills", name)
	}
	folder := filepath.Join(".agents", "skills")
	if t.Runtime == "claude" {
		folder = filepath.Join(".claude", "skills")
	}
	return filepath.Join(base, folder, name)
}
func instructionsDestination(e Environment, t Target) string {
	if t.Scope == "profile" {
		return filepath.Join(profileDir(e, t), "instructions-"+t.Runtime+".md")
	}
	if t.Runtime == "claude" {
		if t.Scope == "project" {
			return filepath.Join(t.Project, "CLAUDE.md")
		}
		return filepath.Join(e.Home, ".claude", "CLAUDE.md")
	}
	if t.Scope == "project" {
		return filepath.Join(t.Project, "AGENTS.md")
	}
	if t.Runtime == "codex" {
		return filepath.Join(e.CodexHome, "AGENTS.md")
	}
	base := filepath.Join(e.Home, ".config")
	if x := e.ConfigHome; x != "" {
		base = x
	}
	return filepath.Join(base, "opencode", "AGENTS.md")
}
func profileInstructions(c Catalog, s Selection) string {
	var b strings.Builder
	b.WriteString("Explicitly selected oxStack workflows. Read the named SKILL.md only when that workflow is requested:\n")
	for _, name := range s.Skills {
		fmt.Fprintf(&b, "- %s: %s\n", name, filepath.Join(c.Skills[name].Path, "SKILL.md"))
	}
	return b.String()
}
