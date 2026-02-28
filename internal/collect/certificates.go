package collect

import (
	"context"
	"encoding/json"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
)

type certList struct {
	Items []struct {
		Metadata struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
		Spec struct {
			SecretName string `json:"secretName"`
			IssuerRef  struct {
				Name string `json:"name"`
				Kind string `json:"kind"`
			} `json:"issuerRef"`
		} `json:"spec"`
		Status struct {
			Conditions []struct {
				Type   string `json:"type"`
				Status string `json:"status"`
			} `json:"conditions"`
			NotAfter string `json:"notAfter"`
		} `json:"status"`
	} `json:"items"`
}

// Certificates collects cert-manager Certificate resources.
// Returns nil (non-fatal) if cert-manager is not installed.
func Certificates(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) error {
	raw, err := cs.RESTClient().
		Get().
		AbsPath("/apis/cert-manager.io/v1/certificates").
		DoRaw(ctx)
	if err != nil {
		// cert-manager not installed or not accessible — non-fatal
		return nil
	}

	var cl certList
	if err := json.Unmarshal(raw, &cl); err != nil {
		return nil // non-fatal parse failure
	}

	now := time.Now().UTC()
	for _, c := range cl.Items {
		if !InScope(c.Metadata.Namespace, b) {
			continue
		}
		ready := false
		for _, cond := range c.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				ready = true
				break
			}
		}
		daysToExpiry := 0
		if c.Status.NotAfter != "" {
			if expiry, err := time.Parse(time.RFC3339, c.Status.NotAfter); err == nil {
				daysToExpiry = int(expiry.Sub(now).Hours() / 24)
			}
		}
		issuer := c.Spec.IssuerRef.Kind + "/" + c.Spec.IssuerRef.Name
		b.Inventory.Certificates = append(b.Inventory.Certificates, model.Certificate{
			Namespace:    c.Metadata.Namespace,
			Name:         c.Metadata.Name,
			SecretName:   c.Spec.SecretName,
			Issuer:       issuer,
			Ready:        ready,
			NotAfter:     c.Status.NotAfter,
			DaysToExpiry: daysToExpiry,
		})
	}
	return nil
}
