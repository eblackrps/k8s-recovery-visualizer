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
            role="tab"
            aria-selected={props.resultTab === tab}
            className={`tab ${props.resultTab === tab ? "is-active" : ""}`}
            onClick={() => props.setResultTab(tab)}
            onKeyDown={(event) => handleRovingTabs(event, activeTabs, index, (next) => props.setResultTab(activeTabs[next]))}
          >
            {tab}
          </button>
        ))}
      </div>

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
          <div className="filter-row">
            {["ALL", "CRITICAL", "HIGH", "MEDIUM", "LOW"].map((filter) => (
              <button key={filter} type="button" className={`chip-button ${props.findingFilter === filter ? "is-active" : ""}`} onClick={() => props.setFindingFilter(filter)}>
                {filter}
              </button>
            ))}
          </div>
          <DataTable
            caption="Findings"
            rows={filteredFindings as Array<Record<string, unknown>>}
            columns={[
              { key: "severity", label: "Severity" },
              { key: "resourceId", label: "Resource" },
              { key: "message", label: "Finding" },
              { key: "recommendation", label: "Recommendation" },
            ]}
          />
        </div>
      )}
      {props.resultTab === "Remediation" && (
        <div className="stack-list">
          {(bundle.inventory.remediationSteps || []).map((step, index) => (
            <article key={`${step.title}-${index}`} className="remediation-card">
              <div className="section-header">
                <div>
                  <p className="eyebrow">Priority {step.priority}</p>
                  <h4>{step.title}</h4>
                </div>
                <span className="chip">{step.category}</span>
              </div>
              <p>{step.detail}</p>
              {step.whyItMatters ? <p className="muted">Why it matters: {step.whyItMatters}</p> : null}
              {step.validation?.length ? <p className="muted">Validate: {step.validation.join(" · ")}</p> : null}
              {step.fixSteps?.length ? <p className="muted">Fix: {step.fixSteps.join(" · ")}</p> : null}
              {step.commands?.length ? <code className="mono-block">{step.commands.join("\n")}</code> : null}
            </article>
          ))}
        </div>
      )}
      {props.resultTab === "Compare" && bundle.comparison ? <ComparePanel bundle={bundle} /> : null}
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
      </div>
      <Card title="Policies">
        <DataTable caption="Policies" rows={(backup?.policies || []) as Array<Record<string, unknown>>} columns={[{ key: "tool", label: "Tool" }, { key: "name", label: "Policy" }, { key: "schedule", label: "Schedule" }, { key: "rpoHours", label: "RPO (h)" }, { key: "lastSuccessAt", label: "Last Success" }]} />
      </Card>
      <Card title="Restore Simulation">
        <DataTable caption="Restore simulation" rows={(backup?.restoreSim?.namespaces || []) as Array<Record<string, unknown>>} columns={[{ key: "namespace", label: "Namespace" }, { key: "hasCoverage", label: "Covered", render: (value) => statusCell(Boolean(value)) }, { key: "rpoHours", label: "RPO (h)" }, { key: "pvcSizeGb", label: "PVC GB" }, { key: "warnings", label: "Warnings", render: (value) => listCell(value) }]} />
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
      </div>
      <Card title="Change Summary">
        <KeyValueGrid items={[["Previous Scan", compare?.previousScannedAt || "n/a"], ["Backup Tool", `${compare?.backupToolPrevious || "none"} -> ${compare?.backupToolCurrent || "none"}`], ["Namespaces Added", (compare?.namespacesAdded || []).join(", ") || "none"], ["Workloads Added", (compare?.workloadsAdded || []).join(", ") || "none"]]} />
      </Card>
      <DataTable caption="New findings" rows={(compare?.findingsNew || []) as Array<Record<string, unknown>>} columns={[{ key: "severity", label: "Severity" }, { key: "resourceId", label: "Resource" }, { key: "message", label: "Message" }]} />
    </div>
  );
}
