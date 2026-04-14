package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeOpenPathResolvesToAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "recovery-report.html")
	if err := os.WriteFile(filePath, []byte("<html></html>"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	resolved, err := normalizeOpenPath(filePath)
	if err != nil {
		t.Fatalf("normalizeOpenPath() error = %v", err)
	}
	if !filepath.IsAbs(resolved) {
		t.Fatalf("normalizeOpenPath() = %q, want absolute path", resolved)
	}
}

func TestOpenPathCommandChoosesPlatformLauncher(t *testing.T) {
	tests := []struct {
		goos        string
		wantCommand string
		wantArgs    []string
	}{
		{goos: "windows", wantCommand: "cmd", wantArgs: []string{"/c", "start", "", "C:/demo/report.html"}},
		{goos: "linux", wantCommand: "xdg-open", wantArgs: []string{"/tmp/report.html"}},
		{goos: "darwin", wantCommand: "open", wantArgs: []string{"/tmp/report.html"}},
	}

	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			target := "/tmp/report.html"
			if tt.goos == "windows" {
				target = "C:/demo/report.html"
			}

			command, args, err := openPathCommand(tt.goos, target)
			if err != nil {
				t.Fatalf("openPathCommand() error = %v", err)
			}
			if command != tt.wantCommand {
				t.Fatalf("command = %q, want %q", command, tt.wantCommand)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args = %#v, want %#v", args, tt.wantArgs)
			}
			for index := range args {
				if args[index] != tt.wantArgs[index] {
					t.Fatalf("args = %#v, want %#v", args, tt.wantArgs)
				}
			}
		})
	}
}
