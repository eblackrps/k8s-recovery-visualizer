package enrich

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s-recovery-visualizer/internal/profile"
	"k8s-recovery-visualizer/internal/risk"
	"k8s-recovery-visualizer/internal/trend"
)

const SchemaVersion = "v1"

type CategoryScore struct {
	Name     string  `json:"name"`
	Raw      float64 `json:"raw"`
	Weight   float64 `json:"weight"`
	Weighted float64 `json:"weighted"`
	Max      float64 `json:"max"`
	Grade    string  `json:"grade"`
}

type CategoryDelta struct {
	Name  string  `json:"name"`
	From  float64 `json:"from"`
	To    float64 `json:"to"`
	Delta float64 `json:"delta"`
}

type HistoryIndex struct {
	Entries []HistoryEntry `json:"entries"`
}

type HistoryEntry struct {
	TimestampUtc string  `json:"timestampUtc"`
	Overall      float64 `json:"overall"`
	Maturity     string  `json:"maturity"`
}

type Enriched struct {
	SchemaVersion string        `json:"schemaVersion"`
	GeneratedUtc  string        `json:"generatedUtc"`
	Profile       string        `json:"profile"`
	Current       HistoryEntry  `json:"current"`
	Previous      *HistoryEntry `json:"previous,omitempty"`
	Trend         *trend.Trend  `json:"trend,omitempty"`
	Risk          risk.Rating   `json:"risk"`
	LastN         []float64     `json:"lastN"`

	Categories         []CategoryScore `json:"categories,omitempty"`
	ProfileOverall     *float64        `json:"profileOverall,omitempty"`
	ProfileRiskPosture *string         `json:"profileRiskPosture,omitempty"`
	CategoryDeltas     []CategoryDelta `json:"categoryDeltas,omitempty"`
}

type Options struct {
	OutDir     string
	LastNCount int
	Profile    string
}

type scanLike struct {
	Categories []CategoryScore `json:"categories"`
}

func Run(opts Options) (*Enriched, error) {
	if opts.OutDir == "" {
		opts.OutDir = "out"
	}
	if opts.LastNCount <= 0 {
		opts.LastNCount = 10
	}

	p := strings.TrimSpace(opts.Profile)
	if p == "" {
		p = os.Getenv("DR_PROFILE")
	}
	pn := profile.Normalize(p)

	// overall history
	historyPath := filepath.Join(opts.OutDir, "history", "index.json")
	b, err := os.ReadFile(historyPath)
	if err != nil {
		return &Enriched{
			SchemaVersion: SchemaVersion,
			GeneratedUtc:  time.Now().UTC().Format(time.RFC3339),
			Profile:       string(pn),
			Risk:          risk.FromScore(0, ""),
			LastN:         []float64{},
		}, nil
	}

	var idx HistoryIndex
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, fmt.Errorf("parse history index: %w", err)
	}
	if len(idx.Entries) == 0 {
		return &Enriched{
			SchemaVersion: SchemaVersion,
			GeneratedUtc:  time.Now().UTC().Format(time.RFC3339),
			Profile:       string(pn),
			Risk:          risk.FromScore(0, ""),
			LastN:         []float64{},
		}, nil
	}

	curr := idx.Entries[len(idx.Entries)-1]
	var prev *HistoryEntry
	if len(idx.Entries) >= 2 {
		p := idx.Entries[len(idx.Entries)-2]
		prev = &p
	}

	start := 0
	if len(idx.Entries) > opts.LastNCount {
		start = len(idx.Entries) - opts.LastNCount
	}
	last := make([]float64, 0, len(idx.Entries)-start)
	for i := start; i < len(idx.Entries); i++ {
		last = append(last, idx.Entries[i].Overall)
	}

	var tr *trend.Trend
	if prev != nil {
		t := trend.Compute(prev.Overall, curr.Overall)
		tr = &t
	}

	en := &Enriched{
		SchemaVersion: SchemaVersion,
		GeneratedUtc:  time.Now().UTC().Format(time.RFC3339),
		Profile:       string(pn),
		Current:       curr,
		Previous:      prev,
		Trend:         tr,
		Risk:          risk.FromScore(curr.Overall, curr.Maturity),
		LastN:         last,
	}

	// best-effort categories from recovery-scan.json
	scanPath := filepath.Join(opts.OutDir, "recovery-scan.json")
	if sb, err := os.ReadFile(scanPath); err == nil {
		var sl scanLike
		if json.Unmarshal(sb, &sl) == nil && len(sl.Categories) > 0 {
			en.Categories = sl.Categories

			if po, pr := computeProfileOverall(sl.Categories, profile.Weights(pn)); po != nil {
				en.ProfileOverall = po
				en.ProfileRiskPosture = pr
			}
			en.CategoryDeltas = computeCategoryDeltas(opts.OutDir, sl.Categories)
		}
	}

	return en, nil
}

func computeProfileOverall(cats []CategoryScore, w map[string]float64) (*float64, *string) {
	if len(w) == 0 {
		return nil, nil
	}
	totalMax := 0.0
	totalGot := 0.0
	for _, c := range cats {
		mul := 1.0
		if v, ok := w[c.Name]; ok {
			mul = v
		}
		totalMax += c.Max * mul
		totalGot += c.Weighted * mul
	}
	if totalMax <= 0.00001 {
		return nil, nil
	}
	score := (totalGot / totalMax) * 100.0
	r := risk.FromScore(score, "")
	p := string(r.Posture)
	return &score, &p
}

type enrichIndex struct {
	Entries []enrichEntry `json:"entries"`
}
type enrichEntry struct {
	TimestampUtc  string          `json:"timestampUtc"`
	Path          string          `json:"path"`
	Categories    []CategoryScore `json:"categories"`
	SchemaVersion string          `json:"schemaVersion"`
	Profile       string          `json:"profile"`
}

func computeCategoryDeltas(outDir string, current []CategoryScore) []CategoryDelta {
	idxPath := filepath.Join(outDir, "history", "enriched-index.json")
	b, err := os.ReadFile(idxPath)
	if err != nil {
		return []CategoryDelta{}
	}
	var ix enrichIndex
	if json.Unmarshal(b, &ix) != nil {
		return []CategoryDelta{}
	}
	if len(ix.Entries) < 1 {
		return []CategoryDelta{}
	}

	prevEntry := ix.Entries[len(ix.Entries)-1]
	prevMap := map[string]CategoryScore{}
	for _, c := range prevEntry.Categories {
		prevMap[c.Name] = c
	}

	deltas := make([]CategoryDelta, 0, len(current))
	for _, c := range current {
		if p, ok := prevMap[c.Name]; ok {
			d := c.Weighted - p.Weighted
			deltas = append(deltas, CategoryDelta{Name: c.Name, From: p.Weighted, To: c.Weighted, Delta: d})
		}
	}
	return deltas
}
