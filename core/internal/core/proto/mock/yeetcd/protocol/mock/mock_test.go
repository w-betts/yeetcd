package mock_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yeetcd/yeetcd/internal/core/proto/mock/yeetcd/protocol/mock"
)

func TestMockWorkRequestSerialization(t *testing.T) {
	// Given: a MockWorkRequest with all fields populated
	req := &mock.MockWorkRequest{
		Image:       "nginx:latest",
		Cmd:         []string{"echo", "hello"},
		EnvVars:     map[string]string{"FOO": "bar"},
		WorkingDir:  "/app",
		InputPaths:  []string{"/inputs/data"},
		OutputPaths: []string{"/outputs/result"},
	}

	// Then: all fields are correctly accessible
	assert.Equal(t, "nginx:latest", req.GetImage())
	assert.Equal(t, []string{"echo", "hello"}, req.GetCmd())
	assert.Equal(t, map[string]string{"FOO": "bar"}, req.GetEnvVars())
	assert.Equal(t, "/app", req.GetWorkingDir())
	assert.Equal(t, []string{"/inputs/data"}, req.GetInputPaths())
	assert.Equal(t, []string{"/outputs/result"}, req.GetOutputPaths())
}

func TestMockWorkRequestWithEmptyFields(t *testing.T) {
	// Given: a minimal MockWorkRequest
	req := &mock.MockWorkRequest{
		Image: "alpine",
		Cmd:   []string{"sh", "-c", "echo test"},
	}

	// Then: default values are as expected
	assert.Equal(t, "alpine", req.GetImage())
	assert.Equal(t, []string{"sh", "-c", "echo test"}, req.GetCmd())
	assert.Nil(t, req.GetEnvVars())
	assert.Equal(t, "", req.GetWorkingDir())
	assert.Nil(t, req.GetInputPaths())
	assert.Nil(t, req.GetOutputPaths())
}

func TestMockWorkResponseSerialization(t *testing.T) {
	// Given: a MockWorkResponse with exit code and outputs
	resp := &mock.MockWorkResponse{
		ExitCode:    0,
		Stdout:      "Hello, World!",
		Stderr:      "",
		OutputPaths: map[string]string{"result": "/outputs/result.txt"},
	}

	// Then: response correctly maps to protobuf fields
	assert.Equal(t, int32(0), resp.GetExitCode())
	assert.Equal(t, "Hello, World!", resp.GetStdout())
	assert.Equal(t, "", resp.GetStderr())
	assert.Equal(t, map[string]string{"result": "/outputs/result.txt"}, resp.GetOutputPaths())
}

func TestMockWorkResponseError(t *testing.T) {
	// Given: a MockWorkResponse with error
	resp := &mock.MockWorkResponse{
		ExitCode:    42,
		Stdout:      "error output",
		Stderr:      "error message",
		OutputPaths: map[string]string{"err": "/outputs/error.txt"},
	}

	// Then: error fields are correctly set
	assert.Equal(t, int32(42), resp.GetExitCode())
	assert.Equal(t, "error output", resp.GetStdout())
	assert.Equal(t, "error message", resp.GetStderr())
	assert.Equal(t, map[string]string{"err": "/outputs/error.txt"}, resp.GetOutputPaths())
}

func TestMockImageBuildRequest(t *testing.T) {
	// Given: a MockImageBuildRequest with all build parameters
	req := &mock.MockImageBuildRequest{
		Image:     "my-app",
		Tag:       "latest",
		Artifacts: []string{"target/classes", "target/dependency"},
		BuildCmd:  "mvn package",
	}

	// Then: all parameters are correctly mapped
	assert.Equal(t, "my-app", req.GetImage())
	assert.Equal(t, "latest", req.GetTag())
	assert.Equal(t, []string{"target/classes", "target/dependency"}, req.GetArtifacts())
	assert.Equal(t, "mvn package", req.GetBuildCmd())
}

func TestMockImageBuildResponse(t *testing.T) {
	// Given: successful build response
	resp := &mock.MockImageBuildResponse{
		Success:  true,
		Error:    "",
		ImageRef: "my-app:latest",
	}

	// Then: fields are correctly set
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, "", resp.GetError())
	assert.Equal(t, "my-app:latest", resp.GetImageRef())

	// Given: failed build response
	respFail := &mock.MockImageBuildResponse{
		Success:  false,
		Error:    "Build failed",
		ImageRef: "",
	}

	// Then: error is captured
	assert.False(t, respFail.GetSuccess())
	assert.Equal(t, "Build failed", respFail.GetError())
}

func TestMockExecutionServiceClientExists(t *testing.T) {
	// Verify the MockExecutionServiceClient interface exists
	var _ mock.MockExecutionServiceClient = nil
}

func TestMockExecutionServiceServerExists(t *testing.T) {
	// Verify the MockExecutionServiceServer interface exists
	var _ mock.MockExecutionServiceServer = nil
}
