package gates

import (
	"fmt"
	"strings"

	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/risk"
)

type Status string

const (
	StatusPass Status = "PASS"
	StatusFail Status = "FAIL"
	StatusSkip Status = "SKIP"
)

type Policy struct {
	MinScore                  int
	MaxRisk                   string
	MaxDrop                   float64
	MaxDropPct                float64
	FailOnNewCritical         bool
	FailOnUncoveredStateful   bool
	FailOnOffsiteLoss         bool
	FailOnMissingBackupPolicy bool
}

type Result struct {
	ID      string `json:"id"`
	Status  Status `json:"status"`
	Message string `json:"message"`
}

type Evaluation struct {
	Status          Status   `json:"status"`
	CurrentScore    int      `json:"currentScore"`
	CurrentMaturity string   `json:"currentMaturity"`
	Results         []Result `json:"results"`
}

func Evaluate(current, previous *model.Bundle, policy Policy) Evaluation {
	eval := Evaluation{
		Status:          StatusPass,
		CurrentScore:    current.Score.Overall.Final,
		CurrentMaturity: current.Score.Maturity,
	}

	add := func(id string, status Status, message string) {
		eval.Results = append(eval.Results, Result{ID: id, Status: status, Message: message})
		if status == StatusFail {
			eval.Status = StatusFail
		}
	}

	if policy.MinScore > 0 {
		if current.Score.Overall.Final < policy.MinScore {
			add("overall-score", StatusFail, fmt.Sprintf("overall score %d is below minimum %d", current.Score.Overall.Final, policy.MinScore))
		} else {
			add("overall-score", StatusPass, fmt.Sprintf("overall score %d meets minimum %d", current.Score.Overall.Final, policy.MinScore))
		}
	}

	if strings.TrimSpace(policy.MaxRisk) != "" {
		actual := risk.FromScore(float64(current.Score.Overall.Final), current.Score.Maturity).Posture
		if postureRank(string(actual)) > postureRank(policy.MaxRisk) {
			add("risk-posture", StatusFail, fmt.Sprintf("risk posture %s exceeds allowed %s", actual, strings.ToUpper(strings.TrimSpace(policy.MaxRisk))))
		} else {
			add("risk-posture", StatusPass, fmt.Sprintf("risk posture %s is within allowed %s", actual, strings.ToUpper(strings.TrimSpace(policy.MaxRisk))))
		}
	}

	if policy.MaxDrop > 0 || policy.MaxDropPct > 0 || policy.FailOnNewCritical || policy.FailOnOffsiteLoss {
		if previous == nil {
			if policy.MaxDrop > 0 {
				add("score-drop", StatusSkip, "no previous bundle supplied; score regression gate skipped")
			}
			if policy.MaxDropPct > 0 {
				add("score-drop-pct", StatusSkip, "no previous bundle supplied; percentage regression gate skipped")
			}
			if policy.FailOnNewCritical {
				add("new-critical-findings", StatusSkip, "no previous bundle supplied; new critical finding gate skipped")
			}
			if policy.FailOnOffsiteLoss {
				add("offsite-loss", StatusSkip, "no previous bundle supplied; offsite loss gate skipped")
			}
		} else {
			prevScore := float64(previous.Score.Overall.Final)
			currScore := float64(current.Score.Overall.Final)
			drop := prevScore - currScore
			if policy.MaxDrop > 0 {
				if drop > policy.MaxDrop {
					add("score-drop", StatusFail, fmt.Sprintf("score regressed by %.2f points (max %.2f)", drop, policy.MaxDrop))
				} else {
					add("score-drop", StatusPass, fmt.Sprintf("score regression %.2f points is within max %.2f", drop, policy.MaxDrop))
				}
			}
			if policy.MaxDropPct > 0 {
				dropPct := 0.0
				if prevScore > 0 {
					dropPct = (drop / prevScore) * 100
				}
				if dropPct > policy.MaxDropPct {
					add("score-drop-pct", StatusFail, fmt.Sprintf("score regressed by %.2f%% (max %.2f%%)", dropPct, policy.MaxDropPct))
				} else {
					add("score-drop-pct", StatusPass, fmt.Sprintf("score regression %.2f%% is within max %.2f%%", dropPct, policy.MaxDropPct))
				}
			}
			if policy.FailOnNewCritical {
				newCritical := newCriticalFindings(previous.Inventory.Findings, current.Inventory.Findings)
				if len(newCritical) > 0 {
					add("new-critical-findings", StatusFail, fmt.Sprintf("new critical findings detected: %s", strings.Join(newCritical, ", ")))
				} else {
					add("new-critical-findings", StatusPass, "no new critical findings were introduced")
				}
			}
			if policy.FailOnOffsiteLoss {
				if previous.Inventory.Backup.HasOffsite && !current.Inventory.Backup.HasOffsite {
					add("offsite-loss", StatusFail, "offsite backup evidence was present previously and is now missing")
				} else {
					add("offsite-loss", StatusPass, "offsite backup posture did not regress")
				}
			}
		}
	}

	if policy.FailOnUncoveredStateful {
		if len(current.Inventory.Backup.UncoveredStatefulNS) > 0 {
			add("uncovered-stateful", StatusFail, fmt.Sprintf("uncovered stateful namespaces: %s", strings.Join(current.Inventory.Backup.UncoveredStatefulNS, ", ")))
		} else {
			add("uncovered-stateful", StatusPass, "all detected stateful namespaces are covered by verified backup scope")
		}
	}

	if policy.FailOnMissingBackupPolicy {
		switch {
		case current.Inventory.Backup.PrimaryTool == "" || current.Inventory.Backup.PrimaryTool == "none":
			add("missing-backup-policies", StatusFail, "no backup tool detected")
		case len(current.Inventory.Backup.Policies) == 0:
			add("missing-backup-policies", StatusFail, "backup tool detected but no policies or schedules were found")
		default:
			add("missing-backup-policies", StatusPass, fmt.Sprintf("backup policies present: %d", len(current.Inventory.Backup.Policies)))
		}
	}

	return eval
}

func postureRank(p string) int {
	switch strings.ToUpper(strings.TrimSpace(p)) {
	case "LOW":
		return 0
	case "MODERATE":
		return 1
	case "HIGH":
		return 2
	case "CRITICAL":
		return 3
	default:
		return 99
	}
}

func newCriticalFindings(previous, current []model.Finding) []string {
	prevSet := map[string]struct{}{}
	for _, finding := range previous {
		if finding.Severity != "CRITICAL" {
			continue
		}
		prevSet[finding.ID+"|"+finding.ResourceID] = struct{}{}
	}

	var out []string
	for _, finding := range current {
		if finding.Severity != "CRITICAL" {
			continue
		}
		key := finding.ID + "|" + finding.ResourceID
		if _, ok := prevSet[key]; ok {
			continue
		}
		out = append(out, finding.ID)
	}
	return out
}
