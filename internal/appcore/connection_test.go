package appcore

import "testing"

func TestSanitizeScanRequestClearsIrrelevantConnectionFields(t *testing.T) {
	req := sanitizeScanRequest(ScanRequest{
		ConnectionMethod:  ConnectionMethodCurrent,
		ContextName:       " prod-east-admin ",
		KubeconfigPath:    "C:/Users/eric/.kube/config",
		KubeconfigContent: "apiVersion: v1",
		APIServerEndpoint: "https://cluster.example.net:6443",
		BearerToken:       "secret-token",
		CACertPath:        "C:/certs/cluster-ca.pem",
		CACertContent:     "-----BEGIN CERTIFICATE-----",
		Namespaces:        []string{" payments ", "frontend", "payments"},
	})

	if req.ContextName != "prod-east-admin" {
		t.Fatalf("ContextName = %q, want trimmed context", req.ContextName)
	}
	if req.KubeconfigPath != "" || req.KubeconfigContent != "" {
		t.Fatalf("expected kubeconfig fields to be cleared, got path=%q content=%q", req.KubeconfigPath, req.KubeconfigContent)
	}
	if req.APIServerEndpoint != "" || req.BearerToken != "" {
		t.Fatalf("expected API endpoint fields to be cleared, got endpoint=%q token=%q", req.APIServerEndpoint, req.BearerToken)
	}
	if len(req.Namespaces) != 2 || req.Namespaces[0] != "payments" || req.Namespaces[1] != "frontend" {
		t.Fatalf("Namespaces = %#v, want trimmed unique namespaces", req.Namespaces)
	}
}
