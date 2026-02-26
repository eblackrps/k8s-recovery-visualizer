package model

import "time"

type Bundle struct {
	Checks        []Check   `json:"checks,omitempty"`
	SchemaVersion string    `json:"schemaVersion"`
	Metadata      Metadata  `json:"metadata"`
	Tool          Tool      `json:"tool"`
	Scan          Scan      `json:"scan"`
	Cluster       Cluster   `json:"cluster"`
	Inventory     Inventory `json:"inventory"`
	Score         Score     `json:"score"`
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
}

type Inventory struct {
	Namespaces   []Namespace             `json:"namespaces"`
	PVCs         []PersistentVolumeClaim `json:"pvcs"`
	PVs          []PersistentVolume      `json:"pvs"`
	Pods         []Pod                   `json:"pods"`
	StatefulSets []StatefulSet           `json:"statefulSets"`
  Nodes          []Node         `json:"nodes"`
  StorageClasses []StorageClass `json:"storageClasses"`
	Findings     []Finding               `json:"findings"`
}

type Namespace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewBundle(scanID string, started time.Time) Bundle {
	return Bundle{
		SchemaVersion: "1.0.0",
		Metadata: Metadata{
			ToolVersion: "0.1.0",
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		},
		Tool: Tool{
			Name:      "k8s-recovery-visualizer",
			Version:   "0.1.0",
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
          Nodes:          []Node{},
          StorageClasses: []StorageClass{},
			Findings:     []Finding{},
		},
		Score: Score{
			Storage:  DomainScore{Max: 100, Final: 100},
			Workload: DomainScore{Max: 100, Final: 100},
			Config:   DomainScore{Max: 100, Final: 100},
			Overall:  DomainScore{Max: 100, Final: 100},
			Maturity: "PLATINUM",
		},
	}
}



