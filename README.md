# k8s-recovery-visualizer

[![CI](https://github.com/eblackrps/k8s-recovery-visualizer/actions/workflows/ci.yml/badge.svg)](https://github.com/eblackrps/k8s-recovery-visualizer/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/eblackrps/k8s-recovery-visualizer)](https://github.com/eblackrps/k8s-recovery-visualizer/releases)
[![License](https://img.shields.io/github/license/eblackrps/k8s-recovery-visualizer)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/eblackrps/k8s-recovery-visualizer)](go.mod)

`k8s-recovery-visualizer` is the repository, release, and archive identity for a Kubernetes disaster recovery assessment toolkit. The desktop product is `K8V`: a Wails desktop workspace for guided scans, live preflight and run feedback, bundle review, history, compare workflows, and offline exports. The Go CLI stays in-repo for contributors, CI gates, smoke tests, automation, and source builds.

<p align="center">
  <img alt="K8V desktop dashboard" src="images/gui-dashboard.png" width="49%" />
  <img alt="K8V findings workspace" src="images/gui-results-findings.png" width="49%" />
</p>

`K8V` can scan a live cluster or open an existing bundle directory, `recovery-scan.json`, `.zip`, `.tar.gz`, or `.tgz` bundle without cluster access. Bundle loading now validates archives and JSON structure up front so operators get clearer corruption or mis-packaging diagnostics instead of a generic open failure.

## What Teams Get

- prioritized findings with impact, likely owner, rough effort, and deterministic ranking
- restore-readiness evidence that goes beyond “backup detected” to show blocked, warning, ready, and unknown namespaces
- a restore drill planner that turns bundle evidence into an operator runbook sequence
- compare and history workflows that surface score drift, severity deltas, regressed findings, and persistent gaps
- offline-friendly exports and schema-validated bundles that still work in CI and air-gapped review flows

## Public Release

Public GitHub releases publish exactly four files:

- `k8s-recovery-visualizer-desktop-linux-amd64.tar.gz`
- `k8s-recovery-visualizer-desktop-windows-amd64.zip`
- `checksums.txt`
- `k8s-recovery-visualizer.spdx.json`

Current supported public release platforms:

| Platform | Status | Notes |
| --- | --- | --- |
| Linux desktop amd64 | Supported | Public tarball release artifact |
| Windows desktop amd64 | Supported | Public zip release artifact |

Deprecated release surfaces and contributor-only build paths are documented in [docs/SUPPORT-MATRIX.md](docs/SUPPORT-MATRIX.md).

## Choose Your Path

| Path | Best for | How |
| --- | --- | --- |
| Download the desktop app | Operators and evaluators | Grab the Linux or Windows desktop package from [GitHub Releases](https://github.com/eblackrps/k8s-recovery-visualizer/releases/latest) |
| Build the CLI from source | Contributors, CI, air-gapped workflows | `make build` |
| Build cross-platform CLI binaries locally | Contributor validation and internal packaging | `make build-cli-cross` |
| Run the desktop app in dev mode | Frontend and UX iteration | `make frontend-install && make dev-gui` |
| Build the current-host desktop app | Local packaging validation | `make frontend-install && make build-gui` |

## Quickstart

### Desktop

Install frontend dependencies and launch the desktop app in development mode:

```bash
make frontend-install
make dev-gui
```

Build the current-host desktop app:

```bash
make build-gui
```

### CLI Source Builds

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
go run ./cmd/check --current ./out/recovery-scan.json --min-score 85 --min-backup-score 80 --max-new-findings 0 --max-regressed-findings 0 --format json
```

Build a host-specific CLI binary into `dist/`:

```bash
make build
```

## Latest Desktop Screenshots

<p align="center">
  <img alt="K8V guided scan wizard" src="images/gui-scan-wizard.png" width="32%" />
  <img alt="K8V live run progress" src="images/gui-live-run.png" width="32%" />
  <img alt="K8V compare workflow" src="images/gui-compare.png" width="32%" />
</p>

The public gallery intentionally uses the current deterministic desktop screenshot set only. See [docs/SCREENSHOTS.md](docs/SCREENSHOTS.md) for the capture workflow and maintained image list.

## CLI And Desktop At A Glance

| Surface | Best for | Strengths |
| --- | --- | --- |
| CLI (`cmd/scan`, `cmd/check`) | CI/CD, repeatable ops workflows, scripting, source builds | Stable flags, schema-validated bundles, policy gating, deterministic smoke flows |
| Desktop (`desktop/`) | Guided scans, bundle review, compare/history exploration | Shared backend, preflight assistant, live progress, cancellation, export controls, prioritized findings, restore drill planning, offline bundle inspection |

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

## Documentation

- Start here: [docs/README.md](docs/README.md)
- CLI usage and source builds: [docs/CLI.md](docs/CLI.md)
- Desktop app guide: [docs/GUI.md](docs/GUI.md)
- Architecture: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- Development: [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
- Release process: [docs/RELEASE.md](docs/RELEASE.md)
- Support policy: [docs/SUPPORT-MATRIX.md](docs/SUPPORT-MATRIX.md)
- Troubleshooting: [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)
- Schemas and compatibility: [docs/SCHEMAS.md](docs/SCHEMAS.md)
- Screenshots: [docs/SCREENSHOTS.md](docs/SCREENSHOTS.md)

## Community

- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Support: [SUPPORT.md](SUPPORT.md)
- Security: [SECURITY.md](SECURITY.md)
- Code of Conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- License: [LICENSE](LICENSE)
