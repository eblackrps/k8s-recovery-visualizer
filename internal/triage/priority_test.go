package triage

import (
	"testing"

	"k8s-recovery-visualizer/internal/model"
)

func TestApplyDecoratesAndRanksFindings(t *testing.T) {
	findings := []model.Finding{
		{
			ID:             "BACKUP_NO_POLICIES",
			Severity:       "HIGH",
			ResourceID:     "cluster/demo",
			Message:        "No schedules were found.",
			Recommendation: "Create a recurring backup policy.",
			Confidence:     model.EvidenceConfidenceConfirmed,
		},
		{
			ID:             "NETPOL_MISSING_NAMESPACE",
			Severity:       "MEDIUM",
			ResourceID:     "namespace/frontend",
			Message:        "Namespace is missing a default policy.",
			Recommendation: "Create default deny plus explicit allow rules.",
			Confidence:     model.EvidenceConfidenceConfirmed,
		},
	}

	got := Apply(findings)
	if got[0].ID != "BACKUP_NO_POLICIES" {
		t.Fatalf("top finding = %s, want BACKUP_NO_POLICIES", got[0].ID)
	}
	if got[0].Rank != 1 || got[1].Rank != 2 {
		t.Fatalf("ranks = %d,%d, want 1,2", got[0].Rank, got[1].Rank)
	}
	if got[0].OwnerHint == "" || got[0].Impact == "" || got[0].Effort == "" || got[0].PriorityScore == 0 {
		t.Fatalf("expected decorated fields on top finding, got %+v", got[0])
	}
	if got[1].Title == "" {
		t.Fatalf("expected scoring title to populate, got %+v", got[1])
	}
}
