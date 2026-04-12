package appcore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"k8s-recovery-visualizer/internal/history"
	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/theme"
)

type Service struct {
	now  func() time.Time
	uuid func() string
}

func NewService() *Service {
	return &Service{
		now: func() time.Time { return time.Now().UTC() },
		uuid: func() string {
			return model.NewUUID()
		},
	}
}

func (s *Service) Bootstrap() Bootstrap {
	return Bootstrap{Theme: theme.Default()}
}

func (s *Service) LoadWorkspace(path string) (Workspace, error) {
	bundle, err := loadBundle(path)
	if err != nil {
		return Workspace{}, err
	}
	dir := filepath.Dir(path)
	artifacts := detectArtifacts(dir)
	artifacts.LoadedBundlePath = path
	artifacts.LoadedBundleDirHint = dir
	return s.workspaceFromBundle(*bundle, artifacts), nil
}

func (s *Service) ListProjects(root string) ([]ProjectSummary, error) {
	if root == "" {
		root = "."
	}
	summaries := map[string]ProjectSummary{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "dist", "build":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Base(path), "recovery-scan.json") {
			bundle, err := loadBundle(path)
			if err != nil {
				return nil
			}
			dir := filepath.Dir(path)
			artifacts := detectArtifacts(dir)
			key := filepath.Clean(dir)
			summaries[key] = ProjectSummary{
				Name:         filepath.Base(dir),
				ClusterName:  bundle.Metadata.ClusterName,
				Environment:  bundle.Metadata.Environment,
				OutputDir:    dir,
				LastScanPath: path,
				ReportPath:   artifacts.HTMLReport,
				Score:        bundle.Score.Overall.Final,
				Maturity:     bundle.Score.Maturity,
				TimestampUTC: bundle.Metadata.GeneratedAt,
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	out := make([]ProjectSummary, 0, len(summaries))
	for _, summary := range summaries {
		out = append(out, summary)
	}
	sort.Slice(out, func(i, j int) bool {
		left, errLeft := time.Parse(time.RFC3339, out[i].TimestampUTC)
		right, errRight := time.Parse(time.RFC3339, out[j].TimestampUTC)
		switch {
		case errLeft == nil && errRight == nil:
			return left.After(right)
		case errLeft == nil:
			return true
		case errRight == nil:
			return false
		default:
			return out[i].OutputDir < out[j].OutputDir
		}
	})
	return out, nil
}

func (s *Service) workspaceFromBundle(bundle model.Bundle, artifacts ArtifactPaths) Workspace {
	entries := history.LoadRecent(artifacts.OutputDir, 10)
	historyEntries := make([]HistoryEntry, 0, len(entries))
	for _, entry := range entries {
		historyEntries = append(historyEntries, HistoryEntry{
			TimestampUTC: entry.TimestampUTC,
			Overall:      entry.Overall,
			Maturity:     entry.Maturity,
		})
	}
	return Workspace{
		Bundle:    bundle,
		Artifacts: artifacts,
		History: HistoryDashboard{
			Entries: historyEntries,
		},
		Source:   "bundle",
		LoadedAt: s.now().Format(time.RFC3339),
	}
}

func detectArtifacts(outDir string) ArtifactPaths {
	layout := artifactLayout(outDir)
	artifacts := ArtifactPaths{OutputDir: outDir}
	artifacts.BundleJSON = existingArtifact(layout.BundleJSON)
	artifacts.EnrichedJSON = existingArtifact(layout.EnrichedJSON)
	artifacts.HTMLReport = existingArtifact(layout.HTMLReport)
	artifacts.MarkdownReport = existingArtifact(layout.MarkdownReport)
	artifacts.SummaryHTML = existingArtifact(layout.SummaryHTML)
	artifacts.RunbookHTML = existingArtifact(layout.RunbookHTML)
	artifacts.RedactedJSON = existingArtifact(layout.RedactedJSON)
	artifacts.RedactedHTML = existingArtifact(layout.RedactedHTML)
	artifacts.CSVDir = existingDir(layout.CSVDir)
	artifacts.HistoryIndex = existingArtifact(layout.HistoryIndex)
	artifacts.HistoryLatestHTML = existingArtifact(layout.HistoryLatestHTML)
	return artifacts
}

func artifactLayout(outDir string) ArtifactPaths {
	return ArtifactPaths{
		OutputDir:         outDir,
		BundleJSON:        filepath.Join(outDir, "recovery-scan.json"),
		EnrichedJSON:      filepath.Join(outDir, "recovery-enriched.json"),
		HTMLReport:        filepath.Join(outDir, "recovery-report.html"),
		MarkdownReport:    filepath.Join(outDir, "recovery-report.md"),
		SummaryHTML:       filepath.Join(outDir, "recovery-summary.html"),
		RunbookHTML:       filepath.Join(outDir, "recovery-runbook.html"),
		RedactedJSON:      filepath.Join(outDir, "recovery-scan-redacted.json"),
		RedactedHTML:      filepath.Join(outDir, "recovery-report-redacted.html"),
		CSVDir:            filepath.Join(outDir, "csv"),
		HistoryIndex:      filepath.Join(outDir, "history", "index.json"),
		HistoryLatestHTML: filepath.Join(outDir, "history", "latest", "recovery-report.html"),
	}
}

func loadBundle(path string) (*model.Bundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var bundle model.Bundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, err
	}
	return &bundle, nil
}

func cloneBundle(bundle *model.Bundle) (*model.Bundle, error) {
	raw, err := json.Marshal(bundle)
	if err != nil {
		return nil, err
	}
	var out model.Bundle
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func existingArtifact(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func existingDir(path string) string {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return path
	}
	return ""
}

func emit(sink EventSink, event RunEvent) {
	if sink == nil {
		return
	}
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	sink.Emit(event)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
