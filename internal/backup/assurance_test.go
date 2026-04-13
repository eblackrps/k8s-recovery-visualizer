package backup

import (
	"strings"
	"testing"

	"k8s-recovery-visualizer/internal/model"
)

func TestAssessAssuranceUsesEvidenceConfirmedConclusion(t *testing.T) {
	b := model.Bundle{
		Inventory: model.Inventory{
			PVCs:                  []model.PersistentVolumeClaim{{Namespace: "prod", Name: "db"}},
			VolumeSnapshotClasses: []model.VolumeSnapshotClass{{Name: "snap"}},
			VolumeSnapshots:       []model.VolumeSnapshot{{Namespace: "prod", Name: "db-snap", PVCName: "db"}},
			Backup: model.BackupInventory{
				PrimaryTool:      "velero",
				CoverageVerified: true,
				HasOffsite:       true,
				Policies: []model.BackupPolicy{
					{Name: "daily", Tool: "velero", LastSuccessAt: "2026-04-10T06:00:00Z", FreshSchedule: true},
				},
				RestoreSim: &model.RestoreSimResult{
					Namespaces: []model.RestoreSimNamespace{{Namespace: "prod", CoverageKnown: true, HasCoverage: true}},
				},
			},
		},
	}

	AssessAssurance(&b)

	if b.Inventory.Backup.Assurance == nil {
		t.Fatal("Assurance = nil, want populated assurance result")
	}
	if b.Inventory.Backup.Assurance.Conclusion != model.BackupAssuranceEvidenceConfirmed {
		t.Fatalf("Conclusion = %q, want %q", b.Inventory.Backup.Assurance.Conclusion, model.BackupAssuranceEvidenceConfirmed)
	}
	if strings.Contains(strings.ToLower(b.Inventory.Backup.Assurance.Summary), "recoverable") {
		t.Fatalf("Summary = %q, should not claim recoverability", b.Inventory.Backup.Assurance.Summary)
	}
}

func TestAssessAssuranceReportsOffsiteNamespaceGap(t *testing.T) {
	b := model.Bundle{
		Inventory: model.Inventory{
			Backup: model.BackupInventory{
				PrimaryTool:      "velero",
				CoverageVerified: true,
				HasOffsite:       false,
				OffsiteMissingNS: []string{"staging"},
				Policies: []model.BackupPolicy{
					{Name: "prod", Tool: "velero", LastSuccessAt: "2026-04-10T06:00:00Z", FreshSchedule: true},
				},
			},
		},
	}

	AssessAssurance(&b)

	if b.Inventory.Backup.Assurance == nil {
		t.Fatal("Assurance = nil, want populated assurance result")
	}
	if b.Inventory.Backup.Assurance.Conclusion != model.BackupAssuranceCoverageGap {
		t.Fatalf("Conclusion = %q, want %q", b.Inventory.Backup.Assurance.Conclusion, model.BackupAssuranceCoverageGap)
	}
	found := false
	for _, signal := range b.Inventory.Backup.Assurance.Signals {
		if signal.ID == "offsite" {
			found = true
			if !strings.Contains(signal.Detail, "staging") {
				t.Fatalf("offsite detail = %q, want missing namespace detail", signal.Detail)
			}
		}
	}
	if !found {
		t.Fatal("missing offsite assurance signal")
	}
}
