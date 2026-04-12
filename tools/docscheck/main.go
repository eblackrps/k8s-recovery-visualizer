package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var linkPattern = regexp.MustCompile(`!?\[[^\]]+\]\(([^)]+)\)`)

func main() {
	files, err := markdownFiles(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "docscheck: %v\n", err)
		os.Exit(1)
	}

	var missing []string
	for _, path := range files {
		refs, err := references(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "docscheck: %v\n", err)
			os.Exit(1)
		}
		for _, ref := range refs {
			if skipReference(ref.target) {
				continue
			}
			target := normalizeTarget(ref.target)
			if target == "" {
				continue
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(path), target))
			if _, err := os.Stat(resolved); err != nil {
				missing = append(missing, fmt.Sprintf("%s:%d -> %s", path, ref.line, ref.target))
			}
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		fmt.Fprintln(os.Stderr, "docscheck: missing references detected:")
		for _, item := range missing {
			fmt.Fprintln(os.Stderr, "  "+item)
		}
		os.Exit(1)
	}

	fmt.Printf("docscheck: validated %d markdown files\n", len(files))
}

type ref struct {
	line   int
	target string
}

func markdownFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".md") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func references(path string) ([]ref, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var refs []ref
	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		matches := linkPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			refs = append(refs, ref{line: lineNo, target: match[1]})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return refs, nil
}

func skipReference(target string) bool {
	trimmed := strings.TrimSpace(target)
	switch {
	case trimmed == "":
		return true
	case strings.HasPrefix(trimmed, "#"):
		return true
	case strings.HasPrefix(trimmed, "http://"):
		return true
	case strings.HasPrefix(trimmed, "https://"):
		return true
	case strings.HasPrefix(trimmed, "mailto:"):
		return true
	default:
		return false
	}
}

func normalizeTarget(target string) string {
	trimmed := strings.TrimSpace(target)
	if idx := strings.Index(trimmed, " \""); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") {
		trimmed = strings.TrimSuffix(strings.TrimPrefix(trimmed, "<"), ">")
	}
	if idx := strings.Index(trimmed, "#"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	return strings.TrimSpace(trimmed)
}
