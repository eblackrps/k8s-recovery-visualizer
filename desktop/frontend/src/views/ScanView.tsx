import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import type { ConnectionMethod, ContextCatalog, PreflightCheck, PreflightReport, ScanRequest } from "../lib/types";
import {
  describeAuthMode,
  describeTrustMode,
  type ScanFieldName,
  type ScanReviewTone,
  type ScanValidation,
} from "../lib/scanForm";
import { Badge, Card, CodeBlock, Field, KeyValueGrid, MetricCard, ReviewCard, SectionHeader, splitList } from "../components/ui";

const connectionModes: Array<{ value: ConnectionMethod; label: string; description: string }> = [
  { value: "current", label: "Current login", description: "Best for desktops and jumpboxes where kubectl or KUBECONFIG already works." },
  { value: "kubeconfig_file", label: "Kubeconfig file", description: "Choose a kubeconfig from disk and optionally select the context to scan." },
  { value: "kubeconfig_inline", label: "Paste kubeconfig", description: "Paste kubeconfig content directly when you do not want to rely on local files." },
  { value: "api_endpoint", label: "API endpoint", description: "Enter the control-plane host or URL directly and authenticate with a bearer token." },
];

export function ScanView(props: {
  busy: boolean;
  scanForm: ScanRequest;
  setScanForm: (updater: (current: ScanRequest) => ScanRequest) => void;
  preflight: PreflightReport | null;
  contextCatalog: ContextCatalog | null;
  detectingContexts: boolean;
  validation: ScanValidation;
  showValidationErrors: boolean;
  validationRequest: { version: number; field?: ScanFieldName };
  insecureAcknowledged: boolean;
  onSetInsecureAcknowledged: (value: boolean) => void;
  onPreflight: () => void;
  onStartScan: () => void;
  onDetectContexts: () => void;
  onBrowseOutput: () => void;
  onBrowseKubeconfig: () => void;
  onBrowseCACert: () => void;
}) {
  const [tokenVisible, setTokenVisible] = useState(false);
  const connectionMethod = props.scanForm.connectionMethod || "current";
  const supportsContextDiscovery = connectionMethod !== "api_endpoint";
  const trustMode = describeTrustMode(props.scanForm);
  const authMode = describeAuthMode(props.scanForm);

  useEffect(() => {
    if (connectionMethod !== "api_endpoint") {
      setTokenVisible(false);
    }
  }, [connectionMethod]);

  useEffect(() => {
    if (!props.validationRequest.version || !props.validationRequest.field) {
      return;
    }
    const target = document.querySelector<HTMLElement>(`[data-scan-field="${props.validationRequest.field}"]`);
    if (!target) {
      return;
    }
    target.focus();
    if (typeof target.scrollIntoView === "function") {
      target.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }, [props.validationRequest.field, props.validationRequest.version]);

  const updateForm = <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => {
    props.setScanForm((current) => ({ ...current, [key]: value }));
  };
  const fieldError = (field: ScanFieldName) => (props.showValidationErrors ? props.validation.fieldErrors[field] : undefined);
  const fieldWarning = (field: ScanFieldName) => props.validation.fieldWarnings[field];
  const reviewCards = buildReviewCards(props.scanForm, authMode, trustMode, props.validation.riskFlags);

  return (
    <section className="page-grid scan-grid">
      <section className="panel">
        <SectionHeader
          eyebrow="New Scan"
          title="Remote cluster scan"
          description="Use the simplest connection that already works on this machine. The setup stays compact, but the rail and review summary now explain what the scanner needs before you launch."
        />
        <div className="scan-stack">
          <Card title="1. Connection" description="Choose the access path, then make endpoint, credentials, and trust explicit before preflight.">
            <div className="connection-modes" role="radiogroup" aria-label="Connection method">
              {connectionModes.map((mode) => (
                <label key={mode.value} className={`mode-card ${connectionMethod === mode.value ? "is-active" : ""}`}>
                  <input type="radio" name="connectionMethod" checked={connectionMethod === mode.value} onChange={() => updateForm("connectionMethod", mode.value)} />
                  <strong>{mode.label}</strong>
                  <span>{mode.description}</span>
                </label>
              ))}
            </div>
            {connectionMethod === "api_endpoint"
              ? renderApiConnection({
                  busy: props.busy,
                  scanForm: props.scanForm,
                  tokenVisible,
                  setTokenVisible,
                  fieldError,
                  fieldWarning,
                  insecureAcknowledged: props.insecureAcknowledged,
                  onSetInsecureAcknowledged: props.onSetInsecureAcknowledged,
                  onBrowseCACert: props.onBrowseCACert,
                  updateForm,
                })
              : renderStandardConnection({ busy: props.busy, connectionMethod, scanForm: props.scanForm, contextCatalog: props.contextCatalog, fieldError, onBrowseKubeconfig: props.onBrowseKubeconfig, updateForm })}
            {supportsContextDiscovery ? renderContextHelper(props, updateForm) : null}
          </Card>
          <Card title="2. Scope and labels" description="Set namespace scope and operator-facing labels so the resulting bundle is easy to recognize later.">
            <div className="form-grid">
              <Field label="Namespaces (optional)" hint="Leave blank to scan every namespace the credentials can read.">
                <input aria-label="Namespaces" placeholder="payments, frontend" value={(props.scanForm.namespaces || []).join(", ")} onChange={(event) => updateForm("namespaces", splitList(event.target.value))} />
              </Field>
              <Field label="Cluster label (optional)" hint="If left blank, the app will derive a label from the context or API server.">
                <input aria-label="Cluster label" placeholder="prod-east" value={props.scanForm.clusterName || ""} onChange={(event) => updateForm("clusterName", event.target.value)} />
              </Field>
              <Field label="Environment (optional)" hint="Examples: production, staging, dev.">
                <input aria-label="Environment" placeholder="production" value={props.scanForm.environment || ""} onChange={(event) => updateForm("environment", event.target.value)} />
              </Field>
            </div>
          </Card>
          <Card title="3. Outputs" description="Choose where artifacts land and which operator-facing exports should be refreshed after the run.">
            <div className="form-grid">
              <Field label="Output directory" hint="All reports and bundles will be written here." error={fieldError("outputDir")}>
                <div className="inline-field">
                  <input aria-label="Output directory" data-scan-field="outputDir" value={props.scanForm.outputDir || ""} onChange={(event) => updateForm("outputDir", event.target.value)} />
                  <button type="button" className="button secondary" onClick={props.onBrowseOutput} disabled={props.busy}>Browse</button>
                </div>
              </Field>
            </div>
            <div className="toggle-grid">
              <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.summary)} onChange={(event) => updateForm("summary", event.target.checked)} />Executive summary</label>
              <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.runbook)} onChange={(event) => updateForm("runbook", event.target.checked)} />Customer runbook</label>
            </div>
          </Card>
          {renderAdvancedOptions({ connectionMethod, scanForm: props.scanForm, fieldError, updateForm })}
          <Card title="5. Review and launch" description="Sanity-check endpoint, auth path, trust mode, scope, exports, and obvious risk flags before you touch the cluster.">
            <div className="review-grid">{reviewCards.map((card) => <ReviewCard key={card.label} label={card.label} value={card.value} detail={card.detail} tone={card.tone} />)}</div>
            {props.showValidationErrors && props.validation.errors.length ? (
              <div className="notice notice-error"><strong>Fix these before preflight or start:</strong><ul className="notice-list">{props.validation.errors.map((error) => <li key={error}>{error}</li>)}</ul></div>
            ) : props.validation.riskFlags.length ? (
              <div className="notice notice-warning"><strong>Review these operator risks before launch</strong><ul className="notice-list">{props.validation.riskFlags.map((flag) => <li key={flag}>{flag}</li>)}</ul></div>
            ) : (
              <div className="notice notice-info"><strong>Launch posture looks good.</strong><p className="muted">Run preflight to verify the selected connection, trust path, and RBAC scope before collecting the full bundle.</p></div>
            )}
            <div className="form-actions">
              <button type="button" className="button secondary" onClick={props.onPreflight} disabled={props.busy}>Run Preflight</button>
              <button type="button" className="button primary" onClick={props.onStartScan} disabled={props.busy}>Start Scan</button>
            </div>
          </Card>
        </div>
      </section>
      <section className="panel">
        <SectionHeader
          eyebrow={props.preflight ? "Preflight" : "Assistant"}
          title={props.preflight ? "Connection and access check" : assistantTitle(connectionMethod)}
          description={props.preflight ? "Preflight makes transport, auth, scope, and reduced-visibility caveats explicit before the full collection run starts." : "The rail adapts to the selected connection method so first-time operators can prepare the endpoint, credentials, and trust chain before preflight."}
        />
        {props.preflight ? <PreflightPanel preflight={props.preflight} connectionMethod={connectionMethod} insecure={Boolean(props.scanForm.insecure)} /> : <ConnectionAssistant connectionMethod={connectionMethod} trustMode={trustMode} insecure={Boolean(props.scanForm.insecure)} />}
      </section>
    </section>
  );
}

function renderApiConnection(props: {
  busy: boolean;
  scanForm: ScanRequest;
  tokenVisible: boolean;
  setTokenVisible: (value: boolean | ((current: boolean) => boolean)) => void;
  fieldError: (field: ScanFieldName) => string | undefined;
  fieldWarning: (field: ScanFieldName) => string | undefined;
  insecureAcknowledged: boolean;
  onSetInsecureAcknowledged: (value: boolean) => void;
  onBrowseCACert: () => void;
  updateForm: <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => void;
}) {
  return (
    <div className="api-connection-stack">
      <div className="form-grid api-connection-grid">
        <Field
          label="API server host or URL"
          hint="Accepted examples: https://cluster.example.net:6443, 10.0.0.15:6443, or control-plane.prod-east.example.net:6443."
          warning={props.fieldWarning("apiServerEndpoint")}
          error={props.fieldError("apiServerEndpoint")}
        >
          <input
            aria-label="API server host or URL"
            data-scan-field="apiServerEndpoint"
            placeholder="10.0.0.15:6443 or https://cluster.example.net:6443"
            value={props.scanForm.apiServerEndpoint || ""}
            onChange={(event) => props.updateForm("apiServerEndpoint", event.target.value)}
          />
        </Field>
        <Field
          label="Bearer token"
          hint="Paste the raw token value. Leading Bearer text, extra spaces, and accidental line breaks are removed automatically."
          error={props.fieldError("bearerToken")}
        >
          <div className="inline-field token-field">
            <input
              aria-label="Bearer token"
              data-scan-field="bearerToken"
              type={props.tokenVisible ? "text" : "password"}
              placeholder="Paste a Kubernetes API bearer token"
              value={props.scanForm.bearerToken || ""}
              onChange={(event) => props.updateForm("bearerToken", event.target.value)}
              autoComplete="off"
              autoCorrect="off"
              autoCapitalize="off"
              spellCheck={false}
            />
            <button type="button" className="button secondary quiet" onClick={() => props.setTokenVisible((current) => !current)}>{props.tokenVisible ? "Hide" : "Show"}</button>
            <button type="button" className="button secondary quiet" onClick={() => props.updateForm("bearerToken", "")}>Clear</button>
          </div>
        </Field>
      </div>
      <section className="connection-trust-panel" aria-label="Connection trust">
        <SectionHeader compact title="Connection trust" description="Decide whether the server certificate should be verified with machine trust, an explicit internal CA, or a temporary insecure override." />
        <div className="trust-guide-grid">
          <article className="trust-guide-card"><strong>Publicly trusted cert</strong><p className="muted">Leave CA inputs empty when the API server certificate chains to a CA already trusted on this machine.</p></article>
          <article className="trust-guide-card"><strong>Private or internal CA</strong><p className="muted">Add the issuing CA as a file or pasted PEM when the cluster uses an internal or self-signed trust chain.</p></article>
          <article className="trust-guide-card warning"><strong>Skip verification</strong><p className="muted">Use this only as a temporary workaround in a trusted environment. It should not be the steady-state path.</p></article>
        </div>
        <div className="form-grid">
          <Field label="CA certificate file (optional)" hint="Point to the PEM that issued the API server certificate when the cluster uses a private CA.">
            <div className="inline-field">
              <input aria-label="CA certificate file" placeholder="C:\\certs\\cluster-ca.pem" value={props.scanForm.caCertPath || ""} onChange={(event) => props.updateForm("caCertPath", event.target.value)} />
              <button type="button" className="button secondary" onClick={props.onBrowseCACert} disabled={props.busy}>Browse</button>
            </div>
          </Field>
          <Field label="Pasted CA certificate (optional)" hint="Paste PEM content only when you do not want to reference a local CA file.">
            <textarea aria-label="Pasted CA certificate" rows={6} placeholder="-----BEGIN CERTIFICATE-----" value={props.scanForm.caCertContent || ""} onChange={(event) => props.updateForm("caCertContent", event.target.value)} spellCheck={false} />
          </Field>
        </div>
        {props.fieldWarning("caTrust") ? <div className="notice notice-warning compact"><strong>Trust check</strong><p className="muted">{props.fieldWarning("caTrust")}</p></div> : null}
        <div className="toggle-stack">
          <label className={`toggle-card ${props.scanForm.insecure ? "is-warning" : ""}`}>
            <input type="checkbox" checked={Boolean(props.scanForm.insecure)} onChange={(event) => props.updateForm("insecure", event.target.checked)} />
            <div><strong>Skip TLS verification</strong><p className="muted">Only use this when you trust the network path and need a temporary workaround for certificate validation.</p></div>
          </label>
          {props.scanForm.insecure ? (
            <label className={`toggle-card acknowledgement-card ${props.fieldError("insecureAcknowledgement") ? "is-error" : "is-warning"}`}>
              <input type="checkbox" data-scan-field="insecureAcknowledgement" checked={props.insecureAcknowledged} onChange={(event) => props.onSetInsecureAcknowledged(event.target.checked)} />
              <div>
                <strong>I understand this disables certificate verification</strong>
                <p className="muted">This should stay confined to trusted, short-lived troubleshooting or recovery preparation scenarios.</p>
                {props.fieldError("insecureAcknowledgement") ? <small className="field-message error">{props.fieldError("insecureAcknowledgement")}</small> : null}
              </div>
            </label>
          ) : null}
        </div>
      </section>
    </div>
  );
}

function renderStandardConnection(props: {
  busy: boolean;
  connectionMethod: ConnectionMethod;
  scanForm: ScanRequest;
  contextCatalog: ContextCatalog | null;
  fieldError: (field: ScanFieldName) => string | undefined;
  onBrowseKubeconfig: () => void;
  updateForm: <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => void;
}) {
  return (
    <div className="form-grid">
      {props.connectionMethod === "current" ? (
        <Field label="Context (optional)" hint="Leave blank to use the current kubectl context.">
          <input aria-label="Context" data-scan-field="contextName" list={props.contextCatalog?.contexts?.length ? "scan-context-options" : undefined} placeholder="Leave blank to use the current context" value={props.scanForm.contextName || ""} onChange={(event) => props.updateForm("contextName", event.target.value)} />
        </Field>
      ) : null}
      {props.connectionMethod === "kubeconfig_file" ? (
        <>
          <Field label="Kubeconfig file" hint="Use this when the machine should not rely on the default kubectl config path." error={props.fieldError("kubeconfigPath")}>
            <div className="inline-field">
              <input aria-label="Kubeconfig file" data-scan-field="kubeconfigPath" placeholder="C:\\Users\\you\\.kube\\config" value={props.scanForm.kubeconfigPath || ""} onChange={(event) => props.updateForm("kubeconfigPath", event.target.value)} />
              <button type="button" className="button secondary" onClick={props.onBrowseKubeconfig} disabled={props.busy}>Browse</button>
            </div>
          </Field>
          <Field label="Context (optional)" hint="Leave blank to use the kubeconfig current context.">
            <input aria-label="Context" data-scan-field="contextName" list={props.contextCatalog?.contexts?.length ? "scan-context-options" : undefined} placeholder="prod-east-admin" value={props.scanForm.contextName || ""} onChange={(event) => props.updateForm("contextName", event.target.value)} />
          </Field>
        </>
      ) : null}
      {props.connectionMethod === "kubeconfig_inline" ? (
        <>
          <Field label="Pasted kubeconfig" hint="Useful when operators receive kubeconfig content from a secure vault or ticket." error={props.fieldError("kubeconfigContent")}>
            <textarea aria-label="Pasted kubeconfig" data-scan-field="kubeconfigContent" rows={10} placeholder={"apiVersion: v1\nclusters: ..."} value={props.scanForm.kubeconfigContent || ""} onChange={(event) => props.updateForm("kubeconfigContent", event.target.value)} />
          </Field>
          <Field label="Context (optional)" hint="Leave blank to use the kubeconfig current context.">
            <input aria-label="Context" data-scan-field="contextName" list={props.contextCatalog?.contexts?.length ? "scan-context-options" : undefined} placeholder="prod-east-admin" value={props.scanForm.contextName || ""} onChange={(event) => props.updateForm("contextName", event.target.value)} />
          </Field>
        </>
      ) : null}
    </div>
  );
}

function renderContextHelper(
  props: {
    busy: boolean;
    scanForm: ScanRequest;
    contextCatalog: ContextCatalog | null;
    detectingContexts: boolean;
    onDetectContexts: () => void;
  },
  updateForm: <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => void,
) {
  return (
    <div className="context-helper">
      <div className="inline-actions">
        <button type="button" className="button secondary quiet" onClick={props.onDetectContexts} disabled={props.busy || props.detectingContexts}>{props.detectingContexts ? "Loading Contexts..." : "Detect Contexts"}</button>
        {props.contextCatalog?.currentContext && props.scanForm.contextName !== props.contextCatalog.currentContext ? <button type="button" className="button secondary quiet" onClick={() => updateForm("contextName", props.contextCatalog?.currentContext || "")} disabled={props.busy}>Use {props.contextCatalog.currentContext}</button> : null}
      </div>
      <p className="muted context-helper-status">
        {props.contextCatalog ? props.contextCatalog.contexts?.length ? `Found ${props.contextCatalog.contexts.length} context${props.contextCatalog.contexts.length === 1 ? "" : "s"} from ${props.contextCatalog.source || "this connection"}${props.contextCatalog.currentContext ? `. Current: ${props.contextCatalog.currentContext}.` : "."}` : `No named contexts were found for ${props.contextCatalog.source || "this connection"}.` : "Load contexts to populate suggestions before preflight."}
      </p>
      {props.contextCatalog?.contexts?.length ? <datalist id="scan-context-options">{props.contextCatalog.contexts.map((context) => <option key={context} value={context} />)}</datalist> : null}
    </div>
  );
}

function renderAdvancedOptions(props: {
  connectionMethod: ConnectionMethod;
  scanForm: ScanRequest;
  fieldError: (field: ScanFieldName) => string | undefined;
  updateForm: <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => void;
}) {
  return (
    <details className="advanced-panel">
      <summary>4. Advanced options</summary>
      <div className="advanced-panel-body">
        <div className="form-grid">
          <Field label="Profile" hint="Standard is the best default for most production scans."><select value={props.scanForm.profileName || "standard"} onChange={(event) => props.updateForm("profileName", event.target.value)}><option value="standard">standard</option><option value="enterprise">enterprise</option><option value="dev">dev</option><option value="airgap">airgap</option></select></Field>
          <Field label="Compare baseline" hint="Optional path to a previous bundle for drift and progress comparison."><input aria-label="Compare baseline" placeholder="C:\\scans\\previous\\recovery-scan.json" value={props.scanForm.compareTo || ""} onChange={(event) => props.updateForm("compareTo", event.target.value)} /></Field>
          <Field label="Recovery target"><select value={props.scanForm.target || "vm"} onChange={(event) => props.updateForm("target", event.target.value)}><option value="vm">vm</option><option value="baremetal">baremetal</option></select></Field>
          <Field label="Timeout (seconds)" hint="Increase this for slower networks or very large clusters." error={props.fieldError("timeoutSeconds")}><input aria-label="Timeout seconds" data-scan-field="timeoutSeconds" type="number" min={10} value={props.scanForm.timeoutSeconds || 60} onChange={(event) => props.updateForm("timeoutSeconds", Number(event.target.value))} /></Field>
        </div>
        <div className="toggle-grid">
          {props.connectionMethod !== "api_endpoint" ? <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.insecure)} onChange={(event) => props.updateForm("insecure", event.target.checked)} />Skip TLS verification</label> : null}
          <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.csvExport)} onChange={(event) => props.updateForm("csvExport", event.target.checked)} />CSV exports</label>
          <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.redact)} onChange={(event) => props.updateForm("redact", event.target.checked)} />Redacted outputs</label>
          <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.includeSecretMetadata)} onChange={(event) => props.updateForm("includeSecretMetadata", event.target.checked)} />Secret metadata</label>
        </div>
      </div>
    </details>
  );
}
function buildReviewCards(
  scanForm: ScanRequest,
  authMode: { label: string; detail: string; tone: ScanReviewTone },
  trustMode: { label: string; detail: string; tone: ScanReviewTone },
  riskFlags: string[],
) {
  const connectionMethod = scanForm.connectionMethod || "current";
  const cards: Array<{ label: string; value: string; detail?: string; tone?: ScanReviewTone }> = [];
  if (connectionMethod === "api_endpoint") {
    cards.push({ label: "Endpoint target", value: scanForm.apiServerEndpoint || "Missing endpoint", detail: "Use the Kubernetes API server address only.", tone: scanForm.apiServerEndpoint ? "neutral" : "critical" });
    cards.push({ label: "Auth mode", value: authMode.label, detail: authMode.detail, tone: authMode.tone });
    cards.push({ label: "TLS trust", value: trustMode.label, detail: trustMode.detail, tone: trustMode.tone });
  } else {
    cards.push({ label: "Connection", value: connectionSummary(scanForm), detail: "Credential and trust handling come from the current login or kubeconfig.", tone: "neutral" });
  }
  cards.push({ label: "Namespace scope", value: scopeSummary(scanForm.namespaces), detail: scanForm.namespaces?.length ? "Only the listed namespaces will be collected." : "All readable namespaces are in scope.", tone: scanForm.namespaces?.length ? "neutral" : "high" });
  cards.push({ label: "Compare baseline", value: scanForm.compareTo ? "Loaded" : "None selected", detail: scanForm.compareTo || "Add a previous bundle if you want delta and regression analysis in Results.", tone: scanForm.compareTo ? "success" : "neutral" });
  cards.push({ label: "Exports", value: reportSummary(scanForm), detail: "These artifacts will be refreshed after the run completes.", tone: "neutral" });
  cards.push({ label: "Output", value: scanForm.outputDir || "Missing output path", detail: "Bundles and refreshed exports land here.", tone: scanForm.outputDir ? "neutral" : "critical" });
  cards.push({ label: "Risk flags", value: riskFlags.length ? `${riskFlags.length} attention point${riskFlags.length === 1 ? "" : "s"}` : "None detected", detail: riskFlags.length ? riskFlags.join(" ") : "No obvious transport, scope, or trust risks are flagged right now.", tone: riskFlags.length ? "high" : "success" });
  return cards;
}

function assistantTitle(connectionMethod: ConnectionMethod) {
  switch (connectionMethod) {
    case "api_endpoint":
      return "API endpoint guide";
    case "kubeconfig_file":
      return "Kubeconfig file guide";
    case "kubeconfig_inline":
      return "Pasted kubeconfig guide";
    default:
      return "Connection assistant";
  }
}

function ConnectionAssistant(props: { connectionMethod: ConnectionMethod; trustMode: { label: string; detail: string; tone: ScanReviewTone }; insecure: boolean }) {
  if (props.connectionMethod === "api_endpoint") {
    return (
      <div className="assistant-stack">
        {props.insecure ? <div className="notice notice-warning compact"><strong>Skip-TLS is enabled</strong><p className="muted">Preflight can still run, but the app will not verify the API server certificate. Treat this as a temporary, trusted-environment workaround.</p></div> : null}
        <AssistantSection title="When to use API endpoint mode" body="Use this when you know the Kubernetes API server URL and can obtain a bearer token directly. If the cluster depends on exec plugins, cloud auth helpers, SSO prompts, or client certificates, kubeconfig mode is usually the better fit." />
        <AssistantSection title="Find the API server URL" body="You need the control-plane API address, not an ingress or an application URL.">
          <ul className="assistant-list">
            <li>From an existing kubectl setup, print the active cluster server directly.</li>
            <li>From platform docs or cluster inventory, use the documented control-plane endpoint.</li>
            <li>When appropriate, use a control-plane host or IP with port 6443.</li>
          </ul>
          <CodeBlock label="From the current kubeconfig" code={"kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'"} />
          <CodeBlock label="Direct endpoint example" code="https://control-plane.prod-east.example.net:6443" />
          <CodeBlock label="Host or IP example" code="10.0.0.15:6443" />
        </AssistantSection>
        <AssistantSection title="Get a bearer token safely" body="Prefer a short-lived token for a least-privilege service account. Do not default to cluster-admin. If you already have a suitable service account, create a fresh token for it instead of reusing long-lived static credentials.">
          <CodeBlock label="Create a short-lived token" code="kubectl create token <service-account> --namespace <namespace>" />
          <CodeBlock label="Example with a duration" code="kubectl create token k8v-scan --namespace k8v-ops --duration 10m" />
          <p className="muted assistant-footnote">Bind the service account to only the RBAC it needs. Preflight will show missing access with kubectl auth can-i guidance and a least-privilege manifest starting point when possible.</p>
        </AssistantSection>
        <AssistantSection title="Choose the trust path" body={`Current trust mode: ${props.trustMode.label}. Use system trust for publicly trusted certs, a CA file or pasted PEM for private/internal CAs, and skip verification only as a last-resort temporary workaround.`} />
        <AssistantSection title="What preflight will validate" body="Preflight will check transport reachability, credential loading, namespace scope, and the RBAC needed to collect the bundle. If access is incomplete, the assistant will separate blockers from degraded-but-runnable gaps." />
      </div>
    );
  }
  if (props.connectionMethod === "kubeconfig_file") {
    return <div className="assistant-stack"><AssistantSection title="Kubeconfig file workflow" body="Use this when the machine should not rely on the default kubectl config path. Pick the file, then optionally pin the context before you run preflight."><CodeBlock label="Show contexts in a kubeconfig" code="kubectl config get-contexts -o=name --kubeconfig /path/to/config" /></AssistantSection></div>;
  }
  if (props.connectionMethod === "kubeconfig_inline") {
    return <div className="assistant-stack"><AssistantSection title="Pasted kubeconfig workflow" body="Use this when operators receive kubeconfig content from a secure vault, ticket, or handoff and do not want to rely on a local file path."><CodeBlock label="Capture a raw kubeconfig for handoff" code="kubectl config view --raw" /></AssistantSection></div>;
  }
  return <div className="assistant-stack"><AssistantSection title="Current login workflow" body="Use this when kubectl already reaches the cluster from this desktop or jumpbox. Detect contexts if you want suggestions before preflight, then run preflight to confirm namespace scope and RBAC."><CodeBlock label="Quick sanity check" code="kubectl config current-context" /></AssistantSection></div>;
}

function AssistantSection(props: { title: string; body: string; children?: ReactNode }) {
  return (
    <section className="assistant-section">
      <div className="section-header compact"><div className="section-heading"><h4>{props.title}</h4><p className="muted section-copy">{props.body}</p></div></div>
      {props.children}
    </section>
  );
}
function PreflightPanel(props: { preflight: PreflightReport; connectionMethod: ConnectionMethod; insecure: boolean }) {
  const categories = groupPreflightChecks(props.preflight.checks);
  const blockers = props.preflight.checks.filter((check) => check.status === "fail").length;
  const warnings = props.preflight.checks.filter((check) => check.status === "warn").length;
  return (
    <div className="results-stack">
      {props.insecure ? <div className="notice notice-warning compact"><strong>Transport trust warning</strong><p className="muted">This preflight used insecure TLS mode. Re-enable certificate verification or add the proper CA material before treating this as a normal operating path.</p></div> : null}
      <div className="inline-metrics">
        <MetricCard label="Can run" value={props.preflight.canRun ? "Yes" : "No"} tone={props.preflight.canRun ? "success" : "critical"} />
        <MetricCard label="Blockers" value={blockers} tone={blockers ? "critical" : "success"} />
        <MetricCard label="Warnings" value={warnings} tone={warnings ? "high" : "success"} />
        <MetricCard label="Scope" value={props.preflight.scope} />
      </div>
      <div className={`notice ${blockers ? "notice-error" : warnings || props.preflight.degraded ? "notice-warning" : "notice-info"}`}>
        <strong>Top next action</strong>
        <p className="muted">{preflightNextAction(props.preflight, props.connectionMethod, props.insecure)}</p>
      </div>
      {props.preflight.warnings?.length ? <div className="notice notice-info"><strong>Operator note</strong><ul className="notice-list">{props.preflight.warnings.map((warning) => <li key={warning}>{warning}</li>)}</ul></div> : null}
      <div className="preflight-category-stack">
        {categories.map((category) => (
          <section key={category.id} className="preflight-category">
            <div className="section-header compact">
              <div className="section-heading"><h4>{category.title}</h4><p className="muted section-copy">{category.description}</p></div>
              <div className="preflight-category-badges"><Badge tone={category.blockers ? "fail" : category.warnings ? "warn" : "pass"}>{category.blockers ? `${category.blockers} blocker${category.blockers === 1 ? "" : "s"}` : category.warnings ? `${category.warnings} warning${category.warnings === 1 ? "" : "s"}` : "ready"}</Badge></div>
            </div>
            <div className="preflight-list">
              {category.checks.map((check) => (
                <article key={check.id} className={`preflight-check ${check.status}`}>
                  <div className="preflight-check-head">
                    <div>
                      <strong>{check.title}</strong>
                      {check.resource ? <p className="muted">{titleForProbe(check.scope, check.resource)}</p> : null}
                    </div>
                    <div className="preflight-check-badges">
                      <Badge tone={toneForCheck(check.status)}>{labelForCheck(check.status)}</Badge>
                      <Badge tone={check.required ? "accent" : "neutral"}>{check.required ? "required" : "optional"}</Badge>
                    </div>
                  </div>
                  <p>{check.detail}</p>
                  {check.hint ? <p className="muted preflight-hint">{check.hint}</p> : null}
                  <KeyValueGrid items={preflightFacts(check)} />
                  {check.commands?.length ? <div className="preflight-code-list">{check.commands.map((command, index) => <CodeBlock key={`${check.id}-command-${index}`} label={index === 0 ? "Check command" : "Additional command"} code={command} />)}</div> : null}
                  {check.manifest ? <CodeBlock label="Least-privilege starting point" code={check.manifest} /> : null}
                </article>
              ))}
            </div>
          </section>
        ))}
      </div>
    </div>
  );
}

function reportSummary(scanForm: ScanRequest) {
  const enabled = [];
  if (scanForm.summary) enabled.push("summary");
  if (scanForm.runbook) enabled.push("runbook");
  if (scanForm.csvExport) enabled.push("csv");
  if (scanForm.redact) enabled.push("redacted");
  return enabled.length ? enabled.join(", ") : "bundle only";
}

function groupPreflightChecks(checks: PreflightCheck[]) {
  const groups = [
    { id: "transport", title: "Transport", description: "Endpoint reachability, TLS path, and low-level server contact.", checks: [] as PreflightCheck[] },
    { id: "auth", title: "Auth", description: "Credential loading, token handling, and login context checks.", checks: [] as PreflightCheck[] },
    { id: "rbac", title: "RBAC", description: "Read permissions and least-privilege gaps that affect collection depth.", checks: [] as PreflightCheck[] },
    { id: "scope", title: "Scope", description: "Namespace targeting and whether the scan is operating at the intended boundary.", checks: [] as PreflightCheck[] },
    { id: "collection", title: "Collection readiness", description: "Collectors that can run now versus data surfaces that will be degraded or skipped.", checks: [] as PreflightCheck[] },
  ];
  for (const check of checks) {
    const target = groups.find((group) => group.id === categoryForCheck(check)) || groups[groups.length - 1];
    target.checks.push(check);
  }
  return groups.map((group) => ({ ...group, blockers: group.checks.filter((check) => check.status === "fail").length, warnings: group.checks.filter((check) => check.status === "warn").length })).filter((group) => group.checks.length > 0);
}

function categoryForCheck(check: PreflightCheck) {
  const text = `${check.id} ${check.title} ${check.detail} ${check.hint || ""} ${check.resource || ""}`.toLowerCase();
  if (text.includes("api") || text.includes("endpoint") || text.includes("tls") || text.includes("certificate") || text.includes("reachability")) return "transport";
  if (text.includes("credential") || text.includes("token") || text.includes("login") || text.includes("context")) return "auth";
  if (check.resource || check.commands?.length || check.manifest || text.includes("permission") || text.includes("rbac") || text.includes("access")) return "rbac";
  if (text.includes("namespace") || text.includes("scope")) return "scope";
  return "collection";
}

function preflightFacts(check: PreflightCheck): Array<[string, string]> {
  return [check.scope ? ["Scope", check.scope === "cluster" ? "Cluster-scope" : "Namespace-scope"] : null, check.resource ? ["Resource", check.resource] : null].filter(Boolean) as Array<[string, string]>;
}

function preflightNextAction(preflight: PreflightReport, connectionMethod: ConnectionMethod, insecure: boolean) {
  if (insecure && connectionMethod === "api_endpoint") return "TLS verification is currently disabled. Add the proper CA material or move back to verified TLS before treating this as a standard operating path.";
  const firstBlocker = preflight.checks.find((check) => check.status === "fail");
  if (firstBlocker) return firstBlocker.hint || firstBlocker.detail;
  const firstWarning = preflight.checks.find((check) => check.status === "warn");
  if (firstWarning) return firstWarning.hint || firstWarning.detail;
  if (preflight.degraded) return "The scan can run, but at least one collector will operate with reduced visibility. Review the warning categories before launch.";
  return "Transport, credential, and collection checks look ready for a full scan.";
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
  return namespaces?.length ? namespaces.join(", ") : "All readable namespaces";
}

function titleForProbe(scope?: string, resource?: string) {
  if (!resource) return "";
  return `${scope === "cluster" ? "Cluster-scope" : "Namespace-scope"} access for ${resource}`;
}

function labelForCheck(status: "pass" | "warn" | "fail") {
  switch (status) {
    case "pass":
      return "ready";
    case "warn":
      return "degraded";
    default:
      return "blocked";
  }
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
