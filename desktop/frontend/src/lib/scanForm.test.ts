import { describe, expect, it } from "vitest";
import {
  describeAuthMode,
  describeTrustMode,
  haveContextInputsChanged,
  inspectScanForm,
  normalizeBearerToken,
  sanitizeScanForm,
  validateContextDiscovery,
} from "./scanForm";
import type { ScanRequest } from "./types";

describe("scanForm helpers", () => {
  it("clears stale kubeconfig and endpoint credentials when current login is selected", () => {
    const scanForm: ScanRequest = {
      connectionMethod: "current",
      contextName: " prod-east ",
      kubeconfigPath: "C:/Users/eric/.kube/config",
      kubeconfigContent: "apiVersion: v1",
      apiServerEndpoint: "https://cluster.example.net:6443",
      bearerToken: "secret-token",
      caCertPath: "C:/certs/cluster-ca.pem",
      caCertContent: "-----BEGIN CERTIFICATE-----",
    };

    expect(sanitizeScanForm(scanForm)).toEqual(expect.objectContaining({
      connectionMethod: "current",
      contextName: "prod-east",
      kubeconfigPath: "",
      kubeconfigContent: "",
      apiServerEndpoint: "",
      bearerToken: "",
      caCertPath: "",
      caCertContent: "",
    }));
  });

  it("normalizes raw token input by removing bearer prefixes and whitespace", () => {
    expect(normalizeBearerToken("  Bearer eyJhbGciOiAi... \n")).toBe("eyJhbGciOiAi...");
    expect(normalizeBearerToken("line-one\nline-two")).toBe("line-oneline-two");
  });

  it("surfaces field-level validation and operator risk flags for direct API scans", () => {
    const validation = inspectScanForm({
      connectionMethod: "api_endpoint",
      apiServerEndpoint: "http://10.0.0.15:6443",
      outputDir: "./out",
      timeoutSeconds: 60,
      insecure: true,
    });

    expect(validation.fieldErrors.bearerToken).toMatch(/bearer token/i);
    expect(validation.fieldErrors.insecureAcknowledgement).toMatch(/skip-TLS/i);
    expect(validation.fieldWarnings.apiServerEndpoint).toMatch(/prefer https/i);
    expect(validation.riskFlags).toContain("Endpoint is explicitly using HTTP instead of HTTPS.");
    expect(validation.riskFlags).toContain("TLS verification is disabled for this connection.");
  });

  it("describes API auth and trust posture without echoing the token", () => {
    expect(describeAuthMode({ connectionMethod: "api_endpoint" })).toEqual(expect.objectContaining({
      label: "Bearer token missing",
      tone: "critical",
    }));

    expect(describeTrustMode({ connectionMethod: "api_endpoint", caCertPath: "C:/certs/cluster-ca.pem" })).toEqual(expect.objectContaining({
      label: "CA file",
      tone: "success",
    }));
  });

  it("validates context discovery inputs separately from the full scan launch", () => {
    expect(validateContextDiscovery({ connectionMethod: "kubeconfig_file" })).toEqual([
      "Choose a kubeconfig file or switch to Use existing access.",
    ]);
  });

  it("detects when connection source inputs changed", () => {
    expect(haveContextInputsChanged(
      { connectionMethod: "current" },
      { connectionMethod: "kubeconfig_file", kubeconfigPath: "C:/Users/eric/.kube/config" },
    )).toBe(true);
    expect(haveContextInputsChanged(
      { connectionMethod: "kubeconfig_file", kubeconfigPath: "C:/Users/eric/.kube/config" },
      { connectionMethod: "kubeconfig_file", kubeconfigPath: "C:/Users/eric/.kube/config", contextName: "prod-east" },
    )).toBe(false);
  });
});
