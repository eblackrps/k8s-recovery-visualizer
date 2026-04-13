package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"k8s-recovery-visualizer/internal/appcore"
)

type blockingDesktopService struct {
	started chan string
}

func (s *blockingDesktopService) Bootstrap() appcore.Bootstrap {
	return appcore.Bootstrap{}
}

func (s *blockingDesktopService) ListProjects(string) ([]appcore.ProjectSummary, error) {
	return nil, nil
}

func (s *blockingDesktopService) Preflight(ctx context.Context, req appcore.ScanRequest) (appcore.PreflightReport, error) {
	return appcore.PreflightReport{CanRun: ctx.Err() == nil, Scope: req.OutputDir}, ctx.Err()
}

func (s *blockingDesktopService) Run(ctx context.Context, req appcore.ScanRequest, sink appcore.EventSink) (appcore.RunResult, error) {
	s.started <- req.RunID
	<-ctx.Done()
	return appcore.RunResult{}, ctx.Err()
}

func (s *blockingDesktopService) LoadWorkspace(path string) (appcore.Workspace, error) {
	return appcore.Workspace{}, nil
}

func (s *blockingDesktopService) ExportBundle(path string, req appcore.ExportRequest) (appcore.ArtifactPaths, error) {
	return appcore.ArtifactPaths{}, nil
}

func TestCancelRunRemovesActiveRunAndCancelsContext(t *testing.T) {
	service := &blockingDesktopService{started: make(chan string, 1)}
	app := NewApp()
	app.service = service
	app.setLifecycleContext(context.Background())

	errCh := make(chan error, 1)
	go func() {
		_, err := app.RunScan(appcore.ScanRequest{RunID: "run-1"})
		errCh <- err
	}()

	select {
	case gotRunID := <-service.started:
		if gotRunID != "run-1" {
			t.Fatalf("RunID = %q, want %q", gotRunID, "run-1")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for RunScan() to start")
	}

	if got := activeRunCount(app); got != 1 {
		t.Fatalf("activeRunCount() = %d, want 1", got)
	}
	if err := app.CancelRun("run-1"); err != nil {
		t.Fatalf("CancelRun() error = %v", err)
	}
	if got := activeRunCount(app); got != 0 {
		t.Fatalf("activeRunCount() after cancel = %d, want 0", got)
	}

	if err := <-errCh; !errors.Is(err, context.Canceled) {
		t.Fatalf("RunScan() error = %v, want context.Canceled", err)
	}
}

func TestShutdownCancelsActiveRuns(t *testing.T) {
	service := &blockingDesktopService{started: make(chan string, 1)}
	app := NewApp()
	app.service = service
	app.setLifecycleContext(context.Background())

	errCh := make(chan error, 1)
	go func() {
		_, err := app.RunScan(appcore.ScanRequest{RunID: "run-2"})
		errCh <- err
	}()

	<-service.started
	app.shutdown(context.Background())

	if got := activeRunCount(app); got != 0 {
		t.Fatalf("activeRunCount() after shutdown = %d, want 0", got)
	}
	if err := <-errCh; !errors.Is(err, context.Canceled) {
		t.Fatalf("RunScan() error = %v, want context.Canceled", err)
	}
}

func TestRunPreflightRequiresLifecycleContext(t *testing.T) {
	app := NewApp()
	app.service = &blockingDesktopService{started: make(chan string, 1)}

	_, err := app.RunPreflight(appcore.ScanRequest{})
	if !errors.Is(err, errLifecycleContextUnavailable) {
		t.Fatalf("RunPreflight() error = %v, want %v", err, errLifecycleContextUnavailable)
	}
}

func TestRunScanRequiresLifecycleContext(t *testing.T) {
	app := NewApp()
	app.service = &blockingDesktopService{started: make(chan string, 1)}

	_, err := app.RunScan(appcore.ScanRequest{RunID: "run-no-lifecycle"})
	if !errors.Is(err, errLifecycleContextUnavailable) {
		t.Fatalf("RunScan() error = %v, want %v", err, errLifecycleContextUnavailable)
	}
	if got := activeRunCount(app); got != 0 {
		t.Fatalf("activeRunCount() = %d, want 0", got)
	}
}

func activeRunCount(app *App) int {
	app.mu.Lock()
	defer app.mu.Unlock()
	return len(app.cancels)
}
