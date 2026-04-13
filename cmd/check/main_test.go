package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/gates"
	"k8s-recovery-visualizer/internal/model"
)

func TestRunCurrentBundleJSONOutput(t *testing.T) {
	current := testBundle(86, "GOLD")
	current.Inventory.Backup.PrimaryTool = "velero"
	current.Inventory.Backup.Policies = []model.BackupPolicy{{Tool: "velero", Name: "daily"}}

	currentPath := writeJSONFixture(t, "current.json", current)

	var out bytes.Buffer
	code := run([]string{"--current", currentPath, "--min-score", "80", "--format", "json"}, &out, os.ReadFile)
	if code != 0 {
		t.Fatalf("run() exit code = %d, want 0\n%s", code, out.String())
	}

	var eval gates.Evaluation
	if err := json.Unmarshal(out.Bytes(), &eval); err != nil {
		t.Fatalf("failed to decode json output: %v\n%s", err, out.String())
	}
	if eval.Status != gates.StatusPass {
		t.Fatalf("status = %s, want %s", eval.Status, gates.StatusPass)
	}
	if eval.CurrentScore != 86 {
		t.Fatalf("current score = %d, want 86", eval.CurrentScore)
	}
}

func TestRunCurrentBundleFailsRegressionAndOffsiteGates(t *testing.T) {
	previous := testBundle(92, "PLATINUM")
	previous.Inventory.Backup.PrimaryTool = "velero"
	previous.Inventory.Backup.HasOffsite = true
	previous.Inventory.Backup.Policies = []model.BackupPolicy{{Tool: "velero", Name: "daily"}}

	current := testBundle(80, "GOLD")
	current.Inventory.Backup.PrimaryTool = "velero"
	current.Inventory.Backup.Policies = []model.BackupPolicy{{Tool: "velero", Name: "daily"}}
	current.Inventory.Findings = []model.Finding{{ID: "RBAC_WILDCARD_VERB", Severity: "CRITICAL", ResourceID: "clusterrole/adminish"}}

	prevPath := writeJSONFixture(t, "previous.json", previous)
	currPath := writeJSONFixture(t, "current.json", current)

	var out bytes.Buffer
	code := run([]string{
		"--current", currPath,
		"--previous", prevPath,
		"--max-drop", "5",
		"--fail-on-new-critical",
		"--fail-on-offsite-loss",
		"--format", "json",
	}, &out, os.ReadFile)
	if code != 1 {
		t.Fatalf("run() exit code = %d, want 1\n%s", code, out.String())
	}

	var eval gates.Evaluation
	if err := json.Unmarshal(out.Bytes(), &eval); err != nil {
		t.Fatalf("failed to decode json output: %v\n%s", err, out.String())
	}
	if eval.Status != gates.StatusFail {
		t.Fatalf("status = %s, want %s", eval.Status, gates.StatusFail)
	}
	wantFailed := map[string]bool{
		"score-drop":            false,
		"new-critical-findings": false,
		"offsite-loss":          false,
	}
	for _, result := range eval.Results {
		if _, ok := wantFailed[result.ID]; ok && result.Status == gates.StatusFail {
			wantFailed[result.ID] = true
		}
	}
	for id, seen := range wantFailed {
		if !seen {
			t.Fatalf("missing failed gate result for %s in %#v", id, eval.Results)
		}
	}
}

func TestRunLegacyInputFallbackStillWorks(t *testing.T) {
	legacy := legacyEnriched{
		SchemaVersion: "v1",
		Trend: &legacyTrend{
			DeltaScore:   -1.5,
			DeltaPercent: -2.0,
		},
	}
	legacy.Risk.Score = 78
	legacy.Risk.Maturity = "SILVER"
	legacy.Risk.Posture = "MODERATE"

	legacyPath := writeJSONFixture(t, "legacy.json", legacy)

	var out bytes.Buffer
	code := run([]string{"--in", legacyPath, "--max-risk", "HIGH", "--max-drop", "5", "--format", "json"}, &out, os.ReadFile)
	if code != 0 {
		t.Fatalf("run() exit code = %d, want 0\n%s", code, out.String())
	}

	var eval gates.Evaluation
	if err := json.Unmarshal(out.Bytes(), &eval); err != nil {
		t.Fatalf("failed to decode json output: %v\n%s", err, out.String())
	}
	if eval.Status != gates.StatusPass {
		t.Fatalf("status = %s, want %s", eval.Status, gates.StatusPass)
	}
}

func TestRunReturnsUsageStyleFailureForMissingInput(t *testing.T) {
	var out bytes.Buffer
	code := run([]string{"--current", filepath.Join(t.TempDir(), "missing.json")}, &out, os.ReadFile)
	if code != 2 {
		t.Fatalf("run() exit code = %d, want 2", code)
	}
	if !bytes.Contains(out.Bytes(), []byte("CHECK FAIL: cannot read")) {
		t.Fatalf("expected missing-file error, got %q", out.String())
	}
}

func TestRunSupportsDomainThresholdsAndFindingBudgets(t *testing.T) {
	previous := testBundle(90, "PLATINUM")
	previous.Score.Workload.Final = 88
	previous.Score.Backup.Final = 90
	previous.Inventory.Findings = []model.Finding{
		{ID: "BACKUP_NO_POLICIES", Severity: "MEDIUM", ResourceID: "cluster/demo"},
	}

	current := testBundle(82, "GOLD")
	current.Score.Workload.Final = 74
	current.Score.Backup.Final = 81
	current.Inventory.Findings = []model.Finding{
		{ID: "BACKUP_NO_POLICIES", Severity: "CRITICAL", ResourceID: "cluster/demo"},
		{ID: "PVC_UNBOUND", Severity: "HIGH", ResourceID: "payments/db"},
	}

	prevPath := writeJSONFixture(t, "previous.json", previous)
	currPath := writeJSONFixture(t, "current.json", current)

	var out bytes.Buffer
	code := run([]string{
		"--current", currPath,
		"--previous", prevPath,
		"--min-workload-score", "80",
		"--min-backup-score", "85",
		"--max-critical-findings", "0",
		"--max-high-findings", "0",
		"--max-new-findings", "0",
		"--max-regressed-findings", "0",
		"--format", "json",
	}, &out, os.ReadFile)
	if code != 1 {
		t.Fatalf("run() exit code = %d, want 1\n%s", code, out.String())
	}

	var eval gates.Evaluation
	if err := json.Unmarshal(out.Bytes(), &eval); err != nil {
		t.Fatalf("failed to decode json output: %v\n%s", err, out.String())
	}
	assertHasGate(t, eval, "workload-score")
	assertHasGate(t, eval, "backup-score")
	assertHasGate(t, eval, "critical-finding-budget")
	assertHasGate(t, eval, "high-finding-budget")
	assertHasGate(t, eval, "new-finding-budget")
	assertHasGate(t, eval, "regressed-finding-budget")
}

func TestRunHelpPrintsUsage(t *testing.T) {
	var out bytes.Buffer
	code := run([]string{"--help"}, &out, os.ReadFile)
	if code != 0 {
		t.Fatalf("run() exit code = %d, want 0\n%s", code, out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Usage of check:")) {
		t.Fatalf("expected usage header, got %q", out.String())
	}
	if strings.Count(out.String(), "Usage of check:") != 1 {
		t.Fatalf("expected usage to print once, got %q", out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("max-regressed-findings")) {
		t.Fatalf("expected extended gate help, got %q", out.String())
	}
}

func writeJSONFixture(t *testing.T, name string, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func testBundle(score int, maturity string) model.Bundle {
	b := model.NewBundle("scan-test", time.Unix(1_700_000_000, 0).UTC())
	b.Score.Overall.Final = score
	b.Score.Storage.Final = score
	b.Score.Workload.Final = score
	b.Score.Config.Final = score
	b.Score.Backup.Final = score
	b.Score.Maturity = maturity
	return b
}

func assertHasGate(t *testing.T, eval gates.Evaluation, id string) {
	t.Helper()
	for _, result := range eval.Results {
		if result.ID == id {
			return
		}
	}
	t.Fatalf("gate %s not found in %#v", id, eval.Results)
}
