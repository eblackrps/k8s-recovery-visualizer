package backup

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s-recovery-visualizer/internal/model"
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
		PrimaryTool: "none",
		Tools:       []model.BackupDetectedTool{},
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

		if tool.Detected && inv.PrimaryTool == "none" {
			inv.PrimaryTool = spec.Name
		}
	}

	// Determine which namespaces with StatefulSets are not covered.
	if inv.PrimaryTool != "none" {
		inv.CoveredNamespaces = coveredNamespaces(ctx, cs, inv.PrimaryTool, b)
		inv.UncoveredStatefulNS = uncoveredStatefulNamespaces(b, inv.CoveredNamespaces)

		// Collect detailed backup policies (Velero, Kasten, Longhorn).
		inv.Policies = collectPolicies(ctx, cs, inv.PrimaryTool)
		for _, p := range inv.Policies {
			if p.HasOffsite {
				inv.HasOffsite = true
				break
			}
		}
	}

	b.Inventory.Backup = inv
}

// ── Policy collection ──────────────────────────────────────────────────────

// collectPolicies fetches backup policies/schedules for supported tools.
func collectPolicies(ctx context.Context, cs *kubernetes.Clientset, tool string) []model.BackupPolicy {
	switch tool {
	case "velero":
		return veleroSchedules(ctx, cs)
	case "kasten":
		return kastenPolicies(ctx, cs)
	case "longhorn":
		return longhornRecurringJobs(ctx, cs)
	default:
		return nil
	}
}

// veleroSchedules reads velero.io/v1 Schedule objects.
func veleroSchedules(ctx context.Context, cs *kubernetes.Clientset) []model.BackupPolicy {
	raw, err := cs.RESTClient().
		Get().
		AbsPath("/apis/velero.io/v1/schedules").
		DoRaw(ctx)
	if err != nil {
		return nil
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
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil
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
		// Non-default storage location is a strong offsite signal.
		loc := strings.ToLower(item.Spec.Template.StorageLocation)
		p.HasOffsite = loc != "" && loc != "default"
		policies = append(policies, p)
	}
	return policies
}

// kastenPolicies reads config.kio.kasten.io/v1alpha1 Policy objects.
func kastenPolicies(ctx context.Context, cs *kubernetes.Clientset) []model.BackupPolicy {
	raw, err := cs.RESTClient().
		Get().
		AbsPath("/apis/config.kio.kasten.io/v1alpha1/policies").
		DoRaw(ctx)
	if err != nil {
		return nil
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
		return nil
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
		}
		policies = append(policies, p)
	}
	return policies
}

// longhornRecurringJobs reads longhorn.io/v1beta2 RecurringJob objects.
// It also checks whether a BackupTarget is configured (offsite signal).
func longhornRecurringJobs(ctx context.Context, cs *kubernetes.Clientset) []model.BackupPolicy {
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
			return nil
		}
	}

	var list struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Spec struct {
				Task   string `json:"task"`   // "backup" or "snapshot"
				Cron   string `json:"cron"`
				Retain int    `json:"retain"`
				Groups []string `json:"groups"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil
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
		}
		policies = append(policies, p)
	}
	return policies
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

// ── Legacy helpers (kept for CoveredNamespaces population) ────────────────

func coveredNamespaces(ctx context.Context, cs *kubernetes.Clientset, tool string, b *model.Bundle) []string {
	switch tool {
	case "velero":
		return veleroScheduledNamespaces(ctx, cs)
	default:
		var ns []string
		for _, n := range b.Inventory.Namespaces {
			ns = append(ns, n.Name)
		}
		return ns
	}
}

func veleroScheduledNamespaces(ctx context.Context, cs *kubernetes.Clientset) []string {
	raw, err := cs.RESTClient().
		Get().
		AbsPath("/apis/velero.io/v1/schedules").
		DoRaw(ctx)
	if err != nil {
		return nil
	}
	content := string(raw)
	if strings.Contains(content, `"includedNamespaces"`) {
		return []string{"*"}
	}
	return nil
}

func uncoveredStatefulNamespaces(b *model.Bundle, covered []string) []string {
	if len(covered) == 1 && covered[0] == "*" {
		return nil
	}
	coveredSet := map[string]struct{}{}
	for _, ns := range covered {
		coveredSet[ns] = struct{}{}
	}
	seen := map[string]struct{}{}
	var uncovered []string
	for _, sts := range b.Inventory.StatefulSets {
		if _, ok := coveredSet[sts.Namespace]; !ok {
			if _, already := seen[sts.Namespace]; !already {
				uncovered = append(uncovered, sts.Namespace)
				seen[sts.Namespace] = struct{}{}
			}
		}
	}
	return uncovered
}
