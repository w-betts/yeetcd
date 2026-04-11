package sample

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	sdk "github.com/yeetcd/yeetcd/sdk/pkg/yeetcd"
	sdktest "github.com/yeetcd/yeetcd/sdk/test/pkg/yeetcd"
)

// TestE2ESamplePipeline runs the sample pipeline end-to-end using FakePipelineRunner.
func TestE2ESamplePipeline(t *testing.T) {
	// Get the pipeline definition
	pipeline := SamplePipeline()

	// Create a fake runner with behavior for the containerised work
	runner := sdktest.NewFakePipelineRunner().
		WithWorkBehavior("containerised-work-definition", sdktest.NewContainerisedWorkBehavior("maven:3.9.9-eclipse-temurin-17").
			WithExecute(func() error {
				// In a real scenario, this would run the container
				// For testing, we just verify the behavior is set up
				return nil
			}).
			WithOutput("Hello from a containerised task\n").
			Build()).
		Build()

	// Run the pipeline (the runner doesn't actually use the pipeline definition yet,
	// but it demonstrates the test pattern)
	result, err := runner.Run(pipeline.Name, nil)

	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", result.Status)
	assert.Equal(t, 0, result.ExitCode)
	assert.Len(t, result.WorkExecutions, 1)
	assert.Equal(t, "containerised-work-definition", result.WorkExecutions[0].Description)
	assert.Equal(t, "Hello from a containerised task\n", result.Output)
}

// TestE2ESampleCompoundPipeline tests a compound pipeline with dependencies.
func TestE2ESampleCompoundPipeline(t *testing.T) {
	pipeline := SampleCompoundPipeline()

	runner := sdktest.NewFakePipelineRunner().
		WithWorkBehavior("sample-pipeline-work-1", sdktest.NewContainerisedWorkBehavior("alpine").
			WithOutput("Work 1 done\n").
			Build()).
		WithWorkBehavior("sample-pipeline-work-2", sdktest.NewContainerisedWorkBehavior("alpine").
			WithOutput("Work 2 done\n").
			Build()).
		Build()

	result, err := runner.Run(pipeline.Name, nil)

	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", result.Status)
	assert.Len(t, result.WorkExecutions, 2)
}

// TestE2ESampleWithCustomWork tests custom work execution.
func TestE2ESampleWithCustomWork(t *testing.T) {
	pipeline := SampleWithCustomWorkPipeline()

	executed := false
	runner := sdktest.NewFakePipelineRunner().
		WithWorkBehavior("custom-work", sdktest.NewCustomWorkBehavior().
			WithExecute(func() error {
				executed = true
				return nil
			}).
			Build()).
		Build()

	result, err := runner.Run(pipeline.Name, nil)

	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", result.Status)
	assert.True(t, executed)
}

// TestE2ESampleWithConditions tests conditional work execution.
func TestE2ESampleWithConditions(t *testing.T) {
	pipeline := SampleWithConditionsPipeline()

	// Default behavior for works that don't have specific behaviors
	runner := sdktest.NewFakePipelineRunner().
		WithDefaultBehavior(sdktest.NewDefaultWorkBehavior().
			WithStatus(sdktest.StatusSuccess).
			Build()).
		Build()

	result, err := runner.Run(pipeline.Name, nil)

	assert.NoError(t, err)
	assert.Equal(t, "SUCCESS", result.Status)
}

// TestE2EPipelineOutput demonstrates how to capture pipeline output.
func TestE2EPipelineOutput(t *testing.T) {
	pipeline := SamplePipeline()

	// Capture output in a buffer-like fashion
	var outputBuilder strings.Builder
	executed := false

	runner := sdktest.NewFakePipelineRunner().
		WithWorkBehavior("containerised-work-definition", sdktest.NewContainerisedWorkBehavior("alpine").
			WithExecute(func() error {
				executed = true
				outputBuilder.WriteString("Executing containerised work\n")
				return nil
			}).
			WithOutput("Executing containerised work\n").
			Build()).
		Build()

	result, err := runner.Run(pipeline.Name, nil)

	assert.NoError(t, err)
	assert.True(t, executed)
	assert.True(t, strings.Contains(result.Output, "containerised"))
}

// TestE2EFailureScenario tests how failures are handled.
func TestE2EFailureScenario(t *testing.T) {
	pipeline := SamplePipeline()

	runner := sdktest.NewFakePipelineRunner().
		WithWorkBehavior("containerised-work-definition", sdktest.NewContainerisedWorkBehavior("alpine").
			WithExecute(func() error {
				return &testError{message: "simulated failure"}
			}).
			WithStatus(sdktest.StatusFailure).
			Build()).
		Build()

	result, err := runner.Run(pipeline.Name, nil)

	assert.NoError(t, err) // Runner doesn't return error, it captures failure in result
	assert.Equal(t, "FAILURE", result.Status)
	assert.Equal(t, 1, result.ExitCode)
	assert.Equal(t, "FAILURE", result.WorkExecutions[0].Status)
}

// testError is a simple error for testing.
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// TestE2EAllPipelines verifies all pipeline functions can be loaded and run.
func TestE2EAllPipelines(t *testing.T) {
	pipelines := []struct {
		name   string
		getter func() interface{}
	}{
		{"sample", func() interface{} { return SamplePipeline() }},
		{"sampleCompound", func() interface{} { return SampleCompoundPipeline() }},
		{"sampleWithWorkContext", func() interface{} { return SampleWithWorkContextPipeline() }},
		{"sampleWithParameters", func() interface{} { return SampleWithParametersPipeline() }},
		{"sampleWithConditions", func() interface{} { return SampleWithConditionsPipeline() }},
		{"sampleWithCustomWork", func() interface{} { return SampleWithCustomWorkPipeline() }},
		{"sampleWithCompound", func() interface{} { return SampleWithCompoundPipeline() }},
	}

	for _, tc := range pipelines {
		t.Run(tc.name, func(t *testing.T) {
			// Call the getter to get the pipeline
			p := tc.getter().(sdk.Pipeline)

			// Verify pipeline has a name
			assert.NotEmpty(t, p.Name)

			// Verify pipeline has final work
			assert.NotEmpty(t, p.FinalWork)

			// Create runner with default behavior
			runner := sdktest.NewFakePipelineRunner().
				WithDefaultBehavior(sdktest.NewDefaultWorkBehavior().
					WithStatus(sdktest.StatusSuccess).
					Build()).
				Build()

			result, err := runner.Run(p.Name, nil)

			assert.NoError(t, err)
			assert.Equal(t, "SUCCESS", result.Status)
		})
	}
}
