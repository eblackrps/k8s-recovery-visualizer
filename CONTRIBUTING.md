# Contributing

Thanks for helping improve `k8s-recovery-visualizer`.

## Development Setup

Prerequisites:

- Go `1.25+`
- Node.js `20+`
- Wails v2 CLI
- Kubernetes credentials if you want to run live scans

Install the frontend dependencies:

```bash
make frontend-install
```

## Common Commands

- `make build`
- `make build-gui`
- `make dev-gui`
- `make test`
- `make frontend-test`
- `make smoke`
- `make docs-check`
- `make ci`

## Contribution Rules

- Keep the CLI in `cmd/scan` working and backward compatible.
- Keep `internal/scanapp` thin.
- Prefer shared Go functionality in `internal/appcore` over GUI-only execution paths.
- Do not break published JSON schemas without versioning and documentation updates.
- Keep generated reports offline-friendly.
- Keep the report and desktop palette aligned through `internal/theme`.
- Preserve existing report screenshots in `images/`.

## Before Opening A PR

Run:

```bash
make fmt
make vet
make test
make frontend-build
make frontend-test
make smoke
make schema-samples
make docs-check
```

If you change the desktop UI, refresh the fixture screenshots:

```bash
make screenshots
```

## Schema Changes

If a contribution changes emitted JSON:

- update the relevant schema file
- update committed schema examples
- document the version bump in `CHANGELOG.md`
- explain compatibility impact in the PR

## Theme Changes

Report and GUI colors should be updated through `internal/theme`.
Avoid hard-coding a new parallel palette in the frontend or report writers.
