import type { RunCompletionSummary, Workspace } from "../lib/types";
import { Card, CodeBlock, RunCompletionCallout, SectionHeader } from "../components/ui";

export function CompletionView(props: {
  workspace: Workspace;
  summary: RunCompletionSummary;
  onOpenPath: (path: string, label: string) => void | Promise<void>;
  onReviewResults: () => void;
  onReviewFindings: () => void;
  onReviewCompare?: () => void;
  onStartAnotherScan: () => void;
}) {
  const artifacts = summarizeArtifacts(props.summary);

  return (
    <section className="panel completion-panel">
      <SectionHeader
        eyebrow="Scan Complete"
        title="Assessment ready"
        description="The live collection finished, the portable bundle was written, and the main report outputs are ready for review or handoff."
      />

      <RunCompletionCallout
        summary={props.summary}
        title="Assessment complete"
        description="Start with findings for blockers and ownership first, or open the output folder if this run needs to be handed off right away."
        onOpenPath={props.onOpenPath}
        onReviewResults={props.onReviewResults}
        onReviewFindings={props.onReviewFindings}
        onReviewCompare={props.onReviewCompare}
        onStartAnotherScan={props.onStartAnotherScan}
      />

      <div className="summary-two-up completion-followup-grid">
        <Card title="What K8V just produced">
          <div className="artifact-list">
            <div className="artifact-row">
              <strong>Portable bundle</strong>
              <p className="muted">Reopen this later to get the same findings, compare data, restore readiness view, and exports without reconnecting to the cluster.</p>
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
            <li>Review findings first to identify blockers, likely owners, and the shortest remediation path.</li>
            <li>
              {props.summary.hasComparison
                ? "Open Compare next if you want the regression and improvement story against the loaded baseline."
                : "Stay in Overview if you want the broad readiness summary before drilling into findings."}
            </li>
            <li>Open the output folder if this run needs to be attached to a ticket, shared, or copied elsewhere.</li>
          </ol>
          <CodeBlock
            label="Reopen this exact assessment later"
            code={props.summary.artifacts.bundleJson || props.summary.artifacts.loadedBundlePath || "Bundle path not recorded"}
          />
        </Card>
      </div>
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
