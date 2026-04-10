package build

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yeetcd/yeetcd/pkg/config"
)

// createTestZipBytes creates a zip archive from the given files
func createTestZipBytes(files map[string][]byte) []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for name, content := range files {
		w, _ := zipWriter.Create(name)
		w.Write(content)
	}

	zipWriter.Close()
	return buf.Bytes()
}

// createTestDirectory creates a temporary directory with the given files
func createTestDirectory(files map[string][]byte) (string, error) {
	dir, err := os.MkdirTemp("", "yeetcd-test-")
	if err != nil {
		return "", err
	}

	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
	}

	return dir, nil
}

// TestSourceExtractor_Extract_CreatesTempDirectoryAndExtractsZip tests that Extract creates temp directory and extracts zip
// Given: Source with Name='test' and Zip containing 'yeetcd.yaml' with valid config
// When: Extract(source) is called
// Then: SourceExtractionResult.Directory is created, SourceExtractionResult.YeetcdDefinitions contains parsed config
func TestSourceExtractor_Extract_CreatesTempDirectoryAndExtractsZip(t *testing.T) {
	// Create a source with zip containing yeetcd.yaml
	yeetcdYaml := `name: test-project
language: JAVA
buildImage: maven:3.9.9
buildCmd: mvn package
artifacts:
  - name: classes
    path: target/classes
`
	zipData := createTestZipBytes(map[string][]byte{
		"yeetcd.yaml": []byte(yeetcdYaml),
	})

	source := Source{
		Name: "test",
		Zip:  zipData,
	}

	// Create extractor
	extractor := NewSourceExtractor()

	// Extract should succeed
	result, err := extractor.Extract(source)
	require.NoError(t, err)
	require.NotNil(t, result)
	defer result.Close()

	// Verify directory was created
	assert.DirExists(t, result.Directory)

	// Verify yeetcd.yaml was extracted
	assert.Contains(t, result.YeetcdDefinitions, ".")
	assert.Equal(t, "test-project", result.YeetcdDefinitions["."].Name)
	assert.Equal(t, config.SourceLanguageJava, result.YeetcdDefinitions["."].Language)
}

// TestSourceExtractor_Extract_ParsesMultipleYeetcdYamlFiles tests that Extract parses multiple yeetcd.yaml files
// Given: Source with Zip containing 'project1/yeetcd.yaml' and 'project2/yeetcd.yaml'
// When: Extract(source) is called
// Then: SourceExtractionResult.YeetcdDefinitions contains entries for 'project1/yeetcd.yaml' and 'project2/yeetcd.yaml'
func TestSourceExtractor_Extract_ParsesMultipleYeetcdYamlFiles(t *testing.T) {
	// Create a source with zip containing multiple yeetcd.yaml files
	yeetcdYaml1 := `name: project1
language: JAVA
buildImage: maven:3.9.9
buildCmd: mvn package
artifacts:
  - name: classes
    path: target/classes
`
	yeetcdYaml2 := `name: project2
language: JAVA
buildImage: maven:3.9.9
buildCmd: mvn package
artifacts:
  - name: classes
    path: target/classes
`
	zipData := createTestZipBytes(map[string][]byte{
		"project1/yeetcd.yaml": []byte(yeetcdYaml1),
		"project2/yeetcd.yaml": []byte(yeetcdYaml2),
	})

	source := Source{
		Name: "test",
		Zip:  zipData,
	}

	// Create extractor
	extractor := NewSourceExtractor()

	// Extract should succeed
	result, err := extractor.Extract(source)
	require.NoError(t, err)
	require.NotNil(t, result)
	defer result.Close()

	// Verify both configs were parsed
	assert.Len(t, result.YeetcdDefinitions, 2)
	assert.Contains(t, result.YeetcdDefinitions, "project1")
	assert.Contains(t, result.YeetcdDefinitions, "project2")
	assert.Equal(t, "project1", result.YeetcdDefinitions["project1"].Name)
	assert.Equal(t, "project2", result.YeetcdDefinitions["project2"].Name)
}

// TestSourceExtractor_Extract_WithEmptyZip tests extraction with empty zip
// Given: Source with empty Zip data
// When: Extract(source) is called
// Then: It should return an error
func TestSourceExtractor_Extract_WithEmptyZip(t *testing.T) {
	// Create a source with empty zip
	source := Source{
		Name: "test",
		Zip:  []byte{},
	}

	// Create extractor
	extractor := NewSourceExtractor()

	// Extract should fail with invalid zip error
	result, err := extractor.Extract(source)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "zip")
	assert.Nil(t, result)
}

// TestSourceExtractor_Extract_WithInvalidZip tests extraction with invalid zip data
// Given: Source with invalid Zip data
// When: Extract(source) is called
// Then: It should return an error
func TestSourceExtractor_Extract_WithInvalidZip(t *testing.T) {
	// Create a source with invalid zip data
	source := Source{
		Name: "test",
		Zip:  []byte("not a valid zip file"),
	}

	// Create extractor
	extractor := NewSourceExtractor()

	// Extract should fail with invalid zip error
	result, err := extractor.Extract(source)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "zip")
	assert.Nil(t, result)
}

// TestSourceExtractionResult_Close_RemovesTempDirectory tests that Close removes temp directory
// Given: SourceExtractionResult with temp directory containing extracted files
// When: Close() is called
// Then: Temp directory is removed from filesystem
func TestSourceExtractionResult_Close_RemovesTempDirectory(t *testing.T) {
	// Create a source with zip containing yeetcd.yaml
	yeetcdYaml := `name: test-project
language: JAVA
buildImage: maven:3.9.9
`
	zipData := createTestZipBytes(map[string][]byte{
		"yeetcd.yaml": []byte(yeetcdYaml),
	})

	source := Source{
		Name: "test",
		Zip:  zipData,
	}

	// Create extractor
	extractor := NewSourceExtractor()

	// Extract should succeed
	result, err := extractor.Extract(source)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Get the directory path before closing
	dir := result.Directory

	// Close should succeed
	err = result.Close()
	require.NoError(t, err)

	// Verify directory was removed
	_, err = os.Stat(dir)
	assert.True(t, os.IsNotExist(err), "temp directory should be removed")
}

// TestSourceExtractor_Extract_FromDirectory tests extraction from a directory source
// Given: Source with Directory pointing to a directory containing 'yeetcd.yaml'
// When: Extract(source) is called
// Then: SourceExtractionResult.Directory is the same directory, SourceExtractionResult.YeetcdDefinitions contains parsed config
func TestSourceExtractor_Extract_FromDirectory(t *testing.T) {
	// Create a test directory with yeetcd.yaml
	yeetcdYaml := `name: test-project
language: JAVA
buildImage: maven:3.9.9
buildCmd: mvn package
artifacts:
  - name: classes
    path: target/classes
`
	testDir, err := createTestDirectory(map[string][]byte{
		"yeetcd.yaml": []byte(yeetcdYaml),
	})
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	source := Source{
		Name:      "test",
		Directory: testDir,
	}

	// Create extractor
	extractor := NewSourceExtractor()

	// Extract should succeed
	result, err := extractor.Extract(source)
	require.NoError(t, err)
	require.NotNil(t, result)
	defer result.Close()

	// Verify directory is the same as source directory
	assert.Equal(t, testDir, result.Directory)

	// Verify yeetcd.yaml was parsed
	assert.Contains(t, result.YeetcdDefinitions, "")
	assert.Equal(t, "test-project", result.YeetcdDefinitions[""].Name)
	assert.Equal(t, config.SourceLanguageJava, result.YeetcdDefinitions[""].Language)
}

// TestSourceExtractor_Extract_FromDirectoryWithMultipleYeetcdYaml tests extraction from directory with multiple yeetcd.yaml files
// Given: Source with Directory containing 'project1/yeetcd.yaml' and 'project2/yeetcd.yaml'
// When: Extract(source) is called
// Then: SourceExtractionResult.YeetcdDefinitions contains entries for both projects
func TestSourceExtractor_Extract_FromDirectoryWithMultipleYeetcdYaml(t *testing.T) {
	// Create a test directory with multiple yeetcd.yaml files
	yeetcdYaml1 := `name: project1
language: JAVA
buildImage: maven:3.9.9
buildCmd: mvn package
artifacts:
  - name: classes
    path: target/classes
`
	yeetcdYaml2 := `name: project2
language: JAVA
buildImage: maven:3.9.9
buildCmd: mvn package
artifacts:
  - name: classes
    path: target/classes
`
	testDir, err := createTestDirectory(map[string][]byte{
		"project1/yeetcd.yaml": []byte(yeetcdYaml1),
		"project2/yeetcd.yaml": []byte(yeetcdYaml2),
	})
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	source := Source{
		Name:      "test",
		Directory: testDir,
	}

	// Create extractor
	extractor := NewSourceExtractor()

	// Extract should succeed
	result, err := extractor.Extract(source)
	require.NoError(t, err)
	require.NotNil(t, result)
	defer result.Close()

	// Verify both configs were parsed
	assert.Len(t, result.YeetcdDefinitions, 2)
	assert.Contains(t, result.YeetcdDefinitions, "project1")
	assert.Contains(t, result.YeetcdDefinitions, "project2")
	assert.Equal(t, "project1", result.YeetcdDefinitions["project1"].Name)
	assert.Equal(t, "project2", result.YeetcdDefinitions["project2"].Name)
}

// TestSourceExtractionResult_Close_DoesNotRemoveUserDirectory tests that Close doesn't remove user-provided directories
// Given: SourceExtractionResult from a directory source
// When: Close() is called
// Then: The directory should still exist
func TestSourceExtractionResult_Close_DoesNotRemoveUserDirectory(t *testing.T) {
	// Create a test directory
	yeetcdYaml := `name: test-project
language: JAVA
buildImage: maven:3.9.9
`
	testDir, err := createTestDirectory(map[string][]byte{
		"yeetcd.yaml": []byte(yeetcdYaml),
	})
	require.NoError(t, err)
	defer os.RemoveAll(testDir) // Clean up after test

	source := Source{
		Name:      "test",
		Directory: testDir,
	}

	// Create extractor
	extractor := NewSourceExtractor()

	// Extract should succeed
	result, err := extractor.Extract(source)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Close should succeed
	err = result.Close()
	require.NoError(t, err)

	// Verify directory still exists
	_, err = os.Stat(testDir)
	assert.NoError(t, err, "user-provided directory should not be removed")
}

// TestSourceExtractor_Extract_FromNonExistentDirectory tests extraction from non-existent directory
// Given: Source with Directory pointing to non-existent path
// When: Extract(source) is called
// Then: It should return an error
func TestSourceExtractor_Extract_FromNonExistentDirectory(t *testing.T) {
	source := Source{
		Name:      "test",
		Directory: "/non/existent/directory",
	}

	// Create extractor
	extractor := NewSourceExtractor()

	// Extract should fail
	result, err := extractor.Extract(source)
	require.Error(t, err)
	assert.Nil(t, result)
}
