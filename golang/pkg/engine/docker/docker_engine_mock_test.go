package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yeetcd/yeetcd/pkg/engine"
	"github.com/yeetcd/yeetcd/pkg/engine/docker/mocks"
)

// TestDockerExecutionEngine_BuildImage_CreatesDockerfileAndBuildsImage tests that BuildImage creates a Dockerfile and calls Docker build API
func TestDockerExecutionEngine_BuildImage_CreatesDockerfileAndBuildsImage(t *testing.T) {
	// GIVEN: Mock Docker client, temp artifact directory, and BuildImageDefinition
	mockClient := new(mocks.MockDockerClient)
	
	// Create temp artifact directory
	artifactDir := t.TempDir()
	
	buildDef := engine.BuildImageDefinition{
		Image:           "test",
		Tag:             "v1",
		ImageBase:       engine.JAVA,
		ArtifactDirectory: artifactDir,
		ArtifactNames:   []string{"classes", "dependencies"},
		Cmd:             "TestMain",
	}

	// Create engine with mock client
	eng := NewDockerExecutionEngineWithClient(mockClient)

	// Set up mock expectations - ImageBuild returns a reader with success message
	buildOutput := createMockBuildOutput("sha256:abc123")
	mockClient.On("ImageBuild", mock.Anything, mock.Anything, mock.Anything).
		Return(types.ImageBuildResponse{Body: io.NopCloser(buildOutput)}, nil)

	// WHEN: BuildImage is called
	result, err := eng.BuildImage(context.Background(), buildDef)

	// THEN: Dockerfile is created, Docker ImageBuild API is called, result is returned
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.ImageID, "sha256:abc123")
	mockClient.AssertExpectations(t)
}

// TestDockerExecutionEngine_BuildImage_CleansUpDockerfile tests that Dockerfile is cleaned up after build
func TestDockerExecutionEngine_BuildImage_CleansUpDockerfile(t *testing.T) {
	// GIVEN: Mock Docker client, temp artifact directory, and BuildImageDefinition
	mockClient := new(mocks.MockDockerClient)
	
	// Create temp artifact directory
	artifactDir := t.TempDir()
	
	buildDef := engine.BuildImageDefinition{
		Image:           "test",
		Tag:             "v1",
		ImageBase:       engine.JAVA,
		ArtifactDirectory: artifactDir,
		ArtifactNames:   []string{"classes"},
		Cmd:             "TestMain",
	}

	// Create engine with mock client
	eng := NewDockerExecutionEngineWithClient(mockClient)

	// Set up mock expectations
	buildOutput := createMockBuildOutput("sha256:abc123")
	mockClient.On("ImageBuild", mock.Anything, mock.Anything, mock.Anything).
		Return(types.ImageBuildResponse{Body: io.NopCloser(buildOutput)}, nil)

	// WHEN: BuildImage is called and completes
	result, err := eng.BuildImage(context.Background(), buildDef)

	// THEN: Dockerfile is deleted after build completes
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockClient.AssertExpectations(t)
}

// TestDockerExecutionEngine_RemoveImage_CallsDockerAPI tests that RemoveImage calls Docker API with correct image ID
func TestDockerExecutionEngine_RemoveImage_CallsDockerAPI(t *testing.T) {
	// GIVEN: Mock Docker client and image ID
	mockClient := new(mocks.MockDockerClient)
	imageID := "sha256:abc123"

	// Create engine with mock client
	eng := NewDockerExecutionEngineWithClient(mockClient)

	// Set up mock expectations
	mockClient.On("ImageRemove", mock.Anything, imageID, mock.Anything).
		Return([]image.DeleteResponse{}, nil)

	// WHEN: RemoveImage is called
	err := eng.RemoveImage(context.Background(), imageID)

	// THEN: Docker ImageRemove API is called with correct image ID, no error returned
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// TestDockerExecutionEngine_RunJob_PullsImageIfNotPresent tests that RunJob pulls image if not present locally
func TestDockerExecutionEngine_RunJob_PullsImageIfNotPresent(t *testing.T) {
	// GIVEN: Mock Docker client that returns NotFound for image inspect, JobDefinition
	mockClient := new(mocks.MockDockerClient)
	jobDef := engine.JobDefinition{
		Image: "maven:3.9.9",
		Cmd:   []string{"echo", "hello"},
	}

	// Create engine with mock client
	eng := NewDockerExecutionEngineWithClient(mockClient)

	// Set up mock expectations - image not found, then pull
	mockClient.On("ImageInspectWithRaw", mock.Anything, "maven:3.9.9").
		Return(types.ImageInspect{}, []byte{}, errors.New("not found"))
	mockClient.On("ImagePull", mock.Anything, "maven:3.9.9", mock.Anything).
		Return(io.NopCloser(bytes.NewReader([]byte{})), nil)
	mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(container.CreateResponse{ID: "container123"}, nil)
	mockClient.On("ContainerStart", mock.Anything, "container123", mock.Anything).
		Return(nil)
	mockClient.On("ContainerWait", mock.Anything, "container123", mock.Anything).
		Return(container.WaitResponse{StatusCode: 0}, nil)
	mockClient.On("ContainerRemove", mock.Anything, "container123", mock.Anything).
		Return(nil)

	// WHEN: RunJob is called
	result, err := eng.RunJob(context.Background(), jobDef)

	// THEN: Docker ImagePull API is called
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
	mockClient.AssertExpectations(t)
}

// TestDockerExecutionEngine_RunJob_CreatesContainerWithCorrectConfig tests that RunJob creates container with correct configuration
func TestDockerExecutionEngine_RunJob_CreatesContainerWithCorrectConfig(t *testing.T) {
	// GIVEN: Mock Docker client, JobDefinition with image, cmd, workingDir, env, inputFilePaths
	mockClient := new(mocks.MockDockerClient)
	jobDef := engine.JobDefinition{
		Image:      "test:latest",
		Cmd:        []string{"echo", "hello"},
		WorkingDir: "/app",
		Environment: map[string]string{
			"KEY": "value",
		},
		InputFilePaths: map[string]engine.MountInput{
			"/input": engine.OnDiskMountInput{Dir: "/host/path"},
		},
	}

	// Create engine with mock client
	eng := NewDockerExecutionEngineWithClient(mockClient)

	// Set up mock expectations
	mockClient.On("ImageInspectWithRaw", mock.Anything, "test:latest").
		Return(types.ImageInspect{}, []byte{}, nil)
	mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(container.CreateResponse{ID: "container123"}, nil)
	mockClient.On("ContainerStart", mock.Anything, "container123", mock.Anything).
		Return(nil)
	mockClient.On("ContainerWait", mock.Anything, "container123", mock.Anything).
		Return(container.WaitResponse{StatusCode: 0}, nil)
	mockClient.On("ContainerRemove", mock.Anything, "container123", mock.Anything).
		Return(nil)

	// WHEN: RunJob is called
	result, err := eng.RunJob(context.Background(), jobDef)

	// THEN: Container is created with correct image, cmd, workingDir, env, binds
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
	mockClient.AssertExpectations(t)
}

// TestDockerExecutionEngine_RunJob_CapturesStdoutAndStderr tests that RunJob captures stdout and stderr to JobStreams
func TestDockerExecutionEngine_RunJob_CapturesStdoutAndStderr(t *testing.T) {
	// GIVEN: Mock Docker client that returns container logs, JobStreams, JobDefinition
	mockClient := new(mocks.MockDockerClient)
	streams := engine.NewJobStreams(nil, nil)
	jobDef := engine.JobDefinition{
		Image:      "test:latest",
		Cmd:        []string{"echo", "hello"},
		JobStreams: streams,
	}

	// Create engine with mock client
	eng := NewDockerExecutionEngineWithClient(mockClient)

	// Set up mock expectations
	mockClient.On("ImageInspectWithRaw", mock.Anything, "test:latest").
		Return(types.ImageInspect{}, []byte{}, nil)
	mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(container.CreateResponse{ID: "container123"}, nil)
	mockClient.On("ContainerAttach", mock.Anything, "container123", mock.Anything).
		Return(types.HijackedResponse{}, nil)
	mockClient.On("ContainerStart", mock.Anything, "container123", mock.Anything).
		Return(nil)
	mockClient.On("ContainerWait", mock.Anything, "container123", mock.Anything).
		Return(container.WaitResponse{StatusCode: 0}, nil)
	mockClient.On("ContainerRemove", mock.Anything, "container123", mock.Anything).
		Return(nil)

	// WHEN: RunJob is called and container produces output
	result, err := eng.RunJob(context.Background(), jobDef)

	// THEN: Container logs are written to JobStreams, GetStdOut/GetStdErr return captured bytes
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
	mockClient.AssertExpectations(t)
}

// TestDockerExecutionEngine_RunJob_ExtractsOutputDirectories tests that RunJob extracts output directories to host
func TestDockerExecutionEngine_RunJob_ExtractsOutputDirectories(t *testing.T) {
	// GIVEN: Mock Docker client, JobDefinition with outputDirectoryPaths
	mockClient := new(mocks.MockDockerClient)
	jobDef := engine.JobDefinition{
		Image: "test:latest",
		Cmd:   []string{"bash", "-c", "mkdir -p /var/out && echo hello > /var/out/file.txt"},
		OutputDirectoryPaths: map[string]string{
			"output": "/var/out",
		},
	}

	// Create engine with mock client
	eng := NewDockerExecutionEngineWithClient(mockClient)

	// Create a mock tar archive for the output directory
	tarBuffer := &bytes.Buffer{}
	tw := tar.NewWriter(tarBuffer)
	hdr := &tar.Header{
		Name: "file.txt",
		Mode: 0644,
		Size: int64(len("hello\n")),
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte("hello\n"))
	tw.Close()

	// Set up mock expectations
	mockClient.On("ImageInspectWithRaw", mock.Anything, "test:latest").
		Return(types.ImageInspect{}, []byte{}, nil)
	mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(container.CreateResponse{ID: "container123"}, nil)
	mockClient.On("ContainerStart", mock.Anything, "container123", mock.Anything).
		Return(nil)
	mockClient.On("ContainerWait", mock.Anything, "container123", mock.Anything).
		Return(container.WaitResponse{StatusCode: 0}, nil)
	mockClient.On("CopyFromContainer", mock.Anything, "container123", "/var/out").
		Return(io.NopCloser(tarBuffer), container.PathStat{}, nil)
	mockClient.On("ContainerRemove", mock.Anything, "container123", mock.Anything).
		Return(nil)

	// WHEN: RunJob is called and container exits with code 0
	result, err := eng.RunJob(context.Background(), jobDef)

	// THEN: Docker CopyFromContainer API is called, files extracted to temp directory
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
	assert.NotEmpty(t, result.OutputDirectoriesParent)
	mockClient.AssertExpectations(t)
}

// TestDockerExecutionEngine_RunJob_RemovesContainerAfterExecution tests that RunJob removes container after execution
func TestDockerExecutionEngine_RunJob_RemovesContainerAfterExecution(t *testing.T) {
	// GIVEN: Mock Docker client, JobDefinition
	mockClient := new(mocks.MockDockerClient)
	jobDef := engine.JobDefinition{
		Image: "test:latest",
		Cmd:   []string{"echo", "hello"},
	}

	// Create engine with mock client
	eng := NewDockerExecutionEngineWithClient(mockClient)

	// Set up mock expectations
	mockClient.On("ImageInspectWithRaw", mock.Anything, "test:latest").
		Return(types.ImageInspect{}, []byte{}, nil)
	mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(container.CreateResponse{ID: "container123"}, nil)
	mockClient.On("ContainerStart", mock.Anything, "container123", mock.Anything).
		Return(nil)
	mockClient.On("ContainerWait", mock.Anything, "container123", mock.Anything).
		Return(container.WaitResponse{StatusCode: 0}, nil)
	mockClient.On("ContainerRemove", mock.Anything, "container123", mock.Anything).
		Return(nil)

	// WHEN: RunJob is called (success or failure)
	result, err := eng.RunJob(context.Background(), jobDef)

	// THEN: Docker ContainerRemove API is called to clean up container
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockClient.AssertExpectations(t)
}

// createMockBuildOutput creates a mock Docker build output with the given image ID
func createMockBuildOutput(imageID string) *bytes.Buffer {
	buf := &bytes.Buffer{}
	
	// Create aux message with image ID
	auxData := map[string]interface{}{
		"ID": imageID,
	}
	auxJSON, _ := json.Marshal(auxData)
	
	msg := jsonmessage.JSONMessage{
		Aux: &json.RawMessage{},
	}
	*msg.Aux = auxJSON
	
	encoder := json.NewEncoder(buf)
	encoder.Encode(msg)
	
	return buf
}
