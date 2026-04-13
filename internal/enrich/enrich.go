package enrich

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s-recovery-visualizer/internal/analyze"
	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/profile"
	"k8s-recovery-visualizer/internal/risk"
	"k8s-recovery-visualizer/internal/trend"
)

const SchemaVersion = model.EnrichedSchemaVersion

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
	RunCount      int           `json:"runCount,omitempty"`
	AverageScore  *float64      `json:"averageScore,omitempty"`
	BestScore     *float64      `json:"bestScore,omitempty"`
	WorstScore    *float64      `json:"worstScore,omitempty"`

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
	currentFromScan := loadCurrentFromScan(opts.OutDir)

	// overall history
	historyPath := filepath.Join(opts.OutDir, "history", "index.json")
	b, err := os.ReadFile(historyPath)
	if err != nil {
		currentAverage := currentFromScan.Overall
		return &Enriched{
			SchemaVersion: SchemaVersion,
			GeneratedUtc:  time.Now().UTC().Format(time.RFC3339),
			Profile:       string(pn),
			Current:       currentFromScan,
			Risk:          risk.FromScore(currentFromScan.Overall, currentFromScan.Maturity),
			LastN:         lastNFromCurrent(currentFromScan),
			RunCount:      1,
			AverageScore:  &currentAverage,
			BestScore:     &currentAverage,
			WorstScore:    &currentAverage,
		}, nil
	}

	var idx HistoryIndex
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, fmt.Errorf("parse history index: %w", err)
	}
	if len(idx.Entries) == 0 {
		currentAverage := currentFromScan.Overall
		return &Enriched{
			SchemaVersion: SchemaVersion,
			GeneratedUtc:  time.Now().UTC().Format(time.RFC3339),
			Profile:       string(pn),
			Current:       currentFromScan,
			Risk:          risk.FromScore(currentFromScan.Overall, currentFromScan.Maturity),
			LastN:         lastNFromCurrent(currentFromScan),
			RunCount:      1,
			AverageScore:  &currentAverage,
			BestScore:     &currentAverage,
			WorstScore:    &currentAverage,
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
	best := idx.Entries[0].Overall
	worst := idx.Entries[0].Overall
	total := 0.0
	for i := start; i < len(idx.Entries); i++ {
		last = append(last, idx.Entries[i].Overall)
	}
	for _, entry := range idx.Entries {
		total += entry.Overall
		if entry.Overall > best {
			best = entry.Overall
		}
		if entry.Overall < worst {
			worst = entry.Overall
		}
	}

	var tr *trend.Trend
	if prev != nil {
		t := trend.Compute(prev.Overall, curr.Overall)
		tr = &t
	}
	avg := total / float64(len(idx.Entries))

	en := &Enriched{
		SchemaVersion: SchemaVersion,
		GeneratedUtc:  time.Now().UTC().Format(time.RFC3339),
		Profile:       string(pn),
		Current:       curr,
		Previous:      prev,
		Trend:         tr,
		Risk:          risk.FromScore(curr.Overall, curr.Maturity),
		LastN:         last,
		RunCount:      len(idx.Entries),
		AverageScore:  &avg,
		BestScore:     &best,
		WorstScore:    &worst,
	}

	// best-effort categories from recovery-scan.json
	scanPath := filepath.Join(opts.OutDir, "recovery-scan.json")
	if sb, err := os.ReadFile(scanPath); err == nil {
		var bundle model.Bundle
		if json.Unmarshal(sb, &bundle) == nil {
			categories := toEnrichedCategories(analyze.BuildCategories(&bundle))
			if len(categories) > 0 {
				en.Categories = categories

				if po, pr := computeProfileOverall(categories, profile.Weights(pn)); po != nil {
					en.ProfileOverall = po
					en.ProfileRiskPosture = pr
				}
				en.CategoryDeltas = computeCategoryDeltas(opts.OutDir, categories)
			}
		}
	}

	return en, nil
}

func loadCurrentFromScan(outDir string) HistoryEntry {
	scanPath := filepath.Join(outDir, "recovery-scan.json")
	raw, err := os.ReadFile(scanPath)
	if err != nil {
		return HistoryEntry{}
	}
	var bundle model.Bundle
	if err := json.Unmarshal(raw, &bundle); err != nil {
		return HistoryEntry{}
	}
	return HistoryEntry{
		TimestampUtc: time.Now().UTC().Format(time.RFC3339),
		Overall:      float64(bundle.Score.Overall.Final),
		Maturity:     bundle.Score.Maturity,
	}
}

func lastNFromCurrent(current HistoryEntry) []float64 {
	if current.Overall == 0 {
		return []float64{}
	}
	return []float64{current.Overall}
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

type scanLike struct {
	Score struct {
		Overall struct {
			Final int `json:"final"`
		} `json:"overall"`
		Maturity string `json:"maturity"`
	} `json:"score"`
}

func toEnrichedCategories(categories []model.CategoryScore) []CategoryScore {
	out := make([]CategoryScore, 0, len(categories))
	for _, category := range categories {
		out = append(out, CategoryScore{
			Name:     category.Name,
			Raw:      category.Raw,
			Weight:   category.Weight,
			Weighted: category.Weighted,
			Max:      category.Max,
			Grade:    category.Grade,
		})
	}
	return out
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
