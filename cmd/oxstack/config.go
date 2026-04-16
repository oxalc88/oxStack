package main

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config is the typed representation of oxstack.toml.
type Config struct {
	Gstack struct {
		ForkedSkills []string `toml:"forked_skills"`
	} `toml:"gstack"`
	Skills struct {
		External map[string]ExternalSkill `toml:"external"`
	} `toml:"skills"`
	MCP struct {
		Servers map[string]map[string]any `toml:"servers"`
	} `toml:"mcp"`
}

// ExternalSkill describes a third-party skill installable via `npx skills add`.
type ExternalSkill struct {
	Repo  string `toml:"repo"`
	Skill string `toml:"skill"`
}

// loadConfig reads and parses oxstack.toml from the repo root.
func loadConfig() *Config {
	path := filepath.Join(repoRoot(), "oxstack.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		errorf("Could not read oxstack.toml: %v", err)
		os.Exit(1)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		errorf("Could not parse oxstack.toml: %v", err)
		os.Exit(1)
	}
	return &cfg
}
