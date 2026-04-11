// Package test provides testing utilities for Yeetcd pipelines in Go.
package test

import (
	"fmt"
	"sync"
)

// WorkExecution represents a single work execution.
type WorkExecution struct {
	WorkID      string
	Description string
	Status      string
}

// FakePipelineRunner is a test runner that executes pipelines without actual containers.
// It mocks the execution engine behavior.
type FakePipelineRunner struct {
	mu              sync.Mutex
	behaviors       map[string]*WorkBehavior
	defaultBehavior *DefaultWorkBehavior
}

// WorkBehavior defines the behavior for a specific work description.
type WorkBehavior struct {
	Execute func() error
	Output  string
	Status  WorkStatus
}

// WorkStatus represents the status of work execution.
type WorkStatus string

const (
	StatusSuccess WorkStatus = "SUCCESS"
	StatusFailure WorkStatus = "FAILURE"
)

// DefaultWorkBehavior defines default behavior for works without specific behaviors.
type DefaultWorkBehavior struct {
	Execute func() error
	Status  WorkStatus
}

// FakePipelineRunnerBuilder builds a FakePipelineRunner.
type FakePipelineRunnerBuilder struct {
	runner FakePipelineRunner
}

// NewFakePipelineRunner creates a new FakePipelineRunnerBuilder.
func NewFakePipelineRunner() *FakePipelineRunnerBuilder {
	return &FakePipelineRunnerBuilder{
		runner: FakePipelineRunner{
			behaviors:       make(map[string]*WorkBehavior),
			defaultBehavior: nil,
		},
	}
}

// WithWorkBehavior sets behavior for a specific work description.
func (b *FakePipelineRunnerBuilder) WithWorkBehavior(description string, behavior *WorkBehavior) *FakePipelineRunnerBuilder {
	b.runner.behaviors[description] = behavior
	return b
}

// WithDefaultBehavior sets default behavior for works without specific behaviors.
func (b *FakePipelineRunnerBuilder) WithDefaultBehavior(behavior *DefaultWorkBehavior) *FakePipelineRunnerBuilder {
	b.runner.defaultBehavior = behavior
	return b
}

// Build returns the FakePipelineRunner.
func (b *FakePipelineRunnerBuilder) Build() *FakePipelineRunner {
	return &b.runner
}

// FakePipelineRunResult represents the result of running a pipeline.
type FakePipelineRunResult struct {
	Status         string
	ExitCode       int
	WorkExecutions []FakeWorkExecution
	Output         string
}

// FakeWorkExecution represents a single work execution in the test runner.
type FakeWorkExecution struct {
	WorkID      string
	Description string
	Status      string
	Output      string
}

// Run executes a pipeline with the given arguments.
func (r *FakePipelineRunner) Run(pipelineName string, args map[string]string) (*FakePipelineRunResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find the pipeline by name
	// For now, we'll just run all the behaviors we can find

	executions := []FakeWorkExecution{}
	output := ""
	allSuccess := true

	// Execute each behavior
	for desc, behavior := range r.behaviors {
		exec := FakeWorkExecution{
			WorkID:      fmt.Sprintf("work-%s", desc),
			Description: desc,
			Status:      string(behavior.Status),
			Output:      behavior.Output,
		}

		if behavior.Execute != nil {
			err := behavior.Execute()
			if err != nil {
				exec.Status = string(StatusFailure)
				allSuccess = false
			}
		}

		output += behavior.Output
		executions = append(executions, exec)
	}

	// Also run default behavior if set
	if r.defaultBehavior != nil {
		exec := FakeWorkExecution{
			WorkID:      "work-default",
			Description: "default",
			Status:      string(r.defaultBehavior.Status),
		}

		if r.defaultBehavior.Execute != nil {
			err := r.defaultBehavior.Execute()
			if err != nil {
				exec.Status = string(StatusFailure)
				allSuccess = false
			}
		}

		executions = append(executions, exec)
	}

	status := "SUCCESS"
	exitCode := 0
	if !allSuccess {
		status = "FAILURE"
		exitCode = 1
	}

	return &FakePipelineRunResult{
		Status:         status,
		ExitCode:       exitCode,
		WorkExecutions: executions,
		Output:         output,
	}, nil
}

// ContainerisedWorkBehaviorBuilder builds behavior for containerised work.
type ContainerisedWorkBehaviorBuilder struct {
	behavior WorkBehavior
}

// NewContainerisedWorkBehavior creates a new ContainerisedWorkBehaviorBuilder.
func NewContainerisedWorkBehavior(image string) *ContainerisedWorkBehaviorBuilder {
	return &ContainerisedWorkBehaviorBuilder{
		behavior: WorkBehavior{Status: StatusSuccess},
	}
}

// WithExecute sets the execute function.
func (b *ContainerisedWorkBehaviorBuilder) WithExecute(execute func() error) *ContainerisedWorkBehaviorBuilder {
	b.behavior.Execute = execute
	return b
}

// WithOutput sets the output.
func (b *ContainerisedWorkBehaviorBuilder) WithOutput(output string) *ContainerisedWorkBehaviorBuilder {
	b.behavior.Output = output
	return b
}

// WithStatus sets the status.
func (b *ContainerisedWorkBehaviorBuilder) WithStatus(status WorkStatus) *ContainerisedWorkBehaviorBuilder {
	b.behavior.Status = status
	return b
}

// Build returns the WorkBehavior.
func (b *ContainerisedWorkBehaviorBuilder) Build() *WorkBehavior {
	return &b.behavior
}

// CustomWorkBehaviorBuilder builds behavior for custom work.
type CustomWorkBehaviorBuilder struct {
	behavior WorkBehavior
}

// NewCustomWorkBehavior creates a new CustomWorkBehaviorBuilder.
func NewCustomWorkBehavior() *CustomWorkBehaviorBuilder {
	return &CustomWorkBehaviorBuilder{
		behavior: WorkBehavior{Status: StatusSuccess},
	}
}

// WithExecute sets the execute function.
func (b *CustomWorkBehaviorBuilder) WithExecute(execute func() error) *CustomWorkBehaviorBuilder {
	b.behavior.Execute = execute
	return b
}

// WithOutput sets the output.
func (b *CustomWorkBehaviorBuilder) WithOutput(output string) *CustomWorkBehaviorBuilder {
	b.behavior.Output = output
	return b
}

// WithStatus sets the status.
func (b *CustomWorkBehaviorBuilder) WithStatus(status WorkStatus) *CustomWorkBehaviorBuilder {
	b.behavior.Status = status
	return b
}

// Build returns the WorkBehavior.
func (b *CustomWorkBehaviorBuilder) Build() *WorkBehavior {
	return &b.behavior
}

// DefaultWorkBehaviorBuilder builds default work behavior.
type DefaultWorkBehaviorBuilder struct {
	behavior DefaultWorkBehavior
}

// NewDefaultWorkBehavior creates a new DefaultWorkBehaviorBuilder.
func NewDefaultWorkBehavior() *DefaultWorkBehaviorBuilder {
	return &DefaultWorkBehaviorBuilder{
		behavior: DefaultWorkBehavior{Status: StatusSuccess},
	}
}

// WithExecute sets the execute function.
func (b *DefaultWorkBehaviorBuilder) WithExecute(execute func() error) *DefaultWorkBehaviorBuilder {
	b.behavior.Execute = execute
	return b
}

// WithStatus sets the status.
func (b *DefaultWorkBehaviorBuilder) WithStatus(status WorkStatus) *DefaultWorkBehaviorBuilder {
	b.behavior.Status = status
	return b
}

// Build returns the DefaultWorkBehavior.
func (b *DefaultWorkBehaviorBuilder) Build() *DefaultWorkBehavior {
	return &b.behavior
}

// MockServer is a mock execution engine server for integration tests.
// In a full implementation, this would be a real server that accepts work execution requests.
type MockServer struct {
	port       int
	executions []WorkExecution
}

// MockServerBuilder builds a MockServer.
type MockServerBuilder struct {
	server MockServer
}

// NewMockServer creates a new MockServerBuilder.
func NewMockServer(port int) *MockServerBuilder {
	return &MockServerBuilder{
		server: MockServer{
			port:       port,
			executions: []WorkExecution{},
		},
	}
}

// Build returns the MockServer.
func (b *MockServerBuilder) Build() *MockServer {
	return &b.server
}

// Start starts the mock server.
func (s *MockServer) Start() error {
	// Placeholder - would start a real server
	return nil
}

// Stop stops the mock server.
func (s *MockServer) Stop() error {
	// Placeholder - would stop the server
	return nil
}

// GetPort returns the server port.
func (s *MockServer) GetPort() int {
	return s.port
}
