package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s-recovery-visualizer/internal/model"
)

// WriteCSV writes one CSV file per inventory section to outDir/csv/.
// Files are UTF-8 with BOM for clean Excel opening on Windows.
func WriteCSV(outDir string, b *model.Bundle) error {
	dir := filepath.Join(outDir, "csv")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("csv: mkdir: %w", err)
	}
	writers := []func(string, *model.Bundle) error{
		writeNodesCSV,
		writeWorkloadsCSV,
		writeStorageCSV,
		writeNetworkingCSV,
		writeConfigCSV,
		writeImagesCSV,
		writeHelmCSV,
		writeCertificatesCSV,
		writeDRScoreCSV,
		writeRemediationCSV,
	}
	for _, fn := range writers {
		if err := fn(dir, b); err != nil {
			return err
		}
	}
	return nil
}

func csvFile(dir, name string) (*os.File, *csv.Writer, error) {
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return nil, nil, err
	}
	// UTF-8 BOM for Excel
	_, _ = f.Write([]byte{0xEF, 0xBB, 0xBF})
	return f, csv.NewWriter(f), nil
}

func writeNodesCSV(dir string, b *model.Bundle) error {
	f, w, err := csvFile(dir, "nodes.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	_ = w.Write([]string{"Name", "Roles", "Ready", "OS Image", "Kernel", "Runtime", "Kubelet Version", "Internal IP", "Taints"})
	for _, n := range b.Inventory.Nodes {
		rd := "false"
		if n.Ready {
			rd = "true"
		}
		_ = w.Write([]string{n.Name, strings.Join(n.Roles, ";"), rd, n.OSImage, n.KernelVersion, n.ContainerRuntime, n.KubeletVersion, n.InternalIP, strings.Join(n.Taints, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeWorkloadsCSV(dir string, b *model.Bundle) error {
	f, w, err := csvFile(dir, "workloads.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	_ = w.Write([]string{"Type", "Namespace", "Name", "Replicas", "Ready", "Images"})
	for _, d := range b.Inventory.Deployments {
		_ = w.Write([]string{"Deployment", d.Namespace, d.Name, fmt.Sprintf("%d", d.Replicas), fmt.Sprintf("%d", d.Ready), strings.Join(d.Images, ";")})
	}
	for _, ds := range b.Inventory.DaemonSets {
		_ = w.Write([]string{"DaemonSet", ds.Namespace, ds.Name, fmt.Sprintf("%d", ds.Desired), fmt.Sprintf("%d", ds.Ready), strings.Join(ds.Images, ";")})
	}
	for _, sts := range b.Inventory.StatefulSets {
		hasPVC := "false"
		if sts.HasVolumeClaim {
			hasPVC = "true"
		}
		_ = w.Write([]string{"StatefulSet", sts.Namespace, sts.Name, fmt.Sprintf("%d", sts.Replicas), hasPVC, ""})
	}
	for _, j := range b.Inventory.Jobs {
		comp := "false"
		if j.Completed {
			comp = "true"
		}
		_ = w.Write([]string{"Job", j.Namespace, j.Name, "", comp, ""})
	}
	for _, cj := range b.Inventory.CronJobs {
		_ = w.Write([]string{"CronJob", cj.Namespace, cj.Name, "", cj.Schedule, ""})
	}
	w.Flush()
	return w.Error()
}

func writeStorageCSV(dir string, b *model.Bundle) error {
	f, w, err := csvFile(dir, "storage.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	_ = w.Write([]string{"Kind", "Namespace", "Name", "StorageClass", "Provisioner", "Capacity", "Backend", "Reclaim Policy", "Access Modes", "Bound To"})
	pvMap := map[string]model.PersistentVolume{}
	for _, pv := range b.Inventory.PVs {
		pvMap[pv.ClaimRef] = pv
	}
	for _, pvc := range b.Inventory.PVCs {
		key := pvc.Namespace + "/" + pvc.Name
		pv := pvMap[key]
		_ = w.Write([]string{"PVC", pvc.Namespace, pvc.Name, pvc.StorageClass, "", pvc.RequestedSize, pv.Backend, pv.ReclaimPolicy, strings.Join(pvc.AccessModes, ";"), pv.Name})
	}
	for _, pv := range b.Inventory.PVs {
		_ = w.Write([]string{"PV", "", pv.Name, pv.StorageClass, "", pv.Capacity, pv.Backend, pv.ReclaimPolicy, "", pv.ClaimRef})
	}
	for _, sc := range b.Inventory.StorageClasses {
		expand := ""
		if sc.AllowVolumeExpansion != nil {
			if *sc.AllowVolumeExpansion {
				expand = "true"
			} else {
				expand = "false"
			}
		}
		_ = w.Write([]string{"StorageClass", "", sc.Name, sc.Name, sc.Provisioner, "", "", sc.ReclaimPolicy, "", expand})
	}
	w.Flush()
	return w.Error()
}

func writeNetworkingCSV(dir string, b *model.Bundle) error {
	f, w, err := csvFile(dir, "networking.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	_ = w.Write([]string{"Kind", "Namespace", "Name", "Type/Class", "Cluster IP", "External IP/TLS", "Rules/Selector"})
	for _, svc := range b.Inventory.Services {
		_ = w.Write([]string{"Service", svc.Namespace, svc.Name, svc.Type, svc.ClusterIP, svc.ExternalIP, ""})
	}
	for _, ing := range b.Inventory.Ingresses {
		tls := "false"
		if ing.TLS {
			tls = "true"
		}
		var rules []string
		for _, r := range ing.Rules {
			rules = append(rules, r.Host+"->"+r.Backend)
		}
		_ = w.Write([]string{"Ingress", ing.Namespace, ing.Name, ing.ClassName, "", tls, strings.Join(rules, ";")})
	}
	for _, np := range b.Inventory.NetworkPolicies {
		hasI := "false"
		if np.HasIngress {
			hasI = "true"
		}
		hasE := "false"
		if np.HasEgress {
			hasE = "true"
		}
		_ = w.Write([]string{"NetworkPolicy", np.Namespace, np.Name, "", "", fmt.Sprintf("ingress=%s egress=%s", hasI, hasE), np.PodSelector})
	}
	w.Flush()
	return w.Error()
}

func writeConfigCSV(dir string, b *model.Bundle) error {
	f, w, err := csvFile(dir, "config.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	_ = w.Write([]string{"Kind", "Namespace", "Name", "Detail"})
	for _, cm := range b.Inventory.ConfigMaps {
		_ = w.Write([]string{"ConfigMap", cm.Namespace, cm.Name, fmt.Sprintf("keys=%d", cm.KeyCount)})
	}
	for _, s := range b.Inventory.Secrets {
		_ = w.Write([]string{"Secret", s.Namespace, s.Name, fmt.Sprintf("type=%s keys=%d", s.Type, s.KeyCount)})
	}
	for _, cr := range b.Inventory.ClusterRoles {
		custom := "built-in"
		if cr.Custom {
			custom = "custom"
		}
		_ = w.Write([]string{"ClusterRole", "", cr.Name, fmt.Sprintf("%s rules=%d", custom, cr.RuleCount)})
	}
	for _, crd := range b.Inventory.CRDs {
		_ = w.Write([]string{"CRD", "", crd.Group, fmt.Sprintf("scope=%s versions=%s", crd.Scope, strings.Join(crd.Versions, ";"))})
	}
	w.Flush()
	return w.Error()
}

func writeImagesCSV(dir string, b *model.Bundle) error {
	f, w, err := csvFile(dir, "images.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	_ = w.Write([]string{"Image", "Registry", "Public", "Used By"})
	for _, img := range b.Inventory.Images {
		pub := "false"
		if img.IsPublic {
			pub = "true"
		}
		_ = w.Write([]string{img.Image, img.Registry, pub, strings.Join(img.Workloads, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeHelmCSV(dir string, b *model.Bundle) error {
	f, w, err := csvFile(dir, "helm.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	_ = w.Write([]string{"Namespace", "Release", "Chart", "Version", "App Version", "Status"})
	for _, h := range b.Inventory.HelmReleases {
		_ = w.Write([]string{h.Namespace, h.Name, h.Chart, h.Version, h.AppVersion, h.Status})
	}
	w.Flush()
	return w.Error()
}

func writeCertificatesCSV(dir string, b *model.Bundle) error {
	f, w, err := csvFile(dir, "certificates.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	_ = w.Write([]string{"Namespace", "Name", "Issuer", "Secret", "Ready", "Expires", "Days To Expiry"})
	for _, c := range b.Inventory.Certificates {
		rd := "false"
		if c.Ready {
			rd = "true"
		}
		_ = w.Write([]string{c.Namespace, c.Name, c.Issuer, c.SecretName, rd, c.NotAfter, fmt.Sprintf("%d", c.DaysToExpiry)})
	}
	w.Flush()
	return w.Error()
}

func writeDRScoreCSV(dir string, b *model.Bundle) error {
	f, w, err := csvFile(dir, "dr-score.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	_ = w.Write([]string{"Domain", "Score", "Max", "Weight"})
	_ = w.Write([]string{"Storage", fmt.Sprintf("%d", b.Score.Storage.Final), "100", "35%"})
	_ = w.Write([]string{"Workload", fmt.Sprintf("%d", b.Score.Workload.Final), "100", "20%"})
	_ = w.Write([]string{"Config", fmt.Sprintf("%d", b.Score.Config.Final), "100", "15%"})
	_ = w.Write([]string{"Backup/Recovery", fmt.Sprintf("%d", b.Score.Backup.Final), "100", "30%"})
	_ = w.Write([]string{"Overall", fmt.Sprintf("%d", b.Score.Overall.Final), "100", "100%"})
	_ = w.Write([]string{})
	_ = w.Write([]string{"Severity", "Resource", "Finding", "Recommendation"})
	for _, finding := range b.Inventory.Findings {
		_ = w.Write([]string{finding.Severity, finding.ResourceID, finding.Message, finding.Recommendation})
	}
	w.Flush()
	return w.Error()
}

func writeRemediationCSV(dir string, b *model.Bundle) error {
	f, w, err := csvFile(dir, "remediation.csv")
	if err != nil {
		return err
	}
	defer f.Close()
	_ = w.Write([]string{"Priority", "Category", "Title", "Detail", "Target Notes", "Commands", "Finding ID"})
	for _, s := range b.Inventory.RemediationSteps {
		_ = w.Write([]string{
			fmt.Sprintf("%d", s.Priority),
			s.Category,
			s.Title,
			s.Detail,
			s.TargetNotes,
			strings.Join(s.Commands, " | "),
			s.FindingID,
		})
	}
	w.Flush()
	return w.Error()
}
