package appcore

import "testing"

func TestDiagnoseFailureCategorizesCommonOperatorIssues(t *testing.T) {
	tests := []struct {
		name    string
		req     ScanRequest
		message string
		want    string
	}{
		{
			name:    "tls trust",
			req:     ScanRequest{ConnectionMethod: ConnectionMethodAPIEndpoint},
			message: "Get \"https://cluster.example.net:6443/version\": x509: certificate signed by unknown authority",
			want:    "tls_trust",
		},
		{
			name:    "endpoint unreachable",
			req:     ScanRequest{ConnectionMethod: ConnectionMethodAPIEndpoint},
			message: "Get \"https://cluster.example.net:6443/version\": dial tcp: lookup cluster.example.net: no such host",
			want:    "endpoint_unreachable",
		},
		{
			name:    "auth helper",
			req:     ScanRequest{ConnectionMethod: ConnectionMethodKubeconfigFile},
			message: "load kube config from \"C:\\temp\\cluster\": exec plugin failed",
			want:    "auth_helper",
		},
		{
			name:    "rbac denied",
			req:     ScanRequest{ConnectionMethod: ConnectionMethodAPIEndpoint},
			message: "pods is forbidden: User \"system:serviceaccount:k8v:scanner\" cannot list resource \"pods\"",
			want:    "rbac_denied",
		},
		{
			name:    "output path",
			req:     ScanRequest{},
			message: "create output directory: access is denied",
			want:    "output_path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnosis := diagnoseFailure(tt.req, tt.message)
			if diagnosis.Code != tt.want {
				t.Fatalf("diagnoseFailure() code = %q, want %q", diagnosis.Code, tt.want)
			}
			if diagnosis.Summary == "" || diagnosis.NextAction == "" {
				t.Fatalf("diagnoseFailure() = %#v, want summary and next action", diagnosis)
			}
		})
	}
}

func TestReportConnectionFailureAppliesFieldFeedbackForKnownFailures(t *testing.T) {
	report := reportConnectionFailure(
		ScanRequest{ConnectionMethod: ConnectionMethodAPIEndpoint},
		"direct API endpoint",
		"https://cluster.example.net:6443",
		"",
		testError("x509: certificate signed by unknown authority"),
	)

	if report.Diagnosis == nil || report.Diagnosis.Code != "tls_trust" {
		t.Fatalf("Diagnosis = %#v, want tls_trust", report.Diagnosis)
	}
	if report.FieldErrors["caTrust"] == "" {
		t.Fatalf("FieldErrors = %#v, want caTrust feedback", report.FieldErrors)
	}
}

type testError string

func (e testError) Error() string {
	return string(e)
}
