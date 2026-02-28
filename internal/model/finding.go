package model

type Finding struct {
	ID             string `json:"id"`
	Severity       string `json:"severity"`
	ResourceID     string `json:"resourceId"`
	Message        string `json:"message"`
	Recommendation string `json:"recommendation"`
}

type DomainScore struct {
	Max   int `json:"max"`
	Final int `json:"final"`
}

type Score struct {
	Storage  DomainScore `json:"storage"`
	Workload DomainScore `json:"workload"`
	Config   DomainScore `json:"config"`
	Backup   DomainScore `json:"backup"`
	Overall  DomainScore `json:"overall"`
	Maturity string      `json:"maturity"`
}
