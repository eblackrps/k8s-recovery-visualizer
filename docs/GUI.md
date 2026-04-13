# Desktop App

The desktop app ships to users as `K8V`. It lives in [`desktop/`](../desktop) and uses Wails v2 for the shell plus React + TypeScript for the frontend. It shares the same Go execution path as the CLI, so live scans, dry runs, exports, and bundle loading stay aligned across both surfaces.

## What The Desktop App Is For

Use the desktop app when you want:

- a guided scan wizard instead of CLI flags
- a preflight and RBAC check before running
- live progress, structured logs, warning surfacing, and cancel support
- a results workspace that mirrors the report tabs
- easy inspection of existing bundles without needing live cluster access

## Main Screens

- `Home / Projects`: recent bundles, quick actions, history, and workspace discovery
- `New Scan`: wizard for kubeconfig/context, namespace scope, profile, compare baseline, outputs, redaction, summary/runbook, and dry-run settings
- `Live Run`: progress events, warnings, structured logs, and cancel
- `Results`: Summary, Nodes, Workloads, Storage, Networking, Config, Images, Backup, DR Score, Findings, Remediation, and Compare
- `Settings`: workspace defaults plus an open-existing-bundle path

## Shared Backend Contract

The Wails backend binds typed methods from `desktop/app.go` and delegates real work to `internal/appcore`.

Key methods:

- `GetBootstrap`
- `GetSettings`
- `SaveSettings`
- `ListProjects`
- `RunPreflight`
- `RunScan`
- `CancelRun`
- `OpenBundle`
- `ExportBundle`

Live run updates are emitted as the `scan:event` Wails event.

## Exports And Bundle Workflows

The desktop app can:

- open an existing bundle directory without cluster access
- export only the requested outputs instead of rewriting everything blindly
- refresh HTML, Markdown, summary, runbook, CSV, redacted, and JSON artifacts from a loaded bundle
- preserve the same theme and offline-friendly output format as the CLI-generated reports

## Accessibility And Navigation

- wizard controls use explicit labels
- primary navigation and tab rows expose semantic tab/tabpanel wiring
- tablists support keyboard arrow navigation plus `Home` and `End`
- focus-visible states are styled intentionally instead of relying on browser defaults
- filter buttons expose state with accessible pressed semantics

## Development

Install dependencies:

```bash
make frontend-install
```

Run dev mode:

```bash
make dev-gui
```

Build the current-host app:

```bash
make build-gui
```

The packaged executable name is `K8V` on supported desktop targets.

Do not use `go build` directly in [`desktop/`](../desktop). Wails requires its own build path, and a plain Go build will produce a binary that exits with a build-tags error dialog instead of launching the app.

Run frontend tests:

```bash
make frontend-test
```

## Fixture Mode

The frontend supports a deterministic mock backend for tests, local previews, and screenshot generation.

- screenshots are generated from the fixture-backed browser build
- the fixture includes history, comparison data, findings, remediation steps, and backup evidence
- fixture timestamps, locale, timezone, and motion settings are fixed so the screenshot set stays reproducible
- no live cluster access is required for the screenshot workflow

Screenshot workflow: [SCREENSHOTS.md](SCREENSHOTS.md)
