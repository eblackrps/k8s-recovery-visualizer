package output

import (
	"strings"

	"k8s-recovery-visualizer/internal/model"
)

func backupCoverageStatusText(inv model.BackupInventory) string {
	switch inv.CoverageStatus {
	case model.BackupCoverageStatusVerified:
		return "verified"
	case model.BackupCoverageStatusUnsupported:
		return "unsupported"
	case model.BackupCoverageStatusPermissionDenied:
		return "permission denied"
	case model.BackupCoverageStatusParseError:
		return "parse error"
	case model.BackupCoverageStatusAPIError:
		return "api error"
	case model.BackupCoverageStatusNotDetected:
		return "not detected"
	default:
		return "unknown"
	}
}

func backupCoverageReasonText(inv model.BackupInventory) string {
	if inv.CoverageReason != "" {
		return inv.CoverageReason
	}
	switch inv.CoverageStatus {
	case model.BackupCoverageStatusVerified:
		return "Policy coverage was verified."
	case model.BackupCoverageStatusUnsupported:
		return "The detected backup tool is not yet supported for policy inspection."
	case model.BackupCoverageStatusPermissionDenied:
		return "The scan could not read backup policy objects with the current credentials."
	case model.BackupCoverageStatusParseError:
		return "The scan reached the backup API, but could not parse the returned objects."
	case model.BackupCoverageStatusAPIError:
		return "The scan could not query the backup product API successfully."
	default:
		return ""
	}
}

func backupInspectionStatusText(status model.BackupCoverageStatus) string {
	switch status {
	case model.BackupCoverageStatusVerified:
		return "verified"
	case model.BackupCoverageStatusUnsupported:
		return "unsupported"
	case model.BackupCoverageStatusPermissionDenied:
		return "permission denied"
	case model.BackupCoverageStatusParseError:
		return "parse error"
	case model.BackupCoverageStatusAPIError:
		return "api error"
	case model.BackupCoverageStatusNotDetected:
		return "not detected"
	default:
		return "not inspected"
	}
}

func backupAssuranceConclusionText(assurance *model.BackupAssurance) string {
	if assurance == nil {
		return "not assessed"
	}
	switch assurance.Conclusion {
	case model.BackupAssuranceEvidenceConfirmed:
		return "evidence confirmed"
	case model.BackupAssuranceEvidenceInferred:
		return "evidence inferred"
	case model.BackupAssuranceCoverageGap:
		return "coverage gap"
	case model.BackupAssuranceUnverified:
		return "unverified"
	case model.BackupAssuranceAtRisk:
		return "at risk"
	default:
		return string(assurance.Conclusion)
	}
}

func backupAssuranceColor(assurance *model.BackupAssurance) string {
	if assurance == nil {
		return "#8b949e"
	}
	switch assurance.Conclusion {
	case model.BackupAssuranceEvidenceConfirmed:
		return "#7ee787"
	case model.BackupAssuranceEvidenceInferred:
		return "#f2cc60"
	case model.BackupAssuranceCoverageGap, model.BackupAssuranceUnverified:
		return "#ffa657"
	case model.BackupAssuranceAtRisk:
		return "#f85149"
	default:
		return "#8b949e"
	}
}

func backupOffsiteDetailText(inv model.BackupInventory) string {
	switch {
	case len(inv.OffsiteMissingNS) > 0:
		return "missing for namespaces: " + strings.Join(inv.OffsiteMissingNS, ", ")
	case inv.HasOffsite:
		return "all covered namespaces have offsite evidence"
	case inv.PrimaryTool == "" || inv.PrimaryTool == "none":
		return "not applicable"
	default:
		return "no offsite evidence"
	}
}
