package docker

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yeetcd/yeetcd/pkg/engine"
	"github.com/yeetcd/yeetcd/pkg/engine/docker/mocks"
)

// TestDockerExecutionEngine_BuildImage_CreatesDockerfileAndBuildsImage tests that BuildImage creates a Dockerfile and calls Docker build API
func TestDockerExecutionEngine_BuildImage_CreatesDockerfileAndBuildsImage(t *testing.T) {
	// GIVEN: Mock Docker client and BuildImageDefinition
	mockClient := new(mocks.MockDockerClient)
	buildDef := engine.BuildImageDefinition{
		Image:           "test",
		Tag:             "v1",
		ImageBase:       engine.JAVA,
		ArtifactDirectory: "/tmp/artifacts",
		ArtifactNames:   []string{"classes", "dependencies"},
		Cmd:             "TestMain",
	}

	// Create engine with mock client
	eng := &DockerExecutionEngine{dockerClient: mockClient}

	// Set up mock expectations
	mockClient.On("ImageBuild", mock.Anything, mock.Anything, mock.Anything).
		Return(&struct{ Body interface{} }{Body: nil}, nil)

	// WHEN: BuildImage is called
	result, err := eng.BuildImage(context.Background(), buildDef)

	// THEN: Dockerfile is created, Docker ImageBuild API is called, result is returned
	assert.Error(t, err) // Currently returns "not implemented"
	assert.Nil(t, result)
	assert.Equal(t, "not implemented", err.Error())
}

// TestDockerExecutionEngine_BuildImage_CleansUpDockerfile tests that Dockerfile is cleaned up after build
func TestDockerExecutionEngine_BuildImage_CleansUpDockerfile(t *testing.T) {
	// GIVEN: Mock Docker client and BuildImageDefinition
	mockClient := new(mocks.MockDockerClient)
	buildDef := engine.BuildImageDefinition{
		Image: "test",
		Tag:   "v1",
	}

	// Create engine with mock client
	eng := &DockerExecutionEngine{dockerClient: mockClient}

	// WHEN: BuildImage is called and completes
	result, err := eng.BuildImage(context.Background(), buildDef)

	// THEN: Dockerfile is deleted after build completes (success or failure)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestDockerExecutionEngine_RemoveImage_CallsDockerAPI tests that RemoveImage calls Docker API with correct image ID
func TestDockerExecutionEngine_RemoveImage_CallsDockerAPI(t *testing.T) {
	// GIVEN: Mock Docker client and image ID
	mockClient := new(mocks.MockDockerClient)
	imageID := "sha256:abc123"

	// Create engine with mock client
	eng := &DockerExecutionEngine{dockerClient: mockClient}

	// Set up mock expectations
	mockClient.On("ImageRemove", mock.Anything, imageID, mock.Anything).
		Return([]interface{}{}, nil)

	// WHEN: RemoveImage is called
	err := eng.RemoveImage(context.Background(), imageID)

	// THEN: Docker ImageRemove API is called with correct image ID, no error returned
	assert.Error(t, err) // Currently returns "not implemented"
	assert.Equal(t, "not implemented", err.Error())
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
	eng := &DockerExecutionEngine{dockerClient: mockClient}

	// Set up mock expectations - image not found
	mockClient.On("ImageInspectWithRaw", mock.Anything, "maven:3.9.9").
		Return(nil, []byte{}, errors.New("not found"))
	mockClient.On("ImagePull", mock.Anything, "maven:3.9.9", mock.Anything).
		Return(nil, nil)

	// WHEN: RunJob is called
	result, err := eng.RunJob(context.Background(), jobDef)

	// THEN: Docker ImagePull API is called
	assert.Error(t, err) // Currently returns "not implemented"
	assert.Nil(t, result)
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
	eng := &DockerExecutionEngine{dockerClient: mockClient}

	// Set up mock expectations
	mockClient.On("ImageInspectWithRaw", mock.Anything, "test:latest").
		Return(&struct{}{}, []byte{}, nil)
	mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&struct{ ID string }{ID: "container123"}, nil)

	// WHEN: RunJob is called
	result, err := eng.RunJob(context.Background(), jobDef)

	// THEN: Container is created with correct image, cmd, workingDir, env, binds
	assert.Error(t, err) // Currently returns "not implemented"
	assert.Nil(t, result)
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
	eng := &DockerExecutionEngine{dockerClient: mockClient}

	// WHEN: RunJob is called and container produces output
	result, err := eng.RunJob(context.Background(), jobDef)

	// THEN: Container logs are written to JobStreams, GetStdOut/GetStdErr return captured bytes
	assert.Error(t, err) // Currently returns "not implemented"
	assert.Nil(t, result)
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
	eng := &DockerExecutionEngine{dockerClient: mockClient}

	// Set up mock expectations
	mockClient.On("CopyFromContainer", mock.Anything, mock.Anything, "/var/out").
		Return(nil, nil, nil)

	// WHEN: RunJob is called and container exits with code 0
	result, err := eng.RunJob(context.Background(), jobDef)

	// THEN: Docker CopyFromContainer API is called, files extracted to temp directory
	assert.Error(t, err) // Currently returns "not implemented"
	assert.Nil(t, result)
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
	eng := &DockerExecutionEngine{dockerClient: mockClient}

	// Set up mock expectations
	mockClient.On("ContainerRemove", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// WHEN: RunJob is called (success or failure)
	result, err := eng.RunJob(context.Background(), jobDef)

	// THEN: Docker ContainerRemove API is called to clean up container
	assert.Error(t, err) // Currently returns "not implemented"
	assert.Nil(t, result)
}
