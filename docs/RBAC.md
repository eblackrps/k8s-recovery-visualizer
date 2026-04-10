# RBAC

The scanner supports two supported deployment modes:

1. Cluster-wide scan
2. Namespace-scoped scan

Use the published manifests instead of inventing ad hoc permissions:

- [`../deploy/rbac/cluster-scan.yaml`](../deploy/rbac/cluster-scan.yaml)
- [`../deploy/rbac/namespace-scan.yaml`](../deploy/rbac/namespace-scan.yaml)

## Cluster-wide scan

Use this when you want the full default scoring model, backup inspection, RBAC audit, and cluster-wide trend reporting.

Example:

```bash
kubectl apply -f deploy/rbac/cluster-scan.yaml
kubectl -n kube-system create token k8vis-scan
scan --kubeconfig ./kubeconfig --out ./out
```

What it unlocks:

- Full default collector coverage
- Cluster-scoped storage and node scoring
- Backup policy inspection for supported tools
- RBAC privilege findings
- Snapshot-class and cluster-wide restore context

## Namespace-scoped scan

Use this when platform teams want the smallest supported permission set for one namespace or a small set of namespaces.

Before applying the manifest:

- Replace `REPLACE_NAMESPACE` in [`../deploy/rbac/namespace-scan.yaml`](../deploy/rbac/namespace-scan.yaml)
- Run the CLI with `--namespace=<namespace>`

Example:

```bash
kubectl apply -f deploy/rbac/namespace-scan.yaml
kubectl -n prod create token k8vis-scan
scan --namespace=prod --kubeconfig ./kubeconfig --out ./out
```

For multiple namespaces, apply a copy of the namespaced `Role` and `RoleBinding` in each scanned namespace and keep the shared `ClusterRoleBinding` if you want node, PV, namespace, and StorageClass context.

## Degraded features when permissions are absent

The scanner records permission failures under `collectorSkips` and renders them in the HTML report. Missing access does not silently disappear.

| Missing permission | What degrades |
| --- | --- |
| `nodes` | node readiness, single-AZ scoring, platform hints |
| `persistentvolumes` | PVC-to-PV binding analysis, hostPath detection, reclaim-policy findings |
| `storageclasses` | storage predictability findings, restore blocker warnings |
| `clusterroles` / `clusterrolebindings` | RBAC escalation and wildcard findings |
| `snapshot.storage.k8s.io/volumesnapshotclasses` | snapshot readiness scoring |
| `velero.io/schedules` / `config.kio.kasten.io/policies` / `longhorn.io/*` | backup inspection falls back to `unverified` |
| `cert-manager.io/certificates` | certificate expiry findings disappear |

## Optional Secret metadata collection

The published manifests intentionally do not grant Secret reads.

That is deliberate:

- the default scan does not need Secret objects
- Kubernetes Secret reads expose full Secret payloads to the client, even if the collector only reports metadata
- the scanner should not ask for that access unless an operator opts in

If you still want `inventory.secrets`, do both of these:

1. Run the CLI with `--include-secret-metadata`
2. Extend the manifest yourself with `get/list/watch` on `secrets`

Treat that as an explicit security tradeoff, not part of the baseline package.

## Safe production deployment notes

- Prefer a dedicated read-only service account instead of reusing an operator token.
- Keep `--insecure` off unless you are scanning a cluster with a known self-signed or broken certificate chain.
- Review redacted artifacts before sharing them outside the cluster team.
- Treat `coverageStatus != verified` as an operator follow-up item, not a cosmetic warning.
- Keep Secret metadata collection off unless you have a specific review need and approved RBAC for it.
