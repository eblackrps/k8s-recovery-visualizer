import { render, screen } from "@testing-library/react";
import { ResultsView } from "./ResultsView";
import { mockWorkspace } from "../lib/backend";
import type { RunCompletionSummary } from "../lib/types";

const completionSummary: RunCompletionSummary = {
  runId: "run-123",
  clusterName: "default",
  environment: "Unknown",
  generatedAt: "2026-04-14T20:54:16Z",
  score: 33,
  findingCount: 80,
  hasComparison: false,
  artifacts: mockWorkspace.artifacts,
};

describe("ResultsView", () => {
  it("shows the bundle and report handoff paths", () => {
    render(
      <ResultsView
        workspace={mockWorkspace}
        resultTab="Overview"
        setResultTab={() => undefined}
        findingFilter="ALL"
        setFindingFilter={() => undefined}
        exportNotice=""
        onExport={() => undefined}
        onOpenPath={() => undefined}
      />,
    );

    expect(screen.getByText("Bundle and reports on disk")).toBeInTheDocument();
    expect(screen.getByText("Reopen later")).toBeInTheDocument();
    expect(screen.getAllByText("./demo-out/recovery-scan.json").length).toBeGreaterThan(0);
    expect(screen.getByText("HTML report")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Open output folder" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Open bundle JSON" })).toBeInTheDocument();
  });

  it("renders prioritized findings and richer compare data", () => {
    render(
      <ResultsView
        workspace={mockWorkspace}
        resultTab="Findings"
        setResultTab={() => undefined}
        findingFilter="ALL"
        setFindingFilter={() => undefined}
        exportNotice=""
        onExport={() => undefined}
        onOpenPath={() => undefined}
      />,
    );

    expect(screen.getByText("Platform engineering")).toBeInTheDocument();
    expect(screen.getByText("degraded recovery")).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
  });

  it("renders compare score and regression summaries", () => {
    render(
      <ResultsView
        workspace={mockWorkspace}
        resultTab="Compare"
        setResultTab={() => undefined}
        findingFilter="ALL"
        setFindingFilter={() => undefined}
        exportNotice=""
        onExport={() => undefined}
        onOpenPath={() => undefined}
      />,
    );

    expect(screen.getByRole("heading", { name: "Domain score drift" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "What got worse" })).toBeInTheDocument();
    expect(screen.getByText("severity-up")).toBeInTheDocument();
  });

  it("does not repeat the artifact handoff card when the completion handoff is already visible", () => {
    render(
      <ResultsView
        workspace={mockWorkspace}
        resultTab="Overview"
        setResultTab={() => undefined}
        findingFilter="ALL"
        setFindingFilter={() => undefined}
        exportNotice=""
        completionSummary={completionSummary}
        onExport={() => undefined}
        onOpenPath={() => undefined}
      />,
    );

    expect(screen.getByText("Scan complete. Outputs ready.")).toBeInTheDocument();
    expect(screen.queryByText("Bundle and reports on disk")).not.toBeInTheDocument();
    expect(screen.getByText("More actions")).toBeInTheDocument();
  });
});
