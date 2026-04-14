import { mockBackend } from "./mock";
import type {
  AppAlert,
  ArtifactPaths,
  Bootstrap,
  ConnectionAdvisor,
  ContextCatalog,
  ConnectionTestReport,
  ExportRequest,
  KubeconfigInspection,
  PreflightReport,
  ProjectSummary,
  RunEvent,
  RunResult,
  ScanRequest,
  Settings,
  Workspace,
} from "./types";

declare global {
  interface Window {
    go?: {
      main?: {
        App?: Record<string, (...args: any[]) => Promise<any>>;
      };
    };
    runtime?: {
      EventsOn?: (eventName: string, callback: (payload: RunEvent) => void) => void;
      EventsOff?: (eventName: string) => void;
    };
  }
}

const isWails = () =>
  typeof window !== "undefined" &&
  Boolean(window.go?.main?.App) &&
  Boolean(window.runtime?.EventsOn);

async function invoke<T>(method: string, ...args: unknown[]): Promise<T> {
  if (isWails()) {
    const target = window.go?.main?.App?.[method];
    if (!target) {
      throw new Error(`Backend method ${method} is unavailable.`);
    }
    return (await target(...args)) as T;
  }
  const target = (mockBackend as unknown as Record<string, (...inner: unknown[]) => Promise<unknown>>)[method];
  return (await target(...args)) as T;
}

export const backend = {
  getBootstrap: () => invoke<Bootstrap>("GetBootstrap"),
  getSettings: () => invoke<Settings>("GetSettings"),
  getStartupAlerts: () => invoke<AppAlert[]>("GetStartupAlerts"),
  saveSettings: (settings: Settings) => invoke<void>("SaveSettings", settings),
  listProjects: (root?: string) => invoke<ProjectSummary[]>("ListProjects", root ?? ""),
  getConnectionAdvisor: () => invoke<ConnectionAdvisor>("GetConnectionAdvisor"),
  inspectKubeconfig: (request: ScanRequest) => invoke<KubeconfigInspection>("InspectKubeconfig", request),
  listConnectionContexts: (request: ScanRequest) => invoke<ContextCatalog>("ListConnectionContexts", request),
  testConnection: (request: ScanRequest) => invoke<ConnectionTestReport>("TestConnection", request),
  runPreflight: (request: ScanRequest) => invoke<PreflightReport>("RunPreflight", request),
  runScan: (request: ScanRequest) => invoke<RunResult>("RunScan", request),
  cancelRun: (runId: string) => invoke<void>("CancelRun", runId),
  openBundle: (path: string) => invoke<Workspace>("OpenBundle", path),
  exportBundle: (path: string, request: ExportRequest) =>
    invoke<ArtifactPaths>("ExportBundle", path, request),
  openPath: (path: string) => invoke<void>("OpenPath", path),
  pickBundleFile: () => invoke<string>("PickBundleFile"),
  pickKubeconfigFile: () => invoke<string>("PickKubeconfigFile"),
  pickCertificateFile: () => invoke<string>("PickCertificateFile"),
  pickOutputDirectory: () => invoke<string>("PickOutputDirectory"),
  onScanEvent(listener: (event: RunEvent) => void) {
    if (isWails()) {
      window.runtime?.EventsOn?.("scan:event", listener);
      return () => window.runtime?.EventsOff?.("scan:event");
    }
    return mockBackend.onScanEvent(listener);
  },
};

export const mockWorkspace = mockBackend.demoWorkspace;
