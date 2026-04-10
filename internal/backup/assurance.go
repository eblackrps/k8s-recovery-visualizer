package backup

import "k8s-recovery-visualizer/internal/model"

// AssessAssurance derives operator-facing assurance signals from backup policy
// inspection, restore simulation, snapshot capability, and recent-success
// evidence. It intentionally stays conservative: missing evidence lowers
// confidence instead of implying recoverability.
func AssessAssurance(b *model.Bundle) {
	inv := &b.Inventory.Backup
	signals := []model.BackupEvidenceSignal{}

	addSignal := func(id, status, summary, detail string, confidence model.EvidenceConfidence) {
		signals = append(signals, model.BackupEvidenceSignal{
			ID:         id,
			Status:     status,
			Summary:    summary,
			Detail:     detail,
			Confidence: confidence,
		})
	}

	if inv.PrimaryTool == "" || inv.PrimaryTool == "none" {
		addSignal("tool-detected", "missing", "No backup tool detected.", "", model.EvidenceConfidenceConfirmed)
		inv.Assurance = &model.BackupAssurance{
			Conclusion: model.BackupAssuranceAtRisk,
			Confidence: model.EvidenceConfidenceConfirmed,
			Summary:    "No backup tool was detected, so recoverability is at risk.",
			Signals:    signals,
		}
		return
	}

	if !inv.CoverageVerified {
		addSignal("coverage", "unverified", "Namespace coverage could not be verified from the detected tool.", inv.CoverageReason, model.EvidenceConfidenceConfirmed)
		inv.Assurance = &model.BackupAssurance{
			Conclusion: model.BackupAssuranceUnverified,
			Confidence: model.EvidenceConfidenceConfirmed,
			Summary:    "A backup product was detected, but policy coverage could not be verified.",
			Signals:    signals,
		}
		return
	}

	addSignal("coverage", "confirmed", "Backup namespace coverage was verified from inspected policies.", inv.CoverageReason, model.EvidenceConfidenceConfirmed)

	recentConfidence := model.EvidenceConfidenceUnknown
	recentStatus := "unknown"
	recentSummary := "Recent successful backup evidence was not exposed by the inspected tool."
	recentDetail := ""
	hasRecentEvidence := false
	hasFreshRun := false
	hasStaleRun := false
	for _, policy := range inv.Policies {
		if policy.LastSuccessAt == "" {
			continue
		}
		hasRecentEvidence = true
		recentConfidence = model.EvidenceConfidenceConfirmed
		if policy.FreshSchedule {
			hasFreshRun = true
		} else {
			hasStaleRun = true
			recentDetail = "At least one inspected policy has a last successful run older than the configured freshness threshold."
		}
	}
	switch {
	case hasFreshRun && !hasStaleRun:
		recentStatus = "confirmed"
		recentSummary = "Recent successful backup evidence is present."
	case hasStaleRun:
		recentStatus = "warning"
		recentSummary = "Backup run history is stale relative to the configured schedule."
	case hasRecentEvidence:
		recentStatus = "confirmed"
		recentSummary = "Successful backup evidence is present, but freshness could not be determined."
	}
	addSignal("recent-success", recentStatus, recentSummary, recentDetail, recentConfidence)

	if inv.HasOffsite {
		addSignal("offsite", "confirmed", "An offsite or secondary backup target was detected.", "", model.EvidenceConfidenceConfirmed)
	} else {
		addSignal("offsite", "missing", "No offsite or secondary backup target was detected.", "", model.EvidenceConfidenceConfirmed)
	}

	if len(b.Inventory.PVCs) == 0 {
		addSignal("snapshot-readiness", "not_applicable", "No PVCs were present, so snapshot readiness is not applicable.", "", model.EvidenceConfidenceConfirmed)
	} else if len(b.Inventory.VolumeSnapshotClasses) == 0 {
		addSignal("snapshot-readiness", "missing", "No VolumeSnapshotClass is available for point-in-time volume protection.", "", model.EvidenceConfidenceConfirmed)
	} else if len(b.Inventory.VolumeSnapshots) > 0 {
		addSignal("snapshot-readiness", "confirmed", "Snapshot classes and at least one VolumeSnapshot are present.", "", model.EvidenceConfidenceConfirmed)
	} else {
		addSignal("snapshot-readiness", "warning", "Snapshot capability exists, but no VolumeSnapshots were found for the current PVC set.", "", model.EvidenceConfidenceInferred)
	}

	if len(inv.UncoveredStatefulNS) > 0 {
		addSignal("coverage-gaps", "missing", "Some stateful namespaces are outside verified backup coverage.", "", model.EvidenceConfidenceConfirmed)
	} else {
		addSignal("coverage-gaps", "confirmed", "No uncovered stateful namespaces were found in verified policy scope.", "", model.EvidenceConfidenceConfirmed)
	}

	blockerCount := 0
	warningCount := 0
	if sim := inv.RestoreSim; sim != nil {
		for _, ns := range sim.Namespaces {
			if len(ns.Blockers) > 0 {
				blockerCount++
			}
			if len(ns.Warnings) > 0 {
				warningCount++
			}
		}
	}
	switch {
	case blockerCount > 0:
		addSignal("restore-dependencies", "missing", "Restore simulation found blockers that prevent clean recovery.", "", model.EvidenceConfidenceConfirmed)
	case warningCount > 0:
		addSignal("restore-dependencies", "warning", "Restore simulation found dependency warnings that should be reviewed before DR.", "", model.EvidenceConfidenceConfirmed)
	default:
		addSignal("restore-dependencies", "confirmed", "Restore simulation did not find dependency blockers.", "", model.EvidenceConfidenceConfirmed)
	}

	conclusion := model.BackupAssuranceConfirmedRecoverable
	confidence := model.EvidenceConfidenceConfirmed
	summary := "Backup coverage, offsite protection, and restore prerequisites are all confirmed."

	switch {
	case len(inv.UncoveredStatefulNS) > 0 || blockerCount > 0 || !inv.HasOffsite || hasStaleRun:
		conclusion = model.BackupAssuranceCoverageGap
		confidence = model.EvidenceConfidenceConfirmed
		summary = "Recoverability has confirmed gaps or blockers that need attention."
	case !hasRecentEvidence:
		conclusion = model.BackupAssuranceInferredRecoverable
		confidence = model.EvidenceConfidenceInferred
		summary = "Coverage looks recoverable, but recent successful backup evidence could not be confirmed."
	}

	inv.Assurance = &model.BackupAssurance{
		Conclusion: conclusion,
		Confidence: confidence,
		Summary:    summary,
		Signals:    signals,
	}
}
