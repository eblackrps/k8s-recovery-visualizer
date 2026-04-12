package scanapp

import (
	"encoding/json"
	"fmt"
	"time"

	"k8s-recovery-visualizer/internal/analyze"
	"k8s-recovery-visualizer/internal/model"
)

// PrintCISummary remains on the supported CLI path so CI callers get the same
// machine-readable summary output after scans complete.
func PrintCISummary(b *model.Bundle, minScore int, trendLabel string, trendDelta int) {
	counts := model.FindingCounts{}
	for _, f := range b.Inventory.Findings {
		counts.Total++
		switch f.Severity {
		case "CRITICAL":
			counts.Critical++
		case "HIGH":
			counts.High++
		case "MEDIUM":
			counts.Medium++
		default:
			counts.Low++
		}
	}
	summary := model.ScanSummary{
		ScanID:       b.Scan.ScanID,
		TimestampUtc: time.Now().UTC().Format(time.RFC3339),
		Overall:      b.Score.Overall.Final,
		Maturity:     b.Score.Maturity,
		Status:       "PASSED",
		MinScore:     minScore,
		Profile:      b.Profile,
		Categories:   analyze.BuildCategories(b),
		Trend:        trendLabel,
		Delta:        trendDelta,
		Findings:     counts,
	}
	if b.Score.Overall.Final < minScore {
		summary.Status = "FAILED"
	}
	raw, _ := json.Marshal(summary)
	fmt.Println()
	fmt.Println(string(raw))
}
