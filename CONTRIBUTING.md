# Contributing

Thanks for helping improve `k8s-recovery-visualizer`.

Please read these first:

- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

## Contribution Priorities

- Keep the CLI in `cmd/scan` working and backward compatible.
- Keep `internal/scanapp` thin.
- Prefer shared functionality in `internal/appcore` over GUI-only execution paths.
- Do not break published JSON schemas without versioning and documentation updates.
- Keep generated reports offline-friendly.
- Route palette changes through `internal/theme`.
- Preserve the supported screenshot set under `images/`.

## Validation Before Opening A PR

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

If you changed the desktop UI, README visuals, or screenshot documentation:

```bash
make screenshots
```

## Schema And Output Changes

If your contribution changes emitted JSON or generated artifacts:

- update the relevant schema files when required
- update committed schema examples when required
- document the compatibility impact in `CHANGELOG.md`
- explain the upgrade or migration impact in the PR

## Pull Request Expectations

- keep PRs focused and reviewable
- include validation results
- include screenshots when user-facing UI changes
- update README or docs when workflows, assets, or terminology change

Need help first? Start with [SUPPORT.md](SUPPORT.md).
