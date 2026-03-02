# k8s-recovery-visualizer

## Quick sanity run

    powershell -NoProfile -ExecutionPolicy Bypass -File .\tools\smoke.ps1

## Scan examples

Text report + HTML:

    .\scan.exe --dry-run --out .\out

CI JSON mode:

    .\scan.exe --dry-run --ci --out .\out

Policy gate example (exit code 2 when score < min-score):

    .\scan.exe --dry-run --ci --min-score 90 --out .\out

Self-signed / RKE2 / k3s clusters (skip TLS cert verification):

    .\scan.exe --insecure --out .\out

With explicit kubeconfig:

    .\scan.exe --insecure --kubeconfig C:\path\to\rke2.yaml --out .\out
