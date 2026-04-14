package appcore

import "testing"

func TestScanRequestNormalizedPreservesExplicitZeroMinScore(t *testing.T) {
	req := ScanRequest{MinScore: 0}
	got := req.Normalized()
	if got.MinScore != 0 {
		t.Fatalf("expected min score 0 to be preserved, got %d", got.MinScore)
	}
}

func TestScanRequestNormalizedAppliesOtherDefaults(t *testing.T) {
	got := (ScanRequest{}).Normalized()
	if got.ConnectionMethod != ConnectionMethodCurrent {
		t.Fatalf("expected default connection method, got %q", got.ConnectionMethod)
	}
	if got.OutputDir != "./out" {
		t.Fatalf("expected default output dir, got %q", got.OutputDir)
	}
	if got.TimeoutSeconds != 60 {
		t.Fatalf("expected default timeout, got %d", got.TimeoutSeconds)
	}
	if got.Target != "vm" {
		t.Fatalf("expected default target, got %q", got.Target)
	}
	if got.ProfileName != "standard" {
		t.Fatalf("expected default profile, got %q", got.ProfileName)
	}
}

func TestScanRequestNormalizedInfersConnectionMethod(t *testing.T) {
	tests := []struct {
		name string
		req  ScanRequest
		want string
	}{
		{name: "file", req: ScanRequest{KubeconfigPath: "C:/kube/config"}, want: ConnectionMethodKubeconfigFile},
		{name: "inline", req: ScanRequest{KubeconfigContent: "apiVersion: v1"}, want: ConnectionMethodKubeconfigInline},
		{name: "endpoint", req: ScanRequest{APIServerEndpoint: "10.0.0.15:6443"}, want: ConnectionMethodAPIEndpoint},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.req.Normalized()
			if got.ConnectionMethod != tt.want {
				t.Fatalf("expected inferred connection method %q, got %q", tt.want, got.ConnectionMethod)
			}
		})
	}
}
