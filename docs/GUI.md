# Desktop App

The desktop app ships to users as `K8V`. It lives in [`desktop/`](../desktop), uses Wails v2 for the shell plus React + TypeScript for the frontend, and shares the same execution path as the CLI so live scans, dry runs, exports, and bundle loading stay aligned across both surfaces.

The current desktop UX intentionally favors a sober, denser operator console over decorative dashboard chrome. Surfaces are flatter, the shell is quieter, and the highest-value DR views are promoted ahead of inventory-heavy detail.

## Public Support

Supported public desktop release packages:

- Linux amd64
- Windows amd64

Public macOS desktop packages are deprecated for this release line. macOS-oriented Wails assets may still exist for source builds and contributor workflow, but no public macOS package is published.

## What The Desktop App Is For

Use the desktop app when you want:

- a remote-first scan setup instead of raw CLI flags
- a preflight and RBAC check before running
- live progress, structured logs, warning surfacing, and cancel support
- a quieter operator shell with one canonical active-bundle context strip in the header
- a results workspace organized by operator priority instead of raw report order
- prioritized findings, restore-readiness summaries, and drill-planning guidance in one place
- easy inspection of existing bundles without needing live cluster access

## Main Screens

- `Home / Projects`: task-first operator workspace with primary actions, current posture, a compact bundle history list, and an operational watch panel for regressions and trend changes
- `New Scan`: guided remote-first scan setup with four connection modes, namespace scope, labels, outputs, inline validation, advanced options, and a dedicated preflight rail
  - `Current login`: best for desktops and jumpboxes where `kubectl` or `KUBECONFIG` already works
  - `Kubeconfig file`: pick a kubeconfig from disk and optionally select a context
  - `Paste kubeconfig`: paste kubeconfig content directly when operators do not want to rely on local files
  - `API endpoint`: enter a control-plane host, IP, or URL directly and authenticate with a bearer token
  - context discovery can load named contexts from the current login or kubeconfig inputs before the scan starts
- `Live Run`: progress events, warnings, structured logs, and cancel
- `Results`: Overview, Findings, Restore Readiness, Compare, Inventory, and Remediation
  - `Overview` centralizes score, trend, backup posture, and top findings instead of forcing users through inventory tabs first
  - `Findings` uses dense filtering and a tighter findings table with severity, owner, impact, effort, resource, and recommendation data optimized for quick scanning
  - `Restore Readiness` consolidates backup coverage, restore simulation, blocker counts, and the drill plan into one operator-facing section
  - `Compare` makes score drift, regressions, improvements, persistent issues, and inventory drift legible at a glance
  - `Inventory` contains secondary navigation for Nodes, Workloads, Storage, Networking, Config, and Images
  - `Remediation` keeps the full guidance available while using denser disclosure rows instead of large always-open cards
- `Settings`: workspace defaults plus open-existing-bundle support

## Shared Backend Contract

The Wails backend keeps [`desktop/app.go`](../desktop/app.go) as wiring and app lifecycle glue, then splits settings, scan control, dialogs, bundle loading, and window helpers into focused files under [`desktop/`](../desktop).

Bound methods include:

- `GetBootstrap`
- `GetSettings`
- `GetStartupAlerts`
- `SaveSettings`
- `ListProjects`
- `RunPreflight`
- `ListConnectionContexts`
- `RunScan`
- `CancelRun`
- `OpenBundle`
- `ExportBundle`

Live run updates are emitted as the `scan:event` Wails event.

## Exports And Bundle Workflows

The desktop app can:

- open a bundle directory, `recovery-scan.json`, or a supported `.zip` / `.tar.gz` / `.tgz` archive without cluster access
- validate archives and JSON content before loading so users get actionable corruption or ambiguity diagnostics
- export only the requested outputs instead of rewriting everything blindly
- refresh HTML, Markdown, summary, runbook, CSV, redacted, and JSON artifacts from a loaded bundle
- preserve the same theme and offline-friendly output format as the CLI-generated reports

## Preflight And RBAC Guidance

The preflight panel is no longer just a pass/fail list. When key permissions are missing, `K8V` now shows:

- scope and resource context for the failing probe
- a suggested `kubectl auth can-i` command to confirm the gap
- a least-privilege manifest snippet when the missing permission maps cleanly to an RBAC rule

That keeps degraded-mode behavior explicit and gives platform teams a concrete starting point for access remediation.

## Connection Tips

- On a desktop or jumpbox, leave the app on `Current login` when `kubectl get nodes` already works from that machine.
- Use `Kubeconfig file` or `Paste kubeconfig` when operators receive a kubeconfig through a secure file handoff or vault workflow.
- Use `API endpoint` when you need to enter a control-plane host or IP directly. The current direct-endpoint mode is intentionally bearer-token based. If the cluster depends on exec plugins, cloud auth helpers, or client certificates, kubeconfig mode remains the better fit.
- Run **Detect Contexts** to load named contexts before preflight when you are using the current login or a kubeconfig source.

## Settings Behavior

- Startup now surfaces saved-settings load failures instead of silently swallowing them.
- Settings-save failures are reported back through the desktop UI without destroying the in-memory session state.
- Linux defaults prefer XDG-friendly locations when available.
- Persisted settings files are written with tighter per-user permissions.

## Accessibility And Navigation

- scan setup controls use explicit labels and inline help tips
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

Package the current-host app:

```bash
make package-gui
```

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

The maintained public screenshot set is the current `gui-*` desktop capture set documented in [SCREENSHOTS.md](SCREENSHOTS.md).
