# DR Scan Prereqs

## Required
- PowerShell 7 (pwsh)
- kubectl in PATH
- kubeconfig configured to the target cluster
- Read-only access to core resources across namespaces (scan can still run partially with limited RBAC)

## Run
Open PowerShell 7 in this folder and run:

pwsh -ExecutionPolicy Bypass -File .\Run-DrScan.ps1 -OutDir .\out

Outputs:
- .\out\drscan-report.json
- .\out\drscan-report.md
- .\out\drscan-report.html
