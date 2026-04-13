package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"k8s-recovery-visualizer/internal/appcore"
)

func (a *App) RunPreflight(req appcore.ScanRequest) (appcore.PreflightReport, error) {
	req = a.applyDefaults(req)

	preflightCtx, cancel := context.WithCancel(a.lifecycleContext())
	defer cancel()

	return a.service.Preflight(preflightCtx, req)
}

func (a *App) RunScan(req appcore.ScanRequest) (appcore.RunResult, error) {
	req = a.applyDefaults(req)
	if req.RunID == "" {
		req.RunID = fmt.Sprintf("run-%d", time.Now().UnixNano())
	}

	runCtx, cancel := context.WithCancel(a.lifecycleContext())
	if err := a.registerRun(req.RunID, cancel); err != nil {
		cancel()
		return appcore.RunResult{}, err
	}
	defer a.finishRun(req.RunID)

	result, err := a.service.Run(runCtx, req, eventSink(func(event appcore.RunEvent) {
		a.emitRunEvent(event)
	}))
	if err != nil {
		if errors.Is(err, context.Canceled) {
			a.emitRunEvent(appcore.RunEvent{
				Type:    "warning",
				RunID:   req.RunID,
				Step:    "cancel",
				Level:   "warn",
				Message: "Scan canceled.",
				Warning: "The active desktop run was canceled before completion.",
			})
		}
		return appcore.RunResult{}, err
	}
	return result, nil
}

func (a *App) CancelRun(runID string) error {
	cancel, ok := a.takeRunCancel(runID)
	if !ok {
		return fmt.Errorf("run %s is not active", runID)
	}
	cancel()
	return nil
}

func (a *App) ExportBundle(path string, req appcore.ExportRequest) (appcore.ArtifactPaths, error) {
	return a.service.ExportBundle(path, req)
}

func (a *App) applyDefaults(req appcore.ScanRequest) appcore.ScanRequest {
	settings := a.currentSettings()
	if req.OutputDir == "" {
		req.OutputDir = settings.DefaultOutputDir
	}
	if req.ProfileName == "" {
		req.ProfileName = settings.DefaultProfile
	}
	return req.Normalized()
}

func (a *App) emitRunEvent(event appcore.RunEvent) {
	if a.ctx == nil || a.ctx.Value("events") == nil {
		return
	}
	runtime.EventsEmit(a.ctx, scanEventName, event)
}

func (a *App) logError(scope string, err error) {
	message := fmt.Sprintf("%s: %v", scope, err)
	log.Printf("desktop error: %s", message)
	if a.ctx != nil && a.ctx.Value("logger") != nil {
		runtime.LogError(a.ctx, message)
	}
}

func (a *App) logWarning(scope string, err error) {
	message := fmt.Sprintf("%s: %v", scope, err)
	log.Printf("desktop warning: %s", message)
	if a.ctx != nil && a.ctx.Value("logger") != nil {
		runtime.LogWarning(a.ctx, message)
	}
}

func (a *App) setLifecycleContext(ctx context.Context) {
	a.lifecycleMu.Lock()
	defer a.lifecycleMu.Unlock()

	if a.lifecycleCancel != nil {
		a.lifecycleCancel()
	}
	a.lifecycleCtx, a.lifecycleCancel = context.WithCancel(ctx)
}

func (a *App) lifecycleContext() context.Context {
	a.lifecycleMu.RLock()
	defer a.lifecycleMu.RUnlock()

	if a.lifecycleCtx != nil {
		return a.lifecycleCtx
	}
	return context.Background()
}

func (a *App) cancelLifecycle() {
	a.lifecycleMu.Lock()
	cancel := a.lifecycleCancel
	a.lifecycleCancel = nil
	a.lifecycleCtx = nil
	a.lifecycleMu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (a *App) registerRun(runID string, cancel context.CancelFunc) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.cancels[runID]; exists {
		return fmt.Errorf("run %s is already active", runID)
	}
	a.cancels[runID] = cancel
	return nil
}

func (a *App) takeRunCancel(runID string) (context.CancelFunc, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cancel, ok := a.cancels[runID]
	if ok {
		delete(a.cancels, runID)
	}
	return cancel, ok
}

func (a *App) finishRun(runID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.cancels, runID)
}

func (a *App) cancelAllRuns() {
	a.mu.Lock()
	cancels := make([]context.CancelFunc, 0, len(a.cancels))
	for runID, cancel := range a.cancels {
		cancels = append(cancels, cancel)
		delete(a.cancels, runID)
	}
	a.mu.Unlock()

	for _, cancel := range cancels {
		cancel()
	}
}

type eventSink func(appcore.RunEvent)

func (sink eventSink) Emit(event appcore.RunEvent) {
	sink(event)
}
