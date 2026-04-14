import { render, screen } from "@testing-library/react";
import { ResultsView } from "./ResultsView";
import { mockWorkspace } from "../lib/backend";

describe("ResultsView", () => {
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
      />,
    );

    expect(screen.getByRole("heading", { name: "Domain score drift" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "What got worse" })).toBeInTheDocument();
    expect(screen.getByText("severity-up")).toBeInTheDocument();
  });
});
