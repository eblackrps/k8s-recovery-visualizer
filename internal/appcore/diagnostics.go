package appcore

import "strings"

func diagnoseFailure(req ScanRequest, message string) FailureDiagnosis {
	detail := strings.TrimSpace(message)
	lower := strings.ToLower(detail)

	switch {
	case isOutputPathFailure(lower):
		return FailureDiagnosis{
			Code:       "output_path",
			Label:      "Output path",
			Summary:    "K8V could not prepare the output directory.",
			Detail:     detail,
			NextAction: "Choose a writable output directory on this machine, then try the scan again.",
		}
	case isArtifactWriteFailure(lower):
		return FailureDiagnosis{
			Code:       "artifact_write",
			Label:      "Report output",
			Summary:    "K8V connected successfully, but writing the bundle or report artifacts failed.",
			Detail:     detail,
			NextAction: "Check that the output directory is writable and that no report files are locked by another process, then rerun the scan.",
		}
	case isMissingKubeconfigDependency(lower):
		return FailureDiagnosis{
			Code:       "kubeconfig_dependencies",
			Label:      "Kubeconfig dependencies",
			Summary:    "This kubeconfig depends on local CA or client-certificate files that are not available here.",
			Detail:     detail,
			NextAction: "Bring the referenced CA or client-certificate files with the kubeconfig, or export a self-contained kubeconfig with embedded *-data fields before testing again.",
		}
	case isExecAuthFailure(lower):
		return FailureDiagnosis{
			Code:       "auth_helper",
			Label:      "External auth helper",
			Summary:    "This connection depends on an external auth helper that is not working on this machine.",
			Detail:     detail,
			NextAction: "Use this kubeconfig on the same machine where its exec or cloud auth helper already works, or switch back to existing access on a prepared jumpbox.",
		}
	case isTLSFailure(lower):
		return FailureDiagnosis{
			Code:       "tls_trust",
			Label:      "TLS trust",
			Summary:    "The cluster is reachable, but this machine does not trust the API server certificate.",
			Detail:     detail,
			NextAction: "Add the issuing CA, use a kubeconfig with embedded CA data, or use skip-TLS only as a temporary workaround in a trusted environment.",
		}
	case isReachabilityFailure(lower):
		if req.ConnectionMethod == ConnectionMethodKubeconfigFile || req.ConnectionMethod == ConnectionMethodKubeconfigInline {
			return FailureDiagnosis{
				Code:       "endpoint_unreachable",
				Label:      "Cluster reachability",
				Summary:    "The kubeconfig was accepted, but the cluster API it points to is not reachable from this machine.",
				Detail:     detail,
				NextAction: "The file is valid. Check VPN, private DNS, firewall path, proxy, or whether the kubeconfig points at an internal-only control-plane endpoint that only works from a work jumpbox or cluster network.",
			}
		}
		return FailureDiagnosis{
			Code:       "endpoint_unreachable",
			Label:      "API reachability",
			Summary:    "The cluster API is not reachable from this machine.",
			Detail:     detail,
			NextAction: "Double-check the API server address, DNS, VPN, firewall path, or whether the kubeconfig points at an internal-only control-plane endpoint.",
		}
	case isRBACFailure(lower):
		return FailureDiagnosis{
			Code:       "rbac_denied",
			Label:      "Access scope",
			Summary:    "Credentials were accepted, but the cluster denied the requested API access.",
			Detail:     detail,
			NextAction: "Run preflight or kubectl auth can-i checks to confirm the minimum required read access before starting the full scan.",
		}
	case isAuthFailure(lower):
		return FailureDiagnosis{
			Code:       "auth_rejected",
			Label:      "Credentials",
			Summary:    "The API server rejected the current credentials.",
			Detail:     detail,
			NextAction: "Use credentials that already work with kubectl on this machine, or switch to kubeconfig mode if the cluster relies on exec helpers, cloud login, or client certificates.",
		}
	case isKubeconfigValidationFailure(req, lower):
		return FailureDiagnosis{
			Code:       "kubeconfig_validation",
			Label:      "Kubeconfig validation",
			Summary:    "The selected kubeconfig could not be used as provided.",
			Detail:     detail,
			NextAction: kubeconfigFailureAction(req),
		}
	default:
		return FailureDiagnosis{
			Code:       "unknown",
			Label:      "Unknown failure",
			Summary:    "K8V could not complete this step.",
			Detail:     detail,
			NextAction: connectionHint(req),
		}
	}
}

func isOutputPathFailure(lower string) bool {
	return containsAny(lower,
		"create output directory",
		"mkdir ",
		"access is denied",
		"permission denied",
		"read-only file system",
		"the system cannot find the path specified",
	)
}

func isArtifactWriteFailure(lower string) bool {
	return containsAny(lower,
		"write json",
		"write html report",
		"write summary",
		"write runbook",
		"write csv",
		"write redacted",
		"write enriched artifacts",
		"enrich outputs",
	)
}

func isMissingKubeconfigDependency(lower string) bool {
	return containsAny(lower,
		"missing on this machine",
		"referenced ca or client certificate files",
		"certificate authority",
		"client certificate",
		"client key",
	) && containsAny(lower,
		"no such file",
		"cannot find the file",
		"missing",
		"not found",
	)
}

func isExecAuthFailure(lower string) bool {
	return containsAny(lower,
		"exec plugin",
		"auth provider",
		"external auth helper",
		"executable file not found",
		"exec: executable",
	)
}

func isTLSFailure(lower string) bool {
	return containsAny(lower,
		"x509",
		"certificate signed by unknown authority",
		"tls",
		"certificate is valid for",
		"unknown authority",
		"certificate verify failed",
	)
}

func isReachabilityFailure(lower string) bool {
	return containsAny(lower,
		"dial tcp",
		"no such host",
		"connection refused",
		"i/o timeout",
		"context deadline exceeded",
		"timeout awaiting headers",
		"server misbehaving",
		"network is unreachable",
	)
}

func isRBACFailure(lower string) bool {
	return containsAny(lower,
		"forbidden",
		"selfsubjectaccessreview",
		"cannot list resource",
		"cannot get resource",
		"cannot watch resource",
	)
}

func isAuthFailure(lower string) bool {
	return containsAny(lower,
		"unauthorized",
		"authentication required",
		"token has expired",
		"invalid bearer token",
		"oauth",
	)
}

func isKubeconfigValidationFailure(req ScanRequest, lower string) bool {
	if req.ConnectionMethod != ConnectionMethodKubeconfigFile && req.ConnectionMethod != ConnectionMethodKubeconfigInline {
		return false
	}
	return containsAny(lower,
		"load kubeconfig",
		"read kubeconfig",
		"invalid yaml",
		"not a kubeconfig",
		"current-context",
		"cluster entry",
		"user entry",
		"context entry",
		"choose a kubeconfig file",
		"paste a kubeconfig",
	)
}

func kubeconfigFailureAction(req ScanRequest) string {
	if req.ConnectionMethod == ConnectionMethodKubeconfigInline {
		return "Paste a full kubeconfig document, or switch to a file-based or existing-access connection path."
	}
	return "Choose another kubeconfig file, drag or paste the kubeconfig content directly, or switch back to existing access on a machine where Kubernetes access already works."
}

func containsAny(lower string, parts ...string) bool {
	for _, part := range parts {
		if strings.Contains(lower, part) {
			return true
		}
	}
	return false
}
