package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s-recovery-visualizer/internal/analyze"
	"k8s-recovery-visualizer/internal/backup"
	"k8s-recovery-visualizer/internal/collect"
	"k8s-recovery-visualizer/internal/compare"
	"k8s-recovery-visualizer/internal/enrich"
	"k8s-recovery-visualizer/internal/history"
	"k8s-recovery-visualizer/internal/kube"
	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/output"
	"k8s-recovery-visualizer/internal/profile"
	"k8s-recovery-visualizer/internal/remediation"
	"k8s-recovery-visualizer/internal/restore"
)

func main() {
	var (
		kubeconfig = flag.String("kubeconfig", "", "Path to kubeconfig")
		outDir     = flag.String("out", "./out", "Output directory")
		dryRun     = flag.Bool("dry-run", false, "Run without Kubernetes")
		ci         = flag.Bool("ci", false, "CI mode (machine-readable output)")
		minScore   = flag.Int("min-score", 90, "Minimum acceptable DR score")
		timeoutSec = flag.Int("timeout", 60, "Timeout in seconds for Kubernetes API calls")
		customerID = flag.String("customer", "", "Customer identifier (optional)")
		site       = flag.String("site", "", "Site/region name (optional)")
		cluster    = flag.String("cluster", "", "Cluster name (optional)")
		env        = flag.String("env", "", "Environment (prod/dev/test) (optional)")
		target     = flag.String("target", "vm", "Recovery target type: baremetal or vm")
		csvExport  = flag.Bool("csv", false, "Also write CSV exports alongside HTML report")
		namespace  = flag.String("namespace", "", "Comma-separated namespaces to scan (empty = all namespaces)")
		compareTo  = flag.String("compare", "", "Path to a previous recovery-scan.json to diff against")
		summary    = flag.Bool("summary", false, "Also write a print-optimised executive summary HTML")
		redactOut   = flag.Bool("redact", false, "Also write redacted JSON and HTML with masked identifiers")
		profileName = flag.String("profile", "standard", "Scoring profile: standard|enterprise|dev|airgap")
		runbook     = flag.Bool("runbook", false, "Also write a customer-facing DR runbook HTML")
	)
	flag.Parse()

	if *target != "baremetal" && *target != "vm" {
		log.Fatalf("--target must be 'baremetal' or 'vm', got %q", *target)
	}

	start := time.Now().UTC()
	scanID := model.NewUUID()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatalf("mkdir failed: %v", err)
	}

	bundle := model.NewBundle(scanID, start)
	bundle.Metadata.CustomerID = *customerID
	bundle.Metadata.Site = *site
	bundle.Metadata.ClusterName = *cluster
	bundle.Metadata.Environment = *env
	bundle.Target = *target
	bundle.Profile = string(profile.Normalize(*profileName))
	if *namespace != "" {
		for _, ns := range strings.Split(*namespace, ",") {
			ns = strings.TrimSpace(ns)
			if ns != "" {
				bundle.ScanNamespaces = append(bundle.ScanNamespaces, ns)
			}
		}
	}

	if !*ci {
		fmt.Printf("Profile: %s\n", bundle.Profile)
	}

	if *dryRun {
		bundle.Inventory.Namespaces = []model.Namespace{
			{ID: "ns:default", Name: "default"},
			{ID: "ns:test", Name: "test"},
		}
		sim := restore.Simulate(&bundle)
		bundle.Inventory.Backup.RestoreSim = &sim
		analyze.Evaluate(&bundle)
		bundle.Inventory.RemediationSteps = remediation.Generate(&bundle, *target)
		applyComparison(&bundle, *compareTo)
		trendLabel, trendDelta := write(&bundle, *outDir, *ci, *minScore, *csvExport, *summary, *redactOut, *runbook)
		if *ci {
			printCISummary(&bundle, *minScore, trendLabel, trendDelta)
		}
		exitWithPolicy(&bundle, *minScore, *ci)
		return
	}

	clientset, restCfg, err := kube.NewClient(*kubeconfig)
	if err != nil {
		log.Fatalf("kube error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutSec)*time.Second)
	defer cancel()
	bundle.Cluster.APIServer.Endpoint = restCfg.Host

	// ── Core collectors ────────────────────────────────────────────────────
	if err := collect.Namespaces(ctx, clientset, &bundle); err != nil {
		log.Fatalf("collect namespaces: %v", err)
	}
	if err := collect.Nodes(ctx, clientset, &bundle); err != nil {
		log.Fatalf("collect nodes: %v", err)
	}
	if err := collect.Pods(ctx, clientset, &bundle); err != nil {
		log.Fatalf("collect pods: %v", err)
	}
	if err := collect.PVCs(ctx, clientset, &bundle); err != nil {
		log.Fatalf("collect pvcs: %v", err)
	}
	if err := collect.PVs(ctx, clientset, &bundle); err != nil {
		log.Fatalf("collect pvs: %v", err)
	}
	if err := collect.StatefulSets(ctx, clientset, &bundle); err != nil {
		log.Fatalf("collect statefulsets: %v", err)
	}
	if err := collect.StorageClasses(ctx, clientset, &bundle); err != nil {
		log.Fatalf("collect storageclasses: %v", err)
	}

	// ── Workload collectors ─────────────────────────────────────────────────
	tryCollect("Deployments", collect.Deployments(ctx, clientset, &bundle), &bundle)
	tryCollect("DaemonSets", collect.DaemonSets(ctx, clientset, &bundle), &bundle)
	tryCollect("Jobs", collect.Jobs(ctx, clientset, &bundle), &bundle)
	tryCollect("CronJobs", collect.CronJobs(ctx, clientset, &bundle), &bundle)

	// ── Networking collectors ───────────────────────────────────────────────
	tryCollect("Services", collect.Services(ctx, clientset, &bundle), &bundle)
	tryCollect("Ingresses", collect.Ingresses(ctx, clientset, &bundle), &bundle)
	tryCollect("NetworkPolicies", collect.NetworkPolicies(ctx, clientset, &bundle), &bundle)

	// ── Config / RBAC collectors ────────────────────────────────────────────
	tryCollect("ConfigMaps", collect.ConfigMaps(ctx, clientset, &bundle), &bundle)
	tryCollect("Secrets", collect.Secrets(ctx, clientset, &bundle), &bundle)
	tryCollect("ClusterRoles", collect.ClusterRoles(ctx, clientset, &bundle), &bundle)
	tryCollect("ClusterRoleBindings", collect.ClusterRoleBindings(ctx, clientset, &bundle), &bundle)
	tryCollect("HPAs", collect.HPAs(ctx, clientset, &bundle), &bundle)
	tryCollect("PodDisruptionBudgets", collect.PodDisruptionBudgets(ctx, clientset, &bundle), &bundle)
	tryCollect("ResourceQuotas", collect.ResourceQuotas(ctx, clientset, &bundle), &bundle)
	tryCollect("CRDs", collect.CRDs(ctx, clientset, &bundle), &bundle)

	// ── Advanced collectors ─────────────────────────────────────────────────
	tryCollect("HelmReleases", collect.HelmReleases(ctx, clientset, &bundle), &bundle)
	tryCollect("Platform", collect.Platform(ctx, clientset, &bundle), &bundle)
	tryCollect("Certificates", collect.Certificates(ctx, clientset, &bundle), &bundle)

	// Images is post-collection (derives data from already-collected workloads)
	tryCollect("Images", collect.Images(ctx, clientset, &bundle), &bundle)

	// ── Backup detection + restore simulation ───────────────────────────────
	backup.Detect(ctx, clientset, &bundle)
	sim := restore.Simulate(&bundle)
	bundle.Inventory.Backup.RestoreSim = &sim

	// ── Scoring + remediation ───────────────────────────────────────────────
	analyze.Evaluate(&bundle)
	bundle.Inventory.RemediationSteps = remediation.Generate(&bundle, *target)

	// ── Comparison (--compare) ───────────────────────────────────────────────
	applyComparison(&bundle, *compareTo)

	// ── Write outputs ───────────────────────────────────────────────────────
	trendLabel, trendDelta := write(&bundle, *outDir, *ci, *minScore, *csvExport, *summary, *redactOut, *runbook)

	if *ci {
		printCISummary(&bundle, *minScore, trendLabel, trendDelta)
	}
	exitWithPolicy(&bundle, *minScore, *ci)
}

// write serialises all outputs and returns trend label + delta for CI summary.
func write(bundle *model.Bundle, outDir string, quiet bool, minScore int, csvExport, summaryOut, redactOut, runbookOut bool) (string, int) {
	bundle.Scan.EndedAt = time.Now().UTC()
	bundle.Scan.DurationSeconds = int(bundle.Scan.EndedAt.Sub(bundle.Scan.StartedAt).Seconds())
	bundle.Checks = analyze.BuildChecks(bundle, minScore)

	jsonPath := filepath.Join(outDir, "recovery-scan.json")
	htmlPath := filepath.Join(outDir, "recovery-report.html")

	if err := output.WriteJSON(jsonPath, bundle); err != nil {
		log.Fatalf("write json: %v", err)
	}

	// Enrich pipeline: trend history, risk, enriched.json, markdown report
	var trendLabel string
	var trendDelta int
	if en, err := enrich.Run(enrich.Options{OutDir: outDir, LastNCount: 10}); err != nil {
		if !quiet {
			fmt.Printf("Enrich: FAILED (%v)\n", err)
		}
	} else if err := enrich.WriteArtifacts(outDir, en); err != nil {
		if !quiet {
			fmt.Printf("Enrich: FAILED writing artifacts (%v)\n", err)
		}
	}

	// Trend (separate from enrich, uses history package)
	if tr, err := history.Record(outDir, bundle); err != nil {
		if !quiet {
			fmt.Println("History: (skipped)", err)
		}
		trendLabel = "HISTORY_SKIPPED"
	} else if tr.Label == "FIRST_RUN" {
		if !quiet {
			fmt.Println("Trend: FIRST RUN (no previous scan found)")
		}
		trendLabel = "FIRST_RUN"
	} else {
		trendLabel = tr.Label
		trendDelta = tr.Delta
		if !quiet {
			sign := ""
			if tr.Delta > 0 {
				sign = "+"
			}
			fmt.Printf("Trend: %s (%s%d) Previous: %d, Current: %d\n",
				tr.Label, sign, tr.Delta, tr.Previous, tr.Current)
		}
	}

	// New tabbed HTML report (overwrites the simple one produced by enrich)
	if err := output.WriteReport(htmlPath, bundle); err != nil {
		log.Fatalf("write html report: %v", err)
	}

	// Optional CSV export
	if csvExport {
		if err := output.WriteCSV(outDir, bundle); err != nil {
			log.Fatalf("write csv: %v", err)
		}
		if !quiet {
			fmt.Println("CSV exports:", filepath.Join(outDir, "csv"))
		}
	}

	// Optional executive summary
	if summaryOut {
		summaryPath := filepath.Join(outDir, "recovery-summary.html")
		if err := output.WriteSummary(summaryPath, bundle); err != nil {
			log.Fatalf("write summary: %v", err)
		}
		if !quiet {
			fmt.Println("Executive Summary:", summaryPath)
		}
	}

	// Optional DR runbook
	if runbookOut {
		runbookPath := filepath.Join(outDir, "recovery-runbook.html")
		if err := output.WriteRunbook(runbookPath, bundle); err != nil {
			log.Fatalf("write runbook: %v", err)
		}
		if !quiet {
			fmt.Println("DR Runbook:", runbookPath)
		}
	}

	// Optional redacted exports
	if redactOut {
		if err := output.WriteRedactedJSON(filepath.Join(outDir, "recovery-scan-redacted.json"), bundle); err != nil {
			log.Fatalf("write redacted json: %v", err)
		}
		if err := output.WriteRedactedReport(filepath.Join(outDir, "recovery-report-redacted.html"), bundle); err != nil {
			log.Fatalf("write redacted html: %v", err)
		}
		if !quiet {
			fmt.Println("Redacted exports: recovery-scan-redacted.json, recovery-report-redacted.html")
		}
	}

	if !quiet {
		fmt.Println("Scan complete.")
		fmt.Println("JSON:", jsonPath)
		fmt.Println("HTML Report:", htmlPath)
		fmt.Println("Enriched:", filepath.Join(outDir, "recovery-enriched.json"))
	}

	return trendLabel, trendDelta
}

func printCISummary(b *model.Bundle, minScore int, trendLabel string, trendDelta int) {
	summary := model.ScanSummary{
		ScanID:       b.Scan.ScanID,
		TimestampUtc: time.Now().UTC().Format(time.RFC3339),
		Overall:      b.Score.Overall.Final,
		Maturity:     b.Score.Maturity,
		Status:       "PASSED",
		MinScore:     minScore,
		Profile:      b.Profile,
		Categories:   analyze.BuildCategories(b),
		Trend:        trendLabel,
		Delta:        trendDelta,
	}
	if b.Score.Overall.Final < minScore {
		summary.Status = "FAILED"
	}
	raw, _ := json.Marshal(summary)
	fmt.Println()
	fmt.Println(string(raw))
}

// applyComparison loads a previous scan and attaches diff to bundle.
func applyComparison(bundle *model.Bundle, compareTo string) {
	if compareTo == "" {
		return
	}
	prev, err := loadBundle(compareTo)
	if err != nil {
		log.Printf("compare: failed to load %s: %v (skipping)", compareTo, err)
		return
	}
	diff := compare.Diff(prev, bundle)
	bundle.Comparison = &diff
}

// loadBundle reads and decodes a recovery-scan.json file.
func loadBundle(path string) (*model.Bundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var b model.Bundle
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// tryCollect records a collector skip when err != nil.
// It logs the error and appends a CollectorSkip to the bundle.
func tryCollect(name string, err error, bundle *model.Bundle) {
	if err == nil {
		return
	}
	msg := err.Error()
	isRBAC := strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "Forbidden") ||
		strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "Unauthorized")
	bundle.CollectorSkips = append(bundle.CollectorSkips, model.CollectorSkip{
		Name:   name,
		Reason: msg,
		RBAC:   isRBAC,
	})
	log.Printf("collect %s: %v (skipping)", name, err)
}

func exitWithPolicy(b *model.Bundle, minScore int, quiet bool) {
	score := b.Score.Overall.Final
	if !quiet {
		fmt.Println("Final Score:", score)
		fmt.Println("DR Maturity:", b.Score.Maturity)
	}
	if score < minScore {
		if !quiet {
			fmt.Printf("DR Status: FAILED (score below %d)\n", minScore)
		}
		os.Exit(2)
	}
	if !quiet {
		fmt.Println("DR Status: PASSED")
	}
	os.Exit(0)
}
