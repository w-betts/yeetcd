package build

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestZipExtractor_Extract_CreatesDirectoryStructure tests that Extract creates directory structure from zip
// Given: Zip data containing 'dir1/dir2/file.txt' with content 'hello'
// When: Extract(zipData, destDir) is called
// Then: Directory 'destDir/dir1/dir2/' is created, file 'destDir/dir1/dir2/file.txt' contains 'hello'
func TestZipExtractor_Extract_CreatesDirectoryStructure(t *testing.T) {
	// Create a temporary destination directory
	destDir := t.TempDir()

	// Create zip data with nested directories
	zipData := createTestZipForExtractor(map[string][]byte{
		"dir1/dir2/file.txt": []byte("hello"),
	})

	// Create extractor
	extractor := NewZipExtractor()

	// Extract the zip - should fail with "not implemented" since it's a stub
	err := extractor.Extract(zipData, destDir)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
}

// TestZipExtractor_Extract_CallsFileHandler tests that Extract calls FileHandler for matching files
// Given: Zip data containing 'yeetcd.yaml' with content 'name: test', FileHandler with ShouldHandle=(name == 'yeetcd.yaml')
// When: Extract(zipData, destDir, handler) is called
// Then: FileHandler.Handle is called with HandledFile{Parent: 'path/to/parent', Contents: []byte('name: test')}
func TestZipExtractor_Extract_CallsFileHandler(t *testing.T) {
	// Create a temporary destination directory
	destDir := t.TempDir()

	// Create zip data with yeetcd.yaml
	yeetcdContent := []byte("name: test")
	zipData := createTestZipForExtractor(map[string][]byte{
		"yeetcd.yaml": yeetcdContent,
	})

	// Track handled files
	var handledFiles []string
	var handledContents [][]byte

	// Create file handler
	handler := FileHandler{
		ShouldHandle: func(name string) bool {
			return name == "yeetcd.yaml"
		},
		Handle: func(parent string, contents []byte) error {
			handledFiles = append(handledFiles, parent)
			handledContents = append(handledContents, contents)
			return nil
		},
	}

	// Create extractor
	extractor := NewZipExtractor()

	// Extract the zip with handler - should fail with "not implemented" since it's a stub
	err := extractor.Extract(zipData, destDir, handler)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
}

// TestZipExtractor_Extract_InvalidZip tests extraction with invalid zip data
// Given: Invalid zip data
// When: Extract is called
// Then: It should return an error indicating the zip is invalid
func TestZipExtractor_Extract_InvalidZip(t *testing.T) {
	// Create a temporary destination directory
	destDir := t.TempDir()

	// Create invalid zip data
	invalidZipData := []byte("this is not a zip file")

	// Create extractor
	extractor := NewZipExtractor()

	// Try to extract - should fail with "not implemented" since it's a stub
	err := extractor.Extract(invalidZipData, destDir)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
}

// TestZipExtractor_Extract_WritesFileContents tests that Extract writes file contents correctly
// Given: Zip data containing 'main.go' with content 'package main'
// When: Extract(zipData, destDir) is called
// Then: File 'destDir/main.go' contains 'package main'
func TestZipExtractor_Extract_WritesFileContents(t *testing.T) {
	// Create a temporary destination directory
	destDir := t.TempDir()

	// Create zip data with a Go file
	zipData := createTestZipForExtractor(map[string][]byte{
		"main.go": []byte("package main\n\nfunc main() {}"),
	})

	// Create extractor
	extractor := NewZipExtractor()

	// Extract the zip - should fail with "not implemented" since it's a stub
	err := extractor.Extract(zipData, destDir)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
}

// TestZipExtractor_Extract_HandlesNestedDirectories tests that Extract handles nested directories
// Given: Zip data containing nested directory structure
// When: Extract(zipData, destDir) is called
// Then: All directories are created correctly
func TestZipExtractor_Extract_HandlesNestedDirectories(t *testing.T) {
	// Create a temporary destination directory
	destDir := t.TempDir()

	// Create zip data with deeply nested structure
	zipData := createTestZipForExtractor(map[string][]byte{
		"a/b/c/d/file.txt": []byte("deep nested file"),
		"a/b/other.txt":    []byte("other file"),
		"root.txt":         []byte("root file"),
	})

	// Create extractor
	extractor := NewZipExtractor()

	// Extract the zip - should fail with "not implemented" since it's a stub
	err := extractor.Extract(zipData, destDir)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
}

// createTestZipForExtractor creates a zip archive from the given files
func createTestZipForExtractor(files map[string][]byte) []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for name, content := range files {
		w, _ := zipWriter.Create(name)
		w.Write(content)
	}

	zipWriter.Close()
	return buf.Bytes()
}
