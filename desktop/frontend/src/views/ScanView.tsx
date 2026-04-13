import type { PreflightReport, ScanRequest } from "../lib/types";
import { Field, MetricCard, ReviewCard, handleRovingTabs, splitList } from "../components/ui";

const wizardSteps = ["Access", "Scope", "Outputs", "Review"];

export function ScanView(props: {
  busy: boolean;
  wizardStep: number;
  setWizardStep: (index: number) => void;
  scanForm: ScanRequest;
  setScanForm: (updater: (current: ScanRequest) => ScanRequest) => void;
  preflight: PreflightReport | null;
  onPreflight: () => void;
  onStartScan: () => void;
  onBrowseOutput: () => void;
}) {
  const updateForm = <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => {
    props.setScanForm((current) => ({ ...current, [key]: value }));
  };

  return (
    <section className="page-grid scan-grid">
      <section className="panel">
        <div className="section-header">
          <div>
            <p className="eyebrow">New Scan</p>
            <h3>Guided scan setup</h3>
          </div>
          <button type="button" className="button secondary" onClick={props.onPreflight} disabled={props.busy}>
            Run Preflight
          </button>
        </div>
        <div className="stepper" role="tablist" aria-label="Scan wizard steps">
          {wizardSteps.map((step, index) => (
            <button
              key={step}
              type="button"
              id={`wizard-tab-${index}`}
              role="tab"
              aria-selected={props.wizardStep === index}
              aria-controls={`wizard-panel-${index}`}
              tabIndex={props.wizardStep === index ? 0 : -1}
              className={`step-pill ${props.wizardStep === index ? "is-active" : ""}`}
              onClick={() => props.setWizardStep(index)}
              onKeyDown={(event) => handleRovingTabs(event, wizardSteps, index, props.setWizardStep)}
            >
              <span className="step-index" aria-hidden="true">{index + 1}</span>
              <span className="step-label">{step}</span>
            </button>
          ))}
        </div>

        {props.wizardStep === 0 && (
          <div id="wizard-panel-0" role="tabpanel" aria-labelledby="wizard-tab-0" className="form-grid">
            <Field label="Kubeconfig Path">
              <input aria-label="Kubeconfig path" value={props.scanForm.kubeconfigPath || ""} onChange={(event) => updateForm("kubeconfigPath", event.target.value)} />
            </Field>
            <Field label="Context">
              <input aria-label="Context" value={props.scanForm.contextName || ""} onChange={(event) => updateForm("contextName", event.target.value)} />
            </Field>
            <Field label="Cluster Name">
              <input aria-label="Cluster name" value={props.scanForm.clusterName || ""} onChange={(event) => updateForm("clusterName", event.target.value)} />
            </Field>
            <Field label="Environment">
              <input aria-label="Environment" value={props.scanForm.environment || ""} onChange={(event) => updateForm("environment", event.target.value)} />
            </Field>
            <label className="toggle">
              <input type="checkbox" checked={Boolean(props.scanForm.dryRun)} onChange={(event) => updateForm("dryRun", event.target.checked)} />
              Dry-run fixture mode
            </label>
            <label className="toggle">
              <input type="checkbox" checked={Boolean(props.scanForm.insecure)} onChange={(event) => updateForm("insecure", event.target.checked)} />
              Skip TLS verification
            </label>
          </div>
        )}

        {props.wizardStep === 1 && (
          <div id="wizard-panel-1" role="tabpanel" aria-labelledby="wizard-tab-1" className="form-grid">
            <Field label="Namespaces (comma separated)">
              <input value={(props.scanForm.namespaces || []).join(", ")} onChange={(event) => updateForm("namespaces", splitList(event.target.value))} />
            </Field>
            <Field label="Profile">
              <select value={props.scanForm.profileName || "standard"} onChange={(event) => updateForm("profileName", event.target.value)}>
                <option value="standard">standard</option>
                <option value="enterprise">enterprise</option>
                <option value="dev">dev</option>
                <option value="airgap">airgap</option>
              </select>
            </Field>
            <Field label="Compare Baseline">
              <input value={props.scanForm.compareTo || ""} onChange={(event) => updateForm("compareTo", event.target.value)} />
            </Field>
            <Field label="Recovery Target">
              <select value={props.scanForm.target || "vm"} onChange={(event) => updateForm("target", event.target.value)}>
                <option value="vm">vm</option>
                <option value="baremetal">baremetal</option>
              </select>
            </Field>
          </div>
        )}

        {props.wizardStep === 2 && (
          <div id="wizard-panel-2" role="tabpanel" aria-labelledby="wizard-tab-2" className="form-grid">
            <Field label="Output Directory">
              <div className="inline-field">
                <input value={props.scanForm.outputDir || ""} onChange={(event) => updateForm("outputDir", event.target.value)} />
                <button type="button" className="button secondary" onClick={props.onBrowseOutput} disabled={props.busy}>
                  Browse
                </button>
              </div>
            </Field>
            <Field label="Minimum Score">
              <input type="number" min={0} max={100} value={props.scanForm.minScore || 90} onChange={(event) => updateForm("minScore", Number(event.target.value))} />
            </Field>
            <Field label="Timeout (seconds)">
              <input type="number" min={10} value={props.scanForm.timeoutSeconds || 60} onChange={(event) => updateForm("timeoutSeconds", Number(event.target.value))} />
            </Field>
            <div className="toggle-grid">
              <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.summary)} onChange={(event) => updateForm("summary", event.target.checked)} /> Executive summary</label>
              <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.runbook)} onChange={(event) => updateForm("runbook", event.target.checked)} /> Customer runbook</label>
              <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.redact)} onChange={(event) => updateForm("redact", event.target.checked)} /> Redacted outputs</label>
              <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.csvExport)} onChange={(event) => updateForm("csvExport", event.target.checked)} /> CSV exports</label>
              <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.includeSecretMetadata)} onChange={(event) => updateForm("includeSecretMetadata", event.target.checked)} /> Secret metadata</label>
            </div>
          </div>
        )}

        {props.wizardStep === 3 && (
          <div id="wizard-panel-3" role="tabpanel" aria-labelledby="wizard-tab-3" className="review-grid">
            <ReviewCard label="Kubeconfig" value={props.scanForm.kubeconfigPath || "default loading rules"} />
            <ReviewCard label="Context" value={props.scanForm.contextName || "default context"} />
            <ReviewCard label="Scope" value={(props.scanForm.namespaces || []).join(", ") || "all namespaces"} />
            <ReviewCard label="Profile" value={props.scanForm.profileName || "standard"} />
            <ReviewCard label="Compare Baseline" value={props.scanForm.compareTo || "none"} />
            <ReviewCard label="Output" value={props.scanForm.outputDir || "./out"} />
          </div>
        )}

        <div className="wizard-actions">
          <button type="button" className="button secondary" onClick={() => props.setWizardStep(Math.max(0, props.wizardStep - 1))} disabled={props.wizardStep === 0}>
            Back
          </button>
          <button type="button" className="button secondary" onClick={() => props.setWizardStep(Math.min(wizardSteps.length - 1, props.wizardStep + 1))} disabled={props.wizardStep === wizardSteps.length - 1}>
            Next
          </button>
          <button type="button" className="button primary" onClick={props.onStartScan} disabled={props.busy}>
            Start Scan
          </button>
        </div>
      </section>

      <section className="panel">
        <div className="section-header">
          <div>
            <p className="eyebrow">Preflight</p>
            <h3>Access validation and degraded mode</h3>
          </div>
        </div>
        {props.preflight ? (
          <>
            <div className="inline-metrics">
              <MetricCard label="Can Run" value={props.preflight.canRun ? "Yes" : "No"} tone={props.preflight.canRun ? "success" : "critical"} />
              <MetricCard label="Degraded" value={props.preflight.degraded ? "Yes" : "No"} tone={props.preflight.degraded ? "high" : "success"} />
              <MetricCard label="Scope" value={props.preflight.scope} />
            </div>
            <div className="stack-list">
              {props.preflight.checks.map((check) => (
                <div key={check.id} className={`status-card ${check.status}`}>
                  <div className="status-card-head">
                    <div>
                      <strong>{check.title}</strong>
                      {check.resource ? <p className="muted">{titleForProbe(check.scope, check.resource)}</p> : null}
                    </div>
                    <span className={`chip chip-${check.status}`}>{check.status}</span>
                  </div>
                  <p>{check.detail}</p>
                  {check.hint ? <p className="muted">{check.hint}</p> : null}
                  {check.commands?.length ? <p className="muted">Check with: {check.commands[0]}</p> : null}
                  {check.manifest ? <code className="mono-block">{check.manifest}</code> : null}
                </div>
              ))}
            </div>
          </>
        ) : (
          <p className="muted">Run preflight to validate access, namespace scope, and expected degraded-mode caveats.</p>
        )}
      </section>
    </section>
  );
}

function titleForProbe(scope?: string, resource?: string) {
  if (!resource) {
    return "";
  }
  return `${scope === "cluster" ? "Cluster-scope" : "Namespace-scope"} access for ${resource}`;
}
