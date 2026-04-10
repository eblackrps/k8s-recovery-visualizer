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
}
