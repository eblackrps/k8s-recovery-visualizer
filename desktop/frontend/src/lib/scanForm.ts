import type { ConnectionMethod, ScanRequest } from "./types";

export type ScanFieldName =
  | "kubeconfigPath"
  | "kubeconfigContent"
  | "contextName"
  | "apiServerEndpoint"
  | "bearerToken"
  | "caTrust"
  | "insecureAcknowledgement"
  | "outputDir"
  | "timeoutSeconds";

export type ScanReviewTone = "neutral" | "success" | "critical" | "high" | "medium";

export type ScanValidation = {
  errors: string[];
  fieldErrors: Partial<Record<ScanFieldName, string>>;
  fieldWarnings: Partial<Record<ScanFieldName, string>>;
  riskFlags: string[];
  firstInvalidField?: ScanFieldName;
};

type ScanValidationOptions = {
  insecureAcknowledged?: boolean;
};

const validationFieldOrder: ScanFieldName[] = [
  "apiServerEndpoint",
  "bearerToken",
  "caTrust",
  "insecureAcknowledgement",
  "contextName",
  "kubeconfigPath",
  "kubeconfigContent",
  "outputDir",
  "timeoutSeconds",
];

export function sanitizeScanForm(scanForm: ScanRequest): ScanRequest {
  const connectionMethod = normalizeConnectionMethod(scanForm);
  const sanitized: ScanRequest = {
    ...scanForm,
    connectionMethod,
    kubeconfigPath: trim(scanForm.kubeconfigPath),
    kubeconfigContent: trimMultiline(scanForm.kubeconfigContent),
    contextName: trim(scanForm.contextName),
    apiServerEndpoint: trim(scanForm.apiServerEndpoint),
    bearerToken: normalizeBearerToken(scanForm.bearerToken),
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

export function inspectConnectionSetup(scanForm: ScanRequest, options: ScanValidationOptions = {}): ScanValidation {
  const sanitized = sanitizeScanForm(scanForm);
  const errors: string[] = [];
  const fieldErrors: Partial<Record<ScanFieldName, string>> = {};
  const fieldWarnings: Partial<Record<ScanFieldName, string>> = {};
  const riskFlags: string[] = [];

  const addError = (field: ScanFieldName, message: string) => {
    if (!fieldErrors[field]) {
      fieldErrors[field] = message;
    }
    errors.push(message);
  };

  const addWarning = (field: ScanFieldName, message: string) => {
    if (!fieldWarnings[field]) {
      fieldWarnings[field] = message;
    }
  };

  switch (sanitized.connectionMethod) {
    case "kubeconfig_file":
      if (!sanitized.kubeconfigPath) {
        addError("kubeconfigPath", "Choose a kubeconfig file or switch to Use existing access.");
      }
      break;
    case "kubeconfig_inline":
      if (!sanitized.kubeconfigContent) {
        addError("kubeconfigContent", "Paste a kubeconfig or switch to another connection method.");
      }
      break;
    case "api_endpoint":
      if (!sanitized.apiServerEndpoint) {
        addError("apiServerEndpoint", "Enter the Kubernetes API server host, IP, or URL.");
      }
      if (!sanitized.bearerToken) {
        addError("bearerToken", "Paste a bearer token for direct API endpoint scans.");
      }
      if (isExplicitHttpEndpoint(sanitized.apiServerEndpoint)) {
        addWarning("apiServerEndpoint", "HTTP endpoints are unusual for Kubernetes APIs. Prefer HTTPS or omit the scheme so the app can assume HTTPS.");
        riskFlags.push("Endpoint is explicitly using HTTP instead of HTTPS.");
      }
      if (sanitized.insecure) {
        riskFlags.push("TLS verification is disabled for this connection.");
        if (!options.insecureAcknowledged) {
          addError("insecureAcknowledgement", "Acknowledge the skip-TLS warning before preflight or launch.");
        }
      }
      if (isLikelyPrivateEndpoint(sanitized.apiServerEndpoint) && !sanitized.insecure && !sanitized.caCertPath && !sanitized.caCertContent) {
        addWarning("caTrust", "If this cluster uses an internal or self-signed certificate, add the issuing CA before preflight.");
        riskFlags.push("No CA material is configured for a likely private or internal API endpoint.");
      }
      break;
    default:
      break;
  }

  return {
    errors,
    fieldErrors,
    fieldWarnings,
    riskFlags,
    firstInvalidField: validationFieldOrder.find((field) => Boolean(fieldErrors[field])),
  };
}

export function inspectScanForm(scanForm: ScanRequest, options: ScanValidationOptions = {}): ScanValidation {
  const sanitized = sanitizeScanForm(scanForm);
  const validation = inspectConnectionSetup(sanitized, options);
  const errors = [...validation.errors];
  const fieldErrors = { ...validation.fieldErrors };
  const fieldWarnings = { ...validation.fieldWarnings };
  const riskFlags = [...validation.riskFlags];

  const addError = (field: ScanFieldName, message: string) => {
    if (!fieldErrors[field]) {
      fieldErrors[field] = message;
    }
    errors.push(message);
  };

  if (!sanitized.outputDir) {
    addError("outputDir", "Choose an output directory for the scan bundle and reports.");
  }
  if ((sanitized.timeoutSeconds || 0) < 10) {
    addError("timeoutSeconds", "Set the timeout to at least 10 seconds.");
  }
  if (sanitized.connectionMethod === "api_endpoint" && !sanitized.namespaces?.length) {
    riskFlags.push("The scan will use all readable namespaces.");
  }

  return {
    errors,
    fieldErrors,
    fieldWarnings,
    riskFlags,
    firstInvalidField: validationFieldOrder.find((field) => Boolean(fieldErrors[field])),
  };
}

export function validateScanForm(scanForm: ScanRequest, options: ScanValidationOptions = {}) {
  return inspectScanForm(scanForm, options).errors;
}

export function validateContextDiscovery(scanForm: ScanRequest) {
  const sanitized = sanitizeScanForm(scanForm);
  const validation = inspectConnectionSetup(sanitized);
  if (validation.errors.length > 0) {
    return validation.errors;
  }
  return [];
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

export function haveConnectionInputsChanged(previous: ScanRequest, next: ScanRequest) {
  const before = sanitizeScanForm(previous);
  const after = sanitizeScanForm(next);
  return (
    before.connectionMethod !== after.connectionMethod ||
    before.kubeconfigPath !== after.kubeconfigPath ||
    before.kubeconfigContent !== after.kubeconfigContent ||
    before.contextName !== after.contextName ||
    before.apiServerEndpoint !== after.apiServerEndpoint ||
    before.bearerToken !== after.bearerToken ||
    before.caCertPath !== after.caCertPath ||
    before.caCertContent !== after.caCertContent ||
    before.insecure !== after.insecure
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

export function normalizeBearerToken(value?: string) {
  return value
    ?.trim()
    .replace(/^bearer\s+/i, "")
    .replace(/\s+/g, "") || "";
}

export function describeTrustMode(scanForm: ScanRequest): { label: string; detail: string; tone: ScanReviewTone } {
  const sanitized = sanitizeScanForm(scanForm);
  if (sanitized.connectionMethod !== "api_endpoint") {
    return {
      label: "Inherited from login or kubeconfig",
      detail: "The active kubectl or kubeconfig trust chain will be used.",
      tone: "neutral",
    };
  }
  if (sanitized.insecure) {
    return {
      label: "Skip verification",
      detail: "TLS certificate verification is disabled. Use this only as a temporary troubleshooting path in a trusted environment.",
      tone: "critical",
    };
  }
  if (sanitized.caCertPath) {
    return {
      label: "CA file",
      detail: "TLS trust will come from the selected CA certificate file.",
      tone: "success",
    };
  }
  if (sanitized.caCertContent) {
    return {
      label: "Pasted CA",
      detail: "TLS trust will come from the pasted PEM certificate.",
      tone: "success",
    };
  }
  return {
    label: "System trust",
    detail: "Use this when the API server certificate chains to a CA already trusted on this machine.",
    tone: "neutral",
  };
}

export function describeAuthMode(scanForm: ScanRequest): { label: string; detail: string; tone: ScanReviewTone } {
  const sanitized = sanitizeScanForm(scanForm);
  if (sanitized.connectionMethod !== "api_endpoint") {
    return {
      label: "Inherited credentials",
      detail: "Authentication comes from the current kubectl login or selected kubeconfig.",
      tone: "neutral",
    };
  }
  if (!sanitized.bearerToken) {
    return {
      label: "Bearer token missing",
      detail: "Paste a short-lived Kubernetes API token before preflight.",
      tone: "critical",
    };
  }
  return {
    label: "Bearer token",
    detail: "The app stores the token only in memory for this session and never echoes it back in the review summary.",
    tone: "success",
  };
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

function isExplicitHttpEndpoint(endpoint?: string) {
  return endpoint?.trim().toLowerCase().startsWith("http://") || false;
}

function isLikelyPrivateEndpoint(endpoint?: string) {
  const host = extractHost(endpoint);
  if (!host) {
    return false;
  }
  const normalized = host.toLowerCase();
  if (
    normalized === "localhost" ||
    normalized.endsWith(".local") ||
    normalized.endsWith(".internal") ||
    normalized.endsWith(".cluster.local") ||
    (!normalized.includes(".") && !isIpAddress(normalized))
  ) {
    return true;
  }
  if (!isIpAddress(normalized)) {
    return false;
  }
  const [first, second] = normalized.split(".").map((segment) => Number(segment));
  return (
    first === 10 ||
    first === 127 ||
    (first === 169 && second === 254) ||
    (first === 172 && second >= 16 && second <= 31) ||
    (first === 192 && second === 168)
  );
}

function extractHost(endpoint?: string) {
  const trimmed = endpoint?.trim();
  if (!trimmed) {
    return "";
  }
  try {
    const candidate = /^[a-z]+:\/\//i.test(trimmed) ? trimmed : `https://${trimmed}`;
    return new URL(candidate).hostname;
  } catch {
    return "";
  }
}

function isIpAddress(value: string) {
  return /^(\d{1,3}\.){3}\d{1,3}$/.test(value);
}
