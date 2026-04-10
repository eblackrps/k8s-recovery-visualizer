package scanapp

import (
	"context"
	"log"
	"strings"

	"k8s-recovery-visualizer/internal/collect"
	"k8s-recovery-visualizer/internal/model"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type collectorStep struct {
	Name     string
	Required bool
	Run      func() error
}

func runCollectors(ctx context.Context, cs *kubernetes.Clientset, dc dynamic.Interface, bundle *model.Bundle) error {
	for _, step := range buildCollectorPipeline(ctx, cs, dc, bundle) {
		if err := step.Run(); err != nil {
			if step.Required {
				return err
			}
			recordCollectorSkip(step.Name, err, bundle)
		}
	}
	return nil
}

func buildCollectorPipeline(ctx context.Context, cs *kubernetes.Clientset, dc dynamic.Interface, bundle *model.Bundle) []collectorStep {
	return []collectorStep{
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
		{Name: "Secrets", Run: func() error { return collect.Secrets(ctx, cs, bundle) }},
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
}

func recordCollectorSkip(name string, err error, bundle *model.Bundle) {
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
