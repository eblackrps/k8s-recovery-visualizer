# Architecture

Scan -> Enrich -> Score -> Report -> Trend -> Normalize -> Theme -> Screenshot (optional)

Key point:
The report post-processing pipeline does NOT scan Kubernetes.
Always run scan.exe first if you want fresh cluster data.
