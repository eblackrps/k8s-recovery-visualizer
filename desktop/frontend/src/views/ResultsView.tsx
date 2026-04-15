import { useEffect, useRef, useState } from "react";
import type { Bundle, Finding, RunCompletionSummary, Workspace } from "../lib/types";
import {
  Badge,
  Card,
  DataTable,
  KeyValueGrid,
  MetricCard,
  RunCompletionCallout,
  SectionHeader,
  TrendRail,
  formatDelta,
  handleRovingTabs,
  listCell,
  normalizeWorkloads,
  statusCell,
  titleCase,
  toneForDelta,
  toneForMaturity,
  toneForSeverity,
} from "../components/ui";

const primaryTabs = [
  "Overview",
  "Findings",
  "Restore Readiness",
  "Compare",
  "Inventory",
  "Remediation",
];

const inventoryTabs = ["Nodes", "Workloads", "Storage", "Networking", "Config", "Images"];

export function ResultsView(props: {
  workspace: Workspace;
  resultTab: string;
  setResultTab: (value: string) => void;
  findingFilter: string;
  setFindingFilter: (value: string) => void;
  exportNotice: string;
  completionSummary?: RunCompletionSummary | null;
  onExport: (kind: "report" | "summary" | "runbook" | "json" | "csv" | "redacted") => void;
  onOpenPath: (path: string, label: string) => void | Promise<void>;
  onDismissCompletion?: () => void;
}) {
  const bundle = props.workspace.bundle;
  const availablePrimaryTabs = bundle.comparison ? primaryTabs : primaryTabs.filter((tab) => tab !== "Compare");
  const currentTab = normalizeResultsTab(props.resultTab, Boolean(bundle.comparison));
  const activePrimary = inventoryTabs.includes(currentTab) ? "Inventory" : currentTab;
  const activeInventory = inventoryTabs.includes(currentTab) ? currentTab : inventoryTabs[0];
  const filteredFindings = (bundle.inventory.findings || []).filter((finding) =>
    props.findingFilter === "ALL" ? true : finding.severity === props.findingFilter,
  );
  const activePrimaryId = `results-panel-${activePrimary.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`;
  const activePrimaryTabId = `results-tab-${activePrimary.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`;
  const activeInventoryId = `results-inventory-panel-${activeInventory.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`;
  const activeInventoryTabId = `results-inventory-tab-${activeInventory.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`;
  const panelBodyRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (panelBodyRef.current) {
      panelBodyRef.current.scrollTop = 0;
    }
  }, [currentTab]);

  return (
    <section className="panel results-panel">
      <SectionHeader
        eyebrow="Results"
        title="Assessment workspace"
        description="Start with overview, findings, restore readiness, and compare. Inventory stays available without taking over the top level."
        actions={
          <div className="toolbar">
            <Badge tone={toneForMaturity(bundle.score.maturity)}>
              {bundle.score.maturity} {bundle.score.overall.final}
            </Badge>
            <details className="export-menu">
              <summary className="button secondary quiet">Export</summary>
              <div className="export-menu-body">
                {(["report", "summary", "runbook", "json", "csv", "redacted"] as const).map((kind) => (
                  <button key={kind} type="button" className="menu-action" onClick={() => props.onExport(kind)}>
                    Export {titleCase(kind)}
                  </button>
                ))}
              </div>
            </details>
          </div>
        }
      />

      {props.exportNotice ? <p className="notice notice-info">{props.exportNotice}</p> : null}
      {props.completionSummary ? (
        <RunCompletionCallout
          summary={props.completionSummary}
          onOpenPath={props.onOpenPath}
          onReviewFindings={() => props.setResultTab("Findings")}
          onReviewCompare={bundle.comparison ? () => props.setResultTab("Compare") : undefined}
          onDismiss={props.onDismissCompletion}
        />
      ) : null}
      {!props.completionSummary ? <ArtifactHandoffPanel workspace={props.workspace} onOpenPath={props.onOpenPath} /> : null}

      <div className="tab-row" role="tablist" aria-label="Results sections">
        {availablePrimaryTabs.map((tab, index) => {
          const selected = activePrimary === tab;
          return (
            <button
              key={tab}
              type="button"
              id={`results-tab-${tab.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`}
              role="tab"
              aria-selected={selected}
              aria-controls={`results-panel-${tab.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`}
              tabIndex={selected ? 0 : -1}
              className={`tab ${selected ? "is-active" : ""}`}
              onClick={() => props.setResultTab(tab === "Inventory" ? activeInventory : tab)}
              onKeyDown={(event) => handleRovingTabs(event, availablePrimaryTabs, index, (next) => props.setResultTab(availablePrimaryTabs[next] === "Inventory" ? activeInventory : availablePrimaryTabs[next]))}
            >
              {tab}
            </button>
          );
        })}
      </div>

      {activePrimary === "Inventory" ? (
        <div className="subtab-row" role="tablist" aria-label="Inventory sections">
          {inventoryTabs.map((tab, index) => (
            <button
              key={tab}
              type="button"
              id={`results-inventory-tab-${tab.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`}
              role="tab"
              aria-selected={activeInventory === tab}
              aria-controls={`results-inventory-panel-${tab.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`}
              tabIndex={activeInventory === tab ? 0 : -1}
              className={`subtab ${activeInventory === tab ? "is-active" : ""}`}
              onClick={() => props.setResultTab(tab)}
              onKeyDown={(event) => handleRovingTabs(event, inventoryTabs, index, (next) => props.setResultTab(inventoryTabs[next]))}
            >
              {tab}
            </button>
          ))}
        </div>
      ) : null}

      <section ref={panelBodyRef} id={activePrimaryId} role="tabpanel" aria-labelledby={activePrimaryTabId} className="results-panel-body">
        {activePrimary === "Overview" && <OverviewPanel bundle={bundle} workspace={props.workspace} />}
        {activePrimary === "Findings" && (
          <FindingsPanel
            bundle={bundle}
            filteredFindings={filteredFindings}
            findingFilter={props.findingFilter}
            setFindingFilter={props.setFindingFilter}
          />
        )}
        {activePrimary === "Restore Readiness" && <RestoreReadinessPanel bundle={bundle} />}
        {activePrimary === "Compare" && bundle.comparison ? <ComparePanel bundle={bundle} /> : null}
        {activePrimary === "Inventory" && (
          <section id={activeInventoryId} role="tabpanel" aria-labelledby={activeInventoryTabId}>
            <InventoryPanel bundle={bundle} tab={activeInventory} />
          </section>
        )}
        {activePrimary === "Remediation" && <RemediationPanel bundle={bundle} />}
      </section>
    </section>
  );
}

function ArtifactHandoffPanel(props: { workspace: Workspace; onOpenPath: (path: string, label: string) => void | Promise<void> }) {
  const artifacts = summarizeArtifacts(props.workspace);
  if (!props.workspace.artifacts.outputDir && !artifacts.length) {
    return null;
  }

  return (
    <Card
      title="Bundle and reports on disk"
      description="This assessment is already written to the output directory below. Reopen the bundle JSON later to return to the same findings, compare view, and restore-readiness analysis."
      actions={
        <div className="toolbar wrap-toolbar">
          {props.workspace.artifacts.outputDir ? (
            <button
              type="button"
              className="button secondary quiet"
              onClick={() => void props.onOpenPath(props.workspace.artifacts.outputDir || "", "output folder")}
            >
              Open output folder
            </button>
          ) : null}
          {props.workspace.artifacts.htmlReport ? (
            <button
              type="button"
              className="button secondary quiet"
              onClick={() => void props.onOpenPath(props.workspace.artifacts.htmlReport || "", "primary report")}
            >
              Open primary report
            </button>
          ) : null}
          {props.workspace.artifacts.bundleJson || props.workspace.artifacts.loadedBundlePath ? (
            <button
              type="button"
              className="button secondary quiet"
              onClick={() =>
                void props.onOpenPath(
                  props.workspace.artifacts.bundleJson || props.workspace.artifacts.loadedBundlePath || "",
                  "bundle JSON",
                )
              }
            >
              Open bundle JSON
            </button>
          ) : null}
        </div>
      }
      className="results-handoff-card"
    >
      <div className="summary-two-up results-handoff-grid">
        <KeyValueGrid
          items={[
            ["Output directory", <code className="path-chip">{props.workspace.artifacts.outputDir || "Not recorded"}</code>],
            ["Reopen later", <code className="path-chip">{props.workspace.artifacts.bundleJson || props.workspace.artifacts.loadedBundlePath || "Bundle path not recorded"}</code>],
            ["Primary report", <code className="path-chip">{props.workspace.artifacts.htmlReport || "HTML report not generated"}</code>],
            ["Bundle source", props.workspace.source || "bundle"],
          ]}
          className="compact-kv-grid"
        />
        <div className="artifact-list results-artifact-list">
          {artifacts.map((artifact) => (
            <article key={artifact.name} className="artifact-row">
              <strong>{artifact.name}</strong>
              <p className="muted">{artifact.detail}</p>
              <code className="path-chip">{artifact.path}</code>
            </article>
          ))}
        </div>
      </div>
    </Card>
  );
}

function OverviewPanel(props: { bundle: Bundle; workspace: Workspace }) {
  const bundle = props.bundle;
  const compare = bundle.comparison;
  const findings = bundle.inventory.findings || [];
  const restoreSim = bundle.inventory.backup?.restoreSim;

  return (
    <div className="results-stack">
      <div className="metric-grid">
        <MetricCard label="Overall score" value={bundle.score.overall.final} tone="accent" />
        <MetricCard label="Maturity" value={bundle.score.maturity} />
        <MetricCard label="Open findings" value={findings.length} tone={findings.length ? "high" : "success"} />
        <MetricCard
          label="Trend delta"
          value={props.workspace.history.entries.length > 1 ? formatDelta(props.workspace.history.trendDelta || 0) : "First run"}
          tone={mapDeltaTone(props.workspace.history.trendDelta || 0)}
        />
        <MetricCard label="Blocked namespaces" value={restoreSim?.blockedNamespaces || 0} tone={(restoreSim?.blockedNamespaces || 0) > 0 ? "critical" : "success"} />
        <MetricCard label="Data at risk GiB" value={restoreSim?.estimatedDataAtRiskGb || 0} tone={(restoreSim?.estimatedDataAtRiskGb || 0) > 0 ? "critical" : "success"} />
      </div>

      <div className="summary-two-up">
        <Card title="Assessment context">
          <KeyValueGrid
            items={[
              ["Cluster", bundle.metadata.clusterName || "unknown"],
              ["Environment", bundle.metadata.environment || "unknown"],
              ["Provider", bundle.cluster.platform?.provider || "unknown"],
              ["Version", bundle.cluster.platform?.k8sVersion || "unknown"],
              ["Profile", bundle.profile || "standard"],
              ["Scope", (bundle.scanNamespaces || []).join(", ") || "all namespaces"],
            ]}
          />
        </Card>
        <Card title="Operator brief">
          <div className="brief-list">
            <div className="brief-row">
              <span>Backup coverage</span>
              <strong>{bundle.inventory.backup?.coverageStatus || "unknown"}</strong>
            </div>
            <div className="brief-row">
              <span>Restore blockers</span>
              <strong>{(restoreSim?.blockingReasons || []).join(" · ") || "No blocking reasons reported"}</strong>
            </div>
            <div className="brief-row">
              <span>Comparison signal</span>
              <strong>
                {compare ? `${formatDelta(compare.scoreDelta || 0)} overall, ${compare.findingsRegressed?.length || 0} regressed` : "No comparison baseline loaded"}
              </strong>
            </div>
            <div className="brief-row">
              <span>Primary tool</span>
              <strong>{bundle.inventory.backup?.primaryTool || "none"}</strong>
            </div>
          </div>
        </Card>
      </div>

      <div className="summary-two-up">
        <Card title="Top findings">
          <div className="finding-list">
            {findings.length ? (
              findings.slice(0, 3).map((finding) => (
                <article key={finding.id || finding.resourceId} className="finding-list-item">
                  <div className="finding-list-head">
                    <Badge tone={toneForSeverity(finding.severity)}>{finding.severity}</Badge>
                    <code className="table-code">{finding.resourceId}</code>
                  </div>
                  <strong>{finding.title || finding.message}</strong>
                  <p className="muted">{finding.title ? finding.message : finding.recommendation}</p>
                </article>
              ))
            ) : (
              <p className="muted">No findings are present in the loaded bundle.</p>
            )}
          </div>
        </Card>
        <Card title="History trend">
          <TrendRail entries={props.workspace.history.entries} />
        </Card>
      </div>

      <DataTable
        caption="Domain score detail"
        dense
        stickyHeader
        rows={[
          { name: "Overall", score: bundle.score.overall.final, previous: compare?.previousScore, delta: compare?.scoreDelta || 0 },
          { name: "Storage", score: bundle.score.storage.final, previous: compare?.domainDeltas?.find((row) => row.name === "storage")?.previous, delta: compare?.domainDeltas?.find((row) => row.name === "storage")?.delta || 0 },
          { name: "Workload", score: bundle.score.workload.final, previous: compare?.domainDeltas?.find((row) => row.name === "workload")?.previous, delta: compare?.domainDeltas?.find((row) => row.name === "workload")?.delta || 0 },
          { name: "Config", score: bundle.score.config.final, previous: compare?.domainDeltas?.find((row) => row.name === "config")?.previous, delta: compare?.domainDeltas?.find((row) => row.name === "config")?.delta || 0 },
          { name: "Backup", score: bundle.score.backup.final, previous: compare?.domainDeltas?.find((row) => row.name === "backup")?.previous, delta: compare?.domainDeltas?.find((row) => row.name === "backup")?.delta || 0 },
        ] as Array<Record<string, unknown>>}
        columns={[
          { key: "name", label: "Domain" },
          { key: "score", label: "Current" },
          { key: "previous", label: "Previous" },
          { key: "delta", label: "Delta", render: (value) => <span className={`text-${toneForDelta(Number(value || 0))}`}>{formatDelta(Number(value || 0))}</span> },
        ]}
      />
    </div>
  );
}

function FindingsPanel(props: {
  bundle: Bundle;
  filteredFindings: Finding[];
  findingFilter: string;
  setFindingFilter: (value: string) => void;
}) {
  const [expandedFindingId, setExpandedFindingId] = useState<string | null>(null);
  const findings = props.bundle.inventory.findings || [];
  const counts = ["CRITICAL", "HIGH", "MEDIUM", "LOW"].map((level) => ({
    level,
    count: findings.filter((finding) => finding.severity === level).length,
  }));

  return (
    <div className="results-stack">
      <div className="findings-toolbar">
        <div className="filter-row" role="group" aria-label="Finding severity filters">
          {["ALL", "CRITICAL", "HIGH", "MEDIUM", "LOW"].map((filter) => (
            <button
              key={filter}
              type="button"
              className={`chip-button ${props.findingFilter === filter ? "is-active" : ""}`}
              onClick={() => props.setFindingFilter(filter)}
              aria-pressed={props.findingFilter === filter}
            >
              {filter}
            </button>
          ))}
        </div>
        <div className="findings-counts">
          {counts.map((item) => (
            <Badge key={item.level} tone={toneForSeverity(item.level)}>
              {item.level} {item.count}
            </Badge>
          ))}
        </div>
      </div>

      {props.filteredFindings.length ? (
        <section className="table-block findings-table">
          <div className="table-heading">
            <h4>Findings</h4>
          </div>
          <div className="table-shell sticky-head dense">
            <table>
              <caption className="sr-only">Findings</caption>
              <thead>
                <tr>
                  <th>Rank</th>
                  <th>Severity</th>
                  <th>Owner</th>
                  <th>Impact</th>
                  <th>Effort</th>
                  <th>Resource</th>
                  <th>Finding</th>
                  <th>Recommendation</th>
                  <th>Detail</th>
                </tr>
              </thead>
              <tbody>
                {props.filteredFindings.map((finding, index) => {
                  const rowId = finding.id || `${finding.resourceId}-${finding.rank || index}`;
                  const expanded = expandedFindingId === rowId;
                  const detailId = `finding-detail-${index}`;
                  return [
                      <tr key={`${rowId}-summary`}>
                        <td className="cell-tight">{finding.rank ?? "—"}</td>
                        <td className="cell-tight">
                          <Badge tone={toneForSeverity(finding.severity)}>{finding.severity}</Badge>
                        </td>
                        <td className="cell-owner">{finding.ownerHint || "—"}</td>
                        <td className="cell-impact">{finding.impact || "—"}</td>
                        <td className="cell-tight">
                          <Badge>{finding.effort || "—"}</Badge>
                        </td>
                        <td className="cell-code">
                          <code className="table-code">{finding.resourceId}</code>
                        </td>
                        <td className="cell-wide">
                          <div className="finding-cell">
                            {finding.title ? <strong>{finding.title}</strong> : null}
                            <p>{finding.message}</p>
                          </div>
                        </td>
                        <td className="cell-wide">
                          <p className="recommendation-preview">{finding.recommendation}</p>
                        </td>
                        <td className="cell-tight">
                          <button
                            type="button"
                            className="button secondary quiet detail-toggle"
                            aria-expanded={expanded}
                            aria-controls={detailId}
                            onClick={() => setExpandedFindingId(expanded ? null : rowId)}
                          >
                            {expanded ? "Hide" : "View"}
                          </button>
                        </td>
                      </tr>,
                      expanded ? (
                        <tr key={`${rowId}-detail`} className="detail-row">
                          <td colSpan={9} id={detailId}>
                            <div className="finding-detail-panel">
                              <div className="finding-detail-grid">
                                <div>
                                  <span className="eyebrow">Recommendation</span>
                                  <p>{finding.recommendation}</p>
                                </div>
                                <div>
                                  <span className="eyebrow">Ownership</span>
                                  <p>{finding.ownerHint || "Not specified"}</p>
                                </div>
                                <div>
                                  <span className="eyebrow">Impact</span>
                                  <p>{finding.impact || "Not specified"}</p>
                                </div>
                                <div>
                                  <span className="eyebrow">Effort</span>
                                  <p>{finding.effort || "Not specified"}</p>
                                </div>
                              </div>
                            </div>
                          </td>
                        </tr>
                      ) : null,
                  ];
                })}
              </tbody>
            </table>
          </div>
        </section>
      ) : (
        <p className="muted">No findings match the current filter.</p>
      )}
    </div>
  );
}

function RestoreReadinessPanel(props: { bundle: Bundle }) {
  const backup = props.bundle.inventory.backup;
  const restoreSim = backup?.restoreSim;

  return (
    <div className="results-stack">
      <div className="metric-grid">
        <MetricCard label="Tool" value={backup?.primaryTool || "none"} />
        <MetricCard label="Coverage" value={backup?.coverageStatus || "unknown"} tone="accent" />
        <MetricCard label="Assurance" value={backup?.assurance?.conclusion || "unknown"} tone="success" />
        <MetricCard label="Offsite" value={backup?.hasOffsite ? "yes" : "no"} tone={backup?.hasOffsite ? "success" : "high"} />
        <MetricCard label="Ready namespaces" value={restoreSim?.readyNamespaces || 0} tone="success" />
        <MetricCard label="Blocked namespaces" value={restoreSim?.blockedNamespaces || 0} tone={(restoreSim?.blockedNamespaces || 0) > 0 ? "critical" : "success"} />
      </div>

      <div className="summary-two-up">
        <Card title="Backup posture">
          <KeyValueGrid
            items={[
              ["Coverage", backup?.coverageStatus || "unknown"],
              ["Coverage reason", backup?.coverageReason || "not reported"],
              ["Covered PVC GiB", String(restoreSim?.coveredPvcsGb || 0)],
              ["Data at risk GiB", String(restoreSim?.estimatedDataAtRiskGb || 0)],
              ["Offsite verified", backup?.hasOffsite ? "yes" : "no"],
            ]}
          />
        </Card>
        <Card title="Restore snapshot">
          <KeyValueGrid
            items={[
              ["Ready namespaces", String(restoreSim?.readyNamespaces || 0)],
              ["Warning namespaces", String(restoreSim?.warningNamespaces || 0)],
              ["Blocked namespaces", String(restoreSim?.blockedNamespaces || 0)],
              ["Unknown namespaces", String(restoreSim?.unknownNamespaces || 0)],
              ["Top blockers", (restoreSim?.blockingReasons || []).join(" · ") || "No blockers reported"],
            ]}
          />
        </Card>
      </div>

      <DataTable
        caption="Backup policies"
        rows={(backup?.policies || []) as Array<Record<string, unknown>>}
        dense
        stickyHeader
        columns={[
          { key: "tool", label: "Tool" },
          { key: "name", label: "Policy" },
          { key: "schedule", label: "Schedule" },
          { key: "rpoHours", label: "RPO (h)" },
          { key: "lastSuccessAt", label: "Last success" },
        ]}
      />

      <DataTable
        caption="Restore simulation"
        rows={(restoreSim?.namespaces || []) as Array<Record<string, unknown>>}
        dense
        stickyHeader
        columns={[
          { key: "namespace", label: "Namespace" },
          { key: "readiness", label: "Readiness", render: (value) => <Badge tone={readinessTone(String(value || ""))}>{titleCase(String(value || "unknown"))}</Badge> },
          { key: "hasCoverage", label: "Covered", render: (value) => statusCell(Boolean(value)) },
          { key: "rpoHours", label: "RPO (h)" },
          { key: "pvcSizeGb", label: "PVC GiB" },
          { key: "blockers", label: "Blockers", render: (value) => listCell(value) },
          { key: "warnings", label: "Warnings", render: (value) => listCell(value) },
        ]}
      />

      <DataTable
        caption="Restore drill plan"
        rows={(backup?.drillPlan || []) as Array<Record<string, unknown>>}
        dense
        stickyHeader
        columns={[
          { key: "phase", label: "Phase", render: (value) => titleCase(String(value || "")) },
          { key: "title", label: "Step" },
          { key: "ownerHint", label: "Owner" },
          { key: "detail", label: "Detail", className: "cell-wide" },
          { key: "validation", label: "Validate", render: (value) => listCell(value), className: "cell-wide" },
        ]}
      />
    </div>
  );
}

function ComparePanel(props: { bundle: Bundle }) {
  const compare = props.bundle.comparison;

  return (
    <div className="results-stack">
      <div className="metric-grid">
        <MetricCard label="Previous score" value={compare?.previousScore || "n/a"} />
        <MetricCard label="Delta" value={formatDelta(compare?.scoreDelta || 0)} tone={mapDeltaTone(compare?.scoreDelta || 0)} />
        <MetricCard label="Regressed" value={(compare?.findingsRegressed || []).length} tone={(compare?.findingsRegressed || []).length ? "critical" : "success"} />
        <MetricCard label="Resolved" value={(compare?.findingsResolved || []).length} tone="success" />
        <MetricCard label="New findings" value={(compare?.findingsNew || []).length} tone="high" />
        <MetricCard label="Persistent" value={compare?.persistentFindingCount || 0} />
      </div>

      <div className="summary-two-up">
        <Card title="Change summary">
          <KeyValueGrid
            items={[
              ["Previous scan", compare?.previousScannedAt || "n/a"],
              ["Maturity", `${compare?.previousMaturity || "n/a"} -> ${compare?.currentMaturity || "n/a"}`],
              ["Backup tool", `${compare?.backupToolPrevious || "none"} -> ${compare?.backupToolCurrent || "none"}`],
              ["Namespaces added", (compare?.namespacesAdded || []).join(", ") || "none"],
              ["Workloads added", (compare?.workloadsAdded || []).join(", ") || "none"],
            ]}
          />
        </Card>
        <Card title="Delta story">
          <div className="brief-list">
            <div className="brief-row">
              <span>What got worse</span>
              <strong>{compare?.findingsRegressed?.[0]?.message || "No regressed findings recorded."}</strong>
            </div>
            <div className="brief-row">
              <span>What improved</span>
              <strong>{compare?.findingsImproved?.[0]?.message || "No improved findings recorded."}</strong>
            </div>
            <div className="brief-row">
              <span>What stayed risky</span>
              <strong>{compare?.persistentFindingCount || 0} persistent findings remain open.</strong>
            </div>
          </div>
        </Card>
      </div>

      <div className="summary-two-up">
        <DataTable
          caption="Domain score drift"
          rows={(compare?.domainDeltas || []) as Array<Record<string, unknown>>}
          dense
          stickyHeader
          columns={[
            { key: "name", label: "Domain", render: (value) => titleCase(String(value || "")) },
            { key: "previous", label: "Previous" },
            { key: "current", label: "Current" },
            { key: "delta", label: "Delta", render: (value) => <span className={`text-${toneForDelta(Number(value || 0))}`}>{formatDelta(Number(value || 0))}</span> },
          ]}
        />
        <DataTable
          caption="Severity drift"
          rows={(compare?.severityDeltas || []) as Array<Record<string, unknown>>}
          dense
          stickyHeader
          columns={[
            { key: "severity", label: "Severity", render: (value) => <Badge tone={toneForSeverity(String(value || ""))}>{String(value || "unknown")}</Badge> },
            { key: "previous", label: "Previous" },
            { key: "current", label: "Current" },
            { key: "delta", label: "Delta", render: (value) => <span className={`text-${toneForDelta(-1 * Number(value || 0))}`}>{formatDelta(Number(value || 0))}</span> },
          ]}
        />
      </div>

      <DataTable
        caption="What got worse"
        rows={(compare?.findingsRegressed || []) as Array<Record<string, unknown>>}
        rowKey="id"
        dense
        stickyHeader
        columns={[
          { key: "previousSeverity", label: "Was", render: (value) => <Badge tone={toneForSeverity(String(value || ""))}>{String(value || "unknown")}</Badge> },
          { key: "currentSeverity", label: "Now", render: (value) => <Badge tone={toneForSeverity(String(value || ""))}>{String(value || "unknown")}</Badge> },
          { key: "change", label: "Change" },
          { key: "resourceId", label: "Resource", render: (value) => <code className="table-code">{String(value || "—")}</code> },
          { key: "message", label: "Message", className: "cell-wide" },
          { key: "ownerHint", label: "Owner" },
        ]}
      />

      <div className="summary-two-up">
        <DataTable
          caption="What improved"
          rows={(compare?.findingsImproved || []) as Array<Record<string, unknown>>}
          rowKey="id"
          dense
          stickyHeader
          columns={[
            { key: "previousSeverity", label: "Was", render: (value) => <Badge tone={toneForSeverity(String(value || ""))}>{String(value || "unknown")}</Badge> },
            { key: "currentSeverity", label: "Now", render: (value) => <Badge tone={toneForSeverity(String(value || ""))}>{String(value || "unknown")}</Badge> },
            { key: "change", label: "Change" },
            { key: "resourceId", label: "Resource", render: (value) => <code className="table-code">{String(value || "—")}</code> },
            { key: "message", label: "Message", className: "cell-wide" },
          ]}
        />
        <DataTable
          caption="Inventory drift"
          rows={(compare?.inventoryDeltas || []) as Array<Record<string, unknown>>}
          dense
          stickyHeader
          columns={[
            { key: "name", label: "Area", render: (value) => titleCase(String(value || "")) },
            { key: "added", label: "Added" },
            { key: "removed", label: "Removed" },
          ]}
        />
      </div>

      <DataTable
        caption="New findings"
        rows={(compare?.findingsNew || []) as Array<Record<string, unknown>>}
        rowKey="id"
        dense
        stickyHeader
        columns={[
          { key: "severity", label: "Severity", render: (value) => <Badge tone={toneForSeverity(String(value || ""))}>{String(value || "unknown")}</Badge> },
          { key: "resourceId", label: "Resource", render: (value) => <code className="table-code">{String(value || "—")}</code> },
          { key: "message", label: "Message", className: "cell-wide" },
        ]}
      />
    </div>
  );
}

function InventoryPanel(props: { bundle: Bundle; tab: string }) {
  if (props.tab === "Nodes") {
    return (
      <DataTable
        caption="Cluster nodes"
        rows={(props.bundle.inventory.nodes || []) as Array<Record<string, unknown>>}
        dense
        stickyHeader
        columns={[
          { key: "name", label: "Name" },
          { key: "roles", label: "Roles", render: (value) => listCell(value) },
          { key: "ready", label: "Ready", render: (value) => statusCell(Boolean(value)) },
          { key: "zone", label: "Zone" },
          { key: "osImage", label: "OS" },
          { key: "kubeletVersion", label: "Kubelet" },
        ]}
      />
    );
  }

  if (props.tab === "Workloads") {
    return (
      <DataTable
        caption="Workload inventory"
        rows={normalizeWorkloads(props.bundle)}
        dense
        stickyHeader
        columns={[
          { key: "kind", label: "Kind" },
          { key: "namespace", label: "Namespace" },
          { key: "name", label: "Name" },
          { key: "summary", label: "Status" },
          { key: "images", label: "Images", render: (value) => listCell(value), className: "cell-wide" },
        ]}
      />
    );
  }

  if (props.tab === "Storage") {
    return (
      <div className="results-stack">
        <DataTable
          caption="PVC inventory"
          rows={(props.bundle.inventory.pvcs || []) as Array<Record<string, unknown>>}
          dense
          stickyHeader
          columns={[
            { key: "namespace", label: "Namespace" },
            { key: "name", label: "PVC" },
            { key: "storageClass", label: "Storage class" },
            { key: "requestedSize", label: "Requested" },
          ]}
        />
        <DataTable
          caption="PV inventory"
          rows={(props.bundle.inventory.pvs || []) as Array<Record<string, unknown>>}
          dense
          stickyHeader
          columns={[
            { key: "name", label: "PV" },
            { key: "claimRef", label: "Claim ref" },
            { key: "capacity", label: "Capacity" },
            { key: "reclaimPolicy", label: "Reclaim" },
            { key: "backend", label: "Backend" },
          ]}
        />
      </div>
    );
  }

  if (props.tab === "Networking") {
    return (
      <div className="results-stack">
        <DataTable
          caption="Services"
          rows={(props.bundle.inventory.services || []) as Array<Record<string, unknown>>}
          dense
          stickyHeader
          columns={[
            { key: "namespace", label: "Namespace" },
            { key: "name", label: "Name" },
            { key: "type", label: "Type" },
            { key: "externalIp", label: "External" },
          ]}
        />
        <DataTable
          caption="Network policies"
          rows={(props.bundle.inventory.networkPolicies || []) as Array<Record<string, unknown>>}
          dense
          stickyHeader
          columns={[
            { key: "namespace", label: "Namespace" },
            { key: "name", label: "Name" },
            { key: "hasIngress", label: "Ingress", render: (value) => statusCell(Boolean(value)) },
            { key: "hasEgress", label: "Egress", render: (value) => statusCell(Boolean(value)) },
          ]}
        />
      </div>
    );
  }

  if (props.tab === "Config") {
    return (
      <div className="results-stack">
        <DataTable
          caption="Config maps"
          rows={(props.bundle.inventory.configMaps || []) as Array<Record<string, unknown>>}
          dense
          stickyHeader
          columns={[
            { key: "namespace", label: "Namespace" },
            { key: "name", label: "Name" },
            { key: "keyCount", label: "Keys" },
          ]}
        />
        <DataTable
          caption="Cluster roles"
          rows={(props.bundle.inventory.clusterRoles || []) as Array<Record<string, unknown>>}
          dense
          stickyHeader
          columns={[
            { key: "name", label: "Role" },
            { key: "ruleCount", label: "Rules" },
            { key: "hasWildcardVerb", label: "Wildcard", render: (value) => statusCell(Boolean(value)) },
            { key: "hasSecretAccess", label: "Secret access", render: (value) => statusCell(Boolean(value)) },
          ]}
        />
      </div>
    );
  }

  return (
    <DataTable
      caption="Image inventory"
      rows={(props.bundle.inventory.images || []) as Array<Record<string, unknown>>}
      dense
      stickyHeader
      columns={[
        { key: "image", label: "Image", className: "cell-wide" },
        { key: "registry", label: "Registry" },
        { key: "isPublic", label: "Public", render: (value) => statusCell(Boolean(value)) },
        { key: "workloads", label: "Workloads", render: (value) => listCell(value), className: "cell-wide" },
      ]}
    />
  );
}

function RemediationPanel(props: { bundle: Bundle }) {
  const steps = props.bundle.inventory.remediationSteps || [];

  if (!steps.length) {
    return <p className="muted">No remediation steps were generated for this bundle.</p>;
  }

  return (
    <div className="results-stack">
      {steps.map((step, index) => (
        <details key={`${step.title}-${index}`} className="remediation-step">
          <summary className="remediation-summary">
            <div className="remediation-head">
              <div>
                <p className="eyebrow">Priority {step.priority}</p>
                <h4>{step.title}</h4>
              </div>
              <div className="toolbar">
                {step.ownerHint ? <Badge>{step.ownerHint}</Badge> : null}
                {step.effort ? <Badge>{`Effort ${step.effort}`}</Badge> : null}
                <Badge>{step.category}</Badge>
              </div>
            </div>
          </summary>
          <div className="remediation-body">
            <p>{step.detail}</p>
            {step.whyItMatters ? <p className="muted">Why it matters: {step.whyItMatters}</p> : null}
            {step.validation?.length ? <p className="muted">Validate: {step.validation.join(" · ")}</p> : null}
            {step.fixSteps?.length ? <p className="muted">Fix: {step.fixSteps.join(" · ")}</p> : null}
            {step.commands?.length ? <code className="mono-block">{step.commands.join("\n")}</code> : null}
          </div>
        </details>
      ))}
    </div>
  );
}

function summarizeArtifacts(workspace: Workspace) {
  const artifacts = workspace.artifacts;
  return [
    artifacts.bundleJson
      ? {
          name: "Bundle JSON",
          detail: "Primary portable assessment bundle for reopening or comparing later.",
          path: artifacts.bundleJson,
        }
      : null,
    artifacts.htmlReport
      ? {
          name: "HTML report",
          detail: "Operator-facing report for offline review and handoff.",
          path: artifacts.htmlReport,
        }
      : null,
    artifacts.summaryHtml
      ? {
          name: "Summary report",
          detail: "Short-form output for quick leadership or ticket updates.",
          path: artifacts.summaryHtml,
        }
      : null,
    artifacts.runbookHtml
      ? {
          name: "Runbook",
          detail: "Recovery-oriented checklist and follow-up guidance.",
          path: artifacts.runbookHtml,
        }
      : null,
    artifacts.csvDir
      ? {
          name: "CSV exports",
          detail: "Tabular exports for spreadsheets, evidence packs, or downstream analysis.",
          path: artifacts.csvDir,
        }
      : null,
    artifacts.redactedHtml || artifacts.redactedJson
      ? {
          name: "Redacted outputs",
          detail: "Shareable lower-sensitivity copies when redaction was enabled.",
          path: artifacts.redactedHtml || artifacts.redactedJson || "",
        }
      : null,
  ].filter(Boolean) as Array<{ name: string; detail: string; path: string }>;
}

function normalizeResultsTab(tab: string, hasComparison: boolean) {
  switch (tab) {
    case "Summary":
      return "Overview";
    case "Backup":
      return "Restore Readiness";
    case "DR Score":
      return "Overview";
    case "Nodes":
    case "Workloads":
    case "Storage":
    case "Networking":
    case "Config":
    case "Images":
    case "Findings":
    case "Remediation":
    case "Inventory":
    case "Overview":
    case "Restore Readiness":
      return tab;
    case "Compare":
      return hasComparison ? tab : "Overview";
    default:
      return "Overview";
  }
}

function mapDeltaTone(value: number) {
  const tone = toneForDelta(value);
  if (tone === "success") {
    return "success";
  }
  if (tone === "critical") {
    return "critical";
  }
  return "medium";
}

function readinessTone(value: string) {
  switch (value.toLowerCase()) {
    case "ready":
      return "success";
    case "blocked":
      return "critical";
    case "warning":
      return "warn";
    default:
      return "neutral";
  }
}
