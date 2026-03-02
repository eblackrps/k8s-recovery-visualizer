# Troubleshooting

## TLS Certificate Verification Errors

If the scan fails with an error like:

```
x509: certificate signed by unknown authority
```

This is common on RKE2, k3s, and bare-metal clusters that use self-signed CA certificates. Use the `--insecure` flag to skip TLS verification:

```bash
# Linux
./dist/scan-linux-amd64 --insecure --kubeconfig /etc/rancher/rke2/rke2.yaml --out ./out

# Windows
.\dist\scan.exe --insecure --kubeconfig C:\path\to\rke2.yaml --out .\out
```

A warning is printed when `--insecure` is active:

```
WARNING: --insecure is set — TLS certificate verification is disabled.
```

This warning is suppressed in `--ci` mode.

> Only use `--insecure` on clusters you trust. It disables verification of the server's certificate chain.

---

## Cluster access check
    kubectl config current-context
    kubectl cluster-info
    kubectl get nodes -o wide
    kubectl get storageclass -o wide
    kubectl auth can-i list storageclass

## Report shows old data
You likely re-ran the report pipeline without running the scan.
Verify timestamps:

```bash
# Linux / macOS
ls -lh out/recovery-scan.json out/recovery-enriched.json out/recovery-report.html
```

```powershell
# Windows
Get-ChildItem .\out\recovery-scan.json, .\out\recovery-enriched.json, .\out\recovery-report.html |
  Select-Object Name, LastWriteTime, Length | Format-Table -AutoSize
```

If `recovery-report.html` is newer than `recovery-scan.json`, rerun the scan:

```bash
# Linux / macOS
./scan --out ./out

# Windows
.\scan.exe --out .\out
```
