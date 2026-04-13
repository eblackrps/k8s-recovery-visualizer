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
	apiGroup        string
	resource        string
	namespaceScoped bool
	required        bool
	hint            string
	rules           []rbacRuleTemplate
}

type rbacRuleTemplate struct {
	apiGroups []string
	resources []string
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
		{id: "pods", title: "Read workloads", resource: "pods", namespaceScoped: true, required: true, hint: "Grant list access to pods, deployments, daemonsets, jobs, cronjobs, and statefulsets.", rules: []rbacRuleTemplate{{apiGroups: []string{""}, resources: []string{"pods"}}, {apiGroups: []string{"apps"}, resources: []string{"deployments", "daemonsets", "statefulsets"}}, {apiGroups: []string{"batch"}, resources: []string{"jobs", "cronjobs"}}}},
		{id: "storage", title: "Read storage inventory", resource: "persistentvolumeclaims", namespaceScoped: true, required: true, hint: "Grant PVC, PV, and StorageClass read access to build restore coverage.", rules: []rbacRuleTemplate{{apiGroups: []string{""}, resources: []string{"persistentvolumeclaims", "persistentvolumes"}}, {apiGroups: []string{"storage.k8s.io"}, resources: []string{"storageclasses"}}}},
		{id: "cluster", title: "Read cluster inventory", resource: "nodes", namespaceScoped: false, required: false, hint: "Node and namespace inventory improve topology and degraded-mode analysis.", rules: []rbacRuleTemplate{{apiGroups: []string{""}, resources: []string{"nodes", "namespaces"}}}},
		{id: "rbac", title: "Read RBAC inventory", apiGroup: "rbac.authorization.k8s.io", resource: "clusterroles", namespaceScoped: false, required: false, hint: "Cluster role visibility improves blast-radius and risk guidance.", rules: []rbacRuleTemplate{{apiGroups: []string{"rbac.authorization.k8s.io"}, resources: []string{"clusterroles", "clusterrolebindings"}}}},
		{id: "backup", title: "Inspect backup evidence", resource: "configmaps", namespaceScoped: true, required: false, hint: "Backup tool inspection may still be partial depending on product support.", rules: []rbacRuleTemplate{{apiGroups: []string{""}, resources: []string{"configmaps"}}, {apiGroups: []string{"velero.io"}, resources: []string{"schedules", "backups"}}, {apiGroups: []string{"config.kio.kasten.io"}, resources: []string{"policies", "profiles"}}, {apiGroups: []string{"longhorn.io"}, resources: []string{"recurringjobs", "backups"}}}},
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
		Scope:    probe.scopeLabel(),
		Resource: probe.resource,
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
		ok, reason, err := canList(ctx, clientset, probe.apiGroup, probe.resource, ns)
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
	check.Manifest = probe.manifest(targets)
	check.Commands = probe.commands(targets)
	if probe.required {
		check.Status = "fail"
		check.Detail = "Required access missing: " + strings.Join(failures, "; ")
		return check
	}
	check.Status = "warn"
	check.Detail = "Optional access missing: " + strings.Join(failures, "; ")
	return check
}

func (probe permissionProbe) scopeLabel() string {
	if probe.namespaceScoped {
		return "namespace"
	}
	return "cluster"
}

func (probe permissionProbe) manifest(namespaces []string) string {
	if len(probe.rules) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("apiVersion: rbac.authorization.k8s.io/v1\n")
	builder.WriteString("kind: ClusterRole\n")
	builder.WriteString("metadata:\n")
	builder.WriteString(fmt.Sprintf("  name: k8v-%s-reader\n", probe.id))
	builder.WriteString("rules:\n")
	for _, rule := range probe.rules {
		builder.WriteString(fmt.Sprintf("- apiGroups: [\"%s\"]\n", strings.Join(rule.apiGroups, "\", \"")))
		builder.WriteString(fmt.Sprintf("  resources: [\"%s\"]\n", strings.Join(rule.resources, "\", \"")))
		builder.WriteString("  verbs: [\"get\", \"list\", \"watch\"]\n")
	}
	if probe.namespaceScoped {
		targets := namespaces
		if len(targets) == 0 || (len(targets) == 1 && targets[0] == "") {
			targets = []string{"<namespace>"}
		}
		builder.WriteString("---\n")
		builder.WriteString("# Bind the ClusterRole with a RoleBinding in each allowed namespace.\n")
		builder.WriteString("apiVersion: rbac.authorization.k8s.io/v1\n")
		builder.WriteString("kind: RoleBinding\n")
		builder.WriteString("metadata:\n")
		builder.WriteString(fmt.Sprintf("  name: k8v-%s-reader\n", probe.id))
		builder.WriteString(fmt.Sprintf("  namespace: %s\n", targets[0]))
		builder.WriteString("subjects:\n")
		builder.WriteString("- kind: ServiceAccount\n")
		builder.WriteString("  name: <scanner-service-account>\n")
		builder.WriteString("  namespace: <service-account-namespace>\n")
		builder.WriteString("roleRef:\n")
		builder.WriteString("  apiGroup: rbac.authorization.k8s.io\n")
		builder.WriteString("  kind: ClusterRole\n")
		builder.WriteString(fmt.Sprintf("  name: k8v-%s-reader\n", probe.id))
		if len(targets) > 1 {
			builder.WriteString(fmt.Sprintf("# Repeat the RoleBinding for namespaces: %s\n", strings.Join(targets[1:], ", ")))
		}
		return builder.String()
	}
	builder.WriteString("---\n")
	builder.WriteString("apiVersion: rbac.authorization.k8s.io/v1\n")
	builder.WriteString("kind: ClusterRoleBinding\n")
	builder.WriteString("metadata:\n")
	builder.WriteString(fmt.Sprintf("  name: k8v-%s-reader\n", probe.id))
	builder.WriteString("subjects:\n")
	builder.WriteString("- kind: ServiceAccount\n")
	builder.WriteString("  name: <scanner-service-account>\n")
	builder.WriteString("  namespace: <service-account-namespace>\n")
	builder.WriteString("roleRef:\n")
	builder.WriteString("  apiGroup: rbac.authorization.k8s.io\n")
	builder.WriteString("  kind: ClusterRole\n")
	builder.WriteString(fmt.Sprintf("  name: k8v-%s-reader\n", probe.id))
	return builder.String()
}

func (probe permissionProbe) commands(namespaces []string) []string {
	target := ""
	if probe.namespaceScoped {
		target = " -n " + firstNamespace(namespaces)
	}
	return []string{
		fmt.Sprintf("kubectl auth can-i list %s%s", probe.resource, target),
		"# Save the manifest snippet below to rbac-reader.yaml and apply it after replacing the service account placeholders.",
		"kubectl apply -f rbac-reader.yaml",
	}
}

func firstNamespace(namespaces []string) string {
	for _, ns := range namespaces {
		if strings.TrimSpace(ns) != "" {
			return ns
		}
	}
	return "default"
}

func canList(ctx context.Context, clientset *kubernetes.Clientset, group, resource, namespace string) (bool, string, error) {
	review, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Verb:      "list",
				Group:     group,
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
