package model

import "time"

type Bundle struct {
	Checks        []Check          `json:"checks,omitempty"`
	SchemaVersion string           `json:"schemaVersion"`
	Metadata      Metadata         `json:"metadata"`
	Tool          Tool             `json:"tool"`
	Scan          Scan             `json:"scan"`
	Cluster       Cluster          `json:"cluster"`
	Inventory     Inventory        `json:"inventory"`
	Score         Score            `json:"score"`
	// Target is the declared recovery destination: "baremetal" or "vm"
	Target        string           `json:"target,omitempty"`
	// Profile is the scoring profile used for this scan: standard|enterprise|dev|airgap
	Profile       string           `json:"profile,omitempty"`
	// CollectorSkips records collectors that were skipped due to RBAC or missing APIs.
	CollectorSkips []CollectorSkip `json:"collectorSkips,omitempty"`
	// ScanNamespaces restricts the scan to specific namespaces. Empty = all namespaces.
	ScanNamespaces []string `json:"scanNamespaces,omitempty"`
	// Comparison holds the diff against a previous scan when --compare is used.
	Comparison *ComparisonSummary `json:"comparison,omitempty"`
	// TrendHistory holds the last N scan scores for sparkline rendering in the report.
	TrendHistory []TrendPoint `json:"trendHistory,omitempty"`
}

// TrendPoint is a single data point for the score trend sparkline.
type TrendPoint struct {
	TimestampUTC string `json:"ts"`
	Overall      int    `json:"score"`
	Maturity     string `json:"maturity"`
}

// ComparisonSummary is a lightweight reference stored in the bundle.
// The full diff lives in internal/compare but we embed a summary here
// so the report generator can access it without a circular import.
type ComparisonSummary struct {
	PreviousScanID    string `json:"previousScanId"`
	PreviousScannedAt string `json:"previousScannedAt"`
	PreviousScore     int    `json:"previousScore"`
	PreviousMaturity  string `json:"previousMaturity"`
	ScoreDelta        int    `json:"scoreDelta"`

	NamespacesAdded   []string `json:"namespacesAdded,omitempty"`
	NamespacesRemoved []string `json:"namespacesRemoved,omitempty"`
	WorkloadsAdded    []string `json:"workloadsAdded,omitempty"`
	WorkloadsRemoved  []string `json:"workloadsRemoved,omitempty"`
	PVCsAdded         []string `json:"pvcsAdded,omitempty"`
	PVCsRemoved       []string `json:"pvcsRemoved,omitempty"`
	ImagesAdded       []string `json:"imagesAdded,omitempty"`
	ImagesRemoved     []string `json:"imagesRemoved,omitempty"`

	BackupToolPrevious string `json:"backupToolPrevious"`
	BackupToolCurrent  string `json:"backupToolCurrent"`
	BackupToolChanged  bool   `json:"backupToolChanged"`

	FindingsNew      []Finding `json:"findingsNew,omitempty"`
	FindingsResolved []Finding `json:"findingsResolved,omitempty"`
}

// CollectorSkip records a collector that was skipped during the scan.
type CollectorSkip struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
	RBAC   bool   `json:"rbac"` // true when error appears to be a permissions/forbidden error
}

type Metadata struct {
	CustomerID  string `json:"customerId,omitempty"`
	Site        string `json:"site,omitempty"`
	ClusterName string `json:"clusterName,omitempty"`
	Environment string `json:"environment,omitempty"`
	ToolVersion string `json:"toolVersion"`
	GeneratedAt string `json:"generatedAt"`
}

type Tool struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	BuildDate string `json:"buildDate"`
}

type Scan struct {
	ScanID          string    `json:"scanId"`
	StartedAt       time.Time `json:"startedAt"`
	EndedAt         time.Time `json:"endedAt"`
	DurationSeconds int       `json:"durationSeconds"`
	Mode            string    `json:"mode"`
}

type Cluster struct {
	APIServer struct {
		Endpoint string `json:"endpoint,omitempty"`
	} `json:"apiServer"`
	Platform Platform `json:"platform,omitempty"`
}

// BackupDetectedTool is one detected backup solution.
type BackupDetectedTool struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace,omitempty"`
	Version   string   `json:"version,omitempty"`
	Detected  bool     `json:"detected"`
	CRDsFound []string `json:"crdsFound,omitempty"`
}

// BackupPolicy represents a detected backup schedule or policy object.
type BackupPolicy struct {
	Tool            string   `json:"tool"`
	Name            string   `json:"name"`
	PolicyNamespace string   `json:"policyNamespace,omitempty"` // namespace the policy object lives in
	IncludedNS      []string `json:"includedNamespaces,omitempty"` // empty = all namespaces
	ExcludedNS      []string `json:"excludedNamespaces,omitempty"`
	Schedule        string   `json:"schedule,omitempty"`      // cron expression or label e.g. "@daily"
	RetentionTTL    string   `json:"retentionTtl,omitempty"` // e.g. "720h0m0s"
	RPOHours        int      `json:"rpoHours"`                // estimated RPO in hours; -1 = unknown
	HasOffsite      bool     `json:"hasOffsite"`
	StorageLocation string   `json:"storageLocation,omitempty"`
}

// RestoreSimNamespace holds the restore feasibility assessment for one namespace.
type RestoreSimNamespace struct {
	Namespace   string   `json:"namespace"`
	HasCoverage bool     `json:"hasCoverage"`
	RPOHours    int      `json:"rpoHours"` // best RPO from applicable policies; -1 = unknown
	PVCSizeGB   float64  `json:"pvcSizeGb"`
	Blockers    []string `json:"blockers,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
}

// RestoreSimResult holds the full restore simulation output.
type RestoreSimResult struct {
	Namespaces    []RestoreSimNamespace `json:"namespaces"`
	TotalPVCsGB   float64               `json:"totalPvcsGb"`
	CoveredPVCsGB float64               `json:"coveredPvcsGb"`
	UncoveredNS   []string              `json:"uncoveredNamespaces,omitempty"`
}

// BackupInventory holds the result of backup tool detection.
type BackupInventory struct {
	Tools               []BackupDetectedTool `json:"tools"`
	PrimaryTool         string               `json:"primaryTool"` // "none" if nothing found
	CoveredNamespaces   []string             `json:"coveredNamespaces,omitempty"`
	UncoveredStatefulNS []string             `json:"uncoveredStatefulNamespaces,omitempty"`
	Policies            []BackupPolicy       `json:"policies,omitempty"`
	HasOffsite          bool                 `json:"hasOffsite"`
	RestoreSim          *RestoreSimResult    `json:"restoreSim,omitempty"`
}

// RemediationStep is one prioritized DR remediation action.
type RemediationStep struct {
	Priority    int      `json:"priority"`    // 1=critical, 2=recommended, 3=optional
	Category    string   `json:"category"`    // Storage, Backup, Workload, Network, Config
	Title       string   `json:"title"`
	Detail      string   `json:"detail"`
	Commands    []string `json:"commands,omitempty"`
	TargetNotes string   `json:"targetNotes,omitempty"`
	FindingID   string   `json:"findingId,omitempty"`
}

type Inventory struct {
	// Core resources (existing)
	Namespaces   []Namespace             `json:"namespaces"`
	PVCs         []PersistentVolumeClaim `json:"pvcs"`
	PVs          []PersistentVolume      `json:"pvs"`
	Pods         []Pod                   `json:"pods"`
	StatefulSets []StatefulSet           `json:"statefulSets"`
	Nodes        []Node                  `json:"nodes"`
	StorageClasses []StorageClass        `json:"storageClasses"`

	// Workload resources
	Deployments []Deployment `json:"deployments,omitempty"`
	DaemonSets  []DaemonSet  `json:"daemonSets,omitempty"`
	Jobs        []Job        `json:"jobs,omitempty"`
	CronJobs    []CronJob    `json:"cronJobs,omitempty"`

	// Network resources
	Services       []Service       `json:"services,omitempty"`
	Ingresses      []Ingress       `json:"ingresses,omitempty"`
	NetworkPolicies []NetworkPolicy `json:"networkPolicies,omitempty"`

	// Config resources
	ConfigMaps     []ConfigMap          `json:"configMaps,omitempty"`
	Secrets        []Secret             `json:"secrets,omitempty"`
	ClusterRoles   []ClusterRole        `json:"clusterRoles,omitempty"`
	ClusterRoleBindings []ClusterRoleBinding `json:"clusterRoleBindings,omitempty"`
	ResourceQuotas []ResourceQuota      `json:"resourceQuotas,omitempty"`
	HPAs           []HPA                `json:"hpas,omitempty"`
	PodDisruptionBudgets []PodDisruptionBudget `json:"podDisruptionBudgets,omitempty"`

	// Extended inventory
	CRDs         []CRD           `json:"crds,omitempty"`
	HelmReleases []HelmRelease   `json:"helmReleases,omitempty"`
	Images       []ContainerImage `json:"images,omitempty"`
	Certificates []Certificate   `json:"certificates,omitempty"`

	// Backup detection result
	Backup BackupInventory `json:"backup,omitempty"`

	// Remediation steps
	RemediationSteps []RemediationStep `json:"remediationSteps,omitempty"`

	Findings []Finding `json:"findings"`
}

type Namespace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewBundle(scanID string, started time.Time) Bundle {
	return Bundle{
		SchemaVersion: "2.0.0",
		Metadata: Metadata{
			ToolVersion: "0.7.0",
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		},
		Tool: Tool{
			Name:      "k8s-recovery-visualizer",
			Version:   "0.7.0",
			BuildDate: time.Now().UTC().Format("2006-01-02"),
		},
		Scan: Scan{
			ScanID:    scanID,
			StartedAt: started,
			Mode:      "auto",
		},
		Inventory: Inventory{
			Namespaces:   []Namespace{},
			PVCs:         []PersistentVolumeClaim{},
			PVs:          []PersistentVolume{},
			Pods:         []Pod{},
			StatefulSets: []StatefulSet{},
			Nodes:        []Node{},
			StorageClasses: []StorageClass{},
			Findings:     []Finding{},
			Backup: BackupInventory{
				PrimaryTool: "none",
				Tools:       []BackupDetectedTool{},
			},
		},
		Score: Score{
			Storage:  DomainScore{Max: 100, Final: 100},
			Workload: DomainScore{Max: 100, Final: 100},
			Config:   DomainScore{Max: 100, Final: 100},
			Backup:   DomainScore{Max: 100, Final: 100},
			Overall:  DomainScore{Max: 100, Final: 100},
			Maturity: "PLATINUM",
		},
	}
}
