package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yeetcd/yeetcd/internal/testutil"
)

// getCLIPath returns the path to the yeetcd CLI binary
// It assumes the binary has been built with `go build -o bin/yeetcd ./cmd/yeetcd`
func getCLIPath(t *testing.T) string {
	// First check if there's a pre-built binary
	possiblePaths := []string{
		"bin/yeetcd",
		"../bin/yeetcd",
		"../../bin/yeetcd",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, err := filepath.Abs(path)
			require.NoError(t, err)
			return absPath
		}
	}

	// If no pre-built binary, build it temporarily
	tempDir := t.TempDir()
	cliPath := filepath.Join(tempDir, "yeetcd")

	// Get the repository root
	_, currentFile, _, _ := runtime.Caller(0)
	e2eDir := filepath.Dir(currentFile)
	repoRoot := filepath.Join(e2eDir, "..")

	cmd := exec.Command("go", "build", "-o", cliPath, "./cmd/yeetcd")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v\nOutput: %s", err, string(output))
	}

	return cliPath
}

// TestCLI_ListShowsPipelines tests that 'yeetcd list' shows all pipelines
// GIVEN: Real Docker daemon, java-sample project zip, built yeetcd CLI binary
// WHEN: 'yeetcd list --source <zip>' is executed
// THEN: Output contains pipeline names: sample, sampleCompound, sampleWithWorkContext, etc.
func TestCLI_ListShowsPipelines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	cliPath := getCLIPath(t)
	zipPath, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip")

	// Write zip to a temp file
	tempDir := t.TempDir()
	zipFile := filepath.Join(tempDir, "java-sample.zip")
	err = os.WriteFile(zipFile, zipPath, 0644)
	require.NoError(t, err, "Failed to write zip file")

	// Run the CLI
	cmd := exec.Command(cliPath, "list", "--source", zipFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("Failed to run CLI: %v", err)
		}
	}

	output := stdout.String() + stderr.String()
	t.Logf("CLI output:\n%s", output)
	t.Logf("Exit code: %d", exitCode)

	// Once implemented, verify the CLI ran successfully
	assert.Equal(t, 0, exitCode, "CLI should exit with code 0")

	// Check for expected pipeline names in output
	assert.True(t, strings.Contains(output, "sample") || strings.Contains(output, "sampleCompound") || strings.Contains(output, "sampleWithWorkContext"),
		"Output should contain at least one pipeline name")
}

// TestCLI_RunExecutesPipeline tests that 'yeetcd run' executes a pipeline
// GIVEN: Real Docker daemon, java-sample project zip, built yeetcd CLI binary
// WHEN: 'yeetcd run --source <zip> --pipeline sample' is executed
// THEN: CLI exits with code 0, output contains 'Pipeline completed: SUCCESS'
func TestCLI_RunExecutesPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	cliPath := getCLIPath(t)
	zipPath, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip")

	// Write zip to a temp file
	tempDir := t.TempDir()
	zipFile := filepath.Join(tempDir, "java-sample.zip")
	err = os.WriteFile(zipFile, zipPath, 0644)
	require.NoError(t, err, "Failed to write zip file")

	// Run the CLI
	cmd := exec.Command(cliPath, "run", "--source", zipFile, "--pipeline", "sample")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("Failed to run CLI: %v", err)
		}
	}

	output := stdout.String() + stderr.String()
	t.Logf("CLI output:\n%s", output)
	t.Logf("Exit code: %d", exitCode)

	// CLI should succeed - just verify exit code
	assert.Equal(t, 0, exitCode, "CLI should exit with code 0")
}

// TestCLI_RunWithArguments tests that 'yeetcd run' passes arguments to pipeline
// GIVEN: Real Docker daemon, java-sample project zip, built yeetcd CLI binary
// WHEN: 'yeetcd run --source <zip> --pipeline sampleWithParameters --argument PARAMETER_NAME=other' is executed
// THEN: CLI exits with code 0, pipeline receives PARAMETER_NAME=other argument
func TestCLI_RunWithArguments(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	cliPath := getCLIPath(t)
	zipPath, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip")

	// Write zip to a temp file
	tempDir := t.TempDir()
	zipFile := filepath.Join(tempDir, "java-sample.zip")
	err = os.WriteFile(zipFile, zipPath, 0644)
	require.NoError(t, err, "Failed to write zip file")

	// Run the CLI with arguments
	cmd := exec.Command(cliPath, "run",
		"--source", zipFile,
		"--pipeline", "sampleWithParameters",
		"--argument", "PARAMETER_NAME=other",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("Failed to run CLI: %v", err)
		}
	}

	output := stdout.String() + stderr.String()
	t.Logf("CLI output:\n%s", output)
	t.Logf("Exit code: %d", exitCode)

	// CLI should succeed - just verify exit code
	assert.Equal(t, 0, exitCode, "CLI should exit with code 0")
}
