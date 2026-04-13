package appcore

import (
	"strings"
	"testing"
)

func TestPermissionProbeManifestUsesRoleBindingForNamespacedAccess(t *testing.T) {
	probe := permissionProbe{
		id:              "pods",
		resource:        "pods",
		namespaceScoped: true,
		rules: []rbacRuleTemplate{
			{apiGroups: []string{""}, resources: []string{"pods"}},
			{apiGroups: []string{"apps"}, resources: []string{"deployments"}},
		},
	}

	manifest := probe.manifest([]string{"payments", "frontend"})
	if !strings.Contains(manifest, "kind: RoleBinding") {
		t.Fatalf("manifest = %q, want RoleBinding guidance", manifest)
	}
	if !strings.Contains(manifest, "namespace: payments") {
		t.Fatalf("manifest = %q, want first namespace binding", manifest)
	}
	if !strings.Contains(manifest, "Repeat the RoleBinding for namespaces: frontend") {
		t.Fatalf("manifest = %q, want multi-namespace hint", manifest)
	}
}

func TestPermissionProbeCommandsUseDefaultNamespaceFallback(t *testing.T) {
	probe := permissionProbe{id: "pods", resource: "pods", namespaceScoped: true}
	commands := probe.commands(nil)
	if len(commands) == 0 || !strings.Contains(commands[0], "-n default") {
		t.Fatalf("commands = %v, want default namespace fallback", commands)
	}
}
