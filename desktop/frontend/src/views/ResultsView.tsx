import type { Bundle, Workspace } from "../lib/types";
import {
  Card,
  DataTable,
  KeyValueGrid,
  MetricCard,
  TrendRail,
  handleRovingTabs,
  listCell,
  normalizeWorkloads,
  statusCell,
  titleCase,
} from "../components/ui";

const resultsTabs = [
  "Summary",
  "Nodes",
  "Workloads",
  "Storage",
  "Networking",
  "Config",
  "Images",
  "Backup",
  "DR Score",
  "Findings",
  "Remediation",
  "Compare",
];

export function ResultsView(props: {
  workspace: Workspace;
  resultTab: string;
  setResultTab: (value: string) => void;
  findingFilter: string;
  setFindingFilter: (value: string) => void;
  exportNotice: string;
  onExport: (kind: "report" | "summary" | "runbook" | "json" | "csv" | "redacted") => void;
}) {
  const bundle = props.workspace.bundle;
  const activeTabs = bundle.comparison ? resultsTabs : resultsTabs.filter((tab) => tab !== "Compare");
  const filteredFindings = (bundle.inventory.findings || []).filter((finding) =>
    props.findingFilter === "ALL" ? true : finding.severity === props.findingFilter,
  );
  const activePanelId = `results-panel-${props.resultTab.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`;
  const activeTabId = `results-tab-${props.resultTab.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`;

  return (
    <section className="panel results-panel">
      <div className="section-header">
        <div>
          <p className="eyebrow">Results Workspace</p>
          <h3>
            {bundle.metadata.clusterName || "Loaded bundle"} · {bundle.score.maturity} · {bundle.score.overall.final}
          </h3>
        </div>
        <div className="toolbar">
          {(["report", "summary", "runbook", "json", "csv", "redacted"] as const).map((kind) => (
            <button key={kind} type="button" className="button secondary" onClick={() => props.onExport(kind)}>
              Export {titleCase(kind)}
            </button>
          ))}
        </div>
      </div>
      {props.exportNotice ? <p className="notice">{props.exportNotice}</p> : null}
      <div className="tab-row" role="tablist" aria-label="Results sections">
        {activeTabs.map((tab, index) => (
          <button
            key={tab}
            type="button"
            id={`results-tab-${tab.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`}
            role="tab"
            aria-selected={props.resultTab === tab}
            aria-controls={`results-panel-${tab.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`}
            tabIndex={props.resultTab === tab ? 0 : -1}
            className={`tab ${props.resultTab === tab ? "is-active" : ""}`}
            onClick={() => props.setResultTab(tab)}
            onKeyDown={(event) => handleRovingTabs(event, activeTabs, index, (next) => props.setResultTab(activeTabs[next]))}
          >
            {tab}
          </button>
        ))}
      </div>

      <section id={activePanelId} role="tabpanel" aria-labelledby={activeTabId} className="results-panel-body">
        {props.resultTab === "Summary" && <SummaryPanel bundle={bundle} workspace={props.workspace} />}
        {props.resultTab === "Nodes" && <NodesPanel bundle={bundle} />}
        {props.resultTab === "Workloads" && <WorkloadsPanel bundle={bundle} />}
        {props.resultTab === "Storage" && <StoragePanel bundle={bundle} />}
        {props.resultTab === "Networking" && <NetworkingPanel bundle={bundle} />}
        {props.resultTab === "Config" && <ConfigPanel bundle={bundle} />}
        {props.resultTab === "Images" && <ImagesPanel bundle={bundle} />}
        {props.resultTab === "Backup" && <BackupPanel bundle={bundle} />}
        {props.resultTab === "DR Score" && <ScorePanel bundle={bundle} />}
        {props.resultTab === "Findings" && (
          <div className="results-stack">
            <div className="filter-row" role="group" aria-label="Finding severity filters">
              {["ALL", "CRITICAL", "HIGH", "MEDIUM", "LOW"].map((filter) => (
                <button key={filter} type="button" className={`chip-button ${props.findingFilter === filter ? "is-active" : ""}`} onClick={() => props.setFindingFilter(filter)} aria-pressed={props.findingFilter === filter}>
                  {filter}
                </button>
              ))}
            </div>
            <DataTable
              caption="Findings"
              rows={filteredFindings as Array<Record<string, unknown>>}
              columns={[
                { key: "rank", label: "Rank" },
                { key: "severity", label: "Severity" },
                { key: "ownerHint", label: "Owner" },
                { key: "impact", label: "Impact" },
                { key: "effort", label: "Effort" },
                { key: "resourceId", label: "Resource" },
                { key: "message", label: "Finding" },
                { key: "recommendation", label: "Recommendation" },
              ]}
            />
          </div>
        )}
        {props.resultTab === "Remediation" && (
          <div className="stack-list">
            {(bundle.inventory.remediationSteps || []).length ? (
              (bundle.inventory.remediationSteps || []).map((step, index) => (
                <article key={`${step.title}-${index}`} className="remediation-card">
                  <div className="section-header">
                    <div>
                      <p className="eyebrow">Priority {step.priority}</p>
                      <h4>{step.title}</h4>
                    </div>
                    <div className="toolbar">
                      {step.ownerHint ? <span className="chip">{step.ownerHint}</span> : null}
                      {step.effort ? <span className="chip">Effort {step.effort}</span> : null}
                      <span className="chip">{step.category}</span>
                    </div>
                  </div>
                  <p>{step.detail}</p>
                  {step.whyItMatters ? <p className="muted">Why it matters: {step.whyItMatters}</p> : null}
                  {step.validation?.length ? <p className="muted">Validate: {step.validation.join(" · ")}</p> : null}
                  {step.fixSteps?.length ? <p className="muted">Fix: {step.fixSteps.join(" · ")}</p> : null}
                  {step.commands?.length ? <code className="mono-block">{step.commands.join("\n")}</code> : null}
                </article>
              ))
            ) : (
              <p className="muted">No remediation steps were generated for this bundle.</p>
            )}
          </div>
        )}
        {props.resultTab === "Compare" && bundle.comparison ? <ComparePanel bundle={bundle} /> : null}
      </section>
    </section>
  );
}

function SummaryPanel(props: { bundle: Bundle; workspace: Workspace }) {
  const bundle = props.bundle;
  return (
    <div className="results-stack">
      <div className="metric-grid">
        <MetricCard label="Overall" value={bundle.score.overall.final} tone="accent" />
        <MetricCard label="Storage" value={bundle.score.storage.final} />
        <MetricCard label="Workload" value={bundle.score.workload.final} />
        <MetricCard label="Config" value={bundle.score.config.final} />
        <MetricCard label="Backup" value={bundle.score.backup.final} />
        <MetricCard label="Findings" value={(bundle.inventory.findings || []).length} tone="high" />
      </div>
      <div className="summary-two-up">
        <Card title="Environment">
          <KeyValueGrid items={[["Provider", bundle.cluster.platform?.provider || "unknown"], ["Version", bundle.cluster.platform?.k8sVersion || "unknown"], ["Profile", bundle.profile || "standard"], ["Target", bundle.target || "vm"], ["Scope", (bundle.scanNamespaces || []).join(", ") || "all namespaces"], ["Generated", bundle.metadata.generatedAt || "unknown"]]} />
        </Card>
        <Card title="Backup Trust">
          <KeyValueGrid items={[["Primary Tool", bundle.inventory.backup?.primaryTool || "none"], ["Coverage", bundle.inventory.backup?.coverageStatus || "unknown"], ["Assurance", bundle.inventory.backup?.assurance?.conclusion || "not assessed"], ["Offsite", bundle.inventory.backup?.hasOffsite ? "verified" : "missing"]]} />
        </Card>
      </div>
      <div className="summary-two-up">
        <Card title="History Summary">
          <KeyValueGrid items={[["Runs", String(props.workspace.history.runCount || props.workspace.history.entries.length || 0)], ["Average", String(props.workspace.history.averageScore || "n/a")], ["Best", String(props.workspace.history.bestScore || "n/a")], ["Worst", String(props.workspace.history.worstScore || "n/a")], ["Recent Trend", `${props.workspace.history.trendLabel || "FIRST_RUN"}${props.workspace.history.trendDelta ? ` (${props.workspace.history.trendDelta > 0 ? "+" : ""}${props.workspace.history.trendDelta})` : ""}`]]} />
        </Card>
        <Card title="Domain Trends">
          <DataTable
            caption="Domain trends"
            rows={(props.workspace.history.domainTrends || []) as Array<Record<string, unknown>>}
            columns={[
              { key: "name", label: "Domain", render: (value) => titleCase(String(value || "")) },
              { key: "current", label: "Current" },
              { key: "delta", label: "Delta", render: (value) => formatDelta(Number(value || 0)) },
              { key: "direction", label: "Direction", render: (value) => titleCase(String(value || "same")) },
            ]}
          />
        </Card>
      </div>
      <Card title="History">
        <TrendRail entries={props.workspace.history.entries} />
      </Card>
    </div>
  );
}

function NodesPanel(props: { bundle: Bundle }) {
  return (
    <DataTable caption="Cluster nodes" rows={(props.bundle.inventory.nodes || []) as Array<Record<string, unknown>>} columns={[{ key: "name", label: "Name" }, { key: "roles", label: "Roles", render: (value) => listCell(value) }, { key: "ready", label: "Ready", render: (value) => statusCell(Boolean(value)) }, { key: "zone", label: "Zone" }, { key: "osImage", label: "OS" }, { key: "kubeletVersion", label: "Kubelet" }]} />
  );
}

function WorkloadsPanel(props: { bundle: Bundle }) {
  return (
    <DataTable caption="Workload inventory" rows={normalizeWorkloads(props.bundle)} columns={[{ key: "kind", label: "Kind" }, { key: "namespace", label: "Namespace" }, { key: "name", label: "Name" }, { key: "summary", label: "Status" }, { key: "images", label: "Images", render: (value) => listCell(value) }]} />
  );
}

function StoragePanel(props: { bundle: Bundle }) {
  return (
    <div className="results-stack">
      <DataTable caption="PVC inventory" rows={(props.bundle.inventory.pvcs || []) as Array<Record<string, unknown>>} columns={[{ key: "namespace", label: "Namespace" }, { key: "name", label: "PVC" }, { key: "storageClass", label: "Storage Class" }, { key: "requestedSize", label: "Requested" }]} />
      <DataTable caption="PV inventory" rows={(props.bundle.inventory.pvs || []) as Array<Record<string, unknown>>} columns={[{ key: "name", label: "PV" }, { key: "claimRef", label: "Claim Ref" }, { key: "capacity", label: "Capacity" }, { key: "reclaimPolicy", label: "Reclaim" }, { key: "backend", label: "Backend" }]} />
    </div>
  );
}

function NetworkingPanel(props: { bundle: Bundle }) {
  return (
    <div className="results-stack">
      <DataTable caption="Services" rows={(props.bundle.inventory.services || []) as Array<Record<string, unknown>>} columns={[{ key: "namespace", label: "Namespace" }, { key: "name", label: "Name" }, { key: "type", label: "Type" }, { key: "externalIp", label: "External" }]} />
      <DataTable caption="Network policies" rows={(props.bundle.inventory.networkPolicies || []) as Array<Record<string, unknown>>} columns={[{ key: "namespace", label: "Namespace" }, { key: "name", label: "Name" }, { key: "hasIngress", label: "Ingress", render: (value) => statusCell(Boolean(value)) }, { key: "hasEgress", label: "Egress", render: (value) => statusCell(Boolean(value)) }]} />
    </div>
  );
}

function ConfigPanel(props: { bundle: Bundle }) {
  return (
    <div className="results-stack">
      <DataTable caption="Config maps" rows={(props.bundle.inventory.configMaps || []) as Array<Record<string, unknown>>} columns={[{ key: "namespace", label: "Namespace" }, { key: "name", label: "Name" }, { key: "keyCount", label: "Keys" }]} />
      <DataTable caption="Cluster roles" rows={(props.bundle.inventory.clusterRoles || []) as Array<Record<string, unknown>>} columns={[{ key: "name", label: "Role" }, { key: "ruleCount", label: "Rules" }, { key: "hasWildcardVerb", label: "Wildcard", render: (value) => statusCell(Boolean(value)) }, { key: "hasSecretAccess", label: "Secret Access", render: (value) => statusCell(Boolean(value)) }]} />
    </div>
  );
}

function ImagesPanel(props: { bundle: Bundle }) {
  return (
    <DataTable caption="Image inventory" rows={(props.bundle.inventory.images || []) as Array<Record<string, unknown>>} columns={[{ key: "image", label: "Image" }, { key: "registry", label: "Registry" }, { key: "isPublic", label: "Public", render: (value) => statusCell(Boolean(value)) }, { key: "workloads", label: "Workloads", render: (value) => listCell(value) }]} />
  );
}

function BackupPanel(props: { bundle: Bundle }) {
  const backup = props.bundle.inventory.backup;
  return (
    <div className="results-stack">
      <div className="metric-grid">
        <MetricCard label="Tool" value={backup?.primaryTool || "none"} />
        <MetricCard label="Coverage" value={backup?.coverageStatus || "unknown"} tone="accent" />
        <MetricCard label="Assurance" value={backup?.assurance?.conclusion || "unknown"} tone="success" />
        <MetricCard label="Offsite" value={backup?.hasOffsite ? "yes" : "no"} tone={backup?.hasOffsite ? "success" : "high"} />
        <MetricCard label="Ready NS" value={backup?.restoreSim?.readyNamespaces || 0} tone="success" />
        <MetricCard label="Blocked NS" value={backup?.restoreSim?.blockedNamespaces || 0} tone="critical" />
      </div>
      <Card title="Policies">
        <DataTable caption="Policies" rows={(backup?.policies || []) as Array<Record<string, unknown>>} columns={[{ key: "tool", label: "Tool" }, { key: "name", label: "Policy" }, { key: "schedule", label: "Schedule" }, { key: "rpoHours", label: "RPO (h)" }, { key: "lastSuccessAt", label: "Last Success" }]} />
      </Card>
      <Card title="Restore Readiness">
        <KeyValueGrid items={[["Ready namespaces", String(backup?.restoreSim?.readyNamespaces || 0)], ["Blocked namespaces", String(backup?.restoreSim?.blockedNamespaces || 0)], ["Warning namespaces", String(backup?.restoreSim?.warningNamespaces || 0)], ["Unknown namespaces", String(backup?.restoreSim?.unknownNamespaces || 0)], ["Covered PVC GiB", String(backup?.restoreSim?.coveredPvcsGb || 0)], ["Data at risk GiB", String(backup?.restoreSim?.estimatedDataAtRiskGb || 0)]]} />
        {(backup?.restoreSim?.blockingReasons || []).length ? <p className="muted">Top blockers: {(backup?.restoreSim?.blockingReasons || []).join(" · ")}</p> : null}
      </Card>
      <Card title="Restore Simulation">
        <DataTable caption="Restore simulation" rows={(backup?.restoreSim?.namespaces || []) as Array<Record<string, unknown>>} columns={[{ key: "namespace", label: "Namespace" }, { key: "readiness", label: "Readiness", render: (value) => titleCase(String(value || "unknown")) }, { key: "hasCoverage", label: "Covered", render: (value) => statusCell(Boolean(value)) }, { key: "rpoHours", label: "RPO (h)" }, { key: "pvcSizeGb", label: "PVC GB" }, { key: "blockers", label: "Blockers", render: (value) => listCell(value) }, { key: "warnings", label: "Warnings", render: (value) => listCell(value) }]} />
      </Card>
      <Card title="Restore Drill Plan">
        <DataTable caption="Restore drill plan" rows={(backup?.drillPlan || []) as Array<Record<string, unknown>>} columns={[{ key: "phase", label: "Phase", render: (value) => titleCase(String(value || "")) }, { key: "title", label: "Step" }, { key: "ownerHint", label: "Owner" }, { key: "detail", label: "Detail" }, { key: "validation", label: "Validate", render: (value) => listCell(value) }]} />
      </Card>
    </div>
  );
}

function ScorePanel(props: { bundle: Bundle }) {
  return (
    <div className="metric-grid">
      <MetricCard label="Overall" value={props.bundle.score.overall.final} tone="accent" />
      <MetricCard label="Storage" value={props.bundle.score.storage.final} />
      <MetricCard label="Workload" value={props.bundle.score.workload.final} />
      <MetricCard label="Config" value={props.bundle.score.config.final} />
      <MetricCard label="Backup" value={props.bundle.score.backup.final} />
    </div>
  );
}

function ComparePanel(props: { bundle: Bundle }) {
  const compare = props.bundle.comparison;
  return (
    <div className="results-stack">
      <div className="metric-grid">
        <MetricCard label="Previous Score" value={compare?.previousScore || "n/a"} />
        <MetricCard label="Delta" value={`${(compare?.scoreDelta || 0) > 0 ? "+" : ""}${compare?.scoreDelta || 0}`} tone="accent" />
        <MetricCard label="New Findings" value={(compare?.findingsNew || []).length} tone="high" />
        <MetricCard label="Resolved" value={(compare?.findingsResolved || []).length} tone="success" />
        <MetricCard label="Regressed" value={(compare?.findingsRegressed || []).length} tone="critical" />
        <MetricCard label="Persistent" value={compare?.persistentFindingCount || 0} />
      </div>
      <Card title="Change Summary">
        <KeyValueGrid items={[["Previous Scan", compare?.previousScannedAt || "n/a"], ["Maturity", `${compare?.previousMaturity || "n/a"} -> ${compare?.currentMaturity || "n/a"}`], ["Backup Tool", `${compare?.backupToolPrevious || "none"} -> ${compare?.backupToolCurrent || "none"}`], ["Namespaces Added", (compare?.namespacesAdded || []).join(", ") || "none"], ["Workloads Added", (compare?.workloadsAdded || []).join(", ") || "none"]]} />
      </Card>
      <Card title="Score Deltas">
        <DataTable caption="Score deltas" rows={(compare?.domainDeltas || []) as Array<Record<string, unknown>>} columns={[{ key: "name", label: "Domain", render: (value) => titleCase(String(value || "")) }, { key: "previous", label: "Previous" }, { key: "current", label: "Current" }, { key: "delta", label: "Delta", render: (value) => formatDelta(Number(value || 0)) }]} />
      </Card>
      <Card title="Severity Deltas">
        <DataTable caption="Severity deltas" rows={(compare?.severityDeltas || []) as Array<Record<string, unknown>>} columns={[{ key: "severity", label: "Severity" }, { key: "previous", label: "Previous" }, { key: "current", label: "Current" }, { key: "delta", label: "Delta", render: (value) => formatDelta(Number(value || 0)) }]} />
      </Card>
      <Card title="Inventory Deltas">
        <DataTable caption="Inventory deltas" rows={(compare?.inventoryDeltas || []) as Array<Record<string, unknown>>} columns={[{ key: "name", label: "Area", render: (value) => titleCase(String(value || "")) }, { key: "added", label: "Added" }, { key: "removed", label: "Removed" }]} />
      </Card>
      <DataTable caption="New findings" rows={(compare?.findingsNew || []) as Array<Record<string, unknown>>} columns={[{ key: "severity", label: "Severity" }, { key: "resourceId", label: "Resource" }, { key: "message", label: "Message" }]} />
      <DataTable caption="Regressed findings" rows={(compare?.findingsRegressed || []) as Array<Record<string, unknown>>} columns={[{ key: "previousSeverity", label: "Was" }, { key: "currentSeverity", label: "Now" }, { key: "change", label: "Change" }, { key: "resourceId", label: "Resource" }, { key: "message", label: "Message" }, { key: "ownerHint", label: "Owner" }]} />
      <DataTable caption="Improved findings" rows={(compare?.findingsImproved || []) as Array<Record<string, unknown>>} columns={[{ key: "previousSeverity", label: "Was" }, { key: "currentSeverity", label: "Now" }, { key: "change", label: "Change" }, { key: "resourceId", label: "Resource" }, { key: "message", label: "Message" }]} />
    </div>
  );
}

function formatDelta(value: number) {
  return `${value > 0 ? "+" : ""}${value}`;
}
