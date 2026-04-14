export type ThemeTokens = {
  palette: {
    background: string;
    surface: string;
    border: string;
    text: string;
    muted: string;
    accent: string;
    success: string;
    critical: string;
    high: string;
    medium: string;
  };
  maturity: {
    platinum: string;
    gold: string;
    silver: string;
    bronze: string;
  };
  typography: {
    body: string;
    title: string;
    mono: string;
  };
  radius: {
    xl: string;
    lg: string;
    md: string;
    sm: string;
  };
};

export type Bootstrap = {
  theme: ThemeTokens;
};

export type AppAlert = {
  tone: "info" | "error";
  message: string;
};

export type ProjectSummary = {
  name: string;
  clusterName?: string;
  environment?: string;
  outputDir: string;
  lastScanPath: string;
  reportPath?: string;
  score: number;
  maturity: string;
  timestampUtc?: string;
};

export type PreflightCheck = {
  id: string;
  title: string;
  status: "pass" | "warn" | "fail";
  required: boolean;
  scope?: string;
  resource?: string;
  detail: string;
  hint?: string;
  manifest?: string;
  commands?: string[];
};

export type PreflightReport = {
  canRun: boolean;
  degraded: boolean;
  server?: string;
  contextName?: string;
  scope: string;
  diagnosis?: FailureDiagnosis;
  checks: PreflightCheck[];
  warnings?: string[];
};

export type ContextCatalog = {
  contexts?: string[];
  currentContext?: string;
  source?: string;
};

export type ConnectionAdvisor = {
  recommendedMethod?: ConnectionMethod;
  recommendedReason?: string;
  kubectlAvailable?: boolean;
  kubectlPath?: string;
  currentLoginAvailable?: boolean;
  currentContext?: string;
  currentLoginDetail?: string;
  currentLoginWarning?: string;
  defaultKubeconfigAvailable?: boolean;
  defaultKubeconfigPath?: string;
  defaultKubeconfigCurrentContext?: string;
  defaultKubeconfigDetail?: string;
  defaultKubeconfigPortable?: boolean;
  defaultKubeconfigWarning?: string;
};

export type KubeconfigInspection = {
  source?: string;
  path?: string;
  currentContext?: string;
  contexts?: string[];
  clusterCount: number;
  userCount: number;
  usesExecAuth?: boolean;
  usesClientCertificate?: boolean;
  usesCertificateAuthorityFile?: boolean;
  usesCertificateAuthorityData?: boolean;
  referencedFiles?: string[];
  missingReferencedFiles?: string[];
  summary?: string;
  nextAction?: string;
};

export type ConnectionTestCheck = {
  id: string;
  title: string;
  status: "pass" | "warn" | "fail";
  detail: string;
  hint?: string;
};

export type FailureDiagnosis = {
  code?: string;
  label?: string;
  summary?: string;
  detail?: string;
  nextAction?: string;
};

export type ConnectionTestReport = {
  canConnect: boolean;
  source?: string;
  server?: string;
  contextName?: string;
  summary?: string;
  nextAction?: string;
  diagnosis?: FailureDiagnosis;
  fieldErrors?: Record<string, string>;
  fieldWarnings?: Record<string, string>;
  checks?: ConnectionTestCheck[];
};

export type ArtifactPaths = {
  outputDir: string;
  bundleJson?: string;
  enrichedJson?: string;
  htmlReport?: string;
  markdownReport?: string;
  summaryHtml?: string;
  runbookHtml?: string;
  redactedJson?: string;
  redactedHtml?: string;
  csvDir?: string;
  historyIndex?: string;
  loadedBundlePath?: string;
};

export type HistoryEntry = {
  timestampUtc: string;
  overall: number;
  storage?: number;
  workload?: number;
  config?: number;
  backup?: number;
  findings?: number;
  maturity: string;
  clusterName?: string;
  environment?: string;
  jsonPath?: string;
  htmlPath?: string;
};

export type HistoryDashboard = {
  entries: HistoryEntry[];
  trendLabel?: string;
  trendDelta?: number;
  averageScore?: number;
  bestScore?: number;
  worstScore?: number;
  runCount?: number;
  domainTrends?: Array<{ name: string; current: number; delta: number; direction: string }>;
};

export type Finding = {
  id?: string;
  title?: string;
  severity: string;
  resourceId: string;
  message: string;
  recommendation: string;
  impact?: string;
  effort?: string;
  ownerHint?: string;
  priorityScore?: number;
  rank?: number;
};

export type RemediationStep = {
  priority: number;
  category: string;
  title: string;
  detail: string;
  ownerHint?: string;
  effort?: string;
  whyItMatters?: string;
  drImpact?: string;
  validation?: string[];
  fixSteps?: string[];
  commands?: string[];
  caveats?: string[];
};

export type BackupPolicy = {
  tool: string;
  name: string;
  includedNamespaces?: string[];
  schedule?: string;
  rpoHours?: number;
  lastSuccessAt?: string;
  confidence?: string;
  hasOffsite?: boolean;
  retentionTtl?: string;
};

export type RestoreNamespace = {
  namespace: string;
  coverageKnown: boolean;
  hasCoverage: boolean;
  rpoHours: number;
  pvcSizeGb: number;
  readiness?: string;
  blockers?: string[];
  warnings?: string[];
};

export type Bundle = {
  metadata: {
    customerId?: string;
    site?: string;
    clusterName?: string;
    environment?: string;
    generatedAt?: string;
    toolVersion?: string;
  };
  tool: {
    name?: string;
    version?: string;
    buildDate?: string;
  };
  cluster: {
    apiServer?: {
      endpoint?: string;
    };
    platform?: {
      provider?: string;
      k8sVersion?: string;
      clusterUID?: string;
    };
  };
  scan: {
    scanId: string;
    startedAt?: string;
    endedAt?: string;
    durationSeconds?: number;
  };
  schemaVersion?: string;
  target?: string;
  profile?: string;
  scanNamespaces?: string[];
  collectorSkips?: Array<{ name: string; reason: string; rbac: boolean }>;
  trendHistory?: Array<{
    ts: string;
    score: number;
    storage?: number;
    workload?: number;
    config?: number;
    backup?: number;
    findings?: number;
    maturity: string;
  }>;
  comparison?: {
    previousScannedAt?: string;
    previousScore?: number;
    previousMaturity?: string;
    currentScore?: number;
    currentMaturity?: string;
    scoreDelta?: number;
    namespacesAdded?: string[];
    namespacesRemoved?: string[];
    workloadsAdded?: string[];
    workloadsRemoved?: string[];
    pvcsAdded?: string[];
    pvcsRemoved?: string[];
    imagesAdded?: string[];
    imagesRemoved?: string[];
    findingsNew?: Finding[];
    findingsResolved?: Finding[];
    domainDeltas?: Array<{ name: string; previous: number; current: number; delta: number }>;
    severityDeltas?: Array<{ severity: string; previous: number; current: number; delta: number }>;
    inventoryDeltas?: Array<{ name: string; added: number; removed: number }>;
    findingsRegressed?: Array<{
      id: string;
      title?: string;
      resourceId: string;
      message: string;
      recommendation?: string;
      previousSeverity?: string;
      currentSeverity?: string;
      change: string;
      ownerHint?: string;
      impact?: string;
      effort?: string;
    }>;
    findingsImproved?: Array<{
      id: string;
      title?: string;
      resourceId: string;
      message: string;
      recommendation?: string;
      previousSeverity?: string;
      currentSeverity?: string;
      change: string;
      ownerHint?: string;
      impact?: string;
      effort?: string;
    }>;
    persistentFindingCount?: number;
    backupToolPrevious?: string;
    backupToolCurrent?: string;
    backupToolChanged?: boolean;
  };
  score: {
    maturity: string;
    overall: { final: number };
    storage: { final: number };
    workload: { final: number };
    config: { final: number };
    backup: { final: number };
  };
  inventory: {
    nodes?: Array<Record<string, unknown>>;
    namespaces?: Array<Record<string, unknown>>;
    pods?: Array<Record<string, unknown>>;
    deployments?: Array<Record<string, unknown>>;
    daemonSets?: Array<Record<string, unknown>>;
    jobs?: Array<Record<string, unknown>>;
    cronJobs?: Array<Record<string, unknown>>;
    pvcs?: Array<Record<string, unknown>>;
    pvs?: Array<Record<string, unknown>>;
    storageClasses?: Array<Record<string, unknown>>;
    services?: Array<Record<string, unknown>>;
    ingresses?: Array<Record<string, unknown>>;
    networkPolicies?: Array<Record<string, unknown>>;
    configMaps?: Array<Record<string, unknown>>;
    secrets?: Array<Record<string, unknown>>;
    clusterRoles?: Array<Record<string, unknown>>;
    clusterRoleBindings?: Array<Record<string, unknown>>;
    resourceQuotas?: Array<Record<string, unknown>>;
    hpas?: Array<Record<string, unknown>>;
    podDisruptionBudgets?: Array<Record<string, unknown>>;
    limitRanges?: Array<Record<string, unknown>>;
    helmReleases?: Array<Record<string, unknown>>;
    images?: Array<Record<string, unknown>>;
    certificates?: Array<Record<string, unknown>>;
    findings?: Finding[];
    remediationSteps?: RemediationStep[];
    backup?: {
      primaryTool?: string;
      coverageStatus?: string;
      coverageReason?: string;
      coverageVerified?: boolean;
      coveredNamespaces?: string[];
      uncoveredStatefulNamespaces?: string[];
      offsiteCoveredNamespaces?: string[];
      offsiteMissingNamespaces?: string[];
      hasOffsite?: boolean;
      policies?: BackupPolicy[];
      assurance?: {
        conclusion?: string;
        confidence?: string;
        summary?: string;
        signals?: Array<{ id: string; status: string; confidence?: string; summary: string; detail?: string }>;
      };
      drillPlan?: Array<{ phase: string; title: string; detail: string; ownerHint?: string; validation?: string[] }>;
      restoreSim?: {
        namespaces: RestoreNamespace[];
        totalPvcsGb: number;
        coveredPvcsGb: number;
        estimatedDataAtRiskGb?: number;
        readyNamespaces?: number;
        blockedNamespaces?: number;
        warningNamespaces?: number;
        unknownNamespaces?: number;
        blockingReasons?: string[];
        uncoveredNamespaces?: string[];
      };
    };
  };
};

export type Workspace = {
  bundle: Bundle;
  artifacts: ArtifactPaths;
  history: HistoryDashboard;
  source: string;
  loadedAt: string;
};

export type RunCompletionSummary = {
  runId: string;
  clusterName?: string;
  environment?: string;
  generatedAt?: string;
  score?: number;
  findingCount?: number;
  hasComparison?: boolean;
  artifacts: ArtifactPaths;
};

export type RunEvent = {
  type: string;
  runId: string;
  timestamp: string;
  step?: string;
  level?: string;
  message: string;
  percent?: number;
  artifact?: string;
  warning?: string;
  skip?: { name: string; reason: string; rbac: boolean };
};

export type RunResult = {
  runId: string;
  exitCode: number;
  trendLabel?: string;
  trendDelta?: number;
  artifacts: ArtifactPaths;
  workspace: Workspace;
  preflight: PreflightReport;
};

export type ConnectionMethod =
  | "current"
  | "kubeconfig_file"
  | "kubeconfig_inline"
  | "api_endpoint";

export type ScanStage = "connect" | "validate" | "outputs" | "launch";

export type ScanRequest = {
  runId?: string;
  connectionMethod?: ConnectionMethod;
  kubeconfigPath?: string;
  kubeconfigContent?: string;
  contextName?: string;
  apiServerEndpoint?: string;
  bearerToken?: string;
  caCertPath?: string;
  caCertContent?: string;
  outputDir?: string;
  dryRun?: boolean;
  ci?: boolean;
  minScore?: number;
  timeoutSeconds?: number;
  customerId?: string;
  site?: string;
  clusterName?: string;
  environment?: string;
  target?: string;
  csvExport?: boolean;
  namespaces?: string[];
  compareTo?: string;
  summary?: boolean;
  redact?: boolean;
  profileName?: string;
  runbook?: boolean;
  insecure?: boolean;
  includeSecretMetadata?: boolean;
};

export type ExportRequest = {
  outputDir: string;
  report?: boolean;
  bundleJson?: boolean;
  csvExport?: boolean;
  summary?: boolean;
  runbook?: boolean;
  redact?: boolean;
};

export type Settings = {
  workspaceRoot: string;
  defaultOutputDir: string;
  defaultProfile: string;
  includeSecretMetadata: boolean;
  summary: boolean;
  runbook: boolean;
  redact: boolean;
  csvExport: boolean;
};
