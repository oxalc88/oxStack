package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bedrockcosts/internal"
)

func main() {
	flags := internal.ParseFetchFlags()

	if flags.Profile == "" {
		fmt.Fprintln(os.Stderr, "ERROR: --profile is required (or set $AWS_PROFILE)")
		os.Exit(1)
	}

	if err := os.MkdirAll(flags.OutputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot create output dir: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Profile  : %s\n", flags.Profile)
	fmt.Printf("Period   : %s → %s\n", flags.Start, flags.End)
	fmt.Printf("Regions  : %s\n", strings.Join(flags.Regions, ", "))
	fmt.Printf("Output   : %s/\n", flags.OutputDir)
	fmt.Println()

	ctx := context.Background()
	cfg, err := internal.LoadAWSConfig(ctx, flags.Profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	save := func(filename string, data interface{}) {
		path := filepath.Join(flags.OutputDir, filename)
		var raw []byte
		switch v := data.(type) {
		case []byte:
			raw = v
		default:
			var err error
			raw, err = json.MarshalIndent(data, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ERROR marshaling %s: %v\n", filename, err)
				return
			}
		}
		if err := os.WriteFile(path, raw, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR writing %s: %v\n", filename, err)
			return
		}
		fmt.Printf("  ✓ %s\n", filename)
	}

	// ── 1. CloudTrail ────────────────────────────────────────────────────────
	fmt.Println("[1/6] Fetching CloudTrail events (all regions)...")
	startDT, err := time.Parse("2006-01-02", flags.Start)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: invalid start date: %v\n", err)
		os.Exit(1)
	}
	endDT, err := time.Parse("2006-01-02", flags.End)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: invalid end date: %v\n", err)
		os.Exit(1)
	}
	startDT = startDT.UTC()
	endDT = endDT.UTC()

	ctEvents, err := internal.FetchCloudTrail(ctx, cfg, flags.Regions, startDT, endDT)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  WARNING: CloudTrail error: %v\n", err)
	}
	fmt.Printf("  Total events: %d\n", len(ctEvents))
	save("cloudtrail_events.json", ctEvents)

	// ── 2-6. CE fetches (concurrent) ─────────────────────────────────────────
	fmt.Println("\n[2-6] Fetching Cost Explorer data concurrently...")
	type ceResult struct {
		filename string
		data     []byte
		err      error
		optional bool
	}
	ceTasks := []struct {
		filename string
		optional bool
		fetch    func() ([]byte, error)
	}{
		{"daily_by_service.json", false, func() ([]byte, error) {
			return internal.FetchDailyByService(ctx, cfg, flags.Start, flags.End)
		}},
		{"bedrock_daily_by_usage_type.json", false, func() ([]byte, error) {
			return internal.FetchBedrockByUsageType(ctx, cfg, flags.Start, flags.End, "DAILY")
		}},
		{"bedrock_daily_by_operation.json", false, func() ([]byte, error) {
			return internal.FetchBedrockByOperation(ctx, cfg, flags.Start, flags.End)
		}},
		{"bedrock_monthly_by_usage_type.json", false, func() ([]byte, error) {
			return internal.FetchBedrockByUsageType(ctx, cfg, flags.Start, flags.End, "MONTHLY")
		}},
		{"costs_by_project_tag.json", true, func() ([]byte, error) {
			return internal.FetchCostsByTag(ctx, cfg, flags.Start, flags.End, flags.TagKey)
		}},
	}
	ceResults := make([]ceResult, len(ceTasks))
	var ceWg sync.WaitGroup
	for i, t := range ceTasks {
		ceWg.Add(1)
		go func(i int, t struct {
			filename string
			optional bool
			fetch    func() ([]byte, error)
		}) {
			defer ceWg.Done()
			data, err := t.fetch()
			ceResults[i] = ceResult{t.filename, data, err, t.optional}
		}(i, t)
	}
	ceWg.Wait()

	for _, r := range ceResults {
		if r.err != nil {
			if r.optional {
				fmt.Fprintf(os.Stderr, "  WARNING: %s failed (%v). Saving empty file.\n", r.filename, r.err)
				save(r.filename, []byte("{}"))
			} else {
				fmt.Fprintf(os.Stderr, "  ERROR: %s: %v\n", r.filename, r.err)
				os.Exit(1)
			}
		} else {
			save(r.filename, r.data)
		}
	}

	// ── Metadata ─────────────────────────────────────────────────────────────
	meta := internal.Metadata{
		Profile:              flags.Profile,
		Start:                flags.Start,
		End:                  flags.End,
		Regions:              flags.Regions,
		TagKey:               flags.TagKey,
		FetchedAt:            time.Now().UTC().Format(time.RFC3339),
		CloudTrailEventCount: len(ctEvents),
	}
	save("metadata.json", meta)

	fmt.Printf("\nDone. Data saved to: %s/\n", flags.OutputDir)
}
