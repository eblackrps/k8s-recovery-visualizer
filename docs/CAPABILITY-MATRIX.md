# Capability Matrix

## Backup systems

| Tool | Detected | Policy inspection | Recent success evidence | Offsite evidence | Coverage confidence |
| --- | --- | --- | --- | --- | --- |
| Velero | Yes | Yes | Yes (`status.lastBackup`) | Yes (`storageLocation`) | Confirmed when schedules parse |
| Kasten K10 | Yes | Yes | Partial / inferred | Yes (`export` action) | Confirmed for policy scope, inferred for run freshness |
| Longhorn | Yes | Yes | Partial / inferred | Yes (`backup-target` setting) | Confirmed for recurring-job scope, inferred for run freshness |
| Rubrik | Yes | Detection only | No | No | Unverified |
| Trilio | Yes | Detection only | No | No | Unverified |
| Stash | Yes | Detection only | No | No | Unverified |
| CloudCasa | Yes | Detection only | No | No | Unverified |

## Collector and output behavior

| Capability | Status | Notes |
| --- | --- | --- |
| Weighted DR scoring | Supported | Now backed by data-driven rule packs and golden scenario tests |
| HTML report | Supported | Self-contained offline artifact |
| JSON scan bundle | Supported | Versioned schema contract under `schemas/` |
| Enriched trend artifact | Supported | Derived from `recovery-scan.json` plus history |
| Redacted artifacts | Supported | Useful for sharing, but still review for contextual leakage |
| Compare / trend gates | Supported | Use `cmd/check` against current and previous bundles |
| Namespace-scoped scan | Supported | Requires `--namespace` plus namespaced RBAC manifest |

## Degraded modes

| Condition | What happens |
| --- | --- |
| Backup tool detected but not inspectable | `coverageStatus` becomes `unsupported` and assurance becomes `unverified` |
| Backup API forbidden | `coverageStatus` becomes `permission_denied`; reports call this out explicitly |
| Collector RBAC failures | Collector skip is recorded in JSON and reports |
| No recent backup run evidence | Assurance downgrades to `evidence_inferred` or `coverage_gap` depending on the rest of the evidence |
| Offsite evidence exists for only part of verified coverage | `hasOffsite=false`, missing namespaces are listed, and `BACKUP_NO_OFFSITE` stays active |
