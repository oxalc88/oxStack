package main

import (
	"os"
	"path/filepath"
)

// Catalog metadata is separate from all runtime selections.
type Config struct {
	Gstack struct {
		ForkedSkills []string `toml:"forked_skills"`
	} `toml:"gstack"`
	Skills struct {
		External map[string]ExternalSkill `toml:"external"`
	} `toml:"skills"`
}
type ExternalSkill struct {
	Repo     string `toml:"repo"`
	Skill    string `toml:"skill"`
	Revision string `toml:"revision"`
}

func loadConfig() *Config {
	var cfg Config
	if err := decodeTOML(filepath.Join(repoRoot(), "oxstack.toml"), &cfg); err != nil {
		errorf("%v", err)
		os.Exit(1)
	}
	return &cfg
}
