import type { FailureDiagnosis, ScanRequest } from "./types";

export function diagnoseFailure(request: Pick<ScanRequest, "connectionMethod">, message: string): FailureDiagnosis {
  const detail = message.trim();
  const lower = detail.toLowerCase();

  if (isOutputPathFailure(lower)) {
    return {
      code: "output_path",
      label: "Output path",
      summary: "K8V could not prepare the output directory.",
      detail,
      nextAction: "Choose a writable output directory on this machine, then try the scan again.",
    };
  }

  if (isArtifactWriteFailure(lower)) {
    return {
      code: "artifact_write",
      label: "Report output",
      summary: "K8V connected successfully, but writing the bundle or report artifacts failed.",
      detail,
      nextAction: "Check that the output directory is writable and that no report files are locked by another process, then rerun the scan.",
    };
  }

  if (isMissingKubeconfigDependency(lower)) {
    return {
      code: "kubeconfig_dependencies",
      label: "Kubeconfig dependencies",
      summary: "This kubeconfig depends on local CA or client-certificate files that are not available here.",
      detail,
      nextAction:
        "Bring the referenced CA or client-certificate files with the kubeconfig, or export a self-contained kubeconfig with embedded *-data fields before testing again.",
    };
  }

  if (isExecAuthFailure(lower)) {
    return {
      code: "auth_helper",
      label: "External auth helper",
      summary: "This connection depends on an external auth helper that is not working on this machine.",
      detail,
      nextAction:
        "Use this kubeconfig on the same machine where its exec or cloud auth helper already works, or switch back to existing access on a prepared jumpbox.",
    };
  }

  if (isTLSFailure(lower)) {
    return {
      code: "tls_trust",
      label: "TLS trust",
      summary: "The cluster is reachable, but this machine does not trust the API server certificate.",
      detail,
      nextAction:
        "Add the issuing CA, use a kubeconfig with embedded CA data, or use skip-TLS only as a temporary workaround in a trusted environment.",
    };
  }

  if (isReachabilityFailure(lower)) {
    if (request.connectionMethod === "kubeconfig_file" || request.connectionMethod === "kubeconfig_inline") {
      return {
        code: "endpoint_unreachable",
        label: "Cluster reachability",
        summary: "The kubeconfig was accepted, but the cluster API it points to is not reachable from this machine.",
        detail,
        nextAction:
          "The file is valid. Check VPN, private DNS, firewall path, proxy, or whether the kubeconfig points at an internal-only control-plane endpoint that only works from a work jumpbox or cluster network.",
      };
    }
    return {
      code: "endpoint_unreachable",
      label: "API reachability",
      summary: "The cluster API is not reachable from this machine.",
      detail,
      nextAction:
        "Double-check the API server address, DNS, VPN, firewall path, or whether the kubeconfig points at an internal-only control-plane endpoint.",
    };
  }

  if (isRBACFailure(lower)) {
    return {
      code: "rbac_denied",
      label: "Access scope",
      summary: "Credentials were accepted, but the cluster denied the requested API access.",
      detail,
      nextAction:
        "Run preflight or kubectl auth can-i checks to confirm the minimum required read access before starting the full scan.",
    };
  }

  if (isAuthFailure(lower)) {
    return {
      code: "auth_rejected",
      label: "Credentials",
      summary: "The API server rejected the current credentials.",
      detail,
      nextAction:
        "Use credentials that already work with kubectl on this machine, or switch to kubeconfig mode if the cluster relies on exec helpers, cloud login, or client certificates.",
    };
  }

  if (isKubeconfigValidationFailure(request, lower)) {
    return {
      code: "kubeconfig_validation",
      label: "Kubeconfig validation",
      summary: "The selected kubeconfig could not be used as provided.",
      detail,
      nextAction:
        request.connectionMethod === "kubeconfig_inline"
          ? "Paste a full kubeconfig document, or switch to a file-based or existing-access connection path."
          : "Choose another kubeconfig file, drag or paste the kubeconfig content directly, or switch back to existing access on a machine where Kubernetes access already works.",
    };
  }

  return {
    code: "unknown",
    label: "Unknown failure",
    summary: "K8V could not complete this step.",
    detail,
    nextAction: "Review the raw detail below, adjust the connection or output settings, and try again.",
  };
}

function isOutputPathFailure(lower: string) {
  return containsAny(lower, "create output directory", "mkdir ", "access is denied", "read-only file system", "the system cannot find the path specified");
}

function isArtifactWriteFailure(lower: string) {
  return containsAny(lower, "write json", "write html report", "write summary", "write runbook", "write csv", "write redacted", "write enriched artifacts", "enrich outputs");
}

function isMissingKubeconfigDependency(lower: string) {
  return (
    containsAny(lower, "missing on this machine", "referenced ca or client certificate files", "certificate authority", "client certificate", "client key") &&
    containsAny(lower, "no such file", "cannot find the file", "missing", "not found")
  );
}

function isExecAuthFailure(lower: string) {
  return containsAny(lower, "exec plugin", "auth provider", "external auth helper", "executable file not found", "exec: executable");
}

function isTLSFailure(lower: string) {
  return containsAny(lower, "x509", "certificate signed by unknown authority", "tls", "certificate is valid for", "unknown authority", "certificate verify failed");
}

function isReachabilityFailure(lower: string) {
  return containsAny(lower, "dial tcp", "no such host", "connection refused", "i/o timeout", "context deadline exceeded", "timeout awaiting headers", "server misbehaving", "network is unreachable");
}

function isRBACFailure(lower: string) {
  return containsAny(lower, "forbidden", "selfsubjectaccessreview", "cannot list resource", "cannot get resource", "cannot watch resource");
}

function isAuthFailure(lower: string) {
  return containsAny(lower, "unauthorized", "authentication required", "token has expired", "invalid bearer token", "oauth");
}

function isKubeconfigValidationFailure(request: Pick<ScanRequest, "connectionMethod">, lower: string) {
  if (request.connectionMethod !== "kubeconfig_file" && request.connectionMethod !== "kubeconfig_inline") {
    return false;
  }
  return containsAny(lower, "load kube config", "load kubeconfig", "read kubeconfig", "invalid yaml", "not a kubeconfig", "current-context", "cluster entry", "user entry", "context entry", "choose a kubeconfig file", "paste a kubeconfig");
}

function containsAny(lower: string, ...parts: string[]) {
  return parts.some((part) => lower.includes(part));
}
