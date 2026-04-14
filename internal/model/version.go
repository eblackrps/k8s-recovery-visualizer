package model

import "time"

const (
	ScanSchemaVersion     = "3.1.0"
	EnrichedSchemaVersion = "1.2.0"
)

var (
	Version   = "1.8.1"
	BuildDate = "unknown"
)

func EffectiveBuildDate(now time.Time) string {
	if BuildDate != "" && BuildDate != "unknown" {
		return BuildDate
	}
	return now.UTC().Format("2006-01-02")
}
