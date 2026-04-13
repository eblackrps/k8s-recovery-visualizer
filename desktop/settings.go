package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"k8s-recovery-visualizer/internal/appcore"
)

const (
	settingsDirMode  = 0o700
	settingsFileMode = 0o600
)

type Settings struct {
	WorkspaceRoot         string `json:"workspaceRoot"`
	DefaultOutputDir      string `json:"defaultOutputDir"`
	DefaultProfile        string `json:"defaultProfile"`
	IncludeSecretMetadata bool   `json:"includeSecretMetadata"`
	Summary               bool   `json:"summary"`
	Runbook               bool   `json:"runbook"`
	Redact                bool   `json:"redact"`
	CSVExport             bool   `json:"csvExport"`
}

type partialSettings struct {
	WorkspaceRoot         *string `json:"workspaceRoot"`
	DefaultOutputDir      *string `json:"defaultOutputDir"`
	DefaultProfile        *string `json:"defaultProfile"`
	IncludeSecretMetadata *bool   `json:"includeSecretMetadata"`
	Summary               *bool   `json:"summary"`
	Runbook               *bool   `json:"runbook"`
	Redact                *bool   `json:"redact"`
	CSVExport             *bool   `json:"csvExport"`
}

func (a *App) GetSettings() Settings {
	return a.currentSettings()
}

func (a *App) GetStartupAlerts() []AppAlert {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()

	alerts := make([]AppAlert, len(a.startupAlerts))
	copy(alerts, a.startupAlerts)
	return alerts
}

func (a *App) SaveSettings(settings Settings) error {
	merged := mergeSettings(defaultSettings(), settings)
	a.setSettings(merged)

	if err := saveSettings(merged); err != nil {
		a.logError("save desktop settings", err)
		return err
	}
	return nil
}

func (a *App) ListProjects(root string) ([]appcore.ProjectSummary, error) {
	if root == "" {
		root = a.currentSettings().WorkspaceRoot
	}
	return a.service.ListProjects(root)
}

func settingsPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return settingsPathForConfigDir(root), nil
}

func settingsPathForConfigDir(configRoot string) string {
	return filepath.Join(configRoot, "k8s-recovery-visualizer", "settings.json")
}

func loadSettings(defaults Settings) (Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return defaults, err
	}
	return loadSettingsFromPath(path, defaults)
}

func loadSettingsFromPath(path string, defaults Settings) (Settings, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaults, nil
		}
		return defaults, err
	}
	settings, err := parseSettings(raw, defaults)
	if err != nil {
		return defaults, err
	}
	return settings, nil
}

func saveSettings(settings Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	return saveSettingsToPath(path, settings)
}

func saveSettingsToPath(path string, settings Settings) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, settingsDirMode); err != nil {
		return err
	}
	if err := restrictPermissions(dir, settingsDirMode); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, settingsFileMode); err != nil {
		return err
	}
	return restrictPermissions(path, settingsFileMode)
}

func restrictPermissions(path string, mode os.FileMode) error {
	if err := os.Chmod(path, mode); err != nil {
		return err
	}
	return nil
}

func defaultSettings() Settings {
	workspaceRoot, defaultOutputDir := defaultWorkspacePaths()
	return Settings{
		WorkspaceRoot:    workspaceRoot,
		DefaultOutputDir: defaultOutputDir,
		DefaultProfile:   "standard",
		Summary:          true,
		Runbook:          true,
		CSVExport:        true,
	}
}

func defaultWorkspacePaths() (string, string) {
	workspaceRoot := "."
	defaultOutputDir := "./out"

	homeDir, _ := os.UserHomeDir()
	configDir, _ := os.UserConfigDir()

	if goruntime.GOOS == "linux" {
		if root := linuxDefaultWorkspaceRoot(homeDir, configDir, os.Getenv); root != "" {
			return root, filepath.Join(root, "out")
		}
	}

	if homeDir != "" {
		workspaceRoot = filepath.Join(homeDir, "Documents", "k8s-recovery-visualizer")
		defaultOutputDir = filepath.Join(workspaceRoot, "out")
	}
	return workspaceRoot, defaultOutputDir
}

func linuxDefaultWorkspaceRoot(homeDir, configDir string, getenv func(string) string) string {
	if documentsDir := linuxDocumentsDir(homeDir, configDir, getenv); documentsDir != "" {
		return filepath.Join(documentsDir, "k8s-recovery-visualizer")
	}
	if dataHome := linuxDataHome(homeDir, getenv); dataHome != "" {
		return filepath.Join(dataHome, "k8s-recovery-visualizer", "workspace")
	}
	return ""
}

func linuxDocumentsDir(homeDir, configDir string, getenv func(string) string) string {
	if dir := cleanExpandedPath(getenv("XDG_DOCUMENTS_DIR"), homeDir); dir != "" {
		return dir
	}
	if configDir == "" {
		return ""
	}

	raw, err := os.ReadFile(filepath.Join(configDir, "user-dirs.dirs"))
	if err != nil {
		return ""
	}
	return parseXDGUserDir(raw, "XDG_DOCUMENTS_DIR", homeDir)
}

func linuxDataHome(homeDir string, getenv func(string) string) string {
	if dir := cleanExpandedPath(getenv("XDG_DATA_HOME"), homeDir); dir != "" {
		return dir
	}
	if homeDir == "" {
		return ""
	}
	return filepath.Join(homeDir, ".local", "share")
}

func parseXDGUserDir(raw []byte, key, homeDir string) string {
	lines := strings.Split(string(raw), "\n")
	prefix := key + "="
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.HasPrefix(line, prefix) {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		return cleanExpandedPath(value, homeDir)
	}
	return ""
}

func cleanExpandedPath(value, homeDir string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	if value == "" {
		return ""
	}

	replacer := strings.NewReplacer(
		"$HOME", homeDir,
		"${HOME}", homeDir,
	)
	value = replacer.Replace(value)
	value = filepath.FromSlash(value)
	if homeDir != "" && strings.HasPrefix(value, "~"+string(os.PathSeparator)) {
		value = filepath.Join(homeDir, strings.TrimPrefix(value, "~"+string(os.PathSeparator)))
	}
	if !filepath.IsAbs(value) {
		if homeDir == "" {
			return ""
		}
		value = filepath.Join(homeDir, value)
	}
	return filepath.Clean(value)
}

func parseSettings(raw []byte, defaults Settings) (Settings, error) {
	var partial partialSettings
	if err := json.Unmarshal(raw, &partial); err != nil {
		return Settings{}, err
	}
	return mergeSettings(defaults, partialSettingsToSettings(partial, defaults)), nil
}

func partialSettingsToSettings(partial partialSettings, defaults Settings) Settings {
	out := defaults
	if partial.WorkspaceRoot != nil {
		out.WorkspaceRoot = *partial.WorkspaceRoot
	}
	if partial.DefaultOutputDir != nil {
		out.DefaultOutputDir = *partial.DefaultOutputDir
	}
	if partial.DefaultProfile != nil {
		out.DefaultProfile = *partial.DefaultProfile
	}
	if partial.IncludeSecretMetadata != nil {
		out.IncludeSecretMetadata = *partial.IncludeSecretMetadata
	}
	if partial.Summary != nil {
		out.Summary = *partial.Summary
	}
	if partial.Runbook != nil {
		out.Runbook = *partial.Runbook
	}
	if partial.Redact != nil {
		out.Redact = *partial.Redact
	}
	if partial.CSVExport != nil {
		out.CSVExport = *partial.CSVExport
	}
	return out
}

func mergeSettings(defaults, overrides Settings) Settings {
	if overrides.WorkspaceRoot != "" {
		defaults.WorkspaceRoot = overrides.WorkspaceRoot
	}
	if overrides.DefaultOutputDir != "" {
		defaults.DefaultOutputDir = overrides.DefaultOutputDir
	}
	if overrides.DefaultProfile != "" {
		defaults.DefaultProfile = overrides.DefaultProfile
	}
	defaults.IncludeSecretMetadata = overrides.IncludeSecretMetadata
	defaults.Summary = overrides.Summary
	defaults.Runbook = overrides.Runbook
	defaults.Redact = overrides.Redact
	defaults.CSVExport = overrides.CSVExport
	return defaults
}

func startupSettingsMessage(err error) string {
	return fmt.Sprintf("Saved desktop settings could not be loaded, so K8V started with defaults. Review the Settings screen and save again after fixing the issue. Details: %v", err)
}

func (a *App) currentSettings() Settings {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()
	return a.settings
}

func (a *App) setSettings(settings Settings) {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	a.settings = settings
}

func (a *App) recordStartupAlert(alert AppAlert) {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	a.startupAlerts = append(a.startupAlerts, alert)
}
