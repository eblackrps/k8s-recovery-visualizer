package gates

import (
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
)

func TestEvaluateFlagsNewCriticalAndOffsiteLoss(t *testing.T) {
	previous := model.NewBundle("prev", time.Now().UTC())
	previous.Score.Overall.Final = 95
	previous.Score.Maturity = "PLATINUM"
	previous.Inventory.Backup.HasOffsite = true

	current := model.NewBundle("curr", time.Now().UTC())
	current.Score.Overall.Final = 88
	current.Score.Maturity = "GOLD"
	current.Inventory.Backup.HasOffsite = false
	current.Inventory.Findings = []model.Finding{
		{ID: "PV_HOSTPATH", Severity: "CRITICAL", ResourceID: "pv-1"},
	}

	eval := Evaluate(&current, &previous, Policy{
		MaxDrop:           5,
		FailOnNewCritical: true,
		FailOnOffsiteLoss: true,
	})

	if eval.Status != StatusFail {
		t.Fatalf("overall status = %s, want FAIL", eval.Status)
	}
	assertGateStatus(t, eval, "score-drop", StatusFail)
	assertGateStatus(t, eval, "new-critical-findings", StatusFail)
	assertGateStatus(t, eval, "offsite-loss", StatusFail)
}

func TestEvaluateFlagsMissingPoliciesAndUncoveredStateful(t *testing.T) {
	current := model.NewBundle("curr", time.Now().UTC())
	current.Score.Overall.Final = 91
	current.Score.Maturity = "PLATINUM"
	current.Inventory.Backup.PrimaryTool = "velero"
	current.Inventory.Backup.UncoveredStatefulNS = []string{"prod"}

	eval := Evaluate(&current, nil, Policy{
		FailOnUncoveredStateful:   true,
		FailOnMissingBackupPolicy: true,
	})

	if eval.Status != StatusFail {
		t.Fatalf("overall status = %s, want FAIL", eval.Status)
	}
	assertGateStatus(t, eval, "uncovered-stateful", StatusFail)
	assertGateStatus(t, eval, "missing-backup-policies", StatusFail)
}

func TestEvaluatePassesWhenPolicyIsSatisfied(t *testing.T) {
	previous := model.NewBundle("prev", time.Now().UTC())
	previous.Score.Overall.Final = 90
	previous.Score.Maturity = "PLATINUM"
	previous.Inventory.Backup.HasOffsite = true

	current := model.NewBundle("curr", time.Now().UTC())
	current.Score.Overall.Final = 92
	current.Score.Maturity = "PLATINUM"
	current.Inventory.Backup.PrimaryTool = "velero"
	current.Inventory.Backup.HasOffsite = true
	current.Inventory.Backup.Policies = []model.BackupPolicy{{Tool: "velero", Name: "daily"}}

	eval := Evaluate(&current, &previous, Policy{
		MinScore:                  90,
		MaxRisk:                   "MODERATE",
		MaxDrop:                   5,
		FailOnNewCritical:         true,
		FailOnOffsiteLoss:         true,
		FailOnMissingBackupPolicy: true,
	})

	if eval.Status != StatusPass {
		t.Fatalf("overall status = %s, want PASS", eval.Status)
	}
}

func TestEvaluateSupportsDomainThresholdsAndRegressionBudgets(t *testing.T) {
	previous := model.NewBundle("prev", time.Now().UTC())
	previous.Score.Overall.Final = 88
	previous.Score.Storage.Final = 86
	previous.Score.Workload.Final = 85
	previous.Score.Config.Final = 84
	previous.Score.Backup.Final = 90
	previous.Score.Maturity = "GOLD"
	previous.Inventory.Findings = []model.Finding{
		{ID: "BACKUP_NO_POLICIES", Severity: "MEDIUM", ResourceID: "cluster/demo"},
	}

	current := model.NewBundle("curr", time.Now().UTC())
	current.Score.Overall.Final = 82
	current.Score.Storage.Final = 86
	current.Score.Workload.Final = 74
	current.Score.Config.Final = 84
	current.Score.Backup.Final = 81
	current.Score.Maturity = "GOLD"
	current.Inventory.Findings = []model.Finding{
		{ID: "BACKUP_NO_POLICIES", Severity: "CRITICAL", ResourceID: "cluster/demo"},
		{ID: "PVC_UNBOUND", Severity: "HIGH", ResourceID: "payments/db"},
	}

	eval := Evaluate(&current, &previous, Policy{
		MinWorkloadScore:     80,
		MinBackupScore:       85,
		MaxCriticalFindings:  0,
		MaxHighFindings:      0,
		MaxNewFindings:       0,
		MaxRegressedFindings: 0,
	})

	if eval.Status != StatusFail {
		t.Fatalf("overall status = %s, want FAIL", eval.Status)
	}
	assertGateStatus(t, eval, "workload-score", StatusFail)
	assertGateStatus(t, eval, "backup-score", StatusFail)
	assertGateStatus(t, eval, "critical-finding-budget", StatusFail)
	assertGateStatus(t, eval, "high-finding-budget", StatusFail)
	assertGateStatus(t, eval, "new-finding-budget", StatusFail)
	assertGateStatus(t, eval, "regressed-finding-budget", StatusFail)
}

func assertGateStatus(t *testing.T, eval Evaluation, id string, want Status) {
	t.Helper()
	for _, result := range eval.Results {
		if result.ID == id {
			if result.Status != want {
				t.Fatalf("gate %s status = %s, want %s", id, result.Status, want)
			}
			return
		}
	}
	t.Fatalf("gate %s not found", id)
}
