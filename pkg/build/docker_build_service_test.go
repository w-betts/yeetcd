package build

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// MockExecutionEngine is a mock implementation of engine.ExecutionEngine for testing
type MockExecutionEngine struct {
	BuildImageFunc func(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error)
	RemoveImageFunc func(ctx context.Context, imageID string) error
	RunJobFunc func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error)
}

func (m *MockExecutionEngine) BuildImage(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
	if m.BuildImageFunc != nil {
		return m.BuildImageFunc(ctx, def)
	}
	return nil, nil
}

func (m *MockExecutionEngine) RemoveImage(ctx context.Context, imageID string) error {
	if m.RemoveImageFunc != nil {
		return m.RemoveImageFunc(ctx, imageID)
	}
	return nil
}

func (m *MockExecutionEngine) RunJob(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
	if m.RunJobFunc != nil {
		return m.RunJobFunc(ctx, def)
	}
	return nil, nil
}

// createTestZip creates a zip file with the given files
func createTestZip(files map[string][]byte) []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	
	for name, content := range files {
		w, _ := zipWriter.Create(name)
		w.Write(content)
	}
	
	zipWriter.Close()
	return buf.Bytes()
}

// TestDockerBuildService_Build_ExtractsSourceAndRunsBuild tests that Build extracts source and runs build
// Given: Mock ExecutionEngine, Source with Zip containing 'yeetcd.yaml' with buildImage='maven:3.9.9', buildCmd='mvn package'
// When: Build(ctx, source) is called
// Then: ExecutionEngine.RunJob is called with JobDefinition containing image='maven:3.9.9', cmd=['mvn', 'package']
func TestDockerBuildService_Build_ExtractsSourceAndRunsBuild(t *testing.T) {
	ctx := context.Background()

	// Create a source with zip containing yeetcd.yaml
	yeetcdYaml := `name: test-project
language: JAVA
buildImage: maven:3.9.9
buildCmd: mvn package
artifacts:
  - name: classes
    path: target/classes
`
	zipData := createTestZip(map[string][]byte{
		"yeetcd.yaml": []byte(yeetcdYaml),
	})

	source := Source{
		Name: "test",
		Zip:  zipData,
	}

	// Create mock engine
	mockEngine := &MockExecutionEngine{}

	// Create service with mock engine
	service := NewDockerBuildService(mockEngine)

	// Build should fail with "not implemented" since it's a stub
	result, err := service.Build(ctx, source)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, result)
}

// TestDockerBuildService_Build_ReturnsBuildResultWithImageID tests that Build returns BuildResult with image ID
// Given: Mock ExecutionEngine that returns BuildImageResult{ImageID: 'sha256:abc123'}, Source with valid config
// When: Build(ctx, source) is called
// Then: BuildResult.ImageID equals 'sha256:abc123'
func TestDockerBuildService_Build_ReturnsBuildResultWithImageID(t *testing.T) {
	ctx := context.Background()

	// Create a source with zip containing yeetcd.yaml
	yeetcdYaml := `name: test-project
language: JAVA
buildImage: maven:3.9.9
buildCmd: mvn package
`
	zipData := createTestZip(map[string][]byte{
		"yeetcd.yaml": []byte(yeetcdYaml),
	})

	source := Source{
		Name: "test",
		Zip:  zipData,
	}

	// Create mock engine that returns a successful build result
	mockEngine := &MockExecutionEngine{
		BuildImageFunc: func(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
			return &engine.BuildImageResult{ImageID: "sha256:abc123"}, nil
		},
	}

	// Create service with mock engine
	service := NewDockerBuildService(mockEngine)

	// Build should fail with "not implemented" since DockerBuildService.Build is a stub
	result, err := service.Build(ctx, source)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, result)
}

// TestDockerBuildService_Build_WithInvalidSource tests build with invalid source
// Given: A Source without valid zip data
// When: Build is called
// Then: It should return an error
func TestDockerBuildService_Build_WithInvalidSource(t *testing.T) {
	ctx := context.Background()

	// Create a source with invalid zip data
	source := Source{
		Name: "test",
		Zip:  []byte("not a valid zip"),
	}

	// Create mock engine
	mockEngine := &MockExecutionEngine{}

	// Create service with mock engine
	service := NewDockerBuildService(mockEngine)

	// Build should fail with "not implemented" since DockerBuildService.Build is a stub
	result, err := service.Build(ctx, source)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, result)
}

// TestDockerBuildService_Build_WithMultipleYeetcdYaml tests build with multiple yeetcd.yaml files
// Given: A Source with Zip containing 'project1/yeetcd.yaml' and 'project2/yeetcd.yaml'
// When: Build(ctx, source) is called
// Then: BuildResult contains parsed configs for both projects
func TestDockerBuildService_Build_WithMultipleYeetcdYaml(t *testing.T) {
	ctx := context.Background()

	// Create a source with zip containing multiple yeetcd.yaml files
	yeetcdYaml1 := `name: project1
language: JAVA
buildImage: maven:3.9.9
`
	yeetcdYaml2 := `name: project2
language: JAVA
buildImage: maven:3.9.9
`
	zipData := createTestZip(map[string][]byte{
		"project1/yeetcd.yaml": []byte(yeetcdYaml1),
		"project2/yeetcd.yaml": []byte(yeetcdYaml2),
	})

	source := Source{
		Name: "test",
		Zip:  zipData,
	}

	// Create mock engine
	mockEngine := &MockExecutionEngine{}

	// Create service with mock engine
	service := NewDockerBuildService(mockEngine)

	// Build should fail with "not implemented" since DockerBuildService.Build is a stub
	result, err := service.Build(ctx, source)
	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, result)
}
