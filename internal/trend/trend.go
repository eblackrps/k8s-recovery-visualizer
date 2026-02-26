package trend

import "math"

type Direction string

const (
	Up   Direction = "up"
	Down Direction = "down"
	Flat Direction = "flat"
)

type Trend struct {
	DeltaScore   float64   `json:"deltaScore"`
	DeltaPercent float64   `json:"deltaPercent"`
	Direction    Direction `json:"direction"`
	From         float64   `json:"from"`
	To           float64   `json:"to"`
}

func Compute(prev, curr float64) Trend {
	d := curr - prev

	dir := Flat
	if d > 0.00001 {
		dir = Up
	} else if d < -0.00001 {
		dir = Down
	}

	dp := 0.0
	if math.Abs(prev) > 0.00001 {
		dp = (d / prev) * 100.0
	}

	return Trend{
		DeltaScore:   round(d, 2),
		DeltaPercent: round(dp, 2),
		Direction:    dir,
		From:         round(prev, 2),
		To:           round(curr, 2),
	}
}

func round(v float64, places int) float64 {
	p := math.Pow(10, float64(places))
	return math.Round(v*p) / p
}
