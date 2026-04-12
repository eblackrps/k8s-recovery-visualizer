import type { KeyboardEvent, ReactNode } from "react";
import type { ArtifactPaths, Bundle } from "../lib/types";

export function Field(props: { label: string; children: ReactNode }) {
  return (
    <label className="field">
      <span>{props.label}</span>
      {props.children}
    </label>
  );
}

export function Card(props: { title: string; children: ReactNode }) {
  return (
    <section className="subpanel">
      <div className="section-header compact">
        <h4>{props.title}</h4>
      </div>
      {props.children}
    </section>
  );
}

export function ReviewCard(props: { label: string; value: string }) {
  return (
    <div className="review-card">
      <span>{props.label}</span>
      <strong>{props.value}</strong>
    </div>
  );
}

export function MetricCard(props: { label: string; value: string | number; tone?: "accent" | "success" | "critical" | "high" }) {
  return (
    <div className={`metric-card ${props.tone || ""}`}>
      <span>{props.label}</span>
      <strong>{props.value}</strong>
    </div>
  );
}

export function KeyValueGrid(props: { items: Array<[string, string]> }) {
  return (
    <dl className="kv-grid">
      {props.items.map(([label, value]) => (
        <div key={label} className="kv-row">
          <dt>{label}</dt>
          <dd>{value}</dd>
        </div>
      ))}
    </dl>
  );
}

export function DataTable(props: {
  caption: string;
  rows: Array<Record<string, unknown>>;
  columns: Array<{ key: string; label: string; render?: (value: unknown, row: Record<string, unknown>) => ReactNode }>;
}) {
  return (
    <div className="table-shell">
      <table>
        <caption>{props.caption}</caption>
        <thead>
          <tr>
            {props.columns.map((column) => (
              <th key={column.key}>{column.label}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {props.rows.length ? (
            props.rows.map((row, rowIndex) => (
              <tr key={rowIndex}>
                {props.columns.map((column) => (
                  <td key={column.key}>{column.render ? column.render(row[column.key], row) : stringifyValue(row[column.key])}</td>
                ))}
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan={props.columns.length} className="empty-cell">
                No data for this section.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

export function TrendRail(props: { entries: Array<{ timestampUtc: string; overall: number; maturity: string }> }) {
  if (!props.entries.length) {
    return <p className="muted trend-empty">No history recorded yet.</p>;
  }
  const max = Math.max(...props.entries.map((entry) => entry.overall), 100);
  return (
    <div className="trend-rail" aria-label="History trend">
      {props.entries.map((entry) => (
        <div key={entry.timestampUtc} className="trend-stop">
          <span style={{ height: `${(entry.overall / max) * 100}%` }} />
          <strong>{entry.overall}</strong>
          <small>{entry.maturity}</small>
        </div>
      ))}
    </div>
  );
}

export function normalizeWorkloads(bundle: Bundle): Array<Record<string, unknown>> {
  return [
    ...((bundle.inventory.deployments || []) as Array<Record<string, unknown>>).map((item) => ({
      kind: "Deployment",
      namespace: item.namespace,
      name: item.name,
      summary: `${item.ready}/${item.replicas} ready`,
      images: item.images,
    })),
    ...((bundle.inventory.daemonSets || []) as Array<Record<string, unknown>>).map((item) => ({
      kind: "DaemonSet",
      namespace: item.namespace,
      name: item.name,
      summary: `${item.ready}/${item.desired} ready`,
      images: item.images,
    })),
    ...((bundle.inventory.jobs || []) as Array<Record<string, unknown>>).map((item) => ({
      kind: "Job",
      namespace: item.namespace,
      name: item.name,
      summary: item.completed ? "completed" : "running",
      images: [],
    })),
    ...((bundle.inventory.cronJobs || []) as Array<Record<string, unknown>>).map((item) => ({
      kind: "CronJob",
      namespace: item.namespace,
      name: item.name,
      summary: item.schedule,
      images: [],
    })),
  ];
}

export function applyTheme(theme: {
  palette: Record<string, string>;
  maturity: Record<string, string>;
  typography: Record<string, string>;
  radius: Record<string, string>;
}) {
  const root = document.documentElement;
  root.style.setProperty("--bg", theme.palette.background);
  root.style.setProperty("--surface", theme.palette.surface);
  root.style.setProperty("--border", theme.palette.border);
  root.style.setProperty("--text", theme.palette.text);
  root.style.setProperty("--muted", theme.palette.muted);
  root.style.setProperty("--accent", theme.palette.accent);
  root.style.setProperty("--success", theme.palette.success);
  root.style.setProperty("--critical", theme.palette.critical);
  root.style.setProperty("--high", theme.palette.high);
  root.style.setProperty("--medium", theme.palette.medium);
  root.style.setProperty("--maturity-platinum", theme.maturity.platinum);
  root.style.setProperty("--maturity-gold", theme.maturity.gold);
  root.style.setProperty("--maturity-silver", theme.maturity.silver);
  root.style.setProperty("--maturity-bronze", theme.maturity.bronze);
  root.style.setProperty("--font-body", theme.typography.body);
  root.style.setProperty("--font-title", theme.typography.title);
  root.style.setProperty("--font-mono", theme.typography.mono);
  root.style.setProperty("--radius-xl", theme.radius.xl);
  root.style.setProperty("--radius-lg", theme.radius.lg);
  root.style.setProperty("--radius-md", theme.radius.md);
  root.style.setProperty("--radius-sm", theme.radius.sm);
}

export function splitList(value: string) {
  return value
    .split(",")
    .map((entry) => entry.trim())
    .filter(Boolean);
}

export function stringifyValue(value: unknown) {
  if (Array.isArray(value)) {
    return value.join(", ");
  }
  if (typeof value === "boolean") {
    return value ? "yes" : "no";
  }
  if (value == null || value === "") {
    return "—";
  }
  return String(value);
}

export function listCell(value: unknown) {
  if (!Array.isArray(value) || !value.length) {
    return "—";
  }
  return value.join(", ");
}

export function statusCell(value: boolean) {
  return <span className={`chip ${value ? "chip-pass" : "chip-fail"}`}>{value ? "yes" : "no"}</span>;
}

export function handleRovingTabs(
  event: KeyboardEvent<HTMLButtonElement>,
  items: string[],
  current: number,
  setIndex: (index: number) => void,
) {
  if (event.key !== "ArrowRight" && event.key !== "ArrowLeft" && event.key !== "Home" && event.key !== "End") {
    return;
  }
  event.preventDefault();
  if (event.key === "Home") {
    setIndex(0);
    return;
  }
  if (event.key === "End") {
    setIndex(items.length - 1);
    return;
  }
  if (event.key === "ArrowRight") {
    setIndex((current + 1) % items.length);
    return;
  }
  setIndex((current - 1 + items.length) % items.length);
}

export function exportMessage(kind: string, artifacts: ArtifactPaths) {
  switch (kind) {
    case "summary":
      return `Summary refreshed at ${artifacts.summaryHtml}`;
    case "runbook":
      return `Runbook refreshed at ${artifacts.runbookHtml}`;
    case "csv":
      return `CSV directory refreshed at ${artifacts.csvDir}`;
    case "redacted":
      return `Redacted bundle refreshed at ${artifacts.redactedHtml}`;
    case "json":
      return `Bundle JSON refreshed at ${artifacts.bundleJson}`;
    default:
      return `Report refreshed at ${artifacts.htmlReport}`;
  }
}

export function titleCase(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1);
}

export function formatTime(value: string) {
  if (!value) {
    return "";
  }
  return new Date(value).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}
