package risk

type Posture string

const (
	Low      Posture = "LOW"
	Moderate Posture = "MODERATE"
	High     Posture = "HIGH"
	Critical Posture = "CRITICAL"
)

type Rating struct {
	Score    float64 `json:"score"`
	Maturity string  `json:"maturity"`
	Posture  Posture `json:"posture"`
}

func FromScore(score float64, maturity string) Rating {
	p := Critical
	switch {
	case score >= 90:
		p = Low
	case score >= 70:
		p = Moderate
	case score >= 50:
		p = High
	default:
		p = Critical
	}
	return Rating{Score: score, Maturity: maturity, Posture: p}
}
