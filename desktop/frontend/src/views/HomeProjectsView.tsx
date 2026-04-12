import type { ProjectSummary, Workspace } from "../lib/types";
import { DataTable, MetricCard, TrendRail } from "../components/ui";

export function HomeView(props: {
  workspace: Workspace | null;
  projects: ProjectSummary[];
  onViewProjects: () => void;
  onOpenBundle: (path?: string) => void;
}) {
  const bundle = props.workspace?.bundle;
  return (
    <section className="page-grid home-grid">
      <section className="panel hero-panel">
        <p className="eyebrow">Release-Ready Desktop</p>
        <h3>One scan engine, two surfaces, zero shell-outs.</h3>
        <p className="lead">
          The desktop app sits on the same Go service layer as the CLI, keeps report exports offline-friendly,
          and preserves the existing report information architecture.
        </p>
        <div className="metric-grid">
          <MetricCard label="Overall Score" value={bundle?.score.overall.final ?? 85} tone="accent" />
          <MetricCard label="Namespaces" value={(bundle?.inventory.namespaces || []).length} />
          <MetricCard label="Findings" value={(bundle?.inventory.findings || []).length} tone="high" />
          <MetricCard label="History Points" value={props.workspace?.history.entries.length || 3} />
        </div>
      </section>

      <section className="panel">
        <div className="section-header">
          <div>
            <p className="eyebrow">Projects</p>
            <h3>Recent assessment bundles</h3>
          </div>
          <button type="button" className="button secondary" onClick={props.onViewProjects}>
            Open Projects
          </button>
        </div>
        <div className="stack-list">
          {props.projects.map((project) => (
            <button
              type="button"
              key={project.lastScanPath}
              className="project-card"
              onClick={() => props.onOpenBundle(project.lastScanPath)}
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
          ))}
        </div>
      </section>

      <section className="panel">
        <div className="section-header">
          <div>
            <p className="eyebrow">History</p>
            <h3>Trend and compare dashboard</h3>
          </div>
        </div>
        <TrendRail entries={props.workspace?.history.entries || []} />
        <div className="compare-callout">
          <strong>
            Delta vs previous run: {(props.workspace?.history.trendDelta || 0) > 0 ? "+" : ""}
            {props.workspace?.history.trendDelta || 0}
          </strong>
          <p className="muted">
            New findings and resolved regressions stay visible in the same Compare model used by the HTML report.
          </p>
        </div>
      </section>
    </section>
  );
}

export function ProjectsView(props: { projects: ProjectSummary[]; onOpenBundle: (path?: string) => void }) {
  return (
    <section className="panel">
      <div className="section-header">
        <div>
          <p className="eyebrow">Projects</p>
          <h3>Bundle history across outputs</h3>
        </div>
        <button type="button" className="button secondary" onClick={() => props.onOpenBundle()}>
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
