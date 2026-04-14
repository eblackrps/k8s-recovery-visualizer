package appcore

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/client-go/kubernetes"

	"k8s-recovery-visualizer/internal/kube"
)

func (s *Service) ConnectionAdvisor() ConnectionAdvisor {
	advisor := kube.DetectConnectionAdvisor()
	return ConnectionAdvisor{
		RecommendedMethod:               advisor.RecommendedMethod,
		RecommendedReason:               advisor.RecommendedReason,
		KubectlAvailable:                advisor.KubectlAvailable,
		KubectlPath:                     advisor.KubectlPath,
		CurrentLoginAvailable:           advisor.CurrentLoginAvailable,
		CurrentContext:                  advisor.CurrentContext,
		CurrentLoginDetail:              advisor.CurrentLoginDetail,
		CurrentLoginWarning:             advisor.CurrentLoginWarning,
		DefaultKubeconfigAvailable:      advisor.DefaultKubeconfigAvailable,
		DefaultKubeconfigPath:           advisor.DefaultKubeconfigPath,
		DefaultKubeconfigCurrentContext: advisor.DefaultKubeconfigCurrentContext,
		DefaultKubeconfigDetail:         advisor.DefaultKubeconfigDetail,
		DefaultKubeconfigPortable:       advisor.DefaultKubeconfigPortable,
		DefaultKubeconfigWarning:        advisor.DefaultKubeconfigWarning,
	}
}

func (s *Service) InspectKubeconfig(req ScanRequest) (KubeconfigInspection, error) {
	switch {
	case strings.TrimSpace(req.KubeconfigContent) != "":
		inspection, err := kube.InspectKubeconfigContent(req.KubeconfigContent)
		if err != nil {
			return KubeconfigInspection{}, err
		}
		return mapKubeconfigInspection(inspection), nil
	case strings.TrimSpace(req.KubeconfigPath) != "":
		inspection, err := kube.InspectKubeconfigFile(req.KubeconfigPath)
		if err != nil {
			return KubeconfigInspection{}, err
		}
		return mapKubeconfigInspection(inspection), nil
	default:
		return KubeconfigInspection{}, fmt.Errorf("choose a kubeconfig file or paste kubeconfig content first")
	}
}

func (s *Service) TestConnection(ctx context.Context, req ScanRequest) (ConnectionTestReport, error) {
	req = sanitizeScanRequest(req)
	if req.DryRun {
		return ConnectionTestReport{
			CanConnect: true,
			Summary:    "Dry-run mode skips live cluster contact.",
			NextAction: "Continue to scope and outputs, then run the full scan when you are ready for live collection.",
			Checks: []ConnectionTestCheck{
				{ID: "mode", Title: "Connection mode", Status: "pass", Detail: "Dry-run mode does not require a cluster connection."},
			},
		}, nil
	}

	switch req.ConnectionMethod {
	case ConnectionMethodKubeconfigFile:
		if _, err := s.InspectKubeconfig(req); err != nil {
			return reportKubeconfigFailure(req, err), nil
		}
	case ConnectionMethodKubeconfigInline:
		if _, err := s.InspectKubeconfig(req); err != nil {
			return reportKubeconfigFailure(req, err), nil
		}
	}

	resolved, err := kube.ResolveConfig(kubeOptionsFromRequest(req))
	if err != nil {
		return reportConnectionFailure(req, "", "", "", err), nil
	}

	report := ConnectionTestReport{
		CanConnect:  false,
		Source:      resolved.Source,
		Server:      resolved.Config.Host,
		ContextName: resolved.ContextName,
		Checks: []ConnectionTestCheck{
			{
				ID:     "config",
				Title:  "Connection settings loaded",
				Status: "pass",
				Detail: connectionStatusDetail(resolved.Source),
			},
		},
	}

	clientset, err := kubernetes.NewForConfig(resolved.Config)
	if err != nil {
		return reportConnectionFailure(req, resolved.Source, resolved.Config.Host, resolved.ContextName, err), nil
	}

	if _, err := clientset.Discovery().ServerVersion(); err != nil {
		return reportConnectionFailure(req, resolved.Source, resolved.Config.Host, resolved.ContextName, err), nil
	}

	report.Checks = append(report.Checks,
		ConnectionTestCheck{
			ID:     "transport",
			Title:  "API server reachability",
			Status: "pass",
			Detail: fmt.Sprintf("Reached %s.", resolved.Config.Host),
		},
		ConnectionTestCheck{
			ID:     "handshake",
			Title:  "Basic API handshake",
			Status: "pass",
			Detail: "The cluster responded to a basic discovery request. Continue to preflight for RBAC and collection readiness.",
		},
	)

	if req.ConnectionMethod != ConnectionMethodAPIEndpoint {
		report.Checks = append(report.Checks, ConnectionTestCheck{
			ID:     "context",
			Title:  "Context resolution",
			Status: "pass",
			Detail: resolvedContextDetail(resolved),
		})
	}

	report.CanConnect = true
	report.Summary = "Connection test succeeded."
	report.NextAction = "Continue to scope and outputs, then run full preflight to check RBAC and collection readiness before the scan."
	return report, nil
}

func mapKubeconfigInspection(inspection kube.KubeconfigInspection) KubeconfigInspection {
	return KubeconfigInspection{
		Source:                       inspection.Source,
		Path:                         inspection.Path,
		CurrentContext:               inspection.CurrentContext,
		Contexts:                     inspection.Contexts,
		ClusterCount:                 inspection.ClusterCount,
		UserCount:                    inspection.UserCount,
		UsesExecAuth:                 inspection.UsesExecAuth,
		UsesClientCertificate:        inspection.UsesClientCertificate,
		UsesCertificateAuthorityFile: inspection.UsesCertificateAuthorityFile,
		UsesCertificateAuthorityData: inspection.UsesCertificateAuthorityData,
		Servers:                      inspection.Servers,
		LoopbackServers:              inspection.LoopbackServers,
		ReferencedFiles:              inspection.ReferencedFiles,
		MissingReferencedFiles:       inspection.MissingReferencedFiles,
		Summary:                      inspection.Summary,
		NextAction:                   inspection.NextAction,
	}
}

func reportKubeconfigFailure(req ScanRequest, err error) ConnectionTestReport {
	field := "kubeconfigPath"
	title := "Kubeconfig file validation"
	diagnosis := diagnoseFailure(req, err.Error())
	if req.ConnectionMethod == ConnectionMethodKubeconfigInline {
		field = "kubeconfigContent"
		title = "Pasted kubeconfig validation"
	}
	return ConnectionTestReport{
		CanConnect:  false,
		Summary:     diagnosis.Summary,
		NextAction:  diagnosis.NextAction,
		Diagnosis:   &diagnosis,
		FieldErrors: map[string]string{field: strings.TrimSpace(err.Error())},
		Checks: []ConnectionTestCheck{
			{
				ID:     "config",
				Title:  title,
				Status: "fail",
				Detail: strings.TrimSpace(err.Error()),
				Hint:   diagnosis.NextAction,
			},
		},
	}
}

func reportConnectionFailure(req ScanRequest, source, server, contextName string, err error) ConnectionTestReport {
	message := strings.TrimSpace(err.Error())
	diagnosis := diagnoseFailure(req, message)
	report := ConnectionTestReport{
		CanConnect:    false,
		Source:        source,
		Server:        server,
		ContextName:   contextName,
		Summary:       diagnosis.Summary,
		NextAction:    diagnosis.NextAction,
		Diagnosis:     &diagnosis,
		FieldErrors:   map[string]string{},
		FieldWarnings: map[string]string{},
	}

	if source != "" {
		report.Checks = append(report.Checks, ConnectionTestCheck{
			ID:     "config",
			Title:  "Connection settings loaded",
			Status: "pass",
			Detail: connectionStatusDetail(source),
		})
	}

	switch diagnosis.Code {
	case "tls_trust":
		report.FieldErrors["caTrust"] = "The API server certificate could not be verified. Add the issuing CA or use skip-TLS only temporarily."
		report.Checks = append(report.Checks, ConnectionTestCheck{
			ID:     "transport",
			Title:  "TLS and API handshake",
			Status: "fail",
			Detail: message,
			Hint:   diagnosis.NextAction,
		})
	case "auth_rejected":
		if req.ConnectionMethod == ConnectionMethodAPIEndpoint {
			report.FieldErrors["bearerToken"] = "The API server rejected the current token or credential set."
		}
		report.Checks = append(report.Checks, ConnectionTestCheck{
			ID:     "auth",
			Title:  "Credential acceptance",
			Status: "fail",
			Detail: message,
			Hint:   diagnosis.NextAction,
		})
	case "auth_helper":
		if req.ConnectionMethod == ConnectionMethodKubeconfigFile {
			report.FieldErrors["kubeconfigPath"] = "This kubeconfig depends on an external auth helper that is not working on this machine."
		}
		if req.ConnectionMethod == ConnectionMethodKubeconfigInline {
			report.FieldErrors["kubeconfigContent"] = "This kubeconfig depends on an external auth helper that is not working on this machine."
		}
		report.Checks = append(report.Checks, ConnectionTestCheck{
			ID:     "auth-helper",
			Title:  "External auth helper",
			Status: "fail",
			Detail: message,
			Hint:   diagnosis.NextAction,
		})
	case "endpoint_unreachable":
		if req.ConnectionMethod == ConnectionMethodAPIEndpoint {
			report.FieldErrors["apiServerEndpoint"] = "K8V could not reach this API server address from the current machine."
		}
		if req.ConnectionMethod == ConnectionMethodKubeconfigFile {
			report.FieldWarnings["kubeconfigPath"] = "The kubeconfig parsed correctly, but the cluster API inside it is not reachable from this machine."
		}
		if req.ConnectionMethod == ConnectionMethodKubeconfigInline {
			report.FieldWarnings["kubeconfigContent"] = "The kubeconfig parsed correctly, but the cluster API inside it is not reachable from this machine."
		}
		if isLoopbackKubeconfigEndpoint(req, server) {
			diagnosis.Label = "Loopback kubeconfig endpoint"
			diagnosis.Summary = "The kubeconfig was accepted, but it points at 127.0.0.1 or localhost, which is only reachable from the machine or tunnel that created it."
			diagnosis.NextAction = "Replace the kubeconfig server with the reachable control-plane DNS/IP for this machine, or export a kubeconfig that already uses the real cluster endpoint before testing again."
			report.Summary = diagnosis.Summary
			report.NextAction = diagnosis.NextAction
			if req.ConnectionMethod == ConnectionMethodKubeconfigFile {
				report.FieldWarnings["kubeconfigPath"] = "This kubeconfig points at 127.0.0.1 or localhost. Replace it with the real control-plane address for this machine."
			}
			if req.ConnectionMethod == ConnectionMethodKubeconfigInline {
				report.FieldWarnings["kubeconfigContent"] = "This kubeconfig points at 127.0.0.1 or localhost. Replace it with the real control-plane address for this machine."
			}
		}
		report.Diagnosis = &diagnosis
		report.Checks = append(report.Checks, ConnectionTestCheck{
			ID:     "transport",
			Title:  "Cluster API reachability",
			Status: "fail",
			Detail: message,
			Hint:   diagnosis.NextAction,
		})
	case "rbac_denied":
		if req.ConnectionMethod == ConnectionMethodAPIEndpoint {
			report.FieldWarnings["bearerToken"] = "The token connected to the cluster, but the API denied the requested access."
		}
		report.Checks = append(report.Checks, ConnectionTestCheck{
			ID:     "access",
			Title:  "API access scope",
			Status: "fail",
			Detail: message,
			Hint:   diagnosis.NextAction,
		})
	default:
		report.Checks = append(report.Checks, ConnectionTestCheck{
			ID:     "connection",
			Title:  "Connection test",
			Status: "fail",
			Detail: message,
			Hint:   diagnosis.NextAction,
		})
	}

	if len(report.FieldErrors) == 0 {
		report.FieldErrors = nil
	}
	if len(report.FieldWarnings) == 0 {
		report.FieldWarnings = nil
	}
	return report
}

func isLoopbackKubeconfigEndpoint(req ScanRequest, server string) bool {
	if req.ConnectionMethod != ConnectionMethodKubeconfigFile && req.ConnectionMethod != ConnectionMethodKubeconfigInline {
		return false
	}
	return kube.IsLoopbackServer(server)
}

func resolvedContextDetail(resolved *kube.ResolvedConnection) string {
	if resolved == nil {
		return "No context was resolved."
	}
	if strings.TrimSpace(resolved.ContextName) != "" {
		return fmt.Sprintf("Using context %s.", resolved.ContextName)
	}
	return "No explicit context override was needed."
}
