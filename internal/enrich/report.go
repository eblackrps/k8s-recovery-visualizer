package enrich

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func WriteArtifacts(outDir string, en *Enriched) error {
	if outDir == "" {
		outDir = "out"
	}

	jpath := filepath.Join(outDir, "recovery-enriched.json")
	jb, _ := json.MarshalIndent(en, "", "  ")
	if err := os.WriteFile(jpath, jb, 0644); err != nil {
		return fmt.Errorf("write enriched json: %w", err)
	}

	// Write enrich snapshot + index (for category deltas)
	// EnrichHistory: noisy by default; caller controls CI/quiet output
	if err := writeEnrichHistory(outDir, en); err != nil {
		return fmt.Errorf("enrich history write failed: %w", err)
	}

	// Markdown append
	mdPath := filepath.Join(outDir, "recovery-report.md")
	if b, err := os.ReadFile(mdPath); err == nil {
		aug := string(b) + "\n\n" + renderMarkdown(en) + "\n"
		if err := os.WriteFile(mdPath, []byte(aug), 0644); err != nil {
			return fmt.Errorf("append md report: %w", err)
		}
	}
	// HTML chart injection (best effort)
	_ = injectTrendHTML(outDir, en)

	return nil
}

func renderMarkdown(en *Enriched) string {
	var sb strings.Builder
	sb.WriteString("## Trend, Risk, Profiles\n\n")
	sb.WriteString(fmt.Sprintf("- **Schema:** %s\n", en.SchemaVersion))
	sb.WriteString(fmt.Sprintf("- **Profile:** %s\n", en.Profile))
	sb.WriteString(fmt.Sprintf("- **DR Risk Posture (base):** %s\n", en.Risk.Posture))
	if en.ProfileOverall != nil && en.ProfileRiskPosture != nil {
		sb.WriteString(fmt.Sprintf("- **Profile Score:** %0.2f\n", *en.ProfileOverall))
		sb.WriteString(fmt.Sprintf("- **Profile Risk Posture:** %s\n", *en.ProfileRiskPosture))
	} else {
		sb.WriteString("- **Profile Score:** (not available)\n")
	}

	if en.Trend == nil || en.Previous == nil {
		sb.WriteString("\n- **Trend:** FIRST RUN (no previous scan found)\n")
	} else {
		arrow := "→"
		switch en.Trend.Direction {
		case "up":
			arrow = "↑"
		case "down":
			arrow = "↓"
		}
		sb.WriteString(fmt.Sprintf("\n- **Trend:** %s %+0.2f (%+0.2f%%)\n", arrow, en.Trend.DeltaScore, en.Trend.DeltaPercent))
	}

	if len(en.Categories) > 0 {
		sb.WriteString("\n### Categories\n")
		for _, c := range en.Categories {
			sb.WriteString(fmt.Sprintf("- %s: %0.2f / %0.2f (w=%0.2f)\n", c.Name, c.Weighted, c.Max, c.Weight))
		}
	}

	if len(en.CategoryDeltas) > 0 {
		sb.WriteString("\n### Category Deltas (vs last enrich snapshot)\n")
		for _, d := range en.CategoryDeltas {
			sb.WriteString(fmt.Sprintf("- %s: %0.2f ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¾Ãƒâ€šÃ‚Â¢ %0.2f (ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Â¦Ãƒâ€šÃ‚Â½ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â %+0.2f)\n", d.Name, d.From, d.To, d.Delta))
		}
	}

	return sb.String()
}

func writeEnrichHistory(outDir string, en *Enriched) error {
	if len(en.Categories) == 0 {
		return nil
	}

	histDir := filepath.Join(outDir, "history", "enriched")
	if err := os.MkdirAll(histDir, 0755); err != nil {
		return err
	}

	ts := time.Now().UTC().Format("20060102T150405Z")
	snapName := fmt.Sprintf("%s.json", ts)
	snapPath := filepath.Join(histDir, snapName)

	snap := map[string]any{
		"timestampUtc":  ts,
		"schemaVersion": en.SchemaVersion,
		"profile":       en.Profile,
		"categories":    en.Categories,
	}

	jb, _ := json.MarshalIndent(snap, "", "  ")
	if err := os.WriteFile(snapPath, jb, 0644); err != nil {
		return err
	}

	idxPath := filepath.Join(outDir, "history", "enriched-index.json")
	var ix enrichIndex
	if b, err := os.ReadFile(idxPath); err == nil {
		_ = json.Unmarshal(b, &ix)
	}

	ix.Entries = append(ix.Entries, enrichEntry{
		TimestampUtc:  ts,
		Path:          filepath.ToSlash(filepath.Join("history", "enriched", snapName)),
		Categories:    en.Categories,
		SchemaVersion: en.SchemaVersion,
		Profile:       en.Profile,
	})

	ib, _ := json.MarshalIndent(ix, "", "  ")
	return os.WriteFile(idxPath, ib, 0644)
}

func injectTrendHTML(outDir string, en *Enriched) error {
	htmlPath := filepath.Join(outDir, "recovery-report.html")
	b, err := os.ReadFile(htmlPath)
	if err != nil {
		return nil // best effort
	}
	// Use LastN for chart if present; otherwise skip
	if len(en.LastN) == 0 {
		// HTML chart injection (best effort)
		_ = injectTrendHTML(outDir, en)

		return nil
	}

	data, _ := json.Marshal(en.LastN)

	block := fmt.Sprintf(`
<section style="margin-top:24px; padding:16px; border:1px solid #ddd; border-radius:12px;">
  <h2>Trend History</h2>
  <canvas id="drTrend" height="110"></canvas>
  <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
  <script>
    const scores = %s;
    const labels = scores.map((_, i) => "Run " + (i+1));
    const ctx = document.getElementById("drTrend");
    new Chart(ctx, {
      type: "line",
      data: { labels: labels, datasets: [{ label: "Score", data: scores }] },
      options: { responsive: true }
    });
  </script>
</section>
`, string(data))

	s := string(b)
	lower := strings.ToLower(s)
	idx := strings.LastIndex(lower, "</body>")
	if idx == -1 {
		s = s + "\n" + block
	} else {
		s = s[:idx] + "\n" + block + "\n" + s[idx:]
	}
	return os.WriteFile(htmlPath, []byte(s), 0644)
}
