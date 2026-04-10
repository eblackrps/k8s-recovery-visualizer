# Architecture

`k8s-recovery-visualizer` runs in a single CLI process, but the pipeline has distinct stages with different trust boundaries:

1. Collect
   Reads Kubernetes resources into `internal/model.Bundle`.
   Required collectors fail the scan.
   Optional collectors are recorded in `collectorSkips` when RBAC or API access is missing.

2. Detect backup tooling
   Detects installed backup products from namespaces, CRDs, and pods.
   Supported tools are inspected for policies and schedules.
   Detection-only tools remain explicitly unverified.

3. Simulate restore
   Builds namespace-level restore assessments from verified backup coverage, PVC volume, and restore blockers.
   If backup coverage is not verified, restore coverage is reported as unknown instead of guessed.

4. Score and remediate
   Applies weighted DR scoring in `internal/analyze`.
   Generates prioritized remediation steps in `internal/remediation`.

5. Write artifacts
   `recovery-scan.json` is the source-of-truth scan bundle.
   `recovery-enriched.json` is derived trend/risk metadata.
   `recovery-report.html`, `recovery-report.md`, `recovery-summary.html`, and `recovery-runbook.html` are presentation outputs.

6. Record history
   Writes historical scan metadata under `out/history/`.
   The final rendered HTML report is snapshotted into history after the tabbed report is generated.

## Key contracts

- `recovery-scan.json` schema: [`../schemas/recovery-scan-2.1.0.schema.json`](../schemas/recovery-scan-2.1.0.schema.json)
- `recovery-enriched.json` schema: [`../schemas/recovery-enriched-1.1.0.schema.json`](../schemas/recovery-enriched-1.1.0.schema.json)

## Design choices

- Conservative backup claims: detection does not equal verified coverage.
- Offline-friendly reports: HTML output is self-contained and does not need a CDN.
- Deterministic scoring: the same bundle produces the same score.
- Additive JSON changes require schema version updates and CI validation.
