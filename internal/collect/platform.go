package collect

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func Platform(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	platform := model.Platform{
		Provider: "vanilla",
	}

	// Detect from nodes
	nodes, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		b.Cluster.Platform = platform
		return nil // non-fatal
	}

	for _, n := range nodes.Items {
		labels := n.Labels
		kubeletVer := n.Status.NodeInfo.KubeletVersion

		// Capture K8s version from first node
		if platform.K8sVersion == "" {
			platform.K8sVersion = kubeletVer
		}

		switch {
		case labels["eks.amazonaws.com/nodegroup"] != "" ||
			labels["alpha.eksctl.io/nodegroup-name"] != "":
			platform.Provider = "EKS"
		case labels["kubernetes.azure.com/agentpool"] != "":
			platform.Provider = "AKS"
		case labels["cloud.google.com/gke-nodepool"] != "":
			platform.Provider = "GKE"
		case labels["cattle.io/creator"] != "" ||
			hasKeyPrefix(labels, "node-role.cattle.io/"):
			platform.Provider = "Rancher"
		case strings.Contains(strings.ToLower(kubeletVer), "k3s"):
			platform.Provider = "k3s"
		}

		if platform.Provider != "vanilla" {
			break
		}
	}

	// Capture cluster UID from kube-system namespace
	if ns, err := cs.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{}); err == nil {
		platform.ClusterUID = string(ns.UID)
	}

	b.Cluster.Platform = platform
	return nil
}

func hasKeyPrefix(labels map[string]string, prefix string) bool {
	for k := range labels {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}
