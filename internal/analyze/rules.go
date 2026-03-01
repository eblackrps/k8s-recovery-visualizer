package analyze

import (
	"strings"

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

	// Round 11 — resource governance (Workload domain)
	penNoRequests = 15 // pods without CPU+memory requests
	penNoLimits   = 8  // pods without CPU+memory limits

	// Round 12 — pod security (Config domain)
	penPrivileged     = 20 // any pod runs a privileged container
	penHostNetworkPID = 10 // any pod uses hostNetwork or hostPID

	// Round 13 — VolumeSnapshot coverage (Storage domain)
	penNoSnapshot = 10 // PVCs with no matching VolumeSnapshot

	// Round 14 — LimitRange enforcement (Config domain)
	penLRMissing = 10 // namespaces with workloads but no LimitRange

	// Round 14 — PSA label coverage (Config domain)
	penPSAMissing = 12 // non-system namespaces with pods but no pod-security.kubernetes.io/enforce label

	// Round 14 — etcd backup (Backup domain)
	penEtcdNoBackup = 25 // no etcd backup evidence found on self-managed cluster

	// Round 15 — NetworkPolicy coverage (Config domain)
	penNPMissing = 12 // namespaces with pods but no NetworkPolicy

	// Round 16 — Node health + zone topology (Workload domain)
	penNodeNotReady = 20 // one or more nodes are NotReady
	penSingleAZ     = 15 // all nodes share the same availability zone (single-AZ risk)

	// Round 17 — StorageClass DR suitability (Storage domain)
	penSCDeletePolicy = 10 // StorageClass(es) use ReclaimPolicy=Delete — data loss on PVC deletion
	penSCHostPath     = 20 // hostPath provisioner detected — node-local storage, not portable
	penSCZoneUnaware  = 8  // StorageClass provisioner may not be zone-aware

	// Round 18 — ServiceAccount token audit (Config domain)
	penDefaultSAOverPriv = 15 // default ServiceAccount has explicit ClusterRoleBinding
	penAutoMountSA       = 10 // pods automount service account token without need
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

	// Round 11 — resource governance: flag pods missing requests or limits (once per scan).
	var noRequestPods, noLimitPods []string
	for _, pod := range b.Inventory.Pods {
		if pod.Namespace == "kube-system" {
			continue // system pods are exempt
		}
		if !pod.HasRequests {
			noRequestPods = append(noRequestPods, pod.Namespace+"/"+pod.Name)
		}
		if !pod.HasLimits {
			noLimitPods = append(noLimitPods, pod.Namespace+"/"+pod.Name)
		}
	}
	if len(noRequestPods) > 0 {
		workload -= penNoRequests
		addFinding(b, "POD_NO_REQUESTS", "HIGH",
			"pods:"+joinFirst(noRequestPods, 3),
			"Pods running without CPU/memory requests — scheduler cannot make placement guarantees",
			"Set requests on all containers; use LimitRange to enforce namespace defaults")
	}
	if len(noLimitPods) > 0 {
		workload -= penNoLimits
		addFinding(b, "POD_NO_LIMITS", "MEDIUM",
			"pods:"+joinFirst(noLimitPods, 3),
			"Pods running without CPU/memory limits — risk of noisy-neighbour resource exhaustion",
			"Set limits on all containers; use LimitRange to enforce namespace defaults")
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

	// Round 12 — pod security audit (Config domain)
	var privilegedPods, hostNSPods []string
	for _, pod := range b.Inventory.Pods {
		if pod.Namespace == "kube-system" {
			continue // control-plane/CNI pods are expected to use elevated privileges
		}
		if pod.Privileged {
			privilegedPods = append(privilegedPods, pod.Namespace+"/"+pod.Name)
		}
		if pod.HostNetwork || pod.HostPID {
			hostNSPods = append(hostNSPods, pod.Namespace+"/"+pod.Name)
		}
	}
	if len(privilegedPods) > 0 {
		config -= penScale(penPrivileged, wSec)
		addFinding(b, "POD_PRIVILEGED", "CRITICAL",
			"pods:"+joinFirst(privilegedPods, 3),
			"Pods run privileged containers — full host kernel access granted",
			"Remove privileged:true; use specific capabilities (CAP_NET_ADMIN etc.) instead")
	}
	if len(hostNSPods) > 0 {
		config -= penScale(penHostNetworkPID, wSec)
		addFinding(b, "POD_HOST_NAMESPACE", "HIGH",
			"pods:"+joinFirst(hostNSPods, 3),
			"Pods share host network or PID namespace — increases blast radius on node compromise",
			"Set hostNetwork:false and hostPID:false unless explicitly required by the workload")
	}

	// Round 13 — VolumeSnapshot coverage (Storage domain)
	if len(b.Inventory.VolumeSnapshotClasses) == 0 && len(b.Inventory.PVCs) > 0 {
		// No snapshot infrastructure at all
		storage -= penNoSnapshot
		addFinding(b, "SNAPSHOT_NO_CLASS", "MEDIUM", "cluster",
			"No VolumeSnapshotClass found — CSI snapshot capability not configured",
			"Install a CSI driver that supports snapshots and create a VolumeSnapshotClass")
	} else if len(b.Inventory.VolumeSnapshotClasses) > 0 {
		// Snapshot infra present — find PVCs with no snapshot
		snappedPVCs := map[string]bool{}
		for _, vs := range b.Inventory.VolumeSnapshots {
			snappedPVCs[vs.Namespace+"/"+vs.PVCName] = true
		}
		var unsnapshottedPVCs []string
		for _, pvc := range b.Inventory.PVCs {
			key := pvc.Namespace + "/" + pvc.Name
			if !snappedPVCs[key] {
				unsnapshottedPVCs = append(unsnapshottedPVCs, key)
			}
		}
		if len(unsnapshottedPVCs) > 0 {
			storage -= penNoSnapshot
			addFinding(b, "SNAPSHOT_PVC_UNCOVERED", "MEDIUM",
				"pvcs:"+joinFirst(unsnapshottedPVCs, 3),
				"PVCs have no VolumeSnapshot — point-in-time recovery not available for these volumes",
				"Create VolumeSnapshots (or a schedule via the snapshot-controller) for all production PVCs")
		}
	}

	// ── Round 14 — LimitRange enforcement (Config domain) ───────────────────
	// Build set of namespaces that have at least one LimitRange.
	nsWithLR := map[string]bool{}
	for _, lr := range b.Inventory.LimitRanges {
		nsWithLR[lr.Namespace] = true
	}
	// Find namespaces that have pods but no LimitRange.
	nsWithPods := map[string]bool{}
	for _, pod := range b.Inventory.Pods {
		if pod.Namespace == "kube-system" {
			continue
		}
		nsWithPods[pod.Namespace] = true
	}
	var lrMissingNS []string
	for ns := range nsWithPods {
		if !nsWithLR[ns] {
			lrMissingNS = append(lrMissingNS, ns)
		}
	}
	if len(lrMissingNS) > 0 {
		config -= penScale(penLRMissing, wSec)
		addFinding(b, "LR_MISSING_NAMESPACE", "MEDIUM",
			"namespaces:"+joinFirst(lrMissingNS, 3),
			"Namespaces have pods but no LimitRange — unbounded resource consumption possible",
			"Add a LimitRange to each namespace to enforce default CPU/memory requests and limits")
	}

	// ── Round 14 — PSA label coverage (Config domain) ────────────────────────
	var psaMissingNS []string
	for _, ns := range b.Inventory.Namespaces {
		if ns.Name == "kube-system" || ns.Name == "kube-public" || ns.Name == "kube-node-lease" {
			continue // control-plane namespaces — PSA labels not expected
		}
		if !nsWithPods[ns.Name] {
			continue // skip namespaces with no running pods
		}
		if ns.PSAEnforce == "" {
			psaMissingNS = append(psaMissingNS, ns.Name)
		}
	}
	if len(psaMissingNS) > 0 {
		config -= penScale(penPSAMissing, wSec)
		addFinding(b, "PSA_MISSING_ENFORCE_LABEL", "MEDIUM",
			"namespaces:"+joinFirst(psaMissingNS, 3),
			"Namespaces lack pod-security.kubernetes.io/enforce label — PSA admission not enforced",
			"Set pod-security.kubernetes.io/enforce=baseline (or restricted) on each namespace to enable Pod Security Admission")
	}

	// ── Round 14 — etcd backup detection (Backup domain) ─────────────────────
	if eb := b.Inventory.EtcdBackup; eb != nil && !eb.Detected {
		backup -= penScale(penEtcdNoBackup, wSec)
		addFinding(b, "ETCD_BACKUP_MISSING", "HIGH",
			"cluster",
			"No etcd backup evidence found — complete cluster state loss is unrecoverable without etcd",
			"Configure periodic etcd snapshots (e.g. etcdctl snapshot save) via a CronJob, or use a managed K8s service that handles this automatically")
	}

	// ── Round 15 — NetworkPolicy coverage (Config domain) ────────────────────
	nsWithNP := map[string]bool{}
	for _, np := range b.Inventory.NetworkPolicies {
		nsWithNP[np.Namespace] = true
	}
	var npMissingNS []string
	for ns := range nsWithPods {
		if ns == "kube-system" {
			continue
		}
		if !nsWithNP[ns] {
			npMissingNS = append(npMissingNS, ns)
		}
	}
	if len(npMissingNS) > 0 {
		config -= penScale(penNPMissing, wSec)
		addFinding(b, "NETPOL_MISSING_NAMESPACE", "MEDIUM",
			"namespaces:"+joinFirst(npMissingNS, 3),
			"Namespaces have pods but no NetworkPolicy — unrestricted east-west traffic between all pods",
			"Add default-deny NetworkPolicies to each namespace and allow only required traffic paths")
	}

	// ── Round 16 — Node health + zone topology (Workload domain) ─────────────
	var notReadyNodes []string
	zoneSet := map[string]bool{}
	for _, node := range b.Inventory.Nodes {
		if !node.Ready {
			notReadyNodes = append(notReadyNodes, node.Name)
		}
		if z := node.Zone; z != "" {
			zoneSet[z] = true
		}
	}
	if len(notReadyNodes) > 0 {
		workload -= penScale(penNodeNotReady, wRepl)
		addFinding(b, "NODE_NOT_READY", "HIGH",
			"nodes:"+joinFirst(notReadyNodes, 3),
			"One or more nodes are NotReady — workload capacity is reduced and DR failover may be impaired",
			"Investigate node conditions (kubectl describe node) and resolve underlying issues; ensure cluster has sufficient spare capacity")
	}
	if len(b.Inventory.Nodes) > 1 && len(zoneSet) == 1 {
		workload -= penScale(penSingleAZ, wRepl)
		addFinding(b, "SINGLE_AZ_CLUSTER", "MEDIUM",
			"cluster",
			"All nodes reside in a single availability zone — an AZ outage would take down the entire cluster",
			"Distribute nodes across at least 3 availability zones; use topology spread constraints on critical workloads")
	}

	// ── Round 17 — StorageClass DR suitability (Storage domain) ──────────────
	var scDelete, scHostPath, scZoneUnaware []string
	for _, sc := range b.Inventory.StorageClasses {
		if sc.ReclaimPolicy == "Delete" {
			scDelete = append(scDelete, sc.Name)
		}
		p := sc.Provisioner
		if p == "kubernetes.io/host-path" || p == "rancher.io/local-path" ||
			p == "docker.io/hostpath" || p == "microk8s.io/hostpath" {
			scHostPath = append(scHostPath, sc.Name)
		}
		// Zone-unaware if WaitForFirstConsumer not set AND no topology params
		if sc.VolumeBindingMode != "WaitForFirstConsumer" {
			_, hasZone := sc.Parameters["zone"]
			_, hasZones := sc.Parameters["zones"]
			_, hasFSType := sc.Parameters["type"] // some provisioners use type-only
			_ = hasFSType
			if !hasZone && !hasZones {
				scZoneUnaware = append(scZoneUnaware, sc.Name)
			}
		}
	}
	if len(scDelete) > 0 {
		storage -= penScale(penSCDeletePolicy, wImmut)
		addFinding(b, "SC_RECLAIM_DELETE", "MEDIUM",
			"storageclasses:"+joinFirst(scDelete, 3),
			"StorageClass uses ReclaimPolicy=Delete — PV (and data) is destroyed when the PVC is deleted",
			"Change reclaimPolicy to Retain on production StorageClasses to prevent accidental data loss")
	}
	if len(scHostPath) > 0 {
		storage -= penScale(penSCHostPath, wImmut)
		addFinding(b, "SC_HOSTPATH_PROVISIONER", "HIGH",
			"storageclasses:"+joinFirst(scHostPath, 3),
			"StorageClass uses a hostPath provisioner — volumes are node-local and cannot be recovered after node failure",
			"Replace hostPath storage with a network-attached CSI driver (e.g. AWS EBS, Azure Disk, Ceph/Rook) for portable, recoverable storage")
	}
	if len(scZoneUnaware) > 0 && len(zoneSet) > 1 {
		// Only penalise when cluster spans multiple zones — single-AZ clusters get single-AZ finding instead
		storage -= penScale(penSCZoneUnaware, wImmut)
		addFinding(b, "SC_ZONE_UNAWARE", "LOW",
			"storageclasses:"+joinFirst(scZoneUnaware, 3),
			"StorageClass may not be zone-aware in a multi-AZ cluster — PVs could be provisioned in a different AZ than the pod",
			"Set volumeBindingMode: WaitForFirstConsumer so volumes are provisioned in the same zone as the consuming pod")
	}

	// ── Round 18 — ServiceAccount token audit (Config domain) ─────────────────
	// Build set of SA names that have an explicit ClusterRoleBinding (covers default SA abuse).
	saBoundCRBs := map[string]bool{} // key: "namespace/sa-name"
	for _, crb := range b.Inventory.ClusterRoleBindings {
		for _, subj := range crb.Subjects {
			// subjects stored as "ServiceAccount:ns/name" or "ServiceAccount:name"
			if !strings.HasPrefix(subj, "ServiceAccount:") {
				continue
			}
			rest := strings.TrimPrefix(subj, "ServiceAccount:")
			saBoundCRBs[rest] = true
		}
	}
	var defaultSAOverPriv []string
	for _, sa := range b.Inventory.ServiceAccounts {
		if sa.Name != "default" {
			continue
		}
		key := sa.Namespace + "/default"
		if saBoundCRBs[key] || saBoundCRBs["default"] {
			defaultSAOverPriv = append(defaultSAOverPriv, sa.Namespace+"/default")
		}
	}
	if len(defaultSAOverPriv) > 0 {
		config -= penScale(penDefaultSAOverPriv, wSec)
		addFinding(b, "SA_DEFAULT_OVERPRIV", "HIGH",
			"serviceaccounts:"+joinFirst(defaultSAOverPriv, 3),
			"Default ServiceAccount has explicit ClusterRoleBinding — any pod in the namespace inherits elevated cluster permissions",
			"Remove the ClusterRoleBinding from the default SA; create dedicated ServiceAccounts with minimal required permissions")
	}
	var autoMountPods []string
	for _, pod := range b.Inventory.Pods {
		if pod.Namespace == "kube-system" {
			continue
		}
		if pod.AutomountSAToken {
			autoMountPods = append(autoMountPods, pod.Namespace+"/"+pod.Name)
		}
	}
	if len(autoMountPods) > 0 {
		config -= penScale(penAutoMountSA, wSec)
		addFinding(b, "SA_AUTOMOUNT_TOKEN", "MEDIUM",
			"pods:"+joinFirst(autoMountPods, 3),
			"Pods have automountServiceAccountToken enabled — token is mounted even when the pod does not call the Kubernetes API",
			"Set automountServiceAccountToken: false on pods (or their ServiceAccount) that do not need API access")
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
