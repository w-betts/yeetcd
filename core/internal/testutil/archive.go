package testutil

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateProjectZip creates a zip archive of a directory, excluding 'target' directories
// and hidden files (files starting with '.')
func CreateProjectZip(rootDir string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip 'target' directories
		if info.IsDir() && info.Name() == "target" {
			return filepath.SkipDir
		}

		// Skip compiled CLI binaries in resources/cli directories (large binaries not needed for tests)
		if info.IsDir() && (info.Name() == "cli") {
			// Also skip parent 'resources' directory - cli binaries shouldn't be in the zip
			relPath, err := filepath.Rel(rootDir, path)
			if err == nil && strings.Contains(relPath, "resources") {
				return filepath.SkipDir
			}
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories (we only add files)
		if info.IsDir() {
			return nil
		}

		// Create zip header
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header: %w", err)
		}
		header.Name = filepath.ToSlash(relPath)
		header.Method = zip.Deflate

		// Add file to zip
		w, err := zw.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip entry: %w", err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		_, err = io.Copy(w, file)
		if err != nil {
			return fmt.Errorf("failed to copy file to zip: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// CreateProjectZipFromCurrentRepo creates a zip of the repository root.
// The repository root is detected by finding the go.mod file.
func CreateProjectZipFromCurrentRepo() ([]byte, error) {
	// Start from current directory and walk up to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	rootDir, err := findRepoRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	return CreateProjectZip(rootDir)
}

// findRepoRoot walks up the directory tree looking for go.mod
func findRepoRoot(startDir string) (string, error) {
	current := startDir
	for {
		goModPath := filepath.Join(current, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		current = parent
	}
}
