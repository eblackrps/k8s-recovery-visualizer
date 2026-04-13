package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
)

func TestParseSettingsMergesPartialFilesWithDefaults(t *testing.T) {
	defaults := defaultSettings()
	raw := []byte(`{"workspaceRoot":"C:/demo","summary":false}`)

	settings, err := parseSettings(raw, defaults)
	if err != nil {
		t.Fatalf("parseSettings() error = %v", err)
	}

	if settings.WorkspaceRoot != "C:/demo" {
		t.Fatalf("WorkspaceRoot = %q, want %q", settings.WorkspaceRoot, "C:/demo")
	}
	if settings.Summary {
		t.Fatal("Summary = true, want false from partial file")
	}
	if settings.Runbook != defaults.Runbook {
		t.Fatalf("Runbook = %v, want default %v", settings.Runbook, defaults.Runbook)
	}
	if settings.CSVExport != defaults.CSVExport {
		t.Fatalf("CSVExport = %v, want default %v", settings.CSVExport, defaults.CSVExport)
	}
}

func TestOpenBundleResolvesRelativePathsAgainstWorkspaceRoot(t *testing.T) {
	workspaceRoot := t.TempDir()
	outputDir := filepath.Join(workspaceRoot, "demo-out")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	bundle := model.NewBundle("scan-demo", time.Date(2026, time.April, 12, 12, 0, 0, 0, time.UTC))
	bundle.Metadata.ClusterName = "prod-east"
	bundle.Metadata.GeneratedAt = "2026-04-12T12:00:00Z"

	raw, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	bundlePath := filepath.Join(outputDir, "recovery-scan.json")
	if err := os.WriteFile(bundlePath, raw, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	app := NewApp()
	app.settings = Settings{
		WorkspaceRoot:    workspaceRoot,
		DefaultOutputDir: filepath.Join(workspaceRoot, "out"),
		DefaultProfile:   "standard",
		Summary:          true,
		Runbook:          true,
		CSVExport:        true,
	}

	workspace, err := app.OpenBundle(filepath.Join("demo-out", "recovery-scan.json"))
	if err != nil {
		t.Fatalf("OpenBundle() error = %v", err)
	}

	if workspace.Artifacts.LoadedBundlePath != bundlePath {
		t.Fatalf("LoadedBundlePath = %q, want %q", workspace.Artifacts.LoadedBundlePath, bundlePath)
	}
}
