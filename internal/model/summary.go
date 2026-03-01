package model

// FindingCounts holds per-severity finding counts for CI output.
type FindingCounts struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Total    int `json:"total"`
}

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
	Findings     FindingCounts   `json:"findings"`
}
