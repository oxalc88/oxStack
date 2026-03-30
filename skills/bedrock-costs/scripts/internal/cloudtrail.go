package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
)

var bedrockEventNames = []string{
	"InvokeModel",
	"InvokeModelWithResponseStream",
	"Converse",
	"ConverseStream",
}

// FetchCloudTrail fetches Bedrock CloudTrail events across all regions in parallel.
// Launches len(regions) × len(eventNames) goroutines.
func FetchCloudTrail(ctx context.Context, cfg aws.Config, regions []string, start, end time.Time) ([]CTEvent, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var allEvents []CTEvent

	for _, region := range regions {
		for _, eventName := range bedrockEventNames {
			wg.Add(1)
			go func(region, eventName string) {
				defer wg.Done()
				events, err := fetchForRegionEvent(ctx, cfg, region, eventName, start, end)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					fmt.Printf("  [CloudTrail] %s/%s ERROR: %v\n", region, eventName, err)
					return
				}
				if len(events) > 0 {
					allEvents = append(allEvents, events...)
				}
			}(region, eventName)
		}
	}
	wg.Wait()

	// Print per-region counts
	regionCounts := map[string]int{}
	for _, e := range allEvents {
		regionCounts[e.Region]++
	}
	for _, region := range regions {
		if cnt, ok := regionCounts[region]; ok && cnt > 0 {
			fmt.Printf("  [CloudTrail] %s: %d events\n", region, cnt)
		}
	}

	return allEvents, nil
}

func fetchForRegionEvent(ctx context.Context, cfg aws.Config, region, eventName string, start, end time.Time) ([]CTEvent, error) {
	// Create a region-specific config copy
	regionCfg := cfg.Copy()
	regionCfg.Region = region
	client := cloudtrail.NewFromConfig(regionCfg)

	var events []CTEvent
	input := &cloudtrail.LookupEventsInput{
		LookupAttributes: []types.LookupAttribute{
			{
				AttributeKey:   types.LookupAttributeKeyEventName,
				AttributeValue: aws.String(eventName),
			},
		},
		StartTime:  aws.Time(start),
		EndTime:    aws.Time(end),
		MaxResults: aws.Int32(50),
	}

	for {
		resp, err := client.LookupEvents(ctx, input)
		if err != nil {
			return events, fmt.Errorf("cloudtrail %s/%s: %w", region, eventName, err)
		}

		for _, e := range resp.Events {
			raw := aws.ToString(e.CloudTrailEvent)
			if raw == "" {
				raw = "{}"
			}
			var ct struct {
				RequestParameters map[string]interface{} `json:"requestParameters"`
				UserAgent         string                 `json:"userAgent"`
				SourceIPAddress   string                 `json:"sourceIPAddress"`
			}
			if err := json.Unmarshal([]byte(raw), &ct); err != nil {
				continue
			}
			rp := ct.RequestParameters
			model := ""
			if v, ok := rp["modelId"]; ok {
				model = fmt.Sprintf("%v", v)
			} else if v, ok := rp["modelIdentifier"]; ok {
				model = fmt.Sprintf("%v", v)
			}
			if model == "" {
				model = "N/A"
			}
			if strings.Contains(model, "inference-profile") {
				parts := strings.Split(model, "/")
				model = parts[len(parts)-1]
			}

			timeStr := ""
			if e.EventTime != nil {
				timeStr = e.EventTime.Format(time.RFC3339)
			}
			user := aws.ToString(e.Username)

			events = append(events, CTEvent{
				Time:      timeStr,
				Event:     eventName,
				Region:    region,
				User:      user,
				Model:     model,
				UserAgent: ct.UserAgent,
				SourceIP:  ct.SourceIPAddress,
			})
		}

		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}

	return events, nil
}
