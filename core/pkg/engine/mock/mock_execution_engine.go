package mock

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/yeetcd/yeetcd/internal/core/proto/mock/yeetcd/protocol/mock"
	"github.com/yeetcd/yeetcd/pkg/engine"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MockExecutionEngine implements the ExecutionEngine interface using gRPC to communicate
// with a mock execution server instead of running Docker containers
type MockExecutionEngine struct {
	address string
	client  mock.MockExecutionServiceClient
	logger  interface {
		Info(msg string, args ...interface{})
		Debug(msg string, args ...interface{})
	}
}

// NewMockExecutionEngine creates a new mock execution engine that connects to the specified address
func NewMockExecutionEngine(address string) (*MockExecutionEngine, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mock execution server at %s: %w", address, err)
	}

	client := mock.NewMockExecutionServiceClient(conn)

	return &MockExecutionEngine{
		address: address,
		client:  client,
		logger:  nil, // Will be set if logger is available
	}, nil
}

// NewMockExecutionEngineWithClient creates a mock execution engine with a specific client (for testing)
func NewMockExecutionEngineWithClient(client mock.MockExecutionServiceClient) *MockExecutionEngine {
	return &MockExecutionEngine{
		address: "",
		client:  client,
	}
}

// BuildImage sends a mock image build request to the mock server
func (m *MockExecutionEngine) BuildImage(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
	req := &mock.MockImageBuildRequest{
		Image:     def.Image,
		Tag:       def.Tag,
		Artifacts: def.ArtifactNames,
		BuildCmd:  def.Cmd,
	}

	resp, err := m.client.BuildImage(ctx, req)
	if err != nil {
		return &engine.BuildImageResult{}, fmt.Errorf("mock build image failed: %w", err)
	}

	if !resp.Success {
		return &engine.BuildImageResult{}, fmt.Errorf("mock build failed: %s", resp.Error)
	}

	return &engine.BuildImageResult{
		ImageID: resp.ImageRef,
	}, nil
}

// RemoveImage is a no-op for mock execution (no actual images to remove)
func (m *MockExecutionEngine) RemoveImage(ctx context.Context, imageID string) error {
	// No-op for mock - there's nothing to remove
	return nil
}

// RunJob sends a mock work request to the mock server and returns the result
func (m *MockExecutionEngine) RunJob(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
	// Convert engine.JobDefinition to MockWorkRequest
	req := &mock.MockWorkRequest{
		Image:      def.Image,
		Cmd:        def.Cmd,
		WorkingDir: def.WorkingDir,
		EnvVars:    def.Environment,
	}

	// Handle input paths
	if len(def.InputFilePaths) > 0 {
		req.InputPaths = make([]string, 0, len(def.InputFilePaths))
		for path := range def.InputFilePaths {
			req.InputPaths = append(req.InputPaths, path)
		}
	}

	// Handle output paths
	if len(def.OutputDirectoryPaths) > 0 {
		req.OutputPaths = make([]string, 0, len(def.OutputDirectoryPaths))
		for path := range def.OutputDirectoryPaths {
			req.OutputPaths = append(req.OutputPaths, path)
		}
	}

	// Send request to mock server
	resp, err := m.client.RunWork(ctx, req)
	if err != nil {
		return &engine.JobResult{}, fmt.Errorf("mock run job failed: %w", err)
	}

	// Write stdout to JobStreams if provided
	if def.JobStreams != nil && resp.Stdout != "" {
		// Try to Base64 decode the stdout (mock server sends Base64 encoded binary)
		stdoutBytes, err := base64.StdEncoding.DecodeString(resp.Stdout)
		if err != nil {
			// If Base64 decoding fails, use the raw string
			stdoutBytes = []byte(resp.Stdout)
		}
		if _, writeErr := def.JobStreams.StdoutWriter().Write(stdoutBytes); writeErr != nil {
			// Log but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to write stdout: %v\n", writeErr)
		}
	}

	// Write stderr to JobStreams if provided
	if def.JobStreams != nil && resp.Stderr != "" {
		if _, writeErr := def.JobStreams.StderrWriter().Write([]byte(resp.Stderr)); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write stderr: %v\n", writeErr)
		}
	}

	// Convert MockWorkResponse to engine.JobResult
	result := &engine.JobResult{
		ExitCode: int(resp.ExitCode),
	}

	// Handle output paths from response
	if len(resp.OutputPaths) > 0 {
		// For mock execution, we don't have real output directories
		// The mock server can provide placeholder paths if needed
		result.OutputDirectoriesParent = "" // Mock doesn't produce real output
	}

	return result, nil
}
