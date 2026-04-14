import { describe, expect, it } from "vitest";
import { haveContextInputsChanged, sanitizeScanForm, validateContextDiscovery, validateScanForm } from "./scanForm";
import type { ScanRequest } from "./types";

describe("sanitizeScanForm", () => {
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

  it("clears kubeconfig state and trims fields for api endpoint mode", () => {
    const scanForm: ScanRequest = {
      connectionMethod: "api_endpoint",
      contextName: "prod-east-admin",
      kubeconfigPath: "C:/Users/eric/.kube/config",
      kubeconfigContent: "apiVersion: v1",
      apiServerEndpoint: " 10.0.0.15:6443 ",
      bearerToken: " token ",
      caCertPath: " C:/certs/cluster-ca.pem ",
      namespaces: [" payments ", "frontend", "payments"],
    };

    expect(sanitizeScanForm(scanForm)).toEqual(expect.objectContaining({
      connectionMethod: "api_endpoint",
      contextName: "",
      kubeconfigPath: "",
      kubeconfigContent: "",
      apiServerEndpoint: "10.0.0.15:6443",
      bearerToken: "token",
      caCertPath: "C:/certs/cluster-ca.pem",
      namespaces: ["payments", "frontend"],
    }));
  });

  it("reports inline validation errors for incomplete direct API scans", () => {
    const errors = validateScanForm({
      connectionMethod: "api_endpoint",
      outputDir: "./out",
      timeoutSeconds: 60,
    });

    expect(errors).toContain("Enter the Kubernetes API server host, IP, or URL.");
    expect(errors).toContain("Paste a bearer token for direct API endpoint scans.");
  });

  it("validates context discovery inputs separately from full scan validation", () => {
    expect(validateContextDiscovery({ connectionMethod: "kubeconfig_file" })).toEqual([
      "Choose a kubeconfig file before loading contexts.",
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
