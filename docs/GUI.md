# Desktop App

The desktop app lives in [`desktop/`](../desktop) and uses Wails v2 for the shell plus React + TypeScript for the frontend.

## Goals

- preserve the supported CLI workflow
- reuse the same Go scan pipeline instead of shelling out
- keep the same palette and overall feel as the generated reports
- make prior bundles explorable without cluster access
- keep report exports offline-friendly

## Screens

- `Home`: score summary, recent history, and quick actions
- `Projects`: discovered scan bundles under the configured workspace root
- `New Scan`: guided wizard for access, scope, outputs, and review
- `Live Run`: progress, structured logs, warnings, and cancel support
- `Results`: report-aligned workspace with Summary, Nodes, Workloads, Storage, Networking, Config, Images, Backup, DR Score, Findings, Remediation, and Compare
- `Settings`: workspace defaults and open-existing-bundle flow

## Backend Contract

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

## Accessibility And Navigation

- wizard controls use explicit labels
- primary navigation and tabs expose semantic labeling
- keyboard arrow navigation is supported for tab rows
- the UI keeps contrast aligned with the shared report palette

## Development

Install dependencies:

```bash
make frontend-install
```

Run dev mode:

```bash
make dev-gui
```

Build the current host app:

```bash
make build-gui
```

Frontend tests:

```bash
make frontend-test
```

## Fixture Mode

The frontend supports a deterministic mock backend for tests, local previews, and screenshot generation.

- screenshots are generated from the fixture-backed browser build
- the fixture includes history, comparison data, findings, remediation steps, and backup evidence
- this keeps the screenshot set stable without requiring cluster access

Screenshot workflow: [SCREENSHOTS.md](SCREENSHOTS.md)
