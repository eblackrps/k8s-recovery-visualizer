# Troubleshooting

## Cluster access check
    kubectl config current-context
    kubectl cluster-info
    kubectl get nodes -o wide
    kubectl get storageclass -o wide
    kubectl auth can-i list storageclass

## Report shows old data
You likely re-ran the report pipeline without running the scan.
Verify timestamps:

    Get-ChildItem .\out\recovery-scan.json, .\out\recovery-enriched.json, .\out\recovery-report.html |
      Select-Object Name, LastWriteTime, Length | Format-Table -AutoSize

If recovery-report.html is newer than recovery-scan.json, rerun:
    .\scan.exe -out (Resolve-Path .\out).Path
