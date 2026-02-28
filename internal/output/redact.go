package output

import (
	"encoding/json"
	"fmt"
	"os"

	"k8s-recovery-visualizer/internal/model"
)

// WriteRedactedJSON writes a copy of the bundle with sensitive identifiers replaced.
func WriteRedactedJSON(path string, b *model.Bundle) error {
	r := redact(b)
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// WriteRedactedReport writes a redacted HTML report.
func WriteRedactedReport(path string, b *model.Bundle) error {
	r := redact(b)
	return WriteReport(path, r)
}

// redact returns a deep copy of the bundle with sensitive fields replaced.
// Namespace names, node names, cluster UID, API endpoint, and customer
// metadata are replaced with opaque tokens. Resource names are preserved
// because they are rarely sensitive; namespace names are masked.
func redact(b *model.Bundle) *model.Bundle {
	// Deep copy via JSON round-trip
	data, _ := json.Marshal(b)
	var r model.Bundle
	_ = json.Unmarshal(data, &r)

	// Build namespace map: real name → redacted label
	nsMap := make(map[string]string, len(r.Inventory.Namespaces))
	for i, ns := range r.Inventory.Namespaces {
		token := fmt.Sprintf("namespace-%d", i+1)
		nsMap[ns.Name] = token
		r.Inventory.Namespaces[i].Name = token
		r.Inventory.Namespaces[i].ID = "ns:" + token
	}

	// Build node map
	nodeMap := make(map[string]string, len(r.Inventory.Nodes))
	for i, n := range r.Inventory.Nodes {
		token := fmt.Sprintf("node-%d", i+1)
		nodeMap[n.Name] = token
		r.Inventory.Nodes[i].Name = token
		r.Inventory.Nodes[i].InternalIP = ""
	}

	// Strip cluster-level identifiers
	r.Cluster.APIServer.Endpoint = "[redacted]"
	r.Cluster.Platform.ClusterUID = "[redacted]"

	// Strip customer metadata
	r.Metadata.CustomerID = "[redacted]"
	r.Metadata.Site = "[redacted]"
	r.Metadata.ClusterName = "[redacted]"
	r.Metadata.Environment = "[redacted]"

	// Redact namespace references in inventory
	replaceNS := func(ns string) string {
		if v, ok := nsMap[ns]; ok {
			return v
		}
		return ns
	}

	for i := range r.Inventory.Pods {
		r.Inventory.Pods[i].Namespace = replaceNS(r.Inventory.Pods[i].Namespace)
	}
	for i := range r.Inventory.PVCs {
		r.Inventory.PVCs[i].Namespace = replaceNS(r.Inventory.PVCs[i].Namespace)
	}
	for i := range r.Inventory.StatefulSets {
		r.Inventory.StatefulSets[i].Namespace = replaceNS(r.Inventory.StatefulSets[i].Namespace)
	}
	for i := range r.Inventory.Deployments {
		r.Inventory.Deployments[i].Namespace = replaceNS(r.Inventory.Deployments[i].Namespace)
	}
	for i := range r.Inventory.DaemonSets {
		r.Inventory.DaemonSets[i].Namespace = replaceNS(r.Inventory.DaemonSets[i].Namespace)
	}
	for i := range r.Inventory.Jobs {
		r.Inventory.Jobs[i].Namespace = replaceNS(r.Inventory.Jobs[i].Namespace)
	}
	for i := range r.Inventory.CronJobs {
		r.Inventory.CronJobs[i].Namespace = replaceNS(r.Inventory.CronJobs[i].Namespace)
	}
	for i := range r.Inventory.Services {
		r.Inventory.Services[i].Namespace = replaceNS(r.Inventory.Services[i].Namespace)
	}
	for i := range r.Inventory.Ingresses {
		r.Inventory.Ingresses[i].Namespace = replaceNS(r.Inventory.Ingresses[i].Namespace)
	}
	for i := range r.Inventory.ConfigMaps {
		r.Inventory.ConfigMaps[i].Namespace = replaceNS(r.Inventory.ConfigMaps[i].Namespace)
	}
	for i := range r.Inventory.Secrets {
		r.Inventory.Secrets[i].Namespace = replaceNS(r.Inventory.Secrets[i].Namespace)
	}
	for i := range r.Inventory.NetworkPolicies {
		r.Inventory.NetworkPolicies[i].Namespace = replaceNS(r.Inventory.NetworkPolicies[i].Namespace)
	}
	for i := range r.Inventory.HPAs {
		r.Inventory.HPAs[i].Namespace = replaceNS(r.Inventory.HPAs[i].Namespace)
	}
	for i := range r.Inventory.PodDisruptionBudgets {
		r.Inventory.PodDisruptionBudgets[i].Namespace = replaceNS(r.Inventory.PodDisruptionBudgets[i].Namespace)
	}
	for i := range r.Inventory.ResourceQuotas {
		r.Inventory.ResourceQuotas[i].Namespace = replaceNS(r.Inventory.ResourceQuotas[i].Namespace)
	}
	for i := range r.Inventory.HelmReleases {
		r.Inventory.HelmReleases[i].Namespace = replaceNS(r.Inventory.HelmReleases[i].Namespace)
	}
	for i := range r.Inventory.Certificates {
		r.Inventory.Certificates[i].Namespace = replaceNS(r.Inventory.Certificates[i].Namespace)
	}

	// Redact node names in findings resourceIds
	for i, f := range r.Inventory.Findings {
		for real, token := range nodeMap {
			if f.ResourceID == real {
				r.Inventory.Findings[i].ResourceID = token
			}
		}
	}

	// Clear scan namespaces (they contain real names)
	r.ScanNamespaces = nil

	return &r
}
