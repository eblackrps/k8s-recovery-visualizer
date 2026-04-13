package restore

import (
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
)

func TestSimulateLeavesCoverageUnknownWhenPoliciesAreUnverified(t *testing.T) {
	b := model.NewBundle("scan-test", time.Now().UTC())
	b.Inventory.Backup.PrimaryTool = "rubrik"
	b.Inventory.Backup.CoverageVerified = false
	b.Inventory.Backup.CoverageStatus = model.BackupCoverageStatusUnsupported
	b.Inventory.StatefulSets = []model.StatefulSet{
		{Namespace: "prod", Name: "db"},
	}

	sim := Simulate(&b)
	if len(sim.Namespaces) != 1 {
		t.Fatalf("Simulate() namespaces len = %d, want 1", len(sim.Namespaces))
	}

	ns := sim.Namespaces[0]
	if ns.CoverageKnown {
		t.Fatal("Simulate() coverageKnown = true, want false for detection-only backup tools")
	}
	if ns.HasCoverage {
		t.Fatal("Simulate() hasCoverage = true, want false when coverage is unverified")
	}
	if len(sim.UncoveredNS) != 0 {
		t.Fatalf("Simulate() uncovered namespaces = %v, want none while coverage is unknown", sim.UncoveredNS)
	}
	if ns.Readiness != "unknown" {
		t.Fatalf("Simulate() readiness = %q, want unknown", ns.Readiness)
	}
	if sim.UnknownNamespaces != 1 {
		t.Fatalf("Simulate() UnknownNamespaces = %d, want 1", sim.UnknownNamespaces)
	}
}

func TestSimulateBuildsReadinessSummaryAndDrillPlan(t *testing.T) {
	b := model.NewBundle("scan-test", time.Now().UTC())
	b.Profile = "enterprise"
	b.Target = "vm"
	b.Metadata.ClusterName = "prod-east"
	b.Inventory.Backup.PrimaryTool = "velero"
	b.Inventory.Backup.CoverageVerified = true
	b.Inventory.Backup.Policies = []model.BackupPolicy{
		{Tool: "velero", Name: "prod", IncludedNS: []string{"prod"}, RPOHours: 4, HasOffsite: true},
	}
	b.Inventory.StatefulSets = []model.StatefulSet{
		{Namespace: "prod", Name: "db"},
		{Namespace: "payments", Name: "queue"},
	}
	b.Inventory.PVCs = []model.PersistentVolumeClaim{
		{Namespace: "prod", Name: "db-data", StorageClass: "fast", RequestedSize: "10Gi"},
		{Namespace: "payments", Name: "queue-data", StorageClass: "fast", RequestedSize: "5Gi"},
	}
	b.Inventory.PVs = []model.PersistentVolume{
		{Name: "pv-prod", ClaimRef: "prod/db-data", Backend: "csi", Capacity: "10Gi"},
	}
	b.Inventory.StorageClasses = []model.StorageClass{{Name: "fast"}}

	sim := Simulate(&b)
	if sim.ReadyNamespaces != 1 || sim.BlockedNamespaces != 1 {
		t.Fatalf("Simulate() ready/blocked = %d/%d, want 1/1", sim.ReadyNamespaces, sim.BlockedNamespaces)
	}
	if sim.EstimatedDataAtRiskGB <= 0 {
		t.Fatalf("Simulate() EstimatedDataAtRiskGB = %.2f, want > 0", sim.EstimatedDataAtRiskGB)
	}
	if len(sim.BlockingReasons) == 0 {
		t.Fatal("Simulate() BlockingReasons = empty, want blocker summary")
	}

	b.Inventory.Backup.RestoreSim = &sim
	plan := BuildDrillPlan(&b)
	if len(plan) < 4 {
		t.Fatalf("BuildDrillPlan() len = %d, want at least 4 steps", len(plan))
	}
	if plan[1].OwnerHint == "" {
		t.Fatalf("BuildDrillPlan() missing owner hint: %+v", plan[1])
	}
}
