import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "./App";

describe("desktop shell", () => {
  it("renders navigation and opens the guided scan wizard", async () => {
    render(<App />);

    expect(await screen.findByText("K8s Recovery Visualizer")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "New Scan" }));

    expect(screen.getByText("Guided Scan Wizard")).toBeInTheDocument();
    expect(screen.getByLabelText("Kubeconfig path")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Run Preflight" })).toBeInTheDocument();
  });
});
