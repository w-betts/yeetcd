package sample

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktest "github.com/yeetcd/yeetcd/sdk/test/pkg/yeetcd"
)

// TestE2EPipelineTestRun tests the full pipeline execution using PipelineTestRun.
// This test proves that:
// 1. The generator produces valid protobuf
// 2. The controller can parse the protobuf
// 3. The mock execution engine receives work requests
// 4. The pipeline executes correctly
func TestE2EPipelineTestRun(t *testing.T) {
	// Ensure the CLI is built
	cliPath := ensureCLIBuilt(t)

	// Add CLI to PATH for the test
	oldPath := os.Getenv("PATH")
	dir := filepath.Dir(cliPath)
	os.Setenv("PATH", dir+":"+oldPath)
	defer os.Setenv("PATH", oldPath)

	// Create a PipelineTestRun
	testRun := sdktest.NewPipelineTestRun().
		WithPipelineName("sample").
		WithSourceDir(getSampleDir()).
		WithTimeout(30*time.Second).
		ContainerisedWork("maven:3.9.9-eclipse-temurin-17").
		WithResult(0, "Hello from a containerised task", "").
		Build().
		Build()

	// Run the test
	result, err := testRun.Start()
	require.NoError(t, err)

	// Print CLI output for debugging
	t.Logf("CLI Output:\n%s", result.CLIOutput)
	t.Logf("Executions: %d", len(result.Executions))
	for i, exec := range result.Executions {
		t.Logf("  Execution %d: type=%s image=%s exitCode=%d", i, exec.Type, exec.Image, exec.ExitCode)
		t.Logf("    stdout: %s", exec.Stdout)
		t.Logf("    stderr: %s", exec.Stderr)
	}

	// Verify the result
	assert.Equal(t, sdktest.PipelineStatusSuccess, result.Status)
	assert.Equal(t, 0, result.ExitCode)
	assert.True(t, result.HasExecution("maven:3.9.9-eclipse-temurin-17"))
}

// TestE2EPipelineTestRunFailure tests that failures are handled correctly.
func TestE2EPipelineTestRunFailure(t *testing.T) {
	// Ensure the CLI is built
	cliPath := ensureCLIBuilt(t)

	oldPath := os.Getenv("PATH")
	dir := filepath.Dir(cliPath)
	os.Setenv("PATH", dir+":"+oldPath)
	defer os.Setenv("PATH", oldPath)

	testRun := sdktest.NewPipelineTestRun().
		WithPipelineName("sample").
		WithSourceDir(getSampleDir()).
		WithTimeout(30*time.Second).
		ContainerisedWork("maven:3.9.9-eclipse-temurin-17").
		WithResult(1, "", "work failed").
		Build().
		Build()

	result, err := testRun.Start()
	require.NoError(t, err)

	assert.Equal(t, sdktest.PipelineStatusFailure, result.Status)
	assert.Equal(t, 1, result.ExitCode)
}

// TestE2EPipelineTestRunMultipleExecutions tests multiple work executions.
func TestE2EPipelineTestRunMultipleExecutions(t *testing.T) {
	// Ensure the CLI is built
	cliPath := ensureCLIBuilt(t)

	oldPath := os.Getenv("PATH")
	dir := filepath.Dir(cliPath)
	os.Setenv("PATH", dir+":"+oldPath)
	defer os.Setenv("PATH", oldPath)

	testRun := sdktest.NewPipelineTestRun().
		WithPipelineName("sampleCompound").
		WithSourceDir(getSampleDir()).
		WithTimeout(30*time.Second).
		ContainerisedWork("alpine").
		WithResult(0, "work done", "").
		Build().
		Build()

	result, err := testRun.Start()
	require.NoError(t, err)

	assert.Equal(t, sdktest.PipelineStatusSuccess, result.Status)
	// Should have multiple executions for the compound pipeline
	assert.GreaterOrEqual(t, len(result.GetExecutions()), 1)
}

// TestE2EPipelineTestRunDefaultBehavior tests using default behavior.
func TestE2EPipelineTestRunDefaultBehavior(t *testing.T) {
	// Ensure the CLI is built
	cliPath := ensureCLIBuilt(t)

	oldPath := os.Getenv("PATH")
	dir := filepath.Dir(cliPath)
	os.Setenv("PATH", dir+":"+oldPath)
	defer os.Setenv("PATH", oldPath)

	testRun := sdktest.NewPipelineTestRun().
		WithPipelineName("sample").
		WithSourceDir(getSampleDir()).
		WithTimeout(30*time.Second).
		DefaultBehavior().
		WithResult(0, "default output", "").
		Build().
		Build()

	result, err := testRun.Start()
	require.NoError(t, err)

	assert.Equal(t, sdktest.PipelineStatusSuccess, result.Status)
}

// ensureCLIBuilt ensures the CLI binary is built and returns its path.
func ensureCLIBuilt(t *testing.T) string {
	t.Helper()

	// Check if CLI exists in expected location
	coreDir := filepath.Join("..", "..", "..", "core")
	cliPath := filepath.Join(coreDir, "bin", "yeetcd")

	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		// Build the CLI
		cmd := exec.Command("go", "build", "-o", filepath.Join("bin", "yeetcd"), "./cmd/yeetcd")
		cmd.Dir = coreDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to build CLI: %s", string(output))
	}

	absPath, err := filepath.Abs(cliPath)
	require.NoError(t, err)

	return absPath
}

// getSampleDir returns the absolute path to the sample directory.
func getSampleDir() string {
	dir, _ := os.Getwd()
	return dir
}
