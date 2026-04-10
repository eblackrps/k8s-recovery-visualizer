# k8s-recovery-visualizer

`k8s-recovery-visualizer` scans a Kubernetes cluster, inventories recovery-relevant resources, scores disaster recovery readiness across weighted domains, and produces offline-friendly reports.

The tool is intentionally conservative. It will not claim backup coverage unless it can actually inspect backup policies for the detected product. Detection-only tools are reported as `unsupported` coverage, not silently treated as protected.

## What it does

- Inventories cluster, workload, storage, network, config, image, Helm, certificate, and backup-related resources.
- Scores DR readiness across `storage`, `workload`, `config`, and `backup / recovery`.
- Detects backup products and inspects policies for supported tools.
- Simulates restore feasibility per namespace using policy coverage, PVC volume, and restore blockers.
- Generates `JSON`, `Markdown`, `HTML`, optional CSV exports, an executive summary, and a customer-facing runbook.
- Maintains scan history for trend reporting.

## Backup trust model

Backup coverage is reported with an explicit status:

| Status | Meaning |
| --- | --- |
| `verified` | The scanner successfully inspected policies or schedules and can reason about namespace coverage. |
| `unsupported` | The product was detected, but the scanner does not yet inspect that tool's policies. |
| `permission_denied` | The scanner knows how to inspect the tool, but the current credentials could not read the policy objects. |
| `api_error` | Policy inspection failed because the backup product API call failed. |
| `parse_error` | Policy inspection reached the API but could not parse the returned objects. |
| `not_detected` | No backup tool was detected. |

When coverage is not `verified`, the score treats backup scope as unknown and avoids pretending that restore coverage is known.

## Supported backup inspection

| Tool | Detection | Policy inspection |
| --- | --- | --- |
| Kasten K10 | Yes | Yes |
| Velero | Yes | Yes |
| Longhorn | Yes | Yes |
| Rubrik | Yes | Detection only |
| Trilio | Yes | Detection only |
| Stash | Yes | Detection only |
| CloudCasa | Yes | Detection only |

## Output files

Running `scan` writes:

```text
out/
├── recovery-scan.json
├── recovery-enriched.json
├── recovery-report.html
├── recovery-report.md
├── recovery-summary.html          # only with --summary
├── recovery-runbook.html          # only with --runbook
├── recovery-scan-redacted.json    # only with --redact
├── recovery-report-redacted.html  # only with --redact
├── csv/                           # only with --csv
└── history/
    └── index.json
```

Published JSON contracts:

- Scan bundle schema: [`schemas/recovery-scan-2.1.0.schema.json`](schemas/recovery-scan-2.1.0.schema.json)
- Enriched bundle schema: [`schemas/recovery-enriched-1.1.0.schema.json`](schemas/recovery-enriched-1.1.0.schema.json)

Compatibility policy:

- Additive fields bump the schema minor version.
- Breaking JSON changes require a new major schema version.
- CI validates the generated artifacts against the published schemas.

## Quick start

### Prerequisites

- Go `1.25+`, or a release binary from [GitHub Releases](https://github.com/eblackrps/k8s-recovery-visualizer/releases)
- A kubeconfig with read access to the target cluster

### Build

Build for your current host:

```bash
make build
```

Build release binaries for all supported targets:

```bash
make release
```

Manual examples:

```bash
GOOS=linux GOARCH=amd64 go build -o dist/scan-linux-amd64 ./cmd/scan
GOOS=darwin GOARCH=arm64 go build -o dist/scan-darwin-arm64 ./cmd/scan
GOOS=windows GOARCH=amd64 go build -o dist/scan-windows-amd64.exe ./cmd/scan
```

### Run

Linux:

```bash
./dist/scan-linux-amd64 --out ./out
./dist/scan-linux-amd64 --namespace=prod,staging --out ./out
./dist/scan-linux-amd64 --target=baremetal --csv --out ./out
./dist/scan-linux-amd64 --profile=enterprise --out ./out
./dist/scan-linux-amd64 --compare=./previous/recovery-scan.json --out ./out
./dist/scan-linux-amd64 --summary --runbook --redact --out ./out
./dist/scan-linux-amd64 --dry-run --out ./out
```

Windows:

```powershell
.\dist\scan-windows-amd64.exe --out .\out
.\dist\scan-windows-amd64.exe --namespace=prod,staging --out .\out
.\dist\scan-windows-amd64.exe --insecure --out .\out
.\dist\scan-windows-amd64.exe --ci --min-score=75 --out .\out
```

### Open the report

```bash
xdg-open ./out/recovery-report.html
open ./out/recovery-report.html
start .\out\recovery-report.html
```

## CLI flags

| Flag | Default | Description |
| --- | --- | --- |
| `--kubeconfig` | `""` | Path to kubeconfig. Uses in-cluster config if empty. |
| `--insecure` | `false` | Skip TLS verification for self-signed clusters. |
| `--out` | `./out` | Output directory. |
| `--target` | `vm` | Recovery target: `vm` or `baremetal`. |
| `--profile` | `standard` | Scoring profile: `standard`, `enterprise`, `dev`, `airgap`. |
| `--namespace` | `""` | Comma-separated namespace scope. Empty means all namespaces. |
| `--compare` | `""` | Compare against a previous `recovery-scan.json`. |
| `--csv` | `false` | Write CSV exports. |
| `--summary` | `false` | Write `recovery-summary.html`. |
| `--runbook` | `false` | Write `recovery-runbook.html`. |
| `--redact` | `false` | Write redacted JSON and HTML artifacts. |
| `--dry-run` | `false` | Run without hitting a live cluster. |
| `--ci` | `false` | Emit a machine-readable summary and exit `2` if score is below threshold. |
| `--min-score` | `90` | Minimum acceptable score in CI mode. |
| `--timeout` | `60` | Kubernetes API timeout in seconds. |
| `--customer` | `""` | Customer identifier stored in metadata. |
| `--site` | `""` | Site or region stored in metadata. |
| `--cluster` | `""` | Cluster name stored in metadata. |
| `--env` | `""` | Environment tag stored in metadata. |

## Scoring

Weighted domains:

| Domain | Weight |
| --- | --- |
| Storage | 35% |
| Workload | 20% |
| Config | 15% |
| Backup / Recovery | 30% |

Profiles adjust penalty emphasis without changing the base domain weights:

| Profile | Multipliers |
| --- | --- |
| `standard` | baseline |
| `enterprise` | restore testing `1.5x`, immutability `1.3x`, replication `1.2x`, security `1.2x` |
| `dev` | restore testing `1.1x`, immutability `0.9x`, replication `0.9x` |
| `airgap` | immutability `1.6x`, airgap `1.6x`, security `1.3x`, restore testing `1.2x` |

The full scoring output is emitted in `recovery-scan.json` and rendered in the `DR Score` tab.

## CI and validation

Run the core checks locally:

```bash
go test ./...
go build ./...
go run ./cmd/schema-validate -schema ./schemas/recovery-scan-2.1.0.schema.json -input ./out/recovery-scan.json
go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.1.0.schema.json -input ./out/recovery-enriched.json
```

`cmd/check` can enforce enriched risk posture and score regression thresholds:

```bash
go run ./cmd/check --in ./out/recovery-enriched.json --max-risk MODERATE --max-drop 5
```

## Release process

- Tags matching `v*` trigger [`.github/workflows/release.yml`](.github/workflows/release.yml).
- Release builds inject version and build date metadata.
- Published assets include Linux, macOS, Windows binaries, and SHA-256 checksums.

## Documentation

- Architecture: [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- Troubleshooting: [`docs/TROUBLESHOOTING.md`](docs/TROUBLESHOOTING.md)
- Schemas: [`docs/SCHEMAS.md`](docs/SCHEMAS.md)

## Repository notes

- The supported implementation is the Go CLI in [`cmd/scan`](cmd/scan).
- The original PowerShell-based workflow has been archived under [`legacy/powershell`](legacy/powershell) for historical reference and is not part of the current CI, release, or support path.

## Limitations

- Backup coverage is only verified for the supported policy-inspection tools listed above.
- Detection-only tools remain useful inventory signals, but they do not prove protection scope.
- Redaction removes secret values and masks identifiers in redacted artifacts, but you should still review shared outputs for customer-specific context.
