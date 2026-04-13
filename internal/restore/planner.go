package restore

import (
	"fmt"
	"sort"
	"strings"

	"k8s-recovery-visualizer/internal/model"
)

func BuildDrillPlan(b *model.Bundle) []model.RestoreDrillStep {
	steps := []model.RestoreDrillStep{
		{
			Phase:     "prepare",
			Title:     "Freeze the recovery scope and evidence set",
			Detail:    fmt.Sprintf("Use the %s profile bundle for cluster %s and confirm the target is %s before the drill starts.", defaultString(b.Profile, "standard"), defaultString(b.Metadata.ClusterName, "unknown cluster"), defaultString(b.Target, "vm")),
			OwnerHint: "Platform / incident lead",
			Validation: []string{
				"Confirm the latest recovery-scan.json is attached to the drill ticket.",
				"Record the target cluster, namespace scope, and declared recovery destination.",
			},
		},
	}

	backup := b.Inventory.Backup
	if backup.PrimaryTool == "" || backup.PrimaryTool == "none" {
		steps = append(steps, model.RestoreDrillStep{
			Phase:     "backup",
			Title:     "Establish a backup evidence baseline before testing restore",
			Detail:    "No backup tool was detected, so the drill should begin with manual evidence capture or installation validation before any simulated restore.",
			OwnerHint: "Platform / backup owner",
			Validation: []string{
				"Capture the backup product, policy name, retention, and offsite target in the drill notes.",
			},
		})
	} else {
		detail := fmt.Sprintf("Validate %s policies, last-success timestamps, and offsite destinations before touching workloads.", backup.PrimaryTool)
		if backup.RestoreSim != nil {
			if backup.RestoreSim.UnknownNamespaces > 0 {
				detail += fmt.Sprintf(" %d namespace(s) still have unverified coverage.", backup.RestoreSim.UnknownNamespaces)
			}
			if len(backup.RestoreSim.UncoveredNS) > 0 {
				detail += fmt.Sprintf(" Uncovered namespaces: %s.", strings.Join(backup.RestoreSim.UncoveredNS, ", "))
			}
		}
		steps = append(steps, model.RestoreDrillStep{
			Phase:     "backup",
			Title:     "Validate backup and offsite evidence",
			Detail:    detail,
			OwnerHint: "Platform / backup owner",
			Validation: []string{
				"Review backup schedule scope and RPO assumptions.",
				"Confirm the most recent successful run and its offsite/export destination.",
			},
		})
	}

	if sim := backup.RestoreSim; sim != nil {
		order := []string{}
		for _, ns := range sim.Namespaces {
			switch ns.Readiness {
			case "ready", "warning":
				order = append(order, ns.Namespace)
			}
		}
		sort.Strings(order)
		detail := "Restore cluster-scoped dependencies first, then stateful namespaces, then stateless services."
		if len(order) > 0 {
			detail = fmt.Sprintf("%s Suggested namespace order: %s.", detail, strings.Join(order, ", "))
		}
		if len(sim.BlockingReasons) > 0 {
			detail += fmt.Sprintf(" Clear blockers before the drill: %s.", strings.Join(sim.BlockingReasons, "; "))
		}
		steps = append(steps, model.RestoreDrillStep{
			Phase:     "restore",
			Title:     "Execute a dependency-aware restore sequence",
			Detail:    detail,
			OwnerHint: "Platform engineering",
			Validation: []string{
				"Restore CRDs, storage classes, and RBAC prerequisites before namespaced workloads.",
				"Restore data-bearing namespaces before frontend or integration tiers.",
			},
		})
		steps = append(steps, model.RestoreDrillStep{
			Phase:     "validate",
			Title:     "Prove application readiness after data restore",
			Detail:    fmt.Sprintf("Validate recovered workloads against the estimated %.1f GiB protected volume and compare service endpoints, certificates, and policies with the source bundle.", sim.CoveredPVCsGB),
			OwnerHint: "Application owner",
			Validation: []string{
				"Run namespace smoke checks and confirm PVC binding, service endpoints, and secrets/config are present.",
				"Capture any warning or blocker differences in a follow-up bundle.",
			},
		})
	}

	steps = append(steps, model.RestoreDrillStep{
		Phase:     "closeout",
		Title:     "Re-scan the recovered environment and compare the results",
		Detail:    "Generate a fresh bundle from the restored target and compare it to the pre-drill baseline to record regressions, resolved gaps, and evidence drift.",
		OwnerHint: "Platform / incident lead",
		Validation: []string{
			"Run cmd/check or load the new bundle into K8V compare view.",
			"Record new, resolved, and regressed findings in the drill retro.",
		},
	})

	return steps
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
