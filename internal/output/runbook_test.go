package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
)

func TestBuildRunbookRendersRestoreDrillAndPrioritizedFindings(t *testing.T) {
	b := model.NewBundle("scan-test", time.Now().UTC())
	b.Inventory.Findings = []model.Finding{
		{
			ID:             "BACKUP_COVERAGE_UNVERIFIED",
			Severity:       "HIGH",
			ResourceID:     "velero",
			Message:        "Backup coverage is unverified",
			Recommendation: "Verify policy scope and rerun the assessment",
			OwnerHint:      "Platform / backup owner",
			Impact:         "coverage gap",
			Effort:         "M",
			Rank:           2,
		},
	}
	b.Inventory.Backup.RestoreSim = &model.RestoreSimResult{
		Namespaces: []model.RestoreSimNamespace{
			{Namespace: "payments", CoverageKnown: false, HasCoverage: false, Readiness: "unknown", Warnings: []string{"policy visibility denied"}},
		},
		ReadyNamespaces:       0,
		BlockedNamespaces:     0,
		WarningNamespaces:     0,
		UnknownNamespaces:     1,
		EstimatedDataAtRiskGB: 24,
		BlockingReasons:       []string{"policy visibility denied"},
	}
	b.Inventory.Backup.DrillPlan = []model.RestoreDrillStep{
		{Phase: "validation", Title: "Validate workload startup", Detail: "Confirm restored applications can serve traffic.", OwnerHint: "Application owner", Validation: []string{"Synthetic checks pass"}},
	}
	b.Inventory.RemediationSteps = []model.RemediationStep{
		{
			Priority:     1,
			Category:     "Backup",
			Title:        "Verify coverage",
			Detail:       "Grant read access to backup policy objects.",
			OwnerHint:    "Platform / security",
			Effort:       "S",
			WhyItMatters: "Coverage cannot be trusted without policy visibility.",
		},
	}

	var buf bytes.Buffer
	buildRunbook(&buf, &b)
	html := buf.String()

	for _, fragment := range []string{"Recommended Restore Drill Plan", "Platform / backup owner", "coverage gap", "Estimated data at risk", "Execution Notes"} {
		if !strings.Contains(html, fragment) {
			t.Fatalf("buildRunbook() missing fragment %q", fragment)
		}
	}
}
