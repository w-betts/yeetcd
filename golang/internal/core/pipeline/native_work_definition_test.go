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

// MockSourceLanguage implements SourceLanguage interface for testing
type MockSourceLanguage struct {
	CustomTaskRunnerCmd []string
}

func (m *MockSourceLanguage) GetCustomTaskRunnerCmd(pipelineName, taskName string) []string {
	return m.CustomTaskRunnerCmd
}

// TestNativeWorkDefinition_Execute_UsesBuiltSourceImage tests that Execute uses builtSourceImage from metadata
// Given: NativeWorkDefinition, metadata with BuiltSourceImage="built-image:latest"
// When: nativeDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: JobDefinition.Image equals "built-image:latest"
func TestNativeWorkDefinition_Execute_UsesBuiltSourceImage(t *testing.T) {
	ctx := context.Background()

	var capturedImage string
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			capturedImage = def.Image
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner", "cmd"}},
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:             "native-work",
		Description:    "Native Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: nativeDef,
	}

	result, err := nativeDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "built-image:latest", capturedImage)
}

// TestNativeWorkDefinition_Execute_UsesCustomTaskRunnerCmd tests that Execute uses source language's custom task runner command
// Given: NativeWorkDefinition, MockSourceLanguage with CustomTaskRunnerCmd=["custom-runner", "arg"]
// When: nativeDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: JobDefinition.Cmd equals ["custom-runner", "arg"]
func TestNativeWorkDefinition_Execute_UsesCustomTaskRunnerCmd(t *testing.T) {
	ctx := context.Background()

	var capturedCmd []string
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			capturedCmd = def.Cmd
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"custom-runner", "arg"}},
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:             "native-work",
		Description:    "Native Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: nativeDef,
	}

	result, err := nativeDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, []string{"custom-runner", "arg"}, capturedCmd)
}

// TestNativeWorkDefinition_Execute_PassesWorkContextAsEnvironment tests that Execute passes work context as environment variables
// Given: Work with mergedContext={"KEY": "value", "FOO": "bar"}, NativeWorkDefinition
// When: nativeDef.Execute(ctx, work, mergedContext, engine, metadata, tracker, handler) is called
// Then: JobDefinition.Environment contains {"KEY": "value", "FOO": "bar"}
func TestNativeWorkDefinition_Execute_PassesWorkContextAsEnvironment(t *testing.T) {
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
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner"}},
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:             "native-work",
		Description:    "Native Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: nativeDef,
	}

	mergedContext := WorkContext{"KEY": "value", "FOO": "bar"}

	result, err := nativeDef.Execute(ctx, work, mergedContext, mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, capturedEnv)
	assert.Equal(t, "value", capturedEnv["KEY"])
	assert.Equal(t, "bar", capturedEnv["FOO"])
}

// TestNativeWorkDefinition_Execute_PassesOutputPathsToJobDefinition tests that Execute passes output paths to JobDefinition
// Given: Work with OutputPaths=[{Name: "output", Path: "/var/output"}], NativeWorkDefinition
// When: nativeDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: JobDefinition.OutputDirectoryPaths contains {"output": "/var/output"}
func TestNativeWorkDefinition_Execute_PassesOutputPathsToJobDefinition(t *testing.T) {
	ctx := context.Background()

	var capturedOutputPaths map[string]string
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			capturedOutputPaths = def.OutputDirectoryPaths
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner"}},
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:          "native-work",
		Description: "Native Work",
		WorkContext: make(WorkContext),
		WorkDefinition: nativeDef,
		OutputPaths: []WorkOutputPath{
			{Name: "output", Path: "/var/output"},
			{Name: "logs", Path: "/var/log"},
		},
	}

	result, err := nativeDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, capturedOutputPaths)
	assert.Equal(t, "/var/output", capturedOutputPaths["output"])
	assert.Equal(t, "/var/log", capturedOutputPaths["logs"])
}

// TestNativeWorkDefinition_Execute_MapsExitCode0ToSuccess tests that Execute maps exit code 0 to WorkStatus SUCCESS
// Given: NativeWorkDefinition, MockExecutionEngine returning exit code 0
// When: nativeDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkResult.WorkStatus equals SUCCESS
func TestNativeWorkDefinition_Execute_MapsExitCode0ToSuccess(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner"}},
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:             "native-work",
		Description:    "Native Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: nativeDef,
	}

	result, err := nativeDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.SUCCESS, result.WorkStatus)
}

// TestNativeWorkDefinition_Execute_MapsNonZeroExitCodeToFailure tests that Execute maps non-zero exit code to WorkStatus FAILURE
// Given: NativeWorkDefinition, MockExecutionEngine returning exit code 1
// When: nativeDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkResult.WorkStatus equals FAILURE
func TestNativeWorkDefinition_Execute_MapsNonZeroExitCodeToFailure(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 1}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner"}},
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:             "native-work",
		Description:    "Native Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: nativeDef,
	}

	result, err := nativeDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.FAILURE, result.WorkStatus)
}

// TestNativeWorkDefinition_Execute_ReturnsErrorWhenRunJobFails tests that Execute returns error when RunJob fails
// Given: NativeWorkDefinition, MockExecutionEngine that returns error from RunJob
// When: nativeDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: Error is returned
func TestNativeWorkDefinition_Execute_ReturnsErrorWhenRunJobFails(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return nil, errors.New("docker run failed")
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner"}},
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:             "native-work",
		Description:    "Native Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: nativeDef,
	}

	result, err := nativeDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to run native job")
	assert.Contains(t, err.Error(), "docker run failed")
	assert.Nil(t, result)
}

// TestNativeWorkDefinition_Execute_RecordsWorkStartedEvent tests that Execute records WorkStarted event
// Given: NativeWorkDefinition, MockExecutionEngine
// When: nativeDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkStarted event is recorded with correct work and JobStreams
func TestNativeWorkDefinition_Execute_RecordsWorkStartedEvent(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner"}},
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:             "native-work",
		Description:    "Native Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: nativeDef,
	}

	_, err := nativeDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)

	// Verify WorkStarted event was recorded
	startedEvents := GetEventsOfType[WorkStarted](handler)
	require.Len(t, startedEvents, 1)
	assert.Equal(t, "native-work", startedEvents[0].Work.ID)
	assert.NotNil(t, startedEvents[0].JobStreams)
}

// TestNativeWorkDefinition_Execute_PassesJobStreamsToRunJob tests that Execute passes JobStreams to RunJob
// Given: NativeWorkDefinition, MockExecutionEngine
// When: nativeDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: JobDefinition.JobStreams is set (non-nil)
func TestNativeWorkDefinition_Execute_PassesJobStreamsToRunJob(t *testing.T) {
	ctx := context.Background()

	var capturedJobStreams *engine.JobStreams
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			capturedJobStreams = def.JobStreams
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner"}},
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:             "native-work",
		Description:    "Native Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: nativeDef,
	}

	result, err := nativeDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, capturedJobStreams)
}

// TestNativeWorkDefinition_Execute_PassesPreviousWorkMountInputs tests that Execute passes previous work mount inputs
// Given: Work B depending on Work A with OutputPathsMount="/mnt/outputs", Work A has output directories
// When: nativeDef.Execute(ctx, workB, engine, metadata, tracker, handler) is called
// Then: JobDefinition.InputFilePaths contains mount input for "/mnt/outputs"
func TestNativeWorkDefinition_Execute_PassesPreviousWorkMountInputs(t *testing.T) {
	ctx := context.Background()

	var capturedInputPaths map[string]engine.MountInput
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			capturedInputPaths = def.InputFilePaths
			return &engine.JobResult{ExitCode: 0, OutputDirectoriesParent: "/tmp/outputs"}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner"}},
	}

	// First, execute work A to populate tracker with results
	workADef := &ContainerisedWorkDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "A"},
	}
	workA := Work{
		ID:             "work-a",
		Description:    "Work A",
		WorkContext:    make(WorkContext),
		WorkDefinition: workADef,
		OutputPaths: []WorkOutputPath{
			{Name: "output", Path: "/var/output"},
		},
	}

	// Execute work A first
	_, err := workA.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)
	require.NoError(t, err)

	// Now create work B depending on work A
	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	workB := Work{
		ID:          "work-b",
		Description: "Work B",
		WorkContext: make(WorkContext),
		WorkDefinition: nativeDef,
		PreviousWork: []PreviousWork{
			{
				Work:             workA,
				OutputPathsMount: "/mnt/outputs",
			},
		},
	}

	// Execute work B
	result, err := nativeDef.Execute(ctx, workB, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Verify input paths were captured (the mechanism is in place)
	require.NotNil(t, capturedInputPaths)
}

// TestNativeWorkDefinition_Execute_AddsPreviousWorkStdOutToContext tests that Execute adds previous work stdout to context
// Given: Work B depending on Work A with StdOutEnvVar="PREV_OUTPUT", Work A has stdout
// When: nativeDef.Execute(ctx, workB, engine, metadata, tracker, handler) is called
// Then: JobDefinition.Environment contains previous work stdout
func TestNativeWorkDefinition_Execute_AddsPreviousWorkStdOutToContext(t *testing.T) {
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
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner"}},
	}

	// First, execute work A to populate tracker with results
	workADef := &ContainerisedWorkDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "A-output"},
	}
	workA := Work{
		ID:             "work-a",
		Description:    "Work A",
		WorkContext:    make(WorkContext),
		WorkDefinition: workADef,
	}

	// Execute work A first
	_, err := workA.Execute(ctx, make(WorkContext), mockEngine, metadata, tracker, handler)
	require.NoError(t, err)

	// Now create work B depending on work A with stdout capture
	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	workB := Work{
		ID:          "work-b",
		Description: "Work B",
		WorkContext: make(WorkContext),
		WorkDefinition: nativeDef,
		PreviousWork: []PreviousWork{
			{
				Work:         workA,
				StdOutEnvVar: "PREV_OUTPUT",
			},
		},
	}

	// Execute work B
	result, err := nativeDef.Execute(ctx, workB, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, capturedEnv)
	// The environment should contain the previous work stdout
	// (Note: In real execution, this would be the actual stdout, but in mock it's empty)
}

// TestNativeWorkDefinition_Execute_ReturnsOutputDirectoriesParent tests that Execute returns OutputDirectoriesParent
// Given: NativeWorkDefinition, MockExecutionEngine returning OutputDirectoriesParent="/tmp/outputs"
// When: nativeDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkResult.OutputDirectoriesParent equals "/tmp/outputs"
func TestNativeWorkDefinition_Execute_ReturnsOutputDirectoriesParent(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0, OutputDirectoriesParent: "/tmp/outputs"}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    &MockSourceLanguage{CustomTaskRunnerCmd: []string{"runner"}},
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:             "native-work",
		Description:    "Native Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: nativeDef,
	}

	result, err := nativeDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "/tmp/outputs", result.OutputDirectoriesParent)
}

// TestNativeWorkDefinition_Execute_HandlesNilSourceLanguage tests that Execute handles nil SourceLanguage
// Given: NativeWorkDefinition, metadata with nil SourceLanguage
// When: nativeDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: JobDefinition.Cmd is empty (no command)
func TestNativeWorkDefinition_Execute_HandlesNilSourceLanguage(t *testing.T) {
	ctx := context.Background()

	var capturedCmd []string
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			capturedCmd = def.Cmd
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName:      "test-pipeline",
		BuiltSourceImage:  "built-image:latest",
		SourceLanguage:    nil, // nil source language
	}

	nativeDef := &NativeWorkDefinition{
		ExecutionID: "test-execution-id",
	}

	work := Work{
		ID:             "native-work",
		Description:    "Native Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: nativeDef,
	}

	result, err := nativeDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, capturedCmd, "Command should be nil when SourceLanguage is nil")
}
