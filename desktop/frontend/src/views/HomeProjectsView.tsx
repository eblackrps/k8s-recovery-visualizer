import type { ProjectSummary, Workspace } from "../lib/types";
import { Card, DataTable, KeyValueGrid, MetricCard, TrendRail } from "../components/ui";

export function HomeView(props: {
  workspace: Workspace | null;
  projects: ProjectSummary[];
  busy: boolean;
  onViewProjects: () => void;
  onPickBundle: () => void;
  onOpenProject: (path: string) => void;
}) {
  const bundle = props.workspace?.bundle;
  const historyEntries = props.workspace?.history.entries || [];
  const latestHistory = historyEntries[historyEntries.length - 1];
  return (
    <section className="page-grid home-grid">
      <section className="panel">
        <div className="section-header">
          <div>
            <p className="eyebrow">Dashboard</p>
            <h3>Current recovery posture</h3>
          </div>
        </div>
        <div className="metric-grid">
          <MetricCard label="Overall Score" value={bundle ? bundle.score.overall.final : "—"} tone="accent" />
          <MetricCard label="Namespaces" value={(bundle?.inventory.namespaces || []).length} />
          <MetricCard label="Findings" value={(bundle?.inventory.findings || []).length} tone="high" />
          <MetricCard label="History Points" value={props.workspace?.history.entries.length ?? 0} />
        </div>
        <div className="summary-two-up">
          <Card title="Bundle">
            <KeyValueGrid
              items={[
                ["Cluster", bundle?.metadata.clusterName || "No bundle loaded"],
                ["Environment", bundle?.metadata.environment || "Unknown"],
                ["Profile", bundle?.profile || "standard"],
                ["Provider", bundle?.cluster.platform?.provider || "Unknown"],
                ["Generated", bundle?.metadata.generatedAt || "Not available"],
              ]}
            />
          </Card>
          <Card title="Recovery Signals">
            <KeyValueGrid
              items={[
                ["Backup Tool", bundle?.inventory.backup?.primaryTool || "none"],
                ["Backup Coverage", bundle?.inventory.backup?.coverageStatus || "unknown"],
                ["Trend", props.workspace?.history.trendLabel || "No trend"],
                ["Delta", `${(props.workspace?.history.trendDelta || 0) > 0 ? "+" : ""}${props.workspace?.history.trendDelta || 0}`],
                ["Current Maturity", bundle?.score.maturity || latestHistory?.maturity || "Unknown"],
              ]}
            />
          </Card>
        </div>
      </section>

      <section className="panel">
        <div className="section-header">
          <div>
            <p className="eyebrow">Bundles</p>
            <h3>Recent assessment outputs</h3>
          </div>
          <button type="button" className="button secondary" onClick={props.onViewProjects}>
            Open Projects
          </button>
        </div>
        <div className="stack-list">
          {props.projects.length ? (
            props.projects.map((project) => (
              <button
                type="button"
                key={project.lastScanPath}
                className="project-card"
                onClick={() => props.onOpenProject(project.lastScanPath)}
                disabled={props.busy}
              >
                <div>
                  <strong>{project.clusterName || project.name}</strong>
                  <p className="muted">{project.environment || "Unknown environment"}</p>
                </div>
                <div className="project-score">
                  <strong>{project.score}</strong>
                  <span className="chip">{project.maturity}</span>
                </div>
              </button>
            ))
          ) : (
            <p className="muted">No bundles found yet. Run a scan or open an existing bundle to populate this workspace.</p>
          )}
        </div>
      </section>

      <section className="panel">
        <div className="section-header">
          <div>
            <p className="eyebrow">History</p>
            <h3>Trend and compare</h3>
          </div>
        </div>
        <TrendRail entries={props.workspace?.history.entries || []} />
        <div className="compare-callout">
          <strong>
            Delta vs previous run: {(props.workspace?.history.trendDelta || 0) > 0 ? "+" : ""}
            {props.workspace?.history.trendDelta || 0}
          </strong>
          <p className="muted">
            Compare uses the same score deltas, findings changes, and baseline model as the generated report bundle.
          </p>
        </div>
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
      <div className="section-header">
        <div>
          <p className="eyebrow">Projects</p>
          <h3>Bundle history</h3>
        </div>
        <button type="button" className="button secondary" onClick={props.onPickBundle} disabled={props.busy}>
          Open Bundle
        </button>
      </div>
      <DataTable
        caption="Known bundle directories"
        rows={props.projects}
        columns={[
          { key: "clusterName", label: "Cluster" },
          { key: "environment", label: "Environment" },
          { key: "timestampUtc", label: "Last Scan" },
          { key: "score", label: "Score" },
          { key: "maturity", label: "Maturity" },
          { key: "outputDir", label: "Output Dir" },
        ]}
      />
    </section>
  );
}
