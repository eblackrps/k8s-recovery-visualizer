import { startTransition, useEffect, useState } from "react";
import { backend, mockWorkspace } from "./lib/backend";
import type {
  AppAlert,
  ContextCatalog,
  ExportRequest,
  PreflightReport,
  ProjectSummary,
  RunEvent,
  ScanRequest,
  Settings,
  Workspace,
} from "./lib/types";
import { haveContextInputsChanged, prepareScanRequest, sanitizeScanForm, validateContextDiscovery, validateScanForm } from "./lib/scanForm";
import { Badge, applyTheme, exportMessage, titleCase, toneForMaturity } from "./components/ui";
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

function normalizeResultTab(value: string | null | undefined) {
  switch (value) {
    case "Summary":
      return "Overview";
    case "Backup":
      return "Restore Readiness";
    case "DR Score":
      return "Overview";
    default:
      return value || "Overview";
  }
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
    connectionMethod: "current",
    outputDir: "./out",
    target: "vm",
    minScore: 0,
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
  };
}

export default function App() {
  const browserDemo = isBrowserDemo();
  const [view, setView] = useState<View>(initialView);
  const [resultTab, setResultTab] = useState(normalizeResultTab(new URLSearchParams(window.location.search).get("tab")));
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
  const [contextCatalog, setContextCatalog] = useState<ContextCatalog | null>(null);
  const [events, setEvents] = useState<RunEvent[]>(initialLiveEvents());
  const [activeRunId, setActiveRunId] = useState<string | null>(
    new URLSearchParams(window.location.search).get("view") === "live" ? "demo-live" : null,
  );
  const [busy, setBusy] = useState(false);
  const [detectingContexts, setDetectingContexts] = useState(false);
  const [statusMessage, setStatusMessage] = useState("Workspace ready.");
  const [actionBanner, setActionBanner] = useState<{ tone: BannerTone; message: string } | null>(null);
  const [openBundlePath, setOpenBundlePath] = useState(defaultBundleInputPath(browserDemo));
  const [exportNotice, setExportNotice] = useState("");
  const [findingFilter, setFindingFilter] = useState("ALL");
  const [scanForm, setScanForm] = useState<ScanRequest>(defaultScanForm(browserDemo));
  const validationErrors = validateScanForm(scanForm);

  function updateScanForm(updater: (current: ScanRequest) => ScanRequest) {
    const next = sanitizeScanForm(updater(scanForm));
    setPreflight(null);
    if (haveContextInputsChanged(scanForm, next)) {
      setContextCatalog(null);
    }
    setScanForm(next);
  }

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
        setScanForm((current) => sanitizeScanForm({
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
  const environmentLabel = bundle?.metadata.environment || (browserDemo ? "production" : "ready");
  const clusterLabel = bundle?.metadata.clusterName || (browserDemo ? "demo-workspace" : "no bundle");
  const providerLabel = bundle?.cluster.platform?.provider || (browserDemo ? "fixture bundle" : "no bundle");
  const versionLabel = bundle?.cluster.platform?.k8sVersion || (browserDemo ? "preview" : "idle");
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
    const request = prepareScanRequest(scanForm);
    const errors = validateScanForm(request);
    if (errors.length > 0) {
      showBanner("info", errors[0]);
      setStatusMessage("Scan setup needs attention before preflight.");
      return;
    }
    clearBanner();
    setBusy(true);
    try {
      setScanForm(request);
      const report = await backend.runPreflight(request);
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
    const request = prepareScanRequest(scanForm);
    const errors = validateScanForm(request);
    if (errors.length > 0) {
      showBanner("info", errors[0]);
      setStatusMessage("Scan setup needs attention before the run can start.");
      return;
    }
    const runId = `run-${Date.now()}`;
    setBusy(true);
    setEvents([]);
    setActiveRunId(runId);
    setExportNotice("");
    clearBanner();
    startTransition(() => setView("live"));
    try {
      setScanForm(request);
      const result = await backend.runScan({ ...request, runId });
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
      setResultTab("Overview");
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
      setResultTab("Overview");
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
        updateScanForm((current) => ({ ...current, outputDir }));
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

  async function handleBrowseKubeconfig() {
    clearBanner();
    try {
      const kubeconfigPath = await backend.pickKubeconfigFile();
      if (kubeconfigPath) {
        updateScanForm((current) => ({ ...current, kubeconfigPath }));
        setStatusMessage("Kubeconfig file updated.");
        return;
      }
      showBanner("info", "Kubeconfig selection canceled.");
      setStatusMessage("Kubeconfig selection canceled.");
    } catch (error) {
      const message = formatActionError(error, "Could not open the kubeconfig picker.");
      showBanner("error", `Kubeconfig selection failed: ${message}`);
      setStatusMessage("Kubeconfig selection failed.");
    }
  }

  async function handleBrowseCACert() {
    clearBanner();
    try {
      const caCertPath = await backend.pickCertificateFile();
      if (caCertPath) {
        updateScanForm((current) => ({ ...current, caCertPath }));
        setStatusMessage("CA certificate updated.");
        return;
      }
      showBanner("info", "CA certificate selection canceled.");
      setStatusMessage("CA certificate selection canceled.");
    } catch (error) {
      const message = formatActionError(error, "Could not open the CA certificate picker.");
      showBanner("error", `CA certificate selection failed: ${message}`);
      setStatusMessage("CA certificate selection failed.");
    }
  }

  async function handleDetectContexts() {
    const request = prepareScanRequest(scanForm);
    const errors = validateContextDiscovery(request);
    if (errors.length > 0) {
      showBanner("info", errors[0]);
      setStatusMessage("Connection details need attention before contexts can be loaded.");
      return;
    }
    clearBanner();
    setDetectingContexts(true);
    try {
      const catalog = await backend.listConnectionContexts(request);
      setContextCatalog(catalog);
      const nextForm = catalog.currentContext && !request.contextName
        ? sanitizeScanForm({ ...request, contextName: catalog.currentContext })
        : request;
      setPreflight(null);
      setScanForm(nextForm);
      if (catalog.contexts?.length) {
        setStatusMessage(`Loaded ${catalog.contexts.length} context${catalog.contexts.length === 1 ? "" : "s"} from ${catalog.source || "the current login"}.`);
      } else {
        setStatusMessage("No named contexts were found for this connection.");
      }
    } catch (error) {
      const message = formatActionError(error, "Could not load connection contexts.");
      showBanner("error", `Context discovery failed: ${message}`);
      setStatusMessage("Context discovery failed.");
    } finally {
      setDetectingContexts(false);
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
        <div className="sidebar-footer">
          <p className="eyebrow">Workspace</p>
          <p className="muted">Run remote assessments, inspect saved bundles, and review restore readiness without leaving the desktop console.</p>
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
            <div className="context-strip" aria-label="Active bundle context">
              <div className="context-cluster">
                <strong>{clusterLabel}</strong>
                <span className="muted">{environmentLabel}</span>
              </div>
              <div className="context-meta">
                <Badge>{providerLabel}</Badge>
                <Badge>{versionLabel}</Badge>
                <Badge tone={toneForMaturity(bundle?.score.maturity || (browserDemo ? "GOLD" : ""))}>
                  {bundle ? `${bundle.score.maturity} ${bundle.score.overall.final}` : "No bundle"}
                </Badge>
              </div>
            </div>
            <button type="button" className="button secondary quiet" onClick={handlePickBundle} disabled={busy}>
              Open Existing Bundle
            </button>
          </div>
        </header>
        {actionBanner ? <p className={`notice notice-${actionBanner.tone}`}>{actionBanner.message}</p> : null}

        <main className="main-content">
          {view === "home" && (
            <HomeView
              workspace={workspace}
              projects={projects}
              busy={busy}
              onViewProjects={() => setView("projects")}
              onPickBundle={handlePickBundle}
              onOpenProject={handleOpenBundlePath}
              onStartScan={() => setView("scan")}
              onReviewFindings={() => {
                setResultTab("Findings");
                setView("results");
              }}
            />
          )}
          {view === "projects" && <ProjectsView projects={projects} busy={busy} onPickBundle={handlePickBundle} />}
          {view === "scan" && <ScanView busy={busy} scanForm={scanForm} setScanForm={updateScanForm} preflight={preflight} contextCatalog={contextCatalog} detectingContexts={detectingContexts} validationErrors={validationErrors} onPreflight={handlePreflight} onStartScan={handleStartScan} onDetectContexts={handleDetectContexts} onBrowseOutput={handleBrowseOutput} onBrowseKubeconfig={handleBrowseKubeconfig} onBrowseCACert={handleBrowseCACert} />}
          {view === "live" && <LiveView events={events} activeRunId={activeRunId} activePercent={activePercent} onCancel={handleCancelRun} />}
          {view === "results" && workspace && <ResultsView workspace={workspace} resultTab={resultTab} setResultTab={setResultTab} findingFilter={findingFilter} setFindingFilter={setFindingFilter} exportNotice={exportNotice} onExport={handleExport} />}
          {view === "settings" && <SettingsView settings={settings} busy={busy} setSettings={setSettings} openBundlePath={openBundlePath} setOpenBundlePath={setOpenBundlePath} onSave={handleSaveSettings} onOpenBundle={() => handleOpenBundlePath(openBundlePath)} />}
        </main>
      </div>
    </div>
  );
}
