import type { RunCompletionSummary, Workspace } from "../lib/types";
import { Card, CodeBlock, KeyValueGrid, MetricCard, RunCompletionCallout, SectionHeader } from "../components/ui";

export function CompletionView(props: {
  workspace: Workspace;
  summary: RunCompletionSummary;
  onOpenPath: (path: string, label: string) => void | Promise<void>;
  onReviewResults: () => void;
  onReviewFindings: () => void;
  onReviewCompare?: () => void;
  onStartAnotherScan: () => void;
}) {
  const bundle = props.workspace.bundle;
  const artifacts = summarizeArtifacts(props.summary);
  const findings = bundle.inventory.findings || [];

  return (
    <section className="page-grid complete-grid">
      <section className="panel">
        <SectionHeader
          eyebrow="Scan Complete"
          title="The bundle is ready for review and handoff"
          description="K8V finished the live collection, wrote the portable bundle and offline report outputs, and preserved the assessment for later reopening."
        />

        <RunCompletionCallout
          summary={props.summary}
          title="Assessment complete"
          description="Start with findings if you want blockers and ownership first, or open the output paths directly if this run needs to be handed off immediately."
          onOpenPath={props.onOpenPath}
          onReviewResults={props.onReviewResults}
          onReviewFindings={props.onReviewFindings}
          onReviewCompare={props.onReviewCompare}
          onStartAnotherScan={props.onStartAnotherScan}
        />

        <div className="summary-two-up">
          <Card title="What K8V just produced">
            <div className="artifact-list">
              <div className="artifact-row">
                <strong>Portable bundle</strong>
                <p className="muted">This is the file you reopen later for findings, compare, restore readiness, and exports.</p>
                <code className="path-chip">{props.summary.artifacts.bundleJson || props.summary.artifacts.loadedBundlePath || "Not recorded"}</code>
              </div>
              {artifacts.map((artifact) => (
                <div key={artifact.label} className="artifact-row">
                  <strong>{artifact.label}</strong>
                  <p className="muted">{artifact.description}</p>
                  <code className="path-chip">{artifact.path}</code>
                </div>
              ))}
            </div>
          </Card>

          <Card title="Recommended next steps">
            <ol className="workflow-list">
              <li>Review findings first to identify blockers, owners, and the shortest remediation path.</li>
              <li>
                {props.summary.hasComparison
                  ? "Open Compare next if you want the regression and improvement story against the loaded baseline."
                  : "Stay in Overview if you want the broad readiness summary before drilling into findings."}
              </li>
              <li>Open the output folder if this run needs to be attached to a ticket, shared, or copied elsewhere.</li>
              <li>Reopen the bundle JSON later if you want the same assessment without reconnecting to the cluster.</li>
            </ol>
          </Card>
        </div>
      </section>

      <section className="panel">
        <SectionHeader
          eyebrow="Assessment Snapshot"
          title="What this run found"
          description="A quick recap so first-time operators know what they can act on immediately and what can wait for deeper review."
        />

        <div className="metric-grid compact-grid">
          <MetricCard label="Score" value={props.summary.score ?? "n/a"} tone="accent" />
          <MetricCard label="Findings" value={props.summary.findingCount ?? findings.length} tone={(props.summary.findingCount ?? findings.length) > 0 ? "high" : "success"} />
          <MetricCard label="Comparison" value={props.summary.hasComparison ? "Loaded" : "None"} tone={props.summary.hasComparison ? "success" : "medium"} />
          <MetricCard label="Environment" value={props.summary.environment || "Unknown"} />
        </div>

        <Card title="Bundle facts" className="compact-card">
          <KeyValueGrid
            items={[
              ["Cluster", props.summary.clusterName || "Unknown"],
              ["Environment", props.summary.environment || "Unknown"],
              ["Generated", bundle.metadata.generatedAt || "Not recorded"],
              ["Output directory", <code className="path-chip">{props.summary.artifacts.outputDir || "Not recorded"}</code>],
            ]}
            className="compact-kv-grid"
          />
        </Card>

        <Card title="Reopen this exact assessment" className="compact-card">
          <p className="muted">
            The bundle is the saved assessment package. Opening it later restores the same findings, compare data,
            inventory, and exports without needing another live cluster connection.
          </p>
          <CodeBlock label="Bundle path" code={props.summary.artifacts.bundleJson || props.summary.artifacts.loadedBundlePath || "Bundle path not recorded"} />
        </Card>
      </section>
    </section>
  );
}

function summarizeArtifacts(summary: RunCompletionSummary) {
  return [
    summary.artifacts.htmlReport
      ? {
          label: "Primary report",
          description: "Operator-facing HTML report for offline review, ticket attachment, or handoff.",
          path: summary.artifacts.htmlReport,
        }
      : null,
    summary.artifacts.summaryHtml
      ? {
          label: "Summary",
          description: "Condensed output for leadership updates or quick distribution.",
          path: summary.artifacts.summaryHtml,
        }
      : null,
    summary.artifacts.runbookHtml
      ? {
          label: "Runbook",
          description: "Recovery-oriented follow-up guidance for operators.",
          path: summary.artifacts.runbookHtml,
        }
      : null,
    summary.artifacts.csvDir
      ? {
          label: "CSV exports",
          description: "Tabular exports for spreadsheets, evidence packs, or downstream analysis.",
          path: summary.artifacts.csvDir,
        }
      : null,
    summary.artifacts.redactedHtml || summary.artifacts.redactedJson
      ? {
          label: "Redacted outputs",
          description: "Lower-sensitivity copies when redaction was enabled for this run.",
          path: summary.artifacts.redactedHtml || summary.artifacts.redactedJson || "",
        }
      : null,
  ].filter(Boolean) as Array<{ label: string; description: string; path: string }>;
}
