import type { ConnectionMethod, ScanRequest } from "./types";

export function sanitizeScanForm(scanForm: ScanRequest): ScanRequest {
  const connectionMethod = normalizeConnectionMethod(scanForm);
  const sanitized: ScanRequest = {
    ...scanForm,
    connectionMethod,
    kubeconfigPath: trim(scanForm.kubeconfigPath),
    kubeconfigContent: trimMultiline(scanForm.kubeconfigContent),
    contextName: trim(scanForm.contextName),
    apiServerEndpoint: trim(scanForm.apiServerEndpoint),
    bearerToken: trim(scanForm.bearerToken),
    caCertPath: trim(scanForm.caCertPath),
    caCertContent: trimMultiline(scanForm.caCertContent),
    outputDir: trim(scanForm.outputDir),
    customerId: trim(scanForm.customerId),
    site: trim(scanForm.site),
    clusterName: trim(scanForm.clusterName),
    environment: trim(scanForm.environment),
    target: trim(scanForm.target),
    compareTo: trim(scanForm.compareTo),
    profileName: trim(scanForm.profileName),
    namespaces: sanitizeList(scanForm.namespaces),
  };

  switch (connectionMethod) {
    case "kubeconfig_file":
      sanitized.kubeconfigContent = "";
      sanitized.apiServerEndpoint = "";
      sanitized.bearerToken = "";
      sanitized.caCertPath = "";
      sanitized.caCertContent = "";
      break;
    case "kubeconfig_inline":
      sanitized.kubeconfigPath = "";
      sanitized.apiServerEndpoint = "";
      sanitized.bearerToken = "";
      sanitized.caCertPath = "";
      sanitized.caCertContent = "";
      break;
    case "api_endpoint":
      sanitized.kubeconfigPath = "";
      sanitized.kubeconfigContent = "";
      sanitized.contextName = "";
      break;
    default:
      sanitized.kubeconfigPath = "";
      sanitized.kubeconfigContent = "";
      sanitized.apiServerEndpoint = "";
      sanitized.bearerToken = "";
      sanitized.caCertPath = "";
      sanitized.caCertContent = "";
      break;
  }

  return sanitized;
}

export function prepareScanRequest(scanForm: ScanRequest): ScanRequest {
  return sanitizeScanForm(scanForm);
}

export function validateScanForm(scanForm: ScanRequest) {
  const sanitized = sanitizeScanForm(scanForm);
  const errors: string[] = [];

  if (!sanitized.outputDir) {
    errors.push("Choose an output directory for the scan bundle and reports.");
  }
  if ((sanitized.timeoutSeconds || 0) < 10) {
    errors.push("Set the timeout to at least 10 seconds.");
  }

  switch (sanitized.connectionMethod) {
    case "kubeconfig_file":
      if (!sanitized.kubeconfigPath) {
        errors.push("Choose a kubeconfig file or switch to Current login.");
      }
      break;
    case "kubeconfig_inline":
      if (!sanitized.kubeconfigContent) {
        errors.push("Paste a kubeconfig or switch to another connection method.");
      }
      break;
    case "api_endpoint":
      if (!sanitized.apiServerEndpoint) {
        errors.push("Enter the Kubernetes API server host, IP, or URL.");
      }
      if (!sanitized.bearerToken) {
        errors.push("Paste a bearer token for direct API endpoint scans.");
      }
      break;
    default:
      break;
  }

  return errors;
}

export function validateContextDiscovery(scanForm: ScanRequest) {
  const sanitized = sanitizeScanForm(scanForm);
  const errors: string[] = [];

  switch (sanitized.connectionMethod) {
    case "kubeconfig_file":
      if (!sanitized.kubeconfigPath) {
        errors.push("Choose a kubeconfig file before loading contexts.");
      }
      break;
    case "kubeconfig_inline":
      if (!sanitized.kubeconfigContent) {
        errors.push("Paste a kubeconfig before loading contexts.");
      }
      break;
    default:
      break;
  }

  return errors;
}

export function haveContextInputsChanged(previous: ScanRequest, next: ScanRequest) {
  const before = sanitizeScanForm(previous);
  const after = sanitizeScanForm(next);
  return (
    before.connectionMethod !== after.connectionMethod ||
    before.kubeconfigPath !== after.kubeconfigPath ||
    before.kubeconfigContent !== after.kubeconfigContent ||
    before.apiServerEndpoint !== after.apiServerEndpoint
  );
}

function normalizeConnectionMethod(scanForm: ScanRequest): ConnectionMethod {
  switch (scanForm.connectionMethod) {
    case "kubeconfig_file":
    case "kubeconfig_inline":
    case "api_endpoint":
      return scanForm.connectionMethod;
    default:
      return "current";
  }
}

function sanitizeList(values?: string[]) {
  if (!values?.length) {
    return [];
  }
  return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean)));
}

function trim(value?: string) {
  return value?.trim() || "";
}

function trimMultiline(value?: string) {
  return value?.trim() || "";
}
