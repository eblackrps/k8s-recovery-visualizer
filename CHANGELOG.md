# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

Release candidate changes for the next minor line focus on operator actionability, compare/history depth, restore-readiness evidence, and stronger shared-core policy/reporting behavior.

### Improved

- Added finding prioritization metadata across the shared bundle model, reports, runbook, and desktop results workspace, including `rank`, `impact`, `ownerHint`, `effort`, and `priorityScore`.
- Expanded restore-readiness evidence to classify namespaces as ready, warning, blocked, uncovered, or unknown and to calculate namespace counts, blocking reasons, and estimated data at risk.
- Added a generated restore drill plan to backup inventory so scan bundles, reports, runbooks, and desktop views can turn evidence into an executable recovery checklist.
- Expanded compare summaries with per-domain score deltas, severity deltas, inventory deltas, regressed findings, improved findings, and persistent finding counts.
- Expanded history/trend data so bundle history now carries per-domain scores and finding counts, plus a richer history dashboard in the desktop workspace.
- Refactored backup detection behind pluggable collectors to make future backup integrations easier to add without rewriting the main detection loop.
- Improved desktop bundle loading so archive imports are validated earlier and surface clearer corruption, ambiguity, and mispackaging diagnostics.
- Expanded `cmd/check` policy gates with domain thresholds, severity budgets, new-finding budgets, and regression budgets.
- Strengthened desktop/frontend regression coverage with focused view tests for findings, compare, restore readiness, and preflight remediation rendering.

### Docs

- Refreshed the main README and supporting docs so the public desktop-only release posture, CLI source-build story, schema references, and support language read consistently across the repo.
- Updated schema docs, committed examples, and CI/release workflow validation to the published `recovery-scan 3.1.0` and `recovery-enriched 1.2.0` contracts.
- Replaced older report-era README screenshots with the maintained deterministic desktop screenshot set and removed deprecated screenshot references.

### Fixed

- Restored usable `cmd/check --help` output so the current gate set can be audited directly from the terminal instead of failing with a bare help error.
- Hardened the release workflow so tagged releases validate the exact four supported assets and prune stale uploaded assets before publishing, preventing stale release-surface drift on reruns.

## [1.7.1] - 2026-04-13

Hotfix release to address Windows launch failures on Win11 by switching the packaged desktop build to the legacy WebView2 loader.

### Fixed

- Built the Windows desktop package with the legacy WebView2 loader tag to improve launch reliability on Windows 11 machines where the new loader silently failed.

## [1.6.1] - 2026-04-13

Patch follow-up for `v1.6.0`: this release ships the post-release review fixes by moving GitHub Actions fully onto Node 24 and tightening the desktop runtime so scan and preflight work only from the app lifecycle context.

### Fixed

- Switched GitHub Actions workflow execution fully onto Node 24, including forcing JavaScript actions onto the Node 24 runner path so CI and release jobs no longer emit Node 20 deprecation warnings.
- Tightened the desktop runtime so preflight and scan entrypoints fail fast when the Wails lifecycle context is unavailable instead of silently falling back to a detached background context.
- Added focused desktop tests that lock in the missing-lifecycle failure path.

### Docs

- Added a repo-root `.nvmrc` and aligned contributor/development guidance around Node 24 and the reproducible frontend install path.

## [1.6.0] - 2026-04-13

Desktop-only public release for `k8s-recovery-visualizer`: this release narrows the published artifact surface to supported Linux and Windows desktop packages, keeps the CLI source-first for contributors and automation, refactors the desktop runtime for maintainability, hardens settings behavior, and refreshes the repository landing page, screenshots, and release documentation to match the supported product.

### Improved

- Simplified GitHub Actions so public releases publish only the Linux amd64 and Windows amd64 desktop packages plus checksums and an SPDX SBOM.
- Updated CI to validate both supported desktop packaging paths before release instead of waiting for tag-only packaging coverage.
- Refactored the desktop backend into focused settings, scan-control, dialog, bundle-loading, and window helpers.
- Improved bundle-open UX so the desktop app can open bundle directories, `recovery-scan.json`, and supported bundle archives.
- Standardized frontend dependency installation on `npm ci` across Wails, local workflows, CI, and release packaging.

### Fixed

- Replaced desktop scan and preflight `context.Background()` usage with app-lifecycle-derived contexts so active runs cancel more coherently.
- Hardened settings load/save behavior so failures are surfaced to the desktop UI instead of being silently ignored.
- Tightened desktop settings file permissions to safer per-user defaults where supported.
- Improved Linux default workspace behavior to prefer XDG-friendly locations when available.

### Docs

- Rewrote the main README to present `K8V` as the desktop product while keeping `k8s-recovery-visualizer` as the repository and archive identity.
- Refreshed the screenshot-backed landing page and aligned README, support, development, GUI, and release docs with the new public support policy.
- Documented the deprecation of public macOS desktop packages, prebuilt CLI binaries, and the GHCR container image.

## [1.5.2] - 2026-04-13

Maintenance release for `k8s-recovery-visualizer`: this patch release brings the latest tag back in sync with `main` after the post-`v1.5.1` automation follow-up and keeps the GitHub Actions stack aligned with GitHub's Node 24 migration path.

### Fixed

- Updated GitHub Actions workflows to use `actions/setup-node@v6` so CI and release automation no longer depend on the deprecated Node 20-based action runtime.
- Cut a follow-up maintenance release so the latest published tag matches the current `main` branch state.

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
