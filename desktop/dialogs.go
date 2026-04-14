package main

import "github.com/wailsapp/wails/v2/pkg/runtime"

func (a *App) PickBundleFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open recovery bundle",
		Filters: []runtime.FileFilter{
			{DisplayName: "Bundle files", Pattern: "*.json;*.zip;*.tar.gz;*.tgz;*.tar"},
			{DisplayName: "JSON bundles", Pattern: "*.json"},
			{DisplayName: "Bundle archives", Pattern: "*.zip;*.tar.gz;*.tgz;*.tar"},
		},
	})
}

func (a *App) PickKubeconfigFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose kubeconfig file (validated by content)",
		Filters: []runtime.FileFilter{
			{DisplayName: "All files", Pattern: "*"},
			{DisplayName: "Common kubeconfig names", Pattern: "config;kubeconfig;*.yaml;*.yml;*.conf;*.config"},
		},
	})
}

func (a *App) PickCertificateFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose CA certificate",
		Filters: []runtime.FileFilter{
			{DisplayName: "Certificate files", Pattern: "*.crt;*.pem;*.cer"},
			{DisplayName: "All files", Pattern: "*"},
		},
	})
}

func (a *App) PickOutputDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose output directory",
	})
}
