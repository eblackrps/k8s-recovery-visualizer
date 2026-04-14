import type { RunEvent } from "../lib/types";
import { Badge, MetricCard, SectionHeader, formatTime } from "../components/ui";

export function LiveView(props: {
  events: RunEvent[];
  activeRunId: string | null;
  activePercent: number;
  onCancel: () => void;
}) {
  const warnings = props.events.filter((event) => event.type === "warning");
  const reverseEvents = props.events.slice().reverse();
  const canCancel = Boolean(props.activeRunId);

  return (
    <section className="page-grid live-grid">
      <section className="panel">
        <SectionHeader
          eyebrow="Run Progress"
          title="Collector events and progress"
          description="Live collection stays focused on progress, event order, and access caveats so operators can decide quickly whether to wait, intervene, or cancel."
          actions={
            <button
              type="button"
              className="button danger"
              onClick={props.onCancel}
              disabled={!canCancel}
              title={canCancel ? "Stop the active collection run." : "No active run is available to cancel."}
            >
              {canCancel ? "Cancel Run" : "Cancel Unavailable"}
            </button>
          }
        />

        <div className="live-summary">
          <div className="progress-card">
            <div className="progress-head">
              <div className="progress-title">
                <strong>{Math.round(props.activePercent * 100)}%</strong>
                <span className="muted">collection complete</span>
              </div>
              <Badge>{props.activeRunId || "No active run"}</Badge>
            </div>
            <div className="progress-bar">
              <span style={{ width: `${Math.round(props.activePercent * 100)}%` }} />
            </div>
          </div>

          <div className="inline-metrics">
            <MetricCard label="Events" value={props.events.length} />
            <MetricCard label="Warnings" value={warnings.length} tone={warnings.length ? "high" : "success"} />
          </div>
        </div>

        <div className="timeline">
          {reverseEvents.map((event, index) => (
            <article key={`${event.timestamp}-${index}`} className={`timeline-item ${event.level || "info"}`}>
              <div className="timeline-title">
                <div className="timeline-step">
                  <strong>{event.step || event.type}</strong>
                  <span className="muted">{event.message}</span>
                </div>
                <span>{formatTime(event.timestamp)}</span>
              </div>
              {event.warning ? <p className="muted">{event.warning}</p> : null}
            </article>
          ))}
        </div>
      </section>

      <section className="panel">
        <SectionHeader
          eyebrow="Warnings"
          title="Skipped collectors and access caveats"
          description="Warnings stay separate from the main timeline so reduced visibility is obvious without overwhelming the normal event stream."
        />
        <div className="stack-list">
          {(warnings.length
            ? warnings
            : [{ message: "No warnings yet.", timestamp: "", runId: "", type: "log" } as RunEvent]
          ).map((event, index) => (
            <div key={`${event.message}-${index}`} className="warning-item">
              <div className="warning-head">
                <strong>{event.step || "warning"}</strong>
                <Badge tone={warnings.length ? "warn" : "pass"}>{warnings.length ? "warning" : "clear"}</Badge>
              </div>
              <p>{event.message}</p>
              {event.skip ? (
                <p className="muted">
                  {event.skip.name}: {event.skip.reason}
                </p>
              ) : null}
            </div>
          ))}
        </div>
      </section>
    </section>
  );
}
