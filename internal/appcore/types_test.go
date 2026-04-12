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
