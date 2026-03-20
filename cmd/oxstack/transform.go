package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// Ordered replacement rules: $B commands → agent-browser equivalents.
// Order matters — more specific patterns must come before general ones.
var browseReplacements = []struct{ old, new string }{
	{"$B goto ", "agent-browser open "},
	{"$B snapshot -i -a -o", "agent-browser screenshot --annotate"},
	{"$B snapshot -D", "agent-browser diff snapshot"},
	{"$B snapshot -i -c", "agent-browser snapshot -i -c"},
	{"$B snapshot -i", "agent-browser snapshot -i"},
	{"$B snapshot -C", "agent-browser snapshot -C"},
	{"$B snapshot", "agent-browser snapshot"},
	{"$B fill @e", "agent-browser fill e"},
	{"$B click @e", "agent-browser click e"},
	{"$B hover @e", "agent-browser hover e"},
	{"$B select @e", "agent-browser select e"},
	{"$B console --errors", "agent-browser console error"},
	{"$B console", "agent-browser console"},
	{"$B js ", "agent-browser eval "},
	{"$B url", "agent-browser get url"},
	{"$B text", "agent-browser get text"},
	{"$B network", "agent-browser network requests"},
	{"$B screenshot ", "agent-browser screenshot "},
	{"$B screenshot", "agent-browser screenshot"},
	{"$B tabs", "agent-browser tab"},
	{"$B newtab ", "agent-browser tab new "},
	{"$B tab ", "agent-browser tab "},
	{"$B press ", "agent-browser press "},
	{"$B type ", "agent-browser type "},
	{"$B reload", "agent-browser reload"},
	{"$B back", "agent-browser back"},
	{"$B forward", "agent-browser forward"},
	{"$B pdf", "agent-browser pdf"},
	{"$B cookie-import", "agent-browser state load"},
	{"$B diff ", "agent-browser diff url "},
	{"`$B ", "`agent-browser "},
}

// Regex replacements
var (
	reViewport = regexp.MustCompile(`\$B viewport (\d+)x(\d+)`)
	reAtRef    = regexp.MustCompile(`@e(\d)`)
	rePerf     = regexp.MustCompile(`\$B perf`)
	reCss      = regexp.MustCompile(`\$B css (\S+) (.+)`)
)

// transformBrowse applies all $B → agent-browser transformations to a single line.
func transformBrowse(line string) string {
	// Apply ordered string replacements
	for _, r := range browseReplacements {
		line = strings.ReplaceAll(line, r.old, r.new)
	}

	// $B viewport WxH → agent-browser set viewport W H
	line = reViewport.ReplaceAllString(line, "agent-browser set viewport $1 $2")

	// $B perf → agent-browser eval "JSON.stringify(performance.getEntriesByType('navigation')[0])"
	line = rePerf.ReplaceAllString(line, `agent-browser eval "JSON.stringify(performance.getEntriesByType('navigation')[0])"`)

	// $B css <sel> <prop> → agent-browser eval "getComputedStyle(document.querySelector('<sel>')).<prop>"
	line = reCss.ReplaceAllStringFunc(line, func(match string) string {
		m := reCss.FindStringSubmatch(match)
		if len(m) == 3 {
			return `agent-browser eval "getComputedStyle(document.querySelector('` + m[1] + `'))` + `.` + m[2] + `"`
		}
		return match
	})

	// @eN → eN (strip @ prefix on element refs)
	line = reAtRef.ReplaceAllString(line, "e$1")

	return line
}

// Boilerplate section headers to strip from gstack SKILL.md.
var stripSections = []string{
	"Preamble (run first)",
	"AskUserQuestion Format",
	"Completeness Principle",
	"Contributor Mode",
	"Completion Status Protocol",
	"Telemetry (run last)",
	"SETUP (run this check BEFORE any browse command)",
	"Test Framework Bootstrap",
}

// extractMethodology reads a gstack SKILL.md and returns the methodology content
// with frontmatter removed and boilerplate sections stripped.
func extractMethodology(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var b strings.Builder
	scanner := bufio.NewScanner(f)

	inFrontmatter := false
	pastFrontmatter := false
	inStrip := false

	for scanner.Scan() {
		line := scanner.Text()

		// Skip frontmatter
		if !pastFrontmatter {
			if line == "---" {
				if inFrontmatter {
					pastFrontmatter = true
				} else {
					inFrontmatter = true
				}
				continue
			}
			continue
		}

		// Skip AUTO-GENERATED comments
		if strings.HasPrefix(line, "<!-- AUTO-GENERATED") {
			continue
		}
		if strings.HasPrefix(line, "<!-- Regenerate:") {
			continue
		}

		// Check if this is a section heading
		if strings.HasPrefix(line, "## ") {
			heading := strings.TrimPrefix(line, "## ")
			if isStripSection(heading) {
				inStrip = true
				continue
			}
			inStrip = false
		}

		if inStrip {
			continue
		}

		b.WriteString(line)
		b.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return b.String(), nil
}

// isStripSection checks if a heading matches any boilerplate section to strip.
func isStripSection(heading string) bool {
	for _, s := range stripSections {
		if strings.HasPrefix(heading, s) {
			return true
		}
	}
	return false
}

// getHeader reads the -byOx SKILL.md and returns everything before the methodology marker.
func getHeader(path, skill string) (string, error) {
	var marker string
	switch skill {
	case "qa":
		marker = "## Step 0: Detect base branch"
	case "design-review":
		marker = "## Setup"
	default:
		return "", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	var b strings.Builder
	for _, line := range lines {
		if line == marker {
			break
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}

	return b.String(), nil
}

// Boilerplate diff filter patterns (for sync command filtering).
var boilerplatePatterns = []string{
	"gstack-update-check",
	"gstack-config",
	"gstack-telemetry",
	"gstack-slug",
	"gstack-analytics",
	".gstack/sessions",
	".gstack/analytics",
	"_CONTRIB",
	"_LAKE_SEEN",
	"_TEL",
	"_SESSION_ID",
	"_PROACTIVE",
	"UPGRADE_AVAILABLE",
	"JUST_UPGRADED",
	"Boil the Lake",
	"Completeness Principle",
	"Contributor Mode",
	"field report",
	"AUTO-GENERATED",
	"Regenerate: bun",
}

// filterBoilerplate removes lines matching boilerplate patterns from diff output.
func filterBoilerplate(diff string) string {
	var filtered strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()
		skip := false
		for _, pat := range boilerplatePatterns {
			if strings.Contains(line, pat) {
				skip = true
				break
			}
		}
		if !skip {
			filtered.WriteString(line)
			filtered.WriteByte('\n')
		}
	}
	return filtered.String()
}
