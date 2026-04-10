package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/analyze"
	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/restore"
)

func TestWritePersistsUnsupportedBackupCoverageArtifacts(t *testing.T) {
	outDir := t.TempDir()

	b := model.NewBundle("scan-test", time.Now().UTC())
	b.Target = "vm"
	b.Profile = "standard"
	b.Inventory.Namespaces = []model.Namespace{
		{Name: "prod"},
	}
	b.Inventory.StatefulSets = []model.StatefulSet{
		{Namespace: "prod", Name: "db", HasVolumeClaim: true},
	}
	b.Inventory.Backup.PrimaryTool = "rubrik"
	b.Inventory.Backup.CoverageStatus = model.BackupCoverageStatusUnsupported
	b.Inventory.Backup.CoverageReason = "rubrik was detected, but this scanner does not yet inspect its policies or schedules."
	b.Inventory.Backup.Tools = []model.BackupDetectedTool{
		{
			Name:                   "rubrik",
			Detected:               true,
			PolicyInspectionStatus: model.BackupCoverageStatusUnsupported,
			PolicyInspectionDetail: "rubrik was detected, but this scanner does not yet inspect its policies or schedules.",
		},
	}

	sim := restore.Simulate(&b)
	b.Inventory.Backup.RestoreSim = &sim
	analyze.Evaluate(&b)

	write(&b, outDir, true, 0, false, false, false, false)

	jsonBytes, err := os.ReadFile(filepath.Join(outDir, "recovery-scan.json"))
	if err != nil {
		t.Fatalf("ReadFile(recovery-scan.json) error = %v", err)
	}

	var persisted model.Bundle
	if err := json.Unmarshal(jsonBytes, &persisted); err != nil {
		t.Fatalf("Unmarshal(recovery-scan.json) error = %v", err)
	}

	if persisted.Inventory.Backup.CoverageStatus != model.BackupCoverageStatusUnsupported {
		t.Fatalf("CoverageStatus = %q, want %q", persisted.Inventory.Backup.CoverageStatus, model.BackupCoverageStatusUnsupported)
	}
	if persisted.Inventory.Backup.CoverageReason == "" {
		t.Fatal("CoverageReason is empty, want explicit unsupported-tool detail")
	}
	if persisted.Inventory.Backup.RestoreSim == nil || len(persisted.Inventory.Backup.RestoreSim.Namespaces) != 1 {
		t.Fatalf("RestoreSim namespaces = %v, want one stateful namespace result", persisted.Inventory.Backup.RestoreSim)
	}
	if persisted.Inventory.Backup.RestoreSim.Namespaces[0].CoverageKnown {
		t.Fatal("RestoreSim namespace coverageKnown = true, want false for unsupported coverage")
	}

	htmlBytes, err := os.ReadFile(filepath.Join(outDir, "recovery-report.html"))
	if err != nil {
		t.Fatalf("ReadFile(recovery-report.html) error = %v", err)
	}
	html := string(htmlBytes)

	if !strings.Contains(html, "policy coverage could not be verified") {
		t.Fatal("HTML report did not explain that backup coverage is unverified")
	}
	if !strings.Contains(html, "unsupported") {
		t.Fatal("HTML report did not surface the unsupported inspection status")
	}
	if !strings.Contains(html, "unverified") {
		t.Fatal("HTML report did not render unverified restore coverage state")
	}
}
