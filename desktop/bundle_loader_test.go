package main

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
)

func TestOpenBundleUsesAbsoluteDirectoryPaths(t *testing.T) {
	bundleDir := createBundleDir(t, "prod-east")

	app := NewApp()
	app.settings = Settings{
		WorkspaceRoot:    filepath.Join(t.TempDir(), "unused"),
		DefaultOutputDir: filepath.Join(t.TempDir(), "unused", "out"),
		DefaultProfile:   "standard",
		Summary:          true,
		Runbook:          true,
		CSVExport:        true,
	}

	workspace, err := app.OpenBundle(bundleDir)
	if err != nil {
		t.Fatalf("OpenBundle() error = %v", err)
	}

	want := filepath.Join(bundleDir, "recovery-scan.json")
	if workspace.Artifacts.LoadedBundlePath != want {
		t.Fatalf("LoadedBundlePath = %q, want %q", workspace.Artifacts.LoadedBundlePath, want)
	}
}

func TestOpenBundleExtractsZipArchives(t *testing.T) {
	sourceDir := createBundleDir(t, "prod-archive")
	archivePath := filepath.Join(t.TempDir(), "bundle.zip")
	if err := zipBundleDir(archivePath, sourceDir); err != nil {
		t.Fatalf("zipBundleDir() error = %v", err)
	}

	app := NewApp()
	workspace, err := app.OpenBundle(archivePath)
	if err != nil {
		t.Fatalf("OpenBundle() error = %v", err)
	}

	if workspace.Bundle.Metadata.ClusterName != "prod-archive" {
		t.Fatalf("ClusterName = %q, want %q", workspace.Bundle.Metadata.ClusterName, "prod-archive")
	}
	if len(app.extractedBundleDirs) != 1 {
		t.Fatalf("len(extractedBundleDirs) = %d, want 1", len(app.extractedBundleDirs))
	}
	if filepath.Dir(workspace.Artifacts.LoadedBundlePath) != app.extractedBundleDirs[0] {
		t.Fatalf("LoadedBundlePath = %q, want extracted dir %q", workspace.Artifacts.LoadedBundlePath, app.extractedBundleDirs[0])
	}
}

func TestOpenBundleRejectsInvalidJSONWithHelpfulMessage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recovery-scan.json")
	if err := os.WriteFile(path, []byte("{not-json"), settingsFileMode); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "recovery-report.html"), []byte("<html></html>"), settingsFileMode); err != nil {
		t.Fatalf("WriteFile(report) error = %v", err)
	}

	app := NewApp()
	_, err := app.OpenBundle(path)
	if err == nil {
		t.Fatal("OpenBundle() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "not a readable recovery-scan.json bundle") {
		t.Fatalf("OpenBundle() error = %q, want readable bundle guidance", err)
	}
}

func TestFindBundleJSONListsCandidatesWhenBundleIsMissing(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "other.json"), []byte("{}"), settingsFileMode); err != nil {
		t.Fatalf("WriteFile(other.json) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "recovery-report.html"), []byte("<html></html>"), settingsFileMode); err != nil {
		t.Fatalf("WriteFile(report) error = %v", err)
	}

	_, err := findBundleJSON(dir)
	if err == nil {
		t.Fatal("findBundleJSON() error = nil, want diagnostic failure")
	}
	if !strings.Contains(err.Error(), "JSON candidates: other.json") || !strings.Contains(err.Error(), "HTML reports: recovery-report.html") {
		t.Fatalf("findBundleJSON() error = %q, want candidate diagnostics", err)
	}
}

func createBundleDir(t *testing.T, clusterName string) string {
	t.Helper()

	dir := t.TempDir()
	bundle := model.NewBundle("scan-demo", time.Date(2026, time.April, 12, 12, 0, 0, 0, time.UTC))
	bundle.Metadata.ClusterName = clusterName
	bundle.Metadata.GeneratedAt = "2026-04-12T12:00:00Z"

	raw, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "recovery-scan.json"), raw, settingsFileMode); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "recovery-report.html"), []byte("<html></html>"), settingsFileMode); err != nil {
		t.Fatalf("WriteFile(report) error = %v", err)
	}
	return dir
}

func zipBundleDir(archivePath, sourceDir string) error {
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	writer := zip.NewWriter(archiveFile)
	defer writer.Close()

	return filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		entry, err := writer.Create(filepath.ToSlash(relativePath))
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(entry, file); err != nil {
			return err
		}
		return nil
	})
}
