# Support Matrix

This project is conservative about support claims. The matrix below describes code-path and test reality, not marketing intent.

## Kubernetes API expectations

| Area | Current reality |
| --- | --- |
| Go client | `client-go v0.35.1` |
| Stable APIs assumed by collectors | `apps/v1`, `batch/v1`, `networking.k8s.io/v1`, `rbac.authorization.k8s.io/v1`, `autoscaling/v2`, `policy/v1`, `snapshot.storage.k8s.io/v1` |
| Fixture coverage in repo | Synthetic bundles modeled on modern Kubernetes clusters with 1.30-style APIs |
| Namespace-scoped mode | Supported with documented degradation and dedicated RBAC packaging |

## Distribution posture

| Distribution / platform | Status | Notes |
| --- | --- | --- |
| Generic upstream Kubernetes | Supported | Primary collector path uses standard Kubernetes APIs. |
| EKS | Supported heuristics | Provider detection and storage guidance exist; not every AWS-specific backup workflow is inspected. |
| AKS | Supported heuristics | Provider detection and Azure CSI guidance exist. |
| GKE | Supported heuristics | Provider detection exists; backup inspection still depends on in-cluster tools. |
| Rancher / RKE2 | Partial | Platform hints exist; self-signed TLS often requires explicit kubeconfig trust or `--insecure`. |
| k3s | Partial | Provider detection exists; local-path / hostPath style storage often produces portability findings. |

## Scope limitations

| Scan mode | Strengths | Limitations |
| --- | --- | --- |
| Cluster-wide | Full default scoring model, backup inspection, cluster RBAC audit | Needs broad read-only access, but default manifests intentionally exclude Secret reads |
| Namespace-scoped | Smallest supported application assessment | Some cluster-wide findings and backup evidence become degraded or skipped |

## Explicit non-claims

- This tool does not prove that a backup can be restored until operators run a real drill.
- Detection-only backup products are inventory signals, not coverage proof.
- A passing score does not replace workload-specific DR runbooks, RTO validation, or business-impact testing.
