import type { Settings } from "../lib/types";
import { Card, Field, SectionHeader } from "../components/ui";

export function SettingsView(props: {
  settings: Settings;
  busy: boolean;
  setSettings: (settings: Settings) => void;
  openBundlePath: string;
  setOpenBundlePath: (value: string) => void;
  onSave: () => void;
  onOpenBundle: () => void;
}) {
  const settings = props.settings;

  return (
    <section className="page-grid settings-grid">
      <section className="panel">
        <SectionHeader
          eyebrow="Settings"
          title="Workspace defaults"
          description="These defaults shape new scans and export behavior without changing the underlying bundle schema or collection workflow."
          actions={
            <button type="button" className="button primary" onClick={props.onSave} disabled={props.busy}>
              Save Settings
            </button>
          }
        />

        <div className="results-stack">
          <Card title="Scan defaults">
            <div className="form-grid">
              <Field label="Workspace Root">
                <input value={settings.workspaceRoot} onChange={(event) => props.setSettings({ ...settings, workspaceRoot: event.target.value })} />
              </Field>
              <Field label="Default Output Directory">
                <input value={settings.defaultOutputDir} onChange={(event) => props.setSettings({ ...settings, defaultOutputDir: event.target.value })} />
              </Field>
              <Field label="Default Profile">
                <select value={settings.defaultProfile} onChange={(event) => props.setSettings({ ...settings, defaultProfile: event.target.value })}>
                  <option value="standard">standard</option>
                  <option value="enterprise">enterprise</option>
                  <option value="dev">dev</option>
                  <option value="airgap">airgap</option>
                </select>
              </Field>
            </div>
          </Card>

          <Card title="Export defaults">
            <div className="toggle-grid">
              <label className="toggle"><input type="checkbox" checked={settings.summary} onChange={(event) => props.setSettings({ ...settings, summary: event.target.checked })} /> Summary export</label>
              <label className="toggle"><input type="checkbox" checked={settings.runbook} onChange={(event) => props.setSettings({ ...settings, runbook: event.target.checked })} /> Runbook export</label>
              <label className="toggle"><input type="checkbox" checked={settings.redact} onChange={(event) => props.setSettings({ ...settings, redact: event.target.checked })} /> Redacted export</label>
              <label className="toggle"><input type="checkbox" checked={settings.csvExport} onChange={(event) => props.setSettings({ ...settings, csvExport: event.target.checked })} /> CSV export</label>
              <label className="toggle"><input type="checkbox" checked={settings.includeSecretMetadata} onChange={(event) => props.setSettings({ ...settings, includeSecretMetadata: event.target.checked })} /> Secret metadata</label>
            </div>
          </Card>
        </div>
      </section>

      <section className="panel">
        <SectionHeader
          eyebrow="Bundle Tools"
          title="Open existing bundle"
          description="Inspect prior results without cluster access by loading a recovery-scan bundle, bundle directory, or supported archive."
        />
        <div className="results-stack">
          <Card title="Bundle path">
            <div className="inline-field">
              <input value={props.openBundlePath} onChange={(event) => props.setOpenBundlePath(event.target.value)} aria-label="Existing bundle path" />
              <button type="button" className="button secondary" onClick={props.onOpenBundle} disabled={props.busy}>
                Open
              </button>
            </div>
            <p className="muted">Open a recovery-scan.json bundle, a bundle directory, or a .zip / .tar.gz / .tgz archive.</p>
          </Card>
        </div>
      </section>
    </section>
  );
}
