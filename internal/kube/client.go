package kube

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	ConnectionMethodCurrent          = "current"
	ConnectionMethodKubeconfigFile   = "kubeconfig_file"
	ConnectionMethodKubeconfigInline = "kubeconfig_inline"
	ConnectionMethodAPIEndpoint      = "api_endpoint"
)

type ConnectionOptions struct {
	Method            string
	KubeconfigPath    string
	KubeconfigContent string
	ContextName       string
	APIServerEndpoint string
	BearerToken       string
	CACertPath        string
	CACertContent     string
	Insecure          bool
}

type ResolvedConnection struct {
	Config      *rest.Config
	ContextName string
	Source      string
}

type ContextCatalog struct {
	Contexts       []string
	CurrentContext string
	Source         string
}

// pickKubeconfigPath chooses the kubeconfig file to load.
// Priority:
//  1. explicitPath
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

	return env
}

func LoadConfig(kubeconfigPath string) (*rest.Config, error) {
	return LoadConfigWithContext(kubeconfigPath, "")
}

func LoadConfigWithContext(kubeconfigPath, contextName string) (*rest.Config, error) {
	method := ConnectionMethodCurrent
	if strings.TrimSpace(kubeconfigPath) != "" {
		method = ConnectionMethodKubeconfigFile
	}
	resolved, err := ResolveConfig(ConnectionOptions{
		Method:         method,
		KubeconfigPath: kubeconfigPath,
		ContextName:    contextName,
	})
	if err != nil {
		return nil, err
	}
	return resolved.Config, nil
}

func NewClient(kubeconfigPath string, insecure bool) (*kubernetes.Clientset, *rest.Config, error) {
	return NewClientWithContext(kubeconfigPath, "", insecure)
}

func NewClientWithContext(kubeconfigPath, contextName string, insecure bool) (*kubernetes.Clientset, *rest.Config, error) {
	method := ConnectionMethodCurrent
	if strings.TrimSpace(kubeconfigPath) != "" {
		method = ConnectionMethodKubeconfigFile
	}
	clientset, resolved, err := NewClientFromOptions(ConnectionOptions{
		Method:         method,
		KubeconfigPath: kubeconfigPath,
		ContextName:    contextName,
		Insecure:       insecure,
	})
	if err != nil {
		return nil, nil, err
	}
	return clientset, resolved.Config, nil
}

func NewClientFromOptions(options ConnectionOptions) (*kubernetes.Clientset, *ResolvedConnection, error) {
	resolved, err := ResolveConfig(options)
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(resolved.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("create kube client: %w", err)
	}

	return clientset, resolved, nil
}

func ListContexts(options ConnectionOptions) (ContextCatalog, error) {
	switch normalizeConnectionMethod(options) {
	case ConnectionMethodKubeconfigFile:
		return listContextsFromFile(options)
	case ConnectionMethodKubeconfigInline:
		return listContextsFromInline(options)
	case ConnectionMethodAPIEndpoint:
		return ContextCatalog{Source: "direct API endpoint"}, nil
	default:
		return listContextsFromCurrent()
	}
}

func ResolveConfig(options ConnectionOptions) (*ResolvedConnection, error) {
	switch normalizeConnectionMethod(options) {
	case ConnectionMethodKubeconfigFile:
		return resolveFromFile(options)
	case ConnectionMethodKubeconfigInline:
		return resolveFromInlineKubeconfig(options)
	case ConnectionMethodAPIEndpoint:
		return resolveFromAPIEndpoint(options)
	default:
		return resolveCurrent(options)
	}
}

func normalizeConnectionMethod(options ConnectionOptions) string {
	switch strings.TrimSpace(options.Method) {
	case ConnectionMethodCurrent, ConnectionMethodKubeconfigFile, ConnectionMethodKubeconfigInline, ConnectionMethodAPIEndpoint:
		return strings.TrimSpace(options.Method)
	}
	switch {
	case strings.TrimSpace(options.APIServerEndpoint) != "":
		return ConnectionMethodAPIEndpoint
	case strings.TrimSpace(options.KubeconfigContent) != "":
		return ConnectionMethodKubeconfigInline
	case strings.TrimSpace(options.KubeconfigPath) != "":
		return ConnectionMethodKubeconfigFile
	default:
		return ConnectionMethodCurrent
	}
}

func resolveFromFile(options ConnectionOptions) (*ResolvedConnection, error) {
	rawCfg, _, err := loadRawConfigFromFile(options.KubeconfigPath)
	if err != nil {
		return nil, err
	}

	return resolveFromRawConfig(*rawCfg, options, "kubeconfig file")
}

func resolveFromInlineKubeconfig(options ConnectionOptions) (*ResolvedConnection, error) {
	content := strings.TrimSpace(options.KubeconfigContent)
	if content == "" {
		return nil, fmt.Errorf("paste a kubeconfig or switch connection mode")
	}

	rawCfg, err := clientcmd.Load([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("load pasted kubeconfig: %w", err)
	}

	return resolveFromRawConfig(*rawCfg, options, "pasted kubeconfig")
}

func resolveCurrent(options ConnectionOptions) (*ResolvedConnection, error) {
	chosen := pickKubeconfigPath("")
	if strings.TrimSpace(chosen) != "" {
		resolved, err := resolveFromFile(ConnectionOptions{
			KubeconfigPath: chosen,
			ContextName:    options.ContextName,
			Insecure:       options.Insecure,
		})
		if err != nil {
			return nil, err
		}
		resolved.Source = "current Kubernetes login"
		return resolved, nil
	}

	if cfg, err := rest.InClusterConfig(); err == nil {
		applyTLSOverrides(cfg, options)
		return &ResolvedConnection{
			Config: cfg,
			Source: "in-cluster service account",
		}, nil
	}

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := configOverrides(options.ContextName)
	loadingConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
	cfg, err := loadingConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load current Kubernetes login: %w", err)
	}

	rawCfg, err := loadingConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("read current Kubernetes login: %w", err)
	}

	applyTLSOverrides(cfg, options)
	return &ResolvedConnection{
		Config:      cfg,
		ContextName: resolvedContextName(rawCfg, options.ContextName),
		Source:      "current Kubernetes login",
	}, nil
}

func resolveFromAPIEndpoint(options ConnectionOptions) (*ResolvedConnection, error) {
	endpoint := normalizeEndpoint(options.APIServerEndpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("enter the Kubernetes API server host or endpoint")
	}
	if strings.TrimSpace(options.BearerToken) == "" {
		return nil, fmt.Errorf("enter a bearer token or switch to kubeconfig mode")
	}

	cfg := &rest.Config{
		Host:        endpoint,
		BearerToken: strings.TrimSpace(options.BearerToken),
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: strings.TrimSpace(options.CACertPath) == "" && strings.TrimSpace(options.CACertContent) == "" && options.Insecure,
		},
	}

	if caPath := strings.TrimSpace(options.CACertPath); caPath != "" {
		abs := caPath
		if a, err := filepath.Abs(caPath); err == nil {
			abs = a
		}
		cfg.TLSClientConfig.CAFile = abs
	}
	if caContent := strings.TrimSpace(options.CACertContent); caContent != "" {
		cfg.TLSClientConfig.CAData = []byte(caContent)
	}

	applyTLSOverrides(cfg, options)
	return &ResolvedConnection{
		Config: cfg,
		Source: "direct API endpoint",
	}, nil
}

func resolveFromRawConfig(rawCfg clientcmdapi.Config, options ConnectionOptions, source string) (*ResolvedConnection, error) {
	overrides := configOverrides(options.ContextName)
	cfg, err := clientcmd.NewDefaultClientConfig(rawCfg, overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kube config from %s: %w", source, err)
	}

	applyTLSOverrides(cfg, options)
	return &ResolvedConnection{
		Config:      cfg,
		ContextName: resolvedContextName(rawCfg, options.ContextName),
		Source:      source,
	}, nil
}

func listContextsFromFile(options ConnectionOptions) (ContextCatalog, error) {
	rawCfg, _, err := loadRawConfigFromFile(options.KubeconfigPath)
	if err != nil {
		return ContextCatalog{}, err
	}
	return buildContextCatalog(*rawCfg, "kubeconfig file"), nil
}

func listContextsFromInline(options ConnectionOptions) (ContextCatalog, error) {
	content := strings.TrimSpace(options.KubeconfigContent)
	if content == "" {
		return ContextCatalog{}, fmt.Errorf("paste a kubeconfig or switch connection mode")
	}
	rawCfg, err := clientcmd.Load([]byte(content))
	if err != nil {
		return ContextCatalog{}, fmt.Errorf("load pasted kubeconfig: %w", err)
	}
	return buildContextCatalog(*rawCfg, "pasted kubeconfig"), nil
}

func listContextsFromCurrent() (ContextCatalog, error) {
	chosen := pickKubeconfigPath("")
	if strings.TrimSpace(chosen) != "" {
		rawCfg, _, err := loadRawConfigFromFile(chosen)
		if err != nil {
			return ContextCatalog{}, err
		}
		catalog := buildContextCatalog(*rawCfg, "current Kubernetes login")
		if catalog.CurrentContext == "" {
			catalog.CurrentContext = rawCfg.CurrentContext
		}
		return catalog, nil
	}

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rawCfg, err := rules.Load()
	if err == nil && rawCfg != nil {
		catalog := buildContextCatalog(*rawCfg, "current Kubernetes login")
		if catalog.CurrentContext == "" {
			catalog.CurrentContext = rawCfg.CurrentContext
		}
		if len(catalog.Contexts) > 0 || catalog.CurrentContext != "" {
			return catalog, nil
		}
	}

	if _, err := rest.InClusterConfig(); err == nil {
		return ContextCatalog{Source: "in-cluster service account"}, nil
	}
	return ContextCatalog{Source: "current Kubernetes login"}, nil
}

func configOverrides(contextName string) *clientcmd.ConfigOverrides {
	overrides := &clientcmd.ConfigOverrides{}
	resolvedContext := strings.TrimSpace(contextName)
	if resolvedContext == "" {
		resolvedContext = strings.TrimSpace(os.Getenv("KUBE_CONTEXT"))
	}
	if resolvedContext != "" {
		overrides.CurrentContext = resolvedContext
	}
	return overrides
}

func resolvedContextName(rawCfg clientcmdapi.Config, explicitContext string) string {
	if ctx := strings.TrimSpace(explicitContext); ctx != "" {
		return ctx
	}
	if ctx := strings.TrimSpace(os.Getenv("KUBE_CONTEXT")); ctx != "" {
		return ctx
	}
	return rawCfg.CurrentContext
}

func normalizeEndpoint(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, "://") {
		return trimmed
	}
	return "https://" + trimmed
}

func applyTLSOverrides(cfg *rest.Config, options ConnectionOptions) {
	if cfg == nil {
		return
	}
	if options.Insecure {
		cfg.TLSClientConfig.Insecure = true
		cfg.TLSClientConfig.CAFile = ""
		cfg.TLSClientConfig.CAData = nil
	}
}

func loadRawConfigFromFile(kubeconfigPath string) (*clientcmdapi.Config, string, error) {
	chosen := pickKubeconfigPath(kubeconfigPath)
	if strings.TrimSpace(chosen) == "" {
		return nil, "", fmt.Errorf("choose a kubeconfig file or switch to current login")
	}

	abs := chosen
	if a, err := filepath.Abs(chosen); err == nil {
		abs = a
	}

	rawCfg, err := clientcmd.LoadFromFile(abs)
	if err != nil {
		return nil, "", fmt.Errorf("load kube config from %q: %w", abs, err)
	}
	return rawCfg, abs, nil
}

func buildContextCatalog(rawCfg clientcmdapi.Config, source string) ContextCatalog {
	contexts := make([]string, 0, len(rawCfg.Contexts))
	for name := range rawCfg.Contexts {
		contexts = append(contexts, name)
	}
	sort.Strings(contexts)
	return ContextCatalog{
		Contexts:       contexts,
		CurrentContext: rawCfg.CurrentContext,
		Source:         source,
	}
}
