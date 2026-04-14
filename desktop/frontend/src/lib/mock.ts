import type {
  AppAlert,
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

const bootstrap: Bootstrap = {
  theme: {
    palette: {
      background: "#0d1117",
      surface: "#161b22",
      border: "#30363d",
      text: "#c9d1d9",
      muted: "#8b949e",
      accent: "#58a6ff",
      success: "#7ee787",
      critical: "#f85149",
      high: "#ffa657",
      medium: "#f2cc60",
    },
    maturity: {
      platinum: "#79c0ff",
      gold: "#f2cc60",
      silver: "#c9d1d9",
      bronze: "#ffa657",
    },
    typography: {
      body: `"Segoe UI Variable Text","Segoe UI","Trebuchet MS",sans-serif`,
      title: `"Segoe UI Variable Display","Segoe UI","Trebuchet MS",sans-serif`,
      mono: `"Consolas","SFMono-Regular","Liberation Mono",monospace`,
    },
    radius: {
      xl: "30px",
      lg: "24px",
      md: "18px",
      sm: "14px",
    },
  },
};

const demoTimestamp = "2026-04-12T14:11:00Z";

const demoWorkspace: Workspace = {
  source: "bundle",
  loadedAt: "2026-04-12T16:00:00Z",
  artifacts: {
    outputDir: "./demo-out",
    bundleJson: "./demo-out/recovery-scan.json",
    enrichedJson: "./demo-out/recovery-enriched.json",
    htmlReport: "./demo-out/recovery-report.html",
    markdownReport: "./demo-out/recovery-report.md",
    summaryHtml: "./demo-out/recovery-summary.html",
    runbookHtml: "./demo-out/recovery-runbook.html",
    redactedJson: "./demo-out/recovery-scan-redacted.json",
    redactedHtml: "./demo-out/recovery-report-redacted.html",
    csvDir: "./demo-out/csv",
    historyIndex: "./demo-out/history/index.json",
    loadedBundlePath: "./demo-out/recovery-scan.json",
  },
  history: {
    trendLabel: "IMPROVING",
    trendDelta: 6,
    averageScore: 79,
    bestScore: 85,
    worstScore: 74,
    runCount: 3,
    domainTrends: [
      { name: "storage", current: 88, delta: 4, direction: "up" },
      { name: "workload", current: 81, delta: 1, direction: "up" },
      { name: "config", current: 83, delta: -2, direction: "down" },
      { name: "backup", current: 86, delta: 6, direction: "up" },
    ],
    entries: [
      { timestampUtc: "2026-04-01T14:11:00Z", overall: 74, storage: 72, workload: 68, config: 79, backup: 77, findings: 6, maturity: "SILVER" },
      { timestampUtc: "2026-04-05T14:11:00Z", overall: 79, storage: 84, workload: 80, config: 85, backup: 80, findings: 4, maturity: "GOLD" },
      { timestampUtc: "2026-04-12T14:11:00Z", overall: 85, storage: 88, workload: 81, config: 83, backup: 86, findings: 2, maturity: "GOLD" },
    ],
  },
  bundle: {
    schemaVersion: "3.1.0",
    target: "vm",
    profile: "enterprise",
    scanNamespaces: ["payments", "frontend", "platform"],
    metadata: {
      customerId: "acme-hospitality",
      site: "us-east-1a",
      clusterName: "prod-east",
      environment: "production",
      generatedAt: "2026-04-12T14:11:00Z",
      toolVersion: "1.9.1",
    },
    tool: {
      name: "k8s-recovery-visualizer",
      version: "1.9.1",
      buildDate: "2026-04-14",
    },
    scan: {
      scanId: "scan-demo-001",
      startedAt: "2026-04-12T14:09:42Z",
      endedAt: "2026-04-12T14:11:00Z",
      durationSeconds: 78,
    },
    cluster: {
      apiServer: { endpoint: "https://prod-east.example.net:6443" },
      platform: {
        provider: "EKS",
        k8sVersion: "1.30",
        clusterUID: "cluster-demo-prod-east",
      },
    },
    score: {
      maturity: "GOLD",
      overall: { final: 85 },
      storage: { final: 88 },
      workload: { final: 81 },
      config: { final: 83 },
      backup: { final: 86 },
    },
    trendHistory: [
      { ts: "2026-04-01T14:11:00Z", score: 74, storage: 72, workload: 68, config: 79, backup: 77, findings: 6, maturity: "SILVER" },
      { ts: "2026-04-05T14:11:00Z", score: 79, storage: 84, workload: 80, config: 85, backup: 80, findings: 4, maturity: "GOLD" },
      { ts: "2026-04-12T14:11:00Z", score: 85, storage: 88, workload: 81, config: 83, backup: 86, findings: 2, maturity: "GOLD" },
    ],
    collectorSkips: [
      { name: "Secrets", reason: "forbidden: secrets access intentionally withheld", rbac: true },
    ],
    comparison: {
      previousScannedAt: "2026-04-05T14:11:00Z",
      previousScore: 79,
      previousMaturity: "GOLD",
      currentScore: 85,
      currentMaturity: "GOLD",
      scoreDelta: 6,
      namespacesAdded: ["frontend"],
      namespacesRemoved: [],
      workloadsAdded: ["frontend/web (Deployment)"],
      workloadsRemoved: [],
      pvcsAdded: ["payments/postgres-data"],
      pvcsRemoved: [],
      imagesAdded: ["ghcr.io/example/frontend:1.8.4"],
      imagesRemoved: ["nginx:1.25"],
      backupToolPrevious: "velero",
      backupToolCurrent: "velero",
      backupToolChanged: false,
      domainDeltas: [
        { name: "overall", previous: 79, current: 85, delta: 6 },
        { name: "storage", previous: 84, current: 88, delta: 4 },
        { name: "workload", previous: 80, current: 81, delta: 1 },
        { name: "config", previous: 85, current: 83, delta: -2 },
        { name: "backup", previous: 80, current: 86, delta: 6 },
      ],
      severityDeltas: [
        { severity: "CRITICAL", previous: 1, current: 0, delta: -1 },
        { severity: "HIGH", previous: 0, current: 1, delta: 1 },
        { severity: "MEDIUM", previous: 2, current: 1, delta: -1 },
        { severity: "LOW", previous: 1, current: 1, delta: 0 },
        { severity: "INFO", previous: 0, current: 0, delta: 0 },
      ],
      inventoryDeltas: [
        { name: "namespaces", added: 1, removed: 0 },
        { name: "workloads", added: 1, removed: 0 },
        { name: "pvcs", added: 1, removed: 0 },
        { name: "images", added: 1, removed: 1 },
      ],
      findingsNew: [
        {
          id: "NP_MISSING_FRONTEND",
          title: "Missing namespace NetworkPolicy",
          severity: "HIGH",
          resourceId: "namespace/frontend",
          message: "The frontend namespace lacks an egress-aware NetworkPolicy.",
          recommendation: "Add namespace-default ingress and egress policies before the next DR drill.",
          impact: "degraded recovery",
          ownerHint: "Platform engineering",
          effort: "S",
          rank: 1,
        },
      ],
      findingsResolved: [
        {
          id: "ETCD_BACKUP",
          title: "Missing etcd backup evidence",
          severity: "CRITICAL",
          resourceId: "cluster/prod-east",
          message: "No etcd backup evidence was found.",
          recommendation: "Keep the scheduled etcd snapshot job healthy.",
        },
      ],
      findingsRegressed: [
        {
          id: "BACKUP_NO_POLICIES",
          title: "No backup schedules detected",
          resourceId: "cluster/prod-east",
          message: "A previously low-severity schedule drift is now a critical policy gap.",
          previousSeverity: "MEDIUM",
          currentSeverity: "CRITICAL",
          change: "severity-up",
          ownerHint: "Platform / backup owner",
          impact: "coverage gap",
          effort: "M",
        },
      ],
      findingsImproved: [
        {
          id: "ETCD_BACKUP",
          title: "Missing etcd backup evidence",
          resourceId: "cluster/prod-east",
          message: "The control plane backup path now has verified evidence.",
          previousSeverity: "CRITICAL",
          currentSeverity: "INFO",
          change: "severity-down",
        },
      ],
      persistentFindingCount: 3,
    },
    inventory: {
      namespaces: [
        { id: "ns:payments", name: "payments", psaEnforce: "restricted" },
        { id: "ns:frontend", name: "frontend", psaEnforce: "restricted" },
        { id: "ns:platform", name: "platform", psaEnforce: "baseline" },
      ],
      nodes: [
        { name: "ip-10-0-0-12", roles: ["worker"], ready: true, zone: "us-east-1a", kubeletVersion: "v1.30.1", osImage: "Amazon Linux 2023" },
        { name: "ip-10-0-0-13", roles: ["worker"], ready: true, zone: "us-east-1b", kubeletVersion: "v1.30.1", osImage: "Amazon Linux 2023" },
      ],
      pods: [
        { namespace: "payments", name: "postgres-0", containerCount: 1, hasRequests: true, hasLimits: true, privileged: false, hostNetwork: false, hostPID: false },
        { namespace: "frontend", name: "web-7bc9d8", containerCount: 2, hasRequests: false, hasLimits: true, privileged: false, hostNetwork: false, hostPID: false },
      ],
      deployments: [
        { namespace: "frontend", name: "web", replicas: 3, ready: 3, images: ["ghcr.io/example/frontend:1.8.4", "ghcr.io/example/sidecar:0.6.0"] },
        { namespace: "platform", name: "grafana", replicas: 2, ready: 2, images: ["grafana/grafana:11.0.0"] },
      ],
      daemonSets: [
        { namespace: "platform", name: "log-agent", desired: 2, ready: 2, images: ["public.ecr.aws/aws-observability/fluent-bit:2.1.9"] },
      ],
      jobs: [{ namespace: "platform", name: "schema-migration", completed: true }],
      cronJobs: [{ namespace: "platform", name: "etcd-snapshot", schedule: "0 */4 * * *" }],
      pvcs: [
        { namespace: "payments", name: "postgres-data", storageClass: "gp3", requestedSize: "200Gi" },
      ],
      pvs: [
        { name: "pvc-0f3a", claimRef: "payments/postgres-data", storageClass: "gp3", capacity: "200Gi", reclaimPolicy: "Retain", backend: "EBS" },
      ],
      storageClasses: [
        { name: "gp3", provisioner: "ebs.csi.aws.com", reclaimPolicy: "Delete", volumeBindingMode: "WaitForFirstConsumer" },
      ],
      services: [
        { namespace: "frontend", name: "web", type: "LoadBalancer", clusterIp: "172.20.12.8", externalIp: "34.228.10.10" },
      ],
      ingresses: [
        { namespace: "frontend", name: "web", className: "nginx", tls: true, rules: [{ host: "app.example.com", backend: "web:443" }] },
      ],
      networkPolicies: [
        { namespace: "payments", name: "default-deny", podSelector: "all", hasIngress: true, hasEgress: true },
      ],
      configMaps: [
        { namespace: "platform", name: "grafana-dashboards", keyCount: 14 },
      ],
      secrets: [
        { namespace: "payments", name: "db-credentials", type: "Opaque", keyCount: 2 },
      ],
      clusterRoles: [
        { name: "ops-admin", custom: true, ruleCount: 17, hasWildcardVerb: true, hasSecretAccess: true },
      ],
      clusterRoleBindings: [
        { name: "ops-admin-binding", roleName: "ops-admin", subjects: ["Group:platform-ops"] },
      ],
      resourceQuotas: [
        { namespace: "frontend", name: "frontend-quota", hardPods: "30", hardCpu: "12" },
      ],
      hpas: [
        { namespace: "frontend", name: "web", target: "Deployment/web", minReplicas: 2, maxReplicas: 8, currentReplicas: 3 },
      ],
      podDisruptionBudgets: [
        { namespace: "frontend", name: "web", minAvailable: "2" },
      ],
      limitRanges: [
        { namespace: "frontend", name: "defaults", type: "Container" },
      ],
      helmReleases: [
        { namespace: "platform", name: "grafana", chart: "grafana", version: "7.3.4", appVersion: "11.0.0", status: "deployed" },
      ],
      images: [
        { image: "ghcr.io/example/frontend:1.8.4", registry: "ghcr.io", isPublic: false, workloads: ["frontend/web"] },
        { image: "postgres:16.4", registry: "docker.io", isPublic: true, workloads: ["payments/postgres"] },
      ],
      certificates: [
        { namespace: "frontend", name: "web-tls", issuer: "letsencrypt-prod", secretName: "web-tls", ready: true, notAfter: "2026-05-23T10:00:00Z", daysToExpiry: 41 },
      ],
      findings: [
        {
          id: "NP_MISSING_FRONTEND",
          title: "Missing namespace NetworkPolicy",
          severity: "HIGH",
          resourceId: "namespace/frontend",
          message: "The frontend namespace lacks an egress-aware NetworkPolicy.",
          recommendation: "Add namespace-default ingress and egress policies before the next DR drill.",
          impact: "degraded recovery",
          effort: "S",
          ownerHint: "Platform engineering",
          priorityScore: 91,
          rank: 1,
        },
        {
          id: "REQUESTS_MISSING",
          title: "Missing workload resource requests",
          severity: "MEDIUM",
          resourceId: "frontend/web",
          message: "Several frontend pods are missing CPU requests.",
          recommendation: "Add sane CPU and memory requests for recovery predictability.",
          impact: "operational risk",
          effort: "S",
          ownerHint: "Application owner",
          priorityScore: 52,
          rank: 2,
        },
      ],
      remediationSteps: [
        {
          priority: 1,
          category: "Networking",
          title: "Add default frontend NetworkPolicies",
          detail: "The frontend namespace should deny-by-default and explicitly allow platform egress dependencies.",
          ownerHint: "Platform engineering",
          effort: "S",
          whyItMatters: "Recovery tests often fail because rebuilt workloads can reach nothing or everything.",
          drImpact: "Namespace rebuilds may expose services broadly or fail to reconnect to dependencies.",
          validation: ["Run kubectl get networkpolicies -n frontend", "Confirm ingress and egress rules exist."],
          fixSteps: ["Create a namespace default deny policy.", "Add explicit egress rules for DNS, database, and observability endpoints."],
          commands: ["kubectl apply -n frontend -f frontend-networkpolicies.yaml"],
        },
        {
          priority: 2,
          category: "Workload",
          title: "Backfill CPU requests for the frontend web deployment",
          detail: "Recovery placement is less predictable when requests are omitted.",
          ownerHint: "Application owner",
          effort: "S",
          validation: ["Inspect rendered Deployment resources after the fix."],
          fixSteps: ["Set requests.cpu and requests.memory on the web containers."],
        },
      ],
      backup: {
        primaryTool: "velero",
        coverageStatus: "verified",
        coverageReason: "Schedules were parsed successfully.",
        coverageVerified: true,
        coveredNamespaces: ["payments", "frontend"],
        uncoveredStatefulNamespaces: [],
        offsiteCoveredNamespaces: ["payments", "frontend"],
        offsiteMissingNamespaces: [],
        hasOffsite: true,
        policies: [
          {
            tool: "velero",
            name: "prod-hourly",
            includedNamespaces: ["payments", "frontend"],
            schedule: "0 * * * *",
            rpoHours: 1,
            lastSuccessAt: "2026-04-12T13:00:00Z",
            confidence: "confirmed",
            hasOffsite: true,
            retentionTtl: "720h0m0s",
          },
        ],
        assurance: {
          conclusion: "evidence_confirmed",
          confidence: "high",
          summary: "Coverage, restore simulation, and offsite evidence are aligned for the current scope.",
          signals: [
            { id: "coverage", status: "verified", confidence: "high", summary: "Velero schedules cover stateful namespaces." },
            { id: "offsite", status: "verified", confidence: "high", summary: "All covered namespaces have offsite evidence." },
          ],
        },
        restoreSim: {
          namespaces: [
            { namespace: "payments", coverageKnown: true, hasCoverage: true, readiness: "ready", rpoHours: 1, pvcSizeGb: 200, blockers: [], warnings: [] },
          ],
          totalPvcsGb: 200,
          coveredPvcsGb: 200,
          estimatedDataAtRiskGb: 0,
          readyNamespaces: 1,
          blockedNamespaces: 0,
          warningNamespaces: 0,
          unknownNamespaces: 0,
          blockingReasons: [],
          uncoveredNamespaces: [],
        },
        drillPlan: [
          {
            phase: "prepare",
            title: "Freeze the evidence set",
            detail: "Use this bundle as the restore baseline and record the target cluster before the drill starts.",
            ownerHint: "Platform / incident lead",
            validation: ["Attach recovery-scan.json to the drill ticket."],
          },
          {
            phase: "validate",
            title: "Re-scan after restore",
            detail: "Generate a fresh bundle from the restored target and compare it to the baseline.",
            ownerHint: "Platform / incident lead",
            validation: ["Review new and regressed findings in Compare."],
          },
        ],
      },
    },
  },
};

const demoProjects: ProjectSummary[] = [
  {
    name: "prod-east",
    clusterName: "prod-east",
    environment: "production",
    outputDir: "./demo-out",
    lastScanPath: "./demo-out/recovery-scan.json",
    reportPath: "./demo-out/recovery-report.html",
    score: 85,
    maturity: "GOLD",
    timestampUtc: "2026-04-12T14:11:00Z",
  },
  {
    name: "staging-west",
    clusterName: "staging-west",
    environment: "staging",
    outputDir: "./staging-out",
    lastScanPath: "./staging-out/recovery-scan.json",
    reportPath: "./staging-out/recovery-report.html",
    score: 72,
    maturity: "SILVER",
    timestampUtc: "2026-04-11T09:04:00Z",
  },
];

const demoConnectionAdvisor: ConnectionAdvisor = {
  recommendedMethod: "current",
  recommendedReason: "This machine already has Kubernetes access configured, so existing access is the simplest place to start.",
  kubectlAvailable: true,
  kubectlPath: "C:/Program Files/Kubernetes/kubectl.exe",
  currentLoginAvailable: true,
  currentContext: "prod-east-admin",
  currentLoginDetail: "Detected local Kubernetes access with current context \"prod-east-admin\".",
  currentLoginWarning: "This kubeconfig depends on an external auth helper. Existing access can still work here, but keep kubeconfig mode handy if the helper is unavailable in the desktop session.",
  defaultKubeconfigAvailable: true,
  defaultKubeconfigPath: "C:/Users/demo/.kube/prod-cluster.backup",
  defaultKubeconfigCurrentContext: "prod-east-admin",
  defaultKubeconfigDetail: "Loaded kubeconfig file with 3 contexts, 2 clusters, and 1 user entry. Current context: prod-east-admin.",
  defaultKubeconfigPortable: true,
  defaultKubeconfigWarning: "The detected kubeconfig uses an external auth helper. Existing access can still work, but it depends on that helper being available and signed in on this machine.",
};

const demoPreflight = (request: ScanRequest): PreflightReport => ({
  canRun: true,
  degraded: Boolean(request.includeSecretMetadata),
  server: request.dryRun
    ? ""
    : request.connectionMethod === "api_endpoint"
      ? request.apiServerEndpoint || "https://prod-east.example.net:6443"
      : "https://prod-east.example.net:6443",
  contextName: request.connectionMethod === "api_endpoint" ? "" : request.contextName || "prod-east-admin",
  scope: request.namespaces?.length ? request.namespaces.join(", ") : "all namespaces",
  checks: [
    {
      id: "config",
      title: "Kubernetes credentials",
      status: "pass",
      required: true,
      detail: request.dryRun
        ? "Dry-run mode skips cluster auth."
        : request.connectionMethod === "kubeconfig_file"
          ? "Kubeconfig file loaded successfully."
          : request.connectionMethod === "kubeconfig_inline"
            ? "Pasted kubeconfig loaded successfully."
            : request.connectionMethod === "api_endpoint"
              ? "Direct API connection details loaded successfully."
              : "Current Kubernetes login loaded successfully.",
    },
    { id: "api", title: "API server reachability", status: request.dryRun ? "pass" : "pass", required: true, detail: request.dryRun ? "No API server contact needed." : "Cluster API is reachable." },
    { id: "pods", title: "Read workloads", status: "pass", required: true, detail: "List access confirmed for pods and workload controllers." },
    { id: "storage", title: "Read storage inventory", status: "pass", required: true, detail: "PVC and PV inventory access confirmed." },
    {
      id: "secrets",
      title: "Secret metadata collection",
      status: request.includeSecretMetadata ? "warn" : "pass",
      required: false,
      scope: "namespace",
      resource: "secrets",
      detail: request.includeSecretMetadata
        ? "Secret metadata is enabled. Granting this access increases scanner visibility into sensitive resources."
        : "Secret metadata remains opt-in and disabled by default.",
      hint: "Only enable this if you intentionally want Secret type and key-count metadata included.",
      commands: request.includeSecretMetadata ? ["kubectl auth can-i list secrets -n payments"] : undefined,
      manifest: request.includeSecretMetadata
        ? "apiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole\nmetadata:\n  name: k8v-secrets-reader\nrules:\n- apiGroups: [\"\"]\n  resources: [\"secrets\"]\n  verbs: [\"get\", \"list\", \"watch\"]"
        : undefined,
    },
  ],
  warnings: request.includeSecretMetadata ? ["Secret metadata collection is more invasive than the default RBAC profile."] : [],
});

const listeners = new Set<(event: RunEvent) => void>();

function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

function emit(event: RunEvent) {
  for (const listener of listeners) {
    listener(event);
  }
}

function currentSettings(): Settings {
  const raw = globalThis.localStorage?.getItem("k8vis:settings");
  if (raw) {
    return JSON.parse(raw) as Settings;
  }
  return {
    workspaceRoot: ".",
    defaultOutputDir: "./out",
    defaultProfile: "enterprise",
    includeSecretMetadata: false,
    summary: true,
    runbook: true,
    redact: false,
    csvExport: true,
  };
}

function inspectKubeconfig(request: ScanRequest): KubeconfigInspection {
  const source = request.kubeconfigContent ? "pasted kubeconfig" : "kubeconfig file";
  const path = request.kubeconfigPath?.trim() || "";
  const rawContent = request.kubeconfigContent?.trim() || "";

  if (!path && !rawContent) {
    throw new Error("choose a kubeconfig file or paste kubeconfig content first");
  }
  if (path.toLowerCase().includes("missing")) {
    throw new Error(`read kubeconfig file "${path}": The system cannot find the file specified.`);
  }
  if (path.toLowerCase().includes("invalid") || rawContent.toLowerCase().includes("not-a-kubeconfig")) {
    throw new Error(`the selected ${source} is not a usable kubeconfig: missing clusters, contexts, and users`);
  }
  if (path.toLowerCase().includes("yaml-error") || rawContent.toLowerCase().includes("yaml-error")) {
    throw new Error(`parse ${source}: yaml: line 3: did not find expected key`);
  }

  return {
    source,
    path: path || undefined,
    currentContext: "prod-east-admin",
    contexts: ["kind-k8v-test", "prod-east-admin", "staging-west-admin"],
    clusterCount: 2,
    userCount: 1,
    summary: `Loaded ${source} with 3 contexts, 2 clusters, and 1 user entry. Current context: prod-east-admin.`,
    nextAction: "Test the connection next. If it succeeds, continue to scope and outputs before running full preflight.",
  };
}

function demoConnectionTest(request: ScanRequest): ConnectionTestReport {
  if (request.connectionMethod === "api_endpoint") {
    const hasTrust = Boolean(request.caCertPath || request.caCertContent || request.insecure);
    if (!request.apiServerEndpoint) {
      return {
        canConnect: false,
        summary: "Connection test failed.",
        nextAction: "Enter the Kubernetes API server host or URL first.",
        diagnosis: {
          code: "endpoint_unreachable",
          label: "API reachability",
          summary: "The cluster API is not reachable from this machine.",
          detail: "No API server endpoint was provided.",
          nextAction: "Enter the Kubernetes API server host or URL first.",
        },
        fieldErrors: { apiServerEndpoint: "Enter the Kubernetes API server host or URL." },
        checks: [{ id: "transport", title: "API server reachability", status: "fail", detail: "No API server endpoint was provided." }],
      };
    }
    if (!request.bearerToken) {
      return {
        canConnect: false,
        summary: "Credentials were rejected.",
        nextAction: "Paste a short-lived bearer token before testing the connection again.",
        diagnosis: {
          code: "auth_rejected",
          label: "Credentials",
          summary: "The API server rejected the current credentials.",
          detail: "No bearer token was provided.",
          nextAction: "Paste a short-lived bearer token before testing the connection again.",
        },
        fieldErrors: { bearerToken: "Paste a bearer token for direct API endpoint scans." },
        checks: [{ id: "auth", title: "Credential acceptance", status: "fail", detail: "No bearer token was provided." }],
      };
    }
    if (!hasTrust) {
      return {
        canConnect: false,
        source: "direct API endpoint",
        server: request.apiServerEndpoint,
        summary: "TLS verification failed.",
        nextAction: "Add the issuing CA certificate or use skip-TLS only as a temporary workaround in a trusted environment.",
        diagnosis: {
          code: "tls_trust",
          label: "TLS trust",
          summary: "The cluster is reachable, but this machine does not trust the API server certificate.",
          detail: "x509: certificate signed by unknown authority",
          nextAction: "Add the issuing CA certificate or use skip-TLS only as a temporary workaround in a trusted environment.",
        },
        fieldErrors: { caTrust: "The API server certificate could not be verified. Add the issuing CA or use skip-TLS only temporarily." },
        checks: [{ id: "transport", title: "TLS and API handshake", status: "fail", detail: "x509: certificate signed by unknown authority" }],
      };
    }
  }

  if (request.connectionMethod === "kubeconfig_file" || request.connectionMethod === "kubeconfig_inline") {
    inspectKubeconfig(request);
  }

  return {
    canConnect: true,
    source: request.connectionMethod === "api_endpoint" ? "direct API endpoint" : request.connectionMethod === "kubeconfig_inline" ? "pasted kubeconfig" : request.connectionMethod === "kubeconfig_file" ? "kubeconfig file" : "current Kubernetes login",
    server: request.connectionMethod === "api_endpoint" ? request.apiServerEndpoint || "https://prod-east.example.net:6443" : "https://prod-east.example.net:6443",
    contextName: request.connectionMethod === "api_endpoint" ? "" : request.contextName || "prod-east-admin",
    summary: "Connection test succeeded.",
    nextAction: "Continue to scope and outputs, then run full preflight to check RBAC and collection readiness before the scan.",
    checks: [
      { id: "config", title: "Connection settings loaded", status: "pass", detail: request.connectionMethod === "kubeconfig_file" ? "Kubeconfig file loaded successfully." : request.connectionMethod === "kubeconfig_inline" ? "Pasted kubeconfig loaded successfully." : request.connectionMethod === "api_endpoint" ? "Direct API connection details loaded successfully." : "Current Kubernetes login loaded successfully." },
      { id: "transport", title: "API server reachability", status: "pass", detail: "Reached the cluster API successfully." },
      { id: "handshake", title: "Basic API handshake", status: "pass", detail: "The cluster responded to a basic discovery request." },
    ],
  };
}

export const mockBackend = {
  async GetBootstrap(): Promise<Bootstrap> {
    return clone(bootstrap);
  },
  async GetSettings(): Promise<Settings> {
    return clone(currentSettings());
  },
  async GetStartupAlerts(): Promise<AppAlert[]> {
    return [];
  },
  async SaveSettings(settings: Settings): Promise<void> {
    globalThis.localStorage?.setItem("k8vis:settings", JSON.stringify(settings));
  },
  async ListProjects(): Promise<ProjectSummary[]> {
    return clone(demoProjects);
  },
  async GetConnectionAdvisor(): Promise<ConnectionAdvisor> {
    return clone(demoConnectionAdvisor);
  },
  async InspectKubeconfig(request: ScanRequest): Promise<KubeconfigInspection> {
    return clone(inspectKubeconfig(request));
  },
  async ListConnectionContexts(request: ScanRequest): Promise<ContextCatalog> {
    if (request.connectionMethod === "api_endpoint") {
      return { source: "direct API endpoint", contexts: [] };
    }
    if (request.connectionMethod === "kubeconfig_inline") {
      return {
        source: "pasted kubeconfig",
        currentContext: request.contextName || "prod-east-admin",
        contexts: ["kind-k8v-test", "prod-east-admin", "staging-west-admin"],
      };
    }
    return {
      source: request.connectionMethod === "kubeconfig_file" ? "kubeconfig file" : "current Kubernetes login",
      currentContext: request.contextName || "prod-east-admin",
      contexts: ["kind-k8v-test", "prod-east-admin", "staging-west-admin"],
    };
  },
  async TestConnection(request: ScanRequest): Promise<ConnectionTestReport> {
    return clone(demoConnectionTest(request));
  },
  async RunPreflight(request: ScanRequest): Promise<PreflightReport> {
    return clone(demoPreflight(request));
  },
  async RunScan(request: ScanRequest): Promise<RunResult> {
    const runId = request.runId || `mock-run-${Date.now()}`;
    const checkpoints: Array<[number, string, string]> = [
      [0.08, "preflight", "Preflight checks complete."],
      [0.18, "connect", "Connecting to the Kubernetes API."],
      [0.31, "Namespaces", "Collecting Namespaces."],
      [0.46, "StatefulSets", "Collecting StatefulSets."],
      [0.61, "Images", "Collecting Images."],
      [0.74, "analysis", "Analyzing inventory and generating remediation guidance."],
      [0.88, "artifacts", "Writing offline report bundle."],
      [1, "complete", "Scan complete."],
    ];
    for (const [percent, step, message] of checkpoints) {
      emit({
        type: step === "complete" ? "complete" : "status",
        runId,
        timestamp: demoTimestamp,
        step,
        level: "info",
        message,
        percent,
        artifact: step === "complete" ? demoWorkspace.artifacts.htmlReport : undefined,
      });
      await new Promise((resolve) => setTimeout(resolve, step === "complete" ? 120 : 160));
    }
    const workspace = clone(demoWorkspace);
    workspace.bundle.metadata.generatedAt = demoTimestamp;
    return {
      runId,
      exitCode: 0,
      trendLabel: "IMPROVING",
      trendDelta: 6,
      artifacts: workspace.artifacts,
      workspace,
      preflight: demoPreflight(request),
    };
  },
  async CancelRun(): Promise<void> {
    emit({
      type: "warning",
      runId: "mock-run",
      timestamp: demoTimestamp,
      step: "cancel",
      level: "warn",
      message: "Cancellation requested.",
      warning: "The mock backend acknowledges the cancel request immediately.",
    });
  },
  async OpenBundle(): Promise<Workspace> {
    return clone(demoWorkspace);
  },
  async ExportBundle(path: string, request: ExportRequest) {
    const outputDir = request.outputDir || demoWorkspace.artifacts.outputDir;
    return {
      outputDir,
      bundleJson: request.bundleJson || request.report ? `${outputDir}/recovery-scan.json` : undefined,
      enrichedJson: request.bundleJson || request.report ? `${outputDir}/recovery-enriched.json` : undefined,
      htmlReport: request.report ? `${outputDir}/recovery-report.html` : undefined,
      markdownReport: request.bundleJson || request.report ? `${outputDir}/recovery-report.md` : undefined,
      summaryHtml: request.summary ? `${outputDir}/recovery-summary.html` : undefined,
      runbookHtml: request.runbook ? `${outputDir}/recovery-runbook.html` : undefined,
      redactedJson: request.redact ? `${outputDir}/recovery-scan-redacted.json` : undefined,
      redactedHtml: request.redact ? `${outputDir}/recovery-report-redacted.html` : undefined,
      csvDir: request.csvExport ? `${outputDir}/csv` : undefined,
      historyIndex: request.bundleJson || request.report ? `${outputDir}/history/index.json` : undefined,
      loadedBundlePath: path || demoWorkspace.artifacts.loadedBundlePath,
    };
  },
  async OpenPath(): Promise<void> {
    return;
  },
  async PickBundleFile(): Promise<string> {
    return demoWorkspace.artifacts.loadedBundlePath || "./demo-out/recovery-scan.json";
  },
  async PickKubeconfigFile(): Promise<string> {
    return "C:/Users/demo/.kube/prod-cluster.backup";
  },
  async PickCertificateFile(): Promise<string> {
    return "C:/Users/demo/certs/cluster-ca.pem";
  },
  async PickOutputDirectory(): Promise<string> {
    return "./demo-out";
  },
  onScanEvent(listener: (event: RunEvent) => void) {
    listeners.add(listener);
    return () => {
      listeners.delete(listener);
    };
  },
  demoWorkspace,
};
