import { useEffect, useId, useState } from "react";
import type { ChangeEvent, DragEvent, ReactNode } from "react";
import type {
  ConnectionAdvisor,
  ConnectionMethod,
  ConnectionTestReport,
  ContextCatalog,
  KubeconfigInspection,
  PreflightCheck,
  PreflightReport,
  ScanRequest,
  ScanStage,
} from "../lib/types";
import {
  describeAuthMode,
  describeTrustMode,
  type ScanFieldName,
  type ScanReviewTone,
  type ScanValidation,
} from "../lib/scanForm";
import {
  Badge,
  Card,
  CodeBlock,
  Field,
  KeyValueGrid,
  MetricCard,
  ReadinessList,
  ReviewCard,
  SectionHeader,
  Stepper,
  splitList,
} from "../components/ui";

const connectionChoices: Record<ConnectionMethod, { label: string; description: string; eyebrow: string }> = {
  current: {
    label: "Use existing access",
    description: "Best when kubectl or the default kubeconfig already works on this desktop or jumpbox.",
    eyebrow: "Recommended",
  },
  kubeconfig_file: {
    label: "Load kubeconfig file",
    description: "Bring a kubeconfig from disk. K8V validates file contents, not the filename or extension.",
    eyebrow: "Bring access",
  },
  kubeconfig_inline: {
    label: "Paste kubeconfig",
    description: "Paste raw kubeconfig YAML directly when you do not want to rely on a local file path.",
    eyebrow: "Bring access",
  },
  api_endpoint: {
    label: "Use API endpoint directly",
    description: "Manual mode for entering the API server URL, bearer token, and TLS trust settings yourself.",
    eyebrow: "Advanced",
  },
};

export function ScanView(props: {
  busy: boolean;
  scanForm: ScanRequest;
  setScanForm: (updater: (current: ScanRequest) => ScanRequest) => void;
  scanStage: ScanStage;
  onSetScanStage: (stage: ScanStage) => void;
  connectionAdvisor: ConnectionAdvisor | null;
  kubeconfigInspection: KubeconfigInspection | null;
  connectionTest: ConnectionTestReport | null;
  preflight: PreflightReport | null;
  contextCatalog: ContextCatalog | null;
  detectingContexts: boolean;
  connectionValidation: ScanValidation;
  validation: ScanValidation;
  showValidationErrors: boolean;
  validationRequest: { version: number; field?: ScanFieldName };
  insecureAcknowledged: boolean;
  onSetInsecureAcknowledged: (value: boolean) => void;
  onTestConnection: () => void | Promise<void>;
  onInspectKubeconfig: () => void | Promise<void>;
  onUseDetectedKubeconfig: () => void | Promise<void>;
  onPreflight: () => void | Promise<void>;
  onStartScan: () => void | Promise<void>;
  onDetectContexts: () => void | Promise<void>;
  onBrowseOutput: () => void | Promise<void>;
  onBrowseKubeconfig: () => void | Promise<void>;
  onBrowseCACert: () => void | Promise<void>;
  onLoadDroppedKubeconfig: (fileName: string, content: string) => void | Promise<void>;
  loadedKubeconfigLabel?: string;
}) {
  const [tokenVisible, setTokenVisible] = useState(false);
  const connectionMethod = props.scanForm.connectionMethod || "current";
  const trustMode = describeTrustMode(props.scanForm);
  const authMode = describeAuthMode(props.scanForm);
  const selectedChoice = connectionChoices[connectionMethod];
  const reviewCards = buildReviewCards(props.scanForm, authMode, trustMode, props.validation.riskFlags);
  const artifacts = plannedArtifactsForScan(props.scanForm);

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
    target.scrollIntoView?.({ behavior: "smooth", block: "center" });
  }, [props.validationRequest.field, props.validationRequest.version]);

  const updateForm = <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => {
    props.setScanForm((current) => ({ ...current, [key]: value }));
  };
  const connectionFieldError = (field: ScanFieldName) =>
    props.connectionTest?.fieldErrors?.[field] ||
    (props.showValidationErrors ? props.connectionValidation.fieldErrors[field] : undefined);
  const scanFieldError = (field: ScanFieldName) =>
    props.showValidationErrors ? props.validation.fieldErrors[field] : undefined;
  const fieldWarning = (field: ScanFieldName) =>
    props.connectionTest?.fieldWarnings?.[field] ||
    props.connectionValidation.fieldWarnings[field] ||
    props.validation.fieldWarnings[field];

  return (
    <section className="page-grid scan-grid scan-workflow-grid">
      <section className="panel scan-main-panel">
        <SectionHeader
          eyebrow="Guided Scan"
          title="Connect, validate, scope, and launch"
          description="Pick the simplest access path, prove the connection works, decide what to collect and write, then run preflight before the full bundle collection starts."
        />

        <Stepper
          current={props.scanStage}
          steps={buildSteps(props.scanStage)}
          onSelect={(stage) => {
            if (stage === "connect") props.onSetScanStage("connect");
            if (stage === "validate") props.onSetScanStage("validate");
            if (stage === "outputs" && props.connectionTest?.canConnect) props.onSetScanStage("outputs");
            if (stage === "launch" && props.connectionTest?.canConnect) props.onSetScanStage("launch");
          }}
        />

        <div className="scan-stage-shell">
          <div className="scan-stage-intro">
            <p className="eyebrow">{selectedChoice.eyebrow}</p>
            <h3>{stageTitle(props.scanStage)}</h3>
            <p className="muted">{stageDescription(props.scanStage)}</p>
          </div>

          {props.scanStage === "connect" ? (
            <div className="scan-stage-stack">
              {props.connectionAdvisor ? (
                <div className="notice notice-info">
                  <strong>Recommended start on this machine</strong>
                  <p className="muted">
                    {props.connectionAdvisor.recommendedReason ||
                      "Start with the simplest access path that already works on this machine."}
                  </p>
                  <div className="toolbar wrap-toolbar">
                    {props.connectionAdvisor.currentLoginAvailable ? (
                      <button
                        type="button"
                        className="button secondary quiet"
                        onClick={() =>
                          props.setScanForm((current) => ({
                            ...current,
                            connectionMethod: "current",
                            contextName: current.contextName || props.connectionAdvisor?.currentContext || "",
                          }))
                        }
                      >
                        Use existing access
                      </button>
                    ) : null}
                    {props.connectionAdvisor.defaultKubeconfigAvailable ? (
                      <button
                        type="button"
                        className="button secondary quiet"
                        onClick={() => void props.onUseDetectedKubeconfig()}
                      >
                        Use detected kubeconfig
                      </button>
                    ) : null}
                  </div>
                  <ReadinessList className="compact-readiness-list" items={machineReadinessItems(props.connectionAdvisor)} />
                </div>
              ) : null}

              <Card title="1. Choose how to connect" description="API endpoint mode stays intentionally manual so the simplest path is easier to recognize first.">
                <div className="connection-choice-groups">
                  <ConnectionChoiceGroup title="Use existing access" description="Best for desktops and jumpboxes where Kubernetes access is already configured.">
                    <ConnectionChoiceCard
                      method="current"
                      active={connectionMethod === "current"}
                      onSelect={(method) => updateForm("connectionMethod", method)}
                      recommended={props.connectionAdvisor?.recommendedMethod === "current"}
                      detail={connectionChoiceDetail("current", props.connectionAdvisor)}
                    />
                  </ConnectionChoiceGroup>
                  <ConnectionChoiceGroup title="Bring a kubeconfig" description="Use this when someone handed you a kubeconfig file or raw kubeconfig YAML.">
                    <ConnectionChoiceCard
                      method="kubeconfig_file"
                      active={connectionMethod === "kubeconfig_file"}
                      onSelect={(method) => updateForm("connectionMethod", method)}
                      recommended={props.connectionAdvisor?.recommendedMethod === "kubeconfig_file"}
                      detail={connectionChoiceDetail("kubeconfig_file", props.connectionAdvisor)}
                    />
                    <ConnectionChoiceCard
                      method="kubeconfig_inline"
                      active={connectionMethod === "kubeconfig_inline"}
                      onSelect={(method) => updateForm("connectionMethod", method)}
                      detail="Useful when the kubeconfig arrives through a vault, ticket, or copy-safe handoff instead of a local file."
                    />
                  </ConnectionChoiceGroup>
                  <ConnectionChoiceGroup title="Manual / advanced" description="Use direct API access only when kubeconfig mode is not the better fit.">
                    <ConnectionChoiceCard
                      method="api_endpoint"
                      active={connectionMethod === "api_endpoint"}
                      onSelect={(method) => updateForm("connectionMethod", method)}
                      detail="Manual endpoint, token, and TLS setup for clusters where you do not have a usable kubeconfig on this machine."
                    />
                  </ConnectionChoiceGroup>
                </div>
              </Card>

              <Card title={`${selectedChoice.label} setup`} description={selectedChoice.description}>
                {connectionMethod === "current" ? (
                  <CurrentAccessSetup
                    connectionAdvisor={props.connectionAdvisor}
                    contextCatalog={props.contextCatalog}
                    detectingContexts={props.detectingContexts}
                    contextName={props.scanForm.contextName || ""}
                    fieldError={connectionFieldError}
                    onDetectContexts={props.onDetectContexts}
                    onChangeContext={(value) => updateForm("contextName", value)}
                  />
                ) : null}
                {connectionMethod === "kubeconfig_file" ? (
                  <KubeconfigFileSetup
                    busy={props.busy}
                    scanForm={props.scanForm}
                    inspection={props.kubeconfigInspection}
                    contextCatalog={props.contextCatalog}
                    detectingContexts={props.detectingContexts}
                    fieldError={connectionFieldError}
                    fieldWarning={fieldWarning}
                    onBrowseKubeconfig={props.onBrowseKubeconfig}
                    onInspectKubeconfig={props.onInspectKubeconfig}
                    onDetectContexts={props.onDetectContexts}
                    onUseDetectedKubeconfig={props.onUseDetectedKubeconfig}
                    connectionAdvisor={props.connectionAdvisor}
                    updateForm={updateForm}
                    onLoadDroppedKubeconfig={props.onLoadDroppedKubeconfig}
                    loadedKubeconfigLabel={props.loadedKubeconfigLabel}
                  />
                ) : null}
                {connectionMethod === "kubeconfig_inline" ? (
                  <KubeconfigInlineSetup
                    inspection={props.kubeconfigInspection}
                    contextCatalog={props.contextCatalog}
                    detectingContexts={props.detectingContexts}
                    scanForm={props.scanForm}
                    fieldError={connectionFieldError}
                    fieldWarning={fieldWarning}
                    onInspectKubeconfig={props.onInspectKubeconfig}
                    onDetectContexts={props.onDetectContexts}
                    updateForm={updateForm}
                    onLoadDroppedKubeconfig={props.onLoadDroppedKubeconfig}
                    loadedKubeconfigLabel={props.loadedKubeconfigLabel}
                    busy={props.busy}
                  />
                ) : null}
                {connectionMethod === "api_endpoint" ? (
                  <ApiEndpointSetup
                    busy={props.busy}
                    scanForm={props.scanForm}
                    tokenVisible={tokenVisible}
                    setTokenVisible={setTokenVisible}
                    fieldError={connectionFieldError}
                    fieldWarning={fieldWarning}
                    insecureAcknowledged={props.insecureAcknowledged}
                    onSetInsecureAcknowledged={props.onSetInsecureAcknowledged}
                    onBrowseCACert={props.onBrowseCACert}
                    updateForm={updateForm}
                  />
                ) : null}
              </Card>

              <div className="form-actions stage-actions">
                <p className="muted stage-action-copy">
                  Step 2 runs only a lightweight connection test. It does not start collection and it does not replace
                  full preflight.
                </p>
                <button type="button" className="button primary" onClick={() => props.onSetScanStage("validate")}>
                  Continue to validation
                </button>
              </div>
            </div>
          ) : null}

          {props.scanStage === "validate" ? (
            <div className="scan-stage-stack">
              <Card title="2. Validate the connection" description="Answer the fast question first: can K8V reach the Kubernetes API with the current transport, credentials, and TLS settings?">
                <div className="review-grid review-grid-compact">
                  <ReviewCard label="Connection path" value={selectedChoice.label} detail={selectedChoice.description} />
                  <ReviewCard label="Connection target" value={connectionSummary(props.scanForm)} detail={connectionMethod === "api_endpoint" ? "Direct API target" : "Current login or kubeconfig source"} />
                  <ReviewCard label="Auth mode" value={authMode.label} detail={authMode.detail} tone={authMode.tone} />
                  <ReviewCard label="TLS trust" value={trustMode.label} detail={trustMode.detail} tone={trustMode.tone} />
                </div>

                {props.showValidationErrors && props.connectionValidation.errors.length ? (
                  <div className="notice notice-error">
                    <strong>Finish these connection details first</strong>
                    <ul className="notice-list">{props.connectionValidation.errors.map((error) => <li key={error}>{error}</li>)}</ul>
                  </div>
                ) : null}

                {!props.connectionTest ? <div className="notice notice-info"><strong>What this test does</strong><p className="muted">The connection test checks reachability, credential acceptance, and TLS trust. It does not run the deeper RBAC and collection readiness checks yet.</p></div> : null}
                {props.connectionTest && props.connectionTest.canConnect ? <div className="notice notice-info"><strong>Connection test succeeded</strong><p className="muted">{props.connectionTest.nextAction || "Transport, auth, and TLS look healthy. Continue to scope and outputs next."}</p></div> : null}
                {props.connectionTest && !props.connectionTest.canConnect ? <div className="notice notice-error"><strong>Connection test needs follow-up</strong><p className="muted">{props.connectionTest.nextAction || "Adjust the connection details and run the test again before moving on."}</p></div> : null}

                <div className="form-actions stage-actions">
                  <button type="button" className="button secondary quiet" onClick={() => props.onSetScanStage("connect")}>Back to connection</button>
                  <button type="button" className="button primary" onClick={() => void props.onTestConnection()} disabled={props.busy}>{props.connectionTest ? "Retest connection" : "Test connection"}</button>
                  <button type="button" className="button secondary" onClick={() => props.onSetScanStage("outputs")} disabled={!props.connectionTest?.canConnect}>Continue to scope and outputs</button>
                </div>
              </Card>
            </div>
          ) : null}

          {props.scanStage === "outputs" ? (
            <div className="scan-stage-stack">
              <Card title="3. Choose scope and outputs" description="Now that the connection works, decide what the scan should cover and which artifacts should be written when the run succeeds.">
                <div className="form-grid">
                  <Field label="Namespaces (optional)" hint="Leave blank to scan every namespace the credentials can read. Add a comma-separated list only when you intentionally want a narrower scope.">
                    <input aria-label="Namespaces" placeholder="payments, frontend" value={(props.scanForm.namespaces || []).join(", ")} onChange={(event) => updateForm("namespaces", splitList(event.target.value))} />
                  </Field>
                  <Field label="Cluster label (optional)" hint="If left blank, K8V derives a label from the active context or API endpoint.">
                    <input aria-label="Cluster label" placeholder="prod-east" value={props.scanForm.clusterName || ""} onChange={(event) => updateForm("clusterName", event.target.value)} />
                  </Field>
                  <Field label="Environment (optional)" hint="Examples: production, staging, dev.">
                    <input aria-label="Environment" placeholder="production" value={props.scanForm.environment || ""} onChange={(event) => updateForm("environment", event.target.value)} />
                  </Field>
                  <Field label="Output directory" hint="A successful run writes the bundle and report files into this directory." error={scanFieldError("outputDir")}>
                    <div className="inline-field">
                      <input aria-label="Output directory" data-scan-field="outputDir" value={props.scanForm.outputDir || ""} onChange={(event) => updateForm("outputDir", event.target.value)} />
                      <button type="button" className="button secondary" onClick={() => void props.onBrowseOutput()} disabled={props.busy}>Browse</button>
                    </div>
                  </Field>
                </div>

                <section className="artifact-section" aria-label="Output artifacts">
                  <SectionHeader compact title="Artifacts to generate" description="The bundle is always written. Use the extra exports only when you need them." />
                  <div className="artifact-toggle-grid">
                    <ArtifactToggle label="Executive summary" description="Short HTML summary for quick leadership or ticket handoff." checked={Boolean(props.scanForm.summary)} onChange={(checked) => updateForm("summary", checked)} />
                    <ArtifactToggle label="Operator runbook" description="HTML runbook focused on follow-up tasks and recovery actions." checked={Boolean(props.scanForm.runbook)} onChange={(checked) => updateForm("runbook", checked)} />
                    <ArtifactToggle label="CSV exports" description="CSV tables for spreadsheets, external analysis, or evidence attachments." checked={Boolean(props.scanForm.csvExport)} onChange={(checked) => updateForm("csvExport", checked)} />
                    <ArtifactToggle label="Redacted outputs" description="Write a redacted copy when you need a shareable bundle/report with reduced sensitive detail." checked={Boolean(props.scanForm.redact)} onChange={(checked) => updateForm("redact", checked)} />
                  </div>
                </section>

                <section className="artifact-section" aria-label="Planned output files">
                  <SectionHeader compact title="What will be written" description="This makes the bundle workflow explicit before the run starts." />
                  <div className="artifact-list">{artifacts.map((artifact) => <div key={artifact.name} className="artifact-row"><strong>{artifact.name}</strong><p className="muted">{artifact.detail}</p></div>)}</div>
                </section>

                <details className="advanced-panel">
                  <summary>Advanced settings</summary>
                  <div className="advanced-panel-body">
                    <div className="form-grid">
                      <Field label="Compare baseline" hint="Optional path to a previous bundle for drift and progress comparisons."><input aria-label="Compare baseline" placeholder="C:\\scans\\previous\\recovery-scan.json" value={props.scanForm.compareTo || ""} onChange={(event) => updateForm("compareTo", event.target.value)} /></Field>
                      <Field label="Profile" hint="Standard is the safest default for most first runs."><select value={props.scanForm.profileName || "standard"} onChange={(event) => updateForm("profileName", event.target.value)}><option value="standard">standard</option><option value="enterprise">enterprise</option><option value="dev">dev</option><option value="airgap">airgap</option></select></Field>
                      <Field label="Recovery target"><select value={props.scanForm.target || "vm"} onChange={(event) => updateForm("target", event.target.value)}><option value="vm">vm</option><option value="baremetal">baremetal</option></select></Field>
                      <Field label="Timeout (seconds)" hint="Increase this for slower networks or especially large clusters." error={scanFieldError("timeoutSeconds")}><input aria-label="Timeout seconds" data-scan-field="timeoutSeconds" type="number" min={10} value={props.scanForm.timeoutSeconds || 60} onChange={(event) => updateForm("timeoutSeconds", Number(event.target.value))} /></Field>
                      <Field label="Customer ID (optional)"><input aria-label="Customer ID" placeholder="customer-123" value={props.scanForm.customerId || ""} onChange={(event) => updateForm("customerId", event.target.value)} /></Field>
                      <Field label="Site (optional)"><input aria-label="Site" placeholder="us-east-1a" value={props.scanForm.site || ""} onChange={(event) => updateForm("site", event.target.value)} /></Field>
                    </div>
                    <div className="toggle-grid">
                      {connectionMethod !== "api_endpoint" ? <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.insecure)} onChange={(event) => updateForm("insecure", event.target.checked)} />Skip TLS verification</label> : null}
                      <label className="toggle"><input type="checkbox" checked={Boolean(props.scanForm.includeSecretMetadata)} onChange={(event) => updateForm("includeSecretMetadata", event.target.checked)} />Secret metadata</label>
                    </div>
                  </div>
                </details>

                <div className="form-actions stage-actions">
                  <button type="button" className="button secondary quiet" onClick={() => props.onSetScanStage("validate")}>Back to validation</button>
                  <button type="button" className="button primary" onClick={() => props.onSetScanStage("launch")}>Continue to review</button>
                </div>
              </Card>
            </div>
          ) : null}

          {props.scanStage === "launch" ? (
            <div className="scan-stage-stack">
              <Card title="4. Review, run preflight, then start the scan" description="Preflight checks the final transport, RBAC, scope, and collection readiness using the exact scope and output settings you selected. The real scan collects evidence and writes the bundle afterwards.">
                <div className="review-grid">{reviewCards.map((card) => <ReviewCard key={card.label} label={card.label} value={card.value} detail={card.detail} tone={card.tone} />)}</div>
                {props.showValidationErrors && props.validation.errors.length ? <div className="notice notice-error"><strong>Fix these before preflight or launch</strong><ul className="notice-list">{props.validation.errors.map((error) => <li key={error}>{error}</li>)}</ul></div> : null}
                {props.validation.riskFlags.length ? <div className="notice notice-warning"><strong>Operator risks to double-check</strong><ul className="notice-list">{props.validation.riskFlags.map((flag) => <li key={flag}>{flag}</li>)}</ul></div> : null}
                {!props.connectionTest?.canConnect ? <div className="notice notice-error"><strong>Connection test still needs to pass</strong><p className="muted">Go back to Step 2 first. Preflight assumes the basic transport, credentials, and TLS path already work.</p></div> : null}
                {props.connectionTest?.canConnect && !props.preflight ? <div className="notice notice-info"><strong>Ready for full preflight</strong><p className="muted">Run preflight now to check RBAC, scope, and collection readiness with the final settings shown above.</p></div> : null}
                {props.preflight ? <div className={`notice ${props.preflight.canRun ? "notice-info" : "notice-error"}`}><strong>{props.preflight.canRun ? "Preflight passed" : "Preflight found blockers"}</strong><p className="muted">{props.preflight.canRun ? "The full scan can start. Results and exports will be written to the chosen output directory." : "Resolve the blocking checks in the rail before starting the scan."}</p></div> : null}
                <div className="form-actions stage-actions">
                  <button type="button" className="button secondary quiet" onClick={() => props.onSetScanStage("outputs")}>Back to scope and outputs</button>
                  <button type="button" className="button secondary" onClick={() => void props.onPreflight()} disabled={!props.connectionTest?.canConnect || props.busy}>{props.preflight ? "Run preflight again" : "Run preflight"}</button>
                  <button type="button" className="button primary" onClick={() => void props.onStartScan()} disabled={!props.preflight?.canRun || props.busy}>Start scan</button>
                </div>
              </Card>
            </div>
          ) : null}
        </div>
      </section>

      <section className="panel scan-rail-panel">
        <SectionHeader
          eyebrow={props.preflight ? "Preflight" : props.connectionTest ? "Connection test" : "Assistant"}
          title={props.preflight ? "Connection, RBAC, and collection readiness" : props.connectionTest ? "Transport, auth, and TLS result" : assistantTitle(connectionMethod)}
          description={props.preflight ? "The rail stays focused on blockers, degraded-but-runnable gaps, next actions, and copyable guidance." : props.connectionTest ? "The lightweight test gives the fastest answer to whether this connection works before the deeper preflight pass." : "The rail adapts to the selected connection method so first-time operators can prepare the right endpoint, credentials, and trust path before testing."}
        />
        {props.preflight ? <PreflightPanel preflight={props.preflight} connectionMethod={connectionMethod} insecure={Boolean(props.scanForm.insecure)} /> : props.connectionTest ? <ConnectionTestPanel report={props.connectionTest} /> : <ConnectionAssistant connectionMethod={connectionMethod} connectionAdvisor={props.connectionAdvisor} kubeconfigInspection={props.kubeconfigInspection} trustMode={trustMode} insecure={Boolean(props.scanForm.insecure)} />}
      </section>
    </section>
  );
}

function CurrentAccessSetup(props: {
  connectionAdvisor: ConnectionAdvisor | null;
  contextCatalog: ContextCatalog | null;
  detectingContexts: boolean;
  contextName: string;
  fieldError: (field: ScanFieldName) => string | undefined;
  onDetectContexts: () => void | Promise<void>;
  onChangeContext: (value: string) => void;
}) {
  return (
    <div className="scan-stage-stack">
      {props.connectionAdvisor?.currentLoginDetail ? (
        <div className="notice notice-info compact">
          <strong>Detected local access</strong>
          <p className="muted">{props.connectionAdvisor.currentLoginDetail}</p>
        </div>
      ) : (
        <div className="notice notice-warning compact">
          <strong>No local access was positively detected</strong>
          <p className="muted">
            {props.connectionAdvisor?.currentLoginWarning ||
              "If this machine does not already have working Kubernetes access, switch to kubeconfig or API endpoint mode."}
          </p>
        </div>
      )}
      {props.connectionAdvisor?.currentLoginWarning && props.connectionAdvisor.currentLoginAvailable ? (
        <div className="notice notice-warning compact">
          <strong>Current access caution</strong>
          <p className="muted">{props.connectionAdvisor.currentLoginWarning}</p>
        </div>
      ) : null}
      <Field label="Context (optional)" hint="Leave blank to use the current kubectl context." error={props.fieldError("contextName")}>
        <input aria-label="Context" data-scan-field="contextName" list={props.contextCatalog?.contexts?.length ? "scan-context-options" : undefined} placeholder={props.connectionAdvisor?.currentContext || "Leave blank to use the current context"} value={props.contextName} onChange={(event) => props.onChangeContext(event.target.value)} />
      </Field>
      <div className="inline-actions">
        <button type="button" className="button secondary" onClick={() => void props.onDetectContexts()}>{props.detectingContexts ? "Loading contexts..." : "Detect contexts"}</button>
        {props.contextCatalog?.currentContext ? <button type="button" className="button secondary quiet" onClick={() => props.onChangeContext(props.contextCatalog?.currentContext || "")}>Use {props.contextCatalog.currentContext}</button> : null}
      </div>
      <ContextCatalogHint catalog={props.contextCatalog} />
    </div>
  );
}

function KubeconfigFileSetup(props: {
  busy: boolean;
  scanForm: ScanRequest;
  inspection: KubeconfigInspection | null;
  contextCatalog: ContextCatalog | null;
  detectingContexts: boolean;
  fieldError: (field: ScanFieldName) => string | undefined;
  fieldWarning: (field: ScanFieldName) => string | undefined;
  onBrowseKubeconfig: () => void | Promise<void>;
  onInspectKubeconfig: () => void | Promise<void>;
  onDetectContexts: () => void | Promise<void>;
  onUseDetectedKubeconfig: () => void | Promise<void>;
  connectionAdvisor: ConnectionAdvisor | null;
  updateForm: <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => void;
  onLoadDroppedKubeconfig: (fileName: string, content: string) => void | Promise<void>;
  loadedKubeconfigLabel?: string;
}) {
  return (
    <div className="scan-stage-stack">
      <Field label="Kubeconfig file" hint="Any valid kubeconfig works here, including config, .yaml, .backup, or a file with no extension. If browsing is awkward, use Paste kubeconfig instead." warning={props.fieldWarning("kubeconfigPath")} error={props.fieldError("kubeconfigPath")}>
        <div className="inline-field">
          <input aria-label="Kubeconfig file" data-scan-field="kubeconfigPath" placeholder="C:\\Users\\you\\.kube\\config" value={props.scanForm.kubeconfigPath || ""} onChange={(event) => props.updateForm("kubeconfigPath", event.target.value)} />
          <button type="button" className="button secondary" onClick={() => void props.onBrowseKubeconfig()} disabled={props.busy}>Browse</button>
        </div>
      </Field>
      <div className="toolbar wrap-toolbar">
        <button type="button" className="button secondary quiet" onClick={() => void props.onInspectKubeconfig()}>Inspect kubeconfig</button>
        <button type="button" className="button secondary quiet" onClick={() => void props.onDetectContexts()}>{props.detectingContexts ? "Loading contexts..." : "Detect contexts"}</button>
        {props.connectionAdvisor?.defaultKubeconfigAvailable ? <button type="button" className="button secondary quiet" onClick={() => void props.onUseDetectedKubeconfig()}>Use detected kubeconfig</button> : null}
      </div>
      <KubeconfigDropzone
        busy={props.busy}
        loadedLabel={props.loadedKubeconfigLabel}
        onLoadDroppedKubeconfig={props.onLoadDroppedKubeconfig}
      />
      <Field label="Context (optional)" hint="Leave blank to use the kubeconfig current context.">
        <input aria-label="Context" data-scan-field="contextName" list={props.contextCatalog?.contexts?.length ? "scan-context-options" : undefined} placeholder={props.inspection?.currentContext || "Leave blank to use the kubeconfig current context"} value={props.scanForm.contextName || ""} onChange={(event) => props.updateForm("contextName", event.target.value)} />
      </Field>
      <InspectionSummary inspection={props.inspection} />
      <ContextCatalogHint catalog={props.contextCatalog} />
    </div>
  );
}

function KubeconfigInlineSetup(props: {
  busy: boolean;
  inspection: KubeconfigInspection | null;
  contextCatalog: ContextCatalog | null;
  detectingContexts: boolean;
  scanForm: ScanRequest;
  fieldError: (field: ScanFieldName) => string | undefined;
  fieldWarning: (field: ScanFieldName) => string | undefined;
  onInspectKubeconfig: () => void | Promise<void>;
  onDetectContexts: () => void | Promise<void>;
  updateForm: <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => void;
  onLoadDroppedKubeconfig: (fileName: string, content: string) => void | Promise<void>;
  loadedKubeconfigLabel?: string;
}) {
  return (
    <div className="scan-stage-stack">
      <KubeconfigDropzone
        busy={props.busy}
        loadedLabel={props.loadedKubeconfigLabel}
        onLoadDroppedKubeconfig={props.onLoadDroppedKubeconfig}
      />
      <Field label="Pasted kubeconfig" hint="Paste raw kubeconfig YAML exactly as provided. No renaming, wrapping, or conversion is needed." warning={props.fieldWarning("kubeconfigContent")} error={props.fieldError("kubeconfigContent")}>
        <textarea aria-label="Pasted kubeconfig" data-scan-field="kubeconfigContent" rows={12} placeholder={"apiVersion: v1\nkind: Config\nclusters:\n- name: prod-east"} value={props.scanForm.kubeconfigContent || ""} onChange={(event) => props.updateForm("kubeconfigContent", event.target.value)} spellCheck={false} />
      </Field>
      <div className="toolbar wrap-toolbar">
        <button type="button" className="button secondary quiet" onClick={() => void props.onInspectKubeconfig()}>Inspect kubeconfig</button>
        <button type="button" className="button secondary quiet" onClick={() => void props.onDetectContexts()}>{props.detectingContexts ? "Loading contexts..." : "Detect contexts"}</button>
      </div>
      <Field label="Context (optional)" hint="Leave blank to use the kubeconfig current context.">
        <input aria-label="Context" data-scan-field="contextName" list={props.contextCatalog?.contexts?.length ? "scan-context-options" : undefined} placeholder={props.inspection?.currentContext || "Leave blank to use the kubeconfig current context"} value={props.scanForm.contextName || ""} onChange={(event) => props.updateForm("contextName", event.target.value)} />
      </Field>
      <InspectionSummary inspection={props.inspection} />
      <ContextCatalogHint catalog={props.contextCatalog} />
    </div>
  );
}

function KubeconfigDropzone(props: {
  busy: boolean;
  loadedLabel?: string;
  onLoadDroppedKubeconfig: (fileName: string, content: string) => void | Promise<void>;
}) {
  const [dragActive, setDragActive] = useState(false);
  const [loading, setLoading] = useState(false);
  const inputId = useId();

  async function loadFile(file: File | null | undefined) {
    if (!file) {
      return;
    }
    setLoading(true);
    try {
      const content = await readFileAsText(file);
      await props.onLoadDroppedKubeconfig(file.name, content);
    } finally {
      setLoading(false);
      setDragActive(false);
    }
  }

  function handleDragOver(event: DragEvent<HTMLElement>) {
    event.preventDefault();
    setDragActive(true);
  }

  function handleDragLeave(event: DragEvent<HTMLElement>) {
    event.preventDefault();
    if (event.currentTarget.contains(event.relatedTarget as Node | null)) {
      return;
    }
    setDragActive(false);
  }

  async function handleDrop(event: DragEvent<HTMLElement>) {
    event.preventDefault();
    await loadFile(event.dataTransfer.files?.[0]);
  }

  async function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    await loadFile(file);
  }

  return (
    <section
      className={`dropzone ${dragActive ? "is-dragging" : ""} ${loading || props.busy ? "is-busy" : ""}`}
      onDragOver={handleDragOver}
      onDragEnter={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={(event) => void handleDrop(event)}
      aria-label="Kubeconfig dropzone"
    >
      <div className="dropzone-head">
        <strong>Drop a kubeconfig here</strong>
        <p className="muted">
          K8V reads the file into paste kubeconfig mode, validates the contents, and ignores filename quirks like
          `.backup` or no extension.
        </p>
      </div>
      <div className="toolbar wrap-toolbar">
        <label htmlFor={inputId} className="button secondary quiet">
          {loading ? "Loading kubeconfig..." : "Choose file into paste mode"}
        </label>
        <input
          id={inputId}
          className="sr-only"
          type="file"
          aria-label="Choose kubeconfig file for paste mode"
          onChange={(event) => void handleFileChange(event)}
        />
        <span className="muted dropzone-note">
          {props.loadedLabel
            ? `Loaded ${props.loadedLabel} into paste mode.`
            : "Useful when the kubeconfig came from another machine or the native picker is awkward."}
        </span>
      </div>
    </section>
  );
}

function readFileAsText(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(typeof reader.result === "string" ? reader.result : "");
    reader.onerror = () => reject(reader.error || new Error("Could not read the selected kubeconfig."));
    reader.readAsText(file);
  });
}

function ApiEndpointSetup(props: {
  busy: boolean;
  scanForm: ScanRequest;
  tokenVisible: boolean;
  setTokenVisible: (value: boolean | ((current: boolean) => boolean)) => void;
  fieldError: (field: ScanFieldName) => string | undefined;
  fieldWarning: (field: ScanFieldName) => string | undefined;
  insecureAcknowledged: boolean;
  onSetInsecureAcknowledged: (value: boolean) => void;
  onBrowseCACert: () => void | Promise<void>;
  updateForm: <K extends keyof ScanRequest>(key: K, value: ScanRequest[K]) => void;
}) {
  return (
    <div className="scan-stage-stack">
      <div className="wizard-substeps">
        <section className="wizard-substep">
          <div className="wizard-substep-copy">
            <p className="eyebrow">Substep 1</p>
            <h4>API server endpoint</h4>
            <p className="muted">Paste the Kubernetes API server address only. This is the control-plane endpoint, not an ingress or application URL.</p>
          </div>
          <Field label="API server host or URL" hint="Accepted examples: https://cluster.example.net:6443, control-plane.prod-east.example.net:6443, or 10.0.0.15:6443." warning={props.fieldWarning("apiServerEndpoint")} error={props.fieldError("apiServerEndpoint")}>
            <input aria-label="API server host or URL" data-scan-field="apiServerEndpoint" placeholder="https://cluster.example.net:6443" value={props.scanForm.apiServerEndpoint || ""} onChange={(event) => props.updateForm("apiServerEndpoint", event.target.value)} />
          </Field>
          <CodeBlock label="Find the current cluster server" code={"kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'"} />
        </section>
        <section className="wizard-substep">
          <div className="wizard-substep-copy">
            <p className="eyebrow">Substep 2</p>
            <h4>Authentication</h4>
            <p className="muted">Paste the raw bearer token value. If the cluster relies on exec plugins, cloud helpers, SSO prompts, or client certificates, kubeconfig mode is usually the better choice.</p>
          </div>
          <Field label="Bearer token" hint="Leading Bearer text, extra spaces, and accidental line breaks are removed automatically." error={props.fieldError("bearerToken")}>
            <div className="inline-field token-field">
              <input aria-label="Bearer token" data-scan-field="bearerToken" type={props.tokenVisible ? "text" : "password"} placeholder="Paste a Kubernetes API bearer token" value={props.scanForm.bearerToken || ""} onChange={(event) => props.updateForm("bearerToken", event.target.value)} autoComplete="off" autoCorrect="off" autoCapitalize="off" spellCheck={false} />
              <button type="button" className="button secondary quiet" onClick={() => props.setTokenVisible((current) => !current)}>{props.tokenVisible ? "Hide" : "Show"}</button>
              <button type="button" className="button secondary quiet" onClick={() => props.updateForm("bearerToken", "")}>Clear</button>
            </div>
          </Field>
          <CodeBlock label="Create a short-lived service-account token" code="kubectl create token <service-account> --namespace <namespace>" />
        </section>
        <section className="wizard-substep">
          <div className="wizard-substep-copy">
            <p className="eyebrow">Substep 3</p>
            <h4>TLS trust</h4>
            <p className="muted">Use system trust for publicly trusted certificates, add a private CA for internal trust chains, and treat skip-TLS as a temporary workaround only.</p>
          </div>
          <div className="trust-guide-grid">
            <article className="trust-guide-card"><strong>Publicly trusted cert</strong><p className="muted">Leave CA inputs empty when the API server certificate chains to a CA already trusted on this machine.</p></article>
            <article className="trust-guide-card"><strong>Private or internal CA</strong><p className="muted">Add the issuing CA as a file or pasted PEM when the cluster uses an internal or self-signed trust chain.</p></article>
            <article className="trust-guide-card warning"><strong>Skip verification</strong><p className="muted">Use only in a trusted environment as a temporary test or break-glass workaround.</p></article>
          </div>
          <div className="form-grid">
            <Field label="CA certificate file (optional)" hint="Point to the PEM that issued the API server certificate when the cluster uses a private CA.">
              <div className="inline-field">
                <input aria-label="CA certificate file" placeholder="C:\\certs\\cluster-ca.pem" value={props.scanForm.caCertPath || ""} onChange={(event) => props.updateForm("caCertPath", event.target.value)} />
                <button type="button" className="button secondary" onClick={() => void props.onBrowseCACert()} disabled={props.busy}>Browse</button>
              </div>
            </Field>
            <Field label="Pasted CA certificate (optional)" hint="Paste PEM content only when you do not want to reference a local CA file.">
              <textarea aria-label="Pasted CA certificate" rows={6} placeholder="-----BEGIN CERTIFICATE-----" value={props.scanForm.caCertContent || ""} onChange={(event) => props.updateForm("caCertContent", event.target.value)} spellCheck={false} />
            </Field>
          </div>
          {props.fieldWarning("caTrust") ? <div className="notice notice-warning compact"><strong>Trust hint</strong><p className="muted">{props.fieldWarning("caTrust")}</p></div> : null}
          <div className="toggle-stack">
            <label className={`toggle-card ${props.scanForm.insecure ? "is-warning" : ""}`}>
              <input type="checkbox" checked={Boolean(props.scanForm.insecure)} onChange={(event) => props.updateForm("insecure", event.target.checked)} />
              <div><strong>Skip TLS verification</strong><p className="muted">This disables server certificate verification. Keep it limited to trusted, temporary troubleshooting or recovery preparation scenarios.</p></div>
            </label>
            {props.scanForm.insecure ? <label className={`toggle-card acknowledgement-card ${props.fieldError("insecureAcknowledgement") ? "is-error" : "is-warning"}`}><input type="checkbox" data-scan-field="insecureAcknowledgement" checked={props.insecureAcknowledged} onChange={(event) => props.onSetInsecureAcknowledged(event.target.checked)} /><div><strong>I understand this disables certificate verification</strong><p className="muted">Keep this off the steady-state path whenever possible.</p>{props.fieldError("insecureAcknowledgement") ? <small className="field-message error">{props.fieldError("insecureAcknowledgement")}</small> : null}</div></label> : null}
          </div>
        </section>
      </div>
    </div>
  );
}

function ConnectionAssistant(props: {
  connectionMethod: ConnectionMethod;
  connectionAdvisor: ConnectionAdvisor | null;
  kubeconfigInspection: KubeconfigInspection | null;
  trustMode: { label: string; detail: string; tone: ScanReviewTone };
  insecure: boolean;
}) {
  if (props.connectionMethod === "api_endpoint") {
    return (
      <div className="assistant-stack">
        {props.insecure ? <div className="notice notice-warning compact"><strong>Skip-TLS is enabled</strong><p className="muted">The connection test can still run, but K8V will not verify the API server certificate. Treat this as a temporary, trusted-environment workaround.</p></div> : null}
        <AssistantSection title="When direct API mode is the right choice" body="Use this when you know the Kubernetes API server URL and can supply a bearer token directly. If the cluster relies on exec plugins, OIDC or cloud auth helpers, SSO prompts, or client certificates, kubeconfig mode is usually the better fit." />
        <AssistantSection title="1. Find the Kubernetes API server URL" body="You need the control-plane API address only. Pull it from an existing kubeconfig, platform docs, cluster inventory, or a control-plane host/IP plus port 6443 when appropriate.">
          <CodeBlock label="From the current kubeconfig" code={"kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'"} />
          <CodeBlock label="Endpoint example" code="https://control-plane.prod-east.example.net:6443" />
          <CodeBlock label="Host or IP example" code="10.0.0.15:6443" />
        </AssistantSection>
        <AssistantSection title="2. Obtain a short-lived bearer token" body="Prefer a least-privilege service account and create a fresh token instead of relying on long-lived static credentials. Do not default to cluster-admin.">
          <CodeBlock label="Create a short-lived token" code="kubectl create token <service-account> --namespace <namespace>" />
          <CodeBlock label="Example with duration" code="kubectl create token k8v-scan --namespace k8v-ops --duration 10m" />
        </AssistantSection>
        <AssistantSection title="3. Choose the trust path" body={`Current trust mode: ${props.trustMode.label}. Use system trust for publicly trusted certs, a CA file or pasted PEM for private/internal CAs, and skip verification only as a last-resort temporary workaround.`} />
        <AssistantSection title="What happens next" body="Step 2 tests reachability, auth, and TLS only. Later, preflight checks RBAC, scope, and collection readiness with the final scan settings." />
      </div>
    );
  }
  if (props.connectionMethod === "kubeconfig_file") {
    return (
      <div className="assistant-stack">
        <AssistantSection title="Kubeconfig file workflow" body="Pick any valid kubeconfig file from disk. K8V validates the contents after selection, so filenames like config, prod-cluster.backup, and extensionless files all work.">
          <CodeBlock label="Show the cluster server from a kubeconfig" code={"kubectl config view --kubeconfig /path/to/file --minify -o jsonpath='{.clusters[0].cluster.server}'"} />
          <CodeBlock label="List available contexts" code="kubectl config get-contexts -o=name --kubeconfig /path/to/file" />
        </AssistantSection>
        {props.kubeconfigInspection?.summary ? <AssistantSection title="What K8V already found" body={props.kubeconfigInspection.summary}>{props.kubeconfigInspection.nextAction ? <p className="muted assistant-footnote">{props.kubeconfigInspection.nextAction}</p> : null}</AssistantSection> : null}
        {props.kubeconfigInspection?.missingReferencedFiles?.length ? <AssistantSection title="Portable kubeconfig warning" body="This kubeconfig is structurally valid, but it still depends on local CA or client-certificate files that are missing on this machine. If the selected context uses those files, the connection test will fail until they are restored or embedded."><ul className="notice-list compact-list">{props.kubeconfigInspection.missingReferencedFiles.map((reference) => <li key={reference}>{reference}</li>)}</ul></AssistantSection> : null}
      </div>
    );
  }
  if (props.connectionMethod === "kubeconfig_inline") {
    return (
      <div className="assistant-stack">
        <AssistantSection title="Pasted kubeconfig workflow" body="Paste raw kubeconfig YAML exactly as provided. This is the simplest fallback when browsing local files is inconvenient or the kubeconfig is being handed off through a secure vault or ticket.">
          <CodeBlock label="Capture raw kubeconfig content" code="kubectl config view --raw" />
        </AssistantSection>
        {props.kubeconfigInspection?.referencedFiles?.length ? <AssistantSection title="Path-based credential warning" body="This pasted kubeconfig still refers to CA or client-certificate files on disk. Paste mode carries only the YAML, so those local files still need to exist on the prepared machine or be embedded as *-data fields."><ul className="notice-list compact-list">{props.kubeconfigInspection.referencedFiles.map((reference) => <li key={reference}>{reference}</li>)}</ul></AssistantSection> : null}
      </div>
    );
  }
  return (
    <div className="assistant-stack">
      <AssistantSection title="Existing access workflow" body="Use this when the machine already reaches the cluster. K8V will rely on the current kubectl login or default kubeconfig.">
        <CodeBlock label="Quick context sanity check" code="kubectl config current-context" />
      </AssistantSection>
      {props.connectionAdvisor?.currentLoginDetail ? <AssistantSection title="Detected on this machine" body={props.connectionAdvisor.currentLoginDetail} /> : null}
      {props.connectionAdvisor?.currentLoginWarning ? <AssistantSection title="Local access caution" body={props.connectionAdvisor.currentLoginWarning} /> : null}
      {props.connectionAdvisor?.defaultKubeconfigAvailable ? <AssistantSection title="Fallback if existing access does not work" body={`A usable default kubeconfig was also detected at ${props.connectionAdvisor.defaultKubeconfigPath}. You can switch to kubeconfig mode without renaming the file.`} /> : null}
    </div>
  );
}

function ConnectionTestPanel(props: { report: ConnectionTestReport }) {
  const failures = props.report.checks?.filter((check) => check.status === "fail").length || 0;
  const warnings = props.report.checks?.filter((check) => check.status === "warn").length || 0;

  return (
    <div className="assistant-stack">
      {props.report.diagnosis ? (
        <div className="notice notice-error">
          <strong>{props.report.diagnosis.label || "Likely cause"}</strong>
          <p className="muted">{props.report.diagnosis.summary || props.report.nextAction}</p>
          {props.report.diagnosis.nextAction ? <p className="muted assistant-footnote">{props.report.diagnosis.nextAction}</p> : null}
        </div>
      ) : null}
      <div className="inline-metrics">
        <MetricCard label="Can connect" value={props.report.canConnect ? "Yes" : "No"} tone={props.report.canConnect ? "success" : "critical"} />
        <MetricCard label="Failures" value={failures} tone={failures ? "critical" : "success"} />
        <MetricCard label="Warnings" value={warnings} tone={warnings ? "high" : "success"} />
      </div>
      <KeyValueGrid items={[["Source", props.report.source || "Connection settings"], ["Server", props.report.server || "Not resolved yet"], ["Context", props.report.contextName || "Not applicable"]]} />
      <div className={`notice ${props.report.canConnect ? "notice-info" : "notice-error"}`}>
        <strong>{props.report.summary || (props.report.canConnect ? "Connection test succeeded." : "Connection test failed.")}</strong>
        <p className="muted">{props.report.nextAction || (props.report.canConnect ? "Continue to scope and outputs, then run full preflight before the scan." : "Adjust the connection details and test again.")}</p>
      </div>
      <div className="preflight-list">
        {(props.report.checks || []).map((check) => <article key={check.id} className={`preflight-check ${check.status}`}><div className="preflight-check-head"><strong>{check.title}</strong><Badge tone={toneForCheck(check.status)}>{labelForCheck(check.status)}</Badge></div><p>{check.detail}</p>{check.hint ? <p className="muted preflight-hint">{check.hint}</p> : null}</article>)}
      </div>
    </div>
  );
}

function AssistantSection(props: { title: string; body: string; children?: ReactNode }) {
  return (
    <section className="assistant-section">
      <div className="section-header compact">
        <div className="section-heading">
          <h4>{props.title}</h4>
          <p className="muted section-copy">{props.body}</p>
        </div>
      </div>
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
      {props.preflight.diagnosis ? <div className="notice notice-error compact"><strong>{props.preflight.diagnosis.label || "Likely blocker"}</strong><p className="muted">{props.preflight.diagnosis.summary || props.preflight.diagnosis.nextAction}</p>{props.preflight.diagnosis.nextAction ? <p className="muted assistant-footnote">{props.preflight.diagnosis.nextAction}</p> : null}</div> : null}
      {props.insecure ? <div className="notice notice-warning compact"><strong>Transport trust warning</strong><p className="muted">This preflight used insecure TLS mode. Re-enable certificate verification or add the proper CA material before treating this as a normal operating path.</p></div> : null}
      <div className="inline-metrics">
        <MetricCard label="Can run" value={props.preflight.canRun ? "Yes" : "No"} tone={props.preflight.canRun ? "success" : "critical"} />
        <MetricCard label="Blockers" value={blockers} tone={blockers ? "critical" : "success"} />
        <MetricCard label="Warnings" value={warnings} tone={warnings ? "high" : "success"} />
        <MetricCard label="Scope" value={props.preflight.scope} />
      </div>
      <div className={`notice ${blockers ? "notice-error" : warnings || props.preflight.degraded ? "notice-warning" : "notice-info"}`}><strong>Top next action</strong><p className="muted">{preflightNextAction(props.preflight, props.connectionMethod, props.insecure)}</p></div>
      {props.preflight.warnings?.length ? <div className="notice notice-info"><strong>Operator note</strong><ul className="notice-list">{props.preflight.warnings.map((warning) => <li key={warning}>{warning}</li>)}</ul></div> : null}
      <div className="preflight-category-stack">
        {categories.map((category) => <section key={category.id} className="preflight-category"><div className="section-header compact"><div className="section-heading"><h4>{category.title}</h4><p className="muted section-copy">{category.description}</p></div><div className="preflight-category-badges"><Badge tone={category.blockers ? "fail" : category.warnings ? "warn" : "pass"}>{category.blockers ? `${category.blockers} blocker${category.blockers === 1 ? "" : "s"}` : category.warnings ? `${category.warnings} warning${category.warnings === 1 ? "" : "s"}` : "ready"}</Badge></div></div><div className="preflight-list">{category.checks.map((check) => <article key={check.id} className={`preflight-check ${check.status}`}><div className="preflight-check-head"><div><strong>{check.title}</strong>{check.resource ? <p className="muted">{titleForProbe(check.scope, check.resource)}</p> : null}</div><div className="preflight-check-badges"><Badge tone={toneForCheck(check.status)}>{labelForCheck(check.status)}</Badge><Badge tone={check.required ? "accent" : "neutral"}>{check.required ? "required" : "optional"}</Badge></div></div><p>{check.detail}</p>{check.hint ? <p className="muted preflight-hint">{check.hint}</p> : null}<KeyValueGrid items={preflightFacts(check)} />{check.commands?.length ? <div className="preflight-code-list">{check.commands.map((command, index) => <CodeBlock key={`${check.id}-command-${index}`} label={index === 0 ? "Check command" : "Additional command"} code={command} />)}</div> : null}{check.manifest ? <CodeBlock label="Least-privilege starting point" code={check.manifest} /> : null}</article>)}</div></section>)}
      </div>
    </div>
  );
}

function ConnectionChoiceGroup(props: { title: string; description: string; children: ReactNode }) {
  return (
    <section className="connection-choice-group">
      <div className="connection-choice-heading">
        <h4>{props.title}</h4>
        <p className="muted">{props.description}</p>
      </div>
      <div className="connection-modes">{props.children}</div>
    </section>
  );
}

function ConnectionChoiceCard(props: {
  method: ConnectionMethod;
  active: boolean;
  onSelect: (method: ConnectionMethod) => void;
  recommended?: boolean;
  detail?: string;
}) {
  const choice = connectionChoices[props.method];
  return (
    <label className={`mode-card ${props.active ? "is-active" : ""} ${props.recommended ? "is-recommended" : ""}`}>
      <input type="radio" name="connectionMethod" checked={props.active} onChange={() => props.onSelect(props.method)} aria-label={choice.label} />
      <div className="mode-card-head">
        <span className="eyebrow">{props.recommended ? "Recommended" : choice.eyebrow}</span>
        {props.method === "api_endpoint" ? <Badge tone="high">Advanced</Badge> : null}
      </div>
      <strong>{choice.label}</strong>
      <span>{props.detail || choice.description}</span>
    </label>
  );
}

function connectionChoiceDetail(method: ConnectionMethod, advisor: ConnectionAdvisor | null) {
  if (!advisor) {
    return connectionChoices[method].description;
  }

  switch (method) {
    case "current":
      return (
        advisor.currentLoginWarning ||
        advisor.currentLoginDetail ||
        "No working local Kubernetes access was detected yet on this machine."
      );
    case "kubeconfig_file":
      if (advisor.defaultKubeconfigAvailable && advisor.defaultKubeconfigPortable === false) {
        return `${advisor.defaultKubeconfigWarning} Load another kubeconfig or one with embedded *-data fields.`;
      }
      if (advisor.defaultKubeconfigPath) {
        return `Detected kubeconfig at ${advisor.defaultKubeconfigPath}. K8V validates content, so names like config, prod-cluster.backup, and files with no extension still work.`;
      }
      return "Valid kubeconfig content is accepted even if the file is named config, prod-cluster.backup, or has no extension.";
    default:
      return connectionChoices[method].description;
  }
}

function machineReadinessItems(connectionAdvisor: ConnectionAdvisor | null) {
  if (!connectionAdvisor) {
    return [
      {
        label: "Local access",
        value: "Checking",
        detail: "K8V is still checking whether current login or a default kubeconfig is available here.",
        state: "neutral" as const,
      },
    ];
  }

  return [
    {
      label: "Existing access",
      value: connectionAdvisor.currentContext || (connectionAdvisor.currentLoginAvailable ? "Detected" : "Not detected"),
      detail:
        connectionAdvisor.currentLoginWarning ||
        connectionAdvisor.currentLoginDetail ||
        "No working local Kubernetes access was detected yet on this machine.",
      state: connectionAdvisor.currentLoginAvailable
        ? connectionAdvisor.currentLoginWarning
          ? ("caution" as const)
          : ("ready" as const)
        : ("missing" as const),
    },
    {
      label: "Default kubeconfig",
      value: connectionAdvisor.defaultKubeconfigPath || "Not found",
      detail:
        connectionAdvisor.defaultKubeconfigWarning ||
        connectionAdvisor.defaultKubeconfigDetail ||
        "No default kubeconfig was detected in the standard local locations.",
      state: connectionAdvisor.defaultKubeconfigAvailable
        ? connectionAdvisor.defaultKubeconfigPortable === false
          ? ("caution" as const)
          : ("ready" as const)
        : ("missing" as const),
    },
    {
      label: "kubectl CLI (optional)",
      value: connectionAdvisor.kubectlAvailable ? "Detected" : "Not detected",
      detail: connectionAdvisor.kubectlAvailable
        ? connectionAdvisor.kubectlPath || "Useful for endpoint discovery and token commands, but not required for the scan itself."
        : "Helpful for endpoint discovery and token commands, but K8V can still scan with a kubeconfig or direct API setup.",
      state: connectionAdvisor.kubectlAvailable ? ("ready" as const) : ("neutral" as const),
    },
  ];
}

function ArtifactToggle(props: { label: string; description: string; checked: boolean; onChange: (checked: boolean) => void }) {
  return (
    <label className="toggle-card artifact-choice">
      <input type="checkbox" checked={props.checked} onChange={(event) => props.onChange(event.target.checked)} />
      <div>
        <strong>{props.label}</strong>
        <p className="muted">{props.description}</p>
      </div>
    </label>
  );
}

function InspectionSummary(props: { inspection: KubeconfigInspection | null }) {
  if (!props.inspection) {
    return null;
  }
  return (
    <div className="notice notice-info compact">
      <strong>Kubeconfig inspection</strong>
      <p className="muted">{props.inspection.summary}</p>
      <KeyValueGrid items={[["Current context", props.inspection.currentContext || "Not set"], ["Contexts", String(props.inspection.contexts?.length || 0)], ["Clusters", String(props.inspection.clusterCount)], ["Users", String(props.inspection.userCount)]]} className="compact-kv-grid" />
      {props.inspection.missingReferencedFiles?.length ? (
        <div className="notice notice-warning compact nested-notice">
          <strong>Missing local file dependencies</strong>
          <p className="muted">This kubeconfig still expects local CA or client-certificate files that are not present here.</p>
          <ul className="notice-list compact-list">
            {props.inspection.missingReferencedFiles.map((reference) => (
              <li key={reference}>{reference}</li>
            ))}
          </ul>
        </div>
      ) : null}
      {props.inspection.referencedFiles?.length && !props.inspection.missingReferencedFiles?.length ? (
        <div className="reference-list">
          <p className="muted assistant-footnote">Referenced local CA or client-certificate files:</p>
          <ul className="notice-list compact-list">
            {props.inspection.referencedFiles.map((reference) => (
              <li key={reference}>{reference}</li>
            ))}
          </ul>
        </div>
      ) : null}
      {props.inspection.nextAction ? <p className="muted assistant-footnote">{props.inspection.nextAction}</p> : null}
    </div>
  );
}

function ContextCatalogHint(props: { catalog: ContextCatalog | null }) {
  return (
    <div className="context-helper">
      <p className="muted context-helper-status">
        {props.catalog
          ? props.catalog.contexts?.length
            ? `Found ${props.catalog.contexts.length} context${props.catalog.contexts.length === 1 ? "" : "s"} from ${props.catalog.source || "this connection"}${props.catalog.currentContext ? `. Current: ${props.catalog.currentContext}.` : "."}`
            : `No named contexts were found for ${props.catalog.source || "this connection"}.`
          : "Load contexts if you want suggestions before the scan."}
      </p>
      {props.catalog?.contexts?.length ? <datalist id="scan-context-options">{props.catalog.contexts.map((context) => <option key={context} value={context} />)}</datalist> : null}
    </div>
  );
}

function buildReviewCards(scanForm: ScanRequest, authMode: { label: string; detail: string; tone: ScanReviewTone }, trustMode: { label: string; detail: string; tone: ScanReviewTone }, riskFlags: string[]) {
  const connectionMethod = scanForm.connectionMethod || "current";
  const cards: Array<{ label: string; value: string; detail?: string; tone?: ScanReviewTone }> = [];

  if (connectionMethod === "api_endpoint") {
    cards.push({ label: "Endpoint target", value: scanForm.apiServerEndpoint || "Missing endpoint", detail: "Use the Kubernetes API server address only.", tone: scanForm.apiServerEndpoint ? "neutral" : "critical" });
    cards.push({ label: "Auth mode", value: authMode.label, detail: authMode.detail, tone: authMode.tone });
    cards.push({ label: "TLS trust", value: trustMode.label, detail: trustMode.detail, tone: trustMode.tone });
  } else {
    cards.push({ label: "Connection", value: connectionSummary(scanForm), detail: "Credential and trust handling come from existing access or the selected kubeconfig.", tone: "neutral" });
  }

  cards.push({ label: "Namespace scope", value: scopeSummary(scanForm.namespaces), detail: scanForm.namespaces?.length ? "Only the listed namespaces will be collected." : "All readable namespaces are in scope.", tone: scanForm.namespaces?.length ? "neutral" : "high" });
  cards.push({ label: "Compare baseline", value: scanForm.compareTo ? "Loaded" : "None selected", detail: scanForm.compareTo || "Add a previous bundle only when you want delta and regression analysis.", tone: scanForm.compareTo ? "success" : "neutral" });
  cards.push({ label: "Exports", value: reportSummary(scanForm), detail: "These artifacts will be refreshed after the run completes.", tone: "neutral" });
  cards.push({ label: "Output", value: scanForm.outputDir || "Missing output path", detail: "Bundles and refreshed exports land here.", tone: scanForm.outputDir ? "neutral" : "critical" });
  cards.push({ label: "Risk flags", value: riskFlags.length ? `${riskFlags.length} attention point${riskFlags.length === 1 ? "" : "s"}` : "None detected", detail: riskFlags.length ? riskFlags.join(" ") : "No obvious transport, scope, or trust risks are flagged right now.", tone: riskFlags.length ? "high" : "success" });
  return cards;
}

function buildSteps(current: ScanStage) {
  const order: ScanStage[] = ["connect", "validate", "outputs", "launch"];
  const index = order.indexOf(current);
  return [
    { id: "connect", label: "Choose connection", description: "Use existing access, kubeconfig, or manual API mode.", status: index === 0 ? "current" : index > 0 ? "complete" : "upcoming" },
    { id: "validate", label: "Validate connection", description: "Test reachability, credentials, and TLS.", status: index === 1 ? "current" : index > 1 ? "complete" : "upcoming" },
    { id: "outputs", label: "Choose scope and outputs", description: "Decide what to collect and what to write.", status: index === 2 ? "current" : index > 2 ? "complete" : "upcoming" },
    { id: "launch", label: "Preflight and start", description: "Run readiness checks, then launch the scan.", status: index === 3 ? "current" : "upcoming" },
  ] as Array<{ id: ScanStage; label: string; description: string; status: "current" | "complete" | "upcoming" }>;
}

function stageTitle(stage: ScanStage) {
  switch (stage) {
    case "connect":
      return "Pick the simplest working connection path";
    case "validate":
      return "Prove the connection works before you go deeper";
    case "outputs":
      return "Set scope and output defaults without over-configuring";
    default:
      return "Run final preflight and launch the scan";
  }
}

function stageDescription(stage: ScanStage) {
  switch (stage) {
    case "connect":
      return "Only the controls for the selected access path stay visible. Advanced manual setup appears only when you choose direct API mode.";
    case "validate":
      return "This lightweight test is the fast answer to whether K8V can reach the cluster with the current connection settings.";
    case "outputs":
      return "Use sensible defaults, then add only the extra exports and advanced knobs you actually need.";
    default:
      return "Preflight checks final transport, RBAC, scope, and collection readiness. The real scan writes the bundle and exports afterwards.";
  }
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

function plannedArtifactsForScan(scanForm: ScanRequest) {
  const artifacts = [
    { name: "recovery-scan.json", detail: "Portable bundle with the collected scan evidence. This is the main file you reopen later." },
    { name: "recovery-report.html", detail: "Primary operator-facing report view for offline review and handoff." },
  ];
  if (scanForm.summary) artifacts.push({ name: "recovery-summary.html", detail: "Condensed summary for leadership, tickets, or quick sharing." });
  if (scanForm.runbook) artifacts.push({ name: "recovery-runbook.html", detail: "Operator-oriented runbook with follow-up and recovery guidance." });
  if (scanForm.csvExport) artifacts.push({ name: "csv/", detail: "CSV tables for spreadsheet analysis and evidence export." });
  if (scanForm.redact) artifacts.push({ name: "recovery-report-redacted.html", detail: "Redacted report for lower-sensitivity sharing." });
  return artifacts;
}

function reportSummary(scanForm: ScanRequest) {
  const enabled = [];
  if (scanForm.summary) enabled.push("summary");
  if (scanForm.runbook) enabled.push("runbook");
  if (scanForm.csvExport) enabled.push("csv");
  if (scanForm.redact) enabled.push("redacted");
  return enabled.length ? enabled.join(", ") : "bundle + report only";
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
  const blocker = preflight.checks.find((check) => check.status === "fail");
  if (blocker) return blocker.hint || blocker.detail;
  const warning = preflight.checks.find((check) => check.status === "warn");
  if (warning) return warning.hint || warning.detail;
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
      return scanForm.contextName ? `Existing access · ${scanForm.contextName}` : "Existing access";
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
