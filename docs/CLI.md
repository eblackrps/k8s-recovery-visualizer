# CLI Guide

The supported command-line entrypoints are [`cmd/scan`](../cmd/scan) and [`cmd/check`](../cmd/check). They stay intentionally thin and delegate real execution to [`internal/scanapp`](../internal/scanapp) and the shared backend in [`internal/appcore`](../internal/appcore).

## Public Release Note

Prebuilt CLI binaries are no longer published as GitHub release assets. The CLI remains fully supported for source builds, CI gates, smoke tests, automation, and contributor workflows.

## Build Or Run From Source

Build a host-specific CLI binary:

```bash
make build
```

Build the contributor cross-platform CLI set into `dist/`:

```bash
make build-cli-cross
```

Or run directly from source without producing a release-style binary first:

```bash
go run ./cmd/scan --dry-run --summary --runbook --out ./out --min-score 0
```

## When To Use The CLI

Use the CLI when you want:

- repeatable automation in CI/CD or scheduled jobs
- air-gapped or terminal-only operation
- direct control over output locations and scan options
- `cmd/check` policy gates against generated bundles
- source-first validation without relying on GitHub release binaries

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
- Public GitHub releases no longer advertise or ship CLI binaries, so operational docs should point users to source builds instead.

For schemas, see [SCHEMAS.md](SCHEMAS.md). For RBAC guidance, see [RBAC.md](RBAC.md).
