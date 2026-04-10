package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"k8s-recovery-visualizer/internal/gates"
	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/risk"
)

type legacyTrend struct {
	DeltaScore   float64 `json:"deltaScore"`
	DeltaPercent float64 `json:"deltaPercent"`
}

type legacyEnriched struct {
	SchemaVersion string       `json:"schemaVersion"`
	Trend         *legacyTrend `json:"trend,omitempty"`
	Risk          struct {
		Score    float64      `json:"score"`
		Maturity string       `json:"maturity"`
		Posture  risk.Posture `json:"posture"`
	} `json:"risk"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.ReadFile))
}

func run(args []string, stdout io.Writer, readFile func(string) ([]byte, error)) int {
	var (
		fs                        = flag.NewFlagSet("check", flag.ContinueOnError)
		currentPath               = fs.String("current", "", "Path to recovery-scan.json (preferred)")
		previousPath              = fs.String("previous", "", "Path to previous recovery-scan.json for regression checks")
		legacyInput               = fs.String("in", "out/recovery-enriched.json", "Legacy path to recovery-enriched.json or recovery-scan.json")
		format                    = fs.String("format", "text", "Output format: text|json")
		minScore                  = fs.Int("min-score", 0, "Fail if overall score is below this value. 0 disables.")
		maxRisk                   = fs.String("max-risk", "MODERATE", "Highest allowed risk posture: LOW|MODERATE|HIGH|CRITICAL")
		maxDrop                   = fs.Float64("max-drop", 0, "Max allowed score drop vs previous run (points). 0 disables.")
		maxDropPct                = fs.Float64("max-drop-pct", 0, "Max allowed score drop vs previous run (percent). 0 disables.")
		failOnNewCritical         = fs.Bool("fail-on-new-critical", false, "Fail if new critical findings appear compared with the previous scan bundle")
		failOnUncoveredStateful   = fs.Bool("fail-on-uncovered-stateful", false, "Fail if verified backup coverage still leaves uncovered stateful namespaces")
		failOnOffsiteLoss         = fs.Bool("fail-on-offsite-loss", false, "Fail if previous scan had offsite backup evidence and current scan does not")
		failOnMissingBackupPolicy = fs.Bool("fail-on-missing-backup-policies", false, "Fail if no backup policies or schedules are present")
	)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stdout, "CHECK FAIL: %v\n", err)
		return 2
	}

	if strings.TrimSpace(*currentPath) == "" {
		*currentPath = *legacyInput
	}

	currentRaw, err := readFile(*currentPath)
	if err != nil {
		fmt.Fprintf(stdout, "CHECK FAIL: cannot read %s (%v)\n", *currentPath, err)
		return 2
	}

	var currentBundle model.Bundle
	if err := json.Unmarshal(currentRaw, &currentBundle); err == nil && currentBundle.SchemaVersion != "" && currentBundle.Tool.Name != "" {
		var previousBundle *model.Bundle
		if strings.TrimSpace(*previousPath) != "" {
			previousRaw, err := readFile(*previousPath)
			if err != nil {
				fmt.Fprintf(stdout, "CHECK FAIL: cannot read %s (%v)\n", *previousPath, err)
				return 2
			}
			var decoded model.Bundle
			if err := json.Unmarshal(previousRaw, &decoded); err != nil {
				fmt.Fprintf(stdout, "CHECK FAIL: invalid JSON in %s (%v)\n", *previousPath, err)
				return 2
			}
			previousBundle = &decoded
		}

		eval := gates.Evaluate(&currentBundle, previousBundle, gates.Policy{
			MinScore:                  *minScore,
			MaxRisk:                   *maxRisk,
			MaxDrop:                   *maxDrop,
			MaxDropPct:                *maxDropPct,
			FailOnNewCritical:         *failOnNewCritical,
			FailOnUncoveredStateful:   *failOnUncoveredStateful,
			FailOnOffsiteLoss:         *failOnOffsiteLoss,
			FailOnMissingBackupPolicy: *failOnMissingBackupPolicy,
		})
		emitResult(stdout, eval, *format)
		if eval.Status == gates.StatusFail {
			return 1
		}
		return 0
	}

	var legacy legacyEnriched
	if err := json.Unmarshal(currentRaw, &legacy); err != nil {
		fmt.Fprintf(stdout, "CHECK FAIL: invalid JSON in %s (%v)\n", *currentPath, err)
		return 2
	}
	legacyEval := evaluateLegacy(legacy, *maxRisk, *maxDrop, *maxDropPct)
	emitResult(stdout, legacyEval, *format)
	if legacyEval.Status == gates.StatusFail {
		return 1
	}
	return 0
}

func evaluateLegacy(en legacyEnriched, maxRisk string, maxDrop, maxDropPct float64) gates.Evaluation {
	eval := gates.Evaluation{
		Status:          gates.StatusPass,
		CurrentScore:    int(en.Risk.Score),
		CurrentMaturity: en.Risk.Maturity,
	}

	add := func(id string, status gates.Status, message string) {
		eval.Results = append(eval.Results, gates.Result{ID: id, Status: status, Message: message})
		if status == gates.StatusFail {
			eval.Status = gates.StatusFail
		}
	}

	if postureRank(string(en.Risk.Posture)) > postureRank(maxRisk) {
		add("risk-posture", gates.StatusFail, fmt.Sprintf("risk posture %s exceeds allowed %s", strings.ToUpper(string(en.Risk.Posture)), strings.ToUpper(maxRisk)))
	} else {
		add("risk-posture", gates.StatusPass, fmt.Sprintf("risk posture %s within allowed %s", strings.ToUpper(string(en.Risk.Posture)), strings.ToUpper(maxRisk)))
	}

	if en.Trend == nil {
		if maxDrop > 0 {
			add("score-drop", gates.StatusSkip, "no trend available (first run or history missing)")
		}
		if maxDropPct > 0 {
			add("score-drop-pct", gates.StatusSkip, "no trend available (first run or history missing)")
		}
		return eval
	}

	drop := zeroSmall(-en.Trend.DeltaScore)
	dropPct := zeroSmall(-en.Trend.DeltaPercent)
	if maxDrop > 0 {
		if drop > maxDrop {
			add("score-drop", gates.StatusFail, fmt.Sprintf("regression %.2f points exceeds max-drop %.2f", drop, maxDrop))
		} else {
			add("score-drop", gates.StatusPass, fmt.Sprintf("regression %.2f points within max-drop %.2f", drop, maxDrop))
		}
	}
	if maxDropPct > 0 {
		if dropPct > maxDropPct {
			add("score-drop-pct", gates.StatusFail, fmt.Sprintf("regression %.2f%% exceeds max-drop-pct %.2f%%", dropPct, maxDropPct))
		} else {
			add("score-drop-pct", gates.StatusPass, fmt.Sprintf("regression %.2f%% within max-drop-pct %.2f%%", dropPct, maxDropPct))
		}
	}
	return eval
}

func emitResult(stdout io.Writer, eval gates.Evaluation, format string) {
	if strings.EqualFold(format, "json") {
		raw, _ := json.MarshalIndent(eval, "", "  ")
		fmt.Fprintln(stdout, string(raw))
		return
	}
	for _, result := range eval.Results {
		fmt.Fprintf(stdout, "%s: %s\n", result.Status, result.Message)
	}
	fmt.Fprintf(stdout, "CHECK %s: score=%d maturity=%s\n", eval.Status, eval.CurrentScore, eval.CurrentMaturity)
}

func zeroSmall(v float64) float64 {
	if v > -0.0000001 && v < 0.0000001 {
		return 0
	}
	return v
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
