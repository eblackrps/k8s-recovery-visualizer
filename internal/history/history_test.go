package history

import (
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
)

func TestLoadDashboardBuildsTrendStats(t *testing.T) {
	outDir := t.TempDir()

	first := model.NewBundle("first", time.Now().UTC())
	first.Score.Overall.Final = 74
	first.Score.Storage.Final = 70
	first.Score.Workload.Final = 72
	first.Score.Config.Final = 76
	first.Score.Backup.Final = 78
	first.Score.Maturity = "SILVER"
	first.Inventory.Findings = []model.Finding{{ID: "A"}}
	if _, err := Record(outDir, &first); err != nil {
		t.Fatalf("Record(first) error = %v", err)
	}

	second := model.NewBundle("second", time.Now().UTC())
	second.Score.Overall.Final = 82
	second.Score.Storage.Final = 84
	second.Score.Workload.Final = 80
	second.Score.Config.Final = 81
	second.Score.Backup.Final = 83
	second.Score.Maturity = "GOLD"
	second.Inventory.Findings = []model.Finding{{ID: "A"}, {ID: "B"}}
	if _, err := Record(outDir, &second); err != nil {
		t.Fatalf("Record(second) error = %v", err)
	}

	dashboard := LoadDashboard(outDir, 10)
	if dashboard.RunCount != 2 {
		t.Fatalf("RunCount = %d, want 2", dashboard.RunCount)
	}
	if dashboard.TrendLabel != "IMPROVING" || dashboard.TrendDelta != 8 {
		t.Fatalf("trend = %s/%d, want IMPROVING/8", dashboard.TrendLabel, dashboard.TrendDelta)
	}
	if dashboard.AverageScore != 78 || dashboard.BestScore != 82 || dashboard.WorstScore != 74 {
		t.Fatalf("stats = %+v", dashboard)
	}
	if len(dashboard.DomainTrends) != 4 {
		t.Fatalf("DomainTrends len = %d, want 4", len(dashboard.DomainTrends))
	}
	if dashboard.Entries[1].Backup != 83 || dashboard.Entries[1].Findings != 2 {
		t.Fatalf("entries = %+v", dashboard.Entries[1])
	}
}
