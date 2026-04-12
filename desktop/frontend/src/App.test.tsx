import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "./App";

describe("desktop shell", () => {
  it("renders the fixture-backed dashboard in browser mode", async () => {
    render(<App />);

    expect(await screen.findByText("Recent assessment bundles")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Open Projects" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /prod-east/i })).toBeInTheDocument();
    expect(screen.getByText("Trend and compare dashboard")).toBeInTheDocument();
  });

  it("renders navigation and opens the guided scan wizard", async () => {
    render(<App />);

    expect(await screen.findByText("K8s Recovery Visualizer")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "New Scan" }));

    expect(screen.getByText("Guided Scan Wizard")).toBeInTheDocument();
    expect(screen.getByLabelText("Kubeconfig path")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Run Preflight" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /1 Access/i })).toHaveAttribute("aria-controls", "wizard-panel-0");
  });
});
