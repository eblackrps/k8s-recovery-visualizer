package remediation

import "k8s-recovery-visualizer/internal/model"

type knowledgeEntry struct {
	WhyItMatters string
	DRImpact     string
	Validation   []string
	FixSteps     []string
	Caveats      []string
}

var knowledgeBase = map[string]knowledgeEntry{
	"BACKUP_NONE": {
		WhyItMatters: "A DR score without any backup platform is mostly theoretical because there is no automated path to reconstruct workload data and cluster objects.",
		DRImpact:     "A site loss, accidental deletion, or storage failure becomes a rebuild exercise instead of a recovery exercise.",
		Validation: []string{
			"Confirm whether any approved backup platform is installed in the cluster or managed externally.",
			"Verify whether both application data and cluster-scoped configuration are covered.",
		},
		FixSteps: []string{
			"Install a supported backup platform.",
			"Create production backup policies and retention settings.",
			"Run and record a recovery drill before treating the environment as protected.",
		},
		Caveats: []string{
			"External enterprise backup products can exist without leaving inspectable in-cluster evidence. If that is the case, document them and verify coverage manually.",
		},
	},
	"BACKUP_PARTIAL_COVERAGE": {
		WhyItMatters: "Backups that skip stateful namespaces create a false sense of coverage because the control plane can be restored while application data is still missing.",
		DRImpact:     "Namespace restore can succeed for stateless components while databases or message queues remain unrecoverable.",
		Validation: []string{
			"Review namespace selectors in every backup policy.",
			"List stateful workloads and verify each namespace is included by at least one policy.",
		},
		FixSteps: []string{
			"Expand policy selectors or create targeted backup jobs for the uncovered namespaces.",
			"Re-run the scan and a restore simulation after policy changes.",
		},
	},
	"BACKUP_NO_POLICIES": {
		WhyItMatters: "Installed backup software without schedules or policies is not operational protection.",
		DRImpact:     "Recoverability depends on ad hoc operator action instead of repeatable automation.",
		Validation: []string{
			"Check whether scheduled jobs or policies exist in the backup product.",
			"Confirm the policies actually run and persist restore points.",
		},
		FixSteps: []string{
			"Create policies for all production namespaces.",
			"Define retention, schedule, and export settings explicitly.",
		},
	},
	"BACKUP_COVERAGE_UNVERIFIED": {
		WhyItMatters: "Detection alone does not prove policy scope, schedule health, or object recoverability.",
		DRImpact:     "The cluster can look protected on paper while operators still have no evidence that the right namespaces are restorable.",
		Validation: []string{
			"Confirm whether the failure is caused by unsupported tooling, missing permissions, API errors, or parser gaps.",
			"Review the detected product directly and confirm namespace coverage.",
		},
		FixSteps: []string{
			"Grant read access to backup policy resources if the scanner was blocked.",
			"Document manual verification evidence if the tool is not yet supported by the scanner.",
			"Re-run the scan once policy inspection is available.",
		},
		Caveats: []string{
			"An unverified result is intentionally conservative. It should not be overridden without evidence.",
		},
	},
	"BACKUP_NO_OFFSITE": {
		WhyItMatters: "Primary-site-only backups do not protect against region, datacenter, or storage-system loss.",
		DRImpact:     "A site-wide incident can destroy both production data and the only restore source.",
		Validation: []string{
			"Inspect backup export targets and replication jobs.",
			"Confirm the secondary target is reachable and contains recent restore points.",
		},
		FixSteps: []string{
			"Configure an offsite object store, secondary cluster, or replicated target.",
			"Test restore from the secondary location instead of only local snapshots.",
		},
	},
	"BACKUP_RECENT_SUCCESS_MISSING": {
		WhyItMatters: "Policies without recent successful run evidence are not trustworthy during an incident.",
		DRImpact:     "Restore planning relies on stale assumptions instead of observable successful jobs.",
		Validation: []string{
			"Inspect last successful run timestamps in the backup product.",
			"Check failed-job logs and alerting for recurring backup failures.",
		},
		FixSteps: []string{
			"Restore visibility into run history and alerting.",
			"Fix failed jobs before counting the environment as recoverable.",
		},
	},
	"BACKUP_SCHEDULE_STALE": {
		WhyItMatters: "A stale last-success timestamp means the schedule is no longer meeting the declared RPO.",
		DRImpact:     "Actual data-loss exposure is larger than the recovery program expects.",
		Validation: []string{
			"Compare the schedule cadence with the last successful execution time.",
			"Review scheduler, storage, and credentials errors in the backup platform.",
		},
		FixSteps: []string{
			"Repair failed schedules and backfill a successful run.",
			"Reconfirm RPO after the schedule has stabilized.",
		},
	},
	"RESTORE_DEPENDENCY_BLOCKER": {
		WhyItMatters: "Backups are not enough when the target cluster is missing storage, portable volume layouts, or other restore prerequisites.",
		DRImpact:     "Data may exist in backup media but still fail to restore into a runnable workload.",
		Validation: []string{
			"Review restore blockers per namespace in the restore simulation output.",
			"Check for hostPath volumes, unbound PVCs, and missing StorageClasses.",
		},
		FixSteps: []string{
			"Resolve every blocker and rerun the restore simulation.",
			"Prefer a tested restore drill over a paper-only signoff.",
		},
	},
	"PVC_UNBOUND": {
		WhyItMatters: "Unbound claims do not have usable backing storage, so they are already in a failed recovery state.",
		DRImpact:     "Pods cannot mount storage and backup tools cannot protect data that was never provisioned.",
		Validation: []string{
			"Describe the PVC and inspect recent FailedBinding events.",
			"Confirm the referenced StorageClass and capacity are available.",
		},
		FixSteps: []string{
			"Resolve the binding failure.",
			"Verify the claim is Bound before re-running the scan.",
		},
	},
	"PV_HOSTPATH": {
		WhyItMatters: "hostPath volumes are pinned to a node filesystem and are not portable across node or site failures.",
		DRImpact:     "A restore can recover manifests while the underlying data remains stranded on the original node.",
		Validation: []string{
			"List the PV backend type and verify whether it is node-local.",
			"Confirm whether a CSI-backed migration path exists for the workload.",
		},
		FixSteps: []string{
			"Migrate the workload to CSI-backed persistent storage.",
			"Retest recovery after the data is no longer node-local.",
		},
		Caveats: []string{
			"Some platform components intentionally use hostPath. Focus first on application data paths and stateful services.",
		},
	},
	"PV_DELETE_POLICY": {
		WhyItMatters: "Delete reclaim policy increases the chance of permanent data loss during cleanup, migration, or failed restore attempts.",
		DRImpact:     "Operators can unintentionally delete the only remaining copy of a volume while trying to recover it.",
		Validation: []string{
			"Inspect the reclaim policy on production PVs and StorageClasses.",
			"Review storage lifecycle automation that may delete PVCs during maintenance.",
		},
		FixSteps: []string{
			"Switch recoverable data paths to Retain.",
			"Document the manual cleanup process for retained PVs.",
		},
	},
	"PVC_NO_STORAGECLASS": {
		WhyItMatters: "Implicit storage defaults are fragile across clusters and often change between production and recovery environments.",
		DRImpact:     "A restore can recreate PVC objects that never bind because the target cluster has different defaults.",
		Validation: []string{
			"Check whether the PVC relies on a default StorageClass.",
			"Compare the current cluster default with the intended DR target.",
		},
		FixSteps: []string{
			"Set storageClassName explicitly on recoverable PVCs or templates.",
			"Standardize the target StorageClass across source and recovery clusters.",
		},
	},
	"STS_NO_PVC": {
		WhyItMatters: "StatefulSets without persistent claims rely on ephemeral pod filesystems for state.",
		DRImpact:     "A pod restart or full-cluster restore can wipe the application’s only durable data copy.",
		Validation: []string{
			"Inspect the StatefulSet for volumeClaimTemplates and mounted data paths.",
			"Confirm whether the application is actually stateless before downgrading the risk.",
		},
		FixSteps: []string{
			"Add a volumeClaimTemplate for durable state.",
			"Move data out of container filesystems before treating the workload as recoverable.",
		},
	},
	"CRD_NO_BACKUP": {
		WhyItMatters: "Custom resources cannot be restored safely if their CRDs are missing or restored in the wrong order.",
		DRImpact:     "Operators can recover application manifests but fail to recreate the control plane objects they depend on.",
		Validation: []string{
			"List CRDs and confirm whether backup content includes both definitions and instances.",
			"Verify restore order for CRDs before custom resources.",
		},
		FixSteps: []string{
			"Capture CRDs in backup scope.",
			"Document CRD-first restore order in the runbook.",
		},
	},
	"CERT_EXPIRING_SOON": {
		WhyItMatters: "Certificate expiry during or immediately after a DR event turns a successful restore into an availability outage.",
		DRImpact:     "Recovered services can fail TLS handshakes, ingress routing, or workload identity flows.",
		Validation: []string{
			"Check the actual certificate notAfter date and renewal automation health.",
			"Confirm whether the secret or cert-manager issuer will exist in the target environment.",
		},
		FixSteps: []string{
			"Renew the certificate before the DR risk window.",
			"Validate that renewal controllers and issuers are backed up as well.",
		},
	},
	"IMAGE_EXTERNAL_REGISTRY": {
		WhyItMatters: "Recovery often happens in constrained networks where public registries are slow, blocked, or unavailable.",
		DRImpact:     "Pods can be restored but stay Pending or ImagePullBackOff because the target cluster cannot pull the image.",
		Validation: []string{
			"Inventory public image dependencies for critical workloads.",
			"Confirm whether the DR environment can reach those registries.",
		},
		FixSteps: []string{
			"Mirror critical images into a private registry reachable from the recovery environment.",
			"Update manifests and Helm values to use mirrored references.",
		},
	},
	"HELM_UNTRACKED": {
		WhyItMatters: "Helm release state is difficult to reconstruct if values files and overrides are missing.",
		DRImpact:     "Teams can recover charts but still lose environment-specific configuration required for a working reinstall.",
		Validation: []string{
			"Export Helm values for critical releases.",
			"Verify whether release configuration is already stored in Git or another durable system.",
		},
		FixSteps: []string{
			"Back up Helm values and chart references.",
			"Document reinstall order for Helm-managed applications.",
		},
	},
}

func applyKnowledge(step *model.RemediationStep, findingID string) {
	if step == nil {
		return
	}
	entry, ok := knowledgeBase[findingID]
	if !ok {
		return
	}
	step.WhyItMatters = entry.WhyItMatters
	step.DRImpact = entry.DRImpact
	step.Validation = append([]string{}, entry.Validation...)
	step.FixSteps = append([]string{}, entry.FixSteps...)
	step.Caveats = append([]string{}, entry.Caveats...)
}
