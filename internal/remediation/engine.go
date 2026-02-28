// Package remediation generates prioritized, tool-specific DR remediation steps
// from a scan bundle.
package remediation

import (
	"fmt"
	"sort"
	"strings"

	"k8s-recovery-visualizer/internal/model"
)

// Generate produces a sorted list of remediation steps from the bundle.
// target is "baremetal" or "vm".
func Generate(b *model.Bundle, target string) []model.RemediationStep {
	tool := b.Inventory.Backup.PrimaryTool
	if tool == "" {
		tool = "none"
	}

	var steps []model.RemediationStep

	for _, f := range b.Inventory.Findings {
		if s := stepForFinding(f, tool, target, b.Cluster.Platform.Provider); s != nil {
			steps = append(steps, *s)
		}
	}

	// Global backup install step if nothing detected
	if tool == "none" {
		steps = append(steps, backupInstallStep(target))
	}

	// Helm values backup reminder
	if len(b.Inventory.HelmReleases) > 0 {
		steps = append(steps, helmValuesStep(b.Inventory.HelmReleases))
	}

	// Image mirroring reminder when public images present
	publicCount := 0
	for _, img := range b.Inventory.Images {
		if img.IsPublic {
			publicCount++
		}
	}
	if publicCount > 0 {
		steps = append(steps, imageMirrorStep(publicCount, target))
	}

	sort.SliceStable(steps, func(i, j int) bool {
		return steps[i].Priority < steps[j].Priority
	})
	return steps
}

func stepForFinding(f model.Finding, tool, target, platform string) *model.RemediationStep {
	switch f.ID {
	case "BACKUP_NONE":
		s := backupInstallStep(target)
		s.FindingID = f.ID
		return &s

	case "BACKUP_PARTIAL_COVERAGE":
		return &model.RemediationStep{
			Priority:  1,
			Category:  "Backup",
			Title:     "Extend backup coverage to all stateful namespaces",
			Detail:    fmt.Sprintf("StatefulSets in uncovered namespaces: %s", f.ResourceID),
			Commands:  backupPolicyCmds(tool, f.ResourceID),
			FindingID: f.ID,
		}

	case "BACKUP_NO_POLICIES":
		return &model.RemediationStep{
			Priority:  1,
			Category:  "Backup",
			Title:     "Create backup schedules for all production namespaces",
			Detail:    "Backup tool detected but no backup policies or schedules found.",
			Commands:  backupScheduleCmds(tool),
			FindingID: f.ID,
		}

	case "PVC_UNBOUND":
		return &model.RemediationStep{
			Priority:  1,
			Category:  "Storage",
			Title:     fmt.Sprintf("Fix unbound PVC: %s", f.ResourceID),
			Detail:    "Unbound PVCs will not be backed up or restored correctly.",
			Commands:  []string{"kubectl describe pvc -n " + nsFromRef(f.ResourceID), "kubectl get events -n " + nsFromRef(f.ResourceID) + " --field-selector reason=FailedBinding"},
			FindingID: f.ID,
		}

	case "PV_HOSTPATH":
		notes := ""
		if target == "baremetal" {
			notes = "Bare metal recovery: consider Longhorn (helm install longhorn longhorn/longhorn -n longhorn-system) or OpenEBS as a CSI replacement."
		} else {
			notes = "VM recovery: use the appropriate cloud or vSphere CSI driver for your target environment."
		}
		if platform == "EKS" {
			notes = "EKS: use the AWS EBS CSI driver (aws-ebs-csi-driver)."
		} else if platform == "AKS" {
			notes = "AKS: use the Azure Disk CSI driver (disk.csi.azure.com)."
		} else if platform == "GKE" {
			notes = "GKE: use the GCE PD CSI driver (pd.csi.storage.gke.io)."
		}
		return &model.RemediationStep{
			Priority:    1,
			Category:    "Storage",
			Title:       fmt.Sprintf("Migrate hostPath PV to CSI storage: %s", f.ResourceID),
			Detail:      "hostPath volumes are node-local and cannot be restored to a different node.",
			Commands:    []string{"# Provision a new PV using a CSI driver", "# Migrate data: kubectl cp or use a migration tool", "# Update PVC to reference new PV"},
			TargetNotes: notes,
			FindingID:   f.ID,
		}

	case "PV_DELETE_POLICY":
		return &model.RemediationStep{
			Priority:  2,
			Category:  "Storage",
			Title:     fmt.Sprintf("Change PV reclaim policy to Retain: %s", f.ResourceID),
			Detail:    "Delete policy means the PV is deleted when the PVC is deleted, making restore impossible.",
			Commands:  []string{fmt.Sprintf("kubectl patch pv %s -p '{\"spec\":{\"persistentVolumeReclaimPolicy\":\"Retain\"}}'", f.ResourceID)},
			FindingID: f.ID,
		}

	case "PVC_NO_STORAGECLASS":
		return &model.RemediationStep{
			Priority:  2,
			Category:  "Storage",
			Title:     fmt.Sprintf("Add explicit storageClass to PVC: %s", f.ResourceID),
			Detail:    "PVCs without a storageClass rely on cluster defaults which may not exist in the DR environment.",
			Commands:  []string{"# Edit PVC spec to set storageClassName explicitly", fmt.Sprintf("kubectl edit pvc %s -n %s", nameFromRef(f.ResourceID), nsFromRef(f.ResourceID))},
			FindingID: f.ID,
		}

	case "STS_NO_PVC":
		return &model.RemediationStep{
			Priority:  2,
			Category:  "Workload",
			Title:     fmt.Sprintf("Add volumeClaimTemplate to StatefulSet: %s", f.ResourceID),
			Detail:    "StatefulSets without persistent storage lose all data on pod restart or DR event.",
			Commands:  []string{fmt.Sprintf("kubectl edit statefulset %s -n %s", nameFromRef(f.ResourceID), nsFromRef(f.ResourceID))},
			FindingID: f.ID,
		}

	case "CRD_NO_BACKUP":
		return &model.RemediationStep{
			Priority:  2,
			Category:  "Config",
			Title:     "Capture CRD definitions before DR",
			Detail:    "Custom Resource Definitions must be restored before custom resources can be recreated.",
			Commands:  []string{"kubectl get crds -o yaml > crds-backup.yaml", "# Store crds-backup.yaml in a safe external location"},
			FindingID: f.ID,
		}

	case "CERT_EXPIRING_SOON":
		return &model.RemediationStep{
			Priority:  1,
			Category:  "Config",
			Title:     fmt.Sprintf("Renew expiring certificate: %s", f.ResourceID),
			Detail:    "A certificate expiring during a DR event will cause service failures.",
			Commands:  []string{"# cert-manager: annotate to force renewal", fmt.Sprintf("kubectl annotate certificate %s -n %s cert-manager.io/force-renewal=true", nameFromRef(f.ResourceID), nsFromRef(f.ResourceID))},
			FindingID: f.ID,
		}

	case "IMAGE_EXTERNAL_REGISTRY":
		return &model.RemediationStep{
			Priority:    2,
			Category:    "Config",
			Title:       "Mirror public container images to a private registry",
			Detail:      "Public registry images may be unavailable in an air-gapped or restricted DR environment.",
			Commands:    []string{"# List all public images used:", "# kubectl get pods -A -o jsonpath='{range .items[*]}{.spec.containers[*].image}{\"\\n\"}{end}' | sort -u | grep -v <private-registry>"},
			TargetNotes: targetImageNote(target),
			FindingID:   f.ID,
		}

	case "HELM_UNTRACKED":
		return &model.RemediationStep{
			Priority:  3,
			Category:  "Config",
			Title:     "Back up Helm release values",
			Detail:    "Helm values are needed to reinstall releases in the DR environment.",
			Commands:  []string{"# Export all release values:", "helm list -A -o json | jq -r '.[].name' | xargs -I{} sh -c 'helm get values {} > {}-values.yaml'"},
			FindingID: f.ID,
		}
	}
	return nil
}

func backupInstallStep(target string) model.RemediationStep {
	kastenNote := "# Kasten K10 install:\nhelm repo add kasten https://charts.kasten.io/\nhelm install k10 kasten/k10 --namespace kasten-io --create-namespace"
	veleroNote := "# Velero install (example with AWS S3):\nvelero install --provider aws --plugins velero/velero-plugin-for-aws:v1.9.0 \\\n  --bucket <BUCKET> --backup-location-config region=<REGION> \\\n  --secret-file ./credentials-velero"
	targetNote := ""
	if target == "baremetal" {
		targetNote = "Bare metal target: Kasten K10 or Velero with MinIO as local S3 backend are good choices. Longhorn provides native snapshot support."
	} else {
		targetNote = "VM target: Kasten K10 or Velero with cloud storage backend (S3, Azure Blob, GCS) are recommended."
	}
	return model.RemediationStep{
		Priority:    1,
		Category:    "Backup",
		Title:       "Install a backup solution — no backup tool detected",
		Detail:      "Without a backup tool, cluster recovery is not possible. Install Kasten K10, Velero, or Rubrik.",
		Commands:    []string{kastenNote, veleroNote},
		TargetNotes: targetNote,
	}
}

func backupScheduleCmds(tool string) []string {
	switch tool {
	case "velero":
		return []string{
			"# Create a daily Velero backup schedule covering all namespaces:",
			"velero schedule create daily-full --schedule='0 2 * * *' --include-namespaces='*'",
		}
	case "kasten":
		return []string{
			"# In the Kasten K10 UI, navigate to Policies → Create Policy",
			"# Set frequency to Daily, include all namespaces, configure export location",
		}
	default:
		return []string{"# Configure backup schedules in your backup tool covering all production namespaces"}
	}
}

func backupPolicyCmds(tool, namespaces string) []string {
	switch tool {
	case "velero":
		return []string{
			fmt.Sprintf("# Add these namespaces to your Velero schedule: %s", namespaces),
			"velero schedule describe daily-full",
			"# Update schedule with --include-namespaces or create additional schedules",
		}
	default:
		return []string{fmt.Sprintf("# Extend backup policies to cover namespaces: %s", namespaces)}
	}
}

func helmValuesStep(releases []model.HelmRelease) model.RemediationStep {
	cmds := []string{"# Export values for each Helm release:"}
	for _, r := range releases {
		cmds = append(cmds, fmt.Sprintf("helm get values %s -n %s > %s-values.yaml", r.Name, r.Namespace, r.Name))
	}
	cmds = append(cmds, "# Store these files in a safe external location (git, S3, etc.)")
	return model.RemediationStep{
		Priority: 3,
		Category: "Config",
		Title:    fmt.Sprintf("Back up values for %d Helm release(s)", len(releases)),
		Detail:   "Helm release values are required to reinstall applications in the DR environment.",
		Commands: cmds,
	}
}

func imageMirrorStep(count int, target string) model.RemediationStep {
	return model.RemediationStep{
		Priority:    2,
		Category:    "Config",
		Title:       fmt.Sprintf("Mirror %d public image(s) to a private registry", count),
		Detail:      "Public registry images may be unavailable in the DR environment if network access is restricted.",
		Commands:    []string{"# List all public images:\nkubectl get pods -A -o jsonpath='{range .items[*]}{.spec.containers[*].image}{\"\\n\"}{end}' | sort -u", "# Pull and push to private registry:\n# docker pull <image> && docker tag <image> <private-registry>/<image> && docker push <private-registry>/<image>"},
		TargetNotes: targetImageNote(target),
	}
}

func targetImageNote(target string) string {
	if target == "baremetal" {
		return "Bare metal target: set up a private registry (Harbor, Nexus) reachable from recovery nodes, or pre-pull images to all nodes."
	}
	return "VM target: use a registry in the same cloud region or VPN-accessible private registry to reduce pull latency."
}

func nsFromRef(ref string) string {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) == 2 {
		return parts[0]
	}
	return "default"
}

func nameFromRef(ref string) string {
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ref
}
