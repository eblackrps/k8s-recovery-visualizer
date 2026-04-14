import type { ProjectSummary, Workspace } from "../lib/types";
import {
  Badge,
  Card,
  DataTable,
  KeyValueGrid,
  MetricCard,
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

  return (
    <section className="page-grid home-grid">
      <section className="panel home-actions-panel">
        <SectionHeader
          eyebrow="Operator Workspace"
          title="Choose the next task"
          description="Start a new assessment, reopen an existing bundle, or jump directly into the findings that affect restore readiness."
        />
        <div className="action-grid">
          <button type="button" className="action-tile action-primary" onClick={props.onStartScan}>
            <span className="eyebrow">Primary</span>
            <strong>Start New Scan</strong>
            <p>Run a fresh remote assessment against a reachable cluster.</p>
          </button>
          <button type="button" className="action-tile" onClick={props.onReviewFindings} disabled={!bundle}>
            <span className="eyebrow">Current Bundle</span>
            <strong>Review Findings</strong>
            <p>{bundle ? `${openFindings} active findings are ready for review.` : "Load a bundle to inspect current findings."}</p>
          </button>
          <button type="button" className="action-tile" onClick={props.onPickBundle} disabled={props.busy}>
            <span className="eyebrow">Offline Analysis</span>
            <strong>Open Existing Bundle</strong>
            <p>Inspect a prior assessment without requiring live cluster access.</p>
          </button>
        </div>
        <div className="panel-footer-actions">
          <button type="button" className="button secondary quiet" onClick={props.onViewProjects}>
            Browse Bundle History
          </button>
        </div>
      </section>

      <section className="panel home-posture-panel">
        <SectionHeader
          eyebrow="Current Posture"
          title={bundle ? `${bundle.metadata.clusterName} · ${bundle.metadata.environment}` : "No bundle loaded"}
          description={bundle ? "Recovery judgment from the latest loaded assessment bundle." : "Open a bundle or run a scan to populate the workspace."}
          actions={
            bundle ? (
              <Badge tone={toneForMaturity(bundle.score.maturity)}>
                {bundle.score.maturity} {bundle.score.overall.final}
              </Badge>
            ) : null
          }
        />
        <div className="home-score-strip">
          <div className="score-panel">
            <span className="eyebrow">Overall readiness</span>
            <strong>{bundle ? bundle.score.overall.final : "—"}</strong>
            <p className="muted">
              {bundle
                ? `${bundle.cluster.platform?.provider || "Unknown"} ${bundle.cluster.platform?.k8sVersion || ""}`.trim()
                : "No active assessment"}
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
        <div className="summary-two-up">
          <Card title="Bundle context">
            <KeyValueGrid
              items={[
                ["Profile", bundle?.profile || "standard"],
                ["Scope", (bundle?.scanNamespaces || []).join(", ") || "all namespaces"],
                ["Generated", bundle?.metadata.generatedAt || "Not available"],
                ["Output", props.workspace?.artifacts.outputDir || "Not available"],
              ]}
            />
          </Card>
          <Card title="Recovery signals">
            <KeyValueGrid
              items={[
                ["Backup tool", bundle?.inventory.backup?.primaryTool || "none"],
                ["Coverage", bundle?.inventory.backup?.coverageStatus || "unknown"],
                ["Restore blockers", String(restoreSim?.blockedNamespaces || 0)],
                ["Current maturity", bundle?.score.maturity || "Unknown"],
              ]}
            />
          </Card>
        </div>
      </section>

      <section className="panel home-bundles-panel">
        <SectionHeader
          eyebrow="Recent Bundles"
          title="Assessment history"
          description="Open a recent bundle directly from the workspace. The list stays compact so clusters, environments, and score changes are easy to scan."
          actions={
            <button type="button" className="button secondary quiet" onClick={props.onViewProjects}>
              Open Projects
            </button>
          }
        />
        <div className="project-list compact">
          {props.projects.length ? (
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

      <section className="panel home-watch-panel">
        <SectionHeader
          eyebrow="Operational Watch"
          title="What changed and what needs attention"
          description="The watchlist keeps regressions, new gaps, and restore-readiness movement visible without forcing you into the full report first."
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
      </section>
    </section>
  );
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
