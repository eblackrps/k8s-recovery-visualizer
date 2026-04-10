package scoring

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed config/*.json
var configFS embed.FS

type RulePack struct {
	Version       string           `json:"version"`
	DomainWeights map[string]int   `json:"domainWeights"`
	Rules         []RuleDefinition `json:"rules"`
}

type RuleDefinition struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Domain        string `json:"domain"`
	Severity      string `json:"severity"`
	BasePenalty   int    `json:"basePenalty"`
	MultiplierKey string `json:"multiplierKey"`
}

type ProfilePack struct {
	Version  string              `json:"version"`
	Profiles []ProfileDefinition `json:"profiles"`
}

type ProfileDefinition struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Multipliers map[string]float64 `json:"multipliers"`
}

type Registry struct {
	rulePack    RulePack
	profilePack ProfilePack
	rules       map[string]RuleDefinition
	profiles    map[string]ProfileDefinition
}

var defaultRegistry = mustLoadDefaultRegistry()

func Default() *Registry {
	return defaultRegistry
}

func mustLoadDefaultRegistry() *Registry {
	ruleRaw, err := configFS.ReadFile("config/rule-pack.v1.json")
	if err != nil {
		panic(err)
	}
	profileRaw, err := configFS.ReadFile("config/profiles.v1.json")
	if err != nil {
		panic(err)
	}
	reg, err := Load(ruleRaw, profileRaw)
	if err != nil {
		panic(err)
	}
	return reg
}

func Load(ruleRaw, profileRaw []byte) (*Registry, error) {
	var rulePack RulePack
	if err := json.Unmarshal(ruleRaw, &rulePack); err != nil {
		return nil, fmt.Errorf("parse rule pack: %w", err)
	}
	var profilePack ProfilePack
	if err := json.Unmarshal(profileRaw, &profilePack); err != nil {
		return nil, fmt.Errorf("parse profile pack: %w", err)
	}

	reg := &Registry{
		rulePack:    rulePack,
		profilePack: profilePack,
		rules:       map[string]RuleDefinition{},
		profiles:    map[string]ProfileDefinition{},
	}
	if err := reg.validate(); err != nil {
		return nil, err
	}
	return reg, nil
}

func (r *Registry) validate() error {
	if r.rulePack.Version == "" {
		return fmt.Errorf("rule pack version is required")
	}
	if r.profilePack.Version == "" {
		return fmt.Errorf("profile pack version is required")
	}

	requiredDomains := []string{"storage", "workload", "config", "backup"}
	totalWeight := 0
	for _, domain := range requiredDomains {
		weight, ok := r.rulePack.DomainWeights[domain]
		if !ok {
			return fmt.Errorf("missing domain weight for %s", domain)
		}
		if weight <= 0 {
			return fmt.Errorf("invalid domain weight for %s", domain)
		}
		totalWeight += weight
	}
	if totalWeight != 100 {
		return fmt.Errorf("domain weights must sum to 100, got %d", totalWeight)
	}

	allowedSeverities := map[string]struct{}{
		"CRITICAL": {},
		"HIGH":     {},
		"MEDIUM":   {},
		"LOW":      {},
		"INFO":     {},
	}
	allowedMultipliers := map[string]struct{}{
		"":               {},
		"restoreTesting": {},
		"immutability":   {},
		"replication":    {},
		"security":       {},
		"airgap":         {},
	}

	for _, rule := range r.rulePack.Rules {
		rule.ID = strings.TrimSpace(rule.ID)
		if rule.ID == "" {
			return fmt.Errorf("rule with empty id")
		}
		if _, exists := r.rules[rule.ID]; exists {
			return fmt.Errorf("duplicate rule id %s", rule.ID)
		}
		if _, ok := r.rulePack.DomainWeights[rule.Domain]; !ok {
			return fmt.Errorf("rule %s has unknown domain %s", rule.ID, rule.Domain)
		}
		if _, ok := allowedSeverities[rule.Severity]; !ok {
			return fmt.Errorf("rule %s has unknown severity %s", rule.ID, rule.Severity)
		}
		if rule.BasePenalty < 0 {
			return fmt.Errorf("rule %s has negative penalty", rule.ID)
		}
		if _, ok := allowedMultipliers[rule.MultiplierKey]; !ok {
			return fmt.Errorf("rule %s has unknown multiplier key %s", rule.ID, rule.MultiplierKey)
		}
		r.rules[rule.ID] = rule
	}

	requiredProfiles := []string{"standard", "enterprise", "dev", "airgap"}
	for _, profile := range r.profilePack.Profiles {
		name := NormalizeProfile(profile.Name)
		if name == "" {
			return fmt.Errorf("profile with empty name")
		}
		if _, exists := r.profiles[name]; exists {
			return fmt.Errorf("duplicate profile %s", name)
		}
		if profile.Multipliers == nil {
			profile.Multipliers = map[string]float64{}
		}
		for key, value := range profile.Multipliers {
			if _, ok := allowedMultipliers[key]; !ok || key == "" {
				return fmt.Errorf("profile %s uses unknown multiplier key %s", name, key)
			}
			if value <= 0 {
				return fmt.Errorf("profile %s has invalid multiplier %s=%v", name, key, value)
			}
		}
		profile.Name = name
		r.profiles[name] = profile
	}
	for _, required := range requiredProfiles {
		if _, ok := r.profiles[required]; !ok {
			return fmt.Errorf("missing required profile %s", required)
		}
	}

	return nil
}

func NormalizeProfile(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "enterprise", "dev", "airgap":
		return name
	default:
		return "standard"
	}
}

func (r *Registry) Rule(id string) (RuleDefinition, bool) {
	rule, ok := r.rules[id]
	return rule, ok
}

func (r *Registry) MustRule(id string) RuleDefinition {
	rule, ok := r.Rule(id)
	if !ok {
		panic(fmt.Sprintf("unknown scoring rule %s", id))
	}
	return rule
}

func (r *Registry) DomainWeight(domain string) int {
	return r.rulePack.DomainWeights[domain]
}

func (r *Registry) Profile(name string) ProfileDefinition {
	return r.profiles[NormalizeProfile(name)]
}

func (r *Registry) ProfileMultipliers(name string) map[string]float64 {
	profile := r.Profile(name)
	out := map[string]float64{}
	for key, value := range profile.Multipliers {
		out[key] = value
	}
	return out
}

func (r *Registry) ProfileNames() []string {
	names := make([]string, 0, len(r.profiles))
	for name := range r.profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) RuleIDs() []string {
	ids := make([]string, 0, len(r.rules))
	for id := range r.rules {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
