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

	// Cache for sdks/java zip (sample + SDK dependencies)
	javaSampleZipWithSdkCache     []byte
	javaSampleZipWithSdkCacheOnce sync.Once
	javaSampleZipWithSdkCacheErr  error
)

// findProjectRoot attempts to find the project root by looking for the sdks/java directory
// from various possible starting points (current directory, core module root, etc.)
func findProjectRoot() (string, error) {
	// Check current working directory first
	if _, err := os.Stat("sdks/java"); err == nil {
		return ".", nil
	}

	// Check if we're inside the core module - go test runs from temp dir
	// so we need to find the core module's directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up looking for the project root
	// The project root contains both 'core' and 'sdks' directories
	for {
		// Check if this looks like project root (has core AND sdks)
		if _, err := os.Stat(filepath.Join(dir, "core")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "sdks")); err == nil {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found")
}

// GetJavaSampleZip returns a cached zip of the java-sample project.
// This only includes the java-sample directory, not its dependencies.
func GetJavaSampleZip() ([]byte, error) {
	javaSampleZipCacheOnce.Do(func() {
		projectRoot, err := findProjectRoot()
		if err != nil {
			javaSampleZipCacheErr = fmt.Errorf("failed to find project root: %w", err)
			return
		}

		javaSampleDir := filepath.Join(projectRoot, "sdks", "java", "sample")
		if _, err := os.Stat(javaSampleDir); os.IsNotExist(err) {
			javaSampleZipCacheErr = fmt.Errorf("java-sample directory not found at %s", javaSampleDir)
			return
		}

		javaSampleZipCache, javaSampleZipCacheErr = CreateProjectZip(javaSampleDir)
	})

	return javaSampleZipCache, javaSampleZipCacheErr
}

// GetJavaSampleZipWithSdk returns a cached zip of sdks/java directory.
// This includes the sample and the SDK dependencies (sdk, protocol, test modules).
func GetJavaSampleZipWithSdk() ([]byte, error) {
	javaSampleZipWithSdkCacheOnce.Do(func() {
		projectRoot, err := findProjectRoot()
		if err != nil {
			javaSampleZipWithSdkCacheErr = fmt.Errorf("failed to find project root: %w", err)
			return
		}

		sdksJavaDir := filepath.Join(projectRoot, "sdks", "java")
		if _, err := os.Stat(sdksJavaDir); os.IsNotExist(err) {
			javaSampleZipWithSdkCacheErr = fmt.Errorf("sdks/java directory not found at %s", sdksJavaDir)
			return
		}

		javaSampleZipWithSdkCache, javaSampleZipWithSdkCacheErr = CreateProjectZip(sdksJavaDir)
	})

	return javaSampleZipWithSdkCache, javaSampleZipWithSdkCacheErr
}

// GetJavaSampleZipWithRepo is an alias for GetJavaSampleZipWithSdk for backwards compatibility.
// It returns a zip of sdks/java (sample + SDK dependencies).
func GetJavaSampleZipWithRepo() ([]byte, error) {
	return GetJavaSampleZipWithSdk()
}

// GetJavaSamplePath returns the path to the java-sample directory
func GetJavaSamplePath() (string, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", fmt.Errorf("failed to find project root: %w", err)
	}

	return filepath.Join(projectRoot, "sdks", "java", "sample"), nil
}
