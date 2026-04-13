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
		minStorageScore           = fs.Int("min-storage-score", 0, "Fail if storage score is below this value. 0 disables.")
		minWorkloadScore          = fs.Int("min-workload-score", 0, "Fail if workload score is below this value. 0 disables.")
		minConfigScore            = fs.Int("min-config-score", 0, "Fail if config score is below this value. 0 disables.")
		minBackupScore            = fs.Int("min-backup-score", 0, "Fail if backup score is below this value. 0 disables.")
		maxRisk                   = fs.String("max-risk", "MODERATE", "Highest allowed risk posture: LOW|MODERATE|HIGH|CRITICAL")
		maxDrop                   = fs.Float64("max-drop", 0, "Max allowed score drop vs previous run (points). 0 disables.")
		maxDropPct                = fs.Float64("max-drop-pct", 0, "Max allowed score drop vs previous run (percent). 0 disables.")
		maxCriticalFindings       = fs.Int("max-critical-findings", -1, "Fail if the current bundle has more than this many critical findings. -1 disables.")
		maxHighFindings           = fs.Int("max-high-findings", -1, "Fail if the current bundle has more than this many high findings. -1 disables.")
		maxNewFindings            = fs.Int("max-new-findings", -1, "Fail if the current bundle introduces more than this many new findings compared with the previous bundle. -1 disables.")
		maxRegressedFindings      = fs.Int("max-regressed-findings", -1, "Fail if the current bundle worsens severity for more than this many existing findings. -1 disables.")
		failOnNewCritical         = fs.Bool("fail-on-new-critical", false, "Fail if new critical findings appear compared with the previous scan bundle")
		failOnUncoveredStateful   = fs.Bool("fail-on-uncovered-stateful", false, "Fail if verified backup coverage still leaves uncovered stateful namespaces")
		failOnOffsiteLoss         = fs.Bool("fail-on-offsite-loss", false, "Fail if previous scan had offsite backup evidence and current scan does not")
		failOnMissingBackupPolicy = fs.Bool("fail-on-missing-backup-policies", false, "Fail if no backup policies or schedules are present")
	)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintln(stdout, "Usage of check:")
		fmt.Fprintln(stdout, "  go run ./cmd/check --current ./out/recovery-scan.json [flags]")
		fmt.Fprintln(stdout, "  go run ./cmd/check --in ./out/recovery-enriched.json [flags]  # legacy input path")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Validate scan bundles against score floors, domain thresholds, finding budgets, and regression budgets.")
		fmt.Fprintln(stdout)
		originalOutput := fs.Output()
		fs.SetOutput(stdout)
		fs.PrintDefaults()
		fs.SetOutput(originalOutput)
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
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
			MinStorageScore:           *minStorageScore,
			MinWorkloadScore:          *minWorkloadScore,
			MinConfigScore:            *minConfigScore,
			MinBackupScore:            *minBackupScore,
			MaxRisk:                   *maxRisk,
			MaxDrop:                   *maxDrop,
			MaxDropPct:                *maxDropPct,
			MaxCriticalFindings:       *maxCriticalFindings,
			MaxHighFindings:           *maxHighFindings,
			MaxNewFindings:            *maxNewFindings,
			MaxRegressedFindings:      *maxRegressedFindings,
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
	for _, skipped := range []struct {
		enabled bool
		id      string
		message string
	}{
		{*minStorageScore > 0, "storage-score", "legacy enriched input does not expose storage-score gating"},
		{*minWorkloadScore > 0, "workload-score", "legacy enriched input does not expose workload-score gating"},
		{*minConfigScore > 0, "config-score", "legacy enriched input does not expose config-score gating"},
		{*minBackupScore > 0, "backup-score", "legacy enriched input does not expose backup-score gating"},
		{*maxCriticalFindings >= 0, "critical-finding-budget", "legacy enriched input does not expose finding severity counts"},
		{*maxHighFindings >= 0, "high-finding-budget", "legacy enriched input does not expose finding severity counts"},
		{*maxNewFindings >= 0, "new-finding-budget", "legacy enriched input does not expose new-finding comparisons"},
		{*maxRegressedFindings >= 0, "regressed-finding-budget", "legacy enriched input does not expose severity regression comparisons"},
	} {
		if skipped.enabled {
			legacyEval.Results = append(legacyEval.Results, gates.Result{ID: skipped.id, Status: gates.StatusSkip, Message: skipped.message})
		}
	}
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
