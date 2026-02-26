package enrich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func WriteArtifacts(outDir string, enriched *Enriched) error {
	if outDir == "" {
		outDir = "."
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("write artifacts: create outDir: %w", err)
	}

	now := time.Now().Format(time.RFC3339)

	// Load recovery-scan.json (source of truth)
	scanPath := filepath.Join(outDir, "recovery-scan.json")
	var scanAny map[string]any
	if b, err := os.ReadFile(scanPath); err == nil {
		_ = json.Unmarshal(b, &scanAny)
	}

	// Extract score/maturity/status from *likely* locations.
	// Your JSON has top-level keys: schemaVersion, metadata, tool, scan, cluster, inventory, ...
	// Score may live under: score, results, summary, or scanResult. We'll probe multiple.
	score := firstIntDeep(scanAny,
		"score.final",
		"score.overall",
		"results.score.final",
		"results.overall",
		"summary.score.final",
		"summary.overall",
		"scanResult.score.final",
		"scanResult.overall",
	)
	maturity := firstStringDeep(scanAny,
		"score.maturity",
		"results.maturity",
		"summary.maturity",
		"scanResult.maturity",
	)
	status := firstStringDeep(scanAny,
		"score.status",
		"results.status",
		"summary.status",
		"scanResult.status",
	)

	// Fallbacks from enriched object (if it contains anything)
	if score == 0 && enriched != nil {
		// enriched.Current.Overall is float64 in your codebase
		if enriched.Current.Overall != 0 {
			score = int(enriched.Current.Overall)
		}
	}
	if maturity == "" && enriched != nil && strings.TrimSpace(enriched.Current.Maturity) != "" {
		maturity = enriched.Current.Maturity
	}

	if status == "" {
		if score >= 90 && score != 0 {
			status = "PASSED"
		} else if score != 0 {
			status = "FAILED"
		} else {
			status = "UNKNOWN"
		}
	}

	issues := extractIssues(scanAny)
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Status != issues[j].Status {
			return issues[i].Status == "FAIL"
		}
		return issues[i].Weight > issues[j].Weight
	})

	// Write recovery-enriched.json (embed scan so it isn't empty)
	enrichedPath := filepath.Join(outDir, "recovery-enriched.json")
	enrichedOut := map[string]any{
		"schemaVersion": "v1",
		"generatedUtc":  time.Now().UTC().Format(time.RFC3339),
		"profile":       "standard",
		"current": map[string]any{
			"timestampUtc": time.Now().UTC().Format(time.RFC3339),
			"overall":      score,
			"maturity":     maturity,
			"status":       status,
		},
		"scan":           scanAny,
		"enrichedObject": enriched,
	}
	j, err := json.MarshalIndent(enrichedOut, "", "  ")
	if err != nil {
		return fmt.Errorf("write artifacts: marshal enriched json: %w", err)
	}
	if err := os.WriteFile(enrichedPath, j, 0o644); err != nil {
		return fmt.Errorf("write artifacts: write %s: %w", enrichedPath, err)
	}

	// Markdown report
	mdPath := filepath.Join(outDir, "recovery-report.md")
	var md bytes.Buffer
	md.WriteString("# Kubernetes DR Recovery Report\n\n")
	md.WriteString(fmt.Sprintf("- Generated: `%s`\n", now))
	if score != 0 {
		md.WriteString(fmt.Sprintf("- Final Score: `%d`\n", score))
	}
	if maturity != "" {
		md.WriteString(fmt.Sprintf("- DR Maturity: `%s`\n", maturity))
	}
	md.WriteString(fmt.Sprintf("- DR Status: `%s`\n", status))

	md.WriteString("\n\n## Top issues\n\n")
	if len(issues) == 0 {
		md.WriteString("- No structured issues detected in `recovery-scan.json`.\n")
		md.WriteString("- Next step: emit structured checks/findings into the scan output (see `score.checks` or `findings`).\n")
	} else {
		limit := 12
		if len(issues) < limit {
			limit = len(issues)
		}
		for i := 0; i < limit; i++ {
			it := issues[i]
			md.WriteString(fmt.Sprintf("- **%s**: %s\n", it.Status, it.Title))
			if it.Detail != "" {
				md.WriteString(fmt.Sprintf("  - %s\n", it.Detail))
			}
			if it.Remediation != "" {
				md.WriteString(fmt.Sprintf("  - Fix: %s\n", it.Remediation))
			}
		}
	}

	md.WriteString("\n## Artifacts\n\n")
	md.WriteString("- `recovery-scan.json`\n")
	md.WriteString("- `recovery-enriched.json`\n")
	md.WriteString("- `recovery-report.html`\n")

	if err := os.WriteFile(mdPath, md.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write artifacts: write %s: %w", mdPath, err)
	}

	// HTML report
	htmlPath := filepath.Join(outDir, "recovery-report.html")
	var h bytes.Buffer
	h.WriteString("<!doctype html><html><head><meta charset=\"utf-8\" />")
	h.WriteString("<title>Kubernetes DR Recovery Report</title>")
	h.WriteString("<style>body{font-family:system-ui,Segoe UI,Arial,sans-serif;max-width:980px;margin:24px;} code{background:#f4f4f4;padding:2px 4px;border-radius:4px;} .fail{color:#b00020;} .pass{color:#0b6;} .warn{color:#b36b00;} li{margin:10px 0;}</style>")
	h.WriteString("</head><body>")
	h.WriteString("<h1>Kubernetes DR Recovery Report</h1><ul>")
	h.WriteString("<li>Generated: <code>" + html.EscapeString(now) + "</code></li>")
	if score != 0 {
		h.WriteString(fmt.Sprintf("<li>Final Score: <code>%d</code></li>", score))
	}
	if maturity != "" {
		h.WriteString("<li>DR Maturity: <code>" + html.EscapeString(maturity) + "</code></li>")
	}
	h.WriteString("<li>DR Status: <code>" + html.EscapeString(status) + "</code></li>")
	h.WriteString("</ul><h2>Top issues</h2>")

	if len(issues) == 0 {
		h.WriteString("<p>No structured issues detected in <code>recovery-scan.json</code>.</p>")
	} else {
		h.WriteString("<ol>")
		limit := 12
		if len(issues) < limit {
			limit = len(issues)
		}
		for i := 0; i < limit; i++ {
			it := issues[i]
			cls := "warn"
			if it.Status == "FAIL" {
				cls = "fail"
			} else if it.Status == "PASS" {
				cls = "pass"
			}
			h.WriteString("<li class=\"" + cls + "\"><b>" + html.EscapeString(it.Status) + "</b>: " + html.EscapeString(it.Title))
			if it.Detail != "" {
				h.WriteString("<br/><small>" + html.EscapeString(it.Detail) + "</small>")
			}
			if it.Remediation != "" {
				h.WriteString("<br/><small><i>Fix:</i> " + html.EscapeString(it.Remediation) + "</small>")
			}
			h.WriteString("</li>")
		}
		h.WriteString("</ol>")
	}

	h.WriteString("<h2>Artifacts</h2><ul>")
	h.WriteString("<li><code>recovery-scan.json</code></li>")
	h.WriteString("<li><code>recovery-enriched.json</code></li>")
	h.WriteString("</ul></body></html>")

	if err := os.WriteFile(htmlPath, h.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write artifacts: write %s: %w", htmlPath, err)
	}

	return nil
}

type Issue struct {
	Status      string
	Title       string
	Detail      string
	Remediation string
	Weight      int
}

func firstIntDeep(root map[string]any, paths ...string) int {
	for _, p := range paths {
		if v, ok := getPath(root, p); ok {
			switch t := v.(type) {
			case float64:
				return int(t)
			case int:
				return t
			case string:
				var i int
				_, _ = fmt.Sscanf(t, "%d", &i)
				if i != 0 {
					return i
				}
			}
		}
	}
	return 0
}

func firstStringDeep(root map[string]any, paths ...string) string {
	for _, p := range paths {
		if v, ok := getPath(root, p); ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
	}
	return ""
}

func getPath(root map[string]any, path string) (any, bool) {
	if root == nil {
		return nil, false
	}
	cur := any(root)
	for _, part := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[part]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

func extractIssues(scanAny map[string]any) []Issue {
	if scanAny == nil {
		return nil
	}
	var issues []Issue

	// Common top-level arrays
	for _, key := range []string{"checks", "findings", "issues", "warnings", "errors"} {
		if arr, ok := scanAny[key].([]any); ok {
			for _, it := range arr {
				if obj, ok := it.(map[string]any); ok {
					issues = append(issues, issueFromObj(obj, key))
				}
			}
		}
	}

	// Common nested locations
	for _, key := range []string{"results", "summary", "scoring", "analysis", "score"} {
		if obj, ok := scanAny[key].(map[string]any); ok {
			issues = append(issues, extractIssues(obj)...)
		}
	}

	return issues
}

func issueFromObj(obj map[string]any, source string) Issue {
	getStr := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := obj[k]; ok {
				if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
					return s
				}
			}
		}
		return ""
	}
	getInt := func(keys ...string) int {
		for _, k := range keys {
			if v, ok := obj[k]; ok {
				switch t := v.(type) {
				case float64:
					return int(t)
				case int:
					return t
				}
			}
		}
		return 0
	}

	status := strings.ToUpper(getStr("status", "result", "state", "severity"))
	if status == "FAILED" {
		status = "FAIL"
	}
	if status == "PASSED" {
		status = "PASS"
	}
	if status == "" {
		status = strings.ToUpper(source)
	}

	title := getStr("name", "title", "check", "id")
	if title == "" {
		title = "(unnamed finding)"
	}
	detail := getStr("message", "reason", "detail", "description")
	rem := getStr("remediation", "fix", "recommendation", "action")

	weight := getInt("weight", "scoreDelta", "delta", "penalty")
	if status == "FAIL" && weight == 0 {
		weight = 10
	}

	return Issue{Status: status, Title: title, Detail: detail, Remediation: rem, Weight: weight}
}
