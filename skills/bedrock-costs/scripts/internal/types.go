package internal

import "encoding/json"

// CTEvent is a single normalized CloudTrail event.
type CTEvent struct {
	Time      string `json:"time"`
	Event     string `json:"event"`
	Region    string `json:"region"`
	User      string `json:"user"`
	Model     string `json:"model"`
	UserAgent string `json:"userAgent"`
	SourceIP  string `json:"sourceIP"`
}

// UserInfo aggregates call counts for one IAM user.
type UserInfo struct {
	Calls  int            `json:"calls"`
	Models map[string]int `json:"models"`
	Tools  map[string]int `json:"tools"`
}

// DailyTotal is one day of AI vs infra spend.
type DailyTotal struct {
	Date    string  `json:"date"`
	AI      float64 `json:"ai"`
	Infra   float64 `json:"infra"`
	Total   float64 `json:"total"`
	IsSpike bool    `json:"is_spike"`
}

// Metadata mirrors metadata.json written by fetch.
type Metadata struct {
	Profile             string   `json:"profile"`
	Start               string   `json:"start"`
	End                 string   `json:"end"`
	Regions             []string `json:"regions"`
	TagKey              string   `json:"tag_key"`
	FetchedAt           string   `json:"fetched_at"`
	CloudTrailEventCount int     `json:"cloudtrail_event_count"`
}

// Analysis holds all computed results for reporting.
type Analysis struct {
	TotalsSvc           map[string]float64            `json:"totals_svc"`
	DailySvc            map[string]map[string]float64 `json:"daily_svc"`
	AISvcs              []string                      `json:"ai_svcs"`
	InfraSvcs           []string                      `json:"infra_svcs"`
	GrandTotal          float64                       `json:"grand_total"`
	GrandAI             float64                       `json:"grand_ai"`
	GrandInfra          float64                       `json:"grand_infra"`
	BedrockUsageTotals  map[string]float64            `json:"bedrock_usage_totals"`
	BedrockMonthlyTotals map[string]float64           `json:"bedrock_monthly_totals"`
	UserData            map[string]*UserInfo          `json:"user_data"`
	ToolCalls           map[string]int                `json:"tool_calls"`
	ModelCalls          map[string]int                `json:"model_calls"`
	TotalCalls          int                           `json:"total_calls"`
	DailyTotals         []DailyTotal                  `json:"daily_totals"`
	SpikeThreshold      float64                       `json:"spike_threshold"`
	CachedCalls         int                           `json:"cached_calls"`
	UncachedCalls       int                           `json:"uncached_calls"`
	CacheReadCost       float64                       `json:"cache_read_cost"`
	CacheWriteCost      float64                       `json:"cache_write_cost"`
	NoCacheCost         float64                       `json:"no_cache_cost"`
	TagCosts            json.RawMessage               `json:"tag_costs,omitempty"`
}
