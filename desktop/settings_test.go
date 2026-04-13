package main

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"testing"
)

func TestSettingsLoadSaveRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "settings.json")
	want := Settings{
		WorkspaceRoot:         filepath.Join(t.TempDir(), "workspace"),
		DefaultOutputDir:      filepath.Join(t.TempDir(), "workspace", "out"),
		DefaultProfile:        "enterprise",
		IncludeSecretMetadata: true,
		Summary:               true,
		Runbook:               false,
		Redact:                true,
		CSVExport:             false,
	}

	if err := saveSettingsToPath(path, want); err != nil {
		t.Fatalf("saveSettingsToPath() error = %v", err)
	}

	got, err := loadSettingsFromPath(path, defaultSettings())
	if err != nil {
		t.Fatalf("loadSettingsFromPath() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("settings mismatch\n got: %#v\nwant: %#v", got, want)
	}

	if goruntime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat(file) error = %v", err)
		}
		if gotPerm := info.Mode().Perm(); gotPerm != settingsFileMode {
			t.Fatalf("settings file mode = %#o, want %#o", gotPerm, settingsFileMode)
		}

		dirInfo, err := os.Stat(filepath.Dir(path))
		if err != nil {
			t.Fatalf("Stat(dir) error = %v", err)
		}
		if gotPerm := dirInfo.Mode().Perm(); gotPerm != settingsDirMode {
			t.Fatalf("settings dir mode = %#o, want %#o", gotPerm, settingsDirMode)
		}
	}
}

func TestStartupSurfacesSettingsLoadFailures(t *testing.T) {
	configRoot := t.TempDir()
	settingsPath := settingsPathForConfigDir(configRoot)
	if err := os.MkdirAll(filepath.Dir(settingsPath), settingsDirMode); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(settingsPath, []byte("{not-json"), settingsFileMode); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("APPDATA", configRoot)
	t.Setenv("AppData", configRoot)
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	app := NewApp()
	app.startup(context.Background())

	alerts := app.GetStartupAlerts()
	if len(alerts) != 1 {
		t.Fatalf("len(GetStartupAlerts()) = %d, want 1", len(alerts))
	}
	if alerts[0].Tone != "error" {
		t.Fatalf("alert tone = %q, want error", alerts[0].Tone)
	}
	if app.GetSettings() != defaultSettings() {
		t.Fatalf("GetSettings() = %#v, want defaults %#v", app.GetSettings(), defaultSettings())
	}
}

func TestLinuxDefaultWorkspaceRootPrefersConfiguredDocumentsDir(t *testing.T) {
	homeDir := filepath.Join(t.TempDir(), "home")
	configDir := filepath.Join(homeDir, ".config")
	documentsDir := filepath.Join(homeDir, "Documents")

	got := linuxDefaultWorkspaceRoot(homeDir, configDir, func(key string) string {
		if key == "XDG_DOCUMENTS_DIR" {
			return documentsDir
		}
		return ""
	})

	want := filepath.Join(documentsDir, "k8s-recovery-visualizer")
	if got != want {
		t.Fatalf("linuxDefaultWorkspaceRoot() = %q, want %q", got, want)
	}
}

func TestLinuxDocumentsDirParsesUserDirsConfig(t *testing.T) {
	homeDir := filepath.Join(t.TempDir(), "home")
	configDir := filepath.Join(homeDir, ".config")
	if err := os.MkdirAll(configDir, settingsDirMode); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	configPath := filepath.Join(configDir, "user-dirs.dirs")
	raw := []byte("XDG_DOCUMENTS_DIR=\"$HOME/Workspace Docs\"\n")
	if err := os.WriteFile(configPath, raw, settingsFileMode); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got := linuxDocumentsDir(homeDir, configDir, func(string) string { return "" })
	want := filepath.Join(homeDir, "Workspace Docs")
	if got != want {
		t.Fatalf("linuxDocumentsDir() = %q, want %q", got, want)
	}
}

func TestLinuxDefaultWorkspaceRootFallsBackToDataHome(t *testing.T) {
	homeDir := filepath.Join(t.TempDir(), "home")
	configDir := filepath.Join(homeDir, ".config")
	dataHome := filepath.Join(homeDir, ".local", "share")

	got := linuxDefaultWorkspaceRoot(homeDir, configDir, func(key string) string {
		if key == "XDG_DATA_HOME" {
			return dataHome
		}
		return ""
	})

	want := filepath.Join(dataHome, "k8s-recovery-visualizer", "workspace")
	if got != want {
		t.Fatalf("linuxDefaultWorkspaceRoot() = %q, want %q", got, want)
	}
}
