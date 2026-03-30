package internal

import "testing"

func TestNormalizeModelID(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"global.anthropic.claude-3-5-sonnet-20241022-v1:0", "claude-3-5-sonnet"},
		{"us.anthropic.claude-opus-4-6-20250929-v1:0", "claude-opus-4-6"},
		{"inference-profile/us.anthropic.claude-3-haiku-20241022-v1:0", "inference-profile/claude-3-haiku"},
		{"claude-3-5-sonnet-20241022-v1:0", "claude-3-5-sonnet"},
		{"claude-sonnet-4-6-20251001-v1:0", "claude-sonnet-4-6"},
		{"N/A", "N/A"},
		{"", ""},
	}
	for _, c := range cases {
		got := NormalizeModelID(c.in)
		if got != c.want {
			t.Errorf("NormalizeModelID(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestClassifyUserAgent(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"claude-cli/1.0 vscode/1.85", "Claude Code (VSCode)"},
		{"claude-cli/1.0", "Claude Code (CLI)"},
		{"KiloCode/0.9", "KiloCode"},
		{"opencode/0.1", "OpenCode"},
		{"cline/2.0", "Cline"},
		{"strands-agent/1.0", "Strands SDK"},
		{"boto3/1.34", "boto3/python"},
		{"aws-sdk-js/2.0", "aws-sdk-js"},
		{"Mozilla/5.0", "Bedrock Console (browser)"},
		{"aws-cli/2.0", "AWS CLI"},
		{"", "(sin userAgent)"},
	}
	for _, c := range cases {
		got := ClassifyUserAgent(c.in)
		if got != c.want {
			t.Errorf("ClassifyUserAgent(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
