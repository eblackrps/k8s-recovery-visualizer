# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

- No unreleased changes yet.

## [1.4.0] - 2026-04-12

### Added

- Shared `internal/appcore` service layer for scan execution, preflight, workspace loading, history, export refresh, and typed run events.
- Wails v2 desktop app in `desktop/` with React + TypeScript screens for Home, Projects, New Scan, Live Run, Results, and Settings.
- Guided scan wizard with kubeconfig/context, namespace scope, profile, baseline compare path, output path, redaction, summary/runbook, CSV export, and dry-run controls.
- RBAC and preflight assistant with degraded-mode guidance.
- Live progress, structured warning surfacing, cancel support, and open-existing-bundle support.
- Deterministic frontend fixtures and automated GUI screenshot generation.
- `docs/GUI.md`, `docs/SCREENSHOTS.md`, `docs/RELEASE.md`, and `CONTRIBUTING.md`.

### Changed

- Centralized report and desktop design tokens in `internal/theme`.
- Refactored `internal/output/report.go`, `summary.go`, and `runbook.go` to consume shared palette tokens, and archived retired legacy writers out of the build.
- Kept CLI orchestration thin by routing `cmd/scan` through `internal/scanapp` and `internal/appcore`.
- Expanded the Makefile and GitHub Actions workflows to cover frontend build/test, desktop packaging, docs validation, screenshots, checksums, and SBOM generation.
- Updated the README and architecture notes to describe the combined CLI + desktop product.
- Hardened shared-core export behavior so GUI exports write only the requested artifact set and do not mutate loaded bundles.
- Fixed dry-run finalization so findings and remediation are generated once, not duplicated.
- Improved desktop defaults, screenshot determinism, accessibility wiring, and release-gate parity.

### Compatibility

- CLI flags remain backward compatible.
- Published JSON schema versions remain `3.0.0` for `recovery-scan.json` and `1.1.0` for `recovery-enriched.json`.
