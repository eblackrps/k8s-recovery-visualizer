package kube

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// pickKubeconfigPath chooses the kubeconfig file to load.
// Priority:
//  1. explicitPath (flag)
//  2. KUBECONFIG env (first existing entry if multiple)
//  3. empty string (caller decides next steps)
func pickKubeconfigPath(explicitPath string) string {
	if strings.TrimSpace(explicitPath) != "" {
		return explicitPath
	}

	env := strings.TrimSpace(os.Getenv("KUBECONFIG"))
	if env == "" {
		return ""
	}

	// KUBECONFIG can contain multiple paths, separated by ';' on Windows and ':' on Linux.
	sep := ";"
	if strings.Contains(env, ":") && !strings.Contains(env, ";") {
		sep = ":"
	}

	for _, p := range strings.Split(env, sep) {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// No existing entry found, return the raw env so errors are descriptive.
	return env
}

// LoadConfig returns a Kubernetes rest.Config.
// It explicitly loads kubeconfig from file when a path is provided (or KUBECONFIG env is set),
// so failures produce real parse errors instead of "no configuration provided".
func LoadConfig(kubeconfigPath string) (*rest.Config, error) {
	chosen := pickKubeconfigPath(kubeconfigPath)

	// 1) If we have a kubeconfig path (explicit or env), load it explicitly.
	if strings.TrimSpace(chosen) != "" {
		abs := chosen
		if a, err := filepath.Abs(chosen); err == nil {
			abs = a
		}

		// Load raw kubeconfig from disk (gives better errors than deferred rules).
		rawCfg, err := clientcmd.LoadFromFile(abs)
		if err != nil {
			return nil, fmt.Errorf("load kube config: read kubeconfig file (path=%q): %w", abs, err)
		}

		// Optional context override via env, since cmd/scan currently only passes kubeconfig path.
		// (If you later add a -context flag to scan.exe, wire it in here.)
		overrides := &clientcmd.ConfigOverrides{}
		if ctx := strings.TrimSpace(os.Getenv("KUBE_CONTEXT")); ctx != "" {
			overrides.CurrentContext = ctx
		}

		// Build the rest.Config from the loaded kubeconfig.
		cfg, err := clientcmd.NewDefaultClientConfig(*rawCfg, overrides).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("load kube config: kubeconfig (path=%q currentContext=%q envKUBECONFIG=%q): %w",
				abs, rawCfg.CurrentContext, os.Getenv("KUBECONFIG"), err)
		}
		return cfg, nil
	}

	// 2) No kubeconfig path: try in-cluster.
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	// 3) Final fallback: default loading rules (HOME/.kube/config etc.)
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kube config: default rules: %w", err)
	}
	return cfg, nil
}

// NewClient matches cmd/scan/main.go usage:
//
//	clientset, cfg, err := kube.NewClient(kubeconfigPath)
func NewClient(kubeconfigPath string) (*kubernetes.Clientset, *rest.Config, error) {
	cfg, err := LoadConfig(kubeconfigPath)
	if err != nil {
		return nil, nil, err
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("create kube client: %w", err)
	}

	return cs, cfg, nil
}
