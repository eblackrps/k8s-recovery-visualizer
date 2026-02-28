package analyze

import "k8s-recovery-visualizer/internal/model"

func BuildCategories(b *model.Bundle) []model.CategoryScore {
	cats := []model.CategoryScore{
		{
			Name:     "storage",
			Raw:      float64(b.Score.Storage.Final),
			Weight:   float64(storageWeight),
			Weighted: float64(b.Score.Storage.Final * storageWeight / 100),
			Max:      float64(b.Score.Storage.Max),
		},
		{
			Name:     "workload",
			Raw:      float64(b.Score.Workload.Final),
			Weight:   float64(workloadWeight),
			Weighted: float64(b.Score.Workload.Final * workloadWeight / 100),
			Max:      float64(b.Score.Workload.Max),
		},
		{
			Name:     "config",
			Raw:      float64(b.Score.Config.Final),
			Weight:   float64(configWeight),
			Weighted: float64(b.Score.Config.Final * configWeight / 100),
			Max:      float64(b.Score.Config.Max),
		},
		{
			Name:     "backup",
			Raw:      float64(b.Score.Backup.Final),
			Weight:   float64(backupWeight),
			Weighted: float64(b.Score.Backup.Final * backupWeight / 100),
			Max:      float64(b.Score.Backup.Max),
		},
		{
			Name:     "overall",
			Raw:      float64(b.Score.Overall.Final),
			Weight:   1,
			Weighted: float64(b.Score.Overall.Final),
			Max:      float64(b.Score.Overall.Max),
		},
	}
	if cats == nil {
		return []model.CategoryScore{}
	}
	return cats
}
