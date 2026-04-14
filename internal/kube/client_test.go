package kube

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleKubeconfig = `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://alpha.example.net:6443
  name: alpha
- cluster:
    server: https://beta.example.net:6443
  name: beta
contexts:
- context:
    cluster: alpha
    user: ops
  name: alpha-admin
- context:
    cluster: beta
    user: ops
  name: beta-admin
current-context: alpha-admin
users:
- name: ops
  user:
    token: demo-token
`

const sampleKubeconfigWithFileRefs = `apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority: ca/cluster-ca.crt
    server: https://alpha.example.net:6443
  name: alpha
contexts:
- context:
    cluster: alpha
    user: ops
  name: alpha-admin
current-context: alpha-admin
users:
- name: ops
  user:
    client-certificate: auth/user.crt
    client-key: auth/user.key
`

func TestResolveConfigWithInlineKubeconfigAndContextOverride(t *testing.T) {
	resolved, err := ResolveConfig(ConnectionOptions{
		Method:            ConnectionMethodKubeconfigInline,
		KubeconfigContent: sampleKubeconfig,
		ContextName:       "beta-admin",
	})
	if err != nil {
		t.Fatalf("ResolveConfig() error = %v", err)
	}
	if resolved.ContextName != "beta-admin" {
		t.Fatalf("resolved.ContextName = %q, want beta-admin", resolved.ContextName)
	}
	if resolved.Config.Host != "https://beta.example.net:6443" {
		t.Fatalf("resolved.Config.Host = %q, want beta endpoint", resolved.Config.Host)
	}
}

func TestResolveConfigWithAPIEndpointNormalizesHost(t *testing.T) {
	resolved, err := ResolveConfig(ConnectionOptions{
		Method:            ConnectionMethodAPIEndpoint,
		APIServerEndpoint: "10.0.0.15:6443",
		BearerToken:       "demo-token",
		CACertContent:     "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
	})
	if err != nil {
		t.Fatalf("ResolveConfig() error = %v", err)
	}
	if resolved.Config.Host != "https://10.0.0.15:6443" {
		t.Fatalf("resolved.Config.Host = %q, want normalized https endpoint", resolved.Config.Host)
	}
	if resolved.Config.BearerToken != "demo-token" {
		t.Fatalf("resolved.Config.BearerToken = %q, want demo-token", resolved.Config.BearerToken)
	}
	if len(resolved.Config.TLSClientConfig.CAData) == 0 {
		t.Fatal("expected CAData to be populated")
	}
}

func TestResolveConfigCurrentUsesKubeconfigEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(sampleKubeconfig), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("KUBECONFIG", path)
	t.Setenv("KUBE_CONTEXT", "")

	resolved, err := ResolveConfig(ConnectionOptions{Method: ConnectionMethodCurrent})
	if err != nil {
		t.Fatalf("ResolveConfig() error = %v", err)
	}
	if resolved.Source != "current Kubernetes login" {
		t.Fatalf("resolved.Source = %q, want current Kubernetes login", resolved.Source)
	}
	if resolved.Config.Host != "https://alpha.example.net:6443" {
		t.Fatalf("resolved.Config.Host = %q, want alpha endpoint", resolved.Config.Host)
	}
}

func TestResolveConfigRejectsEmptyFileMode(t *testing.T) {
	_, err := ResolveConfig(ConnectionOptions{Method: ConnectionMethodKubeconfigFile})
	if err == nil {
		t.Fatal("expected error for empty kubeconfig file mode")
	}
	if !strings.Contains(err.Error(), "choose a kubeconfig file") {
		t.Fatalf("error = %v, want kubeconfig guidance", err)
	}
}

func TestResolveConfigRejectsAPIEndpointWithoutToken(t *testing.T) {
	_, err := ResolveConfig(ConnectionOptions{
		Method:            ConnectionMethodAPIEndpoint,
		APIServerEndpoint: "10.0.0.15:6443",
	})
	if err == nil {
		t.Fatal("expected error for API endpoint mode without a token")
	}
	if !strings.Contains(err.Error(), "bearer token") {
		t.Fatalf("error = %v, want bearer token guidance", err)
	}
}

func TestListContextsWithInlineKubeconfig(t *testing.T) {
	catalog, err := ListContexts(ConnectionOptions{
		Method:            ConnectionMethodKubeconfigInline,
		KubeconfigContent: sampleKubeconfig,
	})
	if err != nil {
		t.Fatalf("ListContexts() error = %v", err)
	}
	if catalog.CurrentContext != "alpha-admin" {
		t.Fatalf("catalog.CurrentContext = %q, want alpha-admin", catalog.CurrentContext)
	}
	if len(catalog.Contexts) != 2 || catalog.Contexts[0] != "alpha-admin" || catalog.Contexts[1] != "beta-admin" {
		t.Fatalf("catalog.Contexts = %#v, want sorted contexts", catalog.Contexts)
	}
}

func TestInspectKubeconfigFileAcceptsBackupExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prod-cluster.backup")
	if err := os.WriteFile(path, []byte(sampleKubeconfig), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	inspection, err := InspectKubeconfigFile(path)
	if err != nil {
		t.Fatalf("InspectKubeconfigFile() error = %v", err)
	}
	if inspection.CurrentContext != "alpha-admin" {
		t.Fatalf("inspection.CurrentContext = %q, want alpha-admin", inspection.CurrentContext)
	}
	if inspection.ClusterCount != 2 || inspection.UserCount != 1 {
		t.Fatalf("unexpected inspection counts: %#v", inspection)
	}
}

func TestInspectKubeconfigFileAcceptsNoExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(path, []byte(sampleKubeconfig), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	inspection, err := InspectKubeconfigFile(path)
	if err != nil {
		t.Fatalf("InspectKubeconfigFile() error = %v", err)
	}
	if len(inspection.Contexts) != 2 {
		t.Fatalf("inspection.Contexts = %#v, want two contexts", inspection.Contexts)
	}
}

func TestInspectKubeconfigFileReportsMissingReferencedFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "ca"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "auth"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ca", "cluster-ca.crt"), []byte("demo-ca"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "auth", "user.crt"), []byte("demo-cert"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	path := filepath.Join(dir, "prod-cluster.backup")
	if err := os.WriteFile(path, []byte(sampleKubeconfigWithFileRefs), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	inspection, err := InspectKubeconfigFile(path)
	if err != nil {
		t.Fatalf("InspectKubeconfigFile() error = %v", err)
	}
	if len(inspection.ReferencedFiles) != 3 {
		t.Fatalf("inspection.ReferencedFiles = %#v, want three references", inspection.ReferencedFiles)
	}
	if len(inspection.MissingReferencedFiles) != 1 {
		t.Fatalf("inspection.MissingReferencedFiles = %#v, want one missing file", inspection.MissingReferencedFiles)
	}
	if !strings.Contains(inspection.MissingReferencedFiles[0], filepath.Join(dir, "auth", "user.key")) {
		t.Fatalf("missing reference = %q, want resolved user.key path", inspection.MissingReferencedFiles[0])
	}
	if !strings.Contains(inspection.NextAction, "self-contained kubeconfig") {
		t.Fatalf("inspection.NextAction = %q, want self-contained kubeconfig guidance", inspection.NextAction)
	}
}

func TestInspectKubeconfigContentWarnsAboutReferencedFiles(t *testing.T) {
	inspection, err := InspectKubeconfigContent(sampleKubeconfigWithFileRefs)
	if err != nil {
		t.Fatalf("InspectKubeconfigContent() error = %v", err)
	}
	if len(inspection.ReferencedFiles) != 3 {
		t.Fatalf("inspection.ReferencedFiles = %#v, want three references", inspection.ReferencedFiles)
	}
	if len(inspection.MissingReferencedFiles) != 0 {
		t.Fatalf("inspection.MissingReferencedFiles = %#v, want no missing paths for pasted content", inspection.MissingReferencedFiles)
	}
	if !strings.Contains(inspection.NextAction, "Paste mode") {
		t.Fatalf("inspection.NextAction = %q, want paste mode guidance", inspection.NextAction)
	}
}

func TestApplyInspectionToConnectionAdvisorFlagsIncompleteLocalKubeconfig(t *testing.T) {
	advisor := applyInspectionToConnectionAdvisor(ConnectionAdvisor{}, "C:/Users/demo/.kube/prod-cluster.backup", KubeconfigInspection{
		CurrentContext:         "prod-east-admin",
		Summary:                "Loaded kubeconfig file with 1 context.",
		MissingReferencedFiles: []string{"User prod-east client key: C:\\Users\\demo\\.kube\\user.key"},
	})

	if advisor.CurrentLoginAvailable {
		t.Fatal("expected current login to be marked unavailable when referenced files are missing")
	}
	if advisor.RecommendedMethod != ConnectionMethodKubeconfigFile {
		t.Fatalf("advisor.RecommendedMethod = %q, want kubeconfig_file", advisor.RecommendedMethod)
	}
	if !strings.Contains(advisor.CurrentLoginWarning, "missing on this machine") {
		t.Fatalf("advisor.CurrentLoginWarning = %q, want missing-files guidance", advisor.CurrentLoginWarning)
	}
}

func TestApplyInspectionToConnectionAdvisorWarnsAboutExecAuth(t *testing.T) {
	advisor := applyInspectionToConnectionAdvisor(ConnectionAdvisor{}, "C:/Users/demo/.kube/config", KubeconfigInspection{
		CurrentContext: "prod-east-admin",
		Summary:        "Loaded kubeconfig file with 1 context.",
		UsesExecAuth:   true,
	})

	if !advisor.CurrentLoginAvailable {
		t.Fatal("expected current login to stay available when exec auth is detected")
	}
	if advisor.RecommendedMethod != ConnectionMethodCurrent {
		t.Fatalf("advisor.RecommendedMethod = %q, want current", advisor.RecommendedMethod)
	}
	if !strings.Contains(advisor.CurrentLoginWarning, "external auth helper") {
		t.Fatalf("advisor.CurrentLoginWarning = %q, want exec-helper guidance", advisor.CurrentLoginWarning)
	}
}
