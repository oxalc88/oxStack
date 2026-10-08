package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Environment struct{ Root, Home, CodexHome, StateHome, ConfigHome string }
type Target struct{ Scope, Runtime, Project, Profile string }
type Selection struct {
	Skills     []string              `toml:"skills"`
	Guidelines []string              `toml:"guidelines"`
	MCP        map[string]Activation `toml:"mcp"`
}
type Activation struct {
	Active   *bool             `toml:"active"`
	Disabled *bool             `toml:"disabled"` // Legacy input only; adapters use a canonical bool.
	Env      map[string]string `toml:"env"`
}
type Server struct {
	InheritEnv []string          `toml:"inherit_env"`
	Command    string            `toml:"command"`
	Args       []string          `toml:"args"`
	URL        string            `toml:"url"`
	Env        map[string]string `toml:"env"`
	Active     *bool             `toml:"active"`
	Disabled   *bool             `toml:"disabled"`
	Source     string            `toml:"source"`
	Version    string            `toml:"version"`
}
type Skill struct{ Path, Source, Version string }
type Catalog struct {
	Skills map[string]Skill
	MCP    map[string]Server
}

var validEnvName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

func newEnvironment(root string) (Environment, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return Environment{}, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return Environment{}, err
	}
	if r, e := filepath.EvalSymlinks(root); e == nil {
		root = r
	}
	ch := os.Getenv("CODEX_HOME")
	if ch == "" {
		ch = filepath.Join(h, ".codex")
	}
	sh := os.Getenv("XDG_STATE_HOME")
	if sh == "" {
		sh = filepath.Join(h, ".local", "state")
	}
	if !filepath.IsAbs(ch) || !filepath.IsAbs(sh) {
		return Environment{}, fmt.Errorf("CODEX_HOME and XDG_STATE_HOME must be absolute")
	}
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(h, ".config")
	}
	if !filepath.IsAbs(configHome) {
		return Environment{}, fmt.Errorf("XDG_CONFIG_HOME must be absolute")
	}
	ch, err = canonicalLocation(ch)
	if err != nil {
		return Environment{}, err
	}
	sh, err = canonicalLocation(sh)
	if err != nil {
		return Environment{}, err
	}
	configHome, err = canonicalLocation(configHome)
	if err != nil {
		return Environment{}, err
	}
	return Environment{Root: root, Home: h, CodexHome: ch, StateHome: filepath.Join(sh, "oxstack"), ConfigHome: configHome}, nil
}
func (t Target) validate() error {
	if t.Runtime != "codex" && t.Runtime != "claude" && t.Runtime != "opencode" {
		return fmt.Errorf("runtime must be codex, claude or opencode")
	}
	switch t.Scope {
	case "global":
		if t.Project != "" || t.Profile != "" {
			return fmt.Errorf("global cannot include project/profile")
		}
	case "project":
		if !filepath.IsAbs(t.Project) || t.Profile != "" {
			return fmt.Errorf("project requires an absolute directory")
		}
	case "profile":
		if !validName.MatchString(t.Profile) || t.Project != "" {
			return fmt.Errorf("profile requires a safe name")
		}
	default:
		return fmt.Errorf("select exactly one scope: --global, --project DIR, or --profile NAME")
	}
	return nil
}
func (t Target) manifestPath(e Environment) string {
	if t.Scope == "project" {
		return filepath.Join(t.Project, ".oxstack", "ownership-"+t.Runtime+".json")
	}
	name := t.Scope
	if t.Scope == "profile" {
		name += "-" + t.Profile
	}
	return filepath.Join(e.StateHome, name+"-"+t.Runtime+".json")
}
func decodeTOML(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return toml.NewDecoder(strings.NewReader(string(b))).DisallowUnknownFields().Decode(v)
}
func readSelection(path string) (Selection, error) {
	var s Selection
	err := decodeTOML(path, &s)
	return s, err
}
func activeState(active, disabled *bool, fallback bool) (bool, error) {
	if active != nil && disabled != nil {
		return false, fmt.Errorf("use active or legacy disabled, not both")
	}
	if active != nil {
		return *active, nil
	}
	if disabled != nil {
		return !*disabled, nil
	}
	return fallback, nil
}
func selectServer(s Server, a Activation) (Server, bool, error) {
	base, err := activeState(s.Active, s.Disabled, true)
	if err != nil {
		return s, false, err
	}
	active, err := activeState(a.Active, a.Disabled, base)
	if err != nil {
		return s, false, err
	}
	env := map[string]string{}
	for k, v := range s.Env {
		env[k] = v
	}
	for k, v := range a.Env {
		env[k] = v
	}
	s.Env = env
	if (s.Command == "") == (s.URL == "") {
		return s, false, fmt.Errorf("server requires exactly one command or URL")
	}
	return s, active, nil
}

// Resolve symlinks in an explicitly supplied location, including macOS /var,
// before appending not-yet-created directories. Destination parents created
// later are still checked before every write.
func canonicalLocation(path string) (string, error) {
	parent := filepath.Clean(path)
	suffix := []string{}
	for {
		_, err := os.Lstat(parent)
		if err == nil {
			resolved, err := filepath.EvalSymlinks(parent)
			if err != nil {
				return "", err
			}
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return resolved, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		if parent == filepath.Dir(parent) {
			return "", fmt.Errorf("no existing parent for %s", path)
		}
		suffix = append(suffix, filepath.Base(parent))
		parent = filepath.Dir(parent)
	}
}
