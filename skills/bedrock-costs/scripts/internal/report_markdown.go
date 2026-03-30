package internal

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// WriteMarkdownReport writes a markdown report matching the analisis_costos_aws_v2.md format.
func WriteMarkdownReport(a *Analysis, path string, meta *Metadata) error {
	today := time.Now().Format("2006-01-02")
	profile := "unknown"
	start := "?"
	end := today
	var regions []string

	if meta != nil {
		if meta.Profile != "" {
			profile = meta.Profile
		}
		if meta.Start != "" {
			start = meta.Start
		}
		if meta.End != "" {
			end = meta.End
		}
		regions = meta.Regions
	}

	var b strings.Builder

	w := func(format string, args ...interface{}) {
		fmt.Fprintf(&b, format+"\n", args...)
	}
	line := func(s string) { b.WriteString(s + "\n") }

	// Header
	w("# Análisis de Costos AWS — %s", profile)
	w("**Período**: %s – %s", start, end)
	w("**Generado**: %s", today)
	sources := "AWS Cost Explorer (todos los servicios)"
	if a.TotalCalls > 0 {
		sources += " + CloudTrail completo (`InvokeModel`, `InvokeModelWithResponseStream`, `Converse`, `ConverseStream`)"
	}
	if len(regions) > 0 {
		sources += fmt.Sprintf(" en regiones: %s", strings.Join(regions, ", "))
	}
	w("**Fuentes**: %s", sources)
	line("")
	line("---")
	line("")

	// Executive summary
	line("## Resumen Ejecutivo")
	line("")
	line("| Concepto | Valor |")
	line("|----------|-------|")
	w("| **Costo total del período** | **$%.2f** |", a.GrandTotal)
	aiPct := 0.0
	infraPct := 0.0
	if a.GrandTotal > 0 {
		aiPct = a.GrandAI / a.GrandTotal * 100
		infraPct = a.GrandInfra / a.GrandTotal * 100
	}
	w("| Costo en modelos de IA | $%.2f (%.2f%%) |", a.GrandAI, aiPct)
	w("| Costo en infraestructura AWS | $%.2f (%.2f%%) |", a.GrandInfra, infraPct)
	if a.TotalCalls > 0 {
		w("| Total llamadas a modelos (CloudTrail) | **%s** |", fmtInt(a.TotalCalls))
		w("| Usuarios / entidades activos | %d |", len(a.UserData))
		w("| Tipos de herramientas identificados | %d |", len(a.ToolCalls))
	}
	line("")
	line("---")
	line("")

	// Section 1: Cost by service
	line("## 1. Costo por Modelo/Servicio")
	line("")
	line("*Fuente: AWS Cost Explorer*")
	line("")
	line("| Modelo / Servicio | Costo USD | % del total |")
	line("|-------------------|----------:|------------:|")
	for _, svc := range sortedKeys(a.TotalsSvc) {
		cost := a.TotalsSvc[svc]
		pct := 0.0
		if a.GrandTotal > 0 {
			pct = cost / a.GrandTotal * 100
		}
		w("| %s | $%.4f | %.2f%% |", svc, cost, pct)
	}
	w("| **TOTAL** | **$%.4f** | **100%%** |", a.GrandTotal)
	line("")
	line("---")
	line("")

	// Section 2: Daily costs
	if len(a.DailyTotals) > 0 {
		line("## 2. Costo Diario")
		line("")
		line("| Fecha | AI (modelos) | Infra | Total | |")
		line("|-------|-------------:|------:|------:|---|")
		for _, d := range a.DailyTotals {
			spike := ""
			if d.IsSpike {
				spike = "← pico"
			}
			w("| %s | $%.2f | $%.4f | $%.2f | %s |", d.Date, d.AI, d.Infra, d.Total, spike)
		}
		w("| **TOTAL** | **$%.2f** | **$%.4f** | **$%.2f** | |", a.GrandAI, a.GrandInfra, a.GrandTotal)
		line("")
		line("---")
		line("")
	}

	// Section 3: Tools
	if len(a.ToolCalls) > 0 {
		line("## 3. Herramientas Identificadas")
		line("")
		line("*Clasificado por userAgent en CloudTrail*")
		line("")
		line("| Herramienta | Llamadas | % |")
		line("|-------------|--------:|---:|")
		for _, tool := range sortedKeysByInt(a.ToolCalls) {
			cnt := a.ToolCalls[tool]
			pct := float64(cnt) / float64(a.TotalCalls) * 100
			w("| %s | %s | %.1f%% |", tool, fmtInt(cnt), pct)
		}
		w("| **TOTAL** | **%s** | **100%%** |", fmtInt(a.TotalCalls))
		line("")
		line("---")
		line("")
	}

	// Section 4: Users
	if len(a.UserData) > 0 {
		line("## 4. Desglose por Usuario / Entidad")
		line("")
		for _, user := range sortedUsersByCallsDesc(a.UserData) {
			data := a.UserData[user]
			w("### %s", user)
			line("")
			w("**Total llamadas**: %s", fmtInt(data.Calls))
			line("")
			if len(data.Tools) > 0 {
				line("| Herramienta | Llamadas |")
				line("|-------------|--------:|")
				for _, t := range sortedKeysByInt(data.Tools) {
					w("| %s | %s |", t, fmtInt(data.Tools[t]))
				}
				line("")
			}
			if len(data.Models) > 0 {
				line("| Modelo | Llamadas |")
				line("|--------|--------:|")
				for _, m := range sortedKeysByInt(data.Models) {
					w("| %s | %s |", m, fmtInt(data.Models[m]))
				}
				line("")
			}
		}
		line("---")
		line("")
	}

	// Section 5: Models consolidated
	if len(a.ModelCalls) > 0 {
		line("## 5. Modelos Usados — Vista Consolidada")
		line("")
		line("| Modelo | Llamadas | % |")
		line("|--------|--------:|---:|")
		for _, m := range sortedKeysByInt(a.ModelCalls) {
			cnt := a.ModelCalls[m]
			pct := float64(cnt) / float64(a.TotalCalls) * 100
			w("| %s | %s | %.1f%% |", m, fmtInt(cnt), pct)
		}
		w("| **TOTAL** | **%s** | **100%%** |", fmtInt(a.TotalCalls))
		line("")
		line("---")
		line("")
	}

	// Section 6: Caching
	line("## 6. Impacto del Caching")
	line("")
	line("*Fuente: Bedrock USAGE_TYPE en Cost Explorer*")
	line("")
	line("| Tipo | Costo USD |")
	line("|------|----------:|")
	w("| Cache read | $%.4f |", a.CacheReadCost)
	w("| Cache write | $%.4f |", a.CacheWriteCost)
	w("| Sin cache (input/output) | $%.4f |", a.NoCacheCost)
	line("")
	if a.TotalCalls > 0 {
		line("*Llamadas por prefijo de modelo:*")
		line("")
		line("| Prefijo | Caching | Llamadas (aprox) |")
		line("|---------|---------|----------------:|")
		w("| `global.*` | ✅ Con cache | ~%s |", fmtInt(a.CachedCalls))
		w("| `us.*` / sin prefijo | ❌ Sin cache | ~%s |", fmtInt(a.UncachedCalls))
		line("")
	}
	line("---")
	line("")

	// Section 7: Infrastructure
	if len(a.InfraSvcs) > 0 {
		line("## 7. Infraestructura (Servicios no-IA)")
		line("")
		line("| Servicio | Costo |")
		line("|----------|------:|")
		infraSorted := make([]string, len(a.InfraSvcs))
		copy(infraSorted, a.InfraSvcs)
		sort.Slice(infraSorted, func(i, j int) bool {
			return a.TotalsSvc[infraSorted[i]] > a.TotalsSvc[infraSorted[j]]
		})
		for _, svc := range infraSorted {
			w("| %s | $%.4f |", svc, a.TotalsSvc[svc])
		}
		w("| **Total** | **$%.4f** |", a.GrandInfra)
		line("")
		line("---")
		line("")
	}

	// Section 8: Tag costs
	if len(a.TagCosts) > 0 && string(a.TagCosts) != "null" {
		line("## 8. Costos por Proyecto (Tag)")
		line("")
		monthly := gjson.GetBytes(a.TagCosts, "monthly_by_project")
		grand := 0.0
		monthly.ForEach(func(_, v gjson.Result) bool { grand += v.Float(); return true })

		tagActive := false
		monthly.ForEach(func(k, _ gjson.Result) bool {
			if k.String() != "(sin tag)" {
				tagActive = true
				return false
			}
			return true
		})

		if !tagActive {
			line("> ⚠ **Todo el costo aparece como '(sin tag)'.**")
			line("> El tag `project` no está activado como cost-allocation tag.")
			line("> Solo se puede activar desde la cuenta root/management en Billing Console.")
			line("")
		}

		line("| Proyecto (tag) | Costo USD | % |")
		line("|----------------|----------:|--:|")
		type kv struct{ k string; v float64 }
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
			w("| %s | $%.2f | %.1f%% |", p.k, p.v, pct)
		}
		w("| **TOTAL** | **$%.2f** | **100%%** |", grand)
		line("")

		svcProj := gjson.GetBytes(a.TagCosts, "service_by_project")
		if svcProj.IsObject() && len(svcProj.Raw) > 2 {
			line("### Costos por Servicio × Proyecto")
			line("")
			line("| Servicio | Total | Desglose |")
			line("|----------|------:|---------|")
			type svcKV struct {
				svc   string
				total float64
				proj  gjson.Result
			}
			var svcs []svcKV
			svcProj.ForEach(func(k, v gjson.Result) bool {
				total := 0.0
				v.ForEach(func(_, c gjson.Result) bool { total += c.Float(); return true })
				svcs = append(svcs, svcKV{k.String(), total, v})
				return true
			})
			sort.Slice(svcs, func(i, j int) bool { return svcs[i].total > svcs[j].total })
			for _, s := range svcs {
				if s.total < 0.01 {
					continue
				}
				type pkv struct{ k string; v float64 }
				var projs []pkv
				s.proj.ForEach(func(k, v gjson.Result) bool {
					projs = append(projs, pkv{k.String(), v.Float()})
					return true
				})
				sort.Slice(projs, func(i, j int) bool { return projs[i].v > projs[j].v })
				var parts []string
				for _, p := range projs {
					parts = append(parts, fmt.Sprintf("%s: $%.2f", p.k, p.v))
				}
				w("| %s | $%.2f | %s |", s.svc, s.total, strings.Join(parts, ", "))
			}
			line("")
		}

		line("---")
		line("")
	}

	// Section 9: Limitations
	line("## 9. Limitaciones")
	line("")
	line("| Limitación | Detalle |")
	line("|------------|---------|")
	line("| Costo por usuario es estimado | Cost Explorer no desglosa por IAM user. |")
	if len(regions) > 0 {
		w("| Regiones consultadas | CloudTrail: %s |", strings.Join(regions, ", "))
	}
	line("| Tags cost-allocation | Requieren activación desde cuenta root/management. |")
	line("")

	return os.WriteFile(path, []byte(b.String()), 0644)
}

// fmtInt formats an integer with comma separators.
func fmtInt(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
