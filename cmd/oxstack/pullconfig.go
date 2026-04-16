package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

func cmdPullConfig() {
	fmt.Printf(bold+"oxstack pull-config"+reset+" — sync MCP disabled flags from ~/.claude/settings.json\n\n")

	root := repoRoot()
	tomlPath := filepath.Join(root, "oxstack.toml")

	// Safety: refuse if oxstack.toml has uncommitted changes.
	if out, _ := runSilentDir(root, "git", "diff", "--quiet", "--", "oxstack.toml"); out != "" {
		errorf("oxstack.toml has uncommitted changes — commit or stash before pulling config")
		os.Exit(1)
	}

	// Read live ~/.claude/settings.json
	liveData, err := os.ReadFile(claudeSettingsPath())
	if err != nil {
		errorf("Could not read %s: %v", claudeSettingsPath(), err)
		os.Exit(1)
	}
	var liveSettings map[string]any
	if err := json.Unmarshal(liveData, &liveSettings); err != nil {
		errorf("Could not parse settings.json: %v", err)
		os.Exit(1)
	}

	liveServers, _ := liveSettings["mcpServers"].(map[string]any)
	if liveServers == nil {
		infof("No mcpServers in settings.json — nothing to pull")
		return
	}

	cfg := loadConfig()

	// Compute per-server diffs (disabled flag only).
	type serverDiff struct {
		name    string
		repoVal any
		liveVal any
	}
	var diffs []serverDiff

	for name, repoServer := range cfg.MCP.Servers {
		repoDisabled := repoServer["disabled"]

		liveEntry, inLive := liveServers[name]
		if !inLive {
			continue
		}
		liveMap, ok := liveEntry.(map[string]any)
		if !ok {
			continue
		}
		liveDisabled := liveMap["disabled"]

		if toBool(repoDisabled) != toBool(liveDisabled) {
			diffs = append(diffs, serverDiff{
				name:    name,
				repoVal: repoDisabled,
				liveVal: liveDisabled,
			})
		}
	}

	if len(diffs) == 0 {
		infof("No changes — oxstack.toml already matches live settings (disabled flags)")
		return
	}

	// Print diff
	fmt.Printf(bold+"Differences (disabled flags):"+reset+"\n\n")
	for _, d := range diffs {
		fmt.Printf("  %s\n", d.name)
		fmt.Printf(red+"    toml: disabled = %v"+reset+"\n", boolStr(d.repoVal))
		fmt.Printf(green+"    live: disabled = %v"+reset+"\n", boolStr(d.liveVal))
		fmt.Println()
	}

	// Prompt
	fmt.Printf("Apply these changes to oxstack.toml? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		warnf("Aborted — no changes written")
		return
	}

	// Load raw TOML as a generic map to preserve comments and key ordering.
	rawData, err := os.ReadFile(tomlPath)
	if err != nil {
		errorf("Could not read %s: %v", tomlPath, err)
		os.Exit(1)
	}
	var raw map[string]any
	if err := toml.Unmarshal(rawData, &raw); err != nil {
		errorf("Could not parse oxstack.toml: %v", err)
		os.Exit(1)
	}

	// Navigate to mcp.servers and update disabled flags.
	mcpSection, _ := raw["mcp"].(map[string]any)
	if mcpSection == nil {
		errorf("No [mcp] section found in oxstack.toml")
		os.Exit(1)
	}
	serversSection, _ := mcpSection["servers"].(map[string]any)
	if serversSection == nil {
		errorf("No [mcp.servers] section found in oxstack.toml")
		os.Exit(1)
	}

	for _, d := range diffs {
		serverEntry, _ := serversSection[d.name].(map[string]any)
		if serverEntry == nil {
			continue
		}
		live, _ := liveServers[d.name].(map[string]any)
		if toBool(live["disabled"]) {
			serverEntry["disabled"] = true
		} else {
			delete(serverEntry, "disabled")
		}
	}

	// Write back with comment preservation.
	out, err := toml.Marshal(raw)
	if err != nil {
		errorf("Could not marshal oxstack.toml: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(tomlPath, out, 0o644); err != nil {
		errorf("Could not write %s: %v", tomlPath, err)
		os.Exit(1)
	}

	infof("Updated oxstack.toml (%d server(s) changed)", len(diffs))
	fmt.Printf("\n  %s\n", dim+"git diff oxstack.toml && git add oxstack.toml && git commit -m 'chore(mcp): sync disabled flags'"+reset)
}

func toBool(v any) bool {
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func boolStr(v any) string {
	if v == nil {
		return "false (absent)"
	}
	return fmt.Sprintf("%v", v)
}
