package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s-recovery-visualizer/internal/model"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type toolSpec struct {
	Name          string
	Namespaces    []string
	CRDGroupParts []string // substrings to match against CRD group names
	PodLabelKey   string
	PodLabelValue string
}

var knownTools = []toolSpec{
	{
		Name:          "kasten",
		Namespaces:    []string{"kasten-io"},
		CRDGroupParts: []string{"kio.kasten.io", "config.kio.kasten.io"},
		PodLabelKey:   "app",
		PodLabelValue: "k10",
	},
	{
		Name:          "velero",
		Namespaces:    []string{"velero"},
		CRDGroupParts: []string{"velero.io"},
		PodLabelKey:   "app.kubernetes.io/name",
		PodLabelValue: "velero",
	},
	{
		Name:          "rubrik",
		Namespaces:    []string{"rubrik", "rbs"},
		CRDGroupParts: []string{"rubrik.com"},
		PodLabelKey:   "app",
		PodLabelValue: "rubrik-backup-service",
	},
	{
		Name:          "longhorn",
		Namespaces:    []string{"longhorn-system"},
		CRDGroupParts: []string{"longhorn.io"},
		PodLabelKey:   "app",
		PodLabelValue: "longhorn-manager",
	},
	{
		Name:          "trilio",
		Namespaces:    []string{"trilio-system"},
		CRDGroupParts: []string{"triliovault.trilio.io"},
		PodLabelKey:   "app",
		PodLabelValue: "trilio",
	},
	{
		Name:          "stash",
		Namespaces:    []string{"stash"},
		CRDGroupParts: []string{"stash.appscode.com"},
		PodLabelKey:   "app",
		PodLabelValue: "stash",
	},
	{
		Name:          "cloudcasa",
		Namespaces:    []string{"cloudcasa-io"},
		CRDGroupParts: []string{"cloudcasa.io"},
		PodLabelKey:   "app",
		PodLabelValue: "cloudcasa",
	},
}

type inspectionResult struct {
	Tool     string
	Status   model.BackupCoverageStatus
	Reason   string
	Policies []model.BackupPolicy
}

type rankedInspection struct {
	index  int
	result inspectionResult
}

// Detect scans the cluster for known backup tools and populates b.Inventory.Backup.
func Detect(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle) {
	// Build quick lookup sets from already-collected data
	nsSet := map[string]struct{}{}
	for _, ns := range b.Inventory.Namespaces {
		nsSet[ns.Name] = struct{}{}
	}
	crdGroups := map[string]struct{}{}
	for _, crd := range b.Inventory.CRDs {
		crdGroups[crd.Group] = struct{}{}
	}

	inv := model.BackupInventory{
		PrimaryTool:    "none",
		Tools:          []model.BackupDetectedTool{},
		CoverageStatus: model.BackupCoverageStatusNotDetected,
	}

	for _, spec := range knownTools {
		tool := model.BackupDetectedTool{
			Name:     spec.Name,
			Detected: false,
		}

		// Check namespace presence
		foundNS := ""
		for _, ns := range spec.Namespaces {
			if _, ok := nsSet[ns]; ok {
				foundNS = ns
				tool.Detected = true
				tool.Namespace = ns
				break
			}
		}

		// Check CRD presence
		for group := range crdGroups {
			for _, part := range spec.CRDGroupParts {
				if strings.Contains(group, part) {
					tool.Detected = true
					tool.CRDsFound = append(tool.CRDsFound, group)
				}
			}
		}

		// If namespace found, check for pods to confirm and get version
		if foundNS != "" && spec.PodLabelKey != "" {
			selector := spec.PodLabelKey + "=" + spec.PodLabelValue
			pods, err := cs.CoreV1().Pods(foundNS).List(ctx, metav1.ListOptions{
				LabelSelector: selector,
				Limit:         1,
			})
			if err == nil && len(pods.Items) > 0 {
				pod := pods.Items[0]
				if v := pod.Labels["app.kubernetes.io/version"]; v != "" {
					tool.Version = v
				} else if v := pod.Labels["helm.sh/chart"]; v != "" {
					tool.Version = v
				}
			}
		}

		inv.Tools = append(inv.Tools, tool)
	}

	inspectDetectedTools(ctx, cs, b, &inv)

	b.Inventory.Backup = inv
}

// ── Policy collection ──────────────────────────────────────────────────────

// collectPolicies fetches backup policies/schedules for supported tools.
func collectPolicies(ctx context.Context, cs *kubernetes.Clientset, tool string) inspectionResult {
	switch tool {
	case "velero":
		return veleroSchedules(ctx, cs)
	case "kasten":
		return kastenPolicies(ctx, cs)
	case "longhorn":
		return longhornRecurringJobs(ctx, cs)
	default:
		return inspectionResult{
			Tool:   tool,
			Status: model.BackupCoverageStatusUnsupported,
			Reason: fmt.Sprintf("%s was detected, but this scanner does not yet inspect its policies or schedules.", tool),
		}
	}
}

// veleroSchedules reads velero.io/v1 Schedule objects.
func veleroSchedules(ctx context.Context, cs *kubernetes.Clientset) inspectionResult {
	raw, err := cs.RESTClient().
		Get().
		AbsPath("/apis/velero.io/v1/schedules").
		DoRaw(ctx)
	if err != nil {
		return inspectionError("velero", err, "Unable to inspect Velero schedules")
	}

	var list struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Spec struct {
				Schedule string `json:"schedule"`
				Template struct {
					IncludedNamespaces []string `json:"includedNamespaces"`
					ExcludedNamespaces []string `json:"excludedNamespaces"`
					TTL                string   `json:"ttl"`
					StorageLocation    string   `json:"storageLocation"`
				} `json:"template"`
			} `json:"spec"`
			Status struct {
				LastBackup string `json:"lastBackup"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return inspectionParseError("velero", err, "Unable to parse Velero schedules")
	}

	var policies []model.BackupPolicy
	for _, item := range list.Items {
		p := model.BackupPolicy{
			Tool:            "velero",
			Name:            item.Metadata.Name,
			PolicyNamespace: item.Metadata.Namespace,
			IncludedNS:      item.Spec.Template.IncludedNamespaces,
			ExcludedNS:      item.Spec.Template.ExcludedNamespaces,
			Schedule:        item.Spec.Schedule,
			RetentionTTL:    item.Spec.Template.TTL,
			RPOHours:        estimateRPOHours(item.Spec.Schedule),
			StorageLocation: item.Spec.Template.StorageLocation,
		}
		p.LastSuccessAt, p.LastSuccessAgeHours, p.FreshSchedule, p.Confidence = parseBackupSuccess(item.Status.LastBackup, p.RPOHours)
		if p.Confidence == "" {
			p.Confidence = model.EvidenceConfidenceInferred
		}
		// Non-default storage location is a strong offsite signal.
		loc := strings.ToLower(item.Spec.Template.StorageLocation)
		p.HasOffsite = loc != "" && loc != "default"
		policies = append(policies, p)
	}
	return inspectionVerified("velero", policies, "schedule")
}

// kastenPolicies reads config.kio.kasten.io/v1alpha1 Policy objects.
func kastenPolicies(ctx context.Context, cs *kubernetes.Clientset) inspectionResult {
	raw, err := cs.RESTClient().
		Get().
		AbsPath("/apis/config.kio.kasten.io/v1alpha1/policies").
		DoRaw(ctx)
	if err != nil {
		return inspectionError("kasten", err, "Unable to inspect Kasten policies")
	}

	var list struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Spec struct {
				Frequency string `json:"frequency"`
				Selector  struct {
					MatchNamespaces []string `json:"matchNamespaces"`
				} `json:"selector"`
				Actions []struct {
					Action string `json:"action"`
				} `json:"actions"`
				RetentionDays int `json:"retentionDays"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return inspectionParseError("kasten", err, "Unable to parse Kasten policies")
	}

	var policies []model.BackupPolicy
	for _, item := range list.Items {
		hasExport := false
		for _, a := range item.Spec.Actions {
			if strings.EqualFold(a.Action, "export") {
				hasExport = true
			}
		}
		retention := ""
		if item.Spec.RetentionDays > 0 {
			retention = strconv.Itoa(item.Spec.RetentionDays) + "d"
		}
		p := model.BackupPolicy{
			Tool:            "kasten",
			Name:            item.Metadata.Name,
			PolicyNamespace: item.Metadata.Namespace,
			IncludedNS:      item.Spec.Selector.MatchNamespaces,
			Schedule:        item.Spec.Frequency,
			RetentionTTL:    retention,
			RPOHours:        estimateRPOHours(item.Spec.Frequency),
			HasOffsite:      hasExport,
			Confidence:      model.EvidenceConfidenceInferred,
		}
		policies = append(policies, p)
	}
	return inspectionVerified("kasten", policies, "policy")
}

// longhornRecurringJobs reads longhorn.io/v1beta2 RecurringJob objects.
// It also checks whether a BackupTarget is configured (offsite signal).
func longhornRecurringJobs(ctx context.Context, cs *kubernetes.Clientset) inspectionResult {
	// Check BackupTarget setting — non-empty = offsite configured.
	hasOffsiteTarget := longhornBackupTargetSet(ctx, cs)

	raw, err := cs.RESTClient().
		Get().
		AbsPath("/apis/longhorn.io/v1beta2/namespaces/longhorn-system/recurringjobs").
		DoRaw(ctx)
	if err != nil {
		// Try v1beta1
		raw, err = cs.RESTClient().
			Get().
			AbsPath("/apis/longhorn.io/v1beta1/namespaces/longhorn-system/recurringjobs").
			DoRaw(ctx)
		if err != nil {
			return inspectionError("longhorn", err, "Unable to inspect Longhorn recurring jobs")
		}
	}

	var list struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Spec struct {
				Task   string   `json:"task"` // "backup" or "snapshot"
				Cron   string   `json:"cron"`
				Retain int      `json:"retain"`
				Groups []string `json:"groups"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return inspectionParseError("longhorn", err, "Unable to parse Longhorn recurring jobs")
	}

	var policies []model.BackupPolicy
	for _, item := range list.Items {
		if !strings.EqualFold(item.Spec.Task, "backup") {
			continue // skip snapshot-only jobs
		}
		retention := ""
		if item.Spec.Retain > 0 {
			retention = strconv.Itoa(item.Spec.Retain) + " snapshots"
		}
		p := model.BackupPolicy{
			Tool:         "longhorn",
			Name:         item.Metadata.Name,
			Schedule:     item.Spec.Cron,
			RetentionTTL: retention,
			RPOHours:     estimateRPOHours(item.Spec.Cron),
			HasOffsite:   hasOffsiteTarget,
			Confidence:   model.EvidenceConfidenceInferred,
		}
		policies = append(policies, p)
	}
	return inspectionVerified("longhorn", policies, "recurring job")
}

// longhornBackupTargetSet checks if Longhorn has a non-empty BackupTarget setting.
func longhornBackupTargetSet(ctx context.Context, cs *kubernetes.Clientset) bool {
	for _, apiVer := range []string{"v1beta2", "v1beta1"} {
		raw, err := cs.RESTClient().
			Get().
			AbsPath("/apis/longhorn.io/" + apiVer + "/namespaces/longhorn-system/settings/backup-target").
			DoRaw(ctx)
		if err != nil {
			continue
		}
		var setting struct {
			Value string `json:"value"`
		}
		if err := json.Unmarshal(raw, &setting); err != nil {
			continue
		}
		return strings.TrimSpace(setting.Value) != ""
	}
	return false
}

// ── RPO estimation ─────────────────────────────────────────────────────────

// estimateRPOHours converts a cron expression or Kasten frequency label into
// an estimated RPO in hours. Returns -1 when the schedule cannot be parsed.
func estimateRPOHours(schedule string) int {
	schedule = strings.ToLower(strings.TrimSpace(schedule))
	if schedule == "" {
		return -1
	}
	// Named schedules
	switch schedule {
	case "@hourly":
		return 1
	case "@daily", "@midnight":
		return 24
	case "@weekly":
		return 168
	case "@monthly":
		return 720
	}
	// Kasten frequency labels
	switch {
	case strings.Contains(schedule, "hourly"):
		return 1
	case strings.Contains(schedule, "daily"):
		return 24
	case strings.Contains(schedule, "weekly"):
		return 168
	case strings.Contains(schedule, "monthly"):
		return 720
	}
	// 5-field cron: minute hour dom month dow
	parts := strings.Fields(schedule)
	if len(parts) != 5 {
		return -1
	}
	hour := parts[1]
	// */N → every N hours
	if strings.HasPrefix(hour, "*/") {
		if n, err := strconv.Atoi(hour[2:]); err == nil && n > 0 {
			return n
		}
	}
	// * in hour field → runs every hour
	if hour == "*" {
		return 1
	}
	// Specific hour value
	if _, err := strconv.Atoi(hour); err == nil {
		dow := parts[4]
		dom := parts[2]
		if dow != "*" {
			return 168 // weekly
		}
		if dom != "*" {
			return 720 // monthly-ish
		}
		return 24 // daily
	}
	return -1
}

// ── Coverage helpers ───────────────────────────────────────────────────────

func coveredNamespacesFromPolicies(b *model.Bundle, policies []model.BackupPolicy) []string {
	if len(policies) == 0 {
		return nil
	}
	var covered []string
	for _, ns := range b.Inventory.Namespaces {
		if policyListCoversNamespace(policies, ns.Name) {
			covered = append(covered, ns.Name)
		}
	}
	return covered
}

func uncoveredStatefulNamespacesFromPolicies(b *model.Bundle, policies []model.BackupPolicy) []string {
	seen := map[string]struct{}{}
	var uncovered []string
	for _, sts := range b.Inventory.StatefulSets {
		if policyListCoversNamespace(policies, sts.Namespace) {
			continue
		}
		if _, already := seen[sts.Namespace]; already {
			continue
		}
		uncovered = append(uncovered, sts.Namespace)
		seen[sts.Namespace] = struct{}{}
	}
	return uncovered
}

func policyListCoversNamespace(policies []model.BackupPolicy, ns string) bool {
	for _, p := range policies {
		if policyCoversNamespace(p, ns) {
			return true
		}
	}
	return false
}

func policyCoversNamespace(p model.BackupPolicy, ns string) bool {
	for _, ex := range p.ExcludedNS {
		if ex == ns {
			return false
		}
	}
	if len(p.IncludedNS) == 0 {
		return true
	}
	for _, incl := range p.IncludedNS {
		if incl == ns || incl == "*" {
			return true
		}
	}
	return false
}

func inspectDetectedTools(ctx context.Context, cs *kubernetes.Clientset, b *model.Bundle, inv *model.BackupInventory) {
	var detected []rankedInspection
	for i := range inv.Tools {
		if !inv.Tools[i].Detected {
			continue
		}
		result := collectPolicies(ctx, cs, inv.Tools[i].Name)
		inv.Tools[i].PolicyInspectionStatus = result.Status
		inv.Tools[i].PolicyInspectionDetail = result.Reason
		detected = append(detected, rankedInspection{index: i, result: result})
	}

	applyInspectionResults(b, inv, detected)
}

func applyInspectionResults(b *model.Bundle, inv *model.BackupInventory, detected []rankedInspection) {
	if len(detected) == 0 {
		inv.PrimaryTool = "none"
		inv.CoverageStatus = model.BackupCoverageStatusNotDetected
		inv.CoverageReason = "No backup tool detected."
		return
	}

	var verified []inspectionResult
	sourceSet := map[string]struct{}{}
	for _, item := range detected {
		if item.result.Status != model.BackupCoverageStatusVerified {
			continue
		}
		verified = append(verified, item.result)
		if _, seen := sourceSet[item.result.Tool]; !seen {
			inv.CoverageSourceTools = append(inv.CoverageSourceTools, item.result.Tool)
			sourceSet[item.result.Tool] = struct{}{}
		}
	}

	if len(verified) > 0 {
		inv.PrimaryTool = verified[0].Tool
		inv.CoverageVerified = true
		inv.CoverageStatus = model.BackupCoverageStatusVerified
		inv.CoverageReason = verifiedCoverageReason(inv.CoverageSourceTools)
		for _, item := range verified {
			inv.Policies = append(inv.Policies, item.Policies...)
		}
		inv.CoveredNamespaces = coveredNamespacesFromPolicies(b, inv.Policies)
		inv.UncoveredStatefulNS = uncoveredStatefulNamespacesFromPolicies(b, inv.Policies)
		for _, p := range inv.Policies {
			if p.HasOffsite {
				inv.HasOffsite = true
				break
			}
		}
		return
	}

	best := detected[0].result
	for _, item := range detected[1:] {
		if inspectionPriority(item.result.Status) > inspectionPriority(best.Status) {
			best = item.result
		}
	}
	inv.PrimaryTool = best.Tool
	inv.CoverageStatus = best.Status
	inv.CoverageReason = best.Reason
}

func inspectionPriority(status model.BackupCoverageStatus) int {
	switch status {
	case model.BackupCoverageStatusVerified:
		return 5
	case model.BackupCoverageStatusPermissionDenied:
		return 4
	case model.BackupCoverageStatusParseError:
		return 3
	case model.BackupCoverageStatusAPIError:
		return 2
	case model.BackupCoverageStatusUnsupported:
		return 1
	default:
		return 0
	}
}

func inspectionVerified(tool string, policies []model.BackupPolicy, objectLabel string) inspectionResult {
	reason := fmt.Sprintf("Parsed %d %s %s.", len(policies), tool, pluralize(objectLabel, len(policies)))
	if len(policies) == 0 {
		reason = fmt.Sprintf("%s policy inspection succeeded, but no %s were found.", tool, pluralize(objectLabel, 2))
	}
	return inspectionResult{
		Tool:     tool,
		Status:   model.BackupCoverageStatusVerified,
		Reason:   reason,
		Policies: policies,
	}
}

func inspectionError(tool string, err error, prefix string) inspectionResult {
	status := model.BackupCoverageStatusAPIError
	if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) || looksLikePermissionError(err) {
		status = model.BackupCoverageStatusPermissionDenied
	}
	return inspectionResult{
		Tool:   tool,
		Status: status,
		Reason: fmt.Sprintf("%s: %v", prefix, err),
	}
}

func inspectionParseError(tool string, err error, prefix string) inspectionResult {
	return inspectionResult{
		Tool:   tool,
		Status: model.BackupCoverageStatusParseError,
		Reason: fmt.Sprintf("%s: %v", prefix, err),
	}
}

func looksLikePermissionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "unauthorized")
}

func verifiedCoverageReason(tools []string) string {
	if len(tools) == 0 {
		return "Policy coverage was verified."
	}
	if len(tools) == 1 {
		return fmt.Sprintf("Policy coverage verified from %s.", tools[0])
	}
	return fmt.Sprintf("Policy coverage verified from multiple tools: %s.", strings.Join(tools, ", "))
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	if strings.HasSuffix(word, "y") {
		return strings.TrimSuffix(word, "y") + "ies"
	}
	return word + "s"
}

func parseBackupSuccess(lastBackup string, rpoHours int) (string, int, bool, model.EvidenceConfidence) {
	lastBackup = strings.TrimSpace(lastBackup)
	if lastBackup == "" {
		return "", 0, false, model.EvidenceConfidenceUnknown
	}
	ts, err := time.Parse(time.RFC3339, lastBackup)
	if err != nil {
		return lastBackup, 0, false, model.EvidenceConfidenceUnknown
	}
	ageHours := int(time.Since(ts).Hours())
	freshThreshold := 48
	if rpoHours > 0 && rpoHours*2 > freshThreshold {
		freshThreshold = rpoHours * 2
	}
	return ts.UTC().Format(time.RFC3339), ageHours, ageHours <= freshThreshold, model.EvidenceConfidenceConfirmed
}
