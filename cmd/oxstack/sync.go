package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var forkedSkills = []string{"qa", "design-review"}

func cmdSync() {
	fmt.Printf(bold+"oxstack sync"+reset+" — checking gstack for updates to forked skills\n\n")

	gstack := gstackDir()
	if !dirExists(gstack) {
		errorf("gstack not found at %s — run 'oxstack install' first", gstack)
		os.Exit(1)
	}

	// Pull latest gstack
	infof("Pulling latest gstack...")
	if out, err := runDir(gstack, "git", "pull", "--ff-only"); err != nil {
		warnf("Could not pull (offline or conflict): %s", strings.TrimSpace(out))
	}
	fmt.Println()

	// Get current and last-synced commits
	currentShort, err := runSilentDir(gstack, "git", "rev-parse", "--short", "HEAD")
	if err != nil {
		errorf("Could not read gstack commit")
		os.Exit(1)
	}
	currentShort = strings.TrimSpace(currentShort)

	currentFull, err := runSilentDir(gstack, "git", "rev-parse", "HEAD")
	if err != nil {
		errorf("Could not read gstack commit")
		os.Exit(1)
	}
	currentFull = strings.TrimSpace(currentFull)

	syncFile := syncFilePath()
	lastSync := ""
	if data, err := os.ReadFile(syncFile); err == nil {
		lastSync = strings.TrimSpace(string(data))
	}

	if lastSync == currentFull {
		infof("Already synced to gstack @ %s — no new changes", currentShort)
		return
	}

	// Show commits since last sync
	if lastSync != "" {
		fmt.Printf(bold+"gstack commits since last sync:"+reset+"\n")
		out, _ := runSilentDir(gstack, "git", "log", "--oneline", lastSync+"..HEAD")
		fmt.Print(out)
		fmt.Println()
	} else {
		warnf("No previous sync recorded — showing current state")
		fmt.Println()
	}

	// Diff each forked skill (methodology only — strips boilerplate)
	changesFound := false
	for _, skill := range forkedSkills {
		gstackFile := fmt.Sprintf("%s/%s/SKILL.md", gstack, skill)
		if !fileExists(gstackFile) {
			warnf("%s: not found in gstack (skipping)", skill)
			continue
		}

		if lastSync != "" {
			diffRange := lastSync + "..HEAD"
			out, _ := runSilentDir(gstack, "git", "diff", diffRange, "--", skill+"/SKILL.md")
			diff := strings.TrimSpace(out)

			if diff != "" {
				filtered := strings.TrimSpace(filterBoilerplate(diff))
				if filtered != "" {
					changesFound = true
					fmt.Printf(bold+"=== %s: methodology changes since last sync ==="+reset+"\n", skill)
					fmt.Println(filtered)
					fmt.Println()
				} else {
					infof("%s: only boilerplate changed (no methodology updates)", skill)
				}
			} else {
				infof("%s: no changes in gstack since last sync", skill)
			}
		} else {
			changesFound = true
			fmt.Printf(bold+"=== %s: gstack has updates ==="+reset+"\n", skill)
			fmt.Println("  Run 'oxstack update' to regenerate your -byOx skills from gstack's latest.")
			fmt.Println()
		}
	}

	if !changesFound {
		infof("No methodology changes in forked skills")
	} else {
		fmt.Printf(yellow+"Run 'oxstack update' to apply gstack methodology changes to your -byOx skills."+reset+"\n\n")
	}

	// Prompt to mark as synced
	fmt.Printf("Mark gstack @ %s as synced? [y/N] ", currentShort)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "y" || answer == "yes" {
		if err := os.WriteFile(syncFile, []byte(currentFull+"\n"), 0o644); err != nil {
			errorf("Could not save sync point: %v", err)
		} else {
			infof("Saved sync point: %s", currentShort)
		}
	}
}
