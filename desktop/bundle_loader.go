package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"k8s-recovery-visualizer/internal/appcore"
)

func (a *App) OpenBundle(path string) (appcore.Workspace, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return appcore.Workspace{}, fmt.Errorf("bundle path is required")
	}

	resolvedPath, err := a.resolveBundlePath(path)
	if err != nil {
		return appcore.Workspace{}, err
	}
	return a.service.LoadWorkspace(resolvedPath)
}

func (a *App) resolveBundlePath(path string) (string, error) {
	resolved := path
	if !filepath.IsAbs(resolved) {
		base := a.currentSettings().WorkspaceRoot
		if base == "" {
			if cwd, err := os.Getwd(); err == nil {
				base = cwd
			}
		}
		resolved = filepath.Join(base, resolved)
	}
	if absolutePath, err := filepath.Abs(resolved); err == nil {
		resolved = absolutePath
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return findBundleJSON(resolved)
	}
	if !isBundleArchive(resolved) {
		return resolved, nil
	}

	bundlePath, extractDir, err := extractBundleArchive(resolved)
	if err != nil {
		return "", err
	}
	a.trackExtractedBundleDir(extractDir)
	return bundlePath, nil
}

func (a *App) trackExtractedBundleDir(path string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.extractedBundleDirs = append(a.extractedBundleDirs, path)
}

func (a *App) cleanupExtractedBundleDirs() {
	a.mu.Lock()
	dirs := append([]string(nil), a.extractedBundleDirs...)
	a.extractedBundleDirs = nil
	a.mu.Unlock()

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			a.logWarning("clean extracted bundle directory", err)
		}
	}
}

func isBundleArchive(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".zip") ||
		strings.HasSuffix(lower, ".tar.gz") ||
		strings.HasSuffix(lower, ".tgz") ||
		strings.HasSuffix(lower, ".tar")
}

func extractBundleArchive(path string) (bundlePath, extractDir string, err error) {
	extractDir, err = os.MkdirTemp("", "k8v-bundle-*")
	if err != nil {
		return "", "", err
	}

	if err := unpackBundleArchive(path, extractDir); err != nil {
		_ = os.RemoveAll(extractDir)
		return "", "", err
	}

	bundlePath, err = findBundleJSON(extractDir)
	if err != nil {
		_ = os.RemoveAll(extractDir)
		return "", "", err
	}
	return bundlePath, extractDir, nil
}

func unpackBundleArchive(path, extractDir string) error {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return extractZipArchive(path, extractDir)
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return extractTarArchive(path, extractDir, true)
	case strings.HasSuffix(lower, ".tar"):
		return extractTarArchive(path, extractDir, false)
	default:
		return fmt.Errorf("unsupported bundle archive format: %s", filepath.Base(path))
	}
}

func extractZipArchive(path, extractDir string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		target, err := safeArchivePath(extractDir, file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, settingsDirMode); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), settingsDirMode); err != nil {
			return err
		}
		in, err := file.Open()
		if err != nil {
			return err
		}
		if err := writeArchiveFile(target, in); err != nil {
			in.Close()
			return err
		}
		if err := in.Close(); err != nil {
			return err
		}
	}
	return nil
}

func extractTarArchive(path, extractDir string, compressed bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var reader io.Reader = file
	if compressed {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		target, err := safeArchivePath(extractDir, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, settingsDirMode); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), settingsDirMode); err != nil {
				return err
			}
			if err := writeArchiveFile(target, tarReader); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported archive entry %q in %s", header.Name, filepath.Base(path))
		}
	}
}

func writeArchiveFile(path string, reader io.Reader) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, settingsFileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return err
	}
	return nil
}

func safeArchivePath(root, name string) (string, error) {
	target := filepath.Clean(filepath.Join(root, filepath.FromSlash(name)))
	cleanRoot := filepath.Clean(root)
	if target != cleanRoot && !strings.HasPrefix(target, cleanRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry %q escapes the extraction directory", name)
	}
	return target, nil
}

func findBundleJSON(root string) (string, error) {
	preferred := filepath.Join(root, "recovery-scan.json")
	if info, err := os.Stat(preferred); err == nil && !info.IsDir() {
		return preferred, nil
	}

	matches := []string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Base(path), "recovery-scan.json") {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("recovery-scan.json was not found in %s", root)
	}

	sort.Strings(matches)
	return matches[0], nil
}
