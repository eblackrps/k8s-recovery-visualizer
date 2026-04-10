# Architecture

`k8s-recovery-visualizer` now keeps the CLI entrypoint thin by routing the scan workflow through `internal/scanapp`. The process is still a single binary, but the main responsibilities are split into explicit layers with clearer trust boundaries:

1. Parse options
   `internal/scanapp/options.go` owns flags and validation.

2. Bootstrap and collect
   `internal/scanapp/collectors.go` builds the collector pipeline.
   Reads Kubernetes resources into `internal/model.Bundle`.
   Required collectors fail the scan.
   Optional collectors are recorded in `collectorSkips` when RBAC or API access is missing.

3. Detect backup tooling
   Detects installed backup products from namespaces, CRDs, and pods.
   Supported tools are inspected for policies and schedules.
   Detection-only tools remain explicitly unverified.

4. Simulate restore and assurance
   Builds namespace-level restore assessments from verified backup coverage, PVC volume, and restore blockers.
   If backup coverage is not verified, restore coverage is reported as unknown instead of guessed.
   `internal/backup/assurance.go` turns policy inspection, recent-success evidence, offsite state, snapshot readiness, and restore blockers into a conservative assurance conclusion.

5. Score and remediate
   Applies weighted DR scoring in `internal/analyze`.
   Rule weights, penalties, and built-in profiles are loaded from `internal/scoring/config/`.
   Generates prioritized remediation steps in `internal/remediation`.

6. Write artifacts
   `internal/scanapp/artifacts.go` owns JSON, HTML, summary, runbook, redaction, enrichment, and history writing.
   `recovery-scan.json` is the source-of-truth scan bundle.
   `recovery-enriched.json` is derived trend/risk metadata.
   `recovery-report.html`, `recovery-report.md`, `recovery-summary.html`, and `recovery-runbook.html` are presentation outputs.

7. Record history and policy gates
   Writes historical scan metadata under `out/history/`.
   The final rendered HTML report is snapshotted into history after the tabbed report is generated.
   `cmd/check` evaluates deterministic CI gates against current and previous scan bundles.

## Key contracts

- `recovery-scan.json` schema: [`../schemas/recovery-scan-2.2.0.schema.json`](../schemas/recovery-scan-2.2.0.schema.json)
- `recovery-enriched.json` schema: [`../schemas/recovery-enriched-1.1.0.schema.json`](../schemas/recovery-enriched-1.1.0.schema.json)

## Design choices

- Conservative backup claims: detection does not equal verified coverage.
- Evidence-backed remediation: finding output now carries operator guidance beyond a one-line recommendation.
- Offline-friendly reports: HTML output is self-contained and does not need a CDN.
- Deterministic scoring: the same bundle produces the same score.
- Additive JSON changes require schema version updates and CI validation.
