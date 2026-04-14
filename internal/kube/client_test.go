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
