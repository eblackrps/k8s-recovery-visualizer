# Development Guide

This project ships a supported Go CLI and a Wails desktop app backed by the same shared Go service layer. The most useful contributor workflow is to keep both surfaces healthy together.

## Prerequisites

- Go `1.25+`
- Node.js `22` recommended
- Wails v2 CLI
- Playwright Chromium for screenshot generation
- Kubernetes credentials if you want to exercise live scans

## Local Setup

Install frontend dependencies:

```bash
make frontend-install
```

Build the CLI:

```bash
make build
```

Run the desktop app in dev mode:

```bash
make dev-gui
```

Build the desktop app with:

```bash
make build-gui
```

Do not run `go build` directly inside `desktop/`. Wails desktop builds require `wails build` or the matching Make target so the correct build tags and packaging steps are applied.

## Repo Map

- `cmd/scan`: supported CLI entrypoint
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
make frontend-build
make frontend-test
make smoke
make schema-samples
make docs-check
make build-gui
```

If you changed the desktop UI or README screenshots:

```bash
make screenshots
```

## Working Rules

- Keep the CLI in `cmd/scan` working and backward compatible.
- Keep `internal/scanapp` thin.
- Prefer shared logic in `internal/appcore` instead of creating GUI-only execution paths.
- Do not break published JSON schemas without versioning and documentation updates.
- Keep reports offline-friendly.
- Route palette and styling changes through `internal/theme`.

## Testing Notes

- `go test ./...` covers the CLI, shared backend, and report generation.
- `go test -race ./...` is part of the local and CI validation stack.
- Frontend tests live under `desktop/frontend/src`.
- Deterministic fixture data powers the frontend tests and screenshot workflow.

## Release-Maintainer Notes

- Update `CHANGELOG.md` for every tagged release.
- Keep README, docs, screenshots, and workflow expectations aligned with the actual shipped artifacts.
- Use [RELEASE.md](RELEASE.md) for the tag-and-publish checklist.
