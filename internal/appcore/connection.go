package appcore

import (
	"net/url"
	"strings"

	"k8s-recovery-visualizer/internal/kube"
)

func sanitizeScanRequest(req ScanRequest) ScanRequest {
	req = req.Normalized()
	req.KubeconfigPath = strings.TrimSpace(req.KubeconfigPath)
	req.KubeconfigContent = strings.TrimSpace(req.KubeconfigContent)
	req.ContextName = strings.TrimSpace(req.ContextName)
	req.APIServerEndpoint = strings.TrimSpace(req.APIServerEndpoint)
	req.BearerToken = strings.TrimSpace(req.BearerToken)
	req.CACertPath = strings.TrimSpace(req.CACertPath)
	req.CACertContent = strings.TrimSpace(req.CACertContent)
	req.OutputDir = strings.TrimSpace(req.OutputDir)
	req.CustomerID = strings.TrimSpace(req.CustomerID)
	req.Site = strings.TrimSpace(req.Site)
	req.ClusterName = strings.TrimSpace(req.ClusterName)
	req.Environment = strings.TrimSpace(req.Environment)
	req.Target = strings.TrimSpace(req.Target)
	req.CompareTo = strings.TrimSpace(req.CompareTo)
	req.ProfileName = strings.TrimSpace(req.ProfileName)
	req.Namespaces = sanitizeStrings(req.Namespaces)

	switch req.ConnectionMethod {
	case ConnectionMethodKubeconfigFile:
		req.KubeconfigContent = ""
		req.APIServerEndpoint = ""
		req.BearerToken = ""
		req.CACertPath = ""
		req.CACertContent = ""
	case ConnectionMethodKubeconfigInline:
		req.KubeconfigPath = ""
		req.APIServerEndpoint = ""
		req.BearerToken = ""
		req.CACertPath = ""
		req.CACertContent = ""
	case ConnectionMethodAPIEndpoint:
		req.KubeconfigPath = ""
		req.KubeconfigContent = ""
		req.ContextName = ""
	default:
		req.KubeconfigPath = ""
		req.KubeconfigContent = ""
		req.APIServerEndpoint = ""
		req.BearerToken = ""
		req.CACertPath = ""
		req.CACertContent = ""
	}

	return req
}

func kubeOptionsFromRequest(req ScanRequest) kube.ConnectionOptions {
	return kube.ConnectionOptions{
		Method:            req.ConnectionMethod,
		KubeconfigPath:    req.KubeconfigPath,
		KubeconfigContent: req.KubeconfigContent,
		ContextName:       req.ContextName,
		APIServerEndpoint: req.APIServerEndpoint,
		BearerToken:       req.BearerToken,
		CACertPath:        req.CACertPath,
		CACertContent:     req.CACertContent,
		Insecure:          req.Insecure,
	}
}

func connectionHint(req ScanRequest) string {
	switch req.Normalized().ConnectionMethod {
	case ConnectionMethodKubeconfigFile:
		return "Choose a kubeconfig file and, if needed, a context with cluster access."
	case ConnectionMethodKubeconfigInline:
		return "Paste a kubeconfig and, if needed, set the context you want to scan."
	case ConnectionMethodAPIEndpoint:
		return "Enter the control-plane host or endpoint and add a bearer token, or switch to kubeconfig mode."
	default:
		return "Leave kubeconfig blank only when kubectl already works on this machine or KUBECONFIG is set."
	}
}

func connectionStatusDetail(source string) string {
	switch source {
	case "kubeconfig file":
		return "Kubeconfig file loaded successfully."
	case "pasted kubeconfig":
		return "Pasted kubeconfig loaded successfully."
	case "direct API endpoint":
		return "Direct API connection details loaded successfully."
	case "in-cluster service account":
		return "In-cluster service account credentials loaded successfully."
	default:
		return "Current Kubernetes login loaded successfully."
	}
}

func inferredClusterName(req ScanRequest, apiServerEndpoint, resolvedContext string) string {
	if label := strings.TrimSpace(req.ClusterName); label != "" {
		return label
	}
	if label := strings.TrimSpace(resolvedContext); label != "" {
		return label
	}
	return hostLabel(apiServerEndpoint)
}

func sanitizeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func hostLabel(apiServerEndpoint string) string {
	trimmed := strings.TrimSpace(apiServerEndpoint)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}
	if host := parsed.Hostname(); host != "" {
		return host
	}
	return trimmed
}
