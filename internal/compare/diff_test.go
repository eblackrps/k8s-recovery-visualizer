package compare

import (
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
)

func TestDiffCapturesRegressionsAndScoreBreakdown(t *testing.T) {
	prev := model.NewBundle("prev", time.Now().UTC())
	prev.Score.Overall.Final = 90
	prev.Score.Storage.Final = 92
	prev.Score.Workload.Final = 88
	prev.Score.Config.Final = 90
	prev.Score.Backup.Final = 91
	prev.Score.Maturity = "PLATINUM"
	prev.Inventory.Findings = []model.Finding{
		{ID: "BACKUP_NO_POLICIES", Severity: "MEDIUM", ResourceID: "cluster/demo", Message: "old"},
		{ID: "NETPOL_MISSING_NAMESPACE", Severity: "HIGH", ResourceID: "namespace/frontend", Message: "old"},
	}
	prev.Inventory.Namespaces = []model.Namespace{{Name: "frontend"}}

	curr := model.NewBundle("curr", time.Now().UTC())
	curr.Score.Overall.Final = 84
	curr.Score.Storage.Final = 92
	curr.Score.Workload.Final = 80
	curr.Score.Config.Final = 83
	curr.Score.Backup.Final = 81
	curr.Score.Maturity = "GOLD"
	curr.Inventory.Findings = []model.Finding{
		{ID: "BACKUP_NO_POLICIES", Severity: "CRITICAL", ResourceID: "cluster/demo", Message: "new", OwnerHint: "Platform / backup owner", Impact: "coverage gap", Effort: "M"},
		{ID: "PVC_UNBOUND", Severity: "HIGH", ResourceID: "payments/db", Message: "new"},
	}
	curr.Inventory.Namespaces = []model.Namespace{{Name: "frontend"}, {Name: "payments"}}

	got := Diff(&prev, &curr)
	if got.CurrentScore != 84 || got.PreviousScore != 90 {
		t.Fatalf("score summary = %+v", got)
	}
	if got.PersistentFinding != 1 {
		t.Fatalf("persistent count = %d, want 1", got.PersistentFinding)
	}
	if len(got.FindingsNew) != 1 || got.FindingsNew[0].ID != "PVC_UNBOUND" {
		t.Fatalf("new findings = %+v, want PVC_UNBOUND", got.FindingsNew)
	}
	if len(got.FindingsResolved) != 1 || got.FindingsResolved[0].ID != "NETPOL_MISSING_NAMESPACE" {
		t.Fatalf("resolved findings = %+v, want NETPOL_MISSING_NAMESPACE", got.FindingsResolved)
	}
	if len(got.FindingsRegressed) != 1 || got.FindingsRegressed[0].ID != "BACKUP_NO_POLICIES" {
		t.Fatalf("regressed findings = %+v, want BACKUP_NO_POLICIES", got.FindingsRegressed)
	}
	if len(got.DomainDeltas) == 0 || got.DomainDeltas[0].Name != "overall" {
		t.Fatalf("domain deltas = %+v", got.DomainDeltas)
	}
	if len(got.SeverityDeltas) != 5 {
		t.Fatalf("severity deltas len = %d, want 5", len(got.SeverityDeltas))
	}
	if len(got.InventoryDeltas) != 4 {
		t.Fatalf("inventory deltas len = %d, want 4", len(got.InventoryDeltas))
	}
	if len(got.NamespacesAdded) != 1 || got.NamespacesAdded[0] != "payments" {
		t.Fatalf("namespaces added = %+v, want payments", got.NamespacesAdded)
	}
}
