import type { ComponentProps } from "react";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";
import { inspectConnectionSetup, inspectScanForm } from "../lib/scanForm";
import type { PreflightReport, ScanRequest } from "../lib/types";
import { ScanView } from "./ScanView";

const baseRequest: ScanRequest = {
  connectionMethod: "current",
  outputDir: "./out",
  profileName: "standard",
  target: "vm",
  timeoutSeconds: 60,
};

function renderScanView(overrides: Partial<ComponentProps<typeof ScanView>> = {}) {
  const scanForm = overrides.scanForm || baseRequest;
  const insecureAcknowledged = overrides.insecureAcknowledged ?? false;
  const validation = overrides.validation || inspectScanForm(scanForm, { insecureAcknowledged });
  const connectionValidation =
    overrides.connectionValidation || inspectConnectionSetup(scanForm, { insecureAcknowledged });

  return render(
    <ScanView
      busy={false}
      scanForm={scanForm}
      setScanForm={() => undefined}
      scanStage="connect"
      onSetScanStage={() => undefined}
      connectionAdvisor={null}
      kubeconfigInspection={null}
      connectionTest={null}
      preflight={null}
      contextCatalog={null}
      detectingContexts={false}
      connectionValidation={connectionValidation}
      validation={validation}
      showValidationErrors={false}
      validationRequest={{ version: 0 }}
      insecureAcknowledged={insecureAcknowledged}
      onSetInsecureAcknowledged={() => undefined}
      onTestConnection={() => undefined}
      onInspectKubeconfig={() => undefined}
      onUseDetectedKubeconfig={() => undefined}
      onPreflight={() => undefined}
      onStartScan={() => undefined}
      onDetectContexts={() => undefined}
      onBrowseOutput={() => undefined}
      onBrowseKubeconfig={() => undefined}
      onBrowseCACert={() => undefined}
      onLoadDroppedKubeconfig={() => undefined}
      loadedKubeconfigLabel=""
      {...overrides}
    />,
  );
}

describe("ScanView", () => {
  it("renders the API endpoint guide before preflight runs", () => {
    renderScanView({
      scanForm: {
        ...baseRequest,
        connectionMethod: "api_endpoint",
        apiServerEndpoint: "https://prod-east.example.net:6443",
      },
    });

    expect(screen.getByText("When direct API mode is the right choice")).toBeInTheDocument();
    expect(screen.getByText("1. Find the Kubernetes API server URL")).toBeInTheDocument();
    expect(screen.getAllByText(/kubectl config view --minify/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/kubectl create token <service-account>/i).length).toBeGreaterThan(0);
  });

  it("supports show and hide behavior for the token field", async () => {
    renderScanView({
      scanForm: {
        ...baseRequest,
        connectionMethod: "api_endpoint",
        apiServerEndpoint: "https://prod-east.example.net:6443",
        bearerToken: "token-value",
      },
    });

    const tokenInput = screen.getByLabelText("Bearer token") as HTMLInputElement;
    expect(tokenInput.type).toBe("password");

    await userEvent.click(screen.getByRole("button", { name: "Show" }));
    expect(tokenInput.type).toBe("text");

    await userEvent.click(screen.getByRole("button", { name: "Hide" }));
    expect(tokenInput.type).toBe("password");
  });

  it("shows field-level API validation after a launch attempt", () => {
    const scanForm: ScanRequest = {
      ...baseRequest,
      connectionMethod: "api_endpoint",
      apiServerEndpoint: "http://10.0.0.15:6443",
      insecure: true,
    };

    renderScanView({
      scanForm,
      validation: inspectScanForm(scanForm),
      showValidationErrors: true,
      validationRequest: { version: 1, field: "bearerToken" },
    });

    expect(within(screen.getByLabelText("Bearer token").closest("label") as HTMLElement).getByText(/paste a bearer token/i)).toBeInTheDocument();
    expect(screen.getAllByText(/acknowledge the skip-TLS warning/i).length).toBeGreaterThan(0);
    expect(within(screen.getByLabelText("API server host or URL").closest("label") as HTMLElement).getByText(/prefer https/i)).toBeInTheDocument();
  });

  it("surfaces missing kubeconfig file dependencies during inspection", () => {
    renderScanView({
      scanForm: {
        ...baseRequest,
        connectionMethod: "kubeconfig_file",
        kubeconfigPath: "C:/handoff/prod-cluster.backup",
      },
      kubeconfigInspection: {
        source: "kubeconfig file",
        path: "C:/handoff/prod-cluster.backup",
        currentContext: "prod-east-admin",
        contexts: ["prod-east-admin"],
        clusterCount: 1,
        userCount: 1,
        referencedFiles: [
          "Cluster prod-east certificate authority: C:\\handoff\\certs\\cluster-ca.crt",
          "User prod-east client key: C:\\handoff\\certs\\user.key",
        ],
        missingReferencedFiles: [
          "User prod-east client key: C:\\handoff\\certs\\user.key",
        ],
        summary: "Loaded kubeconfig file with 1 context, 1 cluster, and 1 user entry. This kubeconfig depends on local CA or client certificate files. 1 referenced file is missing on this machine.",
        nextAction: "This kubeconfig is valid, but some referenced CA or client certificate files are missing on this machine.",
      },
    });

    expect(screen.getByText("Missing local file dependencies")).toBeInTheDocument();
    expect(screen.getAllByText(/user prod-east client key/i).length).toBeGreaterThan(0);
    expect(screen.getByText(/portable kubeconfig warning/i)).toBeInTheDocument();
  });

  it("loads a chosen kubeconfig file into paste mode through the dropzone", async () => {
    const onLoadDroppedKubeconfig = vi.fn();
    renderScanView({
      scanForm: {
        ...baseRequest,
        connectionMethod: "kubeconfig_file",
      },
      onLoadDroppedKubeconfig,
    });

    const input = screen.getByLabelText("Choose kubeconfig file for paste mode") as HTMLInputElement;
    const file = new File(["apiVersion: v1\nkind: Config\nclusters: []\ncontexts: []\nusers: []"], "prod-cluster.backup", {
      type: "application/x-yaml",
    });

    await userEvent.upload(input, file);

    await waitFor(() =>
      expect(onLoadDroppedKubeconfig).toHaveBeenCalledWith(
        "prod-cluster.backup",
        expect.stringContaining("apiVersion: v1"),
      ),
    );
  });

  it("groups preflight checks into operator-facing categories", () => {
    const report: PreflightReport = {
      canRun: false,
      degraded: true,
      scope: "payments",
      warnings: ["Secret metadata is still optional and disabled by default."],
      checks: [
        { id: "api", title: "API server reachability", status: "pass", required: true, detail: "Cluster API is reachable." },
        { id: "config", title: "Bearer token loaded", status: "pass", required: true, detail: "Credentials were accepted for the initial handshake." },
        {
          id: "pods",
          title: "Read workloads",
          status: "fail",
          required: true,
          scope: "namespace",
          resource: "pods",
          detail: "Required access missing.",
          hint: "Grant list access to pods and workload controllers.",
          commands: ["kubectl auth can-i list pods -n payments"],
          manifest: "kind: ClusterRole",
        },
      ],
    };

    renderScanView({
      preflight: report,
      scanForm: { ...baseRequest, connectionMethod: "api_endpoint", apiServerEndpoint: "https://prod-east.example.net:6443" },
    });

    expect(screen.getByText("Transport")).toBeInTheDocument();
    expect(screen.getByText("Auth")).toBeInTheDocument();
    expect(screen.getByText("RBAC")).toBeInTheDocument();
    expect(screen.getByText(/kubectl auth can-i list pods -n payments/i)).toBeInTheDocument();
    expect(screen.getByText(/kind: ClusterRole/i)).toBeInTheDocument();
  });

  it("surfaces structured connection diagnosis when the test fails", () => {
    renderScanView({
      scanForm: {
        ...baseRequest,
        connectionMethod: "api_endpoint",
        apiServerEndpoint: "https://prod-east.example.net:6443",
      },
      connectionTest: {
        canConnect: false,
        source: "direct API endpoint",
        server: "https://prod-east.example.net:6443",
        summary: "TLS verification failed.",
        nextAction: "Add the issuing CA certificate or use skip-TLS only as a temporary workaround in a trusted environment.",
        diagnosis: {
          code: "tls_trust",
          label: "TLS trust",
          summary: "The cluster is reachable, but this machine does not trust the API server certificate.",
          detail: "x509: certificate signed by unknown authority",
          nextAction: "Add the issuing CA certificate or use skip-TLS only as a temporary workaround in a trusted environment.",
        },
        checks: [{ id: "transport", title: "TLS and API handshake", status: "fail", detail: "x509: certificate signed by unknown authority" }],
      },
    });

    expect(screen.getAllByText("TLS trust").length).toBeGreaterThan(0);
    expect(screen.getByText(/does not trust the api server certificate/i)).toBeInTheDocument();
  });

  it("tells kubeconfig users when the file was accepted but the cluster is unreachable from this machine", () => {
    renderScanView({
      scanForm: {
        ...baseRequest,
        connectionMethod: "kubeconfig_file",
        kubeconfigPath: "C:/handoff/prod-cluster.backup",
      },
      connectionTest: {
        canConnect: false,
        source: "kubeconfig file",
        server: "https://prod-api.internal:6443",
        contextName: "prod-east-admin",
        summary: "The kubeconfig was accepted, but the cluster API it points to is not reachable from this machine.",
        nextAction:
          "The file is valid. Check VPN, private DNS, firewall path, proxy, or whether the kubeconfig points at an internal-only control-plane endpoint that only works from a work jumpbox or cluster network.",
        diagnosis: {
          code: "endpoint_unreachable",
          label: "Cluster reachability",
          summary: "The kubeconfig was accepted, but the cluster API it points to is not reachable from this machine.",
          detail: "dial tcp: lookup prod-api.internal: no such host",
          nextAction:
            "The file is valid. Check VPN, private DNS, firewall path, proxy, or whether the kubeconfig points at an internal-only control-plane endpoint that only works from a work jumpbox or cluster network.",
        },
        fieldWarnings: {
          kubeconfigPath:
            "The kubeconfig parsed correctly, but the cluster API inside it is not reachable from this machine.",
        },
        checks: [
          { id: "config", title: "Connection settings loaded", status: "pass", detail: "Kubeconfig loaded." },
          {
            id: "transport",
            title: "Cluster API reachability",
            status: "fail",
            detail: "dial tcp: lookup prod-api.internal: no such host",
          },
        ],
      },
    });

    expect(screen.getAllByText("Cluster reachability").length).toBeGreaterThan(0);
    expect(
      screen.getAllByText(/the kubeconfig was accepted, but the cluster api it points to is not reachable from this machine/i)
        .length,
    ).toBeGreaterThan(0);
    expect(
      within(screen.getByLabelText("Kubeconfig file").closest("label") as HTMLElement).getByText(
        /the kubeconfig parsed correctly, but the cluster api inside it is not reachable from this machine/i,
      ),
    ).toBeInTheDocument();
  });

  it("shows machine readiness guidance when existing access is not ready", () => {
    renderScanView({
      connectionAdvisor: {
        recommendedMethod: "kubeconfig_file",
        recommendedReason:
          "A kubeconfig was found on this machine, but it is not complete enough to use as-is. Bring a complete kubeconfig or another access path instead of relying on existing access.",
        kubectlAvailable: false,
        currentLoginAvailable: false,
        currentContext: "prod-east-admin",
        currentLoginWarning:
          "The detected kubeconfig is valid, but it still references CA or client-certificate files that are missing on this machine.",
        defaultKubeconfigAvailable: true,
        defaultKubeconfigPath: "C:/Users/demo/.kube/prod-cluster.backup",
        defaultKubeconfigPortable: false,
        defaultKubeconfigWarning:
          "The detected kubeconfig is valid, but it still references CA or client-certificate files that are missing on this machine.",
      },
    });

    expect(screen.getByText("Recommended start on this machine")).toBeInTheDocument();
    expect(screen.getByText("Default kubeconfig")).toBeInTheDocument();
    expect(screen.getByText("kubectl CLI (optional)")).toBeInTheDocument();
    expect(screen.getAllByText(/missing on this machine/i).length).toBeGreaterThan(0);
    expect(screen.getByText("No local access was positively detected")).toBeInTheDocument();
  });
});
