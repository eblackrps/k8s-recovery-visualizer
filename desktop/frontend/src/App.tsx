import { startTransition, useEffect, useRef, useState } from "react";
import { backend, mockWorkspace } from "./lib/backend";
import type {
  AppAlert,
  ConnectionAdvisor,
  ConnectionTestReport,
  ContextCatalog,
  ExportRequest,
  FailureDiagnosis,
  KubeconfigInspection,
  PreflightReport,
  ProjectSummary,
  RunEvent,
  RunCompletionSummary,
  ScanRequest,
  ScanStage,
  Settings,
  Workspace,
} from "./lib/types";
import type { ScanFieldName } from "./lib/scanForm";
import { diagnoseFailure as diagnoseUiFailure } from "./lib/diagnostics";
import { haveConnectionInputsChanged, haveContextInputsChanged, inspectConnectionSetup, inspectScanForm, prepareScanRequest, sanitizeScanForm, validateContextDiscovery } from "./lib/scanForm";
import { applyTheme, exportMessage, titleCase } from "./components/ui";
import { NavIcon, type NavIconName, navIcons } from "./components/NavIcons";
import { HomeView, ProjectsView } from "./views/HomeProjectsView";
import { ScanView } from "./views/ScanView";
import { LiveView } from "./views/LiveView";
import { CompletionView } from "./views/CompletionView";
import { ResultsView } from "./views/ResultsView";
import { SettingsView } from "./views/SettingsView";

type View = "home" | "projects" | "scan" | "live" | "complete" | "results" | "settings";
type BannerTone = AppAlert["tone"];

const navItems: Array<{ id: Exclude<View, "complete">; label: string }> = [
  { id: "home", label: "Home" },
  { id: "projects", label: "Projects" },
  { id: "scan", label: "New Scan" },
  { id: "live", label: "Live Run" },
  { id: "results", label: "Results" },
  { id: "settings", label: "Settings" },
];

const navIconNames: Partial<Record<Exclude<View, "complete">, NavIconName>> = {
  home: "home",
  projects: "projects",
  scan: "scan",
  live: "live",
  results: "results",
  settings: "settings",
};

function searchParams() {
  return new URLSearchParams(window.location.search);
}

function initialView(): View {
  const value = searchParams().get("view");
  if (value === "complete") {
    return "complete";
  }
  if (value && navItems.some((item) => item.id === value)) {
    return value as View;
  }
  return "home";
}

function initialScanStage(): ScanStage {
  switch (searchParams().get("scanStage")) {
    case "validate":
    case "outputs":
    case "launch":
      return searchParams().get("scanStage") as ScanStage;
    default:
      return "connect";
  }
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

function isFirstRunDemo() {
  return searchParams().get("firstRun") === "1";
}

function isCompletionDemo() {
  return searchParams().get("view") === "complete";
}

function initialLiveEvents(): RunEvent[] {
  if (searchParams().get("view") !== "live") {
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
  const requestedConnection = browserDemo ? searchParams().get("scanConnection") : "";
  const connectionMethod =
    requestedConnection === "kubeconfig_file" ||
    requestedConnection === "kubeconfig_inline" ||
    requestedConnection === "api_endpoint"
      ? requestedConnection
      : "current";

  return {
    connectionMethod,
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
    contextName: browserDemo && connectionMethod !== "api_endpoint" ? "prod-east-admin" : "",
    apiServerEndpoint: browserDemo && connectionMethod === "api_endpoint" ? "https://prod-east.example.net:6443" : "",
    caCertPath: browserDemo && connectionMethod === "api_endpoint" ? "./demo-certs/prod-east-ca.pem" : "",
  };
}

function buildInitialScanRequest(browserDemo: boolean, settings: Settings): ScanRequest {
  const base = defaultScanForm(browserDemo);
  return sanitizeScanForm({
    ...base,
    outputDir: settings.defaultOutputDir || base.outputDir,
    profileName: settings.defaultProfile || base.profileName,
    summary: settings.summary,
    runbook: settings.runbook,
    redact: settings.redact,
    csvExport: settings.csvExport,
    includeSecretMetadata: settings.includeSecretMetadata,
  });
}

function isEditableTarget(target: EventTarget | null) {
  return target instanceof HTMLInputElement ||
    target instanceof HTMLTextAreaElement ||
    target instanceof HTMLSelectElement ||
    (target instanceof HTMLElement && (
      target.isContentEditable ||
      target.closest("[contenteditable='true'], [contenteditable='']") !== null
    ));
}

function navButtonTitle(view: Exclude<View, "complete">, label: string) {
  switch (view) {
    case "home":
      return `${label} (Ctrl+H)`;
    case "scan":
      return `${label} (Ctrl+N)`;
    default:
      return label;
  }
}

const scanFieldOrder: ScanFieldName[] = [
  "apiServerEndpoint",
  "bearerToken",
  "caTrust",
  "insecureAcknowledgement",
  "contextName",
  "kubeconfigPath",
  "kubeconfigContent",
  "outputDir",
  "timeoutSeconds",
];

function mergeFieldFeedback(
  base: ReturnType<typeof inspectScanForm>,
  fieldErrors: Partial<Record<ScanFieldName, string>>,
  fieldWarnings: Partial<Record<ScanFieldName, string>>,
) {
  const mergedErrors = { ...base.fieldErrors };
  const mergedWarnings = { ...base.fieldWarnings };
  const errors = [...base.errors];

  for (const [field, message] of Object.entries(fieldErrors)) {
    const resolvedField = field as ScanFieldName;
    if (!message) {
      continue;
    }
    if (!mergedErrors[resolvedField]) {
      mergedErrors[resolvedField] = message;
      errors.push(message);
    }
  }

  for (const [field, message] of Object.entries(fieldWarnings)) {
    const resolvedField = field as ScanFieldName;
    if (!message || mergedWarnings[resolvedField]) {
      continue;
    }
    mergedWarnings[resolvedField] = message;
  }

  return {
    ...base,
    errors,
    fieldErrors: mergedErrors,
    fieldWarnings: mergedWarnings,
    firstInvalidField: scanFieldOrder.find((field) => Boolean(mergedErrors[field])) || base.firstInvalidField,
  };
}

function normalizeFieldMap(source?: Record<string, string>) {
  const next: Partial<Record<ScanFieldName, string>> = {};
  if (!source) {
    return next;
  }
  for (const [field, message] of Object.entries(source)) {
    next[field as ScanFieldName] = message;
  }
  return next;
}

function buildRunCompletionSummary(runId: string, workspace: Workspace): RunCompletionSummary {
  return {
    runId,
    clusterName: workspace.bundle.metadata.clusterName,
    environment: workspace.bundle.metadata.environment,
    generatedAt: workspace.bundle.metadata.generatedAt,
    score: workspace.bundle.score.overall.final,
    findingCount: workspace.bundle.inventory.findings?.length || 0,
    hasComparison: Boolean(workspace.bundle.comparison),
    artifacts: workspace.artifacts,
  };
}

function demoCompletionSummary(browserDemo: boolean, firstRunDemo: boolean) {
  if (!browserDemo || firstRunDemo || !isCompletionDemo()) {
    return null;
  }
  return buildRunCompletionSummary("demo-run", mockWorkspace);
}

function describeView(view: View) {
  if (view === "complete") {
    return "Scan Complete";
  }
  return navItems.find((item) => item.id === view)?.label || "Workspace";
}

export default function App() {
  const browserDemo = isBrowserDemo();
  const firstRunDemo = browserDemo && isFirstRunDemo();
  const [view, setView] = useState<View>(initialView);
  const [scanStage, setScanStage] = useState<ScanStage>(initialScanStage);
  const [resultTab, setResultTab] = useState(normalizeResultTab(searchParams().get("tab")));
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
  const [workspace, setWorkspace] = useState<Workspace | null>(browserDemo && !firstRunDemo ? mockWorkspace : null);
  const [preflight, setPreflight] = useState<PreflightReport | null>(null);
  const [connectionAdvisor, setConnectionAdvisor] = useState<ConnectionAdvisor | null>(null);
  const [kubeconfigInspection, setKubeconfigInspection] = useState<KubeconfigInspection | null>(null);
  const [connectionTest, setConnectionTest] = useState<ConnectionTestReport | null>(null);
  const [contextCatalog, setContextCatalog] = useState<ContextCatalog | null>(null);
  const [events, setEvents] = useState<RunEvent[]>(initialLiveEvents());
  const [activeRunId, setActiveRunId] = useState<string | null>(
    searchParams().get("view") === "live" ? "demo-live" : null,
  );
  const [busy, setBusy] = useState(false);
  const [detectingContexts, setDetectingContexts] = useState(false);
  const [insecureAcknowledged, setInsecureAcknowledged] = useState(false);
  const [scanValidationRequest, setScanValidationRequest] = useState<{ version: number; field?: ScanFieldName }>({ version: 0 });
  const [fieldErrors, setFieldErrors] = useState<Partial<Record<ScanFieldName, string>>>({});
  const [fieldWarnings, setFieldWarnings] = useState<Partial<Record<ScanFieldName, string>>>({});
  const [statusMessage, setStatusMessage] = useState(
    view === "scan"
      ? "Choose a connection path and validate it before the scan."
      : view === "complete"
        ? "Scan complete. Review outputs or move into findings."
        : firstRunDemo
          ? "Connect to a cluster or open a saved bundle."
          : "Start a new assessment or open a saved bundle.",
  );
  const [actionBanner, setActionBanner] = useState<{ tone: BannerTone; message: string } | null>(null);
  const [openBundlePath, setOpenBundlePath] = useState(defaultBundleInputPath(browserDemo));
  const [loadedKubeconfigLabel, setLoadedKubeconfigLabel] = useState("");
  const [recentCompletion, setRecentCompletion] = useState<RunCompletionSummary | null>(() =>
    demoCompletionSummary(browserDemo, firstRunDemo),
  );
  const [runFailure, setRunFailure] = useState<FailureDiagnosis | null>(null);
  const [exportNotice, setExportNotice] = useState("");
  const [findingFilter, setFindingFilter] = useState("ALL");
  const [scanFormTouched, setScanFormTouched] = useState(false);
  const [scanForm, setScanForm] = useState<ScanRequest>(defaultScanForm(browserDemo));
  const connectionValidation = mergeFieldFeedback(inspectConnectionSetup(scanForm, { insecureAcknowledged }), fieldErrors, fieldWarnings);
  const scanValidation = mergeFieldFeedback(inspectScanForm(scanForm, { insecureAcknowledged }), fieldErrors, fieldWarnings);
  const canResetScanForm =
    scanFormTouched || Boolean(connectionTest) || Boolean(preflight) || scanStage !== "connect";
  const hasMeaningfulScanState = scanStage !== "connect" && (
    Boolean(connectionTest) ||
    Boolean(preflight) ||
    scanFormTouched
  );

  function updateScanForm(updater: (current: ScanRequest) => ScanRequest) {
    const proposed = updater(scanForm);
    const next = sanitizeScanForm({
      ...proposed,
      ...(proposed.connectionMethod === "kubeconfig_file" &&
      proposed.connectionMethod !== scanForm.connectionMethod &&
      !proposed.kubeconfigPath &&
      connectionAdvisor?.defaultKubeconfigAvailable &&
      connectionAdvisor.defaultKubeconfigPath
        ? { kubeconfigPath: connectionAdvisor.defaultKubeconfigPath }
        : {}),
    });
    setPreflight(null);
    if (haveConnectionInputsChanged(scanForm, next)) {
      setConnectionTest(null);
      setFieldErrors({});
      setFieldWarnings({});
      setKubeconfigInspection(null);
      if (scanStage === "outputs" || scanStage === "launch") {
        setScanStage("validate");
      }
      if (next.connectionMethod !== scanForm.connectionMethod) {
        setScanStage("connect");
      }
    }
    if (haveContextInputsChanged(scanForm, next)) {
      setContextCatalog(null);
    }
    if (next.connectionMethod !== "kubeconfig_inline" || !next.kubeconfigContent) {
      setLoadedKubeconfigLabel("");
    }
    if (next.connectionMethod !== "api_endpoint" || !next.insecure) {
      setInsecureAcknowledged(false);
    }
    setRunFailure(null);
    setScanFormTouched(true);
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
          setProjects(firstRunDemo ? [] : nextProjects);
        }
      } catch (error) {
        if (active) {
          showBanner("error", `Workspace discovery failed: ${formatActionError(error, "Could not load saved projects.")}`);
          setStatusMessage("Workspace discovery failed.");
        }
      }

      try {
        const advisor = await backend.getConnectionAdvisor();
        if (active) {
          setConnectionAdvisor(advisor);
          setScanForm((current) => {
            const isUntouchedConnection =
              (current.connectionMethod || "current") === "current" &&
              !current.kubeconfigPath &&
              !current.kubeconfigContent &&
              !current.apiServerEndpoint;
            if (!isUntouchedConnection || !advisor.recommendedMethod || advisor.recommendedMethod === "current") {
              return current;
            }
            return sanitizeScanForm({
              ...current,
              connectionMethod: advisor.recommendedMethod,
              kubeconfigPath:
                advisor.recommendedMethod === "kubeconfig_file" && advisor.defaultKubeconfigPath
                  ? advisor.defaultKubeconfigPath
                  : current.kubeconfigPath,
            });
          });
        }
      } catch {
        // Connection advice is optional and should not block startup.
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
  const currentViewLabel = describeView(view);
  const isScanView = view === "scan";
  const environmentLabel = bundle?.metadata.environment || (browserDemo ? "production" : "ready");
  const clusterLabel = bundle?.metadata.clusterName || (browserDemo ? "demo-workspace" : "no bundle");

  function showBanner(tone: BannerTone, message: string) {
    setActionBanner({ tone, message });
  }

  function clearBanner() {
    setActionBanner(null);
  }

  function applyFieldFeedback(
    errors?: Partial<Record<ScanFieldName, string>>,
    warnings?: Partial<Record<ScanFieldName, string>>,
  ) {
    setFieldErrors(errors || {});
    setFieldWarnings(warnings || {});
  }

  async function inspectSelectedKubeconfig(request: ScanRequest) {
    const inspection = await backend.inspectKubeconfig(request);
    setKubeconfigInspection(inspection);
    applyFieldFeedback({}, {});
    return inspection;
  }

  async function handleInspectKubeconfig(request: ScanRequest) {
    clearBanner();
    setBusy(true);
    try {
      const inspection = await inspectSelectedKubeconfig(request);
      setStatusMessage(inspection.summary || "Kubeconfig looks valid.");
    } catch (error) {
      const message = formatActionError(error, "Could not inspect the kubeconfig.");
      const targetField = request.connectionMethod === "kubeconfig_inline" ? "kubeconfigContent" : "kubeconfigPath";
      setKubeconfigInspection(null);
      applyFieldFeedback({ [targetField]: message }, {});
      setScanValidationRequest((current) => ({ version: current.version + 1, field: targetField }));
      showBanner("error", message);
      setStatusMessage(`Kubeconfig validation failed: ${message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleUseDetectedKubeconfig() {
    if (!connectionAdvisor?.defaultKubeconfigPath) {
      return;
    }
    const nextRequest = sanitizeScanForm({
      ...scanForm,
      connectionMethod: "kubeconfig_file",
      kubeconfigPath: connectionAdvisor.defaultKubeconfigPath,
    });
    setScanFormTouched(true);
    setScanForm(nextRequest);
    setScanStage("validate");
    await handleInspectKubeconfig(nextRequest);
  }

  async function handleTestConnection() {
    const request = prepareScanRequest(scanForm);
    const validation = inspectConnectionSetup(request, { insecureAcknowledged });
    if (validation.errors.length > 0) {
      setScanValidationRequest((current) => ({ version: current.version + 1, field: validation.firstInvalidField }));
      showBanner("info", validation.errors[0]);
      setStatusMessage("Connection details need attention before testing.");
      return;
    }

    clearBanner();
    applyFieldFeedback({}, {});
    setBusy(true);
    try {
      if (request.connectionMethod === "kubeconfig_file" || request.connectionMethod === "kubeconfig_inline") {
        try {
          await inspectSelectedKubeconfig(request);
        } catch (error) {
          const message = formatActionError(error, "Could not inspect the kubeconfig.");
          const targetField = request.connectionMethod === "kubeconfig_inline" ? "kubeconfigContent" : "kubeconfigPath";
          applyFieldFeedback({ [targetField]: message }, {});
          setScanValidationRequest((current) => ({ version: current.version + 1, field: targetField }));
          setStatusMessage(`Connection test failed: ${message}`);
          showBanner("error", message);
          setConnectionTest(null);
          return;
        }
      }

      const report = await backend.testConnection(request);
      setConnectionTest(report);
      applyFieldFeedback(normalizeFieldMap(report.fieldErrors), normalizeFieldMap(report.fieldWarnings));
      if (report.canConnect) {
        showBanner("info", report.summary || "Connection test succeeded.");
        setStatusMessage(report.summary || "Connection test succeeded.");
        if (scanStage === "validate") {
          setScanStage("outputs");
        }
      } else {
        const firstField = scanFieldOrder.find((field) => Boolean(report.fieldErrors?.[field]));
        if (firstField) {
          setScanValidationRequest((current) => ({ version: current.version + 1, field: firstField }));
        }
        showBanner("error", report.summary || "Connection test failed.");
        setStatusMessage(report.summary || "Connection test failed.");
      }
    } catch (error) {
      const message = formatActionError(error, "Could not test the connection.");
      showBanner("error", `Connection test failed: ${message}`);
      setStatusMessage(`Connection test failed: ${message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handlePreflight() {
    const request = prepareScanRequest(scanForm);
    const validation = scanValidation;
    if (validation.errors.length > 0) {
      setScanValidationRequest((current) => ({ version: current.version + 1, field: validation.firstInvalidField }));
      showBanner("info", validation.errors[0]);
      setStatusMessage("Scan setup needs attention before preflight.");
      return;
    }
    clearBanner();
    setBusy(true);
    try {
      setScanForm(request);
      const report = await backend.runPreflight(request);
      setPreflight(report);
      setScanStage("launch");
      setStatusMessage(report.canRun ? "Preflight checks passed." : "Preflight found blocking issues.");
    } catch (error) {
      const message = formatActionError(error, "Could not run preflight.");
      showBanner("error", `Preflight failed: ${message}`);
      setStatusMessage(`Preflight failed: ${message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleStartScan() {
    const request = prepareScanRequest(scanForm);
    const validation = scanValidation;
    if (validation.errors.length > 0) {
      setScanValidationRequest((current) => ({ version: current.version + 1, field: validation.firstInvalidField }));
      showBanner("info", validation.errors[0]);
      setStatusMessage("Scan setup needs attention before the run can start.");
      return;
    }
    const runId = `run-${Date.now()}`;
    setBusy(true);
    setEvents([]);
    setActiveRunId(runId);
    setExportNotice("");
    setRunFailure(null);
    clearBanner();
    startTransition(() => setView("live"));
    try {
      setScanForm(request);
      const result = await backend.runScan({ ...request, runId });
      setPreflight(result.preflight);
      setWorkspace(result.workspace);
      setRecentCompletion(buildRunCompletionSummary(runId, result.workspace));
      setRunFailure(null);
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
      startTransition(() => setView("complete"));
      setStatusMessage(`Scan finished. Bundle and report outputs were written to ${result.artifacts.outputDir}.`);
    } catch (error) {
      if (isCanceledError(error)) {
        setRunFailure(null);
        showBanner("info", "Scan canceled.");
        setStatusMessage("Scan canceled.");
      } else {
        const message = formatActionError(error, "Could not complete the scan.");
        const diagnosis = diagnoseUiFailure(request, message);
        setRunFailure(diagnosis);
        showBanner("error", `Scan failed: ${diagnosis.summary || message}`);
        setStatusMessage(`Scan failed: ${diagnosis.summary || message}`);
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
      setRecentCompletion(null);
      setRunFailure(null);
      setOpenBundlePath(resolved);
      setResultTab("Overview");
      startTransition(() => setView("results"));
      setStatusMessage(`Loaded bundle from ${resolved}.`);
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
        const nextRequest = sanitizeScanForm({ ...scanForm, connectionMethod: "kubeconfig_file", kubeconfigPath });
        setScanForm(nextRequest);
        setScanStage("validate");
        await handleInspectKubeconfig(nextRequest);
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
      if (request.connectionMethod === "kubeconfig_file" || request.connectionMethod === "kubeconfig_inline") {
        await inspectSelectedKubeconfig(request);
      }
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
      const field = request.connectionMethod === "kubeconfig_inline" ? "kubeconfigContent" : request.connectionMethod === "kubeconfig_file" ? "kubeconfigPath" : undefined;
      if (field) {
        applyFieldFeedback({ [field]: message }, {});
      }
      showBanner("error", `Context discovery failed: ${message}`);
      setStatusMessage("Context discovery failed.");
    } finally {
      setDetectingContexts(false);
    }
  }

  async function handleOpenPath(path: string, label: string) {
    const target = path.trim();
    if (!target) {
      showBanner("info", `No ${label} is available yet.`);
      setStatusMessage(`No ${label} is available yet.`);
      return;
    }
    clearBanner();
    try {
      await backend.openPath(target);
      setStatusMessage(`Opened ${label}.`);
    } catch (error) {
      const message = formatActionError(error, `Could not open the ${label}.`);
      showBanner("error", message);
      setStatusMessage(`Open ${label} failed.`);
    }
  }

  async function handleLoadDroppedKubeconfig(fileName: string, content: string) {
    const label = fileName.trim() || "kubeconfig file";
    clearBanner();
    const nextRequest = sanitizeScanForm({
      ...scanForm,
      connectionMethod: "kubeconfig_inline",
      kubeconfigPath: "",
      kubeconfigContent: content,
      contextName: "",
    });
    setLoadedKubeconfigLabel(label);
    setConnectionTest(null);
    setPreflight(null);
    setContextCatalog(null);
    applyFieldFeedback({}, {});
    setScanFormTouched(true);
    setScanForm(nextRequest);
    setScanStage("validate");
    setBusy(true);
    try {
      const inspection = await inspectSelectedKubeconfig(nextRequest);
      const dependencyNote = inspection.referencedFiles?.length
        ? " Review the inspection note for local CA or client-certificate dependencies."
        : "";
      showBanner("info", `Loaded ${label} into paste kubeconfig mode.${dependencyNote}`);
      setStatusMessage(`Loaded ${label} into paste kubeconfig mode.`);
    } catch (error) {
      const message = formatActionError(error, "Could not inspect the dropped kubeconfig.");
      setLoadedKubeconfigLabel("");
      setKubeconfigInspection(null);
      applyFieldFeedback({ kubeconfigContent: message }, {});
      setScanValidationRequest((current) => ({ version: current.version + 1, field: "kubeconfigContent" }));
      showBanner("error", message);
      setStatusMessage(`Dropped kubeconfig validation failed: ${message}`);
    } finally {
      setBusy(false);
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

  function handleOpenScan(stage: ScanStage = "connect") {
    clearBanner();
    setConnectionTest(null);
    setPreflight(null);
    setFieldErrors({});
    setFieldWarnings({});
    setScanFormTouched(false);
    setRecentCompletion(null);
    setRunFailure(null);
    setScanStage(stage);
    setStatusMessage("Choose a connection path and validate it before the scan.");
    setView("scan");
  }

  function handleResetScanForm() {
    clearBanner();
    setConnectionTest(null);
    setPreflight(null);
    setKubeconfigInspection(null);
    setContextCatalog(null);
    setFieldErrors({});
    setFieldWarnings({});
    setLoadedKubeconfigLabel("");
    setInsecureAcknowledged(false);
    setRunFailure(null);
    setScanValidationRequest({ version: 0 });
    setScanFormTouched(false);
    setScanForm(buildInitialScanRequest(browserDemo, settings));
    setScanStage("connect");
    setStatusMessage("Choose a connection path and validate it before the scan.");
  }

  function confirmLeavingScanSetup(nextView: View) {
    if (view !== "scan" || nextView === "scan" || !hasMeaningfulScanState) {
      return true;
    }
    return window.confirm("Leave scan setup? Your connection test and preflight results will be kept, but form changes may be lost.");
  }

  function handleSidebarNavigation(nextView: Exclude<View, "complete">) {
    if (nextView === "scan") {
      handleOpenScan("connect");
      return;
    }
    if (nextView === view) {
      return;
    }
    if (!confirmLeavingScanSetup(nextView)) {
      return;
    }
    setView(nextView);
  }

  const handlersRef = useRef({
    handleOpenScan,
    handlePickBundle,
    handleSidebarNavigation,
  });

  useEffect(() => {
    handlersRef.current = {
      handleOpenScan,
      handlePickBundle,
      handleSidebarNavigation,
    };
  });

  useEffect(() => {
    const handleKeydown = (event: KeyboardEvent) => {
      if (busy || !event.ctrlKey || event.altKey || event.metaKey || event.shiftKey) {
        return;
      }
      if (isEditableTarget(event.target)) {
        return;
      }
      const { handleOpenScan: openScan, handlePickBundle: pickBundle, handleSidebarNavigation: navigateSidebar } =
        handlersRef.current;
      switch (event.key.toLowerCase()) {
        case "n":
          event.preventDefault();
          openScan("connect");
          break;
        case "o":
          event.preventDefault();
          void pickBundle();
          break;
        case "h":
          event.preventDefault();
          navigateSidebar("home");
          break;
        default:
          break;
      }
    };
    window.addEventListener("keydown", handleKeydown);
    return () => window.removeEventListener("keydown", handleKeydown);
  }, [busy]);

  const topbarActions = view === "results" && bundle ? (
    <>
      <button type="button" className="button primary" onClick={() => handleOpenScan("connect")}>
        Run new scan
      </button>
      <button type="button" className="button secondary quiet" onClick={() => setView("settings")}>
        Settings
      </button>
    </>
  ) : null;

  return (
    <div className={`app-shell${isScanView ? " scan-shell" : ""}`}>
      <aside className="sidebar" aria-label="Primary navigation">
        <div className="brand">
          <div className="brand-mark" aria-hidden="true">K8V</div>
          <div className="brand-copy">
            <h1>K8V</h1>
            <p className="muted">Kubernetes recovery visualizer</p>
          </div>
        </div>
        <nav className="nav">
          {navItems.map((item) => {
            const iconName = navIconNames[item.id];
            const hasIcon = iconName ? Boolean(navIcons[iconName]) : false;
            return (
              <button
                key={item.id}
                type="button"
                className={`nav-link ${view === item.id || (view === "complete" && item.id === "results") ? "is-active" : ""}`}
                onClick={() => handleSidebarNavigation(item.id)}
                aria-current={view === item.id || (view === "complete" && item.id === "results") ? "page" : undefined}
                title={navButtonTitle(item.id, item.label)}
              >
                {hasIcon && iconName ? (
                  <span className="nav-icon" aria-hidden="true">
                    <NavIcon name={iconName} />
                  </span>
                ) : null}
                <span className="nav-label">{item.label}</span>
              </button>
            );
          })}
        </nav>
        <div className="sidebar-footer">
          <p className="eyebrow">Workspace</p>
          <p className="muted">Run remote assessments, inspect saved bundles, and review restore readiness without leaving the desktop console.</p>
        </div>
      </aside>

      <div className={`workspace${isScanView ? " scan-workspace" : ""}`}>
        <header className="topbar">
          <div className="status-stack">
            <p className="eyebrow">{currentViewLabel}</p>
            <h2>{statusMessage}</h2>
          </div>
          <div className="topbar-actions">
            {bundle ? (
              <div className="context-strip" aria-label="Active bundle context">
                <div className="context-cluster">
                  <strong>{clusterLabel}</strong>
                  <span className="muted">{environmentLabel}</span>
                </div>
              </div>
            ) : (
              <div className="context-strip empty" aria-label="Bundle status">
                <div className="context-cluster">
                  <strong>No bundle loaded yet</strong>
                </div>
                <p className="muted">Run a scan to create a portable bundle, or open an existing bundle for offline review.</p>
              </div>
            )}
            {topbarActions}
            {topbarActions ? null : (
              <button type="button" className="button secondary quiet" onClick={handlePickBundle} disabled={busy} title="Open existing bundle (Ctrl+O)">
                Open Existing Bundle
              </button>
            )}
          </div>
        </header>
        <main className={`main-content${isScanView ? " scan-view-content" : ""}`}>
          {actionBanner ? <p className={`notice notice-${actionBanner.tone} action-banner`}>{actionBanner.message}</p> : null}
          {view === "home" && (
            <HomeView
              workspace={workspace}
              projects={projects}
              connectionAdvisor={connectionAdvisor}
              busy={busy}
              onViewProjects={() => setView("projects")}
              onPickBundle={handlePickBundle}
              onOpenProject={handleOpenBundlePath}
              onStartScan={() => handleOpenScan("connect")}
              onReviewFindings={() => {
                setResultTab("Findings");
                setView("results");
              }}
            />
          )}
          {view === "projects" && <ProjectsView projects={projects} busy={busy} onPickBundle={handlePickBundle} />}
          {view === "scan" && (
            <ScanView
              busy={busy}
              scanForm={scanForm}
              setScanForm={updateScanForm}
              scanStage={scanStage}
              onSetScanStage={setScanStage}
              connectionAdvisor={connectionAdvisor}
              kubeconfigInspection={kubeconfigInspection}
              connectionTest={connectionTest}
              preflight={preflight}
              contextCatalog={contextCatalog}
              detectingContexts={detectingContexts}
              connectionValidation={connectionValidation}
              validation={scanValidation}
              showValidationErrors={scanValidationRequest.version > 0}
              validationRequest={scanValidationRequest}
              insecureAcknowledged={insecureAcknowledged}
              onSetInsecureAcknowledged={setInsecureAcknowledged}
              onTestConnection={handleTestConnection}
              onInspectKubeconfig={() => handleInspectKubeconfig(prepareScanRequest(scanForm))}
              onUseDetectedKubeconfig={handleUseDetectedKubeconfig}
              onPreflight={handlePreflight}
              onStartScan={handleStartScan}
              onDetectContexts={handleDetectContexts}
              onBrowseOutput={handleBrowseOutput}
              onBrowseKubeconfig={handleBrowseKubeconfig}
              onBrowseCACert={handleBrowseCACert}
              canReset={canResetScanForm}
              onResetScanForm={handleResetScanForm}
              onLoadDroppedKubeconfig={handleLoadDroppedKubeconfig}
              loadedKubeconfigLabel={loadedKubeconfigLabel}
            />
          )}
          {view === "live" && (
            <LiveView
              events={events}
              activeRunId={activeRunId}
              activePercent={activePercent}
              outputDir={scanForm.outputDir || settings.defaultOutputDir}
              completionSummary={recentCompletion}
              runFailure={runFailure}
              onCancel={handleCancelRun}
              onViewResults={() => setView("results")}
              onFixScan={() =>
                handleOpenScan(
                  runFailure?.code === "output_path" || runFailure?.code === "artifact_write" ? "outputs" : "validate",
                )
              }
              onOpenPath={handleOpenPath}
            />
          )}
          {view === "complete" && workspace && recentCompletion && (
            <CompletionView
              workspace={workspace}
              summary={recentCompletion}
              onOpenPath={handleOpenPath}
              onReviewResults={() => setView("results")}
              onReviewFindings={() => {
                setResultTab("Findings");
                setView("results");
              }}
              onReviewCompare={
                workspace.bundle.comparison
                  ? () => {
                      setResultTab("Compare");
                      setView("results");
                    }
                  : undefined
              }
              onStartAnotherScan={() => handleOpenScan("connect")}
            />
          )}
          {view === "results" && workspace && (
            <ResultsView
              workspace={workspace}
              resultTab={resultTab}
              setResultTab={setResultTab}
              findingFilter={findingFilter}
              setFindingFilter={setFindingFilter}
              exportNotice={exportNotice}
              completionSummary={recentCompletion}
              onExport={handleExport}
              onOpenPath={handleOpenPath}
              onDismissCompletion={() => setRecentCompletion(null)}
            />
          )}
          {view === "settings" && <SettingsView settings={settings} busy={busy} setSettings={setSettings} openBundlePath={openBundlePath} setOpenBundlePath={setOpenBundlePath} onSave={handleSaveSettings} onOpenBundle={() => handleOpenBundlePath(openBundlePath)} />}
        </main>
      </div>
    </div>
  );
}
