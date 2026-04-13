import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, vi } from "vitest";
import App from "./App";
import { mockBackend } from "./lib/mock";

afterEach(() => {
  vi.restoreAllMocks();
  window.history.replaceState({}, "", "/");
  window.localStorage.clear();
});

describe("desktop shell", () => {
  it("renders the fixture-backed dashboard in browser mode", async () => {
    render(<App />);

    expect(await screen.findByText("Current recovery posture")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Open Projects" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /prod-east/i })).toBeInTheDocument();
    expect(screen.getByText("Trend and compare")).toBeInTheDocument();
  });

  it("renders navigation and opens the guided scan wizard", async () => {
    render(<App />);

    expect(await screen.findByRole("heading", { name: "K8V" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "New Scan" }));

    expect(screen.getByText("Guided scan setup")).toBeInTheDocument();
    expect(screen.getByLabelText("Kubeconfig path")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Run Preflight" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Access" })).toHaveAttribute("aria-controls", "wizard-panel-0");
  });

  it("surfaces bundle picker cancellation instead of failing silently", async () => {
    vi.spyOn(mockBackend, "PickBundleFile").mockResolvedValueOnce("");
    render(<App />);

    await userEvent.click(await screen.findByRole("button", { name: "Open Existing Bundle" }));

    expect(await screen.findByText("Bundle open canceled.", { selector: "p.notice" })).toBeInTheDocument();
  });

  it("surfaces startup settings alerts in the desktop shell", async () => {
    vi.spyOn(mockBackend, "GetStartupAlerts").mockResolvedValueOnce([
      { tone: "error", message: "Saved desktop settings could not be loaded." },
    ]);

    render(<App />);

    expect(await screen.findByText("Saved desktop settings could not be loaded.", { selector: "p.notice" })).toBeInTheDocument();
  });

  it("shows cancel as unavailable when no live run is active", async () => {
    render(<App />);

    await userEvent.click(await screen.findByRole("button", { name: "Live Run" }));

    const cancelButton = screen.getByRole("button", { name: "Cancel Unavailable" });
    expect(cancelButton).toBeDisabled();
    expect(cancelButton).toHaveAttribute("title", "No active run is available to cancel.");
  });
});
