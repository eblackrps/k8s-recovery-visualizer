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

    expect(await screen.findByText("Assess a cluster or reopen a saved bundle")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /connect to a cluster/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Open Existing Bundle" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /prod-east/i })).toBeInTheDocument();
    expect(screen.getByText("Saved assessments")).toBeInTheDocument();
  });

  it("renders navigation and opens the remote scan setup", async () => {
    render(<App />);

    expect(await screen.findByRole("heading", { name: "K8V" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "New Scan" }));

    expect(screen.getByText("Connect, validate, scope, and launch")).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /use existing access/i })).toBeChecked();
    expect(screen.getByText(/1\. Choose how to connect/i)).toBeInTheDocument();
    expect(screen.getByText("Connection assistant")).toBeInTheDocument();
  });

  it("uses a first-run home state that explains how the app works", async () => {
    window.history.replaceState({}, "", "/?view=home&firstRun=1");

    render(<App />);

    expect(await screen.findByText("No active bundle loaded yet")).toBeInTheDocument();
    expect(screen.getByText("How K8V works")).toBeInTheDocument();
    expect(screen.getByText("What a scan produces")).toBeInTheDocument();
    expect(screen.getByText("Machine readiness")).toBeInTheDocument();
    expect(screen.getByText("kubectl CLI (optional)")).toBeInTheDocument();
  });

  it("supports an API endpoint scan preset for demo and screenshot routes", async () => {
    window.history.replaceState({}, "", "/?view=scan&scanConnection=api_endpoint");

    render(<App />);

    expect(await screen.findByText("Connect, validate, scope, and launch")).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /api endpoint/i })).toBeChecked();
    expect(screen.getByText("When direct API mode is the right choice")).toBeInTheDocument();
  });

  it("supports a scan-complete preset for demo and screenshot routes", async () => {
    window.history.replaceState({}, "", "/?view=complete");

    render(<App />);

    expect(await screen.findByText("Assessment complete")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Review findings" })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Open bundle JSON" }).length).toBeGreaterThan(0);
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

  it("shows a clear scan-complete handoff after a successful run", async () => {
    render(<App />);

    await userEvent.click(await screen.findByRole("button", { name: "New Scan" }));
    await userEvent.click(screen.getByRole("radio", { name: /load kubeconfig file/i }));
    await userEvent.type(screen.getByLabelText("Kubeconfig file"), "C:\\temp\\prod-cluster.backup");
    await userEvent.click(screen.getByRole("button", { name: "Continue to validation" }));
    await userEvent.click(screen.getByRole("button", { name: "Test connection" }));
    await userEvent.click(screen.getByRole("button", { name: "Continue to review" }));
    await userEvent.click(await screen.findByRole("button", { name: "Run preflight" }));
    await userEvent.click(await screen.findByRole("button", { name: "Start scan" }));

    expect(await screen.findByRole("button", { name: "Review findings" }, { timeout: 4000 })).toBeInTheDocument();
    expect(screen.getByText("Assessment complete")).toBeInTheDocument();
    expect(screen.getByText("The bundle is ready for review and handoff")).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Open output folder" }).length).toBeGreaterThan(0);
    expect(screen.getAllByRole("button", { name: "Open bundle JSON" }).length).toBeGreaterThan(0);
  });

  it("keeps the real scan failure visible in the main status line", async () => {
    vi.spyOn(mockBackend, "RunScan").mockRejectedValueOnce(new Error("load kube config from \"C:\\\\temp\\\\cluster\": exec plugin failed"));

    render(<App />);

    await userEvent.click(await screen.findByRole("button", { name: "New Scan" }));
    await userEvent.click(screen.getByRole("radio", { name: /load kubeconfig file/i }));
    await userEvent.type(screen.getByLabelText("Kubeconfig file"), "C:\\temp\\cluster");
    await userEvent.click(screen.getByRole("button", { name: "Continue to validation" }));
    await userEvent.click(screen.getByRole("button", { name: "Test connection" }));
    await userEvent.click(screen.getByRole("button", { name: "Continue to review" }));
    await userEvent.click(await screen.findByRole("button", { name: "Run preflight" }));
    await userEvent.click(await screen.findByRole("button", { name: "Start scan" }));

    expect(await screen.findByRole("heading", { name: /scan failed: this connection depends on an external auth helper/i })).toBeInTheDocument();
    expect(screen.getByText("External auth helper")).toBeInTheDocument();
    expect(screen.getByText(/exec plugin failed/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Fix scan setup" })).toBeInTheDocument();
  });
});
