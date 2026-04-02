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

// TestCompoundWorkDefinition_Execute_ExecutesAllFinalWorkItems tests that Execute executes all final work items
// Given: CompoundWorkDefinition with two final work items, MockExecutionEngine
// When: compoundDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: Both final work items are executed
func TestCompoundWorkDefinition_Execute_ExecutesAllFinalWorkItems(t *testing.T) {
	ctx := context.Background()

	executionCount := 0
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			executionCount++
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create two work items
	workA := Work{
		ID:             "work-a",
		Description:    "Work A",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "A"}},
	}

	workB := Work{
		ID:             "work-b",
		Description:    "Work B",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "B"}},
	}

	// Create compound work definition
	compoundDef := &CompoundWorkDefinition{
		FinalWork: []Work{workA, workB},
	}

	compoundWork := Work{
		ID:             "compound-work",
		Description:    "Compound Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: compoundDef,
	}

	result, err := compoundDef.Execute(ctx, compoundWork, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, executionCount, "Both final work items should be executed")
}

// TestCompoundWorkDefinition_Execute_ReturnsSuccessWhenAllSucceed tests that Execute returns SUCCESS when all final work succeeds
// Given: CompoundWorkDefinition with two final work items, both succeed
// When: compoundDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkResult.WorkStatus equals SUCCESS
func TestCompoundWorkDefinition_Execute_ReturnsSuccessWhenAllSucceed(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	workA := Work{
		ID:             "work-a",
		Description:    "Work A",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "A"}},
	}

	workB := Work{
		ID:             "work-b",
		Description:    "Work B",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "B"}},
	}

	compoundDef := &CompoundWorkDefinition{
		FinalWork: []Work{workA, workB},
	}

	compoundWork := Work{
		ID:             "compound-work",
		Description:    "Compound Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: compoundDef,
	}

	result, err := compoundDef.Execute(ctx, compoundWork, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.SUCCESS, result.WorkStatus)
}

// TestCompoundWorkDefinition_Execute_ReturnsFailureWhenAnyFails tests that Execute returns FAILURE when any final work fails
// Given: CompoundWorkDefinition with two final work items, one fails
// When: compoundDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkResult.WorkStatus equals FAILURE
func TestCompoundWorkDefinition_Execute_ReturnsFailureWhenAnyFails(t *testing.T) {
	ctx := context.Background()

	callCount := 0
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			callCount++
			if callCount == 2 {
				// Second work item fails
				return &engine.JobResult{ExitCode: 1}, nil
			}
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	workA := Work{
		ID:             "work-a",
		Description:    "Work A",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "A"}},
	}

	workB := Work{
		ID:             "work-b",
		Description:    "Work B",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"false"}},
	}

	compoundDef := &CompoundWorkDefinition{
		FinalWork: []Work{workA, workB},
	}

	compoundWork := Work{
		ID:             "compound-work",
		Description:    "Compound Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: compoundDef,
	}

	result, err := compoundDef.Execute(ctx, compoundWork, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.FAILURE, result.WorkStatus)
}

// TestCompoundWorkDefinition_Execute_RecordsWorkStartedEvent tests that Execute records WorkStarted event
// Given: CompoundWorkDefinition, MockExecutionEngine
// When: compoundDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkStarted event is recorded with correct work and nil JobStreams
func TestCompoundWorkDefinition_Execute_RecordsWorkStartedEvent(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	workA := Work{
		ID:             "work-a",
		Description:    "Work A",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "A"}},
	}

	compoundDef := &CompoundWorkDefinition{
		FinalWork: []Work{workA},
	}

	compoundWork := Work{
		ID:             "compound-work",
		Description:    "Compound Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: compoundDef,
	}

	_, err := compoundDef.Execute(ctx, compoundWork, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)

	// Verify WorkStarted event was recorded for compound work
	startedEvents := GetEventsOfType[WorkStarted](handler)
	require.Len(t, startedEvents, 2) // One for compound work, one for final work
	// First event should be for compound work with nil JobStreams
	assert.Equal(t, "compound-work", startedEvents[0].Work.ID)
	assert.Nil(t, startedEvents[0].JobStreams, "CompoundWorkDefinition should have nil JobStreams")
	// Second event should be for the final work item
	assert.Equal(t, "work-a", startedEvents[1].Work.ID)
	assert.NotNil(t, startedEvents[1].JobStreams, "Final work should have non-nil JobStreams")
}

// TestCompoundWorkDefinition_Execute_ReturnsErrorWhenFinalWorkFails tests that Execute returns error when final work execution fails
// Given: CompoundWorkDefinition, MockExecutionEngine that returns error from RunJob
// When: compoundDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: Error is returned
func TestCompoundWorkDefinition_Execute_ReturnsErrorWhenFinalWorkFails(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return nil, errors.New("docker run failed")
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	workA := Work{
		ID:             "work-a",
		Description:    "Work A",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "A"}},
	}

	compoundDef := &CompoundWorkDefinition{
		FinalWork: []Work{workA},
	}

	compoundWork := Work{
		ID:             "compound-work",
		Description:    "Compound Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: compoundDef,
	}

	result, err := compoundDef.Execute(ctx, compoundWork, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute final work")
	assert.Contains(t, err.Error(), "docker run failed")
	assert.Nil(t, result)
}

// TestCompoundWorkDefinition_Execute_ContinuesExecutionAfterFailure tests that Execute continues executing all final work even after one fails
// Given: CompoundWorkDefinition with three final work items, second one fails
// When: compoundDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: All three work items are executed
func TestCompoundWorkDefinition_Execute_ContinuesExecutionAfterFailure(t *testing.T) {
	ctx := context.Background()

	executionCount := 0
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			executionCount++
			if executionCount == 2 {
				return &engine.JobResult{ExitCode: 1}, nil
			}
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	workA := Work{
		ID:             "work-a",
		Description:    "Work A",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "A"}},
	}

	workB := Work{
		ID:             "work-b",
		Description:    "Work B",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"false"}},
	}

	workC := Work{
		ID:             "work-c",
		Description:    "Work C",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "C"}},
	}

	compoundDef := &CompoundWorkDefinition{
		FinalWork: []Work{workA, workB, workC},
	}

	compoundWork := Work{
		ID:             "compound-work",
		Description:    "Compound Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: compoundDef,
	}

	result, err := compoundDef.Execute(ctx, compoundWork, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 3, executionCount, "All three work items should be executed")
	assert.Equal(t, types.FAILURE, result.WorkStatus)
}

// TestCompoundWorkDefinition_Execute_ReturnsSuccessWhenAllSkipped tests that Execute returns SUCCESS when all final work is skipped
// Given: CompoundWorkDefinition with two final work items, both return SKIPPED status
// When: compoundDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkResult.WorkStatus equals SUCCESS
func TestCompoundWorkDefinition_Execute_ReturnsSuccessWhenAllSkipped(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			// This won't be called for skipped work
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create work items that will be marked as skipped in tracker
	workA := Work{
		ID:             "work-a",
		Description:    "Work A",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "A"}},
	}

	workB := Work{
		ID:             "work-b",
		Description:    "Work B",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"echo", "B"}},
	}

	// Pre-populate tracker with SKIPPED results
	tracker.results["work-a"] = &types.WorkResult{WorkStatus: types.SKIPPED}
	tracker.results["work-b"] = &types.WorkResult{WorkStatus: types.SKIPPED}

	compoundDef := &CompoundWorkDefinition{
		FinalWork: []Work{workA, workB},
	}

	compoundWork := Work{
		ID:             "compound-work",
		Description:    "Compound Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: compoundDef,
	}

	result, err := compoundDef.Execute(ctx, compoundWork, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.SUCCESS, result.WorkStatus)
}

// TestCompoundWorkDefinition_Execute_HandlesEmptyFinalWork tests that Execute handles empty FinalWork slice
// Given: CompoundWorkDefinition with empty FinalWork slice
// When: compoundDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkResult.WorkStatus equals SUCCESS (no work to fail)
func TestCompoundWorkDefinition_Execute_HandlesEmptyFinalWork(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	compoundDef := &CompoundWorkDefinition{
		FinalWork: []Work{},
	}

	compoundWork := Work{
		ID:             "compound-work",
		Description:    "Compound Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: compoundDef,
	}

	result, err := compoundDef.Execute(ctx, compoundWork, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.SUCCESS, result.WorkStatus, "Empty compound work should succeed")
}

// TestCompoundWorkDefinition_Execute_PassesMergedContextToFinalWork tests that Execute passes merged context to final work items
// Given: CompoundWorkDefinition with final work, mergedContext={"KEY": "value"}
// When: compoundDef.Execute(ctx, work, mergedContext, engine, metadata, tracker, handler) is called
// Then: Final work items receive the merged context
func TestCompoundWorkDefinition_Execute_PassesMergedContextToFinalWork(t *testing.T) {
	ctx := context.Background()

	var capturedEnv map[string]string
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			capturedEnv = def.Environment
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	workA := Work{
		ID:             "work-a",
		Description:    "Work A",
		WorkContext:    make(WorkContext),
		WorkDefinition: &ContainerisedWorkDefinition{Image: "alpine:latest", Cmd: []string{"env"}},
	}

	compoundDef := &CompoundWorkDefinition{
		FinalWork: []Work{workA},
	}

	compoundWork := Work{
		ID:             "compound-work",
		Description:    "Compound Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: compoundDef,
	}

	mergedContext := WorkContext{"KEY": "value", "FOO": "bar"}

	result, err := compoundDef.Execute(ctx, compoundWork, mergedContext, mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, capturedEnv)
	assert.Equal(t, "value", capturedEnv["KEY"])
	assert.Equal(t, "bar", capturedEnv["FOO"])
}
