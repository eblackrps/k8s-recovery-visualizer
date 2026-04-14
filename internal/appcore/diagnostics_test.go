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

func TestDiagnoseFailureMarksKubeconfigReachabilitySeparately(t *testing.T) {
	diagnosis := diagnoseFailure(
		ScanRequest{ConnectionMethod: ConnectionMethodKubeconfigFile},
		`Get "https://prod-api.internal:6443/version": dial tcp: lookup prod-api.internal: no such host`,
	)

	if diagnosis.Code != "endpoint_unreachable" {
		t.Fatalf("diagnoseFailure() code = %q, want endpoint_unreachable", diagnosis.Code)
	}
	if diagnosis.Label != "Cluster reachability" {
		t.Fatalf("diagnoseFailure() label = %q, want Cluster reachability", diagnosis.Label)
	}
	if diagnosis.Summary != "The kubeconfig was accepted, but the cluster API it points to is not reachable from this machine." {
		t.Fatalf("diagnoseFailure() summary = %q", diagnosis.Summary)
	}
	if diagnosis.NextAction == "" {
		t.Fatalf("diagnoseFailure() nextAction = empty, want actionable follow-up")
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

func TestReportConnectionFailureWarnsWhenAcceptedKubeconfigCannotReachCluster(t *testing.T) {
	report := reportConnectionFailure(
		ScanRequest{ConnectionMethod: ConnectionMethodKubeconfigFile},
		"kubeconfig file",
		"https://prod-api.internal:6443",
		"prod-east-admin",
		testError(`Get "https://prod-api.internal:6443/version": dial tcp: lookup prod-api.internal: no such host`),
	)

	if report.Diagnosis == nil || report.Diagnosis.Code != "endpoint_unreachable" {
		t.Fatalf("Diagnosis = %#v, want endpoint_unreachable", report.Diagnosis)
	}
	if report.FieldWarnings["kubeconfigPath"] == "" {
		t.Fatalf("FieldWarnings = %#v, want kubeconfigPath feedback", report.FieldWarnings)
	}
	if len(report.Checks) == 0 || report.Checks[len(report.Checks)-1].Title != "Cluster API reachability" {
		t.Fatalf("Checks = %#v, want Cluster API reachability check", report.Checks)
	}
}

func TestReportConnectionFailureExplainsLoopbackKubeconfigEndpoints(t *testing.T) {
	report := reportConnectionFailure(
		ScanRequest{ConnectionMethod: ConnectionMethodKubeconfigFile},
		"kubeconfig file",
		"https://127.0.0.1:6443",
		"prod-east-admin",
		testError(`Get "https://127.0.0.1:6443/version": dial tcp 127.0.0.1:6443: connectex: No connection could be made because the target machine actively refused it.`),
	)

	if report.Diagnosis == nil || report.Diagnosis.Label != "Loopback kubeconfig endpoint" {
		t.Fatalf("Diagnosis = %#v, want loopback kubeconfig diagnosis", report.Diagnosis)
	}
	if report.FieldWarnings["kubeconfigPath"] == "" {
		t.Fatalf("FieldWarnings = %#v, want kubeconfigPath warning", report.FieldWarnings)
	}
	if report.NextAction == "" || report.NextAction == report.Diagnosis.Detail {
		t.Fatalf("NextAction = %q, want actionable loopback guidance", report.NextAction)
	}
}

type testError string

func (e testError) Error() string {
	return string(e)
}
