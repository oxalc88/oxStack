package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var reHunkHeader = regexp.MustCompile(`@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

type diffHunk struct {
	oldStart, oldCount int
	newStart, newCount int
	removed            []string
	added              []string
}

func cmdUpdate() {
	fullDiff := false
	for _, arg := range os.Args[2:] {
		if arg == "--full" {
			fullDiff = true
		}
	}

	fmt.Printf(bold+"oxstack update"+reset+" — regenerate -byOx skills from gstack methodology\n\n")

	gstack := gstackDir()
	if !dirExists(gstack) {
		errorf("gstack not found at %s — run 'oxstack install' first", gstack)
		os.Exit(1)
	}

	root := repoRoot()
	skillsDir := claudeSkillsDir()
	anyUpdated := false

	for _, skill := range forkedSkills {
		fmt.Printf(bold+"--- %s ---"+reset+"\n", skill)

		// a. Read gstack's current SKILL.md
		gstackFile := filepath.Join(gstack, skill, "SKILL.md")
		if !fileExists(gstackFile) {
			warnf("%s: not found in gstack (skipping)", skill)
			fmt.Println()
			continue
		}

		// b. Extract methodology (strip frontmatter + boilerplate sections)
		methodology, err := extractMethodology(gstackFile)
		if err != nil {
			errorf("%s: could not extract methodology: %v", skill, err)
			fmt.Println()
			continue
		}

		// c. Transform $B → agent-browser
		var transformed strings.Builder
		for _, line := range strings.Split(methodology, "\n") {
			transformed.WriteString(transformBrowse(line))
			transformed.WriteByte('\n')
		}
		newMethodology := transformed.String()

		// d. Get header from current -byOx SKILL.md
		byOxSrc := filepath.Join(root, "gstack-ab", "skills", skill, "SKILL.md")
		if !fileExists(byOxSrc) {
			warnf("%s: -byOx source not found at %s (skipping)", skill, byOxSrc)
			fmt.Println()
			continue
		}

		header, err := getHeader(byOxSrc, skill)
		if err != nil {
			errorf("%s: could not read header: %v", skill, err)
			fmt.Println()
			continue
		}

		// e. Combine header + transformed methodology
		newContent := header + newMethodology

		// f. Read current content for comparison
		currentContent, err := os.ReadFile(byOxSrc)
		if err != nil {
			errorf("%s: could not read current SKILL.md: %v", skill, err)
			fmt.Println()
			continue
		}

		if string(currentContent) == newContent {
			infof("%s: already up to date", skill)
			fmt.Println()
			continue
		}

		// Show diff
		showUpdateDiff(skill, string(currentContent), newContent, fullDiff)

		// g. Prompt user
		fmt.Printf("\nApply update to %s-byOx? [y/N] ", skill)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" && answer != "yes" {
			warnf("%s: skipped", skill)
			fmt.Println()
			continue
		}

		// h. Write to gstack-ab source
		if err := os.WriteFile(byOxSrc, []byte(newContent), 0o644); err != nil {
			errorf("%s: could not write %s: %v", skill, byOxSrc, err)
			fmt.Println()
			continue
		}
		infof("%s: updated %s", skill, byOxSrc)

		// Re-copy to ~/.claude/skills/<skill>-byOx/
		byOxDst := filepath.Join(skillsDir, skill+"-byOx", "SKILL.md")
		os.MkdirAll(filepath.Dir(byOxDst), 0o755)
		if err := os.WriteFile(byOxDst, []byte(newContent), 0o644); err != nil {
			errorf("%s: could not copy to %s: %v", skill, byOxDst, err)
		} else {
			infof("%s: copied to %s", skill, byOxDst)
		}

		anyUpdated = true
		fmt.Println()
	}

	// Save sync point
	if anyUpdated {
		gstackFull, err := runSilentDir(gstack, "git", "rev-parse", "HEAD")
		if err == nil {
			syncFile := syncFilePath()
			os.WriteFile(syncFile, []byte(strings.TrimSpace(gstackFull)+"\n"), 0o644)
			gstackShort, _ := runSilentDir(gstack, "git", "rev-parse", "--short", "HEAD")
			infof("Saved sync point: %s", strings.TrimSpace(gstackShort))
		}
	}
}

// showUpdateDiff displays a two-part diff: change summary + colored condensed diff.
// With fullDiff=true, falls back to raw unified diff (old behavior).
func showUpdateDiff(skill, oldContent, newContent string, fullDiff bool) {
	if fullDiff {
		out := runDiffCommand(oldContent, newContent, 3)
		fmt.Printf(bold+"Changes for %s-byOx:"+reset+"\n", skill)
		if out != "" {
			fmt.Print(out)
		} else {
			infof("(no visible diff)")
		}
		return
	}

	hunks := computeDiffHunks(oldContent, newContent)
	if len(hunks) == 0 {
		fmt.Printf(bold+"Changes for %s-byOx:"+reset+"\n", skill)
		infof("(no visible diff)")
		return
	}

	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// Part 1: Change summary
	showChangeSummary(skill, hunks, oldLines, newLines)

	// Part 2: Colored condensed diff
	showColoredDiff(hunks, newLines)
}

// runDiffCommand writes old/new to temp files, runs diff with the given context lines, and returns the output.
func runDiffCommand(oldContent, newContent string, contextLines int) string {
	oldPath, newPath, cleanup := writeTempPair(oldContent, newContent)
	if cleanup == nil {
		return ""
	}
	defer cleanup()

	args := []string{fmt.Sprintf("--unified=%d", contextLines)}
	if contextLines > 0 {
		args = append(args, "--label", "current", "--label", "updated")
	}
	args = append(args, oldPath, newPath)
	out, _ := runSilent("diff", args...)
	return out
}

// computeDiffHunks shells out to diff --unified=0 and parses the output.
func computeDiffHunks(oldContent, newContent string) []diffHunk {
	out := runDiffCommand(oldContent, newContent, 0)
	return parseDiffHunks(out)
}

// writeTempPair creates two temp files with old/new content. Returns paths and a cleanup function.
// Returns empty strings and nil cleanup on error.
func writeTempPair(oldContent, newContent string) (oldPath, newPath string, cleanup func()) {
	oldFile, err := os.CreateTemp("", "oxstack-old-*.md")
	if err != nil {
		return "", "", nil
	}
	oldFile.WriteString(oldContent)
	oldFile.Close()

	newFile, err := os.CreateTemp("", "oxstack-new-*.md")
	if err != nil {
		os.Remove(oldFile.Name())
		return "", "", nil
	}
	newFile.WriteString(newContent)
	newFile.Close()

	return oldFile.Name(), newFile.Name(), func() {
		os.Remove(oldFile.Name())
		os.Remove(newFile.Name())
	}
}

// parseDiffHunks parses unified diff output into structured hunks.
func parseDiffHunks(output string) []diffHunk {
	var hunks []diffHunk
	lines := strings.Split(output, "\n")

	for i := 0; i < len(lines); i++ {
		m := reHunkHeader.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}

		h := diffHunk{
			oldStart: atoi(m[1]),
			oldCount: 1,
			newStart: atoi(m[3]),
			newCount: 1,
		}
		if m[2] != "" {
			h.oldCount = atoi(m[2])
		}
		if m[4] != "" {
			h.newCount = atoi(m[4])
		}

		// Collect removed and added lines following the header
		for i+1 < len(lines) {
			next := lines[i+1]
			if strings.HasPrefix(next, "-") && !strings.HasPrefix(next, "---") {
				h.removed = append(h.removed, next[1:])
				i++
			} else if strings.HasPrefix(next, "+") && !strings.HasPrefix(next, "+++") {
				h.added = append(h.added, next[1:])
				i++
			} else {
				break
			}
		}

		hunks = append(hunks, h)
	}

	return hunks
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// showChangeSummary prints the section-level change summary.
func showChangeSummary(skill string, hunks []diffHunk, oldLines, newLines []string) {
	modified, added, removed := classifySections(oldLines, newLines, hunks)

	totalAdded := 0
	totalRemoved := 0
	for _, h := range hunks {
		totalAdded += len(h.added)
		totalRemoved += len(h.removed)
	}

	fmt.Printf(bold+"Changes for %s-byOx:"+reset+"\n", skill)

	// Section counts
	var parts []string
	if len(modified) > 0 {
		parts = append(parts, fmt.Sprintf("%d sections modified", len(modified)))
	}
	if len(added) > 0 {
		parts = append(parts, fmt.Sprintf("%d sections added", len(added)))
	}
	if len(removed) > 0 {
		parts = append(parts, fmt.Sprintf("%d sections removed", len(removed)))
	}
	if len(parts) == 0 {
		parts = append(parts, "changes outside sections")
	}
	fmt.Printf("  %s\n", strings.Join(parts, ", "))
	fmt.Printf("  "+green+"+%d lines"+reset+", "+red+"-%d lines"+reset+"\n", totalAdded, totalRemoved)

	if len(modified) > 0 {
		fmt.Printf("\n  " + bold + "Modified:" + reset + "\n")
		for _, s := range modified {
			fmt.Printf("    %s\n", s)
		}
	}
	if len(added) > 0 {
		fmt.Printf("\n  " + bold + "Added:" + reset + "\n")
		for _, s := range added {
			fmt.Printf("    %s\n", s)
		}
	}
	if len(removed) > 0 {
		fmt.Printf("\n  " + bold + "Removed:" + reset + "\n")
		for _, s := range removed {
			fmt.Printf("    %s\n", s)
		}
	}
}

// classifySections groups changes by ## headings into modified, added, and removed.
func classifySections(oldLines, newLines []string, hunks []diffHunk) (modified, added, removed []string) {
	oldHeadings := headingsFromLines(oldLines)
	newHeadings := headingsFromLines(newLines)

	oldSet := make(map[string]bool, len(oldHeadings))
	for _, h := range oldHeadings {
		oldSet[h] = true
	}
	newSet := make(map[string]bool, len(newHeadings))
	for _, h := range newHeadings {
		newSet[h] = true
	}

	// Added: in new but not old (preserve document order)
	for _, h := range newHeadings {
		if !oldSet[h] {
			added = append(added, h)
		}
	}

	// Removed: in old but not new (preserve document order)
	for _, h := range oldHeadings {
		if !newSet[h] {
			removed = append(removed, h)
		}
	}

	// Modified: shared headings with hunks falling within them
	modSet := make(map[string]bool)

	for _, h := range hunks {
		if h.oldCount > 0 {
			sec := findSectionAt(oldLines, h.oldStart)
			if sec != "" && oldSet[sec] && newSet[sec] {
				modSet[sec] = true
			}
		}
		if h.newCount > 0 {
			sec := findSectionAt(newLines, h.newStart)
			if sec != "" && oldSet[sec] && newSet[sec] {
				modSet[sec] = true
			}
		}
	}

	// Preserve document order for modified sections
	for _, h := range newHeadings {
		if modSet[h] {
			modified = append(modified, h)
		}
	}

	return
}

// headingsFromLines returns all ## headings from pre-split lines in document order.
func headingsFromLines(lines []string) []string {
	var headings []string
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			headings = append(headings, line)
		}
	}
	return headings
}

// findSectionAt returns the nearest ## heading at or above lineNum (1-based).
func findSectionAt(lines []string, lineNum int) string {
	if lineNum < 1 {
		return ""
	}
	if lineNum > len(lines) {
		lineNum = len(lines)
	}
	for i := lineNum - 1; i >= 0; i-- {
		if strings.HasPrefix(lines[i], "## ") {
			return lines[i]
		}
	}
	return ""
}

// printContextLines prints lines from newLines in dim style within the given range (1-based, inclusive).
func printContextLines(newLines []string, from, to int) {
	for ln := from; ln <= to; ln++ {
		if ln >= 1 && ln <= len(newLines) {
			fmt.Printf(dim+"    %s"+reset+"\n", newLines[ln-1])
		}
	}
}

// showColoredDiff renders a condensed, colored diff with context and section breadcrumbs.
func showColoredDiff(hunks []diffHunk, newLines []string) {
	const ctx = 2
	printed := 0 // last new-content line printed (1-based)
	prevSec := ""

	fmt.Println()

	for i, h := range hunks {
		// Position in new content
		var beforeEnd, afterStart int
		if h.newCount == 0 {
			// Pure removal: gap between newStart and newStart+1
			beforeEnd = h.newStart
			afterStart = h.newStart + 1
		} else {
			beforeEnd = h.newStart - 1
			afterStart = h.newStart + h.newCount
		}

		// Context before: up to ctx lines ending at beforeEnd
		ctxStart := beforeEnd - ctx + 1
		if ctxStart < 1 {
			ctxStart = 1
		}
		if ctxStart <= printed {
			ctxStart = printed + 1
		}

		// Gap indicator between hunks
		if printed > 0 && ctxStart > printed+1 {
			gap := ctxStart - printed - 1
			fmt.Printf(dim+"  ··· %d unchanged lines ···"+reset+"\n", gap)
		}

		// Section breadcrumb
		refLine := h.newStart
		if refLine < 1 {
			refLine = 1
		}
		sec := findSectionAt(newLines, refLine)
		if sec != prevSec && sec != "" {
			fmt.Printf("  "+bold+"%s"+reset+"\n", sec)
			prevSec = sec
		}

		// Print context before (dim)
		printContextLines(newLines, ctxStart, beforeEnd)
		if beforeEnd > printed {
			printed = beforeEnd
		}

		// Removed lines (red)
		for _, line := range h.removed {
			fmt.Printf(red+"  − %s"+reset+"\n", line)
		}

		// Added lines (green)
		for _, line := range h.added {
			fmt.Printf(green+"  + %s"+reset+"\n", line)
		}

		// Context after: up to ctx lines starting at afterStart
		ctxEnd := afterStart + ctx - 1
		if ctxEnd > len(newLines) {
			ctxEnd = len(newLines)
		}
		// Don't overlap with next hunk
		if i+1 < len(hunks) {
			next := hunks[i+1]
			var nextBefore int
			if next.newCount == 0 {
				nextBefore = next.newStart
			} else {
				nextBefore = next.newStart - 1
			}
			if ctxEnd > nextBefore {
				ctxEnd = nextBefore
			}
		}

		printContextLines(newLines, afterStart, ctxEnd)
		if ctxEnd > printed {
			printed = ctxEnd
		}
	}

	fmt.Println()
}
