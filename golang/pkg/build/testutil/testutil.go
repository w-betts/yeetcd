package testutil

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// CreateTestZip creates a test zip file from a map of file names to contents
func CreateTestZip(files map[string][]byte) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	for name, contents := range files {
		f, err := w.Create(name)
		if err != nil {
			panic(err)
		}
		_, err = f.Write(contents)
		if err != nil {
			panic(err)
		}
	}

	err := w.Close()
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "yeetcd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

// CreateTestFile creates a test file with the given contents
func CreateTestFile(dir, name, contents string) string {
	path := filepath.Join(dir, name)
	
	// Create parent directories if needed
	parent := filepath.Dir(path)
	if parent != dir {
		os.MkdirAll(parent, 0755)
	}
	
	err := os.WriteFile(path, []byte(contents), 0644)
	if err != nil {
		panic(err)
	}
	
	return path
}
