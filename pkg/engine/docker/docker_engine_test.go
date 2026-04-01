package docker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// TestDockerExecutionEngine_BuildImage_Integration tests building a Docker image
// Given: A valid BuildImageDefinition with artifacts
// When: BuildImage is called
// Then: It should successfully build the image and return a valid image ID
func TestDockerExecutionEngine_BuildImage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create engine
	eng, err := NewDockerExecutionEngine()
	require.NoError(t, err)

	// Create temp artifact directory with test files
	artifactDir := t.TempDir()
	classesDir := filepath.Join(artifactDir, "classes")
	depsDir := filepath.Join(artifactDir, "dependencies")
	require.NoError(t, os.MkdirAll(classesDir, 0755))
	require.NoError(t, os.MkdirAll(depsDir, 0755))
	// Write a dummy class file
	require.NoError(t, os.WriteFile(filepath.Join(classesDir, "Test.class"), []byte("test"), 0644))

	// Create build definition
	buildDef := engine.BuildImageDefinition{
		Image:             "yeetcd-test-build",
		Tag:               "v1",
		ImageBase:         engine.JAVA,
		ArtifactDirectory: artifactDir,
		ArtifactNames:     []string{"classes", "dependencies"},
		Cmd:               "TestMain",
	}

	// Build the image
	result, err := eng.BuildImage(ctx, buildDef)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ImageID)

	// Cleanup: remove the image
	defer func() {
		_ = eng.RemoveImage(ctx, result.ImageID)
	}()
}

// TestDockerExecutionEngine_RemoveImage_Integration tests removing a Docker image
// Given: A valid image ID
// When: RemoveImage is called
// Then: It should successfully remove the image
func TestDockerExecutionEngine_RemoveImage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create engine
	eng, err := NewDockerExecutionEngine()
	require.NoError(t, err)

	// First build an image to remove
	artifactDir := t.TempDir()
	classesDir := filepath.Join(artifactDir, "classes")
	require.NoError(t, os.MkdirAll(classesDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(classesDir, "Test.class"), []byte("test"), 0644))

	buildDef := engine.BuildImageDefinition{
		Image:             "yeetcd-test-remove",
		Tag:               "v1",
		ImageBase:         engine.JAVA,
		ArtifactDirectory: artifactDir,
		ArtifactNames:     []string{"classes"},
		Cmd:               "TestMain",
	}

	result, err := eng.BuildImage(ctx, buildDef)
	require.NoError(t, err)

	// Remove the image
	err = eng.RemoveImage(ctx, result.ImageID)
	assert.NoError(t, err)
}

// TestDockerExecutionEngine_RunJob_Integration tests running a container job
// Given: A valid JobDefinition with a simple container command
// When: RunJob is called
// Then: It should successfully run the job and return the exit code
func TestDockerExecutionEngine_RunJob_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create engine
	eng, err := NewDockerExecutionEngine()
	require.NoError(t, err)

	// Create job definition
	jobDef := engine.JobDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "hello"},
	}

	// Run the job
	result, err := eng.RunJob(ctx, jobDef)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
}

// TestDockerExecutionEngine_RunJob_WithEnvVars_Integration tests running with environment variables
// Given: A JobDefinition with environment variables
// When: RunJob is called
// Then: It should pass the environment variables to the container
func TestDockerExecutionEngine_RunJob_WithEnvVars_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create engine
	eng, err := NewDockerExecutionEngine()
	require.NoError(t, err)

	// Create job definition with environment variables
	jobDef := engine.JobDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"sh", "-c", "test $MY_VAR = test_value"},
		Environment: map[string]string{
			"MY_VAR": "test_value",
		},
	}

	// Run the job
	result, err := eng.RunJob(ctx, jobDef)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
}

// TestDockerExecutionEngine_RunJob_WithMounts_Integration tests running with volume mounts
// Given: A JobDefinition with volume mounts
// When: RunJob is called
// Then: It should mount the volumes and allow file access
func TestDockerExecutionEngine_RunJob_WithMounts_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create engine
	eng, err := NewDockerExecutionEngine()
	require.NoError(t, err)

	// Create temp directory with test file
	hostDir := t.TempDir()
	testFile := filepath.Join(hostDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello from host"), 0644))

	// Create job definition with mount
	jobDef := engine.JobDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"cat", "/mnt/test.txt"},
		InputFilePaths: map[string]engine.MountInput{
			"/mnt": engine.OnDiskMountInput{Dir: hostDir},
		},
	}

	// Run the job
	result, err := eng.RunJob(ctx, jobDef)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
}

// TestDockerExecutionEngine_RunJob_WithOutputDirs_Integration tests running with output directories
// Given: A JobDefinition with output directory paths
// When: RunJob is called
// Then: It should extract output directories after execution
func TestDockerExecutionEngine_RunJob_WithOutputDirs_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create engine
	eng, err := NewDockerExecutionEngine()
	require.NoError(t, err)

	// Create job definition with output directories
	jobDef := engine.JobDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"sh", "-c", "mkdir -p /output && echo hello > /output/result.txt"},
		OutputDirectoryPaths: map[string]string{
			"output": "/output",
		},
	}

	// Run the job
	result, err := eng.RunJob(ctx, jobDef)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
	assert.NotEmpty(t, result.OutputDirectoriesParent)

	// Verify the output file was extracted
	outputFile := filepath.Join(result.OutputDirectoriesParent, "output", "result.txt")
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "hello")
}

// TestDockerExecutionEngine_RunJob_WithStreams_Integration tests running with job streams
// Given: A JobDefinition with JobStreams
// When: RunJob is called
// Then: It should capture stdout and stderr
func TestDockerExecutionEngine_RunJob_WithStreams_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create engine
	eng, err := NewDockerExecutionEngine()
	require.NoError(t, err)

	// Create job streams
	streams := engine.NewJobStreams(nil, nil)

	// Create job definition
	jobDef := engine.JobDefinition{
		Image:      "alpine:latest",
		Cmd:        []string{"echo", "hello"},
		JobStreams: streams,
	}

	// Run the job
	result, err := eng.RunJob(ctx, jobDef)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)

	// Verify stdout was captured
	stdout := streams.GetStdOut()
	assert.Contains(t, string(stdout), "hello")
}
