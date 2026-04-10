package analyze

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/restore"
)

type goldenExpectation struct {
	Profile      string         `json:"profile"`
	Findings     []string       `json:"findings"`
	DomainScores map[string]int `json:"domainScores"`
	DomainDeltas map[string]int `json:"domainDeltas"`
	Overall      int            `json:"overall"`
	Maturity     string         `json:"maturity"`
}

func TestGoldenRuleScoringScenarios(t *testing.T) {
	scenarioDirs, err := filepath.Glob(filepath.Join("testdata", "golden", "*"))
	if err != nil {
		t.Fatalf("Glob(testdata) error = %v", err)
	}
	if len(scenarioDirs) == 0 {
		t.Fatal("no golden scenarios found")
	}

	for _, dir := range scenarioDirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			bundle := goldenBaseBundle()

			bundleRaw, err := os.ReadFile(filepath.Join(dir, "bundle.json"))
			if err != nil {
				t.Fatalf("ReadFile(bundle.json) error = %v", err)
			}
			if err := json.Unmarshal(bundleRaw, &bundle); err != nil {
				t.Fatalf("Unmarshal(bundle.json) error = %v", err)
			}

			var expected goldenExpectation
			expectedRaw, err := os.ReadFile(filepath.Join(dir, "expected.json"))
			if err != nil {
				t.Fatalf("ReadFile(expected.json) error = %v", err)
			}
			if err := json.Unmarshal(expectedRaw, &expected); err != nil {
				t.Fatalf("Unmarshal(expected.json) error = %v", err)
			}
			if expected.Profile != "" {
				bundle.Profile = expected.Profile
			}

			sim := restore.Simulate(&bundle)
			bundle.Inventory.Backup.RestoreSim = &sim
			Evaluate(&bundle)

			gotFindings := findingIDs(bundle.Inventory.Findings)
			wantFindings := append([]string{}, expected.Findings...)
			sort.Strings(wantFindings)
			if !reflect.DeepEqual(gotFindings, wantFindings) {
				t.Fatalf("findings = %v, want %v", gotFindings, wantFindings)
			}

			gotScores := map[string]int{
				"storage":  bundle.Score.Storage.Final,
				"workload": bundle.Score.Workload.Final,
				"config":   bundle.Score.Config.Final,
				"backup":   bundle.Score.Backup.Final,
			}
			if !reflect.DeepEqual(gotScores, expected.DomainScores) {
				t.Fatalf("domain scores = %v, want %v", gotScores, expected.DomainScores)
			}

			gotDeltas := map[string]int{
				"storage":  100 - bundle.Score.Storage.Final,
				"workload": 100 - bundle.Score.Workload.Final,
				"config":   100 - bundle.Score.Config.Final,
				"backup":   100 - bundle.Score.Backup.Final,
			}
			if !reflect.DeepEqual(gotDeltas, expected.DomainDeltas) {
				t.Fatalf("domain deltas = %v, want %v", gotDeltas, expected.DomainDeltas)
			}

			if bundle.Score.Overall.Final != expected.Overall {
				t.Fatalf("overall = %d, want %d", bundle.Score.Overall.Final, expected.Overall)
			}
			if bundle.Score.Maturity != expected.Maturity {
				t.Fatalf("maturity = %s, want %s", bundle.Score.Maturity, expected.Maturity)
			}
		})
	}
}

func goldenBaseBundle() model.Bundle {
	b := model.NewBundle("golden-base", time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC))
	b.Profile = "standard"
	b.Target = "vm"
	b.Cluster.Platform = model.Platform{
		Provider:   "EKS",
		K8sVersion: "1.30.0",
		ClusterUID: "cluster-healthy",
	}
	b.Inventory.Namespaces = []model.Namespace{
		{Name: "prod", PSAEnforce: "baseline"},
	}
	b.Inventory.Nodes = []model.Node{
		{Name: "node-a", Ready: true, Zone: "us-east-1a"},
		{Name: "node-b", Ready: true, Zone: "us-east-1b"},
	}
	b.Inventory.Pods = []model.Pod{
		{
			Namespace:      "prod",
			Name:           "api-0",
			UsesHostPath:   false,
			ContainerCount: 1,
			HasRequests:    true,
			HasLimits:      true,
		},
	}
	b.Inventory.PVCs = []model.PersistentVolumeClaim{
		{
			ID:            "pvc:prod/db-data",
			Name:          "db-data",
			Namespace:     "prod",
			StorageClass:  "csi-fast",
			RequestedSize: "10Gi",
		},
	}
	b.Inventory.PVs = []model.PersistentVolume{
		{
			Name:          "pv-db-data",
			StorageClass:  "csi-fast",
			Capacity:      "10Gi",
			ReclaimPolicy: "Retain",
			Backend:       "csi",
			ClaimRef:      "prod/db-data",
		},
	}
	b.Inventory.StatefulSets = []model.StatefulSet{
		{
			Namespace:      "prod",
			Name:           "db",
			Replicas:       1,
			HasVolumeClaim: true,
		},
	}
	allowExpansion := true
	b.Inventory.StorageClasses = []model.StorageClass{
		{
			Name:                 "csi-fast",
			Provisioner:          "ebs.csi.aws.com",
			ReclaimPolicy:        "Retain",
			VolumeBindingMode:    "WaitForFirstConsumer",
			AllowVolumeExpansion: &allowExpansion,
		},
	}
	b.Inventory.NetworkPolicies = []model.NetworkPolicy{
		{Name: "default-deny", Namespace: "prod", HasIngress: true, HasEgress: true},
	}
	b.Inventory.LimitRanges = []model.LimitRange{
		{Name: "defaults", Namespace: "prod"},
	}
	b.Inventory.Images = []model.ContainerImage{
		{Image: "registry.internal.example/app:1.0.0", Registry: "registry.internal.example", IsPublic: false},
	}
	b.Inventory.VolumeSnapshotClasses = []model.VolumeSnapshotClass{
		{Name: "csi-snap", Driver: "ebs.csi.aws.com", DeletionPolicy: "Retain"},
	}
	b.Inventory.VolumeSnapshots = []model.VolumeSnapshot{
		{Name: "db-data-snap", Namespace: "prod", PVCName: "db-data", ClassName: "csi-snap", ReadyToUse: true},
	}
	b.Inventory.EtcdBackup = &model.EtcdBackupEvidence{
		Detected: true,
		Source:   "provider-managed",
		Detail:   "Managed control plane snapshot protection.",
	}
	b.Inventory.Backup = model.BackupInventory{
		PrimaryTool:      "velero",
		CoverageVerified: true,
		CoverageStatus:   model.BackupCoverageStatusVerified,
		CoverageReason:   "Policy coverage verified from velero.",
		CoveredNamespaces: []string{
			"prod",
		},
		HasOffsite: true,
		Tools: []model.BackupDetectedTool{
			{
				Name:                   "velero",
				Detected:               true,
				Namespace:              "velero",
				PolicyInspectionStatus: model.BackupCoverageStatusVerified,
				PolicyInspectionDetail: "Parsed 1 velero schedule.",
			},
		},
		Policies: []model.BackupPolicy{
			{
				Tool:                "velero",
				Name:                "prod-half-daily",
				PolicyNamespace:     "velero",
				IncludedNS:          []string{"prod"},
				Schedule:            "0 */12 * * *",
				RPOHours:            12,
				HasOffsite:          true,
				StorageLocation:     "dr-object-store",
				LastSuccessAt:       "2026-04-10T06:00:00Z",
				LastSuccessAgeHours: 6,
				FreshSchedule:       true,
				Confidence:          model.EvidenceConfidenceConfirmed,
			},
		},
	}
	return b
}

func findingIDs(findings []model.Finding) []string {
	out := make([]string, 0, len(findings))
	for _, finding := range findings {
		out = append(out, finding.ID)
	}
	sort.Strings(out)
	return out
}
