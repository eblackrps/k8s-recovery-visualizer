package analyze

import "testing"

func TestWeightedOverall(t *testing.T) {
	// All domains at 100 → overall 100
	if got := weightedOverall(100, 100, 100, 100); got != 100 {
		t.Errorf("weightedOverall(100,100,100,100) = %d, want 100", got)
	}
	// All domains at 0 → overall 0
	if got := weightedOverall(0, 0, 0, 0); got != 0 {
		t.Errorf("weightedOverall(0,0,0,0) = %d, want 0", got)
	}
	// Verify weighting: S=100 W=0 C=0 B=0 → 35
	if got := weightedOverall(100, 0, 0, 0); got != 35 {
		t.Errorf("weightedOverall(100,0,0,0) = %d, want 35", got)
	}
	// B=100 rest=0 → 30
	if got := weightedOverall(0, 0, 0, 100); got != 30 {
		t.Errorf("weightedOverall(0,0,0,100) = %d, want 30", got)
	}
	// W=100 rest=0 → 20
	if got := weightedOverall(0, 100, 0, 0); got != 20 {
		t.Errorf("weightedOverall(0,100,0,0) = %d, want 20", got)
	}
	// C=100 rest=0 → 15
	if got := weightedOverall(0, 0, 100, 0); got != 15 {
		t.Errorf("weightedOverall(0,0,100,0) = %d, want 15", got)
	}
	// Weights sum to 100: 35+20+15+30 = 100 ✓
	if storageWeight+workloadWeight+configWeight+backupWeight != 100 {
		t.Errorf("domain weights must sum to 100, got %d",
			storageWeight+workloadWeight+configWeight+backupWeight)
	}
}

func TestClamp(t *testing.T) {
	if got := clamp(-10); got != 0 {
		t.Errorf("clamp(-10) = %d, want 0", got)
	}
	if got := clamp(0); got != 0 {
		t.Errorf("clamp(0) = %d, want 0", got)
	}
	if got := clamp(50); got != 50 {
		t.Errorf("clamp(50) = %d, want 50", got)
	}
	if got := clamp(100); got != 100 {
		t.Errorf("clamp(100) = %d, want 100", got)
	}
	if got := clamp(110); got != 100 {
		t.Errorf("clamp(110) = %d, want 100", got)
	}
}

func TestPenScale(t *testing.T) {
	// 1.0× multiplier → identity
	if got := penScale(20, 1.0); got != 20 {
		t.Errorf("penScale(20, 1.0) = %d, want 20", got)
	}
	// 2.0× multiplier → doubled
	if got := penScale(10, 2.0); got != 20 {
		t.Errorf("penScale(10, 2.0) = %d, want 20", got)
	}
	// 0.0× multiplier → clamped to 1 (never fully erased)
	if got := penScale(20, 0.0); got != 1 {
		t.Errorf("penScale(20, 0.0) = %d, want 1", got)
	}
	// 0.5× multiplier → halved (rounded)
	if got := penScale(10, 0.5); got != 5 {
		t.Errorf("penScale(10, 0.5) = %d, want 5", got)
	}
}

func TestJoinFirst(t *testing.T) {
	// Fewer than max → all joined with commas
	if got := joinFirst([]string{"a", "b"}, 3); got != "a,b" {
		t.Errorf("joinFirst([a,b], 3) = %q, want %q", got, "a,b")
	}
	// Exactly max → no ellipsis
	if got := joinFirst([]string{"a", "b", "c"}, 3); got != "a,b,c" {
		t.Errorf("joinFirst([a,b,c], 3) = %q, want %q", got, "a,b,c")
	}
	// More than max → truncated with ellipsis
	if got := joinFirst([]string{"a", "b", "c", "d"}, 3); got != "a,b,c..." {
		t.Errorf("joinFirst([a,b,c,d], 3) = %q, want %q", got, "a,b,c...")
	}
}
