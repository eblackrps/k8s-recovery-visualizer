package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s-recovery-visualizer/internal/model"
)

type IndexEntry struct {
	TimestampUTC string            `json:"timestampUtc"`
	CustomerID   string            `json:"customerId,omitempty"`
	Site         string            `json:"site,omitempty"`
	ClusterName  string            `json:"clusterName,omitempty"`
	Environment  string            `json:"environment,omitempty"`
	Overall      int               `json:"overall"`
	Maturity     string            `json:"maturity"`
	Storage      model.DomainScore `json:"storage"`
	Workload     model.DomainScore `json:"workload"`
	Config       model.DomainScore `json:"config"`
	Backup       model.DomainScore `json:"backup"`
	Findings     int               `json:"findings,omitempty"`
	JSONFile     string            `json:"jsonFile"`
	MDFile       string            `json:"mdFile"`
	HTMLFile     string            `json:"htmlFile"`
}

type Index struct {
	Entries []IndexEntry `json:"entries"`
}

type Trend struct {
	Previous int
	Current  int
	Delta    int
	Label    string // IMPROVING / DECLINING / SAME / FIRST_RUN
}

type DomainTrend struct {
	Name      string
	Current   int
	Delta     int
	Direction string
}

type Dashboard struct {
	Entries      []model.TrendPoint
	TrendLabel   string
	TrendDelta   int
	AverageScore int
	BestScore    int
	WorstScore   int
	RunCount     int
	DomainTrends []DomainTrend
}

func Record(outDir string, b *model.Bundle) (Trend, error) {
	historyDir := filepath.Join(outDir, "history")
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return Trend{}, err
	}

	indexPath := filepath.Join(historyDir, "index.json")
	var idx Index

	if raw, err := os.ReadFile(indexPath); err == nil && len(raw) > 0 {
		_ = json.Unmarshal(raw, &idx)
	}

	prev := -1
	if len(idx.Entries) > 0 {
		prev = idx.Entries[len(idx.Entries)-1].Overall
	}

	ts := time.Now().UTC().Format("20060102-150405")

	jsonName := fmt.Sprintf("recovery-scan-%s.json", ts)
	mdName := fmt.Sprintf("recovery-report-%s.md", ts)
	htmlName := fmt.Sprintf("recovery-report-%s.html", ts)

	if err := writeJSON(filepath.Join(historyDir, jsonName), b); err != nil {
		return Trend{}, err
	}

	_ = copyIfExists(filepath.Join(outDir, "recovery-report.md"), filepath.Join(historyDir, mdName))
	_ = copyIfExists(filepath.Join(outDir, "recovery-report.html"), filepath.Join(historyDir, htmlName))

	entry := IndexEntry{
		TimestampUTC: time.Now().UTC().Format(time.RFC3339),
		CustomerID:   b.Metadata.CustomerID,
		Site:         b.Metadata.Site,
		ClusterName:  b.Metadata.ClusterName,
		Environment:  b.Metadata.Environment,
		Overall:      b.Score.Overall.Final,
		Maturity:     b.Score.Maturity,
		Storage:      b.Score.Storage,
		Workload:     b.Score.Workload,
		Config:       b.Score.Config,
		Backup:       b.Score.Backup,
		Findings:     len(b.Inventory.Findings),
		JSONFile:     filepath.ToSlash(filepath.Join("history", jsonName)),
		MDFile:       filepath.ToSlash(filepath.Join("history", mdName)),
		HTMLFile:     filepath.ToSlash(filepath.Join("history", htmlName)),
	}

	idx.Entries = append(idx.Entries, entry)

	if len(idx.Entries) > 200 {
		idx.Entries = idx.Entries[len(idx.Entries)-200:]
	}

	raw, _ := json.MarshalIndent(idx, "", "  ")
	if err := os.WriteFile(indexPath, raw, 0644); err != nil {
		return Trend{}, err
	}

	tr := Trend{Previous: prev, Current: b.Score.Overall.Final, Delta: 0, Label: "FIRST_RUN"}

	if prev >= 0 {
		tr.Delta = tr.Current - tr.Previous
		if tr.Delta > 0 {
			tr.Label = "IMPROVING"
		} else if tr.Delta < 0 {
			tr.Label = "DECLINING"
		} else {
			tr.Label = "SAME"
		}
	}

	return tr, nil
}

// LoadRecent reads the last n scan entries from the history index and returns
// them as TrendPoints for sparkline rendering. Returns nil if no history exists.
func LoadRecent(outDir string, n int) []model.TrendPoint {
	indexPath := filepath.Join(outDir, "history", "index.json")
	raw, err := os.ReadFile(indexPath)
	if err != nil || len(raw) == 0 {
		return nil
	}
	var idx Index
	if err := json.Unmarshal(raw, &idx); err != nil {
		return nil
	}
	entries := idx.Entries
	if len(entries) > n {
		entries = entries[len(entries)-n:]
	}
	pts := make([]model.TrendPoint, len(entries))
	for i, e := range entries {
		pts[i] = model.TrendPoint{
			TimestampUTC: e.TimestampUTC,
			Overall:      e.Overall,
			Storage:      e.Storage.Final,
			Workload:     e.Workload.Final,
			Config:       e.Config.Final,
			Backup:       e.Backup.Final,
			Findings:     e.Findings,
			Maturity:     e.Maturity,
		}
	}
	return pts
}

func LoadDashboard(outDir string, n int) Dashboard {
	entries := LoadRecent(outDir, n)
	if len(entries) == 0 {
		return Dashboard{}
	}

	best := entries[0].Overall
	worst := entries[0].Overall
	total := 0
	for _, entry := range entries {
		total += entry.Overall
		if entry.Overall > best {
			best = entry.Overall
		}
		if entry.Overall < worst {
			worst = entry.Overall
		}
	}

	dashboard := Dashboard{
		Entries:      entries,
		AverageScore: total / len(entries),
		BestScore:    best,
		WorstScore:   worst,
		RunCount:     len(entries),
	}
	if len(entries) >= 2 {
		last := entries[len(entries)-1]
		prev := entries[len(entries)-2]
		delta := last.Overall - prev.Overall
		dashboard.TrendDelta = delta
		switch {
		case delta > 0:
			dashboard.TrendLabel = "IMPROVING"
		case delta < 0:
			dashboard.TrendLabel = "DECLINING"
		default:
			dashboard.TrendLabel = "SAME"
		}
		dashboard.DomainTrends = []DomainTrend{
			buildDomainTrend("storage", prev.Storage, last.Storage),
			buildDomainTrend("workload", prev.Workload, last.Workload),
			buildDomainTrend("config", prev.Config, last.Config),
			buildDomainTrend("backup", prev.Backup, last.Backup),
		}
	} else {
		dashboard.TrendLabel = "FIRST_RUN"
	}
	return dashboard
}

func buildDomainTrend(name string, prev, curr int) DomainTrend {
	trend := DomainTrend{Name: name, Current: curr, Delta: curr - prev, Direction: "same"}
	switch {
	case trend.Delta > 0:
		trend.Direction = "up"
	case trend.Delta < 0:
		trend.Direction = "down"
	}
	return trend
}

// SnapshotLatestHTML updates the most recent history HTML artifact with the
// final rendered report from outDir after report generation completes.
func SnapshotLatestHTML(outDir string) error {
	indexPath := filepath.Join(outDir, "history", "index.json")
	raw, err := os.ReadFile(indexPath)
	if err != nil || len(raw) == 0 {
		return nil
	}
	var idx Index
	if err := json.Unmarshal(raw, &idx); err != nil {
		return nil
	}
	if len(idx.Entries) == 0 {
		return nil
	}

	dst := filepath.Join(outDir, filepath.FromSlash(idx.Entries[len(idx.Entries)-1].HTMLFile))
	if dst == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return copyIfExists(filepath.Join(outDir, "recovery-report.html"), dst)
}

func writeJSON(path string, b *model.Bundle) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(b)
}

func copyIfExists(src, dst string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return nil
	}
	return os.WriteFile(dst, raw, 0644)
}
