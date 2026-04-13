package triage

import (
	"sort"
	"strings"

	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/scoring"
)

var severityWeights = map[string]int{
	"CRITICAL": 90,
	"HIGH":     70,
	"MEDIUM":   40,
	"LOW":      20,
	"INFO":     10,
}

func Apply(findings []model.Finding) []model.Finding {
	out := append([]model.Finding(nil), findings...)
	for i := range out {
		decorate(&out[i])
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PriorityScore != out[j].PriorityScore {
			return out[i].PriorityScore > out[j].PriorityScore
		}
		if out[i].Penalty != out[j].Penalty {
			return out[i].Penalty > out[j].Penalty
		}
		if out[i].Severity != out[j].Severity {
			return severityRank(out[i].Severity) > severityRank(out[j].Severity)
		}
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		return out[i].ResourceID < out[j].ResourceID
	})

	for i := range out {
		out[i].Rank = i + 1
	}
	return out
}

func decorate(f *model.Finding) {
	if rule, ok := scoring.Default().Rule(f.ID); ok {
		if f.Title == "" {
			f.Title = rule.Title
		}
		if f.Domain == "" {
			f.Domain = rule.Domain
		}
		if f.Severity == "" {
			f.Severity = rule.Severity
		}
		if f.Penalty == 0 {
			f.Penalty = rule.BasePenalty
		}
	}
	if f.Impact == "" {
		f.Impact = impactFor(*f)
	}
	if f.OwnerHint == "" {
		f.OwnerHint = ownerHintFor(*f)
	}
	if f.Effort == "" {
		f.Effort = effortFor(*f)
	}
	f.PriorityScore = priorityScore(*f)
}

func severityRank(severity string) int {
	return severityWeights[strings.ToUpper(strings.TrimSpace(severity))]
}

func priorityScore(f model.Finding) int {
	score := severityRank(f.Severity) + f.Penalty
	switch strings.ToLower(strings.TrimSpace(f.Impact)) {
	case "restore block":
		score += 20
	case "data loss risk":
		score += 18
	case "coverage gap":
		score += 14
	case "access gap":
		score += 10
	case "degraded recovery":
		score += 8
	default:
		score += 5
	}

	switch strings.ToUpper(strings.TrimSpace(f.Effort)) {
	case "S":
		score += 8
	case "M":
		score += 4
	}

	switch f.Confidence {
	case model.EvidenceConfidenceConfirmed:
		score += 6
	case model.EvidenceConfidenceInferred:
		score += 2
	}
	return score
}

func impactFor(f model.Finding) string {
	id := strings.ToUpper(strings.TrimSpace(f.ID))
	switch {
	case strings.HasPrefix(id, "RESTORE_"), strings.HasPrefix(id, "PV_HOSTPATH"), strings.HasPrefix(id, "PVC_UNBOUND"):
		return "restore block"
	case strings.HasPrefix(id, "BACKUP_"), strings.HasPrefix(id, "ETCD_"):
		return "coverage gap"
	case strings.HasPrefix(id, "PV_"), strings.HasPrefix(id, "SC_"), strings.HasPrefix(id, "SNAPSHOT_"):
		return "data loss risk"
	case strings.HasPrefix(id, "RBAC_"):
		return "access gap"
	case strings.HasPrefix(id, "NETPOL_"), strings.HasPrefix(id, "CERT_"), strings.HasPrefix(id, "PSA_"):
		return "degraded recovery"
	default:
		switch strings.ToLower(strings.TrimSpace(f.Domain)) {
		case "backup":
			return "coverage gap"
		case "storage":
			return "data loss risk"
		case "config":
			return "degraded recovery"
		default:
			return "operational risk"
		}
	}
}

func ownerHintFor(f model.Finding) string {
	id := strings.ToUpper(strings.TrimSpace(f.ID))
	switch {
	case strings.HasPrefix(id, "BACKUP_"), strings.HasPrefix(id, "RESTORE_"), strings.HasPrefix(id, "ETCD_"):
		return "Platform / backup owner"
	case strings.HasPrefix(id, "PV_"), strings.HasPrefix(id, "PVC_"), strings.HasPrefix(id, "SC_"), strings.HasPrefix(id, "SNAPSHOT_"):
		return "Storage platform"
	case strings.HasPrefix(id, "RBAC_"), strings.HasPrefix(id, "PSA_"), strings.HasPrefix(id, "SA_"):
		return "Platform / security"
	case strings.HasPrefix(id, "NETPOL_"), strings.HasPrefix(id, "CERT_"), strings.HasPrefix(id, "CRD_"):
		return "Platform engineering"
	default:
		switch strings.ToLower(strings.TrimSpace(f.Domain)) {
		case "workload":
			return "Application owner"
		case "backup":
			return "Platform / backup owner"
		case "storage":
			return "Storage platform"
		default:
			return "Platform engineering"
		}
	}
}

func effortFor(f model.Finding) string {
	id := strings.ToUpper(strings.TrimSpace(f.ID))
	switch {
	case strings.HasPrefix(id, "RBAC_"), strings.HasPrefix(id, "NETPOL_"), strings.HasPrefix(id, "LR_"), strings.HasPrefix(id, "PSA_"), strings.HasPrefix(id, "CERT_"):
		return "S"
	case strings.HasPrefix(id, "PV_HOSTPATH"), strings.HasPrefix(id, "BACKUP_NONE"), strings.HasPrefix(id, "IMAGE_EXTERNAL_"):
		return "L"
	case strings.HasPrefix(id, "BACKUP_"), strings.HasPrefix(id, "SC_"), strings.HasPrefix(id, "PVC_"), strings.HasPrefix(id, "STS_"), strings.HasPrefix(id, "CRD_"), strings.HasPrefix(id, "HELM_"):
		return "M"
	default:
		if severityRank(f.Severity) >= severityWeights["HIGH"] {
			return "M"
		}
		return "S"
	}
}
