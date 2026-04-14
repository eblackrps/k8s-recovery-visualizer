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

## Direct API endpoint mode is failing

If `API endpoint` mode in the desktop app is unclear or preflight keeps failing, check these in order:

1. Confirm you are using the Kubernetes API server URL, not an ingress or application URL.
2. Use the in-app `Test connection` step first. It checks reachability, auth, and TLS before full preflight adds RBAC and collection-readiness checks.
3. If `kubectl` already works on that machine, start with `Use existing access` instead of API endpoint mode.
4. If `kubectl` already works on that machine, print the active server directly:

```bash
kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'
```

5. Prefer a short-lived service-account token instead of a copied long-lived credential:

```bash
kubectl create token <service-account> --namespace <namespace>
```

6. If the cluster uses exec plugins, cloud auth helpers, SSO prompts, or client certificates, switch back to kubeconfig mode instead of forcing direct endpoint mode.
7. If the API server certificate is signed by a private or internal CA, add that CA to the desktop form instead of leaving trust on system defaults.
8. Use skip-TLS only as a temporary workaround in a trusted environment. The desktop review card and preflight panel intentionally flag this as a risk.

## Desktop kubeconfig file mode seems to reject the file or starts talking about an endpoint

Two things are easy to misread here:

1. A valid kubeconfig does not need a `.config` extension. Common names are just `config`, `kubeconfig`, `.yaml`, `.yml`, `.backup`, or no extension at all. The desktop app now validates by content, not filename.
2. Even in `Kubeconfig file` mode, the app still reads the cluster server address from inside that kubeconfig and uses it to connect. So a later error that mentions the API server or endpoint does not mean the app switched into direct `API endpoint` mode.

What to check:

1. Confirm the file is a real kubeconfig with `clusters`, `contexts`, and `users` entries.
2. Print the server address from the file:

```bash
kubectl config view --kubeconfig /path/to/config --minify -o jsonpath='{.clusters[0].cluster.server}'
```

```powershell
kubectl config view --kubeconfig C:\path\to\config --minify -o jsonpath='{.clusters[0].cluster.server}'
```

3. If the kubeconfig depends on an exec plugin, cloud auth helper, browser login, or client certificate flow, prefer `Current login` or another kubeconfig that already works with `kubectl` on that machine.
4. If `kubectl --kubeconfig <path> cluster-info` fails, the desktop scan will fail for the same underlying reason.
5. Use the desktop `Test connection` step before preflight when you want the fastest answer to whether the kubeconfig works on that machine at all.
6. If the desktop inspector says the kubeconfig is valid but still lists missing local CA or client-certificate files, the YAML copied successfully but the supporting files did not. Bring those files with the kubeconfig or export a self-contained kubeconfig with embedded `*-data` fields.
7. A copied kubeconfig often fails with `x509` or client-certificate errors when the original file referenced `certificate-authority`, `client-certificate`, or `client-key` paths on another machine. The desktop inspector now surfaces those path-based dependencies before the scan runs.
8. If the file picker itself is getting in the way, drag the kubeconfig onto the desktop scan dropzone. K8V will load it into paste mode and validate the contents without caring about the filename extension.
9. If the Home page or Step 1 shows an existing-access caution, trust that warning. It means K8V found local Kubernetes configuration, but the detected kubeconfig still depends on missing files or an external auth helper that may not work from this desktop session.

## What is the difference between Test connection, Preflight, and Start scan?

- `Test connection` answers whether the cluster is reachable with the current transport, credentials, and TLS settings.
- `Preflight` answers whether the final scope, RBAC, and collectors are ready for the real scan.
- `Start scan` collects evidence and writes the bundle/report artifacts into the chosen output directory.

If a first attempt fails, use that order. It is usually faster and clearer than changing several things at once.

## Desktop failure labels and what they mean

Recent desktop builds classify common failures before showing the raw error text. These are the main labels to look for:

- `Endpoint unreachable`: the API server URL is wrong for this machine, DNS cannot resolve it, or the control plane is not reachable from the current network path.
- `TLS trust`: the cluster is reachable, but the API server certificate is not trusted on this machine. Add the issuing CA or confirm whether the kubeconfig references a CA file that was not copied over.
- `External auth helper`: the kubeconfig depends on an exec plugin, SSO flow, cloud login helper, or another credential source that is not usable from this desktop session.
- `Auth rejected`: the cluster answered, but the provided token or credentials were not accepted.
- `RBAC denied`: transport and auth succeeded, but the current identity cannot read the resources the scan or preflight needs.
- `Output path`: the app connected successfully, but it cannot create or write the requested output files on disk.

The raw error detail still matters, but the label should tell you where to look first.

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
go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.1.0.schema.json -input ./out/recovery-scan.json
go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.2.0.schema.json -input ./out/recovery-enriched.json
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

## Desktop app does nothing on launch (Windows)

If `K8V.exe` does nothing when double-clicked:

1. Confirm you extracted the zip and are not running from inside the archive.
2. If the app was built locally on Windows, rebuild it with the current `make build-gui` or `make package-gui` target. Older local builds could still use the newer WebView2 loader and fail on some managed Windows 11 machines even when they worked on the build box.
3. Verify WebView2 is installed by checking for:
   - `C:\Program Files (x86)\Microsoft\EdgeWebView\Application\msedgewebview2.exe`
   - `C:\Program Files\Microsoft\EdgeWebView\Application\msedgewebview2.exe`
4. Open the startup log created by the app:
   - `%APPDATA%\k8s-recovery-visualizer\logs\k8v-startup.log`

If the log file is missing or shows WebView2 not found, install the Evergreen WebView2 runtime and retry. If the log shows a startup error, attach the log when reporting the issue.

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
