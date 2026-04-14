package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func (a *App) OpenPath(path string) error {
	target, err := normalizeOpenPath(path)
	if err != nil {
		return err
	}

	command, args, err := openPathCommand(runtime.GOOS, target)
	if err != nil {
		return err
	}

	cmd := exec.Command(command, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open %q: %w", target, err)
	}
	return nil
}

func normalizeOpenPath(path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", fmt.Errorf("choose a file or directory first")
	}

	target := trimmed
	if absolute, err := filepath.Abs(trimmed); err == nil {
		target = filepath.Clean(absolute)
	}

	if _, err := os.Stat(target); err != nil {
		return "", fmt.Errorf("open %q: %w", target, err)
	}
	return target, nil
}

func openPathCommand(goos, target string) (string, []string, error) {
	switch goos {
	case "windows":
		return "cmd", []string{"/c", "start", "", target}, nil
	case "linux":
		return "xdg-open", []string{target}, nil
	case "darwin":
		return "open", []string{target}, nil
	default:
		return "", nil, fmt.Errorf("opening files is not supported on %s", goos)
	}
}
