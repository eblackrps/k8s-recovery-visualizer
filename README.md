# k8s-recovery-visualizer

Full Kubernetes cluster inventory and Disaster Recovery assessment tool.

Scan a live cluster to get a complete RVTools-style inventory, a weighted DR readiness score across four domains, backup tool detection, backup policy analysis, restore simulation, and a prioritized remediation plan — all in a single self-contained HTML report.

---

## Why This Exists

Building DR clusters without structured discovery leads to:
- Missing stateful workloads with no backup coverage
- Storage classes that don't exist in the target environment
- Orphaned PVs with Delete reclaim policies that vanish on PVC deletion
- Public registry images that aren't reachable in air-gapped DR sites
- Backup tools installed but no policies or schedules configured
- No offsite target — a site-level failure takes the backups with it

k8s-recovery-visualizer performs deterministic environment analysis and produces a full picture of DR readiness before a rebuild or recovery event.

---

## What It Collects

| Category | Resources |
|----------|-----------|
| **Cluster** | Nodes, namespaces, platform/provider, K8s version |
| **Workloads** | Deployments, DaemonSets, StatefulSets, Jobs, CronJobs |
| **Storage** | PVCs, PVs, StorageClasses |
| **Networking** | Services, Ingresses, NetworkPolicies |
| **Config** | ConfigMaps, Secrets (metadata only), ClusterRoles, CRDs, ResourceQuotas, HPAs, PodDisruptionBudgets |
| **Images** | All container images grouped by registry; public vs. private flag |
| **Helm** | All Helm v3 releases detected via K8s secrets (no Helm CLI required) |
| **Certificates** | cert-manager Certificate resources with expiry and days-to-renewal |
| **Backup** | Auto-detects 7 tools; collects Velero Schedules, Kasten K10 Policies, Longhorn RecurringJobs |

---

## DR Scoring Model

Scoring covers four weighted domains:

| Domain | Weight | What It Measures |
|--------|--------|-----------------|
| **Storage** | 35% | PVC binding, storageClass presence, hostPath usage, reclaim policies |
| **Workload** | 20% | StatefulSet persistence, deployment coverage |
| **Config** | 15% | CRD backup readiness, certificate expiry, image registry risk |
| **Backup/Recovery** | 30% | Tool presence, policy coverage, offsite config, RPO, restore simulation |

**Maturity levels:** PLATINUM (≥90) · GOLD (≥75) · SILVER (≥50) · BRONZE (<50)

### Backup/Recovery Scoring Rules

| Finding ID | Severity | Penalty | Condition |
|---|---|---|---|
| `BACKUP_NONE` | CRITICAL | −60 | No backup tool detected |
| `BACKUP_NO_POLICIES` | HIGH | −30 | Tool detected but no schedules/policies found |
| `BACKUP_PARTIAL_COVERAGE` | HIGH | −20 | StatefulSets in namespaces not covered by any policy |
| `BACKUP_NO_OFFSITE` | HIGH | −15 | No offsite or export location configured |
| `BACKUP_RPO_HIGH` | MEDIUM | −10 | Worst-case RPO across all policies exceeds 24 hours |
| `RESTORE_SIM_UNCOVERED` | HIGH | −20 | Stateful namespaces have no matching backup policy |

### Example Scoring Breakdown

| Domain | Score | Weight | Weighted |
|--------|-------|--------|---------|
| Storage | 85/100 | 35% | 29.75 |
| Workload | 100/100 | 20% | 20.00 |
| Config | 90/100 | 15% | 13.50 |
| Backup/Recovery | 40/100 | 30% | 12.00 |
| **Overall** | **75/100** | — | **GOLD** |

---

## What It Generates

```
out/
├── recovery-scan.json          # Full cluster inventory + scores + findings
├── recovery-enriched.json      # Enriched DR analysis with trend + risk
├── recovery-report.md          # Markdown summary
├── recovery-report.html        # Self-contained dark-mode tabbed HTML report
├── history/
│   └── index.json              # Trend history across scans
└── csv/                        # (--csv flag) one file per inventory tab
    ├── nodes.csv
    ├── workloads.csv
    ├── storage.csv
    ├── networking.csv
    ├── config.csv
    ├── images.csv
    ├── helm.csv
    ├── certificates.csv
    ├── dr-score.csv
    └── remediation.csv
```

The HTML report is fully self-contained (no CDN, no external dependencies) — safe to open offline in customer environments.

---

## Report Tabs

| Tab | Content |
|-----|---------|
| **Summary** | Score card, maturity badge, platform, backup tool status, findings count |
| **Nodes** | Node name, roles, OS image, kernel, container runtime, ready status, taints |
| **Workloads** | All workload types (Deployments, StatefulSets, DaemonSets, Jobs, CronJobs) |
| **Storage** | PVCs + PVs + StorageClasses with binding status, backend, reclaim policy |
| **Networking** | Services, Ingresses with TLS status, NetworkPolicies |
| **Config** | ConfigMaps, Secrets, CRDs, ClusterRoles, Helm releases, Certificates |
| **Images** | Container images grouped by registry; public vs. private |
| **Backup** | Detected tools, backup policies with RPO + offsite flag, restore simulation per namespace |
| **DR Score** | 4-domain scoring breakdown + all findings sorted by severity |
| **Remediation** | Prioritized, tool-specific remediation steps with commands |
| **Compare** | Scan-to-scan diff (only shown when `--compare` is used) |

---

## Quick Start

### Prerequisites

- Go 1.21+ **or** a pre-built binary from [Releases](../../releases)
- A valid `kubeconfig` with read access to the cluster

### Build

**Linux / macOS**
```bash
make build
# binary: ./scan
```

**One-liner without make**
```bash
go build -o scan ./cmd/scan
```

**Windows**
```powershell
go build -o scan.exe ./cmd/scan
```

### Run

**Linux / macOS**
```bash
# Basic scan (VM recovery target)
./scan --out ./out

# Scoped to specific namespaces
./scan --namespace=prod,staging --out ./out

# Bare metal recovery target with CSV export
./scan --target=baremetal --csv --out ./out

# Diff against a previous scan
./scan --compare=./previous/recovery-scan.json --out ./out

# CI mode (exit code 2 if score below threshold)
./scan --ci --min-score=75 --out ./out

# Write a redacted JSON copy (no secret values)
./scan --redact --out ./out

# Dry run (no cluster required)
./scan --dry-run --out ./out
```

**Windows**
```powershell
.\scan.exe --out .\out
.\scan.exe --namespace=prod,staging --out .\out
.\scan.exe --compare=.\previous\recovery-scan.json --out .\out
.\scan.exe --redact --out .\out
.\scan.exe --ci --min-score=75 --out .\out
```

### Open Report

```bash
# Linux
xdg-open ./out/recovery-report.html

# macOS
open ./out/recovery-report.html

# Windows
start .\out\recovery-report.html
```

---

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--kubeconfig` | `""` | Path to kubeconfig (uses in-cluster config if empty) |
| `--out` | `./out` | Output directory |
| `--target` | `vm` | Recovery target: `baremetal` or `vm` |
| `--namespace` | `""` | Comma-separated namespaces to scan (empty = all namespaces) |
| `--compare` | `""` | Path to a previous `recovery-scan.json` to diff against |
| `--csv` | `false` | Write CSV exports to `out/csv/` |
| `--summary` | `false` | Print a one-line summary to stdout on completion |
| `--redact` | `false` | Write a redacted JSON copy with secret values removed |
| `--dry-run` | `false` | Run without a cluster (for testing) |
| `--ci` | `false` | CI mode: emit JSON summary + exit code 2 on failure |
| `--min-score` | `90` | Minimum acceptable overall score for CI pass |
| `--timeout` | `60` | Kubernetes API timeout in seconds |
| `--customer` | `""` | Customer identifier embedded in report metadata |
| `--site` | `""` | Site/region name embedded in report metadata |
| `--cluster` | `""` | Cluster name embedded in report metadata |
| `--env` | `""` | Environment tag (prod/dev/test) embedded in report metadata |

---

## Backup Tool Detection & Policy Analysis

The tool automatically detects these backup solutions and — for supported tools — collects detailed policy data:

| Tool | Detection | Policy Collection |
|------|-----------|-------------------|
| **Kasten K10** | `kasten-io` namespace, `kio.kasten.io` CRDs | Policies: frequency, namespace selector, export actions |
| **Velero** | `velero` namespace, `velero.io` CRDs | Schedules: namespace coverage, cron, TTL, storage location |
| **Longhorn** | `longhorn-system` namespace, `longhorn.io` CRDs | RecurringJobs (backup tasks), BackupTarget setting |
| **Rubrik** | `rubrik`/`rbs` namespace, `rubrik.com` CRDs | Detection only |
| **Trilio** | `trilio-system` namespace, `triliovault.trilio.io` CRDs | Detection only |
| **Stash** | `stash` namespace, `stash.appscode.com` CRDs | Detection only |
| **CloudCasa** | `cloudcasa-io` namespace, `cloudcasa.io` CRDs | Detection only |

RPO is estimated from cron expressions and frequency labels (`@daily`, `@weekly`, `*/6 * * * *`, etc.).

An offsite/export location is detected from Velero storage locations (non-default), Kasten export actions, and Longhorn BackupTarget settings.

---

## Restore Simulation

After backup detection, the tool runs a per-namespace restore feasibility assessment for every namespace containing StatefulSets or PVCs:

| Field | Description |
|-------|-------------|
| **Coverage** | Whether at least one backup policy covers the namespace |
| **RPO (h)** | Best-case RPO in hours from applicable policies |
| **PVC Data (GB)** | Total persistent storage that would need to be restored |
| **Blockers** | hostPath volumes, unbound PVCs — prevent clean restore |
| **Warnings** | StorageClasses referenced in PVCs but not present in cluster |

Results are visible in the **Backup** tab and drive the `BACKUP_NO_OFFSITE`, `BACKUP_RPO_HIGH`, and `RESTORE_SIM_UNCOVERED` scoring rules.

---

## Platform Detection

Provider is detected automatically from node labels:

| Provider | Detection |
|----------|-----------|
| **EKS** | `eks.amazonaws.com/*` node labels |
| **AKS** | `kubernetes.azure.com/*` node labels |
| **GKE** | `cloud.google.com/*` node labels |
| **Rancher** | `rancher.io/*` StorageClass provisioners |
| **k3s** | `k3s.io/*` node labels |
| **Vanilla** | Fallback |

---

## Use Cases

- Pre-migration DR assessment before cluster rebuild
- Customer environment intake validation for DRaaS onboarding
- Identifying backup gaps and offsite coverage before a DR event
- Repeatable DR maturity measurement over time
- Scan-to-scan comparison to track posture changes across sprints
- CSV/Excel inventory export for documentation or handoff

---

## Example Report

### Summary Tab
![Summary Tab](images/report-summary.png)

### DR Score Tab
![DR Score Tab](images/report-dr-score.png)

---

## Design Principles

- No external dependencies at runtime (reads only from the K8s API)
- No Helm CLI required (reads Helm release secrets directly)
- No cert-manager SDK required (reads CRs via raw REST)
- Self-contained HTML output — safe for air-gapped environments
- Deterministic scoring — same cluster always produces the same score
- Historical trend tracking across repeated scans
