package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	// Cache for java-sample zip to avoid repeated builds
	javaSampleZipCache     []byte
	javaSampleZipCacheOnce sync.Once
	javaSampleZipCacheErr  error

	// Cache for full repo zip
	javaSampleZipWithRepoCache     []byte
	javaSampleZipWithRepoCacheOnce sync.Once
	javaSampleZipWithRepoCacheErr  error
)

// GetJavaSampleZip returns a cached zip of the java-sample project.
// This only includes the java-sample directory, not its dependencies.
func GetJavaSampleZip() ([]byte, error) {
	javaSampleZipCacheOnce.Do(func() {
		javaSampleDir := filepath.Join("java-sample")
		if _, err := os.Stat(javaSampleDir); os.IsNotExist(err) {
			// Try from repo root
			javaSampleDir = filepath.Join("..", "java-sample")
			if _, err := os.Stat(javaSampleDir); os.IsNotExist(err) {
				javaSampleZipCacheErr = fmt.Errorf("java-sample directory not found")
				return
			}
		}

		javaSampleZipCache, javaSampleZipCacheErr = CreateProjectZip(javaSampleDir)
	})

	return javaSampleZipCache, javaSampleZipCacheErr
}

// GetJavaSampleZipWithRepo returns a cached zip of the entire repository.
// This is needed because java-sample depends on java-sdk and protocol modules.
func GetJavaSampleZipWithRepo() ([]byte, error) {
	javaSampleZipWithRepoCacheOnce.Do(func() {
		var err error
		javaSampleZipWithRepoCache, err = CreateProjectZipFromCurrentRepo()
		if err != nil {
			javaSampleZipWithRepoCacheErr = fmt.Errorf("failed to create repo zip: %w", err)
		}
	})

	return javaSampleZipWithRepoCache, javaSampleZipWithRepoCacheErr
}

// GetJavaSamplePath returns the path to the java-sample directory
func GetJavaSamplePath() (string, error) {
	// Try current directory first
	if _, err := os.Stat("java-sample"); err == nil {
		return "java-sample", nil
	}

	// Try parent directory
	if _, err := os.Stat("../java-sample"); err == nil {
		return "../java-sample", nil
	}

	// Try to find from current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up looking for java-sample
	for {
		javaSamplePath := filepath.Join(dir, "java-sample")
		if _, err := os.Stat(javaSamplePath); err == nil {
			return javaSamplePath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("java-sample directory not found")
}
