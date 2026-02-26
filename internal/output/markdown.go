package output

import (
	"fmt"
	"os"
	"sort"

	"k8s-recovery-visualizer/internal/model"
)

func WriteMarkdown(path string, b *model.Bundle) error {

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# DR Assessment Report\n\n")
	fmt.Fprintf(f, "## Overall Score: %d / %d\n\n**DR Maturity Level: %s**\n\n", b.Score.Overall.Final, b.Score.Overall.Max, b.Score.Maturity)

	fmt.Fprintf(f, "### Domain Breakdown\n")
	fmt.Fprintf(f, "- Storage:  %d / %d\n", b.Score.Storage.Final, b.Score.Storage.Max)
	fmt.Fprintf(f, "- Workload: %d / %d\n", b.Score.Workload.Final, b.Score.Workload.Max)
	fmt.Fprintf(f, "- Config:   %d / %d\n\n", b.Score.Config.Final, b.Score.Config.Max)

	fmt.Fprintf(f, "## Findings\n\n")

	if len(b.Inventory.Findings) == 0 {
		fmt.Fprintf(f, "No critical DR findings detected.\n")
		return nil
	}

	sort.Slice(b.Inventory.Findings, func(i, j int) bool {
		return severityRank(b.Inventory.Findings[i].Severity) >
			severityRank(b.Inventory.Findings[j].Severity)
	})

	for _, finding := range b.Inventory.Findings {
		fmt.Fprintf(f, "### [%s] %s\n", finding.Severity, finding.ResourceID)
		fmt.Fprintf(f, "- Issue: %s\n", finding.Message)
		fmt.Fprintf(f, "- Recommendation: %s\n\n", finding.Recommendation)
	}

	return nil
}

func severityRank(s string) int {
	switch s {
	case "CRITICAL":
		return 3
	case "HIGH":
		return 2
	case "MEDIUM":
		return 1
	default:
		return 0
	}
}
