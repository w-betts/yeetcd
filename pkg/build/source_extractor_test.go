package build

import (
	"archive/zip"
	"bytes"
	"os"
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
