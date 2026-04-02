package build

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
	"github.com/yeetcd/yeetcd/pkg/engine"
	"google.golang.org/protobuf/proto"
)

// MockExecutionEngine is a mock implementation of engine.ExecutionEngine for testing
type MockExecutionEngine struct {
	BuildImageFunc  func(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error)
	RemoveImageFunc func(ctx context.Context, imageID string) error
	RunJobFunc      func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error)
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

// createEmptyProtobufPipelines creates an empty Pipelines protobuf message
func createEmptyProtobufPipelines() []byte {
	pipelines := &pb.Pipelines{}
	data, _ := proto.Marshal(pipelines)
	return data
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

	// Create temp directory for output directories
	tempDir, err := os.MkdirTemp("", "yeetcd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock engine
	var buildJobCalled bool
	var pipelineGenJobCalled bool
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			// First call is the build job
			if !buildJobCalled {
				buildJobCalled = true
				assert.Equal(t, "maven:3.9.9", def.Image)
				assert.Equal(t, []string{"mvn", "package"}, def.Cmd)
				return &engine.JobResult{
					ExitCode:                0,
					OutputDirectoriesParent: tempDir,
				}, nil
			}
			// Second call is the pipeline generator
			pipelineGenJobCalled = true
			return &engine.JobResult{
				ExitCode: 0,
			}, nil
		},
		BuildImageFunc: func(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
			return &engine.BuildImageResult{ImageID: "sha256:test-image"}, nil
		},
		RemoveImageFunc: func(ctx context.Context, imageID string) error {
			return nil
		},
	}

	// Create service with mock engine
	service := NewDockerBuildService(mockEngine)

	// Build should succeed
	result, err := service.Build(ctx, source)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, buildJobCalled, "build job should have been called")
	assert.True(t, pipelineGenJobCalled, "pipeline generator job should have been called")
}

// TestDockerBuildService_Build_ReturnsBuildResultWithPipelines tests that Build returns BuildResult with pipelines
// Given: Mock ExecutionEngine that returns successful results, Source with valid config
// When: Build(ctx, source) is called
// Then: BuildResult contains pipelines from protobuf output
func TestDockerBuildService_Build_ReturnsBuildResultWithPipelines(t *testing.T) {
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

	// Create temp directory for output directories
	tempDir, err := os.MkdirTemp("", "yeetcd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a valid protobuf Pipelines message
	pipelines := &pb.Pipelines{
		Pipelines: []*pb.Pipeline{
			{
				Name: "test-pipeline",
			},
		},
	}
	protobufData, err := proto.Marshal(pipelines)
	require.NoError(t, err)

	// Create mock engine that returns successful results
	var callCount int
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			callCount++
			// First call is the build job
			if callCount == 1 {
				return &engine.JobResult{
					ExitCode:                0,
					OutputDirectoriesParent: tempDir,
				}, nil
			}
			// Second call is the pipeline generator - write protobuf to stdout
			if def.JobStreams != nil {
				def.JobStreams.StdoutWriter().Write(protobufData)
			}
			return &engine.JobResult{
				ExitCode: 0,
			}, nil
		},
		BuildImageFunc: func(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
			return &engine.BuildImageResult{ImageID: "sha256:abc123"}, nil
		},
		RemoveImageFunc: func(ctx context.Context, imageID string) error {
			return nil
		},
	}

	// Create service with mock engine
	service := NewDockerBuildService(mockEngine)

	// Build should succeed
	result, err := service.Build(ctx, source)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Pipelines)
	assert.Len(t, result.Pipelines, 1)
	assert.Equal(t, "test-pipeline", result.Pipelines[0].Name)
	assert.Len(t, result.SourceBuildResults, 1)
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

	// Build should fail with invalid zip error
	result, err := service.Build(ctx, source)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract source")
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
	zipData := createTestZip(map[string][]byte{
		"project1/yeetcd.yaml": []byte(yeetcdYaml1),
		"project2/yeetcd.yaml": []byte(yeetcdYaml2),
	})

	source := Source{
		Name: "test",
		Zip:  zipData,
	}

	// Create temp directory for output directories
	tempDir, err := os.MkdirTemp("", "yeetcd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{
				ExitCode:                0,
				OutputDirectoriesParent: tempDir,
			}, nil
		},
		BuildImageFunc: func(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
			return &engine.BuildImageResult{ImageID: "sha256:test-image"}, nil
		},
		RemoveImageFunc: func(ctx context.Context, imageID string) error {
			return nil
		},
	}

	// Create service with mock engine
	service := NewDockerBuildService(mockEngine)

	// Build should succeed
	result, err := service.Build(ctx, source)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.SourceBuildResults, 2, "should have built both projects")
}
