# Troubleshooting

## TLS verification errors

If the scan fails with:

```text
x509: certificate signed by unknown authority
```

use `--insecure` only for clusters you trust:

```bash
go run ./cmd/scan --insecure --kubeconfig /path/to/config --out ./out
```

```powershell
go run ./cmd/scan --insecure --kubeconfig C:\path\to\config --out .\out
```

## Backup coverage shows `permission denied`

The scanner knows how to inspect the backup tool, but the current credentials cannot read the policy objects.

Examples:

```bash
kubectl auth can-i list schedules.velero.io --all-namespaces
kubectl auth can-i list policies.config.kio.kasten.io --all-namespaces
kubectl auth can-i list recurringjobs.longhorn.io -n longhorn-system
```

If these checks fail, the report will conservatively mark coverage as unknown.

## Backup coverage shows `unsupported`

The tool was detected, but native policy inspection has not been implemented for that product yet. The report is intentionally conservative in this case.

What to do:

1. Verify namespace scope and schedule coverage in the backup product directly.
2. Treat the score as conservative until support is added.
3. If you are extending the scanner, add native policy collection for that product and update the schema/docs together.

## Schema validation

Validate emitted JSON against the published contracts:

```bash
go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.0.0.schema.json -input ./out/recovery-scan.json
go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.1.0.schema.json -input ./out/recovery-enriched.json
```

## Secret metadata was not collected

This is expected unless you explicitly enable it.

- The default RBAC manifests do not grant Secret reads.
- `inventory.secrets` is only populated when you run with `--include-secret-metadata`.
- If you enable that flag, you must extend RBAC yourself and accept that Kubernetes Secret reads expose full Secret objects to the scanner.

## Namespace-scoped scan skips collectors

This is expected when the service account only has namespace-local permissions.

Check:

- `collectorSkips` in `recovery-scan.json`
- the "Scan Coverage" card in `recovery-report.html`

If you want node, PV, or StorageClass context in namespace mode, use the shared `ClusterRole` from [`RBAC.md`](RBAC.md).

## Report looks stale

Check artifact timestamps:

```bash
ls -lh out/recovery-scan.json out/recovery-enriched.json out/recovery-report.html
```

```powershell
Get-ChildItem .\out\recovery-scan.json, .\out\recovery-enriched.json, .\out\recovery-report.html |
  Select-Object Name, LastWriteTime, Length | Format-Table -AutoSize
```

If `recovery-report.html` is older than the latest scan, rerun the scan.

## Desktop package build says `makensis not found`

On Windows, `make package-gui` requires NSIS for the Wails installer step.

What to do:

1. Install NSIS locally.
2. Re-run `make package-gui`.
3. If you only need a local app binary, use `make build-gui` instead.

The GitHub release workflow installs NSIS automatically on the Windows packaging runner.

## Desktop app opens without recent projects

The desktop home screen discovers bundles under the configured workspace root.

Check:

- the workspace root in Settings
- that the folder contains `recovery-scan.json` or a generated report bundle
- that the scan completed successfully and wrote outputs to disk

## Quick cluster access sanity checks

```bash
kubectl config current-context
kubectl cluster-info
kubectl get nodes -o wide
kubectl get storageclass -o wide
```
