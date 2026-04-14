import { useState } from "react";
import type { KeyboardEvent, ReactNode } from "react";
import type { ArtifactPaths, Bundle, RunCompletionSummary } from "../lib/types";

type Tone =
  | "neutral"
  | "accent"
  | "success"
  | "critical"
  | "high"
  | "medium"
  | "pass"
  | "fail"
  | "warn"
  | "gold"
  | "silver"
  | "platinum"
  | "bronze";

function cx(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(" ");
}

export function HelpTip(props: { label: string; children: ReactNode }) {
  return (
    <details className="help-tip">
      <summary aria-label={props.label}>?</summary>
      <div className="help-tip-panel">{props.children}</div>
    </details>
  );
}

export function SectionHeader(props: {
  eyebrow?: string;
  title: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  compact?: boolean;
  dense?: boolean;
  className?: string;
}) {
  return (
    <div className={cx("section-header", props.compact && "compact", props.className)}>
      <div className="section-heading">
        {props.eyebrow ? <p className="eyebrow">{props.eyebrow}</p> : null}
        <h3>{props.title}</h3>
        {props.description && !props.dense ? <p className="muted section-copy">{props.description}</p> : null}
      </div>
      {props.actions ? <div className="toolbar">{props.actions}</div> : null}
    </div>
  );
}

export function Field(props: {
  label: ReactNode;
  hint?: ReactNode;
  tip?: ReactNode;
  tipLabel?: string;
  error?: ReactNode;
  warning?: ReactNode;
  className?: string;
  children: ReactNode;
}) {
  return (
    <label className={cx("field", Boolean(props.error) && "has-error", Boolean(props.warning) && !props.error && "has-warning", props.className)}>
      <span className="field-head">
        <span className="field-label">{props.label}</span>
        {props.tip ? <HelpTip label={props.tipLabel || "Field help"}>{props.tip}</HelpTip> : null}
      </span>
      {props.children}
      {props.hint ? <small className="field-hint">{props.hint}</small> : null}
      {props.warning ? <small className="field-message warning">{props.warning}</small> : null}
      {props.error ? <small className="field-message error">{props.error}</small> : null}
    </label>
  );
}

export function Card(props: {
  title?: ReactNode;
  eyebrow?: string;
  description?: ReactNode;
  actions?: ReactNode;
  className?: string;
  children: ReactNode;
}) {
  return (
    <section className={cx("subpanel", props.className)}>
      {props.title ? (
        <SectionHeader
          eyebrow={props.eyebrow}
          title={props.title}
          description={props.description}
          actions={props.actions}
          compact
        />
      ) : null}
      {props.children}
    </section>
  );
}

export function EmptyState(props: {
  eyebrow?: string;
  title: ReactNode;
  description: ReactNode;
  actions?: ReactNode;
  children?: ReactNode;
  className?: string;
}) {
  return (
    <section className={cx("empty-state-panel", props.className)}>
      <SectionHeader eyebrow={props.eyebrow} title={props.title} description={props.description} actions={props.actions} />
      {props.children ? <div className="empty-state-body">{props.children}</div> : null}
    </section>
  );
}

export function Stepper(props: {
  current: string;
  steps: Array<{
    id: string;
    label: string;
    description: string;
    status: "current" | "complete" | "upcoming";
  }>;
  onSelect?: (id: string) => void;
}) {
  return (
    <ol className="stepper" aria-label="Scan setup progress">
      {props.steps.map((step, index) => {
        const clickable = step.status !== "upcoming" && props.onSelect;
        const content = (
          <>
            <span className="stepper-index" aria-hidden="true">{index + 1}</span>
            <span className="stepper-copy">
              <strong>{step.label}</strong>
              <small>{step.description}</small>
            </span>
          </>
        );

        return (
          <li key={step.id} className={cx("stepper-item", `is-${step.status}`)}>
            {clickable ? (
              <button
                type="button"
                className="stepper-button"
                aria-current={props.current === step.id ? "step" : undefined}
                onClick={() => props.onSelect?.(step.id)}
              >
                {content}
              </button>
            ) : (
              <div className="stepper-button" aria-current={props.current === step.id ? "step" : undefined}>
                {content}
              </div>
            )}
          </li>
        );
      })}
    </ol>
  );
}

export function Badge(props: { children: ReactNode; tone?: Tone; className?: string }) {
  return <span className={cx("badge", props.tone && `badge-${props.tone}`, props.className)}>{props.children}</span>;
}

export function ReviewCard(props: { label: string; value: string; detail?: ReactNode; tone?: Tone }) {
  return (
    <div className={cx("review-card", props.tone && `review-card-${props.tone}`)}>
      <span>{props.label}</span>
      <strong>{props.value}</strong>
      {props.detail ? <p className="muted">{props.detail}</p> : null}
    </div>
  );
}

export function CodeBlock(props: { code: string; label?: string; className?: string }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard?.writeText(props.code);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1400);
    } catch {
      setCopied(false);
    }
  }

  return (
    <div className={cx("code-block", props.className)}>
      <div className="code-block-head">
        <span>{props.label || "Command"}</span>
        <button type="button" className="button secondary quiet code-copy" onClick={() => void handleCopy()}>
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
      <code className="mono-block">{props.code}</code>
    </div>
  );
}

export function MetricCard(props: {
  label: string;
  value: string | number;
  tone?: "accent" | "success" | "critical" | "high" | "medium";
  detail?: ReactNode;
  className?: string;
}) {
  const valueText = String(props.value);
  const compactValue = valueText.length > 8 || valueText.includes(" ");

  return (
    <div className={cx("metric-card", props.tone, compactValue && "compact-value", props.className)}>
      <span>{props.label}</span>
      <strong>{valueText}</strong>
      {props.detail ? <small>{props.detail}</small> : null}
    </div>
  );
}

export function KeyValueGrid(props: { items: Array<[string, ReactNode]>; className?: string }) {
  return (
    <dl className={cx("kv-grid", props.className)}>
      {props.items.map(([label, value]) => (
        <div key={label} className="kv-row">
          <dt>{label}</dt>
          <dd>{value}</dd>
        </div>
      ))}
    </dl>
  );
}

export function ReadinessList(props: {
  items: Array<{
    label: string;
    value: ReactNode;
    detail?: ReactNode;
    state?: "ready" | "caution" | "missing" | "neutral";
  }>;
  className?: string;
}) {
  return (
    <div className={cx("readiness-list", props.className)}>
      {props.items.map((item) => (
        <article key={item.label} className={cx("readiness-item", item.state && `is-${item.state}`)}>
          <div className="readiness-copy">
            <strong>{item.label}</strong>
            {item.detail ? <p className="muted">{item.detail}</p> : null}
          </div>
          <div className="readiness-meta">
            <span className="readiness-value">{item.value}</span>
            <Badge tone={toneForReadiness(item.state)}>{labelForReadiness(item.state)}</Badge>
          </div>
        </article>
      ))}
    </div>
  );
}

export function RunCompletionCallout(props: {
  summary: RunCompletionSummary;
  title?: ReactNode;
  description?: ReactNode;
  onOpenPath?: (path: string, label: string) => void | Promise<void>;
  onReviewResults?: () => void;
  onReviewFindings?: () => void;
  onReviewCompare?: () => void;
  onStartAnotherScan?: () => void;
  onDismiss?: () => void;
}) {
  const outputDir = props.summary.artifacts.outputDir;
  const bundlePath = props.summary.artifacts.bundleJson || props.summary.artifacts.loadedBundlePath;
  const reportPath = props.summary.artifacts.htmlReport;
  const hasExtraActions =
    Boolean(props.onReviewResults) ||
    Boolean(props.onStartAnotherScan) ||
    Boolean(reportPath) ||
    Boolean(bundlePath) ||
    Boolean(props.onDismiss);

  return (
    <Card
      title={props.title || "Scan complete. Outputs ready."}
      description={
        props.description ||
        "K8V wrote the bundle and refreshed the report outputs. Start with findings, or open the output folder if this run needs to be handed off."
      }
      className="completion-callout"
      actions={
        <div className="toolbar wrap-toolbar completion-actions">
          {props.onReviewFindings ? (
            <button type="button" className="button primary" onClick={props.onReviewFindings}>
              Review findings
            </button>
          ) : null}
          {props.onReviewCompare && props.summary.hasComparison ? (
            <button type="button" className="button secondary" onClick={props.onReviewCompare}>
              Review compare
            </button>
          ) : null}
          {outputDir ? (
            <button type="button" className="button secondary" onClick={() => void props.onOpenPath?.(outputDir, "output folder")}>
              Open output folder
            </button>
          ) : null}
          {hasExtraActions ? (
            <details className="export-menu completion-more-menu">
              <summary className="button secondary quiet">More actions</summary>
              <div className="export-menu-body">
                {props.onReviewResults ? (
                  <button type="button" className="menu-action" onClick={props.onReviewResults}>
                    Review results
                  </button>
                ) : null}
                {reportPath ? (
                  <button
                    type="button"
                    className="menu-action"
                    onClick={() => void props.onOpenPath?.(reportPath, "primary report")}
                  >
                    Open primary report
                  </button>
                ) : null}
                {bundlePath ? (
                  <button
                    type="button"
                    className="menu-action"
                    onClick={() => void props.onOpenPath?.(bundlePath, "bundle JSON")}
                  >
                    Open bundle JSON
                  </button>
                ) : null}
                {props.onStartAnotherScan ? (
                  <button type="button" className="menu-action" onClick={props.onStartAnotherScan}>
                    Start another scan
                  </button>
                ) : null}
                {props.onDismiss ? (
                  <button type="button" className="menu-action" onClick={props.onDismiss}>
                    Dismiss
                  </button>
                ) : null}
              </div>
            </details>
          ) : null}
        </div>
      }
    >
      <div className="completion-grid completion-grid-compact">
        <div className="completion-metrics">
          <MetricCard label="Cluster" value={props.summary.clusterName || "Unknown"} />
          <MetricCard label="Score" value={props.summary.score ?? "n/a"} tone="accent" />
          <MetricCard
            label="Findings"
            value={props.summary.findingCount ?? 0}
            tone={(props.summary.findingCount || 0) > 0 ? "high" : "success"}
          />
        </div>
        <KeyValueGrid
          className="compact-kv-grid"
          items={[
            ["Completed", formatCompletionTimestamp(props.summary.generatedAt)],
            ["Output directory", <code className="path-chip">{outputDir || "Not recorded"}</code>],
            ["Reopen later", <code className="path-chip">{bundlePath || "Not recorded"}</code>],
          ]}
        />
      </div>
    </Card>
  );
}

export function DataTable(props: {
  caption: string;
  rows: Array<Record<string, unknown>>;
  columns: Array<{
    key: string;
    label: string;
    render?: (value: unknown, row: Record<string, unknown>) => ReactNode;
    className?: string;
    headerClassName?: string;
  }>;
  rowKey?: string | ((row: Record<string, unknown>, rowIndex: number) => string);
  className?: string;
  dense?: boolean;
  stickyHeader?: boolean;
  emptyMessage?: string;
}) {
  if (!props.rows.length) {
    return (
      <section className={cx("table-empty-state", props.className)} aria-label={props.caption}>
        <div className="section-header compact">
          <div className="section-heading">
            <h4>{props.caption}</h4>
          </div>
        </div>
        <p className="muted">{props.emptyMessage || "No data for this section."}</p>
      </section>
    );
  }

  return (
    <section className={cx("table-block", props.className)}>
      <div className="table-heading">
        <h4>{props.caption}</h4>
      </div>
      <div className={cx("table-shell", props.stickyHeader && "sticky-head", props.dense && "dense")}>
        <table>
          <caption className="sr-only">{props.caption}</caption>
          <thead>
            <tr>
              {props.columns.map((column) => (
                <th key={column.key} className={column.headerClassName}>
                  {column.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {props.rows.map((row, rowIndex) => {
              const key =
                typeof props.rowKey === "function"
                  ? props.rowKey(row, rowIndex)
                  : typeof props.rowKey === "string"
                    ? String(row[props.rowKey] ?? rowIndex)
                    : String(row.id ?? row.name ?? row.resourceId ?? rowIndex);

              return (
                <tr key={key}>
                  {props.columns.map((column) => (
                    <td key={column.key} className={column.className}>
                      {column.render ? column.render(row[column.key], row) : stringifyValue(row[column.key])}
                    </td>
                  ))}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}

export function TrendRail(props: { entries: Array<{ timestampUtc: string; overall: number; maturity: string }> }) {
  if (!props.entries.length) {
    return <p className="muted trend-empty">No history recorded yet.</p>;
  }

  const max = Math.max(...props.entries.map((entry) => entry.overall), 100);

  return (
    <div className="trend-rail" aria-label="History trend">
      {props.entries.map((entry, index) => {
        const previous = props.entries[index - 1]?.overall;
        const delta = previous == null ? null : entry.overall - previous;

        return (
          <div key={entry.timestampUtc} className="trend-stop">
            <div className="trend-bar">
              <span style={{ height: `${(entry.overall / max) * 100}%` }} />
            </div>
            <div className="trend-meta">
              <strong>{entry.overall}</strong>
              <small>{entry.maturity}</small>
              <span>{formatShortDate(entry.timestampUtc)}</span>
              <em className={cx(delta != null && toneForDelta(delta) && `text-${toneForDelta(delta)}`)}>
                {delta == null ? "Baseline" : formatDelta(delta)}
              </em>
            </div>
          </div>
        );
      })}
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
  const background = theme.palette.background;
  const surface = theme.palette.surface;
  const text = theme.palette.text;
  const accent = theme.palette.accent;
  const success = theme.palette.success;
  const critical = theme.palette.critical;
  const high = theme.palette.high;
  const medium = theme.palette.medium;

  root.style.setProperty("--bg", background);
  root.style.setProperty("--surface", surface);
  root.style.setProperty("--border", theme.palette.border);
  root.style.setProperty("--text", text);
  root.style.setProperty("--muted", theme.palette.muted);
  root.style.setProperty("--accent", accent);
  root.style.setProperty("--success", success);
  root.style.setProperty("--critical", critical);
  root.style.setProperty("--high", high);
  root.style.setProperty("--medium", medium);
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
  root.style.setProperty("--surface-raised", withAlpha(surface, 0.98));
  root.style.setProperty("--panel", withAlpha(surface, 0.94));
  root.style.setProperty("--panel-strong", withAlpha(surface, 1));
  root.style.setProperty("--bg-deep", withAlpha(background, 1));
  root.style.setProperty("--line", withAlpha(theme.palette.border, 0.9));
  root.style.setProperty("--line-soft", withAlpha(theme.palette.border, 0.52));
  root.style.setProperty("--muted-strong", withAlpha(text, 0.92));
  root.style.setProperty("--accent-soft", withAlpha(accent, 0.78));
  root.style.setProperty("--accent-strong", withAlpha(accent, 0.38));
  root.style.setProperty("--accent-faint", withAlpha(accent, 0.1));
  root.style.setProperty("--accent-surface", withAlpha(accent, 0.16));
  root.style.setProperty("--success-faint", withAlpha(success, 0.14));
  root.style.setProperty("--danger", critical);
  root.style.setProperty("--danger-faint", withAlpha(critical, 0.14));
  root.style.setProperty("--warning-high", high);
  root.style.setProperty("--warning-high-faint", withAlpha(high, 0.14));
  root.style.setProperty("--warning", medium);
  root.style.setProperty("--warning-medium", medium);
  root.style.setProperty("--warning-medium-faint", withAlpha(medium, 0.14));
  root.style.setProperty("--shadow", "0 12px 28px rgba(1, 4, 9, 0.18)");
  root.style.setProperty("--shadow-soft", "0 6px 18px rgba(1, 4, 9, 0.12)");
}

function withAlpha(hex: string, alpha: number) {
  const normalized = hex.replace("#", "").trim();
  if (normalized.length !== 6) {
    return hex;
  }
  const red = Number.parseInt(normalized.slice(0, 2), 16);
  const green = Number.parseInt(normalized.slice(2, 4), 16);
  const blue = Number.parseInt(normalized.slice(4, 6), 16);
  if ([red, green, blue].some((value) => Number.isNaN(value))) {
    return hex;
  }
  const clamped = Math.max(0, Math.min(1, alpha));
  return `rgba(${red}, ${green}, ${blue}, ${clamped.toFixed(2)})`;
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
  return <Badge tone={value ? "pass" : "fail"}>{value ? "yes" : "no"}</Badge>;
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
  return value
    .replace(/[_-]+/g, " ")
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}

export function formatTime(value: string) {
  if (!value) {
    return "";
  }
  return new Date(value).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

export function formatShortDate(value: string) {
  if (!value) {
    return "";
  }
  return new Date(value).toLocaleDateString([], { month: "short", day: "numeric" });
}

export function formatCompletionTimestamp(value: string | undefined) {
  if (!value) {
    return "Just now";
  }
  return new Date(value).toLocaleString([], {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatDelta(value: number) {
  return `${value > 0 ? "+" : ""}${value}`;
}

export function toneForDelta(value: number) {
  if (value > 0) {
    return "success";
  }
  if (value < 0) {
    return "critical";
  }
  return "neutral";
}

function toneForReadiness(state?: "ready" | "caution" | "missing" | "neutral"): Tone {
  switch (state) {
    case "ready":
      return "success";
    case "caution":
      return "warn";
    case "missing":
      return "fail";
    default:
      return "neutral";
  }
}

function labelForReadiness(state?: "ready" | "caution" | "missing" | "neutral") {
  switch (state) {
    case "ready":
      return "Ready";
    case "caution":
      return "Caution";
    case "missing":
      return "Needs setup";
    default:
      return "Info";
  }
}

export function toneForMaturity(value: string | undefined): Tone {
  const normalized = String(value || "").toLowerCase();
  switch (normalized) {
    case "platinum":
      return "platinum";
    case "gold":
      return "gold";
    case "silver":
      return "silver";
    case "bronze":
      return "bronze";
    default:
      return "neutral";
  }
}

export function toneForSeverity(value: string | undefined): Tone {
  switch (String(value || "").toUpperCase()) {
    case "CRITICAL":
      return "critical";
    case "HIGH":
      return "high";
    case "MEDIUM":
      return "medium";
    case "LOW":
      return "neutral";
    default:
      return "neutral";
  }
}
