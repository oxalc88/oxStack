package internal

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/tidwall/gjson"
)

var aiServiceKeywords = []string{"Claude", "Amazon Bedrock"}

func isAIService(svc string) bool {
	for _, kw := range aiServiceKeywords {
		if strings.Contains(svc, kw) {
			return true
		}
	}
	return false
}

// CEData holds raw Cost Explorer JSON blobs loaded from disk.
type CEData struct {
	DailyByService          []byte
	BedrockDailyByUsage     []byte
	BedrockMonthlyByUsage   []byte
	BedrockDailyByOperation []byte
}

// Analyze produces an Analysis from CloudTrail events + CE data + tag costs.
func Analyze(events []CTEvent, ce CEData, tagCosts json.RawMessage) *Analysis {
	// ── Cost Explorer: totals by service ─────────────────────────────────────
	totalsSvc := map[string]float64{}
	dailySvc := map[string]map[string]float64{}

	gjson.GetBytes(ce.DailyByService, "ResultsByTime").ForEach(func(_, result gjson.Result) bool {
		day := result.Get("TimePeriod.Start").String()
		result.Get("Groups").ForEach(func(_, g gjson.Result) bool {
			svc := g.Get("Keys.0").String()
			cost := g.Get("Metrics.BlendedCost.Amount").Float()
			if cost > 0.000001 {
				totalsSvc[svc] += cost
				if dailySvc[day] == nil {
					dailySvc[day] = map[string]float64{}
				}
				dailySvc[day][svc] += cost
			}
			return true
		})
		return true
	})

	// Bedrock by usage type (daily + monthly)
	bedrockUsageTotals := sumUsageTotals(ce.BedrockDailyByUsage)
	bedrockMonthlyTotals := sumUsageTotals(ce.BedrockMonthlyByUsage)

	// AI vs infra split
	var aiSvcs, infraSvcs []string
	grandTotal := 0.0
	grandAI := 0.0
	grandInfra := 0.0
	for svc, cost := range totalsSvc {
		grandTotal += cost
		if isAIService(svc) {
			aiSvcs = append(aiSvcs, svc)
			grandAI += cost
		} else {
			infraSvcs = append(infraSvcs, svc)
			grandInfra += cost
		}
	}

	// ── CloudTrail: by user, tool, model ─────────────────────────────────────
	userData := map[string]*UserInfo{}
	toolCalls := map[string]int{}
	modelCalls := map[string]int{}

	for _, e := range events {
		user := e.User
		if idx := strings.Index(user, "@"); idx >= 0 {
			user = user[:idx]
		}
		model := NormalizeModelID(e.Model)
		tool := ClassifyUserAgent(e.UserAgent)

		if userData[user] == nil {
			userData[user] = &UserInfo{
				Models: map[string]int{},
				Tools:  map[string]int{},
			}
		}
		userData[user].Calls++
		userData[user].Models[model]++
		userData[user].Tools[tool]++
		toolCalls[tool]++
		modelCalls[model]++
	}

	totalCalls := 0
	for _, c := range toolCalls {
		totalCalls += c
	}

	// ── Daily AI vs infra ────────────────────────────────────────────────────
	dayKeys := make([]string, 0, len(dailySvc))
	for k := range dailySvc {
		dayKeys = append(dayKeys, k)
	}
	slices.Sort(dayKeys)

	var dailyTotals []DailyTotal
	for _, day := range dayKeys {
		svcs := dailySvc[day]
		ai := 0.0
		infra := 0.0
		for svc, cost := range svcs {
			if isAIService(svc) {
				ai += cost
			} else {
				infra += cost
			}
		}
		if ai+infra > 0.001 {
			dailyTotals = append(dailyTotals, DailyTotal{
				Date:  day,
				AI:    ai,
				Infra: infra,
				Total: ai + infra,
			})
		}
	}

	// Spike detection: days >= max(median*2, 50.0)
	spikeThreshold := 0.0
	if len(dailyTotals) > 0 {
		totals := make([]float64, len(dailyTotals))
		for i, d := range dailyTotals {
			totals[i] = d.Total
		}
		slices.Sort(totals)
		mid := len(totals) / 2
		median := totals[mid]
		spikeThreshold = median * 2
		if spikeThreshold < 50.0 {
			spikeThreshold = 50.0
		}
		for i := range dailyTotals {
			dailyTotals[i].IsSpike = dailyTotals[i].Total >= spikeThreshold
		}
	}

	// ── Caching analysis ─────────────────────────────────────────────────────
	cachedCalls := 0
	uncachedCalls := 0
	for _, e := range events {
		m := e.Model
		if strings.HasPrefix(m, "global.") {
			cachedCalls++
		} else {
			uncachedCalls++
		}
	}

	cacheReadCost := 0.0
	cacheWriteCost := 0.0
	noCacheCost := 0.0
	for k, v := range bedrockUsageTotals {
		kl := strings.ToLower(k)
		if strings.Contains(kl, "cache-read") || strings.Contains(kl, "cacheread") {
			cacheReadCost += v
		} else if strings.Contains(kl, "cache-write") || strings.Contains(kl, "cachewrite") {
			cacheWriteCost += v
		} else if !strings.Contains(kl, "cache") {
			noCacheCost += v
		}
	}

	return &Analysis{
		TotalsSvc:            totalsSvc,
		DailySvc:             dailySvc,
		AISvcs:               aiSvcs,
		InfraSvcs:            infraSvcs,
		GrandTotal:           grandTotal,
		GrandAI:              grandAI,
		GrandInfra:           grandInfra,
		BedrockUsageTotals:   bedrockUsageTotals,
		BedrockMonthlyTotals: bedrockMonthlyTotals,
		UserData:             userData,
		ToolCalls:            toolCalls,
		ModelCalls:           modelCalls,
		TotalCalls:           totalCalls,
		DailyTotals:          dailyTotals,
		SpikeThreshold:       spikeThreshold,
		CachedCalls:          cachedCalls,
		UncachedCalls:        uncachedCalls,
		CacheReadCost:        cacheReadCost,
		CacheWriteCost:       cacheWriteCost,
		NoCacheCost:          noCacheCost,
		TagCosts:             tagCosts,
	}
}

func sumUsageTotals(data []byte) map[string]float64 {
	totals := map[string]float64{}
	gjson.GetBytes(data, "ResultsByTime").ForEach(func(_, result gjson.Result) bool {
		result.Get("Groups").ForEach(func(_, g gjson.Result) bool {
			usageType := g.Get("Keys.0").String()
			cost := g.Get("Metrics.BlendedCost.Amount").Float()
			if cost > 0 {
				totals[usageType] += cost
			}
			return true
		})
		return true
	})
	return totals
}
