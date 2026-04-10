package scanapp

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"k8s-recovery-visualizer/internal/analyze"
	"k8s-recovery-visualizer/internal/enrich"
	"k8s-recovery-visualizer/internal/history"
	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/output"
)

func WriteOutputs(bundle *model.Bundle, outDir string, quiet bool, minScore int, csvExport, summaryOut, redactOut, runbookOut bool) (string, int, error) {
	bundle.Scan.EndedAt = time.Now().UTC()
	bundle.Scan.DurationSeconds = int(bundle.Scan.EndedAt.Sub(bundle.Scan.StartedAt).Seconds())
	bundle.Checks = analyze.BuildChecks(bundle, minScore)

	jsonPath := filepath.Join(outDir, "recovery-scan.json")
	htmlPath := filepath.Join(outDir, "recovery-report.html")

	if err := output.WriteJSON(jsonPath, bundle); err != nil {
		return "", 0, fmt.Errorf("write json: %w", err)
	}

	var trendLabel string
	var trendDelta int
	if en, err := enrich.Run(enrich.Options{OutDir: outDir, LastNCount: 10, Profile: bundle.Profile}); err != nil {
		if !quiet {
			fmt.Printf("Enrich: FAILED (%v)\n", err)
		}
	} else if err := enrich.WriteArtifacts(outDir, en); err != nil {
		if !quiet {
			fmt.Printf("Enrich: FAILED writing artifacts (%v)\n", err)
		}
	}

	if tr, err := history.Record(outDir, bundle); err != nil {
		if !quiet {
			fmt.Println("History: (skipped)", err)
		}
		trendLabel = "HISTORY_SKIPPED"
	} else if tr.Label == "FIRST_RUN" {
		if !quiet {
			fmt.Println("Trend: FIRST RUN (no previous scan found)")
		}
		trendLabel = "FIRST_RUN"
	} else {
		trendLabel = tr.Label
		trendDelta = tr.Delta
		if !quiet {
			sign := ""
			if tr.Delta > 0 {
				sign = "+"
			}
			fmt.Printf("Trend: %s (%s%d) Previous: %d, Current: %d\n",
				tr.Label, sign, tr.Delta, tr.Previous, tr.Current)
		}
	}

	bundle.TrendHistory = history.LoadRecent(outDir, 20)

	if err := output.WriteReport(htmlPath, bundle); err != nil {
		return "", 0, fmt.Errorf("write html report: %w", err)
	}
	if err := history.SnapshotLatestHTML(outDir); err != nil && !quiet {
		fmt.Println("History HTML snapshot: (skipped)", err)
	}

	if csvExport {
		if err := output.WriteCSV(outDir, bundle); err != nil {
			return "", 0, fmt.Errorf("write csv: %w", err)
		}
		if !quiet {
			fmt.Println("CSV exports:", filepath.Join(outDir, "csv"))
		}
	}

	if summaryOut {
		summaryPath := filepath.Join(outDir, "recovery-summary.html")
		if err := output.WriteSummary(summaryPath, bundle); err != nil {
			return "", 0, fmt.Errorf("write summary: %w", err)
		}
		if !quiet {
			fmt.Println("Executive Summary:", summaryPath)
		}
	}

	if runbookOut {
		runbookPath := filepath.Join(outDir, "recovery-runbook.html")
		if err := output.WriteRunbook(runbookPath, bundle); err != nil {
			return "", 0, fmt.Errorf("write runbook: %w", err)
		}
		if !quiet {
			fmt.Println("DR Runbook:", runbookPath)
		}
	}

	if redactOut {
		if err := output.WriteRedactedJSON(filepath.Join(outDir, "recovery-scan-redacted.json"), bundle); err != nil {
			return "", 0, fmt.Errorf("write redacted json: %w", err)
		}
		if err := output.WriteRedactedReport(filepath.Join(outDir, "recovery-report-redacted.html"), bundle); err != nil {
			return "", 0, fmt.Errorf("write redacted html: %w", err)
		}
		if !quiet {
			fmt.Println("Redacted exports: recovery-scan-redacted.json, recovery-report-redacted.html")
		}
	}

	if !quiet {
		fmt.Println("Scan complete.")
		fmt.Println("JSON:", jsonPath)
		fmt.Println("HTML Report:", htmlPath)
		fmt.Println("Enriched:", filepath.Join(outDir, "recovery-enriched.json"))
	}

	return trendLabel, trendDelta, nil
}

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
