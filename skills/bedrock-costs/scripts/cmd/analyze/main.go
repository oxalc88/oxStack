package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"bedrockcosts/internal"
)

func loadJSON(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}

func main() {
	flags := internal.ParseAnalyzeFlags()
	d := flags.DataDir

	// Load raw JSON files
	ctRaw := loadJSON(filepath.Join(d, "cloudtrail_events.json"))
	dbsRaw := loadJSON(filepath.Join(d, "daily_by_service.json"))
	bduRaw := loadJSON(filepath.Join(d, "bedrock_daily_by_usage_type.json"))
	bmuRaw := loadJSON(filepath.Join(d, "bedrock_monthly_by_usage_type.json"))
	bdoRaw := loadJSON(filepath.Join(d, "bedrock_daily_by_operation.json"))
	tagRaw := loadJSON(filepath.Join(d, "costs_by_project_tag.json"))
	metaRaw := loadJSON(filepath.Join(d, "metadata.json"))

	if dbsRaw == nil {
		fmt.Fprintf(os.Stderr, "ERROR: No data found in %s/. Run fetch first.\n", d)
		os.Exit(1)
	}

	// Parse CloudTrail events
	var ctEvents []internal.CTEvent
	if ctRaw != nil {
		if err := json.Unmarshal(ctRaw, &ctEvents); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: could not parse cloudtrail_events.json: %v\n", err)
		}
	}

	// Parse metadata
	var meta *internal.Metadata
	if metaRaw != nil {
		var m internal.Metadata
		if err := json.Unmarshal(metaRaw, &m); err == nil {
			meta = &m
		}
	}

	fmt.Printf("CloudTrail events: %d\n", len(ctEvents))
	if meta != nil {
		fmt.Printf("Profile: %s\n", meta.Profile)
		fmt.Printf("Period:  %s → %s\n", meta.Start, meta.End)
	}
	fmt.Println()

	ce := internal.CEData{
		DailyByService:          orEmpty(dbsRaw),
		BedrockDailyByUsage:     orEmpty(bduRaw),
		BedrockMonthlyByUsage:   orEmpty(bmuRaw),
		BedrockDailyByOperation: orEmpty(bdoRaw),
	}

	analysis := internal.Analyze(ctEvents, ce, json.RawMessage(tagRaw))

	if flags.Format == "console" || flags.Format == "both" {
		internal.PrintReport(analysis)
	}

	if flags.Format == "markdown" || flags.Format == "both" {
		if err := internal.WriteMarkdownReport(analysis, flags.Output, meta); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR writing markdown report: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Reporte markdown guardado: %s\n", flags.Output)
	}
}

func orEmpty(b []byte) []byte {
	if b == nil {
		return []byte("{}")
	}
	return b
}
