package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test: YeetcdConfig.Load() successfully loads valid configuration
// Given: A valid yeetcd.yaml file with all required fields
// When: Calling Load() with the file path
// Then: Returns YeetcdConfig with parsed values matching the YAML content
func TestYeetcdConfig_Load_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "yeetcd.yaml")
	
	configContent := `
name: test-pipeline
language: JAVA
buildImage: maven:3.9-eclipse-temurin-17
buildCmd: mvn clean package
artifacts:
  - name: classes
    path: target/classes
  - name: dependencies
    path: target/dependency
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := Load(configPath)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test-pipeline", config.Name)
	assert.Equal(t, SourceLanguageJava, config.Language)
	assert.Equal(t, "maven:3.9-eclipse-temurin-17", config.BuildImage)
	assert.Equal(t, "mvn clean package", config.BuildCmd)
	assert.Len(t, config.Artifacts, 2)
}

// Test: YeetcdConfig.Load() returns error when file does not exist
// Given: A path to a non-existent file
// When: Calling Load() with the invalid path
// Then: Returns error indicating file not found
func TestYeetcdConfig_Load_FileNotFound(t *testing.T) {
	nonExistentPath := "/non/existent/path/yeetcd.yaml"

	_, err := Load(nonExistentPath)

	assert.Error(t, err)
}

// Test: YeetcdConfig.Load() returns error for invalid YAML syntax
// Given: A yeetcd.yaml file with malformed YAML syntax
// When: Calling Load() with the file path
// Then: Returns error indicating YAML parsing failure
func TestYeetcdConfig_Load_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "yeetcd.yaml")
	
	invalidContent := `
name: test-pipeline
language: JAVA
buildImage: maven:3.9
  - invalid: yaml: syntax: here
`
	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	_, err = Load(configPath)

	assert.Error(t, err)
}

// Test: YeetcdConfig.Load() returns error for missing required fields
// Given: A yeetcd.yaml file missing required fields (name, language, buildImage)
// When: Calling Load() with the file path
// Then: Returns error indicating missing required field
func TestYeetcdConfig_Load_MissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "yeetcd.yaml")
	
	incompleteContent := `
name: test-pipeline
# missing language and buildImage
`
	err := os.WriteFile(configPath, []byte(incompleteContent), 0644)
	require.NoError(t, err)

	_, err = Load(configPath)

	assert.Error(t, err)
}

// Test: YeetcdConfig.Load() parses artifacts section correctly
// Given: A yeetcd.yaml file with artifacts array containing name and path
// When: Calling Load() with the file path
// Then: Returns YeetcdConfig with Artifacts slice populated correctly
func TestYeetcdConfig_Load_Artifacts(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "yeetcd.yaml")
	
	configContent := `
name: test-pipeline
language: JAVA
buildImage: maven:3.9
buildCmd: mvn clean package
artifacts:
  - name: output1
    path: build/output1
  - name: output2
    path: build/output2
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := Load(configPath)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Len(t, config.Artifacts, 2)
	assert.Equal(t, "output1", config.Artifacts[0].Name)
	assert.Equal(t, "build/output1", config.Artifacts[0].Path)
	assert.Equal(t, "output2", config.Artifacts[1].Name)
	assert.Equal(t, "build/output2", config.Artifacts[1].Path)
}

// Test: YeetcdConfig.Load() handles empty artifacts list
// Given: A yeetcd.yaml file with empty artifacts array
// When: Calling Load() with the file path
// Then: Returns YeetcdConfig with empty Artifacts slice
func TestYeetcdConfig_Load_EmptyArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "yeetcd.yaml")
	
	configContent := `
name: test-pipeline
language: JAVA
buildImage: maven:3.9
buildCmd: mvn clean package
artifacts: []
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := Load(configPath)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Empty(t, config.Artifacts)
}

// Test: YeetcdConfig.Load() handles optional buildCmd field
// Given: A yeetcd.yaml file without buildCmd field
// When: Calling Load() with the file path
// Then: Returns YeetcdConfig with empty buildCmd
func TestYeetcdConfig_Load_OptionalBuildCmd(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "yeetcd.yaml")
	
	configContent := `
name: test-pipeline
language: JAVA
buildImage: maven:3.9
artifacts: []
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := Load(configPath)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Empty(t, config.BuildCmd)
}

// Test: YeetcdConfig.Load() validates language enum values
// Given: A yeetcd.yaml file with invalid language value
// When: Calling Load() with the file path
// Then: Returns error indicating invalid language
func TestYeetcdConfig_Load_InvalidLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "yeetcd.yaml")
	
	configContent := `
name: test-pipeline
language: INVALID_LANGUAGE
buildImage: maven:3.9
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	_, err = Load(configPath)

	assert.Error(t, err)
}
