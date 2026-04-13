import { render, screen } from "@testing-library/react";
import { ScanView } from "./ScanView";
import type { PreflightReport, ScanRequest } from "../lib/types";

const baseRequest: ScanRequest = {
  outputDir: "./out",
  profileName: "standard",
  target: "vm",
};

const report: PreflightReport = {
  canRun: false,
  degraded: true,
  scope: "payments",
  checks: [
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
      manifest:
        "apiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole\nmetadata:\n  name: k8v-pods-reader",
    },
  ],
};

describe("ScanView", () => {
  it("shows RBAC remediation details for failed preflight probes", () => {
    render(
      <ScanView
        busy={false}
        wizardStep={0}
        setWizardStep={() => undefined}
        scanForm={baseRequest}
        setScanForm={() => undefined}
        preflight={report}
        onPreflight={() => undefined}
        onStartScan={() => undefined}
        onBrowseOutput={() => undefined}
      />,
    );

    expect(screen.getByText("Namespace-scope access for pods")).toBeInTheDocument();
    expect(screen.getByText(/kubectl auth can-i list pods -n payments/i)).toBeInTheDocument();
    expect(screen.getByText(/kind: ClusterRole/i)).toBeInTheDocument();
  });
});
