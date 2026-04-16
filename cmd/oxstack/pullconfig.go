package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cmdPullConfig() {
	fmt.Printf(bold+"oxstack pull-config"+reset+" — sync MCP disabled flags from ~/.claude/settings.json\n\n")

	root := repoRoot()
	mcpPath := filepath.Join(root, "mcp", "claude.json")

	// Safety: refuse if mcp/claude.json has uncommitted changes.
	if dirty, _ := runSilentDir(root, "git", "diff", "--quiet", "--", "mcp/claude.json"); dirty != "" {
		errorf("mcp/claude.json has uncommitted changes — commit or stash before pulling config")
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

	// Read mcp/claude.json
	repoData, err := os.ReadFile(mcpPath)
	if err != nil {
		errorf("Could not read %s: %v", mcpPath, err)
		os.Exit(1)
	}
	var repoConfig map[string]any
	if err := json.Unmarshal(repoData, &repoConfig); err != nil {
		errorf("Could not parse mcp/claude.json: %v", err)
		os.Exit(1)
	}

	liveServers, _ := liveSettings["mcpServers"].(map[string]any)
	repoServers, _ := repoConfig["mcpServers"].(map[string]any)

	if liveServers == nil {
		infof("No mcpServers in settings.json — nothing to pull")
		return
	}
	if repoServers == nil {
		errorf("No mcpServers in mcp/claude.json — unexpected")
		os.Exit(1)
	}

	// Compute per-server diffs (disabled flag only).
	type serverDiff struct {
		name    string
		repoVal any // current value in repo (nil = absent)
		liveVal any // current value in live settings (nil = absent)
	}
	var diffs []serverDiff

	for name, repoEntry := range repoServers {
		repoMap, ok := repoEntry.(map[string]any)
		if !ok {
			continue
		}
		repoDisabled := repoMap["disabled"]

		liveEntry, inLive := liveServers[name]
		if !inLive {
			continue // server not yet merged into live settings — skip
		}
		liveMap, ok := liveEntry.(map[string]any)
		if !ok {
			continue
		}
		liveDisabled := liveMap["disabled"]

		// Normalize: missing key == false
		repoBool := toBool(repoDisabled)
		liveBool := toBool(liveDisabled)

		if repoBool != liveBool {
			diffs = append(diffs, serverDiff{
				name:    name,
				repoVal: repoDisabled,
				liveVal: liveDisabled,
			})
		}
	}

	if len(diffs) == 0 {
		infof("No changes — mcp/claude.json already matches live settings (disabled flags)")
		return
	}

	// Print diff
	fmt.Printf(bold+"Differences (disabled flags):"+reset+"\n\n")
	for _, d := range diffs {
		fmt.Printf("  %s\n", d.name)
		fmt.Printf(red+"    repo:  disabled = %v"+reset+"\n", boolStr(d.repoVal))
		fmt.Printf(green+"    live:  disabled = %v"+reset+"\n", boolStr(d.liveVal))
		fmt.Println()
	}

	// Prompt
	fmt.Printf("Apply these changes to mcp/claude.json? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		warnf("Aborted — no changes written")
		return
	}

	// Apply: set disabled flag on matching servers in repo config.
	for _, d := range diffs {
		serverMap, ok := repoServers[d.name].(map[string]any)
		if !ok {
			continue
		}
		live, _ := liveServers[d.name].(map[string]any)
		liveDisabled := live["disabled"]
		if liveDisabled == nil || liveDisabled == false {
			delete(serverMap, "disabled")
		} else {
			serverMap["disabled"] = true
		}
	}

	// Write back mcp/claude.json
	out, err := json.MarshalIndent(repoConfig, "", "  ")
	if err != nil {
		errorf("Could not marshal mcp/claude.json: %v", err)
		os.Exit(1)
	}
	if err := os.WriteFile(mcpPath, append(out, '\n'), 0o644); err != nil {
		errorf("Could not write %s: %v", mcpPath, err)
		os.Exit(1)
	}

	infof("Updated mcp/claude.json (%d server(s) changed)", len(diffs))
	fmt.Printf("\n  %s\n", dim+"git diff mcp/claude.json && git add mcp/claude.json && git commit -m 'chore(mcp): sync disabled flags'"+reset)
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
