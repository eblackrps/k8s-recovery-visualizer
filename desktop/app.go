package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"k8s-recovery-visualizer/internal/appcore"
)

const scanEventName = "scan:event"

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

type App struct {
	ctx      context.Context
	service  *appcore.Service
	settings Settings

	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func NewApp() *App {
	return &App{
		service:  appcore.NewService(),
		cancels:  map[string]context.CancelFunc{},
		settings: defaultSettings(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if settings, err := loadSettings(); err == nil {
		a.settings = settings
	}
}

func (a *App) GetBootstrap() appcore.Bootstrap {
	return a.service.Bootstrap()
}

func (a *App) GetSettings() Settings {
	return a.settings
}

func (a *App) SaveSettings(settings Settings) error {
	a.settings = settings
	return saveSettings(settings)
}

func (a *App) ListProjects(root string) ([]appcore.ProjectSummary, error) {
	if root == "" {
		root = a.settings.WorkspaceRoot
	}
	return a.service.ListProjects(root)
}

func (a *App) RunPreflight(req appcore.ScanRequest) (appcore.PreflightReport, error) {
	req = a.applyDefaults(req)
	return a.service.Preflight(context.Background(), req)
}

func (a *App) RunScan(req appcore.ScanRequest) (appcore.RunResult, error) {
	req = a.applyDefaults(req)
	if req.RunID == "" {
		req.RunID = fmt.Sprintf("run-%d", time.Now().UnixNano())
	}
	runCtx, cancel := context.WithCancel(context.Background())
	a.mu.Lock()
	a.cancels[req.RunID] = cancel
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		delete(a.cancels, req.RunID)
		a.mu.Unlock()
	}()

	return a.service.Run(runCtx, req, eventSink(func(event appcore.RunEvent) {
		runtime.EventsEmit(a.ctx, scanEventName, event)
	}))
}

func (a *App) CancelRun(runID string) error {
	a.mu.Lock()
	cancel, ok := a.cancels[runID]
	a.mu.Unlock()
	if !ok {
		return fmt.Errorf("run %s is not active", runID)
	}
	cancel()
	return nil
}

func (a *App) OpenBundle(path string) (appcore.Workspace, error) {
	return a.service.LoadWorkspace(path)
}

func (a *App) ExportBundle(path string, req appcore.ExportRequest) (appcore.ArtifactPaths, error) {
	return a.service.ExportBundle(path, req)
}

func (a *App) PickBundleFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open recovery bundle",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON", Pattern: "*.json"},
		},
	})
}

func (a *App) PickOutputDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose output directory",
	})
}

func (a *App) applyDefaults(req appcore.ScanRequest) appcore.ScanRequest {
	if req.OutputDir == "" {
		req.OutputDir = a.settings.DefaultOutputDir
	}
	if req.ProfileName == "" {
		req.ProfileName = a.settings.DefaultProfile
	}
	return req.Normalized()
}

type eventSink func(appcore.RunEvent)

func (sink eventSink) Emit(event appcore.RunEvent) {
	sink(event)
}

func settingsPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "k8s-recovery-visualizer", "settings.json"), nil
}

func loadSettings() (Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return Settings{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}
	var settings Settings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

func saveSettings(settings Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func defaultSettings() Settings {
	workspaceRoot := "."
	defaultOutputDir := "./out"
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		workspaceRoot = filepath.Join(home, "Documents", "k8s-recovery-visualizer")
		defaultOutputDir = filepath.Join(workspaceRoot, "out")
	}
	return Settings{
		WorkspaceRoot:    workspaceRoot,
		DefaultOutputDir: defaultOutputDir,
		DefaultProfile:   "standard",
		Summary:          true,
		Runbook:          true,
		CSVExport:        true,
	}
}
