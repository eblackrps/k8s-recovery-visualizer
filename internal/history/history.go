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
