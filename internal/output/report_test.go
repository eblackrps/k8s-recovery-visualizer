package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
)

func TestBuildReportMarksDetectionOnlyBackupCoverageAsUnverified(t *testing.T) {
	b := model.NewBundle("scan-test", time.Now().UTC())
	b.Inventory.Backup.PrimaryTool = "rubrik"
	b.Inventory.Backup.CoverageVerified = false
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
	b.Inventory.Backup.RestoreSim = &model.RestoreSimResult{
		Namespaces: []model.RestoreSimNamespace{
			{
				Namespace:     "prod",
				CoverageKnown: false,
				HasCoverage:   false,
				RPOHours:      -1,
			},
		},
	}

	var buf bytes.Buffer
	buildReport(&buf, &b)
	html := buf.String()

	if !strings.Contains(html, "policy coverage could not be verified") {
		t.Fatal("buildReport() did not explain that backup coverage is unverified")
	}
	if !strings.Contains(html, "unverified") {
		t.Fatal("buildReport() did not render unverified restore coverage state")
	}
	if !strings.Contains(html, "unsupported") {
		t.Fatal("buildReport() did not surface the inspection status")
	}
}

func TestBuildReportRendersStructuredRemediationGuidance(t *testing.T) {
	b := model.NewBundle("scan-test", time.Now().UTC())
	b.Inventory.RemediationSteps = []model.RemediationStep{
		{
			Priority:     1,
			Category:     "Backup",
			Title:        "Resolve unverified backup coverage",
			Detail:       "Coverage could not be verified safely.",
			WhyItMatters: "Detection is not evidence of recoverability.",
			DRImpact:     "Teams may believe namespaces are protected when they are not.",
			Validation:   []string{"Check backup policy visibility."},
			FixSteps:     []string{"Grant policy read access and rerun the scan."},
			Commands:     []string{"kubectl auth can-i list schedules.velero.io --all-namespaces"},
			Caveats:      []string{"Treat the current score as conservative."},
		},
	}

	var buf bytes.Buffer
	buildReport(&buf, &b)
	html := buf.String()

	for _, fragment := range []string{"Why It Matters", "DR Impact", "Validate", "Fix Steps", "Example Commands", "Caveats"} {
		if !strings.Contains(html, fragment) {
			t.Fatalf("buildReport() missing remediation section %q", fragment)
		}
	}
}

func TestSharedOutputCSSUsesSharedThemePalette(t *testing.T) {
	css := sharedOutputCSS()
	for _, want := range []string{"#0d1117", "#161b22", "#58a6ff", "#7ee787", "#f85149"} {
		if !strings.Contains(css, want) {
			t.Fatalf("sharedOutputCSS() missing theme token %s", want)
		}
	}
	for _, old := range []string{"#a78bfa", "#8b5cf6", "#c4b5fd", "rgba(196,181,253", "rgba(190,172,255", "rgba(167,139,250", "rgba(139,92,246"} {
		if strings.Contains(css, old) {
			t.Fatalf("sharedOutputCSS() still contains legacy color %s", old)
		}
	}
}
