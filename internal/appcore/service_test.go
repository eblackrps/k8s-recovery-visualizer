package appcore

import (
	"path/filepath"
	"testing"
)

func TestListProjectsReturnsEmptyForMissingRoot(t *testing.T) {
	service := NewService()
	root := filepath.Join(t.TempDir(), "missing-projects-root")

	projects, err := service.ListProjects(root)
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("len(ListProjects()) = %d, want 0", len(projects))
	}
}
