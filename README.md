# k8s-recovery-visualizer

Full Kubernetes cluster inventory and Disaster Recovery assessment tool.

Scan a live cluster to get a complete RVTools-style inventory, a weighted DR readiness score across four domains, backup tool detection, platform identification, and a prioritized remediation plan — all in a single self-contained HTML report.

---

## Why This Exists

Building DR clusters without structured discovery leads to:
- Missing stateful workloads with no backup coverage
- Storage classes that don't exist in the target environment
- Orphaned PVs with Delete reclaim policies that vanish on PVC deletion
- Public registry images that aren't reachable in air-gapped DR sites
- No backup tool, or a backup tool with no policies configured

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
| **Backup** | Auto-detects Kasten K10, Velero, Rubrik, Longhorn, Trilio, Stash, CloudCasa |

---

## DR Scoring Model

Scoring covers four weighted domains:

| Domain | Weight | What It Measures |
|--------|--------|-----------------|
| **Storage** | 35% | PVC binding, storageClass presence, hostPath usage, reclaim policies |
| **Workload** | 20% | StatefulSet persistence, deployment coverage |
| **Config** | 15% | CRD backup readiness, certificate expiry, image registry risk |
| **Backup/Recovery** | 30% | Backup tool presence, policy coverage, Helm values, public images |

**Maturity levels:** PLATINUM (≥90) · GOLD (≥75) · SILVER (≥50) · BRONZE (<50)

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
| **DR Score** | 4-domain scoring breakdown + all findings sorted by severity |
| **Remediation** | Prioritized, tool-specific remediation steps with commands |

---

## Quick Start

### Prerequisites

- Go 1.21+
- A valid `kubeconfig` with read access to the cluster

### Build

```powershell
go build -o scan.exe ./cmd/scan
```

### Run

```powershell
# Basic scan (VM recovery target, no CSV)
.\scan.exe --out .\out

# Bare metal recovery target with CSV export
.\scan.exe --target=baremetal --csv --out .\out

# With explicit kubeconfig
.\scan.exe --kubeconfig C:\Users\you\.kube\config --target=vm --csv --out .\out

# CI mode (exit code 2 if score below threshold)
.\scan.exe --ci --min-score=75 --out .\out

# Dry run (no cluster required)
.\scan.exe --dry-run --out .\out
```

### Open Report

```powershell
Start-Process .\out\recovery-report.html
```

---

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--kubeconfig` | `""` | Path to kubeconfig (uses in-cluster config if empty) |
| `--out` | `./out` | Output directory |
| `--target` | `vm` | Recovery target: `baremetal` or `vm` |
| `--csv` | `false` | Write CSV exports to `out/csv/` |
| `--dry-run` | `false` | Run without a cluster (for testing) |
| `--ci` | `false` | CI mode: emit JSON summary + exit code 2 on failure |
| `--min-score` | `90` | Minimum acceptable overall score for CI pass |
| `--timeout` | `60` | Kubernetes API timeout in seconds |
| `--customer` | `""` | Customer identifier embedded in report metadata |
| `--site` | `""` | Site/region name embedded in report metadata |
| `--cluster` | `""` | Cluster name embedded in report metadata |
| `--env` | `""` | Environment tag (prod/dev/test) embedded in report metadata |

---

## Backup Tool Detection

The tool automatically detects these backup solutions — no configuration required:

| Tool | Detection Method |
|------|-----------------|
| **Kasten K10** | `kasten-io` namespace, `policies.config.kio.kasten.io` CRD |
| **Velero** | `velero` namespace, `backups.velero.io` CRD |
| **Rubrik** | `rubrik`/`rbs` namespace, `rubrik-backup-service` pods |
| **Longhorn** | `longhorn-system` namespace, `driver.longhorn.io` StorageClass provisioner |
| **Trilio** | `trilio-system` namespace, `backupplans.triliovault.trilio.io` CRD |
| **Stash** | `stash.appscode.com` CRD group |
| **CloudCasa** | `cloudcasa-io` namespace |

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
- Identifying backup gaps before a DR event
- Repeatable DR maturity measurement over time
- CSV/Excel inventory export for documentation or handoff

---

## Example Report

![Sample DR Report](images/sample-report.png)

---

## Design Principles

- No external dependencies at runtime (reads only from the K8s API)
- No Helm CLI required (reads Helm release secrets directly)
- No cert-manager SDK required (reads CRs via raw REST)
- Self-contained HTML output — safe for air-gapped environments
- Deterministic scoring — same cluster always produces the same score
- Historical trend tracking across repeated scans
