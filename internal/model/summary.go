package model

type ScanSummary struct {
	ScanID       string          `json:"scanId"`
	TimestampUtc string          `json:"timestampUtc"`
	Overall      int             `json:"overall"`
	Maturity     string          `json:"maturity"`
	Status       string          `json:"status"` // PASSED/FAILED
	MinScore     int             `json:"minScore"`
	Profile      string          `json:"profile"`
	Categories   []CategoryScore `json:"categories"`
	Trend        string          `json:"trend"`
	Delta        int             `json:"delta"`
}
