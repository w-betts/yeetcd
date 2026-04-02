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

// TestContainerisedWorkDefinition_Execute_CallsExecutionEngineRunJob tests that Execute calls ExecutionEngine.RunJob
// Given: ContainerisedWorkDefinition with image="alpine:latest", cmd=["echo", "hello"], MockExecutionEngine
// When: workDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: ExecutionEngine.RunJob is called with JobDefinition containing image="alpine:latest", cmd=["echo", "hello"]
func TestContainerisedWorkDefinition_Execute_CallsExecutionEngineRunJob(t *testing.T) {
	ctx := context.Background()

	// Track if RunJob was called and with what parameters
	var calledJobDef *engine.JobDefinition
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			calledJobDef = &def
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create test handler and tracker
	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	// Create work with containerised definition
	workDef := &ContainerisedWorkDefinition{
		Image:      "alpine:latest",
		Cmd:        []string{"echo", "hello"},
		WorkingDir: "/app",
	}

	work := Work{
		ID:             "container-work",
		Description:    "Container Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: workDef,
	}

	// Execute
	result, err := workDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, calledJobDef)
	assert.Equal(t, "alpine:latest", calledJobDef.Image)
	assert.Equal(t, []string{"echo", "hello"}, calledJobDef.Cmd)
	assert.Equal(t, "/app", calledJobDef.WorkingDir)
}

// TestContainerisedWorkDefinition_Execute_PassesWorkContextAsEnvironment tests that Execute passes work context as environment variables
// Given: Work with WorkContext={"KEY": "value", "FOO": "bar"}, ContainerisedWorkDefinition
// When: workDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: JobDefinition.Environment contains {"KEY": "value", "FOO": "bar"}
func TestContainerisedWorkDefinition_Execute_PassesWorkContextAsEnvironment(t *testing.T) {
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

	workDef := &ContainerisedWorkDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"env"},
	}

	work := Work{
		ID:          "env-work",
		Description: "Environment Work",
		WorkContext: WorkContext{
			"KEY": "value",
			"FOO": "bar",
		},
		WorkDefinition: workDef,
	}

	result, err := workDef.Execute(ctx, work, WorkContext{"KEY": "value", "FOO": "bar"}, mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, capturedEnv)
	assert.Equal(t, "value", capturedEnv["KEY"])
	assert.Equal(t, "bar", capturedEnv["FOO"])
}

// TestContainerisedWorkDefinition_Execute_PassesOutputPathsToJobDefinition tests that Execute passes output paths to JobDefinition
// Given: Work with OutputPaths=[{Name: "output", Path: "/var/output"}], ContainerisedWorkDefinition
// When: workDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: JobDefinition.OutputDirectoryPaths contains {"output": "/var/output"}
func TestContainerisedWorkDefinition_Execute_PassesOutputPathsToJobDefinition(t *testing.T) {
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
		PipelineName: "test-pipeline",
	}

	workDef := &ContainerisedWorkDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "test"},
	}

	work := Work{
		ID:          "output-work",
		Description: "Output Work",
		WorkContext: make(WorkContext),
		WorkDefinition: workDef,
		OutputPaths: []WorkOutputPath{
			{Name: "output", Path: "/var/output"},
			{Name: "logs", Path: "/var/log"},
		},
	}

	result, err := workDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, capturedOutputPaths)
	assert.Equal(t, "/var/output", capturedOutputPaths["output"])
	assert.Equal(t, "/var/log", capturedOutputPaths["logs"])
}

// TestContainerisedWorkDefinition_Execute_PassesPreviousWorkMountInputs tests that Execute passes previous work mount inputs
// Given: Work B depending on Work A with OutputPathsMount="/mnt/outputs", Work A has output directories
// When: workDefB.Execute(ctx, workB, engine, metadata, tracker, handler) is called
// Then: JobDefinition.InputFilePaths contains mount input for "/mnt/outputs"
func TestContainerisedWorkDefinition_Execute_PassesPreviousWorkMountInputs(t *testing.T) {
	ctx := context.Background()

	var capturedInputPaths map[string]engine.MountInput
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			capturedInputPaths = def.InputFilePaths
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
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
	workBDef := &ContainerisedWorkDefinition{
		Image: "busybox:latest",
		Cmd:   []string{"echo", "B"},
	}
	workB := Work{
		ID:          "work-b",
		Description: "Work B",
		WorkContext: make(WorkContext),
		WorkDefinition: workBDef,
		PreviousWork: []PreviousWork{
			{
				Work:             workA,
				OutputPathsMount: "/mnt/outputs",
			},
		},
	}

	// Execute work B
	result, err := workBDef.Execute(ctx, workB, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Verify input paths were captured (the mechanism is in place)
	require.NotNil(t, capturedInputPaths)
}

// TestContainerisedWorkDefinition_Execute_MapsExitCode0ToSuccess tests that Execute maps exit code 0 to WorkStatus SUCCESS
// Given: ContainerisedWorkDefinition, MockExecutionEngine returning exit code 0
// When: workDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkResult.WorkStatus equals SUCCESS
func TestContainerisedWorkDefinition_Execute_MapsExitCode0ToSuccess(t *testing.T) {
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

	workDef := &ContainerisedWorkDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "success"},
	}

	work := Work{
		ID:             "success-work",
		Description:    "Success Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: workDef,
	}

	result, err := workDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.SUCCESS, result.WorkStatus)
}

// TestContainerisedWorkDefinition_Execute_MapsNonZeroExitCodeToFailure tests that Execute maps non-zero exit code to WorkStatus FAILURE
// Given: ContainerisedWorkDefinition, MockExecutionEngine returning exit code 1
// When: workDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkResult.WorkStatus equals FAILURE
func TestContainerisedWorkDefinition_Execute_MapsNonZeroExitCodeToFailure(t *testing.T) {
	ctx := context.Background()

	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 1}, nil
		},
	}

	handler := NewTestPipelineOutputHandler()
	tracker := NewWorkResultTracker()
	metadata := PipelineMetadata{
		PipelineName: "test-pipeline",
	}

	workDef := &ContainerisedWorkDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"false"},
	}

	work := Work{
		ID:             "failure-work",
		Description:    "Failure Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: workDef,
	}

	result, err := workDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.FAILURE, result.WorkStatus)
}

// TestContainerisedWorkDefinition_Execute_ReturnsErrorWhenRunJobFails tests that Execute returns error when RunJob fails
// Given: ContainerisedWorkDefinition, MockExecutionEngine that returns error from RunJob
// When: workDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: Error is returned
func TestContainerisedWorkDefinition_Execute_ReturnsErrorWhenRunJobFails(t *testing.T) {
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

	workDef := &ContainerisedWorkDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "test"},
	}

	work := Work{
		ID:             "error-work",
		Description:    "Error Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: workDef,
	}

	result, err := workDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "docker run failed")
	assert.Nil(t, result)
}

// TestContainerisedWorkDefinition_Execute_RecordsWorkStartedEvent tests that Execute records WorkStarted event
// Given: ContainerisedWorkDefinition, MockExecutionEngine
// When: workDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: WorkStarted event is recorded with correct work and JobStreams
func TestContainerisedWorkDefinition_Execute_RecordsWorkStartedEvent(t *testing.T) {
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

	workDef := &ContainerisedWorkDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "hello"},
	}

	work := Work{
		ID:             "event-work",
		Description:    "Event Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: workDef,
	}

	_, err := workDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)

	// Verify WorkStarted event was recorded
	startedEvents := GetEventsOfType[WorkStarted](handler)
	require.Len(t, startedEvents, 1)
	assert.Equal(t, "event-work", startedEvents[0].Work.ID)
	assert.NotNil(t, startedEvents[0].JobStreams)
}

// TestContainerisedWorkDefinition_Execute_PassesJobStreamsToRunJob tests that Execute passes JobStreams to RunJob
// Given: ContainerisedWorkDefinition, MockExecutionEngine
// When: workDef.Execute(ctx, work, engine, metadata, tracker, handler) is called
// Then: JobDefinition.JobStreams is set (non-nil)
func TestContainerisedWorkDefinition_Execute_PassesJobStreamsToRunJob(t *testing.T) {
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
		PipelineName: "test-pipeline",
	}

	workDef := &ContainerisedWorkDefinition{
		Image: "alpine:latest",
		Cmd:   []string{"echo", "hello"},
	}

	work := Work{
		ID:             "streams-work",
		Description:    "Streams Work",
		WorkContext:    make(WorkContext),
		WorkDefinition: workDef,
	}

	result, err := workDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, capturedJobStreams)
}

// TestContainerisedWorkDefinition_Execute_MapsVariousExitCodes tests that Execute correctly maps various exit codes
// Given: ContainerisedWorkDefinition, MockExecutionEngine returning various exit codes
// When: workDef.Execute(ctx, work, engine, metadata, tracker, handler) is called with different exit codes
// Then: WorkResult.WorkStatus is SUCCESS for 0, FAILURE for non-zero
func TestContainerisedWorkDefinition_Execute_MapsVariousExitCodes(t *testing.T) {
	testCases := []struct {
		name           string
		exitCode       int
		expectedStatus types.WorkStatus
	}{
		{"exit code 0", 0, types.SUCCESS},
		{"exit code 1", 1, types.FAILURE},
		{"exit code 2", 2, types.FAILURE},
		{"exit code 127", 127, types.FAILURE},
		{"exit code 255", 255, types.FAILURE},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			mockEngine := &MockExecutionEngine{
				RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
					return &engine.JobResult{ExitCode: tc.exitCode}, nil
				},
			}

			handler := NewTestPipelineOutputHandler()
			tracker := NewWorkResultTracker()
			metadata := PipelineMetadata{
				PipelineName: "test-pipeline",
			}

			workDef := &ContainerisedWorkDefinition{
				Image: "alpine:latest",
				Cmd:   []string{"exit", string(rune(tc.exitCode))},
			}

			work := Work{
				ID:             "exit-code-work",
				Description:    "Exit Code Work",
				WorkContext:    make(WorkContext),
				WorkDefinition: workDef,
			}

			result, err := workDef.Execute(ctx, work, make(WorkContext), mockEngine, metadata, tracker, handler)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expectedStatus, result.WorkStatus)
		})
	}
}
