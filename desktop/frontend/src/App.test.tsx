import { fireEvent, render, screen } from "@testing-library/react";
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
    expect(screen.getByText("Machine readiness")).toBeInTheDocument();
  });

  it("renders navigation and opens the remote scan setup", async () => {
    render(<App />);

    expect(await screen.findByRole("heading", { name: "K8 Visualizer" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "New Scan" }));

    expect(screen.getByText("Connect, validate, scope, and launch")).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /use existing access/i })).toBeChecked();
    expect(screen.getByText(/1\. Choose how to connect/i)).toBeInTheDocument();
    expect(screen.getByText("Connection assistant")).toBeInTheDocument();
    expect(screen.queryByText("No bundle loaded yet")).not.toBeInTheDocument();
  });

  it("supports keyboard shortcuts for new scan, home, and open bundle", async () => {
    vi.spyOn(mockBackend, "PickBundleFile").mockResolvedValueOnce("");

    render(<App />);

    expect(await screen.findByText("Assess a cluster or reopen a saved bundle")).toBeInTheDocument();

    fireEvent.keyDown(window, { key: "n", ctrlKey: true });
    expect(screen.getByText("Connect, validate, scope, and launch")).toBeInTheDocument();

    fireEvent.keyDown(window, { key: "h", ctrlKey: true });
    expect(await screen.findByText("Assess a cluster or reopen a saved bundle")).toBeInTheDocument();

    fireEvent.keyDown(window, { key: "o", ctrlKey: true });
    expect(mockBackend.PickBundleFile).toHaveBeenCalledTimes(1);
    expect(await screen.findByText("Bundle open canceled.", { selector: "p.notice" })).toBeInTheDocument();
  });

  it("ignores keyboard shortcuts while focus is inside contenteditable text", async () => {
    render(<App />);

    expect(await screen.findByText("Assess a cluster or reopen a saved bundle")).toBeInTheDocument();

    const editor = document.createElement("div");
    editor.setAttribute("contenteditable", "true");
    editor.textContent = "notes";
    document.body.appendChild(editor);

    fireEvent.keyDown(editor, { key: "n", ctrlKey: true });

    expect(screen.queryByText("Connect, validate, scope, and launch")).not.toBeInTheDocument();
    expect(screen.getByText("Assess a cluster or reopen a saved bundle")).toBeInTheDocument();

    editor.remove();
  });

  it("uses a first-run home state that explains how the app works", async () => {
    window.history.replaceState({}, "", "/?view=home&firstRun=1");

    render(<App />);

    expect(await screen.findByText("Run one scan, then work from the bundle")).toBeInTheDocument();
    expect(screen.getByText("How K8V works")).toBeInTheDocument();
    expect(screen.getByText("What a scan produces")).toBeInTheDocument();
    expect(screen.getAllByText("Machine readiness").length).toBeGreaterThan(0);
    expect(screen.getAllByText("kubectl CLI (optional)").length).toBeGreaterThan(0);
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
    expect(screen.getByText("More actions")).toBeInTheDocument();
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

  it("resets the scan form after confirmation", async () => {
    vi.spyOn(window, "confirm").mockReturnValueOnce(true);

    render(<App />);

    await userEvent.click(await screen.findByRole("button", { name: "New Scan" }));
    expect(screen.queryByRole("button", { name: "Reset form" })).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("radio", { name: /load kubeconfig file/i }));
    expect(screen.getByRole("button", { name: "Reset form" })).toBeInTheDocument();
    await userEvent.type(screen.getByLabelText("Kubeconfig file"), "C:\\temp\\prod-cluster.backup");
    await userEvent.click(screen.getByRole("button", { name: "Reset form" }));

    expect(window.confirm).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("radio", { name: /use existing access/i })).toBeChecked();
    expect(screen.queryByLabelText("Kubeconfig file")).not.toBeInTheDocument();
  });

  it("confirms before leaving a mid-wizard scan from the sidebar", async () => {
    const confirmSpy = vi.spyOn(window, "confirm")
      .mockReturnValueOnce(false)
      .mockReturnValueOnce(true);

    render(<App />);

    await userEvent.click(await screen.findByRole("button", { name: "New Scan" }));
    await userEvent.click(screen.getByRole("radio", { name: /load kubeconfig file/i }));
    await userEvent.type(screen.getByLabelText("Kubeconfig file"), "C:\\temp\\prod-cluster.backup");
    await userEvent.click(screen.getByRole("button", { name: "Continue to validation" }));
    await userEvent.click(screen.getByRole("button", { name: "Home" }));

    expect(confirmSpy).toHaveBeenCalledWith("Leave scan setup? Your connection test and preflight results will be kept, but form changes may be lost.");
    expect(screen.getByText("Connect, validate, scope, and launch")).toBeInTheDocument();

    fireEvent.keyDown(window, { key: "h", ctrlKey: true });
    expect(await screen.findByText("Assess a cluster or reopen a saved bundle")).toBeInTheDocument();
  });

  it("does not prompt when scan setup was only auto-populated by the connection advisor", async () => {
    vi.spyOn(mockBackend, "GetConnectionAdvisor").mockResolvedValueOnce({
      recommendedMethod: "kubeconfig_file",
      recommendedReason: "Use the detected kubeconfig on this machine.",
      currentLoginAvailable: true,
      currentContext: "prod-east-admin",
      defaultKubeconfigAvailable: true,
      defaultKubeconfigPath: "C:/Users/demo/.kube/prod-cluster.backup",
      defaultKubeconfigCurrentContext: "prod-east-admin",
    });
    const confirmSpy = vi.spyOn(window, "confirm");

    render(<App />);

    await userEvent.click(await screen.findByRole("button", { name: "New Scan" }));
    const kubeconfigRadio = await screen.findByRole("radio", { name: /load kubeconfig file/i });
    expect(kubeconfigRadio).toBeChecked();

    await userEvent.click(screen.getByRole("button", { name: "Continue to validation" }));
    await userEvent.click(screen.getByRole("button", { name: "Home" }));

    expect(confirmSpy).not.toHaveBeenCalled();
    expect(await screen.findByText("Assess a cluster or reopen a saved bundle")).toBeInTheDocument();
  });

  it("does not show Live Run as a permanent sidebar destination", async () => {
    render(<App />);

    expect(await screen.findByRole("button", { name: "Home" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Live Run" })).not.toBeInTheDocument();
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
    expect(screen.getByText("Assessment ready")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Open output folder" })).toBeInTheDocument();
    expect(screen.getByText("More actions")).toBeInTheDocument();
  });

  it("shows a run new scan action from the results topbar", async () => {
    window.history.replaceState({}, "", "/?view=results");

    render(<App />);

    const runNewScan = await screen.findByRole("button", { name: "Run new scan" });
    expect(runNewScan).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Open Existing Bundle" })).not.toBeInTheDocument();

    await userEvent.click(runNewScan);
    expect(screen.getByText("Connect, validate, scope, and launch")).toBeInTheDocument();
  });

  it("dismisses the results completion callout when reviewing findings", async () => {
    render(<App />);

    await userEvent.click(await screen.findByRole("button", { name: "New Scan" }));
    await userEvent.click(screen.getByRole("radio", { name: /load kubeconfig file/i }));
    await userEvent.type(screen.getByLabelText("Kubeconfig file"), "C:\\temp\\prod-cluster.backup");
    await userEvent.click(screen.getByRole("button", { name: "Continue to validation" }));
    await userEvent.click(screen.getByRole("button", { name: "Test connection" }));
    await userEvent.click(screen.getByRole("button", { name: "Continue to review" }));
    await userEvent.click(await screen.findByRole("button", { name: "Run preflight" }));
    await userEvent.click(await screen.findByRole("button", { name: "Start scan" }));

    expect(await screen.findByRole("heading", { name: "Assessment complete" }, { timeout: 4000 })).toBeInTheDocument();
    await userEvent.click(screen.getByText("More actions"));
    await userEvent.click(await screen.findByRole("button", { name: "Review results" }));
    await userEvent.click(await screen.findByRole("button", { name: "Review findings" }));

    expect(await screen.findByRole("group", { name: "Finding severity filters" })).toBeInTheDocument();
    expect(screen.queryByText("Scan complete. Outputs ready.")).not.toBeInTheDocument();
  });

  it("opens findings from the home watch panel without leaving the completion callout behind", async () => {
    render(<App />);

    await userEvent.click(await screen.findByRole("button", { name: "New Scan" }));
    await userEvent.click(screen.getByRole("radio", { name: /load kubeconfig file/i }));
    await userEvent.type(screen.getByLabelText("Kubeconfig file"), "C:\\temp\\prod-cluster.backup");
    await userEvent.click(screen.getByRole("button", { name: "Continue to validation" }));
    await userEvent.click(screen.getByRole("button", { name: "Test connection" }));
    await userEvent.click(screen.getByRole("button", { name: "Continue to review" }));
    await userEvent.click(await screen.findByRole("button", { name: "Run preflight" }));
    await userEvent.click(await screen.findByRole("button", { name: "Start scan" }));

    await userEvent.click(await screen.findByRole("button", { name: "Home" }));
    await userEvent.click(await screen.findByRole("button", { name: "Review Findings" }));

    expect(await screen.findByRole("group", { name: "Finding severity filters" })).toBeInTheDocument();
    expect(screen.queryByText("Scan complete. Outputs ready.")).not.toBeInTheDocument();
  });

  it("switches results tabs into visible section content", async () => {
    window.history.replaceState({}, "", "/?view=results");

    render(<App />);

    const findingsTab = await screen.findByRole("tab", { name: "Findings" });
    await userEvent.click(findingsTab);
    expect(screen.getByRole("group", { name: "Finding severity filters" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("tab", { name: "Restore Readiness" }));
    expect(screen.getByText("Backup posture")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("tab", { name: "Inventory" }));
    expect(screen.getByRole("tab", { name: "Nodes" })).toBeInTheDocument();
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
