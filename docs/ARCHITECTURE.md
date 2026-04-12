# Architecture

`k8s-recovery-visualizer` is now organized around a shared application core so the CLI and desktop app execute the same scan and artifact pipeline.

## Layers

1. `cmd/scan`
   The supported CLI entrypoint.
   It remains intentionally thin and delegates orchestration to `internal/scanapp`.

2. `internal/scanapp`
   CLI-facing option parsing, validation, profile selection, and result printing.
   Converts command-line flags into `internal/appcore.ScanRequest`.

3. `internal/appcore`
   Shared service layer for CLI and desktop.
   Owns:
   - scan orchestration
   - deterministic dry-run fixtures
   - preflight and RBAC probing
   - comparison loading
   - output writing and export refresh
   - workspace and history loading
   - typed progress and warning events

4. `desktop/`
   Wails v2 desktop shell.
   The Go backend binds typed methods and emits scan events.
   The React + TypeScript frontend renders the Home, Projects, New Scan, Live Run, Results, and Settings surfaces.

5. `internal/theme`
   Shared theme tokens used by report generation and desktop bootstrap.
   This keeps the offline reports and GUI on the same palette, typography, and radius system.

6. `internal/output`
   HTML, Markdown, CSV, summary, runbook, and redacted artifact writers.
   `report.go`, `summary.go`, `runbook.go`, and the legacy `html.go` now consume the centralized theme tokens.

7. Analysis and remediation packages
   - `internal/analyze`
   - `internal/backup`
   - `internal/compare`
   - `internal/enrich`
   - `internal/history`
   - `internal/remediation`
   - `internal/restore`

## Scan Flow

1. Parse options
   `internal/scanapp/options.go` owns CLI flags and validation.

2. Build a shared request
   `ScanRequest` carries kubeconfig/context, namespace scope, output options, compare path, profile, and dry-run flags.

3. Run preflight
   `internal/appcore.Preflight` validates kubeconfig loading, API reachability, and key RBAC capabilities using `SelfSubjectAccessReview`.
   Optional failures are surfaced as degraded mode instead of silent omission.

4. Collect inventory
   Live runs use Kubernetes clients directly.
   Dry runs use deterministic fixtures so screenshots, smoke tests, and frontend previews stay stable.

5. Analyze and compare
   Restore simulation, backup assurance, scoring, findings, remediation, and bundle comparison are applied to the same `model.Bundle`.

6. Write artifacts
   The shared service writes:
   - `recovery-scan.json`
   - `recovery-enriched.json`
   - `recovery-report.html`
   - `recovery-report.md`
   - optional summary, runbook, redacted, and CSV artifacts

7. Record history
   History snapshots are written under `out/history/` and surfaced in both the report and the desktop history dashboard.

## Desktop Data Flow

- The Wails backend binds typed methods such as `RunScan`, `RunPreflight`, `OpenBundle`, `ExportBundle`, and `ListProjects`.
- Live progress is delivered as structured Wails events instead of shelling out to the CLI.
- The frontend can also run in deterministic mock mode for fixture-backed testing and screenshot generation.

## Compatibility Contracts

- `recovery-scan.json` schema: [`../schemas/recovery-scan-3.0.0.schema.json`](../schemas/recovery-scan-3.0.0.schema.json)
- `recovery-enriched.json` schema: [`../schemas/recovery-enriched-1.1.0.schema.json`](../schemas/recovery-enriched-1.1.0.schema.json)

`v1.4.0` preserves both published schema versions.

## Design Choices

- Conservative backup claims: detection does not equal verified recoverability.
- Offline-first outputs: reports and desktop bundles do not depend on a CDN.
- Shared core: the GUI does not shell out to the CLI for normal operation.
- Deterministic fixtures: screenshot and smoke-test output stays stable across runs.
- Centralized theme tokens: the desktop app and reports share one backend-defined visual system.
