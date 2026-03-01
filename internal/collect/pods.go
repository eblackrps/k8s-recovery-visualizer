package collect

import (
	"context"

	"k8s-recovery-visualizer/internal/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Pods(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {

	list, err := cs.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range list.Items {
		if !InScope(pod.Namespace, b) {
			continue
		}

		usesHostPath := false
		for _, vol := range pod.Spec.Volumes {
			if vol.HostPath != nil {
				usesHostPath = true
				break
			}
		}

		// Round 11 — resource governance: check every container (including init)
		allContainers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
		allHaveRequests := true
		allHaveLimits := true
		for _, c := range allContainers {
			if !containerHasRequests(c) {
				allHaveRequests = false
			}
			if !containerHasLimits(c) {
				allHaveLimits = false
			}
		}

		// Round 12 — pod security: privileged containers
		hasPrivileged := false
		for _, c := range pod.Spec.Containers {
			if c.SecurityContext != nil &&
				c.SecurityContext.Privileged != nil &&
				*c.SecurityContext.Privileged {
				hasPrivileged = true
				break
			}
		}

		// Round 18 — ServiceAccount token: automount enabled when field is nil (default) or explicitly true
		automount := pod.Spec.AutomountServiceAccountToken == nil || *pod.Spec.AutomountServiceAccountToken

		b.Inventory.Pods = append(b.Inventory.Pods, model.Pod{
			Namespace:        pod.Namespace,
			Name:             pod.Name,
			UsesHostPath:     usesHostPath,
			ContainerCount:   len(pod.Spec.Containers),
			HasRequests:      allHaveRequests,
			HasLimits:        allHaveLimits,
			Privileged:       hasPrivileged,
			HostNetwork:      pod.Spec.HostNetwork,
			HostPID:          pod.Spec.HostPID,
			AutomountSAToken: automount,
		})
	}

	return nil
}

// containerHasRequests returns true when the container defines non-zero CPU and memory requests.
func containerHasRequests(c corev1.Container) bool {
	req := c.Resources.Requests
	if req == nil {
		return false
	}
	cpu, hasCPU := req[corev1.ResourceCPU]
	mem, hasMem := req[corev1.ResourceMemory]
	return hasCPU && !cpu.IsZero() && hasMem && !mem.IsZero()
}

// containerHasLimits returns true when the container defines non-zero CPU and memory limits.
func containerHasLimits(c corev1.Container) bool {
	lim := c.Resources.Limits
	if lim == nil {
		return false
	}
	cpu, hasCPU := lim[corev1.ResourceCPU]
	mem, hasMem := lim[corev1.ResourceMemory]
	return hasCPU && !cpu.IsZero() && hasMem && !mem.IsZero()
}
