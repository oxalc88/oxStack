package internal

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

const DefaultRegions = "us-east-1,us-west-2,eu-west-1,ap-southeast-1,us-east-2"

// FetchFlags holds parsed CLI flags for the fetch command.
type FetchFlags struct {
	Profile   string
	Start     string
	End       string
	OutputDir string
	Regions   []string
	TagKey    string
}

// AnalyzeFlags holds parsed CLI flags for the analyze command.
type AnalyzeFlags struct {
	DataDir string
	Format  string
	Output  string
}

func ParseFetchFlags() *FetchFlags {
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	profile := flag.String("profile", os.Getenv("AWS_PROFILE"), "AWS profile name (default: $AWS_PROFILE)")
	start := flag.String("start", firstOfMonth.Format("2006-01-02"), "Start date YYYY-MM-DD (default: first day of current month)")
	end := flag.String("end", now.Format("2006-01-02"), "End date YYYY-MM-DD (default: today)")
	outputDir := flag.String("output-dir", "./bedrock-costs", "Output directory (default: ./bedrock-costs)")
	regions := flag.String("regions", DefaultRegions, "Comma-separated regions for CloudTrail")
	tagKey := flag.String("tag-key", "project", "Cost-allocation tag key for project breakdown (default: project)")
	flag.Parse()

	regionList := []string{}
	for _, r := range strings.Split(*regions, ",") {
		r = strings.TrimSpace(r)
		if r != "" {
			regionList = append(regionList, r)
		}
	}

	return &FetchFlags{
		Profile:   *profile,
		Start:     *start,
		End:       *end,
		OutputDir: *outputDir,
		Regions:   regionList,
		TagKey:    *tagKey,
	}
}

func ParseAnalyzeFlags() *AnalyzeFlags {
	dataDir := flag.String("data-dir", "./bedrock-costs", "Directory with JSON data files from fetch (default: ./bedrock-costs)")
	format := flag.String("format", "both", "Output format: console, markdown, or both (default: both)")
	output := flag.String("output", "report.md", "Markdown output file name (default: report.md)")
	flag.Parse()

	return &AnalyzeFlags{
		DataDir: *dataDir,
		Format:  *format,
		Output:  *output,
	}
}

// LoadAWSConfig loads an AWS config for the given profile.
// If profile is empty, uses the default credential chain.
func LoadAWSConfig(ctx context.Context, profile string) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("load AWS config: %w", err)
	}
	return cfg, nil
}

// NormalizeModelID strips region prefix and version suffix from Bedrock model IDs.
// Mirrors Python: normalize_model_id()
func NormalizeModelID(model string) string {
	m := model
	m = strings.ReplaceAll(m, "global.anthropic.", "")
	m = strings.ReplaceAll(m, "us.anthropic.", "")
	for _, suf := range []string{
		"-20251001-v1:0",
		"-20250929-v1:0",
		"-20241022-v1:0",
		"-20251101-v1:0",
		"-v1",
		"-v2:0",
	} {
		m = strings.TrimSuffix(m, suf)
	}
	return m
}

// ClassifyUserAgent maps a raw userAgent string to a human-readable tool name.
// Mirrors Python: classify_user_agent()
func ClassifyUserAgent(ua string) string {
	l := strings.ToLower(ua)
	if strings.Contains(l, "claude-cli") && strings.Contains(l, "vscode") {
		return "Claude Code (VSCode)"
	}
	if strings.Contains(l, "claude-cli") {
		return "Claude Code (CLI)"
	}
	if strings.Contains(l, "kilocode") {
		return "KiloCode"
	}
	if strings.Contains(l, "opencode") {
		return "OpenCode"
	}
	if strings.Contains(l, "cline") {
		return "Cline"
	}
	if strings.Contains(l, "strands") {
		return "Strands SDK"
	}
	if strings.Contains(l, "ai-sdk") && !strings.Contains(l, "lambda") {
		return "ai-sdk (Lambda)"
	}
	if strings.Contains(l, "boto3") {
		return "boto3/python"
	}
	if strings.Contains(l, "aws-sdk-js") {
		return "aws-sdk-js"
	}
	if strings.Contains(l, "mozilla") {
		return "Bedrock Console (browser)"
	}
	if strings.Contains(l, "aws-cli") {
		return "AWS CLI"
	}
	if ua == "" {
		return "(sin userAgent)"
	}
	if len(ua) > 50 {
		return ua[:50]
	}
	return ua
}
