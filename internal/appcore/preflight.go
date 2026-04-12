package appcore

import (
	"context"
	"fmt"
	"strings"

	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s-recovery-visualizer/internal/kube"
)

type permissionProbe struct {
	id              string
	title           string
	resource        string
	namespaceScoped bool
	required        bool
	hint            string
}

func (s *Service) Preflight(ctx context.Context, req ScanRequest) (PreflightReport, error) {
	req = req.Normalized()
	scope := "all namespaces"
	if len(req.Namespaces) > 0 {
		scope = strings.Join(req.Namespaces, ", ")
	}
	if req.DryRun {
		return PreflightReport{
			CanRun:      true,
			Degraded:    false,
			ContextName: req.ContextName,
			Scope:       scope,
			Checks: []PreflightCheck{
				{ID: "mode", Title: "Dry-run mode", Status: "pass", Required: true, Detail: "No cluster connection is required."},
				{ID: "output", Title: "Output path", Status: "pass", Required: true, Detail: fmt.Sprintf("Artifacts will be written to %s.", req.OutputDir)},
			},
		}, nil
	}

	cfg, err := kube.LoadConfigWithContext(req.KubeconfigPath, req.ContextName)
	if err != nil {
		return PreflightReport{
			CanRun:      false,
			Degraded:    true,
			ContextName: req.ContextName,
			Scope:       scope,
			Checks: []PreflightCheck{
				{ID: "config", Title: "Kubernetes credentials", Status: "fail", Required: true, Detail: err.Error(), Hint: "Choose a kubeconfig file and context with cluster access."},
			},
		}, nil
	}

	if req.Insecure {
		cfg.TLSClientConfig.Insecure = true
		cfg.TLSClientConfig.CAFile = ""
		cfg.TLSClientConfig.CAData = nil
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return PreflightReport{}, fmt.Errorf("create kube client: %w", err)
	}

	checks := []PreflightCheck{
		{ID: "config", Title: "Kubernetes credentials", Status: "pass", Required: true, Detail: "Kubeconfig loaded successfully."},
	}
	warnings := []string{}
	canRun := true
	degraded := false

	discoveryCheck := PreflightCheck{
		ID:       "api",
		Title:    "API server reachability",
		Status:   "pass",
		Required: true,
		Detail:   fmt.Sprintf("Connected to %s.", cfg.Host),
	}
	if _, err := clientset.Discovery().ServerVersion(); err != nil {
		discoveryCheck.Status = "fail"
		discoveryCheck.Detail = err.Error()
		discoveryCheck.Hint = "Verify cluster reachability and kubeconfig credentials."
		canRun = false
		degraded = true
	}
	checks = append(checks, discoveryCheck)

	probes := []permissionProbe{
		{id: "pods", title: "Read workloads", resource: "pods", namespaceScoped: true, required: true, hint: "Grant list access to pods, deployments, daemonsets, jobs, cronjobs, and statefulsets."},
		{id: "storage", title: "Read storage inventory", resource: "persistentvolumeclaims", namespaceScoped: true, required: true, hint: "Grant PVC and PV list access to build restore coverage."},
		{id: "cluster", title: "Read cluster inventory", resource: "nodes", namespaceScoped: false, required: false, hint: "Node and cluster-scope reads improve topology and degraded-mode analysis."},
		{id: "rbac", title: "Read RBAC inventory", resource: "clusterroles", namespaceScoped: false, required: false, hint: "Cluster role visibility improves blast-radius and risk guidance."},
		{id: "backup", title: "Inspect backup evidence", resource: "configmaps", namespaceScoped: true, required: false, hint: "Backup tool inspection may still be partial depending on product support."},
	}
	for _, probe := range probes {
		check := probe.run(ctx, clientset, req.Namespaces)
		checks = append(checks, check)
		if check.Required && check.Status == "fail" {
			canRun = false
			degraded = true
		}
		if !check.Required && check.Status != "pass" {
			degraded = true
			warnings = append(warnings, check.Detail)
		}
	}

	return PreflightReport{
		CanRun:      canRun,
		Degraded:    degraded,
		Server:      cfg.Host,
		ContextName: req.ContextName,
		Scope:       scope,
		Checks:      checks,
		Warnings:    warnings,
	}, nil
}

func (probe permissionProbe) run(ctx context.Context, clientset *kubernetes.Clientset, namespaces []string) PreflightCheck {
	check := PreflightCheck{
		ID:       probe.id,
		Title:    probe.title,
		Status:   "pass",
		Required: probe.required,
		Detail:   "Access verified.",
		Hint:     probe.hint,
	}

	targets := []string{""}
	if probe.namespaceScoped {
		targets = namespaces
		if len(targets) == 0 {
			targets = []string{"default"}
		}
	}

	failures := []string{}
	for _, ns := range targets {
		ok, reason, err := canList(ctx, clientset, probe.resource, ns)
		if err != nil {
			check.Status = "warn"
			check.Detail = fmt.Sprintf("Permission probe could not be completed: %v", err)
			return check
		}
		if !ok {
			label := probe.resource
			if ns != "" {
				label = fmt.Sprintf("%s in %s", probe.resource, ns)
			}
			failures = append(failures, fmt.Sprintf("%s (%s)", label, strings.TrimSpace(reason)))
		}
	}

	if len(failures) == 0 {
		return check
	}
	if probe.required {
		check.Status = "fail"
		check.Detail = "Required access missing: " + strings.Join(failures, "; ")
		return check
	}
	check.Status = "warn"
	check.Detail = "Optional access missing: " + strings.Join(failures, "; ")
	return check
}

func canList(ctx context.Context, clientset *kubernetes.Clientset, resource, namespace string) (bool, string, error) {
	review, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Verb:      "list",
				Group:     "",
				Resource:  resource,
				Namespace: namespace,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return false, "", err
	}
	return review.Status.Allowed, review.Status.Reason, nil
}
