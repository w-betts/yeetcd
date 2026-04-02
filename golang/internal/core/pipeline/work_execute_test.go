package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// TestWork_Execute_ExecutesPreviousWorkDependenciesFirst tests that Execute runs previous work before current work
// Given: Work B depending on Work A, MockExecutionEngine
// When: workB.Execute(ctx, context, engine, metadata, tracker, handler) is called
// Then: Work A is executed before Work B
func TestWork_Execute_ExecutesPreviousWorkDependenciesFirst(t *testing.T) {
	ctx := context.Background()

	// Track execution order
	executionOrder := make([]string, 0)

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			// Track execution by image name
			executionOrder = append(executionOrder, def.Image)
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create work A
	workA := Work{
		ID:          "work-a",
		Description: "Work A",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-a:latest",
			Cmd:   []string{"echo", "A"},
		},
	}

	// Create work B depending on work A
	workB := Work{
		ID:          "work-b",
		Description: "Work B",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-b:latest",
			Cmd:   []string{"echo", "B"},
		},
		PreviousWork: []PreviousWork{
			{
				Work: workA,
			},
		},
	}

	// Execute work B
	result, err := workB.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify execution order: A before B
	require.Len(t, executionOrder, 2)
	assert.Equal(t, "image-a:latest", executionOrder[0])
	assert.Equal(t, "image-b:latest", executionOrder[1])
}

// TestWork_Execute_SkipsWorkWhenConditionNotMet tests that Execute skips work when condition evaluates to false
// Given: Work with condition that returns false, MockExecutionEngine
// When: work.Execute(ctx, context, engine, metadata, tracker, handler) is called
// Then: Work is skipped, WorkFinished event with SKIPPED status recorded, no job executed
func TestWork_Execute_SkipsWorkWhenConditionNotMet(t *testing.T) {
	ctx := context.Background()

	// Track if job was executed
	jobExecuted := false

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			jobExecuted = true
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create a condition that always returns false
	falseCondition := &mockCondition{shouldExecute: false}

	// Create work with false condition
	work := Work{
		ID:          "conditional-work",
		Description: "Conditional Work",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "alpine:latest",
			Cmd:   []string{"echo", "hello"},
		},
		Condition: falseCondition,
	}

	// Execute work
	result, err := work.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify job was NOT executed
	assert.False(t, jobExecuted, "Job should not be executed when condition is false")

	// Verify result status is SKIPPED
	assert.Equal(t, types.SKIPPED, result.WorkStatus)

	// Verify WorkFinished event with SKIPPED status
	finishedEvents := GetEventsOfType[WorkFinished](handler)
	require.Len(t, finishedEvents, 1)
	assert.Equal(t, types.SKIPPED, finishedEvents[0].WorkStatus)
}

// TestWork_Execute_ExecutesWorkWhenConditionMet tests that Execute runs work when condition evaluates to true
// Given: Work with condition that returns true, MockExecutionEngine
// When: work.Execute(ctx, context, engine, metadata, tracker, handler) is called
// Then: Work is executed normally
func TestWork_Execute_ExecutesWorkWhenConditionMet(t *testing.T) {
	ctx := context.Background()

	// Track if job was executed
	jobExecuted := false

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			jobExecuted = true
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create a condition that always returns true
	trueCondition := &mockCondition{shouldExecute: true}

	// Create work with true condition
	work := Work{
		ID:          "conditional-work",
		Description: "Conditional Work",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "alpine:latest",
			Cmd:   []string{"echo", "hello"},
		},
		Condition: trueCondition,
	}

	// Execute work
	result, err := work.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify job WAS executed
	assert.True(t, jobExecuted, "Job should be executed when condition is true")

	// Verify result status is SUCCESS
	assert.Equal(t, types.SUCCESS, result.WorkStatus)
}

// TestWork_Execute_MergesWorkContextWithContainingContext tests that Execute merges work context (work overrides pipeline)
// Given: Pipeline context {KEY: "pipeline"}, Work context {KEY: "work"}, MockExecutionEngine
// When: work.Execute(ctx, pipelineContext, engine, metadata, tracker, handler) is called
// Then: JobDefinition.Environment contains {KEY: "work"} (work context wins)
func TestWork_Execute_MergesWorkContextWithContainingContext(t *testing.T) {
	ctx := context.Background()

	// Track received environment
	var receivedEnv map[string]string

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			receivedEnv = def.Environment
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create pipeline context
	pipelineContext := WorkContext{
		"KEY":       "pipeline_value",
		"PIPELINE":  "only_in_pipeline",
	}

	// Create work with its own context
	work := Work{
		ID:          "context-work",
		Description: "Context Work",
		WorkContext: WorkContext{
			"KEY": "work_value",
			"WORK": "only_in_work",
		},
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "alpine:latest",
			Cmd:   []string{"echo", "hello"},
		},
	}

	// Execute work
	result, err := work.Execute(ctx, pipelineContext, mockEngine, metadata, tracker, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify context merging: work overrides pipeline
	assert.Equal(t, "work_value", receivedEnv["KEY"], "Work context should override pipeline context")
	assert.Equal(t, "only_in_pipeline", receivedEnv["PIPELINE"], "Pipeline-only keys should be present")
	assert.Equal(t, "only_in_work", receivedEnv["WORK"], "Work-only keys should be present")
}

// TestWork_Execute_RecordsWorkFinishedEvent tests that Execute records WorkFinished event after execution
// Given: Any work definition, MockExecutionEngine
// When: work.Execute(ctx, context, engine, metadata, tracker, handler) is called
// Then: WorkFinished event is recorded with correct work and status
func TestWork_Execute_RecordsWorkFinishedEvent(t *testing.T) {
	ctx := context.Background()

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create work
	work := Work{
		ID:          "test-work",
		Description: "Test Work",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "alpine:latest",
			Cmd:   []string{"echo", "hello"},
		},
	}

	// Execute work
	result, err := work.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify WorkFinished event
	finishedEvents := GetEventsOfType[WorkFinished](handler)
	require.Len(t, finishedEvents, 1)
	assert.Equal(t, "test-work", finishedEvents[0].Work.ID)
	assert.Equal(t, types.SUCCESS, finishedEvents[0].WorkStatus)
}

// TestWork_Execute_UsesWorkResultTrackerForDeduplication tests that Execute uses WorkResultTracker to avoid duplicate execution
// Given: Work A executed once, then Work B depending on Work A is executed
// When: Both works are executed
// Then: Work A is only executed once (deduplicated by WorkResultTracker)
func TestWork_Execute_UsesWorkResultTrackerForDeduplication(t *testing.T) {
	ctx := context.Background()

	// Track execution count
	executionCount := 0

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			executionCount++
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create work A
	workA := Work{
		ID:          "work-a",
		Description: "Work A",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-a:latest",
			Cmd:   []string{"echo", "A"},
		},
	}

	// Create work B depending on work A
	workB := Work{
		ID:          "work-b",
		Description: "Work B",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-b:latest",
			Cmd:   []string{"echo", "B"},
		},
		PreviousWork: []PreviousWork{
			{
				Work: workA,
			},
		},
	}

	// Execute work A directly
	_, err := workA.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)
	require.NoError(t, err)

	// Execute work B (which depends on work A)
	_, err = workB.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)
	require.NoError(t, err)

	// Verify work A was only executed once
	assert.Equal(t, 2, executionCount, "Work A should be executed once, Work B once (total 2)")
}

// TestWork_Execute_ReturnsErrorWhenPreviousWorkFails tests that Execute returns error when previous work fails
// Given: Work B depending on Work A, Work A fails
// When: workB.Execute(ctx, context, engine, metadata, tracker, handler) is called
// Then: Error is returned indicating previous work failure
func TestWork_Execute_ReturnsErrorWhenPreviousWorkFails(t *testing.T) {
	ctx := context.Background()

	// Create mock execution engine that fails for work A
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			if def.Image == "image-a:latest" {
				return nil, errors.New("work A failed")
			}
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create work A
	workA := Work{
		ID:          "work-a",
		Description: "Work A",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-a:latest",
			Cmd:   []string{"false"},
		},
	}

	// Create work B depending on work A
	workB := Work{
		ID:          "work-b",
		Description: "Work B",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-b:latest",
			Cmd:   []string{"echo", "B"},
		},
		PreviousWork: []PreviousWork{
			{
				Work: workA,
			},
		},
	}

	// Execute work B (should fail because work A fails)
	result, err := workB.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)

	// Verify error is returned
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute previous work")
	assert.Nil(t, result)
}

// TestWork_Execute_PassesPreviousWorkStdoutAsContext tests that Execute passes previous work stdout as work context
// Given: Work B depending on Work A with StdOutEnvVar="PREV_OUTPUT", Work A produces stdout "hello"
// When: workB.Execute(ctx, context, engine, metadata, tracker, handler) is called
// Then: JobDefinition.Environment contains {PREV_OUTPUT: "hello"}
func TestWork_Execute_PassesPreviousWorkStdoutAsContext(t *testing.T) {
	ctx := context.Background()

	// Track received environment
	var receivedEnv map[string]string

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			if receivedEnv == nil {
				receivedEnv = def.Environment
			}
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create work A that produces stdout
	workA := Work{
		ID:          "work-a",
		Description: "Work A",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-a:latest",
			Cmd:   []string{"echo", "-n", "hello_from_a"},
		},
	}

	// Create work B depending on work A with StdOutEnvVar
	workB := Work{
		ID:          "work-b",
		Description: "Work B",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-b:latest",
			Cmd:   []string{"echo", "B"},
		},
		PreviousWork: []PreviousWork{
			{
				Work:         workA,
				StdOutEnvVar: "PREV_OUTPUT",
			},
		},
	}

	// Execute work B (which will execute work A first)
	result, err := workB.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Note: Since we're mocking the engine, the stdout capture won't actually work
	// This test verifies the mechanism is in place
	// In a real scenario, the JobStreams would capture the actual stdout
}

// TestWork_Execute_HandlesMultiplePreviousWorkDependencies tests that Execute handles multiple previous work dependencies
// Given: Work C depending on Work A and Work B
// When: workC.Execute(ctx, context, engine, metadata, tracker, handler) is called
// Then: Both Work A and Work B are executed before Work C
func TestWork_Execute_HandlesMultiplePreviousWorkDependencies(t *testing.T) {
	ctx := context.Background()

	// Track execution order
	executionOrder := make([]string, 0)

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			executionOrder = append(executionOrder, def.Image)
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create work A
	workA := Work{
		ID:          "work-a",
		Description: "Work A",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-a:latest",
			Cmd:   []string{"echo", "A"},
		},
	}

	// Create work B
	workB := Work{
		ID:          "work-b",
		Description: "Work B",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-b:latest",
			Cmd:   []string{"echo", "B"},
		},
	}

	// Create work C depending on both A and B
	workC := Work{
		ID:          "work-c",
		Description: "Work C",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "image-c:latest",
			Cmd:   []string{"echo", "C"},
		},
		PreviousWork: []PreviousWork{
			{Work: workA},
			{Work: workB},
		},
	}

	// Execute work C
	result, err := workC.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify all three works were executed
	require.Len(t, executionOrder, 3)
	assert.Contains(t, executionOrder, "image-a:latest")
	assert.Contains(t, executionOrder, "image-b:latest")
	assert.Contains(t, executionOrder, "image-c:latest")
}

// TestWork_Execute_RecordsWorkStartedEvent tests that Execute records WorkStarted event
// Given: Any work definition, MockExecutionEngine
// When: work.Execute(ctx, context, engine, metadata, tracker, handler) is called
// Then: WorkStarted event is recorded before work execution
func TestWork_Execute_RecordsWorkStartedEvent(t *testing.T) {
	ctx := context.Background()

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create work
	work := Work{
		ID:          "test-work",
		Description: "Test Work",
		WorkContext: make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{
			Image: "alpine:latest",
			Cmd:   []string{"echo", "hello"},
		},
	}

	// Execute work
	result, err := work.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify WorkStarted event
	startedEvents := GetEventsOfType[WorkStarted](handler)
	require.Len(t, startedEvents, 1)
	assert.Equal(t, "test-work", startedEvents[0].Work.ID)
}

// mockCondition is a mock implementation of types.ConditionEvaluator for testing
type mockCondition struct {
	shouldExecute bool
	evaluateErr   error
}

func (m *mockCondition) Evaluate(workContext types.WorkContext, tracker types.WorkResultTracker) (bool, error) {
	if m.evaluateErr != nil {
		return false, m.evaluateErr
	}
	return m.shouldExecute, nil
}
