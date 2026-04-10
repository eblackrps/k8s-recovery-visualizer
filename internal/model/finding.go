package model

type EvidenceConfidence string

const (
	EvidenceConfidenceConfirmed EvidenceConfidence = "confirmed"
	EvidenceConfidenceInferred  EvidenceConfidence = "inferred"
	EvidenceConfidenceUnknown   EvidenceConfidence = "unknown"
)

type Finding struct {
	ID             string             `json:"id"`
	Domain         string             `json:"domain,omitempty"`
	Severity       string             `json:"severity"`
	ResourceID     string             `json:"resourceId"`
	Message        string             `json:"message"`
	Recommendation string             `json:"recommendation"`
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
