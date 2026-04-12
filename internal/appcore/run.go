package appcore

import (
	"context"
	"fmt"
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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func (s *Service) Run(ctx context.Context, req ScanRequest, sink EventSink) (RunResult, error) {
	req = req.Normalized()
	runID := req.RunID
	if strings.TrimSpace(runID) == "" {
		runID = s.uuid()
	}
	emit(sink, RunEvent{Type: "status", RunID: runID, Step: "prepare", Level: "info", Message: "Preparing scan bundle.", Percent: 0.02})

	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return RunResult{}, fmt.Errorf("create output directory: %w", err)
	}

	preflight, err := s.Preflight(ctx, req)
	if err != nil {
		return RunResult{}, err
	}
	emit(sink, RunEvent{Type: "status", RunID: runID, Step: "preflight", Level: "info", Message: "Preflight checks complete.", Percent: 0.08})

	bundle := prepareBundle(req, s.now(), runID)
	if req.DryRun {
		emit(sink, RunEvent{Type: "status", RunID: runID, Step: "dry-run", Level: "info", Message: "Generating deterministic dry-run bundle.", Percent: 0.14})
		if err := runDryScan(&bundle, req); err != nil {
			return RunResult{}, err
		}
	} else {
		if req.Insecure {
			emit(sink, RunEvent{Type: "warning", RunID: runID, Step: "connect", Level: "warn", Message: "TLS verification is disabled for this run.", Warning: "TLS verification disabled because insecure mode is enabled."})
		}
		if err := s.runLiveScan(ctx, &bundle, req, sink); err != nil {
			return RunResult{}, err
		}
	}

	emit(sink, RunEvent{Type: "status", RunID: runID, Step: "analysis", Level: "info", Message: "Analyzing inventory and generating remediation guidance.", Percent: 0.74})
	finalizeBundle(&bundle, req.Target)

	if err := applyComparison(&bundle, req.CompareTo); err != nil {
		emit(sink, RunEvent{Type: "warning", RunID: runID, Step: "compare", Level: "warn", Message: fmt.Sprintf("Comparison skipped: %v", err), Warning: err.Error()})
	}

	trendLabel, trendDelta, artifacts, err := s.writeOutputs(&bundle, req, sink, true)
	if err != nil {
		return RunResult{}, err
	}

	workspace := s.workspaceFromBundle(bundle, artifacts)
	workspace.History.TrendLabel = trendLabel
	workspace.History.TrendDelta = trendDelta
	emit(sink, RunEvent{Type: "complete", RunID: runID, Step: "complete", Level: "info", Message: "Scan complete.", Percent: 1, Artifact: artifacts.HTMLReport})

	return RunResult{
		RunID:      runID,
		ExitCode:   exitCode(&bundle, req.MinScore),
		TrendLabel: trendLabel,
		TrendDelta: trendDelta,
		Artifacts:  artifacts,
		Workspace:  workspace,
		Preflight:  preflight,
	}, nil
}

func (s *Service) ExportBundle(path string, req ExportRequest) (ArtifactPaths, error) {
	bundle, err := loadBundle(path)
	if err != nil {
		return ArtifactPaths{}, err
	}
	if req.OutputDir == "" {
		req.OutputDir = filepath.Dir(path)
	}
	runReq := ScanRequest{
		OutputDir: req.OutputDir,
		CSVExport: req.CSVExport,
		Summary:   req.Summary,
		Runbook:   req.Runbook,
		Redact:    req.Redact,
		MinScore:  0,
	}
	_, _, artifacts, err := s.writeOutputs(bundle, runReq, nil, false)
	return artifacts, err
}

func prepareBundle(req ScanRequest, startedAt time.Time, scanID string) model.Bundle {
	normalized := req.Normalized()
	bundle := model.NewBundle(scanID, startedAt)
	bundle.Metadata.CustomerID = normalized.CustomerID
	bundle.Metadata.Site = normalized.Site
	bundle.Metadata.ClusterName = normalized.ClusterName
	bundle.Metadata.Environment = normalized.Environment
	bundle.Target = normalized.Target
	bundle.Profile = string(profile.Normalize(normalized.ProfileName))
	if len(normalized.Namespaces) > 0 {
		bundle.ScanNamespaces = append(bundle.ScanNamespaces, normalized.Namespaces...)
	}
	return bundle
}

func runDryScan(bundle *model.Bundle, req ScanRequest) error {
	bundle.Inventory.Namespaces = []model.Namespace{
		{ID: "ns:default", Name: "default"},
		{ID: "ns:payments", Name: "payments"},
	}
	bundle.Inventory.Nodes = []model.Node{
		{Name: "demo-control-plane", Ready: true, Roles: []string{"control-plane"}, OSImage: "Ubuntu 24.04", KubeletVersion: "v1.30.0"},
		{Name: "demo-worker-1", Ready: true, Roles: []string{"worker"}, OSImage: "Ubuntu 24.04", KubeletVersion: "v1.30.0"},
	}
	bundle.Inventory.Deployments = []model.Deployment{
		{Namespace: "default", Name: "frontend", Replicas: 3, Ready: 3, Images: []string{"ghcr.io/example/frontend:1.8.4"}},
	}
	bundle.Inventory.StatefulSets = []model.StatefulSet{
		{Namespace: "payments", Name: "postgres", Replicas: 1, HasVolumeClaim: true},
	}
	bundle.Inventory.PVCs = []model.PersistentVolumeClaim{
		{ID: "pvc:payments/postgres-data", Namespace: "payments", Name: "postgres-data", StorageClass: "fast-ssd", RequestedSize: "200Gi"},
	}
	bundle.Inventory.PVs = []model.PersistentVolume{
		{Name: "pv-payments-postgres", ClaimRef: "payments/postgres-data", StorageClass: "fast-ssd", Capacity: "200Gi", ReclaimPolicy: "Delete", Backend: "EBS"},
	}
	bundle.Inventory.Images = []model.ContainerImage{
		{Image: "ghcr.io/example/frontend:1.8.4", Registry: "ghcr.io", Workloads: []string{"default/frontend"}},
		{Image: "postgres:16.4", Registry: "docker.io", IsPublic: true, Workloads: []string{"payments/postgres"}},
	}
	bundle.Inventory.HelmReleases = []model.HelmRelease{
		{Name: "platform", Namespace: "default", Chart: "platform", Version: "2.6.0", AppVersion: "2.6.0", Status: "deployed"},
	}
	bundle.Inventory.Backup.PrimaryTool = "velero"
	bundle.Inventory.Backup.CoverageVerified = true
	bundle.Inventory.Backup.CoverageStatus = model.BackupCoverageStatusVerified
	bundle.Inventory.Backup.CoveredNamespaces = []string{"payments"}
	finalizeBundle(bundle, req.Target)
	return nil
}

func (s *Service) runLiveScan(ctx context.Context, bundle *model.Bundle, req ScanRequest, sink EventSink) error {
	emit(sink, RunEvent{Type: "status", RunID: bundle.Scan.ScanID, Step: "connect", Level: "info", Message: "Connecting to the Kubernetes API.", Percent: 0.12})
	clientset, restCfg, err := kube.NewClientWithContext(req.KubeconfigPath, req.ContextName, req.Insecure)
	if err != nil {
		return fmt.Errorf("kube error: %w", err)
	}

	dc, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("dynamic client error: %w", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, req.Timeout())
	defer cancel()
	bundle.Cluster.APIServer.Endpoint = restCfg.Host

	steps := buildCollectorPipeline(runCtx, clientset, dc, bundle, req)
	for i, step := range steps {
		select {
		case <-runCtx.Done():
			return runCtx.Err()
		default:
		}
		progressBase := 0.18
		progressSpan := 0.45
		percent := progressBase + (float64(i)/float64(max(1, len(steps))))*progressSpan
		emit(sink, RunEvent{Type: "status", RunID: bundle.Scan.ScanID, Step: step.Name, Level: "info", Message: fmt.Sprintf("Collecting %s.", step.Name), Percent: percent})
		if err := step.Run(); err != nil {
			if step.Required {
				return err
			}
			skip := recordCollectorSkip(step.Name, err, bundle)
			emit(sink, RunEvent{Type: "warning", RunID: bundle.Scan.ScanID, Step: step.Name, Level: "warn", Message: fmt.Sprintf("%s skipped.", step.Name), Warning: err.Error(), Skip: &skip, Percent: percent})
		} else {
			emit(sink, RunEvent{Type: "log", RunID: bundle.Scan.ScanID, Step: step.Name, Level: "info", Message: fmt.Sprintf("%s collected.", step.Name), Percent: percent})
		}
	}

	emit(sink, RunEvent{Type: "status", RunID: bundle.Scan.ScanID, Step: "backup", Level: "info", Message: "Inspecting backup tooling and evidence.", Percent: 0.67})
	backup.Detect(runCtx, clientset, bundle)
	return nil
}

type collectorStep struct {
	Name     string
	Required bool
	Run      func() error
}

func buildCollectorPipeline(ctx context.Context, cs *kubernetes.Clientset, dc dynamic.Interface, bundle *model.Bundle, req ScanRequest) []collectorStep {
	steps := []collectorStep{
		{Name: "Namespaces", Required: true, Run: func() error { return collect.Namespaces(ctx, cs, bundle) }},
		{Name: "Nodes", Required: true, Run: func() error { return collect.Nodes(ctx, cs, bundle) }},
		{Name: "Pods", Required: true, Run: func() error { return collect.Pods(ctx, cs, bundle) }},
		{Name: "PVCs", Required: true, Run: func() error { return collect.PVCs(ctx, cs, bundle) }},
		{Name: "PVs", Required: true, Run: func() error { return collect.PVs(ctx, cs, bundle) }},
		{Name: "StatefulSets", Required: true, Run: func() error { return collect.StatefulSets(ctx, cs, bundle) }},
		{Name: "StorageClasses", Required: true, Run: func() error { return collect.StorageClasses(ctx, cs, bundle) }},
		{Name: "Deployments", Run: func() error { return collect.Deployments(ctx, cs, bundle) }},
		{Name: "DaemonSets", Run: func() error { return collect.DaemonSets(ctx, cs, bundle) }},
		{Name: "Jobs", Run: func() error { return collect.Jobs(ctx, cs, bundle) }},
		{Name: "CronJobs", Run: func() error { return collect.CronJobs(ctx, cs, bundle) }},
		{Name: "Services", Run: func() error { return collect.Services(ctx, cs, bundle) }},
		{Name: "Ingresses", Run: func() error { return collect.Ingresses(ctx, cs, bundle) }},
		{Name: "NetworkPolicies", Run: func() error { return collect.NetworkPolicies(ctx, cs, bundle) }},
		{Name: "ConfigMaps", Run: func() error { return collect.ConfigMaps(ctx, cs, bundle) }},
		{Name: "ClusterRoles", Run: func() error { return collect.ClusterRoles(ctx, cs, bundle) }},
		{Name: "ClusterRoleBindings", Run: func() error { return collect.ClusterRoleBindings(ctx, cs, bundle) }},
		{Name: "HPAs", Run: func() error { return collect.HPAs(ctx, cs, bundle) }},
		{Name: "PodDisruptionBudgets", Run: func() error { return collect.PodDisruptionBudgets(ctx, cs, bundle) }},
		{Name: "ResourceQuotas", Run: func() error { return collect.ResourceQuotas(ctx, cs, bundle) }},
		{Name: "CRDs", Run: func() error { return collect.CRDs(ctx, cs, bundle) }},
		{Name: "HelmReleases", Run: func() error { return collect.HelmReleases(ctx, cs, bundle) }},
		{Name: "Platform", Run: func() error { return collect.Platform(ctx, cs, bundle) }},
		{Name: "Certificates", Run: func() error { return collect.Certificates(ctx, cs, bundle) }},
		{Name: "Images", Run: func() error { return collect.Images(ctx, cs, bundle) }},
		{Name: "VolumeSnapshotClasses", Run: func() error { return collect.VolumeSnapshotClasses(ctx, dc, bundle) }},
		{Name: "VolumeSnapshots", Run: func() error { return collect.VolumeSnapshots(ctx, dc, bundle) }},
		{Name: "LimitRanges", Run: func() error { return collect.LimitRanges(ctx, cs, bundle) }},
		{Name: "EtcdBackup", Run: func() error { return collect.EtcdBackup(ctx, cs, bundle) }},
		{Name: "ServiceAccounts", Run: func() error { return collect.ServiceAccounts(ctx, cs, bundle) }},
	}
	if req.IncludeSecretMetadata {
		steps = append(steps, collectorStep{Name: "Secrets", Run: func() error { return collect.Secrets(ctx, cs, bundle) }})
	}
	return steps
}

func recordCollectorSkip(name string, err error, bundle *model.Bundle) model.CollectorSkip {
	msg := err.Error()
	isRBAC := strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "Forbidden") ||
		strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "Unauthorized")
	skip := model.CollectorSkip{
		Name:   name,
		Reason: msg,
		RBAC:   isRBAC,
	}
	bundle.CollectorSkips = append(bundle.CollectorSkips, skip)
	return skip
}

func finalizeBundle(bundle *model.Bundle, target string) {
	sim := restore.Simulate(bundle)
	bundle.Inventory.Backup.RestoreSim = &sim
	backup.AssessAssurance(bundle)
	analyze.Evaluate(bundle)
	bundle.Inventory.RemediationSteps = remediation.Generate(bundle, target)
}

func applyComparison(bundle *model.Bundle, compareTo string) error {
	if compareTo == "" {
		return nil
	}
	prev, err := loadBundle(compareTo)
	if err != nil {
		return err
	}
	diff := compare.Diff(prev, bundle)
	bundle.Comparison = &diff
	return nil
}

func (s *Service) writeOutputs(bundle *model.Bundle, req ScanRequest, sink EventSink, recordHistory bool) (string, int, ArtifactPaths, error) {
	outDir := req.OutputDir
	emit(sink, RunEvent{Type: "status", RunID: bundle.Scan.ScanID, Step: "artifacts", Level: "info", Message: "Writing offline report bundle.", Percent: 0.82})
	bundle.Scan.EndedAt = s.now()
	bundle.Scan.DurationSeconds = int(bundle.Scan.EndedAt.Sub(bundle.Scan.StartedAt).Seconds())
	bundle.Checks = analyze.BuildChecks(bundle, req.MinScore)

	artifacts := detectArtifacts(outDir)
	if err := output.WriteJSON(artifacts.BundleJSON, bundle); err != nil {
		return "", 0, ArtifactPaths{}, fmt.Errorf("write json: %w", err)
	}
	if en, err := enrich.Run(enrich.Options{OutDir: outDir, LastNCount: 10, Profile: bundle.Profile}); err == nil {
		if err := enrich.WriteArtifacts(outDir, en); err != nil {
			return "", 0, ArtifactPaths{}, fmt.Errorf("write enriched artifacts: %w", err)
		}
	} else {
		return "", 0, ArtifactPaths{}, fmt.Errorf("enrich outputs: %w", err)
	}

	var trendLabel string
	var trendDelta int
	if recordHistory {
		if tr, err := history.Record(outDir, bundle); err == nil {
			trendLabel = tr.Label
			trendDelta = tr.Delta
		}
		bundle.TrendHistory = history.LoadRecent(outDir, 20)
	}

	if err := output.WriteReport(artifacts.HTMLReport, bundle); err != nil {
		return "", 0, ArtifactPaths{}, fmt.Errorf("write html report: %w", err)
	}
	if recordHistory {
		_ = history.SnapshotLatestHTML(outDir)
	}
	emit(sink, RunEvent{Type: "artifact", RunID: bundle.Scan.ScanID, Step: "report", Level: "info", Message: "Report written.", Artifact: artifacts.HTMLReport, Percent: 0.9})

	if req.CSVExport {
		artifacts.CSVDir = filepath.Join(outDir, "csv")
		if err := output.WriteCSV(outDir, bundle); err != nil {
			return "", 0, ArtifactPaths{}, fmt.Errorf("write csv: %w", err)
		}
	}
	if req.Summary {
		artifacts.SummaryHTML = filepath.Join(outDir, "recovery-summary.html")
		if err := output.WriteSummary(artifacts.SummaryHTML, bundle); err != nil {
			return "", 0, ArtifactPaths{}, fmt.Errorf("write summary: %w", err)
		}
	}
	if req.Runbook {
		artifacts.RunbookHTML = filepath.Join(outDir, "recovery-runbook.html")
		if err := output.WriteRunbook(artifacts.RunbookHTML, bundle); err != nil {
			return "", 0, ArtifactPaths{}, fmt.Errorf("write runbook: %w", err)
		}
	}
	if req.Redact {
		artifacts.RedactedJSON = filepath.Join(outDir, "recovery-scan-redacted.json")
		artifacts.RedactedHTML = filepath.Join(outDir, "recovery-report-redacted.html")
		if err := output.WriteRedactedJSON(artifacts.RedactedJSON, bundle); err != nil {
			return "", 0, ArtifactPaths{}, fmt.Errorf("write redacted json: %w", err)
		}
		if err := output.WriteRedactedReport(artifacts.RedactedHTML, bundle); err != nil {
			return "", 0, ArtifactPaths{}, fmt.Errorf("write redacted report: %w", err)
		}
	}

	return trendLabel, trendDelta, artifacts, nil
}

func exitCode(b *model.Bundle, minScore int) int {
	if b.Score.Overall.Final < minScore {
		return 2
	}
	return 0
}
