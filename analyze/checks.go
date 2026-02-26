package analyze

import (
  "fmt"
  "sort"
  "strings"

  "k8s-recovery-visualizer/internal/model"
)

type CheckResult struct {
  ID       string
  Title    string
  Category string
  Score    int
  Note     string
}

func RunChecks(cluster *model.Cluster, inv *model.Inventory) []CheckResult {
  var out []CheckResult

  out = append(out,
    checkNodesPresent(inv),
    checkNodeReadiness(inv),
    checkStorageClassesPresent(inv),
    checkDefaultStorageClass(inv),
  )

  sort.Slice(out, func(i, j int) bool {
    if out[i].Category == out[j].Category {
      return out[i].ID < out[j].ID
    }
    return out[i].Category < out[j].Category
  })

  return out
}

func checkNodesPresent(inv *model.Inventory) CheckResult {
  n := len(inv.Nodes)
  if n == 0 {
    return CheckResult{
      ID:       "cluster.nodes.present",
      Title:    "Nodes discovered",
      Category: "Cluster",
      Score:    0,
      Note:     "No nodes found in inventory. Either node collection is not implemented yet or the scan lacks permissions.",
    }
  }
  return CheckResult{
    ID:       "cluster.nodes.present",
    Title:    "Nodes discovered",
    Category: "Cluster",
    Score:    100,
    Note:     fmt.Sprintf("Found %d node(s).", n),
  }
}

func checkNodeReadiness(inv *model.Inventory) CheckResult {
  total := len(inv.Nodes)
  if total == 0 {
    return CheckResult{
      ID:       "cluster.nodes.ready",
      Title:    "Node readiness",
      Category: "Cluster",
      Score:    0,
      Note:     "No nodes found, cannot assess readiness.",
    }
  }

  ready := 0
  var notReady []string
  for _, node := range inv.Nodes {
    if node.Ready {
      ready++
    } else {
      if node.Name != "" {
        notReady = append(notReady, node.Name)
      } else {
        notReady = append(notReady, "<unnamed>")
      }
    }
  }

  score := int(float64(ready) / float64(total) * 100.0)

  note := fmt.Sprintf("%d/%d node(s) Ready.", ready, total)
  if len(notReady) > 0 {
    if len(notReady) > 5 {
      note += fmt.Sprintf(" Not Ready: %s (+%d more).", strings.Join(notReady[:5], ", "), len(notReady)-5)
    } else {
      note += fmt.Sprintf(" Not Ready: %s.", strings.Join(notReady, ", "))
    }
  }

  return CheckResult{
    ID:       "cluster.nodes.ready",
    Title:    "Node readiness",
    Category: "Cluster",
    Score:    score,
    Note:     note,
  }
}

func checkStorageClassesPresent(inv *model.Inventory) CheckResult {
  n := len(inv.StorageClasses)
  if n == 0 {
    return CheckResult{
      ID:       "storage.classes.present",
      Title:    "StorageClasses discovered",
      Category: "Storage",
      Score:    0,
      Note:     "No StorageClasses found. Either storageclass collection is not implemented yet or dynamic provisioning is not configured.",
    }
  }
  return CheckResult{
    ID:       "storage.classes.present",
    Title:    "StorageClasses discovered",
    Category: "Storage",
    Score:    100,
    Note:     fmt.Sprintf("Found %d StorageClass(es).", n),
  }
}

func checkDefaultStorageClass(inv *model.Inventory) CheckResult {
  if len(inv.StorageClasses) == 0 {
    return CheckResult{
      ID:       "storage.classes.default",
      Title:    "Default StorageClass",
      Category: "Storage",
      Score:    0,
      Note:     "No StorageClasses found, cannot determine default.",
    }
  }

  defaults := 0
  var names []string
  for _, sc := range inv.StorageClasses {
    if isDefaultStorageClass(sc.Annotations) {
      defaults++
      if sc.Name != "" {
        names = append(names, sc.Name)
      } else {
        names = append(names, "<unnamed>")
      }
    }
  }

  if defaults == 0 {
    return CheckResult{
      ID:       "storage.classes.default",
      Title:    "Default StorageClass",
      Category: "Storage",
      Score:    60,
      Note:     "No default StorageClass detected. PVCs without storageClassName may fail to bind depending on workload manifests.",
    }
  }

  return CheckResult{
    ID:       "storage.classes.default",
    Title:    "Default StorageClass",
    Category: "Storage",
    Score:    100,
    Note:     fmt.Sprintf("%d default StorageClass(es): %s.", defaults, strings.Join(names, ", ")),
  }
}

func isDefaultStorageClass(ann map[string]string) bool {
  if ann == nil {
    return false
  }
  v := strings.ToLower(strings.TrimSpace(ann["storageclass.kubernetes.io/is-default-class"]))
  if v == "true" {
    return true
  }
  v = strings.ToLower(strings.TrimSpace(ann["storageclass.beta.kubernetes.io/is-default-class"]))
  return v == "true"
}
