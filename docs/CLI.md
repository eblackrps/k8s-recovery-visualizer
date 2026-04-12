# CLI Guide

The supported command-line entrypoint is [`cmd/scan`](../cmd/scan). It stays intentionally thin and delegates real execution to [`internal/scanapp`](../internal/scanapp) and the shared backend in [`internal/appcore`](../internal/appcore).

## When To Use The CLI

Use the CLI when you want:

- repeatable automation in CI/CD or scheduled jobs
- air-gapped or terminal-only operation
- direct control over output locations and scan options
- `cmd/check` policy gates against generated bundles

## Build Or Download

- Download a prebuilt binary from [GitHub Releases](https://github.com/eblackrps/k8s-recovery-visualizer/releases/latest)
- Or build a host-specific binary locally with:

```bash
make build
```

`make build` writes a host-specific binary into `dist/`. `make release-cli` builds the full cross-platform CLI set used for releases.

## Common Examples

Deterministic dry run:

```bash
go run ./cmd/scan --dry-run --summary --runbook --out ./out --min-score 0
```

Cluster-wide scan:

```bash
go run ./cmd/scan --context prod-east-admin --profile enterprise --summary --runbook --out ./out
```

Namespace-scoped scan:

```bash
go run ./cmd/scan --namespace payments,frontend --context prod-east-admin --out ./out
```

Compare against a previous bundle:

```bash
go run ./cmd/scan --compare ./previous/recovery-scan.json --out ./out
```

Generate share-safe artifacts:

```bash
go run ./cmd/scan --redact --csv --summary --runbook --out ./out
```

Use the CI gate engine:

```bash
go run ./cmd/check --current ./out/recovery-scan.json --min-score 85 --format json
go run ./cmd/check --current ./out/recovery-scan.json --previous ./previous/recovery-scan.json --max-drop 5 --fail-on-new-critical
```

## Output Artifacts

The CLI can write:

- `recovery-scan.json`
- `recovery-enriched.json`
- `recovery-report.html`
- `recovery-report.md`
- optional summary, runbook, CSV, and redacted artifacts

All HTML outputs remain self-contained and offline-friendly.

## High-Value Flags

| Flag | Purpose |
| --- | --- |
| `--dry-run` | Run against deterministic fixtures instead of a live cluster |
| `--context` / `--kubeconfig` | Select the kubeconfig and context to use |
| `--namespace` | Limit the scan to one or more namespaces |
| `--profile` | Apply `standard`, `enterprise`, `dev`, or `airgap` scoring weights |
| `--compare` | Diff the current scan against a previous bundle |
| `--summary` / `--runbook` | Generate additional HTML deliverables |
| `--redact` | Produce masked JSON and HTML outputs for safer sharing |
| `--csv` | Emit CSV exports for spreadsheet or downstream analysis |
| `--include-secret-metadata` | Opt in to Secret metadata collection when you have approved RBAC for it |
| `--insecure` | Skip TLS verification for clusters with a known self-signed or broken trust chain |

## Compatibility Notes

- CLI flags remain backward compatible in the current release line.
- Published JSON schemas remain the contract boundary for downstream tooling.
- Namespace-scoped scans are supported, but intentionally degrade some cluster-wide findings and backup evidence.

For schemas, see [SCHEMAS.md](SCHEMAS.md). For RBAC guidance, see [RBAC.md](RBAC.md).
