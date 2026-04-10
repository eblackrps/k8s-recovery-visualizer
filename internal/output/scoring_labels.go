package output

import (
	"fmt"

	"k8s-recovery-visualizer/internal/scoring"
)

var outputScoreRegistry = scoring.Default()

func domainWeightLabel(domain string) string {
	return fmt.Sprintf("%d%%", outputScoreRegistry.DomainWeight(domain))
}
