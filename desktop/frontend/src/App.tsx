import { startTransition, useEffect, useState } from "react";
import { backend, mockWorkspace } from "./lib/backend";
import type {
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
  const [statusMessage, setStatusMessage] = useState("Desktop workspace ready.");
  const [openBundlePath, setOpenBundlePath] = useState(mockWorkspace.artifacts.loadedBundlePath || "");
  const [exportNotice, setExportNotice] = useState("");
  const [findingFilter, setFindingFilter] = useState("ALL");
  const [scanForm, setScanForm] = useState<ScanRequest>({
    outputDir: "./out",
    target: "vm",
    minScore: 90,
    timeoutSeconds: 60,
    profileName: "enterprise",
    namespaces: ["payments", "frontend", "platform"],
    compareTo: "./demo-out/history/recovery-scan-20260405-141100.json",
    summary: true,
    runbook: true,
    redact: false,
    csvExport: true,
    dryRun: false,
    includeSecretMetadata: false,
    clusterName: "prod-east",
    environment: "production",
    customerId: "acme-hospitality",
    site: "us-east-1a",
    contextName: "prod-east-admin",
    kubeconfigPath: "~/.kube/config",
  });

  useEffect(() => {
    backend.getBootstrap().then((bootstrap) => applyTheme(bootstrap.theme));
    backend.getSettings().then((saved) => {
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
    });
    backend.listProjects().then(setProjects);
    const unsubscribe = backend.onScanEvent((event) => {
      setEvents((current) => [...current, event]);
      setStatusMessage(event.message);
      if (event.type === "complete") {
        setActiveRunId(null);
      }
    });
    return unsubscribe;
  }, []);

  const bundle = workspace?.bundle;
  const activePercent = events.length > 0 ? events[events.length - 1].percent || 0 : 0;

  async function handlePreflight() {
    setBusy(true);
    try {
      const report = await backend.runPreflight(scanForm);
      setPreflight(report);
      setStatusMessage(report.canRun ? "Preflight checks passed." : "Preflight found blocking issues.");
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
    } finally {
      setBusy(false);
      setActiveRunId(null);
    }
  }

  async function handleCancelRun() {
    if (!activeRunId) {
      return;
    }
    await backend.cancelRun(activeRunId);
    setStatusMessage("Cancel request sent.");
    setActiveRunId(null);
  }

  async function handleOpenBundle(path?: string) {
    const resolved = path || openBundlePath || (await backend.pickBundleFile());
    if (!resolved) {
      return;
    }
    setBusy(true);
    try {
      const loaded = await backend.openBundle(resolved);
      setWorkspace(loaded);
      setOpenBundlePath(resolved);
      setResultTab("Summary");
      startTransition(() => setView("results"));
      setStatusMessage("Existing bundle loaded into the workspace.");
    } finally {
      setBusy(false);
    }
  }

  async function handleBrowseOutput() {
    const outputDir = await backend.pickOutputDirectory();
    if (outputDir) {
      setScanForm((current) => ({ ...current, outputDir }));
    }
  }

  async function handleSaveSettings() {
    await backend.saveSettings(settings);
    setStatusMessage("Desktop settings saved.");
  }

  async function handleExport(kind: "report" | "summary" | "runbook" | "json" | "csv" | "redacted") {
    const bundlePath = workspace?.artifacts.loadedBundlePath || workspace?.artifacts.bundleJson;
    if (!bundlePath) {
      return;
    }
    const request: ExportRequest = {
      outputDir: workspace?.artifacts.outputDir || settings.defaultOutputDir || "./out",
      report: kind === "report",
      bundleJson: kind === "json",
      summary: kind === "summary",
      runbook: kind === "runbook",
      csvExport: kind === "csv",
      redact: kind === "redacted",
    };
    const artifacts = await backend.exportBundle(bundlePath, request);
    setExportNotice(exportMessage(kind, artifacts));
    setStatusMessage(`${titleCase(kind)} export refreshed in ${artifacts.outputDir}.`);
  }

  return (
    <div className="app-shell">
      <aside className="sidebar" aria-label="Primary navigation">
        <div className="brand">
          <div className="brand-mark">DR</div>
          <div>
            <p className="eyebrow">Desktop Shell</p>
            <h1>K8s Recovery Visualizer</h1>
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
          <p className="eyebrow">Current Posture</p>
          {bundle ? (
            <div className="hero-score-inline">
              <strong>{bundle.score.overall.final}</strong>
              <span className={`tone tone-${bundle.score.maturity.toLowerCase()}`}>{bundle.score.maturity}</span>
            </div>
          ) : browserDemo ? (
            <div className="hero-score-inline">
              <strong>85</strong>
              <span className="tone tone-gold">GOLD</span>
            </div>
          ) : (
            <div className="hero-score-inline">
              <strong>—</strong>
              <span className="chip">No bundle loaded</span>
            </div>
          )}
          <p className="muted">Reports and GUI now share the same dark, offline-friendly token system.</p>
        </div>
      </aside>

      <div className="workspace">
        <header className="topbar">
          <div>
            <p className="eyebrow">Status</p>
            <h2>{statusMessage}</h2>
          </div>
          <div className="topbar-actions">
            <span className="chip">{bundle?.metadata.clusterName || (browserDemo ? "Demo Workspace" : "No bundle loaded")}</span>
            <span className="chip">{bundle?.metadata.environment || (browserDemo ? "Production" : "Ready")}</span>
            <button type="button" className="button secondary" onClick={() => handleOpenBundle()}>
              Open Existing Bundle
            </button>
          </div>
        </header>

        <main className="main-content">
          {view === "home" && <HomeView workspace={workspace} projects={projects} onViewProjects={() => setView("projects")} onOpenBundle={handleOpenBundle} />}
          {view === "projects" && <ProjectsView projects={projects} onOpenBundle={handleOpenBundle} />}
          {view === "scan" && <ScanView busy={busy} wizardStep={wizardStep} setWizardStep={setWizardStep} scanForm={scanForm} setScanForm={setScanForm} preflight={preflight} onPreflight={handlePreflight} onStartScan={handleStartScan} onBrowseOutput={handleBrowseOutput} />}
          {view === "live" && <LiveView events={events} activeRunId={activeRunId} activePercent={activePercent} onCancel={handleCancelRun} />}
          {view === "results" && workspace && <ResultsView workspace={workspace} resultTab={resultTab} setResultTab={setResultTab} findingFilter={findingFilter} setFindingFilter={setFindingFilter} exportNotice={exportNotice} onExport={handleExport} />}
          {view === "settings" && <SettingsView settings={settings} setSettings={setSettings} openBundlePath={openBundlePath} setOpenBundlePath={setOpenBundlePath} onSave={handleSaveSettings} onOpenBundle={() => handleOpenBundle()} />}
        </main>
      </div>
    </div>
  );
}
