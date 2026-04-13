package remediation

import (
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
)

func TestGenerateCarriesFindingOwnerAndEffort(t *testing.T) {
	b := model.NewBundle("scan-test", time.Now().UTC())
	b.Inventory.Findings = []model.Finding{
		{
			ID:             "BACKUP_NO_POLICIES",
			Severity:       "HIGH",
			ResourceID:     "cluster/demo",
			Message:        "No backup schedules found.",
			Recommendation: "Create a recurring backup schedule.",
			OwnerHint:      "Platform / backup owner",
			Effort:         "M",
		},
	}

	steps := Generate(&b, "vm")
	if len(steps) == 0 {
		t.Fatal("Generate() returned no remediation steps")
	}
	if steps[0].OwnerHint != "Platform / backup owner" || steps[0].Effort != "M" {
		t.Fatalf("step metadata = %+v, want owner/effort from finding", steps[0])
	}
}
