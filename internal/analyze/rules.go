package analyze

import (
	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/profile"
)

const (
	storageWeight  = 35
	workloadWeight = 20
	configWeight   = 15
	backupWeight   = 30
)

// Penalty constants — named for readability and auditing.
const (
	penPVCUnbound          = 25
	penPVCNoStorageClass   = 10
	penPVHostPath          = 30
	penPVDeletePolicy      = 15
	penPVOrphan            = 5
	penSTSNoPVC            = 15
	penBackupNone          = 60
	penBackupNoPolicies    = 30
	penBackupPartial       = 20
	penBackupNoOffsite     = 15
	penBackupRPOHigh       = 10 // worst-case RPO > 24 h
	penRestoreSimUncovered = 20 // namespaces with stateful workloads not covered by any policy
	penCRDNoBackup         = 10
	penCertExpiring        = 10
	penImageExternal       = 5
	penHelmUntracked       = 5
	penRBACWildcard        = 20 // custom ClusterRole with wildcard verb
	penRBACEscalate        = 10 // custom ClusterRole with escalate/bind/impersonate
	penRBACSecrets         = 10 // custom ClusterRole with broad secrets read
)

// profileGet returns the weight multiplier for key from a profile weight map.
// Returns 1.0 when the key is absent (no scaling).
func profileGet(weights map[string]float64, key string) float64 {
	if v, ok := weights[key]; ok {
		return v
	}
	return 1.0
}

// penScale multiplies a base penalty by a profile multiplier, rounding to nearest int.
// The result is clamped to a minimum of 1 so a penalty is never fully erased.
func penScale(base int, multiplier float64) int {
	v := int(float64(base)*multiplier + 0.5)
	if v < 1 {
		return 1
	}
	return v
}

func Evaluate(b *model.Bundle) {
	storage := 100
	workload := 100
	config := 100
	backup := 100

	// Resolve profile weights once; missing keys fall back to 1.0 via profileGet.
	p := profile.Normalize(b.Profile)
	weights := profile.Weights(p)
	wImmut := profileGet(weights, "immutability")
	wRepl := profileGet(weights, "replication")
	wRestore := profileGet(weights, "restoreTesting")
	wSec := profileGet(weights, "security")
	wAirgap := profileGet(weights, "airgap")

	pvMap := map[string]model.PersistentVolume{}
	for _, pv := range b.Inventory.PVs {
		pvMap[pv.ClaimRef] = pv
	}

	// ── Storage domain ──────────────────────────────────────────────────────
	for _, pvc := range b.Inventory.PVCs {
		key := pvc.Namespace + "/" + pvc.Name
		pv, bound := pvMap[key]

		if !bound {
			storage -= penPVCUnbound
			addFinding(b, "PVC_UNBOUND", "CRITICAL", key,
				"PVC is not bound to a PV",
				"Investigate binding failure before DR onboarding")
		}
		if pvc.StorageClass == "" {
			storage -= penPVCNoStorageClass
			addFinding(b, "PVC_NO_STORAGECLASS", "HIGH", key,
				"PVC has no storageClass",
				"Define explicit storageClass for DR predictability")
		}
		if bound && pv.Backend == "hostPath" {
			storage -= penScale(penPVHostPath, wImmut)
			addFinding(b, "PV_HOSTPATH", "CRITICAL", pv.Name,
				"PV uses hostPath storage",
				"Migrate to CSI/network storage before DR onboarding")
		}
		if bound && pv.ReclaimPolicy == "Delete" {
			storage -= penScale(penPVDeletePolicy, wImmut)
			addFinding(b, "PV_DELETE_POLICY", "HIGH", pv.Name,
				"PV reclaimPolicy is Delete",
				"Consider Retain for DR recoverability")
		}
	}

	for _, pv := range b.Inventory.PVs {
		if pv.ClaimRef == "" {
			storage -= penPVOrphan
			addFinding(b, "PV_ORPHAN", "MEDIUM", pv.Name,
				"PV is not bound to any PVC",
				"Validate if orphaned storage should be cleaned up")
		}
	}

	// ── Config domain ───────────────────────────────────────────────────────
	// hostPath in kube-system is INFO (control plane/CNI is expected behaviour).
	for _, pod := range b.Inventory.Pods {
		if !pod.UsesHostPath {
			continue
		}
		sev := "CRITICAL"
		rec := "Replace hostPath with CSI-backed persistent storage"
		if pod.Namespace == "kube-system" {
			sev = "INFO"
			rec = "System pod uses hostPath (common for control plane/CNI). Review if acceptable for DR posture."
		}
		addFinding(b, "POD_HOSTPATH", sev, pod.Namespace+"/"+pod.Name,
			"Pod uses hostPath volume", rec)
	}

	// ── Workload domain ─────────────────────────────────────────────────────
	for _, sts := range b.Inventory.StatefulSets {
		if !sts.HasVolumeClaim {
			workload -= penSTSNoPVC
			addFinding(b, "STS_NO_PVC", "HIGH",
				sts.Namespace+"/"+sts.Name,
				"StatefulSet has no volumeClaimTemplate",
				"Stateful workloads should use persistent storage")
		}
	}

	// ── Backup/Recovery domain ──────────────────────────────────────────────
	inv := b.Inventory.Backup
	if inv.PrimaryTool == "none" || inv.PrimaryTool == "" {
		backup -= penBackupNone
		addFinding(b, "BACKUP_NONE", "CRITICAL", "cluster",
			"No backup tool detected in cluster",
			"Install a backup solution (Kasten K10, Velero, Rubrik, Longhorn) before DR onboarding")
	} else {
		// Tool present — check for coverage gaps
		if len(inv.UncoveredStatefulNS) > 0 {
			backup -= penBackupPartial
			addFinding(b, "BACKUP_PARTIAL_COVERAGE", "HIGH",
				"namespaces:"+joinFirst(inv.UncoveredStatefulNS, 3),
				"StatefulSets found in namespaces not covered by backup policy",
				"Extend backup policies to cover all stateful namespaces")
		}
		if len(inv.CoveredNamespaces) == 0 {
			backup -= penScale(penBackupNoPolicies, wRestore)
			addFinding(b, "BACKUP_NO_POLICIES", "HIGH", inv.PrimaryTool,
				"Backup tool detected but no backup policies or schedules found",
				"Create backup schedules covering all production namespaces")
		}
	}

	// Offsite backup check — tool present but no offsite/export policy found.
	if inv.PrimaryTool != "none" && inv.PrimaryTool != "" && !inv.HasOffsite {
		backup -= penScale(penBackupNoOffsite, wRepl)
		addFinding(b, "BACKUP_NO_OFFSITE", "HIGH", inv.PrimaryTool,
			"Backup tool detected but no offsite/export location configured",
			"Configure an offsite or cloud export target to protect against site-level failures")
	}

	// RPO check — flag when worst-case RPO exceeds 24 hours.
	worstRPO := 0
	for _, p := range inv.Policies {
		if p.RPOHours > worstRPO {
			worstRPO = p.RPOHours
		}
	}
	if inv.PrimaryTool != "none" && len(inv.Policies) > 0 && worstRPO > 24 {
		backup -= penBackupRPOHigh
		addFinding(b, "BACKUP_RPO_HIGH", "MEDIUM", inv.PrimaryTool,
			"Backup schedule results in RPO exposure greater than 24 hours",
			"Increase backup frequency to reduce potential data loss window")
	}

	// Restore simulation — penalise when stateful namespaces have no coverage.
	if sim := b.Inventory.Backup.RestoreSim; sim != nil && len(sim.UncoveredNS) > 0 {
		backup -= penScale(penRestoreSimUncovered, wRestore)
		addFinding(b, "RESTORE_SIM_UNCOVERED", "HIGH",
			"namespaces:"+joinFirst(sim.UncoveredNS, 3),
			"Restore simulation: stateful namespaces have no backup policy coverage",
			"Add backup policies covering all namespaces with PVCs or StatefulSets")
	}

	// CRDs present with no backup = extra risk
	if len(b.Inventory.CRDs) > 0 && (inv.PrimaryTool == "none" || inv.PrimaryTool == "") {
		backup -= penCRDNoBackup
		addFinding(b, "CRD_NO_BACKUP", "MEDIUM", "crds",
			"Custom Resource Definitions present but no backup tool detected",
			"Ensure backup solution captures CRD definitions and CR data")
	}

	// Certificates expiring within 30 days
	for _, cert := range b.Inventory.Certificates {
		if cert.DaysToExpiry >= 0 && cert.DaysToExpiry <= 30 {
			backup -= penScale(penCertExpiring, wSec)
			addFinding(b, "CERT_EXPIRING_SOON", "HIGH",
				cert.Namespace+"/"+cert.Name,
				"Certificate expires within 30 days",
				"Renew certificate before DR event window")
			break // penalise once per scan
		}
	}

	// External public-registry images
	externalCount := 0
	for _, img := range b.Inventory.Images {
		if img.IsPublic {
			externalCount++
		}
	}
	if externalCount > 0 {
		backup -= penScale(penImageExternal, wAirgap)
		addFinding(b, "IMAGE_EXTERNAL_REGISTRY", "MEDIUM", "images",
			"Workloads depend on public container registries",
			"Mirror critical images to a private registry accessible from the DR environment")
	}

	// Helm releases present (flag for values backup)
	if len(b.Inventory.HelmReleases) > 0 && (inv.PrimaryTool == "none" || inv.PrimaryTool == "") {
		backup -= penHelmUntracked
		addFinding(b, "HELM_UNTRACKED", "LOW", "helm",
			"Helm releases detected with no backup tool to capture release values",
			"Back up Helm values (helm get values <release>) for each release before DR")
	}

	// ── Config domain — RBAC privilege audit ────────────────────────────────
	var wildRoles, escalateRoles, secretRoles []string
	for _, cr := range b.Inventory.ClusterRoles {
		if !cr.Custom {
			continue // skip built-in system: roles
		}
		if cr.HasWildcardVerb {
			wildRoles = append(wildRoles, cr.Name)
		}
		if cr.HasEscalatePriv {
			escalateRoles = append(escalateRoles, cr.Name)
		}
		if cr.HasSecretAccess {
			secretRoles = append(secretRoles, cr.Name)
		}
	}
	if len(wildRoles) > 0 {
		config -= penScale(penRBACWildcard, wSec)
		addFinding(b, "RBAC_WILDCARD_VERB", "CRITICAL",
			"roles:"+joinFirst(wildRoles, 3),
			"Custom ClusterRole grants wildcard verb permissions",
			"Scope roles to specific resources and verbs; wildcard permissions are equivalent to cluster-admin")
	}
	if len(escalateRoles) > 0 {
		config -= penScale(penRBACEscalate, wSec)
		addFinding(b, "RBAC_ESCALATE_PRIV", "HIGH",
			"roles:"+joinFirst(escalateRoles, 3),
			"Custom ClusterRole grants escalate, bind, or impersonate verbs",
			"Remove privilege escalation verbs unless explicitly required by the workload")
	}
	if len(secretRoles) > 0 {
		config -= penScale(penRBACSecrets, wSec)
		addFinding(b, "RBAC_SECRET_ACCESS", "HIGH",
			"roles:"+joinFirst(secretRoles, 3),
			"Custom ClusterRole grants broad read access to Secrets",
			"Restrict secret access to specific secrets by name; prefer Role/RoleBinding scoped to a namespace")
	}

	storage = clamp(storage)
	workload = clamp(workload)
	config = clamp(config)
	backup = clamp(backup)

	overall := weightedOverall(storage, workload, config, backup)

	b.Score.Storage.Final = storage
	b.Score.Workload.Final = workload
	b.Score.Config.Final = config
	b.Score.Backup.Final = backup
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

func weightedOverall(storage, workload, config, backup int) int {
	// Integer math, deterministic:
	// overall = round((S*35 + W*20 + C*15 + B*30) / 100)
	sum := storage*storageWeight + workload*workloadWeight + config*configWeight + backup*backupWeight
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

func joinFirst(ss []string, max int) string {
	if len(ss) <= max {
		result := ""
		for i, s := range ss {
			if i > 0 {
				result += ","
			}
			result += s
		}
		return result
	}
	result := ""
	for i := 0; i < max; i++ {
		if i > 0 {
			result += ","
		}
		result += ss[i]
	}
	return result + "..."
}
