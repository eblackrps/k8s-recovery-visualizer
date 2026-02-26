package analyze

import (
  "k8s-recovery-visualizer/internal/model"
)

func statusFromScore(score int) string {
  if score >= 90 {
    return "PASS"
  }
  if score >= 60 {
    return "WARN"
  }
  return "FAIL"
}

func BuildChecks(b *model.Bundle, minScore int) []model.Check {
  _ = minScore

  results := RunChecks(&b.Cluster, &b.Inventory)
  out := make([]model.Check, 0, len(results))
  for _, r := range results {
    msg := r.Note
    if r.Category != "" {
      if msg != "" {
        msg = r.Category + ": " + msg
      } else {
        msg = r.Category
      }
    }
    out = append(out, model.Check{
      ID:      r.ID,
      Title:   r.Title,
      Status:  statusFromScore(r.Score),
      Weight:  1,
      Message: msg,
    })
  }
  return out
}

