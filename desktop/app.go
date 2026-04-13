package main

import (
	"context"
	"sync"

	"k8s-recovery-visualizer/internal/appcore"
)

const scanEventName = "scan:event"

const (
	preferredWindowWidth  = 1240
	preferredWindowHeight = 720
	minimumWindowWidth    = 960
	minimumWindowHeight   = 620
)

type desktopService interface {
	Bootstrap() appcore.Bootstrap
	ListProjects(root string) ([]appcore.ProjectSummary, error)
	Preflight(ctx context.Context, req appcore.ScanRequest) (appcore.PreflightReport, error)
	Run(ctx context.Context, req appcore.ScanRequest, sink appcore.EventSink) (appcore.RunResult, error)
	LoadWorkspace(path string) (appcore.Workspace, error)
	ExportBundle(path string, req appcore.ExportRequest) (appcore.ArtifactPaths, error)
}

type AppAlert struct {
	Tone    string `json:"tone"`
	Message string `json:"message"`
}

type App struct {
	ctx context.Context

	lifecycleMu     sync.RWMutex
	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc

	service desktopService

	stateMu       sync.RWMutex
	settings      Settings
	startupAlerts []AppAlert

	mu                  sync.Mutex
	cancels             map[string]context.CancelFunc
	extractedBundleDirs []string
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
	a.setLifecycleContext(ctx)

	defaults := defaultSettings()
	settings, err := loadSettings(defaults)
	a.setSettings(settings)
	if err != nil {
		a.recordStartupAlert(AppAlert{
			Tone:    "error",
			Message: startupSettingsMessage(err),
		})
		a.logError("load desktop settings", err)
	}
}

func (a *App) domReady(ctx context.Context) {
	a.fitWindowToScreen(ctx)
}

func (a *App) shutdown(context.Context) {
	a.cancelAllRuns()
	a.cancelLifecycle()
	a.cleanupExtractedBundleDirs()
}

func (a *App) GetBootstrap() appcore.Bootstrap {
	return a.service.Bootstrap()
}
