package kube

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type KubeconfigInspection struct {
	Source                       string   `json:"source,omitempty"`
	Path                         string   `json:"path,omitempty"`
	CurrentContext               string   `json:"currentContext,omitempty"`
	Contexts                     []string `json:"contexts,omitempty"`
	ClusterCount                 int      `json:"clusterCount"`
	UserCount                    int      `json:"userCount"`
	UsesExecAuth                 bool     `json:"usesExecAuth,omitempty"`
	UsesClientCertificate        bool     `json:"usesClientCertificate,omitempty"`
	UsesCertificateAuthorityFile bool     `json:"usesCertificateAuthorityFile,omitempty"`
	UsesCertificateAuthorityData bool     `json:"usesCertificateAuthorityData,omitempty"`
	Servers                      []string `json:"servers,omitempty"`
	LoopbackServers              []string `json:"loopbackServers,omitempty"`
	ReferencedFiles              []string `json:"referencedFiles,omitempty"`
	MissingReferencedFiles       []string `json:"missingReferencedFiles,omitempty"`
	Summary                      string   `json:"summary,omitempty"`
	NextAction                   string   `json:"nextAction,omitempty"`
}

type ConnectionAdvisor struct {
	RecommendedMethod               string `json:"recommendedMethod,omitempty"`
	RecommendedReason               string `json:"recommendedReason,omitempty"`
	KubectlAvailable                bool   `json:"kubectlAvailable,omitempty"`
	KubectlPath                     string `json:"kubectlPath,omitempty"`
	CurrentLoginAvailable           bool   `json:"currentLoginAvailable,omitempty"`
	CurrentContext                  string `json:"currentContext,omitempty"`
	CurrentLoginDetail              string `json:"currentLoginDetail,omitempty"`
	CurrentLoginWarning             string `json:"currentLoginWarning,omitempty"`
	DefaultKubeconfigAvailable      bool   `json:"defaultKubeconfigAvailable,omitempty"`
	DefaultKubeconfigPath           string `json:"defaultKubeconfigPath,omitempty"`
	DefaultKubeconfigCurrentContext string `json:"defaultKubeconfigCurrentContext,omitempty"`
	DefaultKubeconfigDetail         string `json:"defaultKubeconfigDetail,omitempty"`
	DefaultKubeconfigPortable       bool   `json:"defaultKubeconfigPortable,omitempty"`
	DefaultKubeconfigWarning        string `json:"defaultKubeconfigWarning,omitempty"`
}

func DetectConnectionAdvisor() ConnectionAdvisor {
	advisor := ConnectionAdvisor{
		RecommendedMethod: ConnectionMethodKubeconfigFile,
		RecommendedReason: "No local Kubernetes access was detected yet. Start by bringing a kubeconfig file or pasting kubeconfig content.",
	}
	if kubectlPath, err := exec.LookPath("kubectl"); err == nil {
		advisor.KubectlAvailable = true
		if absolute, absErr := filepath.Abs(kubectlPath); absErr == nil {
			advisor.KubectlPath = filepath.Clean(absolute)
		} else {
			advisor.KubectlPath = filepath.Clean(kubectlPath)
		}
	}

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	defaultPath := DefaultKubeconfigPath()
	if defaultPath != "" {
		if inspection, err := InspectKubeconfigFile(defaultPath); err == nil {
			advisor = applyInspectionToConnectionAdvisor(advisor, defaultPath, inspection)
			if advisor.CurrentLoginAvailable || advisor.DefaultKubeconfigAvailable {
				return advisor
			}
		}
	}

	rawCfg, err := rules.Load()
	if err == nil && rawCfg != nil && (len(rawCfg.Contexts) > 0 || strings.TrimSpace(rawCfg.CurrentContext) != "") {
		advisor.CurrentLoginAvailable = true
		advisor.CurrentContext = rawCfg.CurrentContext
		if rawCfg.CurrentContext != "" {
			advisor.CurrentLoginDetail = fmt.Sprintf("K8V found Kubernetes access with current context %q.", rawCfg.CurrentContext)
		} else {
			advisor.CurrentLoginDetail = "K8V found Kubernetes access data on this machine."
		}
		advisor.RecommendedMethod = ConnectionMethodCurrent
		advisor.RecommendedReason = "This machine already has Kubernetes access configured. Start with existing access before trying manual connection modes."
	}
	if !advisor.CurrentLoginAvailable && advisor.KubectlAvailable && !advisor.DefaultKubeconfigAvailable {
		advisor.RecommendedReason = "kubectl is installed on this machine, but no usable local kubeconfig was detected. Start by loading or pasting a kubeconfig."
	}

	return advisor
}

func applyInspectionToConnectionAdvisor(advisor ConnectionAdvisor, defaultPath string, inspection KubeconfigInspection) ConnectionAdvisor {
	advisor.DefaultKubeconfigAvailable = true
	advisor.DefaultKubeconfigPath = defaultPath
	advisor.DefaultKubeconfigCurrentContext = inspection.CurrentContext
	advisor.DefaultKubeconfigDetail = inspection.Summary
	advisor.DefaultKubeconfigPortable = len(inspection.MissingReferencedFiles) == 0
	advisor.DefaultKubeconfigWarning = defaultKubeconfigWarning(inspection)
	advisor.CurrentContext = inspection.CurrentContext

	if !advisor.DefaultKubeconfigPortable {
		advisor.CurrentLoginAvailable = false
		advisor.CurrentLoginDetail = fmt.Sprintf("Detected a kubeconfig at %s, but it is missing supporting CA or client-certificate files on this machine.", defaultPath)
		advisor.CurrentLoginWarning = advisor.DefaultKubeconfigWarning
		advisor.RecommendedMethod = ConnectionMethodKubeconfigFile
		advisor.RecommendedReason = "A kubeconfig was found on this machine, but it is not complete enough to use as-is. Bring a complete kubeconfig or another access path instead of relying on existing access."
		return advisor
	}

	advisor.CurrentLoginAvailable = true
	if inspection.CurrentContext != "" {
		advisor.CurrentLoginDetail = fmt.Sprintf("Detected local Kubernetes access with current context %q.", inspection.CurrentContext)
	} else {
		advisor.CurrentLoginDetail = fmt.Sprintf("Detected local Kubernetes access from %s.", defaultPath)
	}
	advisor.RecommendedMethod = ConnectionMethodCurrent
	advisor.RecommendedReason = "This machine already has Kubernetes access configured. Start with existing access before trying manual connection modes."

	if inspection.UsesExecAuth {
		advisor.CurrentLoginWarning = "This kubeconfig depends on an external auth helper or login plugin. Existing access can still work, but it may require the same helper or SSO session to be available in this desktop session."
		advisor.RecommendedReason = "This machine already has Kubernetes access configured. Start with existing access, but keep kubeconfig mode handy if the external auth helper cannot run in the desktop session."
	} else if len(inspection.LoopbackServers) > 0 {
		advisor.CurrentLoginWarning = "The detected kubeconfig points at a loopback API endpoint such as 127.0.0.1 or localhost. Existing access will only work when the same local tunnel, proxy, or jumpbox path is active on this machine."
		advisor.RecommendedReason = "This machine already has Kubernetes access configured, but it appears to rely on a local loopback API endpoint. Start with existing access only if the same tunnel or jumpbox path is active here."
	}
	return advisor
}

func defaultKubeconfigWarning(inspection KubeconfigInspection) string {
	switch {
	case len(inspection.MissingReferencedFiles) > 0:
		return "The detected kubeconfig is valid, but it still references CA or client-certificate files that are missing on this machine."
	case inspection.UsesExecAuth:
		return "The detected kubeconfig uses an external auth helper. Existing access can still work, but it depends on that helper being available and signed in on this machine."
	case len(inspection.LoopbackServers) > 0:
		return "The detected kubeconfig points at a loopback API endpoint such as 127.0.0.1 or localhost. That usually only works on the machine, jumpbox, or tunnel that created it."
	case inspection.UsesClientCertificate && len(inspection.ReferencedFiles) > 0:
		return "The detected kubeconfig relies on local client-certificate material. Keep the kubeconfig and its referenced certificate files together when moving between machines."
	default:
		return ""
	}
}

func DefaultKubeconfigPath() string {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	for _, candidate := range rules.GetLoadingPrecedence() {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			if absolute, absErr := filepath.Abs(candidate); absErr == nil {
				return filepath.Clean(absolute)
			}
			return filepath.Clean(candidate)
		}
	}
	return ""
}

func InspectKubeconfigFile(kubeconfigPath string) (KubeconfigInspection, error) {
	chosen := strings.TrimSpace(kubeconfigPath)
	if chosen == "" {
		return KubeconfigInspection{}, fmt.Errorf("choose a kubeconfig file or switch to existing access")
	}

	abs := chosen
	if absolute, err := filepath.Abs(chosen); err == nil {
		abs = filepath.Clean(absolute)
	}

	raw, err := os.ReadFile(abs)
	if err != nil {
		return KubeconfigInspection{}, fmt.Errorf("read kubeconfig file %q: %w", abs, err)
	}

	return inspectKubeconfigBytes(raw, "kubeconfig file", abs)
}

func InspectKubeconfigContent(content string) (KubeconfigInspection, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return KubeconfigInspection{}, fmt.Errorf("paste kubeconfig content or switch connection mode")
	}
	return inspectKubeconfigBytes([]byte(trimmed), "pasted kubeconfig", "")
}

func inspectKubeconfigBytes(raw []byte, source, path string) (KubeconfigInspection, error) {
	if strings.TrimSpace(string(raw)) == "" {
		return KubeconfigInspection{}, fmt.Errorf("the selected %s is empty", source)
	}

	cfg, err := clientcmd.Load(raw)
	if err != nil {
		return KubeconfigInspection{}, fmt.Errorf("parse %s: %w", source, err)
	}

	missing := missingKubeconfigSections(*cfg)
	if len(missing) > 0 {
		return KubeconfigInspection{}, fmt.Errorf("the selected %s is not a usable kubeconfig: missing %s", source, formatMissingSections(missing))
	}

	contexts := make([]string, 0, len(cfg.Contexts))
	for name := range cfg.Contexts {
		contexts = append(contexts, name)
	}
	sort.Strings(contexts)

	inspection := KubeconfigInspection{
		Source:         source,
		Path:           path,
		CurrentContext: strings.TrimSpace(cfg.CurrentContext),
		Contexts:       contexts,
		ClusterCount:   len(cfg.Clusters),
		UserCount:      len(cfg.AuthInfos),
	}

	for _, authInfo := range cfg.AuthInfos {
		if authInfo.Exec != nil {
			inspection.UsesExecAuth = true
		}
		if authInfo.ClientCertificate != "" || len(authInfo.ClientCertificateData) > 0 || authInfo.ClientKey != "" || len(authInfo.ClientKeyData) > 0 {
			inspection.UsesClientCertificate = true
		}
	}
	for _, cluster := range cfg.Clusters {
		if cluster.CertificateAuthority != "" {
			inspection.UsesCertificateAuthorityFile = true
		}
		if len(cluster.CertificateAuthorityData) > 0 {
			inspection.UsesCertificateAuthorityData = true
		}
	}
	inspection.Servers, inspection.LoopbackServers = inspectClusterServers(*cfg)

	baseDir := ""
	if path != "" {
		baseDir = filepath.Dir(path)
	}
	inspection.ReferencedFiles, inspection.MissingReferencedFiles = inspectReferencedFiles(*cfg, baseDir)
	inspection.Summary = kubeconfigSummary(inspection)
	inspection.NextAction = kubeconfigNextAction(inspection)
	return inspection, nil
}

type kubeconfigPathReference struct {
	label    string
	display  string
	canCheck bool
}

func inspectReferencedFiles(cfg clientcmdapi.Config, baseDir string) ([]string, []string) {
	references := make([]string, 0, len(cfg.Clusters)+len(cfg.AuthInfos)*2)
	missing := []string{}
	seenReferences := map[string]struct{}{}
	seenMissing := map[string]struct{}{}

	recordReference := func(label, rawPath string) {
		reference := resolveReference(baseDir, rawPath)
		description := fmt.Sprintf("%s: %s", label, reference.display)
		if _, ok := seenReferences[description]; !ok {
			references = append(references, description)
			seenReferences[description] = struct{}{}
		}
		if !reference.canCheck {
			return
		}
		if _, err := os.Stat(reference.display); err != nil {
			if _, ok := seenMissing[description]; ok {
				return
			}
			missing = append(missing, description)
			seenMissing[description] = struct{}{}
		}
	}

	for name, cluster := range cfg.Clusters {
		if ref := strings.TrimSpace(cluster.CertificateAuthority); ref != "" {
			recordReference(fmt.Sprintf("Cluster %s certificate authority", name), ref)
		}
	}
	for name, authInfo := range cfg.AuthInfos {
		if ref := strings.TrimSpace(authInfo.ClientCertificate); ref != "" {
			recordReference(fmt.Sprintf("User %s client certificate", name), ref)
		}
		if ref := strings.TrimSpace(authInfo.ClientKey); ref != "" {
			recordReference(fmt.Sprintf("User %s client key", name), ref)
		}
	}

	sort.Strings(references)
	sort.Strings(missing)
	return references, missing
}

func inspectClusterServers(cfg clientcmdapi.Config) ([]string, []string) {
	servers := make([]string, 0, len(cfg.Clusters))
	loopback := []string{}
	seenServers := map[string]struct{}{}
	seenLoopback := map[string]struct{}{}

	for _, cluster := range cfg.Clusters {
		server := strings.TrimSpace(cluster.Server)
		if server == "" {
			continue
		}
		if _, ok := seenServers[server]; !ok {
			servers = append(servers, server)
			seenServers[server] = struct{}{}
		}
		if isLoopbackServer(server) {
			if _, ok := seenLoopback[server]; !ok {
				loopback = append(loopback, server)
				seenLoopback[server] = struct{}{}
			}
		}
	}

	sort.Strings(servers)
	sort.Strings(loopback)
	return servers, loopback
}

func isLoopbackServer(server string) bool {
	return IsLoopbackServer(server)
}

func IsLoopbackServer(server string) bool {
	parsed, err := url.Parse(strings.TrimSpace(server))
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	if host == "" {
		host = strings.TrimSpace(server)
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func resolveReference(baseDir, rawPath string) kubeconfigPathReference {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" {
		return kubeconfigPathReference{}
	}
	if filepath.IsAbs(trimmed) {
		return kubeconfigPathReference{
			display:  filepath.Clean(trimmed),
			canCheck: true,
		}
	}
	if strings.TrimSpace(baseDir) != "" {
		return kubeconfigPathReference{
			display:  filepath.Clean(filepath.Join(baseDir, trimmed)),
			canCheck: true,
		}
	}
	return kubeconfigPathReference{
		display:  filepath.Clean(trimmed),
		canCheck: false,
	}
}

func missingKubeconfigSections(cfg clientcmdapi.Config) []string {
	missing := []string{}
	if len(cfg.Clusters) == 0 {
		missing = append(missing, "clusters")
	}
	if len(cfg.Contexts) == 0 {
		missing = append(missing, "contexts")
	}
	if len(cfg.AuthInfos) == 0 {
		missing = append(missing, "users")
	}
	return missing
}

func formatMissingSections(missing []string) string {
	if len(missing) == 1 {
		return missing[0]
	}
	if len(missing) == 2 {
		return missing[0] + " and " + missing[1]
	}
	return strings.Join(missing[:len(missing)-1], ", ") + ", and " + missing[len(missing)-1]
}

func kubeconfigSummary(inspection KubeconfigInspection) string {
	parts := []string{
		fmt.Sprintf("%d context%s", len(inspection.Contexts), pluralize(len(inspection.Contexts))),
		fmt.Sprintf("%d cluster%s", inspection.ClusterCount, pluralize(inspection.ClusterCount)),
		fmt.Sprintf("%d user entry%s", inspection.UserCount, pluralize(inspection.UserCount)),
	}
	summary := fmt.Sprintf("Loaded %s with %s.", inspection.Source, strings.Join(parts, ", "))
	if inspection.CurrentContext != "" {
		summary += fmt.Sprintf(" Current context: %s.", inspection.CurrentContext)
	}
	if inspection.UsesExecAuth {
		summary += " This kubeconfig uses an external exec auth helper."
	}
	if inspection.UsesClientCertificate {
		summary += " This kubeconfig includes client certificate settings."
	}
	if len(inspection.LoopbackServers) > 0 {
		summary += fmt.Sprintf(" It points at loopback API endpoint%s (%s).", pluralize(len(inspection.LoopbackServers)), strings.Join(inspection.LoopbackServers, ", "))
	}
	if len(inspection.ReferencedFiles) > 0 {
		summary += " This kubeconfig depends on local CA or client certificate files."
	}
	if len(inspection.MissingReferencedFiles) > 0 {
		summary += fmt.Sprintf(" %d referenced file%s are missing on this machine.", len(inspection.MissingReferencedFiles), pluralize(len(inspection.MissingReferencedFiles)))
	}
	return summary
}

func kubeconfigNextAction(inspection KubeconfigInspection) string {
	if len(inspection.MissingReferencedFiles) > 0 {
		return "This kubeconfig is valid, but some referenced CA or client certificate files are missing on this machine. Bring those files with the kubeconfig or export a self-contained kubeconfig with embedded *-data fields before testing again."
	}
	if len(inspection.LoopbackServers) > 0 {
		return "This kubeconfig is valid, but it points at 127.0.0.1, localhost, or another loopback API endpoint. Replace the server with the reachable control-plane DNS/IP for this machine, or export a kubeconfig that already uses the real cluster endpoint before testing again."
	}
	if inspection.UsesExecAuth {
		return "Use this kubeconfig on the same machine where its auth helper already works. If the exec helper or cloud login is unavailable here, switch back to existing access on a prepared jumpbox."
	}
	if len(inspection.ReferencedFiles) > 0 && inspection.Path == "" {
		return "This kubeconfig refers to local CA or client-certificate files. Paste mode does not move those files; use the original kubeconfig file on the prepared machine or export a self-contained kubeconfig with embedded *-data fields."
	}
	if inspection.CurrentContext == "" {
		return "Pick a context explicitly before testing the connection or running the scan."
	}
	return "Test the connection next. If it succeeds, continue to scope and outputs before running full preflight."
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
