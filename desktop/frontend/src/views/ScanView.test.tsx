import type { ComponentProps } from "react";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { inspectScanForm } from "../lib/scanForm";
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
  const validation = overrides.validation || inspectScanForm(scanForm, { insecureAcknowledged: overrides.insecureAcknowledged });

  return render(
    <ScanView
      busy={false}
      scanForm={scanForm}
      setScanForm={() => undefined}
      preflight={null}
      contextCatalog={null}
      detectingContexts={false}
      validation={validation}
      showValidationErrors={false}
      validationRequest={{ version: 0 }}
      insecureAcknowledged={false}
      onSetInsecureAcknowledged={() => undefined}
      onPreflight={() => undefined}
      onStartScan={() => undefined}
      onDetectContexts={() => undefined}
      onBrowseOutput={() => undefined}
      onBrowseKubeconfig={() => undefined}
      onBrowseCACert={() => undefined}
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

    expect(screen.getByText("When to use API endpoint mode")).toBeInTheDocument();
    expect(screen.getByText("Find the API server URL")).toBeInTheDocument();
    expect(screen.getByText(/kubectl config view --minify/i)).toBeInTheDocument();
    expect(screen.getByText(/kubectl create token <service-account>/i)).toBeInTheDocument();
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
});
