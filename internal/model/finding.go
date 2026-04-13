package model

type EvidenceConfidence string

const (
	EvidenceConfidenceConfirmed EvidenceConfidence = "confirmed"
	EvidenceConfidenceInferred  EvidenceConfidence = "inferred"
	EvidenceConfidenceUnknown   EvidenceConfidence = "unknown"
)

type Finding struct {
	ID             string             `json:"id"`
	Title          string             `json:"title,omitempty"`
	Domain         string             `json:"domain,omitempty"`
	Severity       string             `json:"severity"`
	ResourceID     string             `json:"resourceId"`
	Message        string             `json:"message"`
	Recommendation string             `json:"recommendation"`
	Impact         string             `json:"impact,omitempty"`
	Effort         string             `json:"effort,omitempty"`
	OwnerHint      string             `json:"ownerHint,omitempty"`
	PriorityScore  int                `json:"priorityScore,omitempty"`
	Rank           int                `json:"rank,omitempty"`
	Penalty        int                `json:"penalty,omitempty"`
	Confidence     EvidenceConfidence `json:"confidence,omitempty"`
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
