package appcore

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/output"
)

func TestBuildCollectorPipelineOmitsSecretsByDefault(t *testing.T) {
	steps := buildCollectorPipeline(context.Background(), nil, nil, &model.Bundle{}, ScanRequest{})
	for _, step := range steps {
		if step.Name == "Secrets" {
			t.Fatal("Secrets collector present by default, want opt-in only")
		}
	}
}

func TestBuildCollectorPipelineIncludesSecretsWhenEnabled(t *testing.T) {
	steps := buildCollectorPipeline(context.Background(), nil, nil, &model.Bundle{}, ScanRequest{IncludeSecretMetadata: true})
	for _, step := range steps {
		if step.Name == "Secrets" {
			return
		}
	}
	t.Fatal("Secrets collector missing when IncludeSecretMetadata is true")
}

func TestDryRunFinalizationMatchesSinglePass(t *testing.T) {
	startedAt := time.Date(2026, time.April, 12, 14, 10, 0, 0, time.UTC)
	service := &Service{
		now:  func() time.Time { return startedAt.Add(45 * time.Second) },
		uuid: func() string { return "scan-dry-run" },
	}

	req := ScanRequest{
		DryRun:    true,
		OutputDir: t.TempDir(),
		Target:    "vm",
		MinScore:  0,
	}
	result, err := service.Run(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	expected := prepareBundle(req.Normalized(), startedAt.Add(45*time.Second), "scan-dry-run")
	if err := runDryScan(&expected, req.Normalized()); err != nil {
		t.Fatalf("runDryScan() error = %v", err)
	}
	finalizeBundle(&expected, req.Target)

	if got, want := findingCountsByID(result.Workspace.Bundle.Inventory.Findings), findingCountsByID(expected.Inventory.Findings); !reflect.DeepEqual(got, want) {
		t.Fatalf("finding counts mismatch after dry run\n got: %#v\nwant: %#v", got, want)
	}
	if got, want := remediationCountsByTitle(result.Workspace.Bundle.Inventory.RemediationSteps), remediationCountsByTitle(expected.Inventory.RemediationSteps); !reflect.DeepEqual(got, want) {
		t.Fatalf("remediation counts mismatch after dry run\n got: %#v\nwant: %#v", got, want)
	}
}

func TestExportBundleWritesOnlyRequestedArtifacts(t *testing.T) {
	sourceDir := t.TempDir()
	outputDir := t.TempDir()
	sourcePath := filepath.Join(sourceDir, "bundle-source.json")

	bundle := prepareBundle(ScanRequest{Target: "vm"}.Normalized(), time.Date(2026, time.April, 12, 14, 11, 0, 0, time.UTC), "scan-export")
	if err := runDryScan(&bundle, ScanRequest{Target: "vm"}); err != nil {
		t.Fatalf("runDryScan() error = %v", err)
	}
	finalizeBundle(&bundle, "vm")
	if err := output.WriteJSON(sourcePath, &bundle); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	originalRaw, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile(sourcePath) error = %v", err)
	}

	service := NewService()
	artifacts, err := service.ExportBundle(sourcePath, ExportRequest{
		OutputDir: outputDir,
		Summary:   true,
	})
	if err != nil {
		t.Fatalf("ExportBundle() error = %v", err)
	}

	if artifacts.SummaryHTML == "" {
		t.Fatal("SummaryHTML = empty, want generated summary path")
	}
	for field, value := range map[string]string{
		"BundleJSON":   artifacts.BundleJSON,
		"EnrichedJSON": artifacts.EnrichedJSON,
		"HTMLReport":   artifacts.HTMLReport,
		"RunbookHTML":  artifacts.RunbookHTML,
		"RedactedJSON": artifacts.RedactedJSON,
		"RedactedHTML": artifacts.RedactedHTML,
		"CSVDir":       artifacts.CSVDir,
	} {
		if value != "" {
			t.Fatalf("%s = %q, want empty for summary-only export", field, value)
		}
	}

	if _, err := os.Stat(artifacts.SummaryHTML); err != nil {
		t.Fatalf("Stat(summary) error = %v", err)
	}
	for _, unexpected := range []string{
		filepath.Join(outputDir, "recovery-scan.json"),
		filepath.Join(outputDir, "recovery-enriched.json"),
		filepath.Join(outputDir, "recovery-report.html"),
		filepath.Join(outputDir, "recovery-runbook.html"),
		filepath.Join(outputDir, "recovery-report-redacted.html"),
	} {
		if _, err := os.Stat(unexpected); err == nil {
			t.Fatalf("unexpected artifact written: %s", unexpected)
		}
	}

	rawAfter, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile(sourcePath after export) error = %v", err)
	}
	if !reflect.DeepEqual(originalRaw, rawAfter) {
		t.Fatal("ExportBundle() mutated the source bundle file")
	}
}

func TestExportBundleRequiresSelectedOutput(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "bundle-source.json")
	bundle := prepareBundle(ScanRequest{Target: "vm"}.Normalized(), time.Now().UTC(), "scan-export-empty")
	if err := output.WriteJSON(sourcePath, &bundle); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	service := NewService()
	if _, err := service.ExportBundle(sourcePath, ExportRequest{OutputDir: t.TempDir()}); err == nil {
		t.Fatal("ExportBundle() error = nil, want explicit validation failure")
	}
}

func findingCountsByID(findings []model.Finding) map[string]int {
	out := map[string]int{}
	for _, finding := range findings {
		out[finding.ID]++
	}
	return out
}

func remediationCountsByTitle(steps []model.RemediationStep) map[string]int {
	out := map[string]int{}
	for _, step := range steps {
		out[step.Title]++
	}
	return out
}
