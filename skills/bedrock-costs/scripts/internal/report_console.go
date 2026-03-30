package internal

import (
	"fmt"
	"sort"

	"github.com/tidwall/gjson"
)

// PrintReport prints the analysis as formatted Spanish-language console tables.
func PrintReport(a *Analysis) {
	sep := "────────────────────────────────────────────────────────────"

	fmt.Printf("\n%s\n", "============================================================")
	fmt.Println("  ANÁLISIS DE COSTOS AWS")
	fmt.Printf("%s\n", "============================================================")

	// Cost by service
	fmt.Printf("\n%s\n", sep)
	fmt.Println("  COSTO TOTAL POR SERVICIO (Cost Explorer)")
	fmt.Println(sep)
	svcs := sortedKeys(a.TotalsSvc)
	for _, svc := range svcs {
		fmt.Printf("  $%10.4f  %s\n", a.TotalsSvc[svc], svc)
	}
	fmt.Printf("  %s\n", "──────────────────────────────────────────────────")
	fmt.Printf("  $%10.4f  TOTAL\n", a.GrandTotal)

	// Tool calls
	if a.TotalCalls > 0 {
		fmt.Printf("\n%s\n", sep)
		fmt.Println("  LLAMADAS POR HERRAMIENTA (CloudTrail)")
		fmt.Println(sep)
		tools := sortedKeysByInt(a.ToolCalls)
		for _, tool := range tools {
			cnt := a.ToolCalls[tool]
			pct := float64(cnt) / float64(a.TotalCalls) * 100
			fmt.Printf("  %6d  (%5.1f%%)  %s\n", cnt, pct, tool)
		}
		fmt.Printf("  %s\n", "──────────────────────────────────────────────────")
		fmt.Printf("  %6d  TOTAL\n", a.TotalCalls)
	}

	// Model calls
	if a.TotalCalls > 0 {
		fmt.Printf("\n%s\n", sep)
		fmt.Println("  LLAMADAS POR MODELO (CloudTrail)")
		fmt.Println(sep)
		models := sortedKeysByInt(a.ModelCalls)
		for _, m := range models {
			cnt := a.ModelCalls[m]
			pct := float64(cnt) / float64(a.TotalCalls) * 100
			fmt.Printf("  %6d  (%5.1f%%)  %s\n", cnt, pct, m)
		}
	}

	// Users
	if len(a.UserData) > 0 {
		fmt.Printf("\n%s\n", sep)
		fmt.Println("  DESGLOSE POR USUARIO")
		fmt.Println(sep)
		users := sortedUsersByCallsDesc(a.UserData)
		for _, user := range users {
			data := a.UserData[user]
			fmt.Printf("\n  %s: %d llamadas\n", user, data.Calls)
			fmt.Println("    Herramientas:")
			toolsSorted := sortedKeysByInt(data.Tools)
			for _, t := range toolsSorted {
				fmt.Printf("      %5d  %s\n", data.Tools[t], t)
			}
			fmt.Println("    Modelos:")
			modelsSorted := sortedKeysByInt(data.Models)
			for _, m := range modelsSorted {
				fmt.Printf("      %5d  %s\n", data.Models[m], m)
			}
		}
	}

	// Daily totals
	if len(a.DailyTotals) > 0 {
		fmt.Printf("\n%s\n", sep)
		fmt.Println("  COSTO DIARIO")
		fmt.Println(sep)
		fmt.Printf("  %-12s %12s %8s %10s\n", "Fecha", "AI", "Infra", "Total")
		fmt.Printf("  %s\n", "────────────────────────────────────────────")
		for _, d := range a.DailyTotals {
			spike := ""
			if d.IsSpike {
				spike = " ← SPIKE"
			}
			fmt.Printf("  %-12s $%11.2f $%7.4f $%9.2f%s\n", d.Date, d.AI, d.Infra, d.Total, spike)
		}
		fmt.Printf("  %s\n", "────────────────────────────────────────────")
		fmt.Printf("  %-12s $%11.2f $%7.4f $%9.2f\n", "TOTAL", a.GrandAI, a.GrandInfra, a.GrandTotal)
	}

	// Cache
	fmt.Printf("\n%s\n", sep)
	fmt.Println("  CACHE vs NO-CACHE (Bedrock usage types)")
	fmt.Println(sep)
	fmt.Printf("  Cache read:  $%.4f\n", a.CacheReadCost)
	fmt.Printf("  Cache write: $%.4f\n", a.CacheWriteCost)
	fmt.Printf("  Sin cache:   $%.4f\n", a.NoCacheCost)

	// Tag costs
	if len(a.TagCosts) > 0 && string(a.TagCosts) != "null" {
		fmt.Printf("\n%s\n", sep)
		fmt.Println("  COSTOS POR PROYECTO (tag)")
		fmt.Println(sep)
		monthly := gjson.GetBytes(a.TagCosts, "monthly_by_project")
		grand := 0.0
		monthly.ForEach(func(_, v gjson.Result) bool {
			grand += v.Float()
			return true
		})
		type kv struct {
			k string
			v float64
		}
		var pairs []kv
		monthly.ForEach(func(k, v gjson.Result) bool {
			pairs = append(pairs, kv{k.String(), v.Float()})
			return true
		})
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].v > pairs[j].v })
		for _, p := range pairs {
			pct := 0.0
			if grand > 0 {
				pct = p.v / grand * 100
			}
			fmt.Printf("  %-25s $%10.2f  %.1f%%\n", p.k, p.v, pct)
		}
	}
}

// sortedKeys returns keys of a float64 map sorted by value descending.
func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return m[keys[i]] > m[keys[j]] })
	return keys
}

// sortedKeysByInt returns keys of an int map sorted by value descending.
func sortedKeysByInt(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return m[keys[i]] > m[keys[j]] })
	return keys
}

// sortedUsersByCallsDesc returns user keys sorted by call count descending.
func sortedUsersByCallsDesc(m map[string]*UserInfo) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return m[keys[i]].Calls > m[keys[j]].Calls })
	return keys
}
