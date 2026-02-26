package analyze

import "k8s-recovery-visualizer/internal/model"

const (
	storageWeight  = 50
	workloadWeight = 30
	configWeight   = 20
)

func Evaluate(b *model.Bundle) {

	storage := 100
	workload := 100
	config := 100

	pvMap := map[string]model.PersistentVolume{}
	for _, pv := range b.Inventory.PVs {
		pvMap[pv.ClaimRef] = pv
	}

	// PVC rules (Storage)
	for _, pvc := range b.Inventory.PVCs {

		key := pvc.Namespace + "/" + pvc.Name
		pv, bound := pvMap[key]

		if !bound {
			storage -= 25
			addFinding(b, "PVC_UNBOUND", "CRITICAL", key,
				"PVC is not bound to a PV",
				"Investigate binding failure before DR onboarding")
		}

		if pvc.StorageClass == "" {
			storage -= 10
			addFinding(b, "PVC_NO_STORAGECLASS", "HIGH", key,
				"PVC has no storageClass",
				"Define explicit storageClass for DR predictability")
		}

		if bound && pv.Backend == "hostPath" {
			storage -= 30
			addFinding(b, "PV_HOSTPATH", "CRITICAL", pv.Name,
				"PV uses hostPath storage",
				"Migrate to CSI/network storage before DR onboarding")
		}

		if bound && pv.ReclaimPolicy == "Delete" {
			storage -= 15
			addFinding(b, "PV_DELETE_POLICY", "HIGH", pv.Name,
				"PV reclaimPolicy is Delete",
				"Consider Retain for DR recoverability")
		}
	}

	// PV orphan (Storage)
	for _, pv := range b.Inventory.PVs {
		if pv.ClaimRef == "" {
			storage -= 5
			addFinding(b, "PV_ORPHAN", "MEDIUM", pv.Name,
				"PV is not bound to any PVC",
				"Validate if orphaned storage should be cleaned up")
		}
	}

	// Pod hostPath (Config)
	for _, pod := range b.Inventory.Pods {
		if pod.UsesHostPath {
			config -= 20
			addFinding(b, "POD_HOSTPATH", "CRITICAL",
				pod.Namespace+"/"+pod.Name,
				"Pod uses hostPath volume",
				"Replace hostPath with CSI-backed persistent storage")
		}
	}

	// StatefulSet persistence (Workload)
	for _, sts := range b.Inventory.StatefulSets {
		if !sts.HasVolumeClaim {
			workload -= 15
			addFinding(b, "STS_NO_PVC", "HIGH",
				sts.Namespace+"/"+sts.Name,
				"StatefulSet has no volumeClaimTemplate",
				"Stateful workloads should use persistent storage")
		}
	}

	storage = clamp(storage)
	workload = clamp(workload)
	config = clamp(config)

	overall := weightedOverall(storage, workload, config)

	b.Score.Storage.Final = storage
	b.Score.Workload.Final = workload
	b.Score.Config.Final = config
	b.Score.Overall.Final = overall

	if overall >= 90 {
		b.Score.Maturity = "PLATINUM"
	} else if overall >= 75 {
		b.Score.Maturity = "GOLD"
	} else if overall >= 50 {
		b.Score.Maturity = "SILVER"
	} else {
		b.Score.Maturity = "BRONZE"
	}
}

func weightedOverall(storage, workload, config int) int {
	// Integer math, deterministic:
	// overall = round((S*50 + W*30 + C*20)/100)
	sum := storage*storageWeight + workload*workloadWeight + config*configWeight
	return (sum + 50) / 100
}

func addFinding(b *model.Bundle, id, severity, resource, message, recommendation string) {
	b.Inventory.Findings = append(b.Inventory.Findings, model.Finding{
		ID:             id,
		Severity:       severity,
		ResourceID:     resource,
		Message:        message,
		Recommendation: recommendation,
	})
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
