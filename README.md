# k8s-recovery-visualizer

[![CI](https://github.com/eblackrps/k8s-recovery-visualizer/actions/workflows/ci.yml/badge.svg)](https://github.com/eblackrps/k8s-recovery-visualizer/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/eblackrps/k8s-recovery-visualizer)](https://github.com/eblackrps/k8s-recovery-visualizer/releases)
[![License](https://img.shields.io/github/license/eblackrps/k8s-recovery-visualizer)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/eblackrps/k8s-recovery-visualizer)](go.mod)

`k8s-recovery-visualizer` is a Kubernetes disaster recovery assessment tool for platform teams, SREs, consultants, and audit-driven operators. It combines a Go CLI for automation and CI policy gates with a Wails desktop app for guided scans, live progress, bundle review, history, and exports. Both surfaces run on the same shared Go service layer and produce offline-friendly reports, versioned JSON contracts, compare/diff views, and remediation guidance.

<p align="center">
  <img alt="Desktop dashboard" src="images/gui-dashboard.png" width="49%" />
  <img alt="Offline HTML report summary" src="images/report-summary.png" width="49%" />
</p>

## What You Get

- A supported Go CLI for scripted scans, CI/CD gates, and air-gapped workflows
- A desktop app with Home, Projects, New Scan, Live Run, Results, and Settings views
- Offline HTML reports, executive summaries, runbooks, JSON bundles, CSV exports, and redacted artifacts
- Conservative backup evidence scoring, compare/history workflows, and schema-validated output contracts

## Install Options

| Option | Best for | How |
| --- | --- | --- |
| GitHub Releases | Operators and evaluators | Download the matching CLI binary or desktop package from [Releases](https://github.com/eblackrps/k8s-recovery-visualizer/releases/latest) |
| Build the CLI locally | Contributors and air-gapped environments | `make build` |
| Run the desktop app in dev mode | UI iteration and local evaluation | `make frontend-install && make dev-gui` |
| Build the desktop app locally | Packaging validation | `make frontend-install && make build-gui` |

## Quickstart

### CLI

Run a deterministic dry run:

```bash
go run ./cmd/scan --dry-run --summary --runbook --out ./out --min-score 0
```

Run a live scan with a named context and profile:

```bash
go run ./cmd/scan --context prod-east-admin --profile enterprise --summary --runbook --out ./out
```

Evaluate the generated bundle in CI:

```bash
go run ./cmd/check --current ./out/recovery-scan.json --min-score 85 --format json
```

If you prefer prebuilt binaries, use the release asset for your platform. `make build` writes a host-specific binary into `dist/`.

### Desktop

Install frontend dependencies and start the desktop app in dev mode:

```bash
make frontend-install
make dev-gui
```

Build the current-host desktop app:

```bash
make build-gui
```

Do not run `go build` directly inside `desktop/`. Wails desktop binaries must be produced with `make build-gui`, `make package-gui`, or `wails build`.

The desktop app can scan a live cluster or open an existing output bundle without cluster access.

## CLI And Desktop At A Glance

| Surface | Best for | Strengths |
| --- | --- | --- |
| CLI (`cmd/scan`) | CI/CD, repeatable ops workflows, scripting | Stable flags, schema-validated bundles, easy automation, policy gating with `cmd/check` |
| Desktop (`desktop/`) | Guided scans, bundle review, compare/history exploration | Shared backend, preflight assistant, live progress, export controls, accessible tabbed workspace |

## Output Artifacts

| Artifact | Purpose |
| --- | --- |
| `recovery-scan.json` | Primary machine-readable DR bundle |
| `recovery-enriched.json` | Enriched bundle used for history, compare, and follow-on tooling |
| `recovery-report.html` | Offline tabbed HTML report |
| `recovery-report.md` | Markdown export of the report |
| `recovery-summary.html` | Optional executive summary |
| `recovery-runbook.html` | Optional customer-facing DR runbook |
| `csv/` | Optional CSV exports for spreadsheet or downstream analysis |
| `*-redacted.*` | Optional share-safe exports with masked identifiers |

## Who This Tool Is For

- Platform and SRE teams who need a repeatable DR readiness baseline
- Consultants and MSPs who need offline deliverables for customer reviews
- Security, audit, and resilience owners who want evidence-backed reporting instead of optimistic backup claims
- CI/CD owners who want release gates based on real recovery posture signals

## Documentation

- Start here: [docs/README.md](docs/README.md)
- CLI usage: [docs/CLI.md](docs/CLI.md)
- Desktop app: [docs/GUI.md](docs/GUI.md)
- Architecture: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- Development: [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
- Troubleshooting: [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)
- RBAC guidance: [docs/RBAC.md](docs/RBAC.md)
- Schemas and compatibility: [docs/SCHEMAS.md](docs/SCHEMAS.md)
- Screenshots: [docs/SCREENSHOTS.md](docs/SCREENSHOTS.md)
- Release process: [docs/RELEASE.md](docs/RELEASE.md)

## Trust And Compatibility

- The supported CLI implementation remains [`cmd/scan`](cmd/scan).
- The desktop app uses the same shared backend in [`internal/appcore`](internal/appcore) and does not shell out to the CLI for normal runs.
- Published schema versions remain [`3.0.0`](schemas/recovery-scan-3.0.0.schema.json) for `recovery-scan.json` and [`1.1.0`](schemas/recovery-enriched-1.1.0.schema.json) for `recovery-enriched.json`.
- Generated HTML outputs stay self-contained and offline-friendly.
- Backup detection is not treated as proof of recoverability. Unsupported or permission-limited tooling is surfaced explicitly in both JSON and HTML outputs.

## Community

- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Support: [SUPPORT.md](SUPPORT.md)
- Security: [SECURITY.md](SECURITY.md)
- Code of Conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- License: [LICENSE](LICENSE)
