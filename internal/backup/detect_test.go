package backup

import (
	"context"
	"strings"
	"testing"

	"k8s-recovery-visualizer/internal/model"
)

func TestSignatureCollectorDefaultsUnsupportedInspection(t *testing.T) {
	collector := signatureCollector{spec: toolSpec{Name: "rubrik"}}
	got := collector.inspect(context.Background(), nil)
	if got.Status != model.BackupCoverageStatusUnsupported {
		t.Fatalf("inspect() status = %q, want unsupported", got.Status)
	}
	if !strings.Contains(got.Reason, "does not yet inspect") {
		t.Fatalf("inspect() reason = %q, want unsupported guidance", got.Reason)
	}
}

func TestCoveredNamespacesFromPolicies(t *testing.T) {
	b := model.Bundle{
		Inventory: model.Inventory{
			Namespaces: []model.Namespace{
				{Name: "prod"},
				{Name: "staging"},
				{Name: "dev"},
			},
		},
	}
	policies := []model.BackupPolicy{
		{
			IncludedNS: []string{"*"},
			ExcludedNS: []string{"dev"},
		},
	}

	got := coveredNamespacesFromPolicies(&b, policies)
	want := []string{"prod", "staging"}

	if len(got) != len(want) {
		t.Fatalf("coveredNamespacesFromPolicies() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("coveredNamespacesFromPolicies()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestUncoveredStatefulNamespacesFromPolicies(t *testing.T) {
	b := model.Bundle{
		Inventory: model.Inventory{
			StatefulSets: []model.StatefulSet{
				{Namespace: "prod", Name: "db"},
				{Namespace: "staging", Name: "queue"},
				{Namespace: "staging", Name: "cache"},
			},
		},
	}
	policies := []model.BackupPolicy{
		{
			IncludedNS: []string{"prod"},
		},
	}

	got := uncoveredStatefulNamespacesFromPolicies(&b, policies)
	if len(got) != 1 || got[0] != "staging" {
		t.Fatalf("uncoveredStatefulNamespacesFromPolicies() = %v, want [staging]", got)
	}
}

func TestPolicyCoversNamespaceHonorsExclusions(t *testing.T) {
	policy := model.BackupPolicy{
		IncludedNS: []string{"*"},
		ExcludedNS: []string{"kube-system"},
	}

	if !policyCoversNamespace(policy, "prod") {
		t.Fatal("policyCoversNamespace() = false for included namespace, want true")
	}
	if policyCoversNamespace(policy, "kube-system") {
		t.Fatal("policyCoversNamespace() = true for excluded namespace, want false")
	}
}

func TestApplyInspectionResultsPrefersVerifiedCoverageSource(t *testing.T) {
	b := model.Bundle{
		Inventory: model.Inventory{
			Namespaces: []model.Namespace{
				{Name: "prod"},
				{Name: "staging"},
			},
			StatefulSets: []model.StatefulSet{
				{Namespace: "prod", Name: "db"},
			},
		},
	}
	inv := model.BackupInventory{
		PrimaryTool: "none",
		Tools: []model.BackupDetectedTool{
			{Name: "rubrik", Detected: true},
			{Name: "longhorn", Detected: true},
		},
	}

	applyInspectionResults(&b, &inv, []rankedInspection{
		{
			index: 0,
			result: inspectionResult{
				Tool:   "rubrik",
				Status: model.BackupCoverageStatusUnsupported,
				Reason: "rubrik was detected, but this scanner does not yet inspect its policies or schedules.",
			},
		},
		{
			index: 1,
			result: inspectionResult{
				Tool:   "longhorn",
				Status: model.BackupCoverageStatusVerified,
				Reason: "Parsed 1 longhorn recurring job.",
				Policies: []model.BackupPolicy{
					{Tool: "longhorn", Name: "daily", IncludedNS: []string{"prod"}, HasOffsite: true},
				},
			},
		},
	})

	if inv.PrimaryTool != "longhorn" {
		t.Fatalf("PrimaryTool = %q, want longhorn", inv.PrimaryTool)
	}
	if !inv.CoverageVerified {
		t.Fatal("CoverageVerified = false, want true")
	}
	if inv.CoverageStatus != model.BackupCoverageStatusVerified {
		t.Fatalf("CoverageStatus = %q, want %q", inv.CoverageStatus, model.BackupCoverageStatusVerified)
	}
	if len(inv.Policies) != 1 {
		t.Fatalf("Policies len = %d, want 1", len(inv.Policies))
	}
	if len(inv.CoveredNamespaces) != 1 || inv.CoveredNamespaces[0] != "prod" {
		t.Fatalf("CoveredNamespaces = %v, want [prod]", inv.CoveredNamespaces)
	}
	if len(inv.CoverageSourceTools) != 1 || inv.CoverageSourceTools[0] != "longhorn" {
		t.Fatalf("CoverageSourceTools = %v, want [longhorn]", inv.CoverageSourceTools)
	}
	if !inv.HasOffsite {
		t.Fatal("HasOffsite = false, want true when all covered namespaces have offsite evidence")
	}
}

func TestApplyInspectionResultsSurfacesPermissionDeniedReason(t *testing.T) {
	inv := model.BackupInventory{
		PrimaryTool: "none",
		Tools: []model.BackupDetectedTool{
			{Name: "velero", Detected: true},
		},
	}

	applyInspectionResults(&model.Bundle{}, &inv, []rankedInspection{
		{
			index: 0,
			result: inspectionResult{
				Tool:   "velero",
				Status: model.BackupCoverageStatusPermissionDenied,
				Reason: "Unable to inspect Velero schedules: schedules.velero.io is forbidden",
			},
		},
	})

	if inv.PrimaryTool != "velero" {
		t.Fatalf("PrimaryTool = %q, want velero", inv.PrimaryTool)
	}
	if inv.CoverageStatus != model.BackupCoverageStatusPermissionDenied {
		t.Fatalf("CoverageStatus = %q, want %q", inv.CoverageStatus, model.BackupCoverageStatusPermissionDenied)
	}
	if !strings.Contains(inv.CoverageReason, "forbidden") {
		t.Fatalf("CoverageReason = %q, want permission detail", inv.CoverageReason)
	}
	if inv.CoverageVerified {
		t.Fatal("CoverageVerified = true, want false")
	}
}

func TestApplyInspectionResultsRequiresOffsiteForEveryCoveredNamespace(t *testing.T) {
	b := model.Bundle{
		Inventory: model.Inventory{
			Namespaces: []model.Namespace{
				{Name: "prod"},
				{Name: "staging"},
			},
		},
	}
	inv := model.BackupInventory{
		PrimaryTool: "none",
		Tools: []model.BackupDetectedTool{
			{Name: "velero", Detected: true},
		},
	}

	applyInspectionResults(&b, &inv, []rankedInspection{
		{
			index: 0,
			result: inspectionResult{
				Tool:   "velero",
				Status: model.BackupCoverageStatusVerified,
				Reason: "Parsed 2 velero schedules.",
				Policies: []model.BackupPolicy{
					{Tool: "velero", Name: "prod-offsite", IncludedNS: []string{"prod"}, HasOffsite: true},
					{Tool: "velero", Name: "staging-local", IncludedNS: []string{"staging"}, HasOffsite: false},
				},
			},
		},
	})

	if inv.HasOffsite {
		t.Fatal("HasOffsite = true, want false when any covered namespace lacks offsite coverage")
	}
	if len(inv.OffsiteCoveredNS) != 1 || inv.OffsiteCoveredNS[0] != "prod" {
		t.Fatalf("OffsiteCoveredNS = %v, want [prod]", inv.OffsiteCoveredNS)
	}
	if len(inv.OffsiteMissingNS) != 1 || inv.OffsiteMissingNS[0] != "staging" {
		t.Fatalf("OffsiteMissingNS = %v, want [staging]", inv.OffsiteMissingNS)
	}
}
