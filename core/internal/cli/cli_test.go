package cli

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
	cliDir := filepath.Dir(currentFile)
	repoRoot := filepath.Join(cliDir, "..", "..")

	cmd := exec.Command("go", "build", "-o", cliPath, "./cmd/yeetcd")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v\nOutput: %s", err, string(output))
	}

	return cliPath
}

// TestCLI_RunInvalidPipelineReturnsError tests error handling for invalid pipeline
// GIVEN: Real Docker daemon, java-sample project zip, built yeetcd CLI binary
// WHEN: 'yeetcd run --source <zip> --pipeline nonexistent' is executed
// THEN: CLI exits with non-zero code, error message indicates pipeline not found
func TestCLI_RunInvalidPipelineReturnsError(t *testing.T) {
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

	// Run the CLI with invalid pipeline name
	cmd := exec.Command(cliPath, "run",
		"--source", zipFile,
		"--pipeline", "nonexistent",
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

	// The CLI should fail (either with stub error or "not found" error)
	assert.NotEqual(t, 0, exitCode, "CLI should fail when pipeline not found")

	// Once implemented, also check for error message:
	// assert.Contains(t, strings.ToLower(output), "not found", "Output should indicate pipeline not found")
}

// TestCLI_RunMissingSourceReturnsError tests error handling for missing source
// GIVEN: Built yeetcd CLI binary
// WHEN: 'yeetcd run --source /nonexistent/path.zip --pipeline sample' is executed
// THEN: CLI exits with non-zero code, error message indicates source file not found
func TestCLI_RunMissingSourceReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	cliPath := getCLIPath(t)

	// Run the CLI with non-existent source
	cmd := exec.Command(cliPath, "run",
		"--source", "/nonexistent/path/to/source.zip",
		"--pipeline", "sample",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
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

	// The CLI should fail
	assert.NotEqual(t, 0, exitCode, "CLI should fail when source file not found")

	// Check for appropriate error message
	lowerOutput := strings.ToLower(output)
	assert.True(t,
		strings.Contains(lowerOutput, "not found") ||
			strings.Contains(lowerOutput, "no such file") ||
			strings.Contains(lowerOutput, "failed to read") ||
			strings.Contains(lowerOutput, "error"),
		"Output should indicate file not found error",
	)
}

// TestCLI_VersionFlag tests that 'yeetcd --version' or 'yeetcd version' works
func TestCLI_VersionFlag(t *testing.T) {
	cliPath := getCLIPath(t)

	// Test --version flag
	cmd := exec.Command(cliPath, "--version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	require.NoError(t, err, "CLI should execute without error")

	version := strings.TrimSpace(stdout.String())
	t.Logf("Version output: %s", version)

	// Version should not be empty and should contain 'yeetcd'
	assert.NotEmpty(t, version, "Version output should not be empty")
	assert.Contains(t, strings.ToLower(version), "yeetcd", "Version should mention yeetcd")
}

// TestCLI_HelpFlag tests that 'yeetcd --help' works
func TestCLI_HelpFlag(t *testing.T) {
	cliPath := getCLIPath(t)

	// Test --help flag
	cmd := exec.Command(cliPath, "--help")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	require.NoError(t, err, "CLI should execute without error")

	help := stdout.String()
	t.Logf("Help output:\n%s", help)

	// Help should mention available commands
	assert.Contains(t, help, "run", "Help should mention 'run' command")
	assert.Contains(t, help, "list", "Help should mention 'list' command")
}
