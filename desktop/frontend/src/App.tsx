import { startTransition, useEffect, useState } from "react";
import { backend, mockWorkspace } from "./lib/backend";
import type {
  AppAlert,
  ExportRequest,
  PreflightReport,
  ProjectSummary,
  RunEvent,
  ScanRequest,
  Settings,
  Workspace,
} from "./lib/types";
import { applyTheme, exportMessage, titleCase } from "./components/ui";
import { HomeView, ProjectsView } from "./views/HomeProjectsView";
import { ScanView } from "./views/ScanView";
import { LiveView } from "./views/LiveView";
import { ResultsView } from "./views/ResultsView";
import { SettingsView } from "./views/SettingsView";

type View = "home" | "projects" | "scan" | "live" | "results" | "settings";
type BannerTone = AppAlert["tone"];

const navItems: Array<{ id: View; label: string }> = [
  { id: "home", label: "Home" },
  { id: "projects", label: "Projects" },
  { id: "scan", label: "New Scan" },
  { id: "live", label: "Live Run" },
  { id: "results", label: "Results" },
  { id: "settings", label: "Settings" },
];

function initialView(): View {
  const value = new URLSearchParams(window.location.search).get("view");
  if (value && navItems.some((item) => item.id === value)) {
    return value as View;
  }
  return "home";
}

const demoTimestamp = "2026-04-12T14:11:00Z";

function isBrowserDemo() {
  return typeof window !== "undefined" && !window.go?.main?.App;
}

function initialLiveEvents(): RunEvent[] {
  if (new URLSearchParams(window.location.search).get("view") !== "live") {
    return [];
  }
  return [
    { type: "status", runId: "demo-live", timestamp: demoTimestamp, step: "preflight", level: "info", message: "Preflight checks complete.", percent: 0.12 },
    { type: "status", runId: "demo-live", timestamp: demoTimestamp, step: "connect", level: "info", message: "Connecting to the Kubernetes API.", percent: 0.28 },
    { type: "status", runId: "demo-live", timestamp: demoTimestamp, step: "Images", level: "info", message: "Collecting Images.", percent: 0.62 },
    { type: "warning", runId: "demo-live", timestamp: demoTimestamp, step: "Secrets", level: "warn", message: "Secrets skipped.", percent: 0.72, warning: "forbidden: secrets access intentionally withheld" },
  ];
}

function formatActionError(error: unknown, fallback: string) {
  if (error instanceof Error && error.message.trim()) {
    return error.message.trim();
  }
  if (typeof error === "string" && error.trim()) {
    return error.trim();
  }
  return fallback;
}

function isCanceledError(error: unknown) {
  const message = formatActionError(error, "");
  return /\bcancel(?:ed|led)\b/i.test(message);
}

function defaultBundleInputPath(browserDemo: boolean) {
  if (!browserDemo) {
    return "";
  }
  return mockWorkspace.artifacts.loadedBundlePath || "";
}

function defaultScanForm(browserDemo: boolean): ScanRequest {
  return {
    outputDir: "./out",
    target: "vm",
    minScore: 90,
    timeoutSeconds: 60,
    profileName: browserDemo ? "enterprise" : "standard",
    namespaces: browserDemo ? ["payments", "frontend", "platform"] : [],
    compareTo: browserDemo ? "./demo-out/history/recovery-scan-20260405-141100.json" : "",
    summary: true,
    runbook: true,
    redact: false,
    csvExport: true,
    dryRun: false,
    includeSecretMetadata: false,
    clusterName: browserDemo ? "prod-east" : "",
    environment: browserDemo ? "production" : "",
    customerId: browserDemo ? "acme-hospitality" : "",
    site: browserDemo ? "us-east-1a" : "",
    contextName: browserDemo ? "prod-east-admin" : "",
    kubeconfigPath: browserDemo ? "~/.kube/config" : "",
  };
}

export default function App() {
  const browserDemo = isBrowserDemo();
  const [view, setView] = useState<View>(initialView);
  const [wizardStep, setWizardStep] = useState(0);
  const [resultTab, setResultTab] = useState(new URLSearchParams(window.location.search).get("tab") || "Summary");
  const [settings, setSettings] = useState<Settings>({
    workspaceRoot: ".",
    defaultOutputDir: "./out",
    defaultProfile: "enterprise",
    includeSecretMetadata: false,
    summary: true,
    runbook: true,
    redact: false,
    csvExport: true,
  });
  const [projects, setProjects] = useState<ProjectSummary[]>([]);
  const [workspace, setWorkspace] = useState<Workspace | null>(browserDemo ? mockWorkspace : null);
  const [preflight, setPreflight] = useState<PreflightReport | null>(null);
  const [events, setEvents] = useState<RunEvent[]>(initialLiveEvents());
  const [activeRunId, setActiveRunId] = useState<string | null>(
    new URLSearchParams(window.location.search).get("view") === "live" ? "demo-live" : null,
  );
  const [busy, setBusy] = useState(false);
  const [statusMessage, setStatusMessage] = useState("Workspace ready.");
  const [actionBanner, setActionBanner] = useState<{ tone: BannerTone; message: string } | null>(null);
  const [openBundlePath, setOpenBundlePath] = useState(defaultBundleInputPath(browserDemo));
  const [exportNotice, setExportNotice] = useState("");
  const [findingFilter, setFindingFilter] = useState("ALL");
  const [scanForm, setScanForm] = useState<ScanRequest>(defaultScanForm(browserDemo));

  useEffect(() => {
    let active = true;

    void (async () => {
      try {
        const bootstrap = await backend.getBootstrap();
        if (active) {
          applyTheme(bootstrap.theme);
        }
      } catch (error) {
        if (active) {
          showBanner("error", `Theme load failed: ${formatActionError(error, "Could not load the desktop theme.")}`);
        }
      }

      try {
        const saved = await backend.getSettings();
        if (!active) {
          return;
        }
        setSettings(saved);
        setScanForm((current) => ({
          ...current,
          outputDir: saved.defaultOutputDir || current.outputDir,
          profileName: saved.defaultProfile || current.profileName,
          summary: saved.summary,
          runbook: saved.runbook,
          redact: saved.redact,
          csvExport: saved.csvExport,
          includeSecretMetadata: saved.includeSecretMetadata,
        }));
      } catch (error) {
        if (active) {
          showBanner("error", `Settings load failed: ${formatActionError(error, "Could not load desktop settings.")}`);
          setStatusMessage("Desktop settings failed to load.");
        }
      }

      try {
        const alerts = await backend.getStartupAlerts();
        if (active && alerts.length > 0) {
          setActionBanner(alerts[0]);
          if (alerts[0].tone === "error") {
            setStatusMessage("Desktop settings need attention.");
          }
        }
      } catch {
        // Startup alerts are optional and should not block the shell.
      }

      try {
        const nextProjects = await backend.listProjects();
        if (active) {
          setProjects(nextProjects);
        }
      } catch (error) {
        if (active) {
          showBanner("error", `Workspace discovery failed: ${formatActionError(error, "Could not load saved projects.")}`);
          setStatusMessage("Workspace discovery failed.");
        }
      }
    })();

    const unsubscribe = backend.onScanEvent((event) => {
      setEvents((current) => [...current, event]);
      setStatusMessage(event.message);
      if (event.type === "complete") {
        setActiveRunId(null);
      }
    });
    return () => {
      active = false;
      unsubscribe();
    };
  }, []);

  const bundle = workspace?.bundle;
  const activePercent = events.length > 0 ? events[events.length - 1].percent || 0 : 0;
  const currentViewLabel = navItems.find((item) => item.id === view)?.label || "Workspace";
  const maturityTone = (bundle?.score.maturity || (browserDemo ? "GOLD" : "SILVER")).toLowerCase();
  const bundleScore = bundle?.score.overall.final ?? (browserDemo ? 85 : "—");
  const environmentLabel = bundle?.metadata.environment || (browserDemo ? "production" : "ready");
  const clusterLabel = bundle?.metadata.clusterName || (browserDemo ? "demo-workspace" : "no bundle");
  const statusDetail = bundle
    ? `${bundle.cluster.platform?.provider || "Unknown platform"} ${bundle.cluster.platform?.k8sVersion || ""}`.trim()
    : browserDemo
      ? "Fixture bundle loaded for preview."
      : "Open an existing bundle or start a new scan.";

  function showBanner(tone: BannerTone, message: string) {
    setActionBanner({ tone, message });
  }

  function clearBanner() {
    setActionBanner(null);
  }

  async function handlePreflight() {
    clearBanner();
    setBusy(true);
    try {
      const report = await backend.runPreflight(scanForm);
      setPreflight(report);
      setStatusMessage(report.canRun ? "Preflight checks passed." : "Preflight found blocking issues.");
    } catch (error) {
      const message = formatActionError(error, "Could not run preflight.");
      showBanner("error", `Preflight failed: ${message}`);
      setStatusMessage("Preflight failed.");
    } finally {
      setBusy(false);
    }
  }

  async function handleStartScan() {
    const runId = `run-${Date.now()}`;
    setBusy(true);
    setEvents([]);
    setActiveRunId(runId);
    setExportNotice("");
    clearBanner();
    startTransition(() => setView("live"));
    try {
      const result = await backend.runScan({ ...scanForm, runId });
      setPreflight(result.preflight);
      setWorkspace(result.workspace);
      setProjects((current) => {
        const nextProject: ProjectSummary = {
          name: result.workspace.bundle.metadata.clusterName || "latest-run",
          clusterName: result.workspace.bundle.metadata.clusterName,
          environment: result.workspace.bundle.metadata.environment,
          outputDir: result.artifacts.outputDir,
          lastScanPath: result.artifacts.bundleJson || "",
          reportPath: result.artifacts.htmlReport,
          score: result.workspace.bundle.score.overall.final,
          maturity: result.workspace.bundle.score.maturity,
          timestampUtc: result.workspace.bundle.metadata.generatedAt,
        };
        return [nextProject, ...current.filter((item) => item.lastScanPath !== nextProject.lastScanPath)];
      });
      setResultTab("Summary");
      startTransition(() => setView("results"));
      setStatusMessage("Scan finished. Results workspace is ready.");
    } catch (error) {
      if (isCanceledError(error)) {
        showBanner("info", "Scan canceled.");
        setStatusMessage("Scan canceled.");
      } else {
        const message = formatActionError(error, "Could not complete the scan.");
        showBanner("error", `Scan failed: ${message}`);
        setStatusMessage("Scan failed.");
      }
    } finally {
      setBusy(false);
      setActiveRunId(null);
    }
  }

  async function handleCancelRun() {
    if (!activeRunId) {
      showBanner("info", "No active run is available to cancel.");
      setStatusMessage("No active run to cancel.");
      return;
    }
    clearBanner();
    try {
      await backend.cancelRun(activeRunId);
      setStatusMessage("Cancel request sent.");
      setActiveRunId(null);
    } catch (error) {
      const message = formatActionError(error, "Could not cancel the active run.");
      showBanner("error", `Cancel failed: ${message}`);
      setStatusMessage("Cancel failed.");
    }
  }

  async function handleOpenBundlePath(path: string) {
    clearBanner();
    const resolved = path.trim();
    if (!resolved) {
      showBanner("info", "Enter a bundle path or use Open Existing Bundle.");
      setStatusMessage("No bundle path provided.");
      return;
    }
    setBusy(true);
    setExportNotice("");
    try {
      const loaded = await backend.openBundle(resolved);
      setWorkspace(loaded);
      setOpenBundlePath(resolved);
      setResultTab("Summary");
      startTransition(() => setView("results"));
      setStatusMessage("Existing bundle loaded into the workspace.");
    } catch (error) {
      const message = formatActionError(error, "Could not open the selected bundle.");
      showBanner("error", `Open bundle failed: ${message}`);
      setStatusMessage("Open bundle failed.");
    } finally {
      setBusy(false);
    }
  }

  async function handlePickBundle() {
    clearBanner();
    let resolved = "";
    try {
      resolved = await backend.pickBundleFile();
    } catch (error) {
      const message = formatActionError(error, "Could not open the file picker.");
      showBanner("error", `Open bundle failed: ${message}`);
      setStatusMessage("Open bundle failed.");
      return;
    }
    if (!resolved) {
      showBanner("info", "Bundle open canceled.");
      setStatusMessage("Bundle open canceled.");
      return;
    }
    await handleOpenBundlePath(resolved);
  }

  async function handleBrowseOutput() {
    clearBanner();
    try {
      const outputDir = await backend.pickOutputDirectory();
      if (outputDir) {
        setScanForm((current) => ({ ...current, outputDir }));
        setStatusMessage("Output directory updated.");
        return;
      }
      showBanner("info", "Output directory selection canceled.");
      setStatusMessage("Output directory selection canceled.");
    } catch (error) {
      const message = formatActionError(error, "Could not open the output directory picker.");
      showBanner("error", `Output directory failed: ${message}`);
      setStatusMessage("Output directory selection failed.");
    }
  }

  async function handleSaveSettings() {
    clearBanner();
    try {
      await backend.saveSettings(settings);
      setStatusMessage("Desktop settings saved.");
    } catch (error) {
      const message = formatActionError(error, "Could not save desktop settings.");
      showBanner("error", `Save settings failed: ${message}`);
      setStatusMessage("Save settings failed.");
    }
  }

  async function handleExport(kind: "report" | "summary" | "runbook" | "json" | "csv" | "redacted") {
    const bundlePath = workspace?.artifacts.loadedBundlePath || workspace?.artifacts.bundleJson;
    if (!bundlePath) {
      showBanner("info", "Load a bundle before exporting artifacts.");
      setStatusMessage("No bundle loaded for export.");
      return;
    }
    clearBanner();
    const request: ExportRequest = {
      outputDir: workspace?.artifacts.outputDir || settings.defaultOutputDir || "./out",
      report: kind === "report",
      bundleJson: kind === "json",
      summary: kind === "summary",
      runbook: kind === "runbook",
      csvExport: kind === "csv",
      redact: kind === "redacted",
    };
    try {
      const artifacts = await backend.exportBundle(bundlePath, request);
      setExportNotice(exportMessage(kind, artifacts));
      setStatusMessage(`${titleCase(kind)} export refreshed in ${artifacts.outputDir}.`);
    } catch (error) {
      const message = formatActionError(error, `Could not export ${kind}.`);
      showBanner("error", `Export failed: ${message}`);
      setStatusMessage(`${titleCase(kind)} export failed.`);
    }
  }

  return (
    <div className="app-shell">
      <aside className="sidebar" aria-label="Primary navigation">
        <div className="brand">
          <div className="brand-mark" aria-hidden="true">K8V</div>
          <div className="brand-copy">
            <h1>K8V</h1>
            <p className="muted">Kubernetes recovery visualizer</p>
          </div>
        </div>
        <nav className="nav">
          {navItems.map((item) => (
            <button key={item.id} type="button" className={`nav-link ${view === item.id ? "is-active" : ""}`} onClick={() => setView(item.id)} aria-current={view === item.id ? "page" : undefined}>
              {item.label}
            </button>
          ))}
        </nav>
        <div className="sidebar-card">
          <p className="eyebrow">Active Bundle</p>
          <div className="hero-score-inline">
            <strong>{bundleScore}</strong>
            <span className={`tone tone-${maturityTone}`}>{bundle?.score.maturity || (browserDemo ? "GOLD" : "UNLOADED")}</span>
          </div>
          <dl className="sidebar-kv">
            <div>
              <dt>Cluster</dt>
              <dd>{clusterLabel}</dd>
            </div>
            <div>
              <dt>Environment</dt>
              <dd>{environmentLabel}</dd>
            </div>
            <div>
              <dt>Output</dt>
              <dd>{workspace?.artifacts.outputDir || "not loaded"}</dd>
            </div>
          </dl>
        </div>
      </aside>

      <div className="workspace">
        <header className="topbar">
          <div className="status-stack">
            <p className="eyebrow">{currentViewLabel}</p>
            <h2>{statusMessage}</h2>
            <p className="muted">{statusDetail}</p>
          </div>
          <div className="topbar-actions">
            <span className="chip">{clusterLabel}</span>
            <span className="chip">{environmentLabel}</span>
            {bundle ? <span className="chip">{bundle.score.maturity} · {bundle.score.overall.final}</span> : null}
            <button type="button" className="button secondary" onClick={handlePickBundle} disabled={busy}>
              Open Existing Bundle
            </button>
          </div>
        </header>
        {actionBanner ? <p className={`notice notice-${actionBanner.tone}`}>{actionBanner.message}</p> : null}

        <main className="main-content">
          {view === "home" && <HomeView workspace={workspace} projects={projects} busy={busy} onViewProjects={() => setView("projects")} onPickBundle={handlePickBundle} onOpenProject={handleOpenBundlePath} />}
          {view === "projects" && <ProjectsView projects={projects} busy={busy} onPickBundle={handlePickBundle} />}
          {view === "scan" && <ScanView busy={busy} wizardStep={wizardStep} setWizardStep={setWizardStep} scanForm={scanForm} setScanForm={setScanForm} preflight={preflight} onPreflight={handlePreflight} onStartScan={handleStartScan} onBrowseOutput={handleBrowseOutput} />}
          {view === "live" && <LiveView events={events} activeRunId={activeRunId} activePercent={activePercent} onCancel={handleCancelRun} />}
          {view === "results" && workspace && <ResultsView workspace={workspace} resultTab={resultTab} setResultTab={setResultTab} findingFilter={findingFilter} setFindingFilter={setFindingFilter} exportNotice={exportNotice} onExport={handleExport} />}
          {view === "settings" && <SettingsView settings={settings} busy={busy} setSettings={setSettings} openBundlePath={openBundlePath} setOpenBundlePath={setOpenBundlePath} onSave={handleSaveSettings} onOpenBundle={() => handleOpenBundlePath(openBundlePath)} />}
        </main>
      </div>
    </div>
  );
}
