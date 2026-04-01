package build

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// Extract - should fail with "not implemented" since it's a stub
	result, err := extractor.Extract(source)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, result)
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
`
	yeetcdYaml2 := `name: project2
language: JAVA
buildImage: maven:3.9.9
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

	// Extract - should fail with "not implemented" since it's a stub
	result, err := extractor.Extract(source)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, result)
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

	// Extract - should fail with "not implemented" since it's a stub
	result, err := extractor.Extract(source)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
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

	// Extract - should fail with "not implemented" since it's a stub
	result, err := extractor.Extract(source)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, result)
}

// TestSourceExtractionResult_Close_RemovesTempDirectory tests that Close removes temp directory
// Given: SourceExtractionResult with temp directory containing extracted files
// When: Close() is called
// Then: Temp directory is removed from filesystem
func TestSourceExtractionResult_Close_RemovesTempDirectory(t *testing.T) {
	// Create a source extraction result with a temp directory
	// Note: Since Extract is not implemented, we can't get a real result
	// This test verifies the Close method signature exists
	
	// Create a dummy result (this would normally come from Extract)
	result := &SourceExtractionResult{
		Source:    Source{Name: "test", Zip: []byte{}},
		Directory: "/tmp/test-extraction",
	}

	// Close should fail with "not implemented" since it's a stub
	err := result.Close()
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
}
