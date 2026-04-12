package test

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	pb "github.com/yeetcd/yeetcd/pkg/proto/mock"
	"google.golang.org/grpc"
)

// PipelineTestRun represents a test run of a pipeline.
type PipelineTestRun struct {
	pipelineName string
	sourceDir    string
	behaviors    *behaviorChain
	timeout      time.Duration
}

// PipelineTestRunBuilder builds a PipelineTestRun.
type PipelineTestRunBuilder struct {
	pipelineName string
	sourceDir    string
	behaviors    *behaviorChain
	timeout      time.Duration
}

// NewPipelineTestRun creates a new PipelineTestRunBuilder.
func NewPipelineTestRun() *PipelineTestRunBuilder {
	return &PipelineTestRunBuilder{
		timeout:   60 * time.Second,
		behaviors: newBehaviorChain(),
	}
}

// WithPipelineName sets the pipeline name.
func (b *PipelineTestRunBuilder) WithPipelineName(name string) *PipelineTestRunBuilder {
	b.pipelineName = name
	return b
}

// WithSourceDir sets the source directory.
func (b *PipelineTestRunBuilder) WithSourceDir(dir string) *PipelineTestRunBuilder {
	b.sourceDir = dir
	return b
}

// WithTimeout sets the timeout.
func (b *PipelineTestRunBuilder) WithTimeout(timeout time.Duration) *PipelineTestRunBuilder {
	b.timeout = timeout
	return b
}

// ContainerisedWork starts a containerised work behavior builder.
func (b *PipelineTestRunBuilder) ContainerisedWork(image string) *containerisedWorkBehaviorBuilder {
	return &containerisedWorkBehaviorBuilder{
		parent: b,
		image:  image,
	}
}

// CustomWork starts a custom work behavior builder.
func (b *PipelineTestRunBuilder) CustomWork(executionID string) *customWorkBehaviorBuilder {
	return &customWorkBehaviorBuilder{
		parent:      b,
		executionID: executionID,
	}
}

// DefaultBehavior starts a default behavior builder.
func (b *PipelineTestRunBuilder) DefaultBehavior() *defaultBehaviorBuilder {
	return &defaultBehaviorBuilder{parent: b}
}

// Build returns the PipelineTestRun.
func (b *PipelineTestRunBuilder) Build() *PipelineTestRun {
	if b.pipelineName == "" {
		panic("pipelineName is required")
	}
	if b.sourceDir == "" {
		cwd, _ := os.Getwd()
		b.sourceDir = cwd
	}
	return &PipelineTestRun{
		pipelineName: b.pipelineName,
		sourceDir:    b.sourceDir,
		behaviors:    b.behaviors,
		timeout:      b.timeout,
	}
}

// Start runs the pipeline test.
func (r *PipelineTestRun) Start() (*PipelineTestRunResult, error) {
	mockServer := newMockServer()

	// Start the mock server
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMockExecutionServiceServer(grpcServer, mockServer)

	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	// Get the actual port
	addr := lis.Addr().(*net.TCPAddr)
	mockAddress := fmt.Sprintf("localhost:%d", addr.Port)

	// Register behaviors
	r.behaviors.registerWith(mockServer)

	// Run the CLI
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yeetcd", "run",
		"--source", r.sourceDir,
		"--pipeline", r.pipelineName,
		"--mock-execution-engine-address", mockAddress,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("pipeline test timed out after %v", r.timeout)
		}
		// Non-zero exit code is not an error for the test framework
		// It's captured in the result
	}

	// Get executions from mock server
	executions := mockServer.getExecutions()
	matchedBehaviors := mockServer.getMatchedBehaviors()

	// Determine status
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	status := PipelineStatusSuccess
	if exitCode != 0 {
		status = PipelineStatusFailure
	}

	return &PipelineTestRunResult{
		Status:           status,
		ExitCode:         exitCode,
		Executions:       executions,
		MatchedBehaviors: matchedBehaviors,
		CLIOutput:        string(output),
	}, nil
}

// mockServer implements the MockExecutionService gRPC service.
type mockServer struct {
	pb.UnimplementedMockExecutionServiceServer
	mu                sync.Mutex
	executions        []*PipelineWorkExecution
	matchedBehaviors  []*MatchedBehavior
	containerisedResp map[string]*WorkResponse
	customResp        map[string]*WorkResponse
	defaultResp       *WorkResponse
}

func newMockServer() *mockServer {
	return &mockServer{
		containerisedResp: make(map[string]*WorkResponse),
		customResp:        make(map[string]*WorkResponse),
		defaultResp:       &WorkResponse{ExitCode: 0, Stdout: "", Stderr: ""},
	}
}

func (s *mockServer) registerContainerisedBehavior(image string, resp *WorkResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.containerisedResp[image] = resp
}

func (s *mockServer) registerCustomBehavior(executionID string, resp *WorkResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.customResp[executionID] = resp
}

func (s *mockServer) setDefaultResponse(resp *WorkResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defaultResp = resp
}

func (s *mockServer) getExecutions() []*PipelineWorkExecution {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]*PipelineWorkExecution, len(s.executions))
	copy(result, s.executions)
	return result
}

func (s *mockServer) getMatchedBehaviors() []*MatchedBehavior {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]*MatchedBehavior, len(s.matchedBehaviors))
	copy(result, s.matchedBehaviors)
	return result
}

func (s *mockServer) RunWork(ctx context.Context, req *pb.MockWorkRequest) (*pb.MockWorkResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	image := req.GetImage()
	cmd := req.GetCmd()
	envVars := req.GetEnvVars()
	workingDir := req.GetWorkingDir()

	var resp *WorkResponse
	var exec *PipelineWorkExecution
	var behaviorType WorkBehaviorType
	var matchKey string

	// Check if this is a build command (go mod tidy && go build)
	if isBuildCommand(cmd) {
		behaviorType = WorkBehaviorTypeContainerised
		matchKey = image
		resp = s.runBuildCommand(cmd, workingDir, envVars)
		exec = &PipelineWorkExecution{
			Type:       WorkBehaviorTypeContainerised,
			Image:      image,
			Cmd:        cmd,
			EnvVars:    envVars,
			WorkingDir: workingDir,
			ExitCode:   resp.ExitCode,
			Stdout:     resp.Stdout,
			Stderr:     resp.Stderr,
		}
	} else if isPipelineGenerator(cmd) {
		// Check if this is a pipeline generator call
		behaviorType = WorkBehaviorTypeContainerised
		matchKey = image
		resp = s.runPipelineGenerator(cmd, workingDir, envVars)
		exec = &PipelineWorkExecution{
			Type:       WorkBehaviorTypeContainerised,
			Image:      image,
			Cmd:        cmd,
			EnvVars:    envVars,
			WorkingDir: workingDir,
			ExitCode:   resp.ExitCode,
			Stdout:     resp.Stdout,
			Stderr:     resp.Stderr,
		}
	} else if isCustomWork(cmd) {
		executionID := extractExecutionID(cmd)
		behaviorType = WorkBehaviorTypeCustom
		matchKey = executionID

		resp = s.customResp[executionID]
		if resp == nil {
			resp = s.defaultResp
		}

		exec = &PipelineWorkExecution{
			Type:        WorkBehaviorTypeCustom,
			ExecutionID: executionID,
			Image:       image,
			Cmd:         cmd,
			EnvVars:     envVars,
			WorkingDir:  workingDir,
			ExitCode:    resp.ExitCode,
			Stdout:      resp.Stdout,
			Stderr:      resp.Stderr,
		}
	} else {
		behaviorType = WorkBehaviorTypeContainerised
		matchKey = image

		resp = s.containerisedResp[image]
		if resp == nil {
			resp = s.defaultResp
		}

		exec = &PipelineWorkExecution{
			Type:       WorkBehaviorTypeContainerised,
			Image:      image,
			Cmd:        cmd,
			EnvVars:    envVars,
			WorkingDir: workingDir,
			ExitCode:   resp.ExitCode,
			Stdout:     resp.Stdout,
			Stderr:     resp.Stderr,
		}
	}

	s.executions = append(s.executions, exec)
	s.matchedBehaviors = append(s.matchedBehaviors, &MatchedBehavior{
		Type:      behaviorType,
		MatchKey:  matchKey,
		Execution: exec,
		Response:  resp,
	})

	return &pb.MockWorkResponse{
		ExitCode: int32(resp.ExitCode),
		Stdout:   resp.Stdout,
		Stderr:   resp.Stderr,
	}, nil
}

func (s *mockServer) BuildImage(ctx context.Context, req *pb.MockImageBuildRequest) (*pb.MockImageBuildResponse, error) {
	return &pb.MockImageBuildResponse{
		Success:  true,
		ImageRef: req.GetImage() + ":" + req.GetTag(),
	}, nil
}

func (s *mockServer) runPipelineGenerator(cmd []string, workingDir string, envVars map[string]string) *WorkResponse {
	// For Go, we need to run the generator and capture its output
	// The generator outputs protobuf to stdout

	// Build the generator command
	generatorCmd := exec.Command("go", "run", "github.com/yeetcd/yeetcd/sdk/generator/cmd/generate")
	generatorCmd.Dir = workingDir

	// Set environment variables
	env := os.Environ()
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	generatorCmd.Env = env

	output, err := generatorCmd.CombinedOutput()
	if err != nil {
		// Include both the error and the output in the stderr
		stderr := fmt.Sprintf("Error: %v\nOutput: %s", err, string(output))
		return &WorkResponse{ExitCode: 1, Stdout: "", Stderr: stderr}
	}

	// Base64 encode the binary protobuf output for transmission over gRPC
	// (gRPC string fields must be valid UTF-8, but protobuf is binary)
	encodedOutput := base64.StdEncoding.EncodeToString(output)
	return &WorkResponse{ExitCode: 0, Stdout: encodedOutput, Stderr: ""}
}

func (s *mockServer) runBuildCommand(cmd []string, workingDir string, envVars map[string]string) *WorkResponse {
	// Run the build command locally
	// The build command is typically "go mod tidy && go build ./..."

	// For simplicity, we'll just run the build command as a shell command
	shellCmd := exec.Command("sh", "-c", strings.Join(cmd, " "))
	shellCmd.Dir = workingDir

	// Set environment variables
	env := os.Environ()
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	shellCmd.Env = env

	output, err := shellCmd.CombinedOutput()
	if err != nil {
		return &WorkResponse{ExitCode: 1, Stdout: "", Stderr: string(output)}
	}

	return &WorkResponse{ExitCode: 0, Stdout: string(output), Stderr: ""}
}

func isBuildCommand(cmd []string) bool {
	// Check if this is a build command
	// Build commands typically contain "go mod tidy" or "go build"
	for _, arg := range cmd {
		if strings.Contains(arg, "go mod tidy") || strings.Contains(arg, "go build") {
			return true
		}
	}
	return false
}

func isPipelineGenerator(cmd []string) bool {
	// Check if this is a generator call
	// For Go, the generator is run via "go run github.com/yeetcd/yeetcd/sdk/generator/cmd/generate"
	for _, arg := range cmd {
		if arg == "github.com/yeetcd/yeetcd/sdk/generator/cmd/generate" {
			return true
		}
	}
	// Empty cmd means it's the built image running the generator
	return len(cmd) == 0
}

func isCustomWork(cmd []string) bool {
	// For Go, custom work would be run via a specific command
	// This needs to match how the controller runs custom work
	return len(cmd) >= 1 && (cmd[0] == "go" || cmd[0] == "/custom-work-runner")
}

func extractExecutionID(cmd []string) string {
	// Extract execution ID from the command
	// This depends on how the controller passes the execution ID
	for i, arg := range cmd {
		if arg == "--execution-id" && i+1 < len(cmd) {
			return cmd[i+1]
		}
	}
	return ""
}

// behaviorChain manages a chain of behaviors.
type behaviorChain struct {
	containerised map[string]*WorkResponse
	custom        map[string]*WorkResponse
	defaultResp   *WorkResponse
}

func newBehaviorChain() *behaviorChain {
	return &behaviorChain{
		containerised: make(map[string]*WorkResponse),
		custom:        make(map[string]*WorkResponse),
		defaultResp:   &WorkResponse{ExitCode: 0, Stdout: "", Stderr: ""},
	}
}

func (c *behaviorChain) registerWith(server *mockServer) {
	for image, resp := range c.containerised {
		server.registerContainerisedBehavior(image, resp)
	}
	for executionID, resp := range c.custom {
		server.registerCustomBehavior(executionID, resp)
	}
	server.setDefaultResponse(c.defaultResp)
}

// containerisedWorkBehaviorBuilder builds a containerised work behavior.
type containerisedWorkBehaviorBuilder struct {
	parent *PipelineTestRunBuilder
	image  string
	resp   *WorkResponse
}

func (b *containerisedWorkBehaviorBuilder) WithExitCode(code int) *containerisedWorkBehaviorBuilder {
	b.resp = &WorkResponse{ExitCode: code, Stdout: "", Stderr: ""}
	return b
}

func (b *containerisedWorkBehaviorBuilder) WithResult(exitCode int, stdout, stderr string) *containerisedWorkBehaviorBuilder {
	b.resp = &WorkResponse{ExitCode: exitCode, Stdout: stdout, Stderr: stderr}
	return b
}

func (b *containerisedWorkBehaviorBuilder) Build() *PipelineTestRunBuilder {
	b.parent.behaviors.containerised[b.image] = b.resp
	return b.parent
}

// customWorkBehaviorBuilder builds a custom work behavior.
type customWorkBehaviorBuilder struct {
	parent      *PipelineTestRunBuilder
	executionID string
	resp        *WorkResponse
}

func (b *customWorkBehaviorBuilder) WithExitCode(code int) *customWorkBehaviorBuilder {
	b.resp = &WorkResponse{ExitCode: code, Stdout: "", Stderr: ""}
	return b
}

func (b *customWorkBehaviorBuilder) WithResult(exitCode int, stdout, stderr string) *customWorkBehaviorBuilder {
	b.resp = &WorkResponse{ExitCode: exitCode, Stdout: stdout, Stderr: stderr}
	return b
}

func (b *customWorkBehaviorBuilder) Build() *PipelineTestRunBuilder {
	b.parent.behaviors.custom[b.executionID] = b.resp
	return b.parent
}

// defaultBehaviorBuilder builds a default behavior.
type defaultBehaviorBuilder struct {
	parent *PipelineTestRunBuilder
	resp   *WorkResponse
}

func (b *defaultBehaviorBuilder) WithExitCode(code int) *defaultBehaviorBuilder {
	b.resp = &WorkResponse{ExitCode: code, Stdout: "", Stderr: ""}
	return b
}

func (b *defaultBehaviorBuilder) WithResult(exitCode int, stdout, stderr string) *defaultBehaviorBuilder {
	b.resp = &WorkResponse{ExitCode: exitCode, Stdout: stdout, Stderr: stderr}
	return b
}

func (b *defaultBehaviorBuilder) Build() *PipelineTestRunBuilder {
	b.parent.behaviors.defaultResp = b.resp
	return b.parent
}

// PipelineTestRunResult represents the result of a pipeline test run.
type PipelineTestRunResult struct {
	Status           PipelineStatus
	ExitCode         int
	Executions       []*PipelineWorkExecution
	MatchedBehaviors []*MatchedBehavior
	CLIOutput        string
}

// HasExecution checks if an execution with the given image exists.
func (r *PipelineTestRunResult) HasExecution(image string) bool {
	for _, exec := range r.Executions {
		if exec.Image == image {
			return true
		}
	}
	return false
}

// HasNoExecution checks if no execution with the given image exists.
func (r *PipelineTestRunResult) HasNoExecution(image string) bool {
	return !r.HasExecution(image)
}

// GetExecutionCount returns the number of executions with the given image.
func (r *PipelineTestRunResult) GetExecutionCount(image string) int {
	count := 0
	for _, exec := range r.Executions {
		if exec.Image == image {
			count++
		}
	}
	return count
}

// GetExecutions returns all executions.
func (r *PipelineTestRunResult) GetExecutions() []*PipelineWorkExecution {
	return r.Executions
}

// FindByImage returns all executions with the given image.
func (r *PipelineTestRunResult) FindByImage(image string) []*PipelineWorkExecution {
	var result []*PipelineWorkExecution
	for _, exec := range r.Executions {
		if exec.Image == image {
			result = append(result, exec)
		}
	}
	return result
}

// GetCustomExecutions returns all custom work executions.
func (r *PipelineTestRunResult) GetCustomExecutions() []*PipelineWorkExecution {
	var result []*PipelineWorkExecution
	for _, exec := range r.Executions {
		if exec.Type == WorkBehaviorTypeCustom {
			result = append(result, exec)
		}
	}
	return result
}

// PipelineStatus represents the status of a pipeline run.
type PipelineStatus string

const (
	PipelineStatusSuccess PipelineStatus = "SUCCESS"
	PipelineStatusFailure PipelineStatus = "FAILURE"
)

// WorkBehaviorType represents the type of work behavior.
type WorkBehaviorType string

const (
	WorkBehaviorTypeContainerised WorkBehaviorType = "CONTAINERISED"
	WorkBehaviorTypeCustom        WorkBehaviorType = "CUSTOM"
	WorkBehaviorTypeDynamic       WorkBehaviorType = "DYNAMIC"
)

// PipelineWorkExecution represents a single work execution.
type PipelineWorkExecution struct {
	Type       WorkBehaviorType
	Image      string
	Cmd        []string
	EnvVars    map[string]string
	WorkingDir string
	ExitCode   int
	Stdout     string
	Stderr     string
	// For custom work
	ExecutionID string
}

// MatchedBehavior represents a behavior that was matched during execution.
type MatchedBehavior struct {
	Type      WorkBehaviorType
	MatchKey  string
	Execution *PipelineWorkExecution
	Response  *WorkResponse
}

// WorkResponse represents the response from a work execution.
type WorkResponse struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// ReadAll reads all content from an io.Reader.
func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
