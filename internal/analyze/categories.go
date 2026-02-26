package analyze

import "k8s-recovery-visualizer/internal/model"

func BuildCategories(b *model.Bundle) []model.CategoryScore {
	// Explicit and stable. No reflection. Uses Score domain Max/Final.
	cats := []model.CategoryScore{
		{
			Name:     "storage",
			Raw:      float64(b.Score.Storage.Final),
			Weight:   1,
			Weighted: float64(b.Score.Storage.Final),
			Max:      float64(b.Score.Storage.Max),
			Grade:    "",
		},
		{
			Name:     "workload",
			Raw:      float64(b.Score.Workload.Final),
			Weight:   1,
			Weighted: float64(b.Score.Workload.Final),
			Max:      float64(b.Score.Workload.Max),
			Grade:    "",
		},
		{
			Name:     "config",
			Raw:      float64(b.Score.Config.Final),
			Weight:   1,
			Weighted: float64(b.Score.Config.Final),
			Max:      float64(b.Score.Config.Max),
			Grade:    "",
		},
		{
			Name:     "overall",
			Raw:      float64(b.Score.Overall.Final),
			Weight:   1,
			Weighted: float64(b.Score.Overall.Final),
			Max:      float64(b.Score.Overall.Max),
			Grade:    "",
		},
	}

	// Guarantee JSON array output even if something is missing.
	if cats == nil {
		return []model.CategoryScore{}
	}
	return cats
}
