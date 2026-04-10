package backup

import (
	"testing"

	"k8s-recovery-visualizer/internal/model"
)

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
