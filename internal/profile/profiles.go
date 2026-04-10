package profile

import "k8s-recovery-visualizer/internal/scoring"

type Name string

const (
	Standard   Name = "standard"
	Enterprise Name = "enterprise"
	Dev        Name = "dev"
	Airgap     Name = "airgap"
)

func Normalize(s string) Name {
	switch scoring.NormalizeProfile(s) {
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
	return scoring.Default().ProfileMultipliers(string(p))
}
