import type { Settings } from "../lib/types";
import { Field } from "../components/ui";

export function SettingsView(props: {
  settings: Settings;
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
        <div className="section-header">
          <div>
            <p className="eyebrow">Settings</p>
            <h3>Desktop defaults</h3>
          </div>
          <button type="button" className="button primary" onClick={props.onSave}>
            Save Settings
          </button>
        </div>
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
          <div className="toggle-grid">
            <label className="toggle"><input type="checkbox" checked={settings.summary} onChange={(event) => props.setSettings({ ...settings, summary: event.target.checked })} /> Summary export</label>
            <label className="toggle"><input type="checkbox" checked={settings.runbook} onChange={(event) => props.setSettings({ ...settings, runbook: event.target.checked })} /> Runbook export</label>
            <label className="toggle"><input type="checkbox" checked={settings.redact} onChange={(event) => props.setSettings({ ...settings, redact: event.target.checked })} /> Redacted export</label>
            <label className="toggle"><input type="checkbox" checked={settings.csvExport} onChange={(event) => props.setSettings({ ...settings, csvExport: event.target.checked })} /> CSV export</label>
            <label className="toggle"><input type="checkbox" checked={settings.includeSecretMetadata} onChange={(event) => props.setSettings({ ...settings, includeSecretMetadata: event.target.checked })} /> Secret metadata</label>
          </div>
        </div>
      </section>

      <section className="panel">
        <div className="section-header">
          <div>
            <p className="eyebrow">Open Existing Bundle</p>
            <h3>Inspect prior results without cluster access</h3>
          </div>
        </div>
        <div className="inline-field">
          <input value={props.openBundlePath} onChange={(event) => props.setOpenBundlePath(event.target.value)} aria-label="Existing bundle path" />
          <button type="button" className="button secondary" onClick={props.onOpenBundle}>
            Open
          </button>
        </div>
      </section>
    </section>
  );
}
