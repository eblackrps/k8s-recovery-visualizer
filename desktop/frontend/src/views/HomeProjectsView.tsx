import type { ConnectionAdvisor, ProjectSummary, Workspace } from "../lib/types";
import {
  Badge,
  Card,
  DataTable,
  EmptyState,
  KeyValueGrid,
  MetricCard,
  ReadinessList,
  SectionHeader,
  TrendRail,
  formatDelta,
  formatShortDate,
  toneForDelta,
  toneForMaturity,
} from "../components/ui";

export function HomeView(props: {
  workspace: Workspace | null;
  projects: ProjectSummary[];
  connectionAdvisor: ConnectionAdvisor | null;
  busy: boolean;
  onViewProjects: () => void;
  onPickBundle: () => void;
  onOpenProject: (path: string) => void;
  onStartScan: () => void;
  onReviewFindings: () => void;
}) {
  const bundle = props.workspace?.bundle;
  const comparison = bundle?.comparison;
  const restoreSim = bundle?.inventory.backup?.restoreSim;
  const historyEntries = props.workspace?.history.entries || [];
  const trendDelta = props.workspace?.history.trendDelta || 0;
  const trendTone = toneForDelta(trendDelta);
  const openFindings = (bundle?.inventory.findings || []).length;
  const hasHistory = props.projects.length > 0;
  const isFirstRun = !bundle && !hasHistory;

  return (
    <section className="page-grid home-grid">
      <section className="panel home-start-panel">
        <SectionHeader
          eyebrow="Start Here"
          title="Assess a cluster or reopen a saved bundle"
          description="K8V checks Kubernetes disaster-recovery readiness. A live scan validates access, collects cluster evidence, and writes a portable bundle you can reopen later without cluster access."
          dense={!isFirstRun}
        />
        <div className="action-grid action-grid-two">
          <button type="button" className="action-tile action-primary" onClick={props.onStartScan}>
            <span className="eyebrow">Primary path</span>
            <strong>Connect to a cluster</strong>
            <p>Test access, run preflight, then collect a fresh recovery bundle.</p>
          </button>
          <button type="button" className="action-tile" onClick={props.onPickBundle} disabled={props.busy}>
            <span className="eyebrow">Offline review</span>
            <strong>Open an existing bundle</strong>
            <p>Review findings, compare drift, or export reports offline.</p>
          </button>
        </div>
        {isFirstRun ? (
          <div className="summary-three-up onboarding-grid">
            <Card title="How K8V works">
              <ol className="workflow-list">
                <li>Choose the access path that is most likely to work on this machine.</li>
                <li>Test the connection, then run preflight to check RBAC and collection readiness.</li>
                <li>Run the scan to write a portable bundle and offline report outputs.</li>
                <li>Reopen that bundle later for findings, compare, inventory, and exports.</li>
              </ol>
            </Card>
            <Card title="What a scan produces">
              <div className="artifact-list">
                <div className="artifact-row">
                  <strong>Portable bundle</strong>
                  <p className="muted">`recovery-scan.json` plus enriched metadata for offline review and comparison.</p>
                </div>
                <div className="artifact-row">
                  <strong>Offline reports</strong>
                  <p className="muted">HTML and Markdown outputs that can be shared without staying connected to the cluster.</p>
                </div>
                <div className="artifact-row">
                  <strong>Optional extras</strong>
                  <p className="muted">Summary, runbook, CSV, and redacted exports when you need them.</p>
                </div>
              </div>
            </Card>
            <Card title="Machine readiness" description="K8V checks what is already available on this machine so first-time operators know where to start.">
              <ReadinessList items={machineReadinessItems(props.connectionAdvisor)} />
            </Card>
          </div>
        ) : null}
      </section>

      <section className="panel home-posture-panel">
        {bundle ? (
          <>
            <SectionHeader
              eyebrow="Current Bundle"
              title={`${bundle.metadata.clusterName} · ${bundle.metadata.environment}`}
              description="The latest loaded bundle is ready for findings, restore readiness, compare, and export work."
              actions={
                <Badge tone={toneForMaturity(bundle.score.maturity)}>
                  {bundle.score.maturity} {bundle.score.overall.final}
                </Badge>
              }
            />
            <div className="home-score-strip">
              <div className="score-panel">
                <span className="eyebrow">Overall readiness</span>
                <strong>{bundle.score.overall.final}</strong>
                <p className="muted">
                  {`${bundle.cluster.platform?.provider || "Unknown"} ${bundle.cluster.platform?.k8sVersion || ""}`.trim()}
                </p>
              </div>
              <div className="metric-grid compact-grid">
                <MetricCard label="Open findings" value={openFindings} tone={openFindings ? "high" : "success"} />
                <MetricCard
                  label="Trend"
                  value={historyEntries.length > 1 ? formatDelta(trendDelta) : "First run"}
                  tone={trendTone === "critical" ? "critical" : trendTone === "success" ? "success" : "medium"}
                />
                <MetricCard label="Ready namespaces" value={restoreSim?.readyNamespaces || 0} tone="success" />
                <MetricCard
                  label="Data at risk GiB"
                  value={restoreSim?.estimatedDataAtRiskGb || 0}
                  tone={(restoreSim?.estimatedDataAtRiskGb || 0) > 0 ? "critical" : "success"}
                />
              </div>
            </div>
            <div className="summary-two-up home-posture-summary">
              <Card title="Bundle context">
                <KeyValueGrid
                  className="home-posture-kv"
                  items={[
                    ["Profile", bundle.profile || "standard"],
                    ["Scope", (bundle.scanNamespaces || []).join(", ") || "all namespaces"],
                    ["Generated", bundle.metadata.generatedAt || "Not available"],
                    ["Output", props.workspace?.artifacts.outputDir || "Not available"],
                  ]}
                />
              </Card>
              <Card title="Recovery signals">
                <KeyValueGrid
                  className="home-posture-kv"
                  items={[
                    ["Backup tool", bundle.inventory.backup?.primaryTool || "none"],
                    ["Coverage", bundle.inventory.backup?.coverageStatus || "unknown"],
                    ["Restore blockers", String(restoreSim?.blockedNamespaces || 0)],
                    ["Current maturity", bundle.score.maturity || "Unknown"],
                  ]}
                />
              </Card>
            </div>
          </>
        ) : isFirstRun ? (
          <EmptyState
            eyebrow="First Run"
            title="Run one scan, then work from the bundle"
            description="K8V writes a portable bundle and offline reports so operators can reopen results, compare drift, and export handoff material without staying connected to the cluster."
          >
            <div className="results-stack">
              {props.connectionAdvisor?.recommendedReason ? (
                <div className="notice notice-info compact">
                  <strong>Recommended on this machine</strong>
                  <p className="muted">{props.connectionAdvisor.recommendedReason}</p>
                </div>
              ) : null}
              <div className="stack-list compact-stack">
                <div className="empty-state-note">
                  <strong>Start with existing access</strong>
                  <p className="muted">Best when `kubectl` or the default kubeconfig already works on this desktop or jumpbox.</p>
                </div>
                <div className="empty-state-note">
                  <strong>Fallback to kubeconfig when needed</strong>
                  <p className="muted">Bring a kubeconfig file or paste raw kubeconfig content if the local login path is not the better fit here.</p>
                </div>
                <div className="empty-state-note">
                  <strong>Use direct API mode only when necessary</strong>
                  <p className="muted">Manual endpoint, bearer token, and CA setup is still available for clusters where kubeconfig mode is not the better option.</p>
                </div>
                <div className="empty-state-note">
                  <strong>Reopen the bundle later</strong>
                  <p className="muted">The scan output becomes the offline review package for findings, compare, inventory, and export refreshes.</p>
                </div>
              </div>
            </div>
          </EmptyState>
        ) : (
          <EmptyState
            eyebrow="Bundle Needed"
            title="No active bundle loaded yet"
            description="Start with a cluster connection if you want a fresh assessment, or open an existing bundle if someone already ran the scan."
          >
            <div className="results-stack">
              {props.connectionAdvisor?.recommendedReason ? (
                <div className="notice notice-info compact">
                  <strong>Recommended on this machine</strong>
                  <p className="muted">{props.connectionAdvisor.recommendedReason}</p>
                </div>
              ) : null}
              <Card title="If the first connection fails">
                <div className="stack-list compact-stack">
                  <div className="empty-state-note">
                    <strong>Use existing access</strong>
                    <p className="muted">Best when `kubectl` or the default kubeconfig already works on this desktop or jumpbox.</p>
                  </div>
                  <div className="empty-state-note">
                    <strong>Bring a kubeconfig</strong>
                    <p className="muted">Use this when you were handed a kubeconfig file or raw kubeconfig text. K8V validates the contents, not the filename.</p>
                  </div>
                  <div className="empty-state-note">
                    <strong>Direct API access</strong>
                    <p className="muted">Use this only for manual endpoint, bearer token, and CA setup when kubeconfig mode is not the better fit.</p>
                  </div>
                </div>
              </Card>
            </div>
          </EmptyState>
        )}
      </section>

      <section className="panel home-bundles-panel">
        <SectionHeader
          eyebrow="Recent Bundles"
          title={hasHistory ? "Saved assessments" : "No saved bundles yet"}
          description={
            hasHistory
              ? "Recent bundles stay available for offline review, comparisons, and export refreshes."
              : "Once a scan succeeds, the bundle written to the output directory will show up here for reopening later."
          }
          actions={
            hasHistory ? (
              <button type="button" className="button secondary quiet" onClick={props.onViewProjects}>
                Open Projects
              </button>
            ) : null
          }
          dense={!isFirstRun}
        />
        <div className="project-list compact">
          {hasHistory ? (
            props.projects.map((project) => (
              <button
                type="button"
                key={project.lastScanPath}
                className="project-row"
                onClick={() => props.onOpenProject(project.lastScanPath)}
                disabled={props.busy}
              >
                <div className="project-main">
                  <strong>{project.clusterName || project.name}</strong>
                  <span>{project.environment || "Unknown environment"}</span>
                </div>
                <div className="project-meta">
                  <span>{formatShortDate(project.timestampUtc || "") || "No date"}</span>
                  <Badge tone={toneForMaturity(project.maturity)}>{project.maturity}</Badge>
                  <strong>{project.score}</strong>
                </div>
              </button>
            ))
          ) : (
            <p className="muted">No bundles found yet. Run a scan or open an existing bundle to populate this workspace.</p>
          )}
        </div>
      </section>

      {isFirstRun ? null : (
        <section className="panel home-watch-panel">
          {bundle ? (
            <>
              <SectionHeader
                eyebrow="Operational Watch"
                title="What changed and what needs attention"
                description="The watchlist keeps regressions, new gaps, and restore-readiness movement visible without forcing you into the full report first."
                actions={
                  <button type="button" className="button secondary quiet" onClick={props.onReviewFindings}>
                    Review Findings
                  </button>
                }
                dense={!isFirstRun}
              />
              <div className="watch-grid">
                <article className="watch-item">
                  <span>Regressed findings</span>
                  <strong>{comparison?.findingsRegressed?.length || 0}</strong>
                  <p>{comparison?.findingsRegressed?.[0]?.message || "No regressions detected in the loaded comparison set."}</p>
                </article>
                <article className="watch-item">
                  <span>New findings</span>
                  <strong>{comparison?.findingsNew?.length || 0}</strong>
                  <p>{comparison?.findingsNew?.[0]?.message || "No new findings in the current comparison set."}</p>
                </article>
                <article className="watch-item">
                  <span>Persistent issues</span>
                  <strong>{comparison?.persistentFindingCount || 0}</strong>
                  <p>Persistent issues stay visible across runs and usually deserve scheduling or ownership cleanup.</p>
                </article>
              </div>
              <Card title="Trend history" className="compact-card">
                <TrendRail entries={historyEntries} />
              </Card>
            </>
          ) : (
            <EmptyState
              eyebrow="Why Bundles Matter"
              title="Load a bundle to unlock deeper review"
              description="K8V is built around portable bundle review so operators can compare, export, and triage recovery evidence without staying connected to the cluster."
            >
              <Card title="Bundle workflow">
                <div className="stack-list compact-stack">
                  <div className="empty-state-note">
                    <strong>Run a scan once</strong>
                    <p className="muted">The scan writes bundle and report files into the output directory you choose.</p>
                  </div>
                  <div className="empty-state-note">
                    <strong>Reopen it later</strong>
                    <p className="muted">Use Open Existing Bundle to review findings, compare drift, and refresh exports without live access.</p>
                  </div>
                  <div className="empty-state-note">
                    <strong>Share carefully</strong>
                    <p className="muted">Optional summary, runbook, CSV, and redacted outputs are available when you need to hand off results.</p>
                  </div>
                </div>
              </Card>
            </EmptyState>
          )}
        </section>
      )}
    </section>
  );
}

function machineReadinessItems(connectionAdvisor: ConnectionAdvisor | null) {
  if (!connectionAdvisor) {
    return [
      {
        label: "Local access",
        value: "Checking",
        detail: "K8V is still checking whether current login or a default kubeconfig is available here.",
        state: "neutral" as const,
      },
    ];
  }

  return [
    {
      label: "Existing access",
      value: connectionAdvisor.currentContext || (connectionAdvisor.currentLoginAvailable ? "Detected" : "Not detected"),
      detail:
        connectionAdvisor.currentLoginWarning ||
        connectionAdvisor.currentLoginDetail ||
        "No working local Kubernetes access was detected yet on this machine.",
      state: connectionAdvisor.currentLoginAvailable
        ? connectionAdvisor.currentLoginWarning
          ? ("caution" as const)
          : ("ready" as const)
        : ("missing" as const),
    },
    {
      label: "Default kubeconfig",
      value: connectionAdvisor.defaultKubeconfigPath || "Not found",
      detail:
        connectionAdvisor.defaultKubeconfigWarning ||
        connectionAdvisor.defaultKubeconfigDetail ||
        "No default kubeconfig was detected in the standard local locations.",
      state: connectionAdvisor.defaultKubeconfigAvailable
        ? connectionAdvisor.defaultKubeconfigPortable === false
          ? ("caution" as const)
          : ("ready" as const)
        : ("missing" as const),
    },
    {
      label: "kubectl CLI (optional)",
      value: connectionAdvisor.kubectlAvailable ? "Detected" : "Not detected",
      detail: connectionAdvisor.kubectlAvailable
        ? connectionAdvisor.kubectlPath || "Useful for endpoint discovery and token commands, but not required for the scan itself."
        : "Helpful for endpoint discovery and token commands, but K8V can still scan with a kubeconfig or direct API setup.",
      state: connectionAdvisor.kubectlAvailable ? ("ready" as const) : ("neutral" as const),
    },
  ];
}

export function ProjectsView(props: {
  projects: ProjectSummary[];
  busy: boolean;
  onPickBundle: () => void;
}) {
  return (
    <section className="panel">
      <SectionHeader
        eyebrow="Projects"
        title="Bundle history"
        description="Saved bundles stay available for offline review, export refreshes, and longitudinal comparison."
        actions={
          <button type="button" className="button secondary quiet" onClick={props.onPickBundle} disabled={props.busy}>
            Open Bundle
          </button>
        }
      />
      <DataTable
        caption="Known bundle directories"
        rows={props.projects as Array<Record<string, unknown>>}
        rowKey="lastScanPath"
        dense
        stickyHeader
        columns={[
          { key: "clusterName", label: "Cluster" },
          { key: "environment", label: "Environment" },
          { key: "timestampUtc", label: "Last scan" },
          { key: "score", label: "Score" },
          { key: "maturity", label: "Maturity" },
          { key: "outputDir", label: "Output dir" },
        ]}
      />
    </section>
  );
}
