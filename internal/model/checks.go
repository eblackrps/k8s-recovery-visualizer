package model

// Check is a human-readable scoring signal for DR readiness.
// Keep it simple: this is meant to be shown directly in reports.
type Check struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
  Category string `json:"category"`
  Score    int    `json:"score"`
  Note     string `json:"note,omitempty"`
	Status      string `json:"status"` // PASS/WARN/FAIL
	Weight      int    `json:"weight,omitempty"`
	Message     string `json:"message,omitempty"`
	Remediation string `json:"remediation,omitempty"`
}


