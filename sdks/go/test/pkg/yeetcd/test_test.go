package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFakePipelineRunner(t *testing.T) {
	runner := NewFakePipelineRunner().Build()

	assert.NotNil(t, runner)
}

func TestFakePipelineRunnerWithWorkBehavior(t *testing.T) {
	executed := false

	runner := NewFakePipelineRunner().
		WithWorkBehavior("test-work", NewContainerisedWorkBehavior("alpine").
			WithExecute(func() error {
				executed = true
				return nil
			}).
			WithOutput("test output").
			Build()).
		Build()

	result, err := runner.Run("test-pipeline", nil)

	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", result.Status)
	assert.Equal(t, 0, result.ExitCode)
	assert.True(t, executed)
	assert.Len(t, result.WorkExecutions, 1)
	assert.Equal(t, "test-work", result.WorkExecutions[0].Description)
	assert.Equal(t, "test output", result.Output)
}

func TestFakePipelineRunnerWithDefaultBehavior(t *testing.T) {
	executed := false

	runner := NewFakePipelineRunner().
		WithDefaultBehavior(NewDefaultWorkBehavior().
			WithExecute(func() error {
				executed = true
				return nil
			}).
			Build()).
		Build()

	result, err := runner.Run("test-pipeline", nil)

	assert.NoError(t, err)
	assert.True(t, executed)
	assert.Len(t, result.WorkExecutions, 1)
}

func TestFakePipelineRunnerFailure(t *testing.T) {
	runner := NewFakePipelineRunner().
		WithWorkBehavior("failing-work", NewContainerisedWorkBehavior("alpine").
			WithExecute(func() error {
				return assert.AnError
			}).
			Build()).
		Build()

	result, err := runner.Run("test-pipeline", nil)

	assert.NoError(t, err)
	assert.Equal(t, "FAILURE", result.Status)
	assert.Equal(t, 1, result.ExitCode)
	assert.Equal(t, "FAILURE", result.WorkExecutions[0].Status)
}

func TestNewContainerisedWorkBehavior(t *testing.T) {
	behavior := NewContainerisedWorkBehavior("alpine").
		WithExecute(func() error { return nil }).
		WithOutput("output").
		WithStatus(StatusFailure).
		Build()

	assert.NotNil(t, behavior.Execute)
	assert.Equal(t, "output", behavior.Output)
	assert.Equal(t, StatusFailure, behavior.Status)
}

func TestNewCustomWorkBehavior(t *testing.T) {
	behavior := NewCustomWorkBehavior().
		WithExecute(func() error { return nil }).
		WithOutput("custom output").
		Build()

	assert.NotNil(t, behavior.Execute)
	assert.Equal(t, "custom output", behavior.Output)
}

func TestNewDefaultWorkBehavior(t *testing.T) {
	behavior := NewDefaultWorkBehavior().
		WithExecute(func() error { return nil }).
		WithStatus(StatusFailure).
		Build()

	assert.NotNil(t, behavior.Execute)
	assert.Equal(t, StatusFailure, behavior.Status)
}

func TestMockServer(t *testing.T) {
	server := NewMockServer(8080).Build()

	assert.Equal(t, 8080, server.GetPort())

	err := server.Start()
	assert.NoError(t, err)

	err = server.Stop()
	assert.NoError(t, err)
}
