package scanapp

import (
	"encoding/json"
	"os"

	"k8s-recovery-visualizer/internal/compare"
	"k8s-recovery-visualizer/internal/model"
)

func applyComparison(bundle *model.Bundle, compareTo string) error {
	if compareTo == "" {
		return nil
	}
	prev, err := loadBundle(compareTo)
	if err != nil {
		return err
	}
	diff := compare.Diff(prev, bundle)
	bundle.Comparison = &diff
	return nil
}

func loadBundle(path string) (*model.Bundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var b model.Bundle
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return &b, nil
}
