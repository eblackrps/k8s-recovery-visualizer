# Development Guide

This project ships a source-first Go CLI and a Wails desktop app, packaged in-product as `K8V`, backed by the same shared Go service layer. Public GitHub releases now focus on Linux and Windows desktop packages, while the CLI remains available for local and CI source builds.

## Prerequisites

- Go `1.25+`
- Node.js `24` recommended (`.nvmrc` is provided at the repo root)
- Wails v2 CLI
- Playwright Chromium for screenshot generation
- NSIS if you want to package the Windows desktop app locally
- Kubernetes credentials if you want to exercise live scans

## Local Setup

Install frontend dependencies:

```bash
make frontend-install
```

Build the CLI for the current host:

```bash
make build
```

Build the contributor cross-platform CLI set:

```bash
make build-cli-cross
```

Run the desktop app in dev mode:

```bash
make dev-gui
```

Build the desktop app:

```bash
make build-gui
```

Package the current-host desktop app:

```bash
make package-gui
```

Do not run `go build` directly inside `desktop/`. Wails desktop builds require `wails build` or the matching Make target so the correct build tags and packaging steps are applied.

On Windows, the maintained `make build-gui` and `make package-gui` targets include the `native_webview2loader` tag so local packages match the validated release runtime behavior on managed Windows 11 machines.

## Repo Map

- `cmd/scan`: source-first CLI scan entrypoint
- `cmd/check`: CLI gate and policy evaluation entrypoint
- `internal/scanapp`: CLI-facing option parsing and orchestration
- `internal/appcore`: shared scan, preflight, workspace, history, and export logic
- `desktop/`: Wails shell and React frontend
- `internal/theme`: shared report and GUI design tokens
- `internal/output`: report, summary, runbook, CSV, and redacted artifact writers
- `docs/`: operator, contributor, and release documentation

Architecture details: [ARCHITECTURE.md](ARCHITECTURE.md)

## Validation Workflow

Run the core validation stack before opening a PR:

```bash
make fmt
go build ./...
make vet
make test
make race
make frontend-install
make frontend-build
make frontend-test
make screenshots
make smoke
make schema-samples
make docs-check
make build-gui
```

If you changed the desktop UI or README screenshots, keep the generated images under `images/` in sync with the app state.

## Docs And Screenshot Hygiene

- Keep the main [../README.md](../README.md) aligned with the actual public release scope and supported assets.
- Refresh screenshots with `make screenshots` instead of editing or exporting ad hoc images by hand.
- Treat the maintained screenshot inventory in [SCREENSHOTS.md](SCREENSHOTS.md) as the source of truth for public docs.
- Remove deprecated screenshots from docs and README references when the maintained gallery changes.

## Working Rules

- Keep the CLI in `cmd/scan` and `cmd/check` working and backward compatible.
- Keep `internal/scanapp` thin.
- Prefer shared logic in `internal/appcore` instead of creating GUI-only execution paths.
- Do not break published JSON schemas without versioning and documentation updates.
- Keep schema references aligned with the current published contracts (`recovery-scan 3.1.0`, `recovery-enriched 1.2.0`) when the JSON surface changes.
- Keep reports offline-friendly.
- Route palette and styling changes through `internal/theme`.
- Keep public release messaging aligned with the actual shipped artifacts.
- Keep docs and support messaging aligned with the maintained screenshot set and desktop-only public release posture.

## Testing Notes

- `go test ./...` covers the CLI, shared backend, and report generation.
- `go test -race ./...` is part of the local and CI validation stack.
- Frontend tests live under `desktop/frontend/src`.
- Deterministic fixture data powers the frontend tests and screenshot workflow.
- CI validates Linux and Windows desktop packaging before any tag-triggered release is cut.

## Release-Maintainer Notes

- Update `CHANGELOG.md` for every tagged release.
- Keep README, docs, screenshots, and workflow expectations aligned with the actual shipped artifacts.
- Use [RELEASE.md](RELEASE.md) for the tag-and-publish checklist.
