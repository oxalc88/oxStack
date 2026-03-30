package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/tidwall/gjson"
)

var bedrockFilter = &cetypes.Expression{
	Dimensions: &cetypes.DimensionValues{
		Key:    cetypes.DimensionService,
		Values: []string{"Amazon Bedrock"},
	},
}

func ceClient(cfg aws.Config) *costexplorer.Client {
	// Cost Explorer is always us-east-1
	c := cfg.Copy()
	c.Region = "us-east-1"
	return costexplorer.NewFromConfig(c)
}

// FetchDailyByService fetches daily cost grouped by service for all AWS services.
func FetchDailyByService(ctx context.Context, cfg aws.Config, start, end string) ([]byte, error) {
	ce := ceClient(cfg)
	resp, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod:  &cetypes.DateInterval{Start: aws.String(start), End: aws.String(end)},
		Granularity: cetypes.GranularityDaily,
		Metrics:     []string{"BlendedCost"},
		GroupBy:     []cetypes.GroupDefinition{{Type: cetypes.GroupDefinitionTypeDimension, Key: aws.String("SERVICE")}},
	})
	if err != nil {
		return nil, fmt.Errorf("FetchDailyByService: %w", err)
	}
	return json.Marshal(resp)
}

// FetchBedrockByUsageType fetches Bedrock cost grouped by USAGE_TYPE at given granularity.
func FetchBedrockByUsageType(ctx context.Context, cfg aws.Config, start, end, granularity string) ([]byte, error) {
	ce := ceClient(cfg)
	gran := cetypes.GranularityDaily
	if strings.ToUpper(granularity) == "MONTHLY" {
		gran = cetypes.GranularityMonthly
	}
	resp, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod:  &cetypes.DateInterval{Start: aws.String(start), End: aws.String(end)},
		Granularity: gran,
		Metrics:     []string{"BlendedCost", "UsageQuantity"},
		Filter:      bedrockFilter,
		GroupBy:     []cetypes.GroupDefinition{{Type: cetypes.GroupDefinitionTypeDimension, Key: aws.String("USAGE_TYPE")}},
	})
	if err != nil {
		return nil, fmt.Errorf("FetchBedrockByUsageType(%s): %w", granularity, err)
	}
	return json.Marshal(resp)
}

// FetchBedrockByOperation fetches Bedrock cost grouped by OPERATION, daily.
func FetchBedrockByOperation(ctx context.Context, cfg aws.Config, start, end string) ([]byte, error) {
	ce := ceClient(cfg)
	resp, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod:  &cetypes.DateInterval{Start: aws.String(start), End: aws.String(end)},
		Granularity: cetypes.GranularityDaily,
		Metrics:     []string{"BlendedCost", "UsageQuantity"},
		Filter:      bedrockFilter,
		GroupBy:     []cetypes.GroupDefinition{{Type: cetypes.GroupDefinitionTypeDimension, Key: aws.String("OPERATION")}},
	})
	if err != nil {
		return nil, fmt.Errorf("FetchBedrockByOperation: %w", err)
	}
	return json.Marshal(resp)
}

// FetchCostsByTag fetches cost breakdown by a cost-allocation tag.
// Returns structured JSON matching the Python format:
//
//	{monthly_by_project, monthly_periods, daily_by_project, service_by_project}
func FetchCostsByTag(ctx context.Context, cfg aws.Config, start, end, tagKey string) ([]byte, error) {
	ce := ceClient(cfg)

	getCost := func(gran cetypes.Granularity, extraGroups []cetypes.GroupDefinition) ([]byte, error) {
		groupBy := append(append([]cetypes.GroupDefinition(nil), extraGroups...), cetypes.GroupDefinition{
			Type: cetypes.GroupDefinitionTypeTag,
			Key:  aws.String(tagKey),
		})
		resp, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
			TimePeriod:  &cetypes.DateInterval{Start: aws.String(start), End: aws.String(end)},
			Granularity: gran,
			Metrics:     []string{"BlendedCost"},
			GroupBy:     groupBy,
		})
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}

	projectFromKey := func(key string) string {
		if idx := strings.Index(key, "$"); idx >= 0 {
			v := key[idx+1:]
			if v == "" {
				return "(sin tag)"
			}
			return v
		}
		if key == "" {
			return "(sin tag)"
		}
		return key
	}

	// Monthly by tag
	monthlyRaw, err := getCost(cetypes.GranularityMonthly, nil)
	if err != nil {
		return nil, fmt.Errorf("tag monthly: %w", err)
	}

	monthlyByProject := map[string]float64{}
	var monthlyPeriods []map[string]interface{}
	gjson.GetBytes(monthlyRaw, "ResultsByTime").ForEach(func(_, period gjson.Result) bool {
		periodTotals := map[string]float64{}
		pStart := period.Get("TimePeriod.Start").String()
		pEnd := period.Get("TimePeriod.End").String()
		period.Get("Groups").ForEach(func(_, g gjson.Result) bool {
			proj := projectFromKey(g.Get("Keys.0").String())
			cost := g.Get("Metrics.BlendedCost.Amount").Float()
			if cost > 0.0001 {
				periodTotals[proj] += cost
				monthlyByProject[proj] += cost
			}
			return true
		})
		monthlyPeriods = append(monthlyPeriods, map[string]interface{}{
			"start":      pStart,
			"end":        pEnd,
			"by_project": periodTotals,
		})
		return true
	})

	// Daily by tag
	dailyRaw, err := getCost(cetypes.GranularityDaily, nil)
	if err != nil {
		return nil, fmt.Errorf("tag daily: %w", err)
	}

	var dailyByProject []map[string]interface{}
	gjson.GetBytes(dailyRaw, "ResultsByTime").ForEach(func(_, period gjson.Result) bool {
		dayTotals := map[string]float64{}
		dateStr := period.Get("TimePeriod.Start").String()
		period.Get("Groups").ForEach(func(_, g gjson.Result) bool {
			proj := projectFromKey(g.Get("Keys.0").String())
			cost := g.Get("Metrics.BlendedCost.Amount").Float()
			if cost > 0.0001 {
				dayTotals[proj] += cost
			}
			return true
		})
		if len(dayTotals) > 0 {
			dailyByProject = append(dailyByProject, map[string]interface{}{
				"date":       dateStr,
				"by_project": dayTotals,
			})
		}
		return true
	})

	// Service × tag (monthly)
	svcTagRaw, err := getCost(cetypes.GranularityMonthly, []cetypes.GroupDefinition{
		{Type: cetypes.GroupDefinitionTypeDimension, Key: aws.String("SERVICE")},
	})
	if err != nil {
		return nil, fmt.Errorf("tag service×tag: %w", err)
	}

	serviceByProject := map[string]map[string]float64{}
	gjson.GetBytes(svcTagRaw, "ResultsByTime").ForEach(func(_, period gjson.Result) bool {
		period.Get("Groups").ForEach(func(_, g gjson.Result) bool {
			svc := g.Get("Keys.0").String()
			proj := projectFromKey(g.Get("Keys.1").String())
			cost := g.Get("Metrics.BlendedCost.Amount").Float()
			if cost > 0.0001 {
				if serviceByProject[svc] == nil {
					serviceByProject[svc] = map[string]float64{}
				}
				serviceByProject[svc][proj] += cost
			}
			return true
		})
		return true
	})

	result := map[string]interface{}{
		"monthly_by_project": monthlyByProject,
		"monthly_periods":    monthlyPeriods,
		"daily_by_project":   dailyByProject,
		"service_by_project": serviceByProject,
	}
	return json.Marshal(result)
}
