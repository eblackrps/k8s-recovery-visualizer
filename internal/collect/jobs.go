package collect

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

func Jobs(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, j := range list.Items {
		if !InScope(j.Namespace, b) {
			continue
		}
		completed := j.Status.CompletionTime != nil
		b.Inventory.Jobs = append(b.Inventory.Jobs, model.Job{
			Namespace: j.Namespace,
			Name:      j.Name,
			Succeeded: j.Status.Succeeded,
			Failed:    j.Status.Failed,
			Active:    j.Status.Active,
			Completed: completed,
		})
	}
	return nil
}

func CronJobs(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	list, err := cs.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, cj := range list.Items {
		if !InScope(cj.Namespace, b) {
			continue
		}
		lastRun := ""
		if cj.Status.LastScheduleTime != nil {
			lastRun = cj.Status.LastScheduleTime.UTC().Format("2006-01-02T15:04:05Z")
		}
		b.Inventory.CronJobs = append(b.Inventory.CronJobs, model.CronJob{
			Namespace:   cj.Namespace,
			Name:        cj.Name,
			Schedule:    cj.Spec.Schedule,
			Suspended:   cj.Spec.Suspend != nil && *cj.Spec.Suspend,
			LastRunTime: lastRun,
			ActiveJobs:  len(cj.Status.Active),
		})
	}
	return nil
}
