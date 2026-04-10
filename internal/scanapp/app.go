package scanapp

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"k8s-recovery-visualizer/internal/analyze"
	"k8s-recovery-visualizer/internal/backup"
	"k8s-recovery-visualizer/internal/kube"
	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/profile"
	"k8s-recovery-visualizer/internal/remediation"
	"k8s-recovery-visualizer/internal/restore"
	"k8s.io/client-go/dynamic"
)

func Main(args []string) int {
	opts, err := ParseArgs(args, os.Stderr)
	if err != nil {
		log.Printf("scan: %v", err)
		return 2
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		log.Printf("mkdir failed: %v", err)
		return 1
	}

	bundle := prepareBundle(opts, time.Now().UTC(), model.NewUUID())
	if !opts.CI {
		fmt.Printf("Profile: %s\n", bundle.Profile)
	}

	if opts.DryRun {
		if err := runDryScan(&bundle, opts); err != nil {
			log.Printf("dry-run failed: %v", err)
			return 1
		}
	} else {
		if opts.Insecure && !opts.CI {
			fmt.Println("WARNING: --insecure is set - TLS certificate verification is disabled.")
		}
		if err := runLiveScan(&bundle, opts); err != nil {
			log.Printf("scan failed: %v", err)
			return 1
		}
	}

	if err := applyComparison(&bundle, opts.CompareTo); err != nil {
		log.Printf("compare: failed to load %s: %v (skipping)", opts.CompareTo, err)
	}

	trendLabel, trendDelta, err := WriteOutputs(&bundle, opts.OutDir, opts.CI, opts.MinScore, opts.CSVExport, opts.Summary, opts.RedactOut, opts.Runbook)
	if err != nil {
		log.Printf("write outputs failed: %v", err)
		return 1
	}
	if opts.CI {
		PrintCISummary(&bundle, opts.MinScore, trendLabel, trendDelta)
	}
	return exitCode(&bundle, opts.MinScore, opts.CI)
}

func prepareBundle(opts Options, startedAt time.Time, scanID string) model.Bundle {
	bundle := model.NewBundle(scanID, startedAt)
	bundle.Metadata.CustomerID = opts.CustomerID
	bundle.Metadata.Site = opts.Site
	bundle.Metadata.ClusterName = opts.Cluster
	bundle.Metadata.Environment = opts.Environment
	bundle.Target = opts.Target
	bundle.Profile = string(profile.Normalize(opts.ProfileName))
	if len(opts.Namespaces) > 0 {
		bundle.ScanNamespaces = append(bundle.ScanNamespaces, opts.Namespaces...)
	}
	return bundle
}

func runDryScan(bundle *model.Bundle, opts Options) error {
	bundle.Inventory.Namespaces = []model.Namespace{
		{ID: "ns:default", Name: "default"},
		{ID: "ns:test", Name: "test"},
	}
	finalizeBundle(bundle, opts.Target)
	return nil
}

func runLiveScan(bundle *model.Bundle, opts Options) error {
	clientset, restCfg, err := kube.NewClient(opts.Kubeconfig, opts.Insecure)
	if err != nil {
		return fmt.Errorf("kube error: %w", err)
	}

	dc, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("dynamic client error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	bundle.Cluster.APIServer.Endpoint = restCfg.Host

	if err := runCollectors(ctx, clientset, dc, bundle); err != nil {
		return err
	}

	backup.Detect(ctx, clientset, bundle)
	finalizeBundle(bundle, opts.Target)
	return nil
}

func finalizeBundle(bundle *model.Bundle, target string) {
	sim := restore.Simulate(bundle)
	bundle.Inventory.Backup.RestoreSim = &sim
	backup.AssessAssurance(bundle)
	analyze.Evaluate(bundle)
	bundle.Inventory.RemediationSteps = remediation.Generate(bundle, target)
}

func exitCode(b *model.Bundle, minScore int, quiet bool) int {
	score := b.Score.Overall.Final
	if !quiet {
		fmt.Println("Final Score:", score)
		fmt.Println("DR Maturity:", b.Score.Maturity)
	}
	if score < minScore {
		if !quiet {
			fmt.Printf("DR Status: FAILED (score below %d)\n", minScore)
		}
		return 2
	}
	if !quiet {
		fmt.Println("DR Status: PASSED")
	}
	return 0
}
