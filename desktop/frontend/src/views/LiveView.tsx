import type { RunEvent } from "../lib/types";
import { MetricCard, formatTime } from "../components/ui";

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
        <div className="section-header">
          <div>
            <p className="eyebrow">Run Progress</p>
            <h3>Collector events and progress</h3>
          </div>
          <button
            type="button"
            className="button danger"
            onClick={props.onCancel}
            disabled={!canCancel}
            title={canCancel ? "Stop the active collection run." : "No active run is available to cancel."}
          >
            {canCancel ? "Cancel Run" : "Cancel Unavailable"}
          </button>
        </div>
        <div className="progress-card">
          <div className="progress-head">
            <strong>{Math.round(props.activePercent * 100)}%</strong>
            <span className="chip">{props.activeRunId || "No active run"}</span>
          </div>
          <div className="progress-bar">
            <span style={{ width: `${Math.round(props.activePercent * 100)}%` }} />
          </div>
        </div>
        <div className="timeline">
          {reverseEvents.map((event, index) => (
            <article key={`${event.timestamp}-${index}`} className={`timeline-item ${event.level || "info"}`}>
              <div className="timeline-title">
                <strong>{event.step || event.type}</strong>
                <span>{formatTime(event.timestamp)}</span>
              </div>
              <p>{event.message}</p>
              {event.warning ? <p className="muted">{event.warning}</p> : null}
            </article>
          ))}
        </div>
      </section>

      <section className="panel">
        <div className="section-header">
          <div>
            <p className="eyebrow">Warnings</p>
            <h3>Skipped collectors and access caveats</h3>
          </div>
        </div>
        <div className="inline-metrics">
          <MetricCard label="Warnings" value={warnings.length} tone={warnings.length ? "high" : "success"} />
          <MetricCard label="Event Count" value={props.events.length} />
        </div>
        <div className="stack-list">
          {(warnings.length
            ? warnings
            : [{ message: "No warnings yet.", timestamp: "", runId: "", type: "log" } as RunEvent]
          ).map((event, index) => (
            <div key={`${event.message}-${index}`} className="status-card warn">
              <strong>{event.step || "warning"}</strong>
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
