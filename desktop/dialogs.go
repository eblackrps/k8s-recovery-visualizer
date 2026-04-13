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

func (a *App) PickOutputDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose output directory",
	})
}
