package scanapp

import (
	"context"
	"testing"

	"k8s-recovery-visualizer/internal/model"
)

func TestBuildCollectorPipelineOmitsSecretsByDefault(t *testing.T) {
	steps := buildCollectorPipeline(context.Background(), nil, nil, &model.Bundle{}, Options{})
	for _, step := range steps {
		if step.Name == "Secrets" {
			t.Fatal("Secrets collector present by default, want opt-in only")
		}
	}
}

func TestBuildCollectorPipelineIncludesSecretsWhenEnabled(t *testing.T) {
	steps := buildCollectorPipeline(context.Background(), nil, nil, &model.Bundle{}, Options{IncludeSecretMetadata: true})
	for _, step := range steps {
		if step.Name == "Secrets" {
			return
		}
	}
	t.Fatal("Secrets collector missing when IncludeSecretMetadata is true")
}
