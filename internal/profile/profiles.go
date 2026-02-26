package profile

import "strings"

type Name string

const (
	Standard   Name = "standard"
	Enterprise Name = "enterprise"
	Dev        Name = "dev"
	Airgap     Name = "airgap"
)

func Normalize(s string) Name {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "enterprise":
		return Enterprise
	case "dev":
		return Dev
	case "airgap":
		return Airgap
	default:
		return Standard
	}
}

func Weights(p Name) map[string]float64 {
	switch p {
	case Enterprise:
		return map[string]float64{
			"restoreTesting": 1.50,
			"immutability":   1.30,
			"replication":    1.20,
			"security":       1.20,
		}
	case Dev:
		return map[string]float64{
			"restoreTesting": 1.10,
			"immutability":   0.90,
			"replication":    0.90,
		}
	case Airgap:
		return map[string]float64{
			"immutability":   1.60,
			"airgap":         1.60,
			"security":       1.30,
			"restoreTesting": 1.20,
		}
	default:
		return map[string]float64{}
	}
}
