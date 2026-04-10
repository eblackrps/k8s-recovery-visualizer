package model

import "time"

const (
	ScanSchemaVersion     = "2.2.0"
	EnrichedSchemaVersion = "1.1.0"
)

var (
	Version   = "1.2.0-dev"
	BuildDate = "unknown"
)

func EffectiveBuildDate(now time.Time) string {
	if BuildDate != "" && BuildDate != "unknown" {
		return BuildDate
	}
	return now.UTC().Format("2006-01-02")
}
