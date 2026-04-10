package scoring

import (
	"strings"
	"testing"
)

func TestDefaultRegistryHasRequiredProfilesAndWeights(t *testing.T) {
	reg := Default()

	if got := reg.DomainWeight("storage"); got != 35 {
		t.Fatalf("storage weight = %d, want 35", got)
	}
	if got := reg.DomainWeight("backup"); got != 30 {
		t.Fatalf("backup weight = %d, want 30", got)
	}

	requiredProfiles := []string{"standard", "enterprise", "dev", "airgap"}
	for _, name := range requiredProfiles {
		profile := reg.Profile(name)
		if profile.Name != name {
			t.Fatalf("profile %s not loaded correctly", name)
		}
	}
}

func TestLoadRejectsInvalidDomainWeightSum(t *testing.T) {
	ruleRaw := []byte(`{"version":"1.0.0","domainWeights":{"storage":40,"workload":20,"config":15,"backup":30},"rules":[]}`)
	profileRaw := []byte(`{"version":"1.0.0","profiles":[{"name":"standard","description":"","multipliers":{}},{"name":"enterprise","description":"","multipliers":{}},{"name":"dev","description":"","multipliers":{}},{"name":"airgap","description":"","multipliers":{}}]}`)

	_, err := Load(ruleRaw, profileRaw)
	if err == nil || !strings.Contains(err.Error(), "sum to 100") {
		t.Fatalf("Load() error = %v, want weight sum validation", err)
	}
}

func TestLoadRejectsUnknownMultiplierKey(t *testing.T) {
	ruleRaw := []byte(`{"version":"1.0.0","domainWeights":{"storage":35,"workload":20,"config":15,"backup":30},"rules":[{"id":"PVC_UNBOUND","title":"PVC","domain":"storage","severity":"CRITICAL","basePenalty":25,"multiplierKey":"mystery"}]}`)
	profileRaw := []byte(`{"version":"1.0.0","profiles":[{"name":"standard","description":"","multipliers":{}},{"name":"enterprise","description":"","multipliers":{}},{"name":"dev","description":"","multipliers":{}},{"name":"airgap","description":"","multipliers":{}}]}`)

	_, err := Load(ruleRaw, profileRaw)
	if err == nil || !strings.Contains(err.Error(), "unknown multiplier key") {
		t.Fatalf("Load() error = %v, want multiplier validation", err)
	}
}
