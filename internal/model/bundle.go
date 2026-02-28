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
	// CollectorSkips records collectors that were skipped due to RBAC or missing APIs.
	CollectorSkips []CollectorSkip `json:"collectorSkips,omitempty"`
	// ScanNamespaces restricts the scan to specific namespaces. Empty = all namespaces.
	ScanNamespaces []string `json:"scanNamespaces,omitempty"`
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

// BackupInventory holds the result of backup tool detection.
type BackupInventory struct {
	Tools               []BackupDetectedTool `json:"tools"`
	PrimaryTool         string               `json:"primaryTool"` // "none" if nothing found
	CoveredNamespaces   []string             `json:"coveredNamespaces,omitempty"`
	UncoveredStatefulNS []string             `json:"uncoveredStatefulNamespaces,omitempty"`
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
			ToolVersion: "0.4.0",
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		},
		Tool: Tool{
			Name:      "k8s-recovery-visualizer",
			Version:   "0.4.0",
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
