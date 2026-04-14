import type { ConnectionMethod, ContextCatalog, PreflightReport, ScanRequest } from "../lib/types";
import { Badge, Card, Field, HelpTip, MetricCard, ReviewCard, SectionHeader, splitList } from "../components/ui";

const connectionModes: Array<{ value: ConnectionMethod; label: string; description: string }> = [
  {
    value: "current",
    label: "Current login",
    description: "Best for desktops and jumpboxes where kubectl or KUBECONFIG already works.",
  },
  {
    value: "kubeconfig_file",
    label: "Kubeconfig file",
    description: "Choose a kubeconfig from disk and optionally select the context to scan.",
  },
  {
    value: "kubeconfig_inline",
    label: "Paste kubeconfig",
    description: "Paste kubeconfig content directly when you do not want to rely on local files.",
  },
  {
    value: "api_endpoint",
    label: "API endpoint",
    description: "Enter the control-plane host or IP and scan with a bearer token.",
  },
];

export function ScanView(props: {
  busy: boolean;
  scanForm: ScanRequest;
  setScanForm: (updater: (current: ScanRequest) => ScanRequest) => void;
  preflight: PreflightReport | null;
  contextCatalog: ContextCatalog | null;
  detectingContexts: boolean;
  validationErrors: string[];
  onPreflight: () => void;
  onStartScan: () => void;
  onDetectContexts: () => void;
  onBrowseOutput: () => void;
  onBrowseKubeconfig: () => void;
  onBrowseCACert: () => void;
}) {
  const updateForm = <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => {
    props.setScanForm((current) => ({ ...current, [key]: value }));
  };
  const connectionMethod = props.scanForm.connectionMethod || "current";
  const supportsContextDiscovery = connectionMethod !== "api_endpoint";

  return (
    <section className="page-grid scan-grid">
      <section className="panel">
        <SectionHeader
          eyebrow="New Scan"
          title="Remote cluster scan"
          description="Use the simplest connection that already works on this machine. Scope, outputs, and advanced collection options stay explicit and close to the launch controls."
          actions={
            <HelpTip label="Scan setup help">
              <p>The desktop app runs locally, but it can scan any Kubernetes cluster your machine can reach.</p>
              <p>Use the current login on a desktop or jumpbox, a kubeconfig file, pasted kubeconfig, or a direct API endpoint with a token.</p>
            </HelpTip>
          }
        />

        <div className="scan-stack">
          <Card title="1. Connection" description="Choose the access path, then validate that the app can see the right context before you launch.">
            <div className="connection-modes" role="radiogroup" aria-label="Connection method">
              {connectionModes.map((mode) => (
                <label key={mode.value} className={`mode-card ${connectionMethod === mode.value ? "is-active" : ""}`}>
                  <input
                    type="radio"
                    name="connectionMethod"
                    checked={connectionMethod === mode.value}
                    onChange={() => updateForm("connectionMethod", mode.value)}
                  />
                  <strong>{mode.label}</strong>
                  <span>{mode.description}</span>
                </label>
              ))}
            </div>

            <div className="form-grid">
              {connectionMethod === "current" && (
                <Field
                  label="Context (optional)"
                  hint="Leave blank to use the current kubectl context."
                  tipLabel="Context help"
                  tip={<p>Use this when the machine has access to multiple clusters and you want to force one context for the scan.</p>}
                >
                  <input
                    aria-label="Context"
                    list={props.contextCatalog?.contexts?.length ? "scan-context-options" : undefined}
                    placeholder="Leave blank to use the current context"
                    value={props.scanForm.contextName || ""}
                    onChange={(event) => updateForm("contextName", event.target.value)}
                  />
                </Field>
              )}

              {connectionMethod === "kubeconfig_file" && (
                <>
                  <Field
                    label="Kubeconfig file"
                    hint="Use this when the jumpbox or desktop does not rely on the default kubectl config path."
                    tipLabel="Kubeconfig file help"
                    tip={<p>Point to a kubeconfig on disk. The app will use the selected context or the kubeconfig current context.</p>}
                  >
                    <div className="inline-field">
                      <input
                        aria-label="Kubeconfig file"
                        placeholder="C:\\Users\\you\\.kube\\config"
                        value={props.scanForm.kubeconfigPath || ""}
                        onChange={(event) => updateForm("kubeconfigPath", event.target.value)}
                      />
                      <button type="button" className="button secondary" onClick={props.onBrowseKubeconfig} disabled={props.busy}>
                        Browse
                      </button>
                    </div>
                  </Field>
                  <Field label="Context (optional)" hint="Leave blank to use the kubeconfig current context.">
                    <input
                      aria-label="Context"
                      list={props.contextCatalog?.contexts?.length ? "scan-context-options" : undefined}
                      placeholder="prod-east-admin"
                      value={props.scanForm.contextName || ""}
                      onChange={(event) => updateForm("contextName", event.target.value)}
                    />
                  </Field>
                </>
              )}

              {connectionMethod === "kubeconfig_inline" && (
                <>
                  <Field
                    label="Pasted kubeconfig"
                    hint="Useful when operators receive a kubeconfig snippet from a secure vault or ticket."
                    tipLabel="Pasted kubeconfig help"
                    tip={<p>Paste the full kubeconfig content, including clusters, users, and contexts. The app does not need a local file in this mode.</p>}
                  >
                    <textarea
                      aria-label="Pasted kubeconfig"
                      rows={10}
                      placeholder={"apiVersion: v1\nclusters: ..."}
                      value={props.scanForm.kubeconfigContent || ""}
                      onChange={(event) => updateForm("kubeconfigContent", event.target.value)}
                    />
                  </Field>
                  <Field label="Context (optional)" hint="Leave blank to use the kubeconfig current context.">
                    <input
                      aria-label="Context"
                      list={props.contextCatalog?.contexts?.length ? "scan-context-options" : undefined}
                      placeholder="prod-east-admin"
                      value={props.scanForm.contextName || ""}
                      onChange={(event) => updateForm("contextName", event.target.value)}
                    />
                  </Field>
                </>
              )}

              {connectionMethod === "api_endpoint" && (
                <>
                  <Field
                    label="API server host or URL"
                    hint="You can enter a host or IP like 10.0.0.15:6443 or a full https:// URL."
                    tipLabel="API endpoint help"
                    tip={<p>Use the Kubernetes control-plane address. If you only enter a host or IP, the app will assume HTTPS.</p>}
                  >
                    <input
                      aria-label="API server host or URL"
                      placeholder="10.0.0.15:6443 or https://cluster.example.net:6443"
                      value={props.scanForm.apiServerEndpoint || ""}
                      onChange={(event) => updateForm("apiServerEndpoint", event.target.value)}
                    />
                  </Field>
                  <Field
                    label="Bearer token"
                    hint="Use a service-account or other token accepted by the API server."
                    tipLabel="Bearer token help"
                    tip={<p>If the cluster uses exec plugins, client certificates, or cloud auth helpers, kubeconfig mode is usually the better fit.</p>}
                  >
                    <input
                      aria-label="Bearer token"
                      type="password"
                      placeholder="Paste a Kubernetes API bearer token"
                      value={props.scanForm.bearerToken || ""}
                      onChange={(event) => updateForm("bearerToken", event.target.value)}
                    />
                  </Field>
                </>
              )}
            </div>

            {supportsContextDiscovery ? (
              <div className="context-helper">
                <div className="inline-actions">
                  <button type="button" className="button secondary quiet" onClick={props.onDetectContexts} disabled={props.busy || props.detectingContexts}>
                    {props.detectingContexts ? "Loading Contexts..." : "Detect Contexts"}
                  </button>
                  {props.contextCatalog?.currentContext && props.scanForm.contextName !== props.contextCatalog.currentContext ? (
                    <button type="button" className="button secondary quiet" onClick={() => updateForm("contextName", props.contextCatalog?.currentContext || "")} disabled={props.busy}>
                      Use {props.contextCatalog.currentContext}
                    </button>
                  ) : null}
                </div>
                <p className="muted context-helper-status">
                  {props.contextCatalog
                    ? props.contextCatalog.contexts?.length
                      ? `Found ${props.contextCatalog.contexts.length} context${props.contextCatalog.contexts.length === 1 ? "" : "s"} from ${props.contextCatalog.source || "this connection"}${props.contextCatalog.currentContext ? `. Current: ${props.contextCatalog.currentContext}.` : "."}`
                      : `No named contexts were found for ${props.contextCatalog.source || "this connection"}.`
                    : "Load contexts to populate suggestions before preflight."}
                </p>
                {props.contextCatalog?.contexts?.length ? (
                  <datalist id="scan-context-options">
                    {props.contextCatalog.contexts.map((context) => (
                      <option key={context} value={context} />
                    ))}
                  </datalist>
                ) : null}
              </div>
            ) : null}
          </Card>

          <Card title="2. Scope and labels" description="Set namespace scope and operator-facing labels so the resulting bundle is easy to recognize later.">
            <div className="form-grid">
              <Field
                label="Namespaces (optional)"
                hint="Leave blank to scan every namespace the credentials can read."
                tipLabel="Namespaces help"
                tip={<p>Enter a comma-separated list such as payments, frontend, platform to narrow the scan.</p>}
              >
                <input
                  aria-label="Namespaces"
                  placeholder="payments, frontend"
                  value={(props.scanForm.namespaces || []).join(", ")}
                  onChange={(event) => updateForm("namespaces", splitList(event.target.value))}
                />
              </Field>
              <Field label="Cluster label (optional)" hint="If left blank, the app will derive a label from the context or API server.">
                <input
                  aria-label="Cluster label"
                  placeholder="prod-east"
                  value={props.scanForm.clusterName || ""}
                  onChange={(event) => updateForm("clusterName", event.target.value)}
                />
              </Field>
              <Field label="Environment (optional)" hint="Examples: production, staging, dev.">
                <input
                  aria-label="Environment"
                  placeholder="production"
                  value={props.scanForm.environment || ""}
                  onChange={(event) => updateForm("environment", event.target.value)}
                />
              </Field>
            </div>
          </Card>

          <Card title="3. Outputs" description="Choose where artifacts land and which operator-facing exports should be refreshed after the run.">
            <div className="form-grid">
              <Field label="Output directory" hint="All reports and bundles will be written here.">
                <div className="inline-field">
                  <input
                    aria-label="Output directory"
                    value={props.scanForm.outputDir || ""}
                    onChange={(event) => updateForm("outputDir", event.target.value)}
                  />
                  <button type="button" className="button secondary" onClick={props.onBrowseOutput} disabled={props.busy}>
                    Browse
                  </button>
                </div>
              </Field>
            </div>
            <div className="toggle-grid">
              <label className="toggle">
                <input type="checkbox" checked={Boolean(props.scanForm.summary)} onChange={(event) => updateForm("summary", event.target.checked)} />
                Executive summary
              </label>
              <label className="toggle">
                <input type="checkbox" checked={Boolean(props.scanForm.runbook)} onChange={(event) => updateForm("runbook", event.target.checked)} />
                Customer runbook
              </label>
            </div>
          </Card>

          <details className="advanced-panel">
            <summary>4. Advanced options</summary>
            <div className="advanced-panel-body">
              <div className="form-grid">
                <Field label="Profile" hint="Standard is the best default for most production scans.">
                  <select value={props.scanForm.profileName || "standard"} onChange={(event) => updateForm("profileName", event.target.value)}>
                    <option value="standard">standard</option>
                    <option value="enterprise">enterprise</option>
                    <option value="dev">dev</option>
                    <option value="airgap">airgap</option>
                  </select>
                </Field>
                <Field label="Compare baseline" hint="Optional path to a previous bundle for drift and progress comparison.">
                  <input
                    aria-label="Compare baseline"
                    placeholder="C:\\scans\\previous\\recovery-scan.json"
                    value={props.scanForm.compareTo || ""}
                    onChange={(event) => updateForm("compareTo", event.target.value)}
                  />
                </Field>
                <Field label="Recovery target">
                  <select value={props.scanForm.target || "vm"} onChange={(event) => updateForm("target", event.target.value)}>
                    <option value="vm">vm</option>
                    <option value="baremetal">baremetal</option>
                  </select>
                </Field>
                <Field label="Timeout (seconds)" hint="Increase this for slower networks or very large clusters.">
                  <input
                    aria-label="Timeout seconds"
                    type="number"
                    min={10}
                    value={props.scanForm.timeoutSeconds || 60}
                    onChange={(event) => updateForm("timeoutSeconds", Number(event.target.value))}
                  />
                </Field>
                {connectionMethod === "api_endpoint" && (
                  <>
                    <Field label="CA certificate file (optional)" hint="Use this when the API server certificate is signed by a private CA.">
                      <div className="inline-field">
                        <input
                          aria-label="CA certificate file"
                          placeholder="C:\\certs\\cluster-ca.pem"
                          value={props.scanForm.caCertPath || ""}
                          onChange={(event) => updateForm("caCertPath", event.target.value)}
                        />
                        <button type="button" className="button secondary" onClick={props.onBrowseCACert} disabled={props.busy}>
                          Browse
                        </button>
                      </div>
                    </Field>
                    <Field label="Pasted CA certificate (optional)" hint="Paste PEM content only if you do not want to reference a local CA file.">
                      <textarea
                        aria-label="Pasted CA certificate"
                        rows={6}
                        placeholder="-----BEGIN CERTIFICATE-----"
                        value={props.scanForm.caCertContent || ""}
                        onChange={(event) => updateForm("caCertContent", event.target.value)}
                      />
                    </Field>
                  </>
                )}
              </div>
              <div className="toggle-grid">
                <label className="toggle">
                  <input type="checkbox" checked={Boolean(props.scanForm.insecure)} onChange={(event) => updateForm("insecure", event.target.checked)} />
                  Skip TLS verification
                </label>
                <label className="toggle">
                  <input type="checkbox" checked={Boolean(props.scanForm.csvExport)} onChange={(event) => updateForm("csvExport", event.target.checked)} />
                  CSV exports
                </label>
                <label className="toggle">
                  <input type="checkbox" checked={Boolean(props.scanForm.redact)} onChange={(event) => updateForm("redact", event.target.checked)} />
                  Redacted outputs
                </label>
                <label className="toggle">
                  <input type="checkbox" checked={Boolean(props.scanForm.includeSecretMetadata)} onChange={(event) => updateForm("includeSecretMetadata", event.target.checked)} />
                  Secret metadata
                </label>
              </div>
            </div>
          </details>

          <Card title="5. Review and launch" description="Use the compact launch summary to sanity-check access, scope, artifacts, and output location before you run.">
            <div className="review-grid">
              <ReviewCard label="Connection" value={connectionSummary(props.scanForm)} />
              <ReviewCard label="Scope" value={scopeSummary(props.scanForm.namespaces)} />
              <ReviewCard label="Reports" value={reportSummary(props.scanForm)} />
              <ReviewCard label="Output" value={props.scanForm.outputDir || "./out"} />
            </div>

            {props.validationErrors.length ? (
              <div className="notice notice-error">
                <strong>Fix these before preflight or start:</strong>
                <ul className="notice-list">
                  {props.validationErrors.map((error) => (
                    <li key={error}>{error}</li>
                  ))}
                </ul>
              </div>
            ) : (
              <div className="notice notice-info">
                <strong>Launch posture looks good.</strong>
                <p className="muted">Run preflight to verify the selected connection and RBAC scope before collecting the full bundle.</p>
              </div>
            )}

            <div className="form-actions">
              <button type="button" className="button secondary" onClick={props.onPreflight} disabled={props.busy}>
                Run Preflight
              </button>
              <button type="button" className="button primary" onClick={props.onStartScan} disabled={props.busy}>
                Start Scan
              </button>
            </div>
          </Card>
        </div>
      </section>

      <section className="panel">
        <SectionHeader
          eyebrow="Preflight"
          title="Connection and access check"
          description="Preflight makes scope, degraded visibility, and permission blockers explicit before the full collection run starts."
        />
        {props.preflight ? (
          <div className="results-stack">
            <div className="inline-metrics">
              <MetricCard label="Can run" value={props.preflight.canRun ? "Yes" : "No"} tone={props.preflight.canRun ? "success" : "critical"} />
              <MetricCard label="Degraded" value={props.preflight.degraded ? "Yes" : "No"} tone={props.preflight.degraded ? "high" : "success"} />
              <MetricCard label="Scope" value={props.preflight.scope} />
            </div>
            {props.preflight.warnings?.length ? (
              <div className="notice notice-info">
                <strong>Operator note</strong>
                <p className="muted">{props.preflight.warnings.join(" ")}</p>
              </div>
            ) : null}
            <div className="preflight-list">
              {props.preflight.checks.map((check) => (
                <article key={check.id} className={`preflight-check ${check.status}`}>
                  <div className="preflight-check-head">
                    <div>
                      <strong>{check.title}</strong>
                      {check.resource ? <p className="muted">{titleForProbe(check.scope, check.resource)}</p> : null}
                    </div>
                    <Badge tone={toneForCheck(check.status)}>{check.status}</Badge>
                  </div>
                  <p>{check.detail}</p>
                  {check.hint ? <p className="muted">{check.hint}</p> : null}
                  {check.commands?.length ? <p className="muted">Check with: {check.commands[0]}</p> : null}
                  {check.manifest ? <code className="mono-block">{check.manifest}</code> : null}
                </article>
              ))}
            </div>
          </div>
        ) : (
          <div className="empty-state">
            <p className="muted">Run preflight before the scan to validate remote access, namespace scope, and any reduced-visibility caveats.</p>
          </div>
        )}
      </section>
    </section>
  );
}

function connectionSummary(scanForm: ScanRequest) {
  switch (scanForm.connectionMethod || "current") {
    case "kubeconfig_file":
      return scanForm.kubeconfigPath || "Selected kubeconfig file";
    case "kubeconfig_inline":
      return scanForm.contextName ? `Pasted kubeconfig · ${scanForm.contextName}` : "Pasted kubeconfig";
    case "api_endpoint":
      return scanForm.apiServerEndpoint || "Direct API endpoint";
    default:
      return scanForm.contextName ? `Current login · ${scanForm.contextName}` : "Current login";
  }
}

function scopeSummary(namespaces?: string[]) {
  return namespaces?.length ? namespaces.join(", ") : "all namespaces";
}

function reportSummary(scanForm: ScanRequest) {
  const enabled = [];
  if (scanForm.summary) {
    enabled.push("summary");
  }
  if (scanForm.runbook) {
    enabled.push("runbook");
  }
  if (scanForm.csvExport) {
    enabled.push("csv");
  }
  if (scanForm.redact) {
    enabled.push("redacted");
  }
  return enabled.length ? enabled.join(", ") : "bundle only";
}

function titleForProbe(scope?: string, resource?: string) {
  if (!resource) {
    return "";
  }
  return `${scope === "cluster" ? "Cluster-scope" : "Namespace-scope"} access for ${resource}`;
}

function toneForCheck(status: "pass" | "warn" | "fail") {
  switch (status) {
    case "pass":
      return "pass";
    case "warn":
      return "warn";
    default:
      return "fail";
  }
}
