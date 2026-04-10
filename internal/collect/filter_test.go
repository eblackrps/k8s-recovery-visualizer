package collect

import (
	"reflect"
	"testing"

	"k8s-recovery-visualizer/internal/model"
)

func TestScopeNamespaces(t *testing.T) {
	b := &model.Bundle{}
	if got := ScopeNamespaces(b); !reflect.DeepEqual(got, []string{""}) {
		t.Fatalf("ScopeNamespaces() = %v, want [\"\"]", got)
	}

	b.ScanNamespaces = []string{"prod", "staging"}
	if got := ScopeNamespaces(b); !reflect.DeepEqual(got, []string{"prod", "staging"}) {
		t.Fatalf("ScopeNamespaces() = %v, want [prod staging]", got)
	}
}

func TestInScope(t *testing.T) {
	b := &model.Bundle{}
	if !InScope("prod", b) {
		t.Fatal("InScope() = false, want true when scan scope is empty")
	}

	b.ScanNamespaces = []string{"prod"}
	if !InScope("prod", b) {
		t.Fatal("InScope(prod) = false, want true")
	}
	if InScope("staging", b) {
		t.Fatal("InScope(staging) = true, want false")
	}
}
