package main

import "testing"

func TestParseSettingsMergesPartialFilesWithDefaults(t *testing.T) {
	defaults := defaultSettings()
	raw := []byte(`{"workspaceRoot":"C:/demo","summary":false}`)

	settings, err := parseSettings(raw, defaults)
	if err != nil {
		t.Fatalf("parseSettings() error = %v", err)
	}

	if settings.WorkspaceRoot != "C:/demo" {
		t.Fatalf("WorkspaceRoot = %q, want %q", settings.WorkspaceRoot, "C:/demo")
	}
	if settings.Summary {
		t.Fatal("Summary = true, want false from partial file")
	}
	if settings.Runbook != defaults.Runbook {
		t.Fatalf("Runbook = %v, want default %v", settings.Runbook, defaults.Runbook)
	}
	if settings.CSVExport != defaults.CSVExport {
		t.Fatalf("CSVExport = %v, want default %v", settings.CSVExport, defaults.CSVExport)
	}
}
