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
