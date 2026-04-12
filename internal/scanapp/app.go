package scanapp

import (
	"context"
	"fmt"
	"log"
	"os"

	"k8s-recovery-visualizer/internal/appcore"
)

func Main(args []string) int {
	opts, err := ParseArgs(args, os.Stderr)
	if err != nil {
		log.Printf("scan: %v", err)
		return 2
	}
	if !opts.CI {
		fmt.Printf("Profile: %s\n", opts.ProfileName)
	}

	service := appcore.NewService()
	result, err := service.Run(context.Background(), appcore.ScanRequest{
		KubeconfigPath:        opts.Kubeconfig,
		ContextName:           opts.ContextName,
		OutputDir:             opts.OutDir,
		DryRun:                opts.DryRun,
		CI:                    opts.CI,
		MinScore:              opts.MinScore,
		TimeoutSeconds:        int(opts.Timeout.Seconds()),
		CustomerID:            opts.CustomerID,
		Site:                  opts.Site,
		ClusterName:           opts.Cluster,
		Environment:           opts.Environment,
		Target:                opts.Target,
		CSVExport:             opts.CSVExport,
		Namespaces:            opts.Namespaces,
		CompareTo:             opts.CompareTo,
		Summary:               opts.Summary,
		Redact:                opts.RedactOut,
		ProfileName:           opts.ProfileName,
		Runbook:               opts.Runbook,
		Insecure:              opts.Insecure,
		IncludeSecretMetadata: opts.IncludeSecretMetadata,
	}, nil)
	if err != nil {
		log.Printf("scan failed: %v", err)
		return 1
	}

	if opts.CI {
		PrintCISummary(&result.Workspace.Bundle, opts.MinScore, result.TrendLabel, result.TrendDelta)
		return result.ExitCode
	}

	if opts.Insecure {
		fmt.Println("WARNING: --insecure is set - TLS certificate verification is disabled.")
	}
	if result.TrendLabel != "" {
		sign := ""
		if result.TrendDelta > 0 {
			sign = "+"
		}
		switch result.TrendLabel {
		case "FIRST_RUN":
			fmt.Println("Trend: FIRST RUN (no previous scan found)")
		default:
			fmt.Printf("Trend: %s (%s%d)\n", result.TrendLabel, sign, result.TrendDelta)
		}
	}
	fmt.Println("Final Score:", result.Workspace.Bundle.Score.Overall.Final)
	fmt.Println("DR Maturity:", result.Workspace.Bundle.Score.Maturity)
	if result.ExitCode == 2 {
		fmt.Printf("DR Status: FAILED (score below %d)\n", opts.MinScore)
	} else {
		fmt.Println("DR Status: PASSED")
	}
	fmt.Println("Scan complete.")
	fmt.Println("JSON:", result.Artifacts.BundleJSON)
	fmt.Println("HTML Report:", result.Artifacts.HTMLReport)
	fmt.Println("Enriched:", result.Artifacts.EnrichedJSON)
	if result.Artifacts.SummaryHTML != "" {
		fmt.Println("Executive Summary:", result.Artifacts.SummaryHTML)
	}
	if result.Artifacts.RunbookHTML != "" {
		fmt.Println("DR Runbook:", result.Artifacts.RunbookHTML)
	}
	if result.Artifacts.CSVDir != "" {
		fmt.Println("CSV exports:", result.Artifacts.CSVDir)
	}
	if result.Artifacts.RedactedJSON != "" || result.Artifacts.RedactedHTML != "" {
		fmt.Println("Redacted exports:", result.Artifacts.RedactedJSON, result.Artifacts.RedactedHTML)
	}
	return result.ExitCode
}
