package output

import "k8s-recovery-visualizer/internal/model"

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
