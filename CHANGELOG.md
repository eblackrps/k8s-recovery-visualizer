# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

- No unreleased changes yet.

## [1.5.1] - 2026-04-13

Desktop polish release for `k8s-recovery-visualizer`: this patch release tightens the Windows desktop shell, brings the app back into alignment with the shipped report visual language, fixes broken bundle-open actions, and refreshes the public README gallery to match the current product.

### Improved

- Tightened the Windows desktop shell so the title bar stays visible on smaller displays and the packaged app presents itself consistently as `K8V`.
- Refined the desktop workspace styling to better match the supported report palette and reduce the earlier website-in-a-window feel.
- Updated the main README screenshot gallery so GitHub now shows the current guided scan, live run, results, compare, and report visuals.

### Fixed

- Corrected desktop bundle-open flows so top-level open actions launch the file picker instead of falling back to stale demo bundle paths.
- Improved disabled and error states for desktop actions such as run cancellation and bundle open so unavailable actions no longer fail silently.
- Cleaned up the scan wizard stepper, metric cards, empty states, and desktop icon resources for a more production-ready Windows experience.

### Docs

- Refreshed release-facing docs and issue templates to target the current patch release.
- Added the updated desktop screenshot set to the main repository landing page.

## [1.5.0] - 2026-04-12

Desktop workspace release for `k8s-recovery-visualizer`: the project now ships as a polished dual-surface product with a supported Go CLI, a Wails desktop app, shared backend execution, refreshed public docs, and a more professional GitHub/release presentation.

### Added

- Shared `internal/appcore` service layer for scan execution, preflight, workspace loading, history, export refresh, and typed run events.
- Wails v2 desktop app in `desktop/` with React + TypeScript screens for Home, Projects, New Scan, Live Run, Results, and Settings.
- Guided scan wizard with kubeconfig/context, namespace scope, profile, baseline compare path, output path, redaction, summary/runbook, CSV export, and dry-run controls.
- RBAC and preflight assistant with degraded-mode guidance.
- Live progress, structured warning surfacing, cancel support, and open-existing-bundle support.
- Deterministic frontend fixtures and automated GUI screenshot generation.
- New public-facing docs for CLI usage, development, release maintenance, screenshots, support, and community workflow.
- GitHub community files including Code of Conduct, support guidance, CODEOWNERS, issue templates, PR template, Dependabot config, and editor configuration.

### Improved

- Centralized report and desktop design tokens in `internal/theme`.
- Refactored `internal/output/report.go`, `summary.go`, and `runbook.go` to consume shared palette tokens.
- Kept CLI orchestration thin by routing `cmd/scan` through `internal/scanapp` and `internal/appcore`.
- Refreshed the README so the repository presents clearly as a CLI plus desktop product with stronger install, quickstart, output, and docs navigation.
- Hardened workflow and docs consistency across release notes, screenshots, support guidance, and packaging expectations.

### Fixed

- Shared-core export behavior so GUI exports write only the requested artifact set and do not mutate loaded bundles.
- Dry-run finalization so findings and remediation are generated once, not duplicated.
- Desktop settings persistence so partial settings files do not wipe newer defaults.
- Accessibility wiring for tabs, tabpanels, keyboard navigation, focus states, and empty-state behavior in the desktop app.
- Deterministic screenshot timestamps, locale, timezone, and fixture content so GUI captures stay reproducible.
- CI and release gate parity so the important verification steps match before publish.

### Docs

- Added a documentation index and clearer navigation between operator, contributor, and release-maintainer docs.
- Updated support, security, and contributing guidance to match the current CLI plus desktop product.
- Removed stale public screenshots that no longer matched the current visual language.

### Packaging And Release

- Standardized desktop packaging names around `k8s-recovery-visualizer-desktop`.
- Added predictable GitHub issue and PR templates for public repository workflow.
- Improved release notes generation and Windows desktop asset archiving for cleaner release presentation.
- Continued publishing checksums, SBOMs, and container metadata in GitHub Actions.

### Compatibility

- CLI flags remain backward compatible.
- Published JSON schema versions remain `3.0.0` for `recovery-scan.json` and `1.1.0` for `recovery-enriched.json`.
- Generated report outputs remain offline-friendly.
