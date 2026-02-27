# k8s-recovery-visualizer

Kubernetes Disaster Recovery scoring and readiness analysis tool.

## Outputs
Generated under .\out:
- recovery-scan.json
- recovery-enriched.json
- recovery-report.md
- recovery-report.html
- history\index.json

## Quick Start
1) Run scan (required):

    .\scan.exe -out (Resolve-Path .\out).Path

2) Run report post-processing (trend + normalize + dark theme):

    $repoRoot = (Resolve-Path ".").Path
    $outDir   = (Resolve-Path ".\out").Path

    pwsh -NoProfile -ExecutionPolicy Bypass -Command @"
    . `"$repoRoot\scripts\report\Bootstrap-ReportLib.ps1`"
    & `"$repoRoot\scripts\report\Append-Trend-To-Reports.ps1`" -OutDir `"$outDir`" -Window 10
    & `"$repoRoot\scripts\report\Normalize-ReportHtml.ps1`" -OutDir `"$outDir`"
    & `"$repoRoot\scripts\report\Apply-DarkTheme.ps1`" -OutDir `"$outDir`"
    "@

3) Open report:

    Start-Process .\out\recovery-report.html

## Important
The report pipeline does NOT scan Kubernetes.
If the report looks stale, check timestamps:

    Get-ChildItem .\out\recovery-scan.json, .\out\recovery-enriched.json, .\out\recovery-report.html |
      Select-Object Name, LastWriteTime, Length | Format-Table -AutoSize
