package pipeline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/yeetcd/yeetcd/pkg/proto/pipeline"
	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/pkg/build"
	"github.com/yeetcd/yeetcd/pkg/config"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// mockSourceExtractor is a no-op source extractor for tests that don't use Assemble
var mockSourceExtractor = build.NewSourceExtractor()

// MockExecutionEngine is a mock implementation of engine.ExecutionEngine for testing
type MockExecutionEngine struct {
	BuildImageFunc  func(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error)
	RemoveImageFunc func(ctx context.Context, imageID string) error
	RunJobFunc      func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error)
}

func (m *MockExecutionEngine) BuildImage(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
	if m.BuildImageFunc != nil {
		return m.BuildImageFunc(ctx, def)
	}
	return nil, nil
}

func (m *MockExecutionEngine) RemoveImage(ctx context.Context, imageID string) error {
	if m.RemoveImageFunc != nil {
		return m.RemoveImageFunc(ctx, imageID)
	}
	return nil
}

func (m *MockExecutionEngine) RunJob(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
	if m.RunJobFunc != nil {
		return m.RunJobFunc(ctx, def)
	}
	return nil, nil
}

// MockBuildService is a mock implementation of build.BuildService for testing
type MockBuildService struct {
	BuildFunc func(ctx context.Context, source build.Source) (*build.BuildResult, error)
}

func (m *MockBuildService) Build(ctx context.Context, source build.Source) (*build.BuildResult, error) {
	if m.BuildFunc != nil {
		return m.BuildFunc(ctx, source)
	}
	return nil, nil
}

// TestPipelineController_Assemble_PopulatesBuiltSourceImage tests that Assemble populates BuiltSourceImage in PipelineMetadata
// Given: A MockBuildService that returns pipelines with source build results containing image IDs
// When: Assemble(ctx, source) is called
// Then: Each pipeline's Metadata.BuiltSourceImage is populated with the corresponding image ID from SourceBuildResult
func TestPipelineController_Assemble_PopulatesBuiltSourceImage(t *testing.T) {
	ctx := context.Background()

	// Create mock build service that returns pipelines with image IDs
	mockBuildService := &MockBuildService{
		BuildFunc: func(ctx context.Context, source build.Source) (*build.BuildResult, error) {
			// Create mock pipelines and source build results with image IDs
			pipelines := []*pb.Pipeline{
				{Name: "pipeline-1"},
				{Name: "pipeline-2"},
			}
			sourceBuildResults := []build.SourceBuildResult{
				{ImageID: "sha256:abc123", YeetcdConfig: config.YeetcdConfig{Language: config.SourceLanguageJava}},
				{ImageID: "sha256:def456", YeetcdConfig: config.YeetcdConfig{Language: config.SourceLanguageGo}},
			}
			return &build.BuildResult{
				Pipelines:          pipelines,
				SourceBuildResults: sourceBuildResults,
			}, nil
		},
	}

	// Create pipeline controller
	controller := NewPipelineController(mockBuildService, mockSourceExtractor, nil)

	// Assemble the pipelines
	pipelines, err := controller.Assemble(ctx, build.Source{})

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, pipelines)
	require.Len(t, pipelines, 2)

	// Verify each pipeline has the correct BuiltSourceImage
	assert.Equal(t, "sha256:abc123", pipelines[0].Metadata.BuiltSourceImage)
	assert.Equal(t, "sha256:def456", pipelines[1].Metadata.BuiltSourceImage)

	// Verify each pipeline has the correct SourceLanguage
	assert.Equal(t, config.SourceLanguageJava, pipelines[0].Metadata.SourceLanguage)
	assert.Equal(t, config.SourceLanguageGo, pipelines[1].Metadata.SourceLanguage)
}

// TestPipelineController_Assemble_PopulatesPipelineName tests that Assemble populates PipelineName in PipelineMetadata
// Given: A MockBuildService that returns pipelines with names
// When: Assemble(ctx, source) is called
// Then: Each pipeline's Metadata.PipelineName is populated with the pipeline name
func TestPipelineController_Assemble_PopulatesPipelineName(t *testing.T) {
	ctx := context.Background()

	// Create mock build service that returns pipelines with names
	mockBuildService := &MockBuildService{
		BuildFunc: func(ctx context.Context, source build.Source) (*build.BuildResult, error) {
			// Create mock pipelines with names
			pipelines := []*pb.Pipeline{
				{Name: "my-pipeline"},
				{Name: "another-pipeline"},
			}
			sourceBuildResults := []build.SourceBuildResult{
				{ImageID: "sha256:abc123", YeetcdConfig: config.YeetcdConfig{Language: config.SourceLanguageJava}},
				{ImageID: "sha256:def456", YeetcdConfig: config.YeetcdConfig{Language: config.SourceLanguageJava}},
			}
			return &build.BuildResult{
				Pipelines:          pipelines,
				SourceBuildResults: sourceBuildResults,
			}, nil
		},
	}

	// Create pipeline controller
	controller := NewPipelineController(mockBuildService, mockSourceExtractor, nil)

	// Assemble the pipelines
	pipelines, err := controller.Assemble(ctx, build.Source{})

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, pipelines)
	require.Len(t, pipelines, 2)

	// Verify each pipeline has the correct PipelineName in metadata
	assert.Equal(t, "my-pipeline", pipelines[0].Metadata.PipelineName)
	assert.Equal(t, "another-pipeline", pipelines[1].Metadata.PipelineName)
}

// TestPipelineController_Execute_RecordsPipelineStartedEvent tests that Execute records PipelineStarted event
// Given: A Pipeline with final work, MockExecutionEngine, TestPipelineOutputHandler
// When: Execute(ctx, pipeline, handler) is called
// Then: PipelineStarted event is recorded with the pipeline
func TestPipelineController_Execute_RecordsPipelineStartedEvent(t *testing.T) {
	ctx := context.Background()

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create pipeline controller
	controller := NewPipelineController(nil, mockSourceExtractor, mockEngine)

	// Create test handler
	handler := NewTestPipelineOutputHandler()

	// Create a simple pipeline with containerised work
	pipeline := &Pipeline{
		Name:        "test-pipeline",
		WorkContext: make(WorkContext),
		FinalWork: []*Work{
			{
				ID:          "work-1",
				Description: "Test work",
				WorkContext: make(WorkContext),
				WorkDefinition: &ContainerisedWorkDefinition{
					Image: "alpine:latest",
					Cmd:   []string{"echo", "hello"},
				},
			},
		},
		Metadata: PipelineMetadata{
			PipelineName: "test-pipeline",
		},
	}

	// Execute the pipeline
	result, err := controller.Execute(ctx, pipeline, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify PipelineStarted event was recorded
	startedEvents := GetEventsOfType[PipelineStarted](handler)
	require.Len(t, startedEvents, 1)
	assert.Equal(t, "test-pipeline", startedEvents[0].Pipeline.Name)
}

// TestPipelineController_Execute_RecordsPipelineFinishedEvent tests that Execute records PipelineFinished event
// Given: A Pipeline with final work, MockExecutionEngine, TestPipelineOutputHandler
// When: Execute(ctx, pipeline, handler) is called
// Then: PipelineFinished event is recorded after all work completes
func TestPipelineController_Execute_RecordsPipelineFinishedEvent(t *testing.T) {
	ctx := context.Background()

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create pipeline controller
	controller := NewPipelineController(nil, mockSourceExtractor, mockEngine)

	// Create test handler
	handler := NewTestPipelineOutputHandler()

	// Create a simple pipeline
	pipeline := &Pipeline{
		Name:        "test-pipeline",
		WorkContext: make(WorkContext),
		FinalWork: []*Work{
			{
				ID:          "work-1",
				Description: "Test work",
				WorkContext: make(WorkContext),
				WorkDefinition: &ContainerisedWorkDefinition{
					Image: "alpine:latest",
					Cmd:   []string{"echo", "hello"},
				},
			},
		},
		Metadata: PipelineMetadata{
			PipelineName: "test-pipeline",
		},
	}

	// Execute the pipeline
	result, err := controller.Execute(ctx, pipeline, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify PipelineFinished event was recorded
	finishedEvents := GetEventsOfType[PipelineFinished](handler)
	require.Len(t, finishedEvents, 1)
	assert.Equal(t, types.PipelineSuccess, finishedEvents[0].PipelineStatus)
}

// TestPipelineController_Execute_ExecutesAllFinalWorkItems tests that Execute runs all final work items
// Given: A Pipeline with multiple final work items, MockExecutionEngine
// When: Execute(ctx, pipeline, handler) is called
// Then: All work items are executed and their results are tracked
func TestPipelineController_Execute_ExecutesAllFinalWorkItems(t *testing.T) {
	ctx := context.Background()

	// Track executed jobs
	executedJobs := make([]string, 0)

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			// Track which image was executed
			executedJobs = append(executedJobs, def.Image)
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create pipeline controller
	controller := NewPipelineController(nil, mockSourceExtractor, mockEngine)

	// Create test handler
	handler := NewTestPipelineOutputHandler()

	// Create a pipeline with multiple final work items
	pipeline := &Pipeline{
		Name:        "test-pipeline",
		WorkContext: make(WorkContext),
		FinalWork: []*Work{
			{
				ID:          "work-1",
				Description: "First work",
				WorkContext: make(WorkContext),
				WorkDefinition: &ContainerisedWorkDefinition{
					Image: "alpine:latest",
					Cmd:   []string{"echo", "first"},
				},
			},
			{
				ID:          "work-2",
				Description: "Second work",
				WorkContext: make(WorkContext),
				WorkDefinition: &ContainerisedWorkDefinition{
					Image: "busybox:latest",
					Cmd:   []string{"echo", "second"},
				},
			},
		},
		Metadata: PipelineMetadata{
			PipelineName: "test-pipeline",
		},
	}

	// Execute the pipeline
	result, err := controller.Execute(ctx, pipeline, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify both work items were executed
	assert.Len(t, executedJobs, 2)
	assert.Contains(t, executedJobs, "alpine:latest")
	assert.Contains(t, executedJobs, "busybox:latest")

	// Verify work results are tracked
	assert.Len(t, result.WorkResults, 2)
}

// TestPipelineController_Execute_ContinuesOnWorkFailure tests that Execute continues even if some work fails
// Given: A Pipeline with multiple final work items where one fails, MockExecutionEngine
// When: Execute(ctx, pipeline, handler) is called
// Then: All work items are attempted, failed work is recorded, pipeline continues
func TestPipelineController_Execute_ContinuesOnWorkFailure(t *testing.T) {
	ctx := context.Background()

	// Create mock execution engine that fails for specific image
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			if def.Image == "failing-image:latest" {
				return &engine.JobResult{ExitCode: 1}, nil
			}
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create pipeline controller
	controller := NewPipelineController(nil, mockSourceExtractor, mockEngine)

	// Create test handler
	handler := NewTestPipelineOutputHandler()

	// Create a pipeline with one failing and one succeeding work item
	pipeline := &Pipeline{
		Name:        "test-pipeline",
		WorkContext: make(WorkContext),
		FinalWork: []*Work{
			{
				ID:          "work-1",
				Description: "Failing work",
				WorkContext: make(WorkContext),
				WorkDefinition: &ContainerisedWorkDefinition{
					Image: "failing-image:latest",
					Cmd:   []string{"false"},
				},
			},
			{
				ID:          "work-2",
				Description: "Succeeding work",
				WorkContext: make(WorkContext),
				WorkDefinition: &ContainerisedWorkDefinition{
					Image: "alpine:latest",
					Cmd:   []string{"echo", "success"},
				},
			},
		},
		Metadata: PipelineMetadata{
			PipelineName: "test-pipeline",
		},
	}

	// Execute the pipeline
	result, err := controller.Execute(ctx, pipeline, handler)

	// Verify no error (pipeline continues even if work fails)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify both work results are tracked
	assert.Len(t, result.WorkResults, 2)

	// Verify WorkFinished events were recorded for both
	finishedEvents := GetEventsOfType[WorkFinished](handler)
	assert.Len(t, finishedEvents, 2)
}

// TestPipelineController_Execute_PassesWorkContextToWorkItems tests that Execute passes pipeline work context to work items
// Given: A Pipeline with work context, MockExecutionEngine
// When: Execute(ctx, pipeline, handler) is called
// Then: Work context is available during work execution
func TestPipelineController_Execute_PassesWorkContextToWorkItems(t *testing.T) {
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

	// Create pipeline controller
	controller := NewPipelineController(nil, mockSourceExtractor, mockEngine)

	// Create test handler
	handler := NewTestPipelineOutputHandler()

	// Create a pipeline with work context
	pipeline := &Pipeline{
		Name: "test-pipeline",
		WorkContext: WorkContext{
			"PIPELINE_KEY": "pipeline_value",
		},
		FinalWork: []*Work{
			{
				ID:          "work-1",
				Description: "Test work",
				WorkContext: WorkContext{
					"WORK_KEY": "work_value",
				},
				WorkDefinition: &ContainerisedWorkDefinition{
					Image: "alpine:latest",
					Cmd:   []string{"echo", "hello"},
				},
			},
		},
		Metadata: PipelineMetadata{
			PipelineName: "test-pipeline",
		},
	}

	// Execute the pipeline
	result, err := controller.Execute(ctx, pipeline, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify work context was passed
	assert.NotNil(t, receivedEnv)
	assert.Equal(t, "pipeline_value", receivedEnv["PIPELINE_KEY"])
	assert.Equal(t, "work_value", receivedEnv["WORK_KEY"])
}

// TestPipelineController_Execute_ReturnsPipelineResultWithWorkResults tests that Execute returns PipelineResult with all work results
// Given: A Pipeline with multiple work items, MockExecutionEngine
// When: Execute(ctx, pipeline, handler) is called
// Then: PipelineResult contains WorkResults map with all work IDs and their results
func TestPipelineController_Execute_ReturnsPipelineResultWithWorkResults(t *testing.T) {
	ctx := context.Background()

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create pipeline controller
	controller := NewPipelineController(nil, mockSourceExtractor, mockEngine)

	// Create test handler
	handler := NewTestPipelineOutputHandler()

	// Create a pipeline with multiple work items
	pipeline := &Pipeline{
		Name:        "test-pipeline",
		WorkContext: make(WorkContext),
		FinalWork: []*Work{
			{
				ID:          "work-1",
				Description: "First work",
				WorkContext: make(WorkContext),
				WorkDefinition: &ContainerisedWorkDefinition{
					Image: "alpine:latest",
					Cmd:   []string{"echo", "first"},
				},
			},
			{
				ID:          "work-2",
				Description: "Second work",
				WorkContext: make(WorkContext),
				WorkDefinition: &ContainerisedWorkDefinition{
					Image: "busybox:latest",
					Cmd:   []string{"echo", "second"},
				},
			},
		},
		Metadata: PipelineMetadata{
			PipelineName: "test-pipeline",
		},
	}

	// Execute the pipeline
	result, err := controller.Execute(ctx, pipeline, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify PipelineResult contains work results
	assert.NotNil(t, result.WorkResults)
	assert.Len(t, result.WorkResults, 2)

	// Verify each work has a result
	for _, work := range pipeline.FinalWork {
		workResult, exists := result.WorkResults[work.ID]
		assert.True(t, exists, "Work %s should have a result", work.ID)
		assert.NotNil(t, workResult)
	}
}

// TestPipelineController_Execute_EmptyPipeline tests that Execute handles empty pipeline
// Given: A Pipeline with no final work items
// When: Execute(ctx, pipeline, handler) is called
// Then: PipelineStarted and PipelineFinished events are recorded, empty result returned
func TestPipelineController_Execute_EmptyPipeline(t *testing.T) {
	ctx := context.Background()

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{}

	// Create pipeline controller
	controller := NewPipelineController(nil, mockSourceExtractor, mockEngine)

	// Create test handler
	handler := NewTestPipelineOutputHandler()

	// Create an empty pipeline
	pipeline := &Pipeline{
		Name:        "empty-pipeline",
		WorkContext: make(WorkContext),
		FinalWork:   []*Work{},
		Metadata: PipelineMetadata{
			PipelineName: "empty-pipeline",
		},
	}

	// Execute the pipeline
	result, err := controller.Execute(ctx, pipeline, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify events were recorded
	startedEvents := GetEventsOfType[PipelineStarted](handler)
	finishedEvents := GetEventsOfType[PipelineFinished](handler)
	require.Len(t, startedEvents, 1)
	require.Len(t, finishedEvents, 1)

	// Verify empty work results
	assert.Empty(t, result.WorkResults)
}

// TestPipelineController_Execute_EventOrdering tests that events are recorded in correct order
// Given: A Pipeline with work items, TestPipelineOutputHandler
// When: Execute(ctx, pipeline, handler) is called
// Then: Events are recorded in order: PipelineStarted, WorkStarted(s), WorkFinished(s), PipelineFinished
func TestPipelineController_Execute_EventOrdering(t *testing.T) {
	ctx := context.Background()

	// Create mock execution engine
	mockEngine := &MockExecutionEngine{
		RunJobFunc: func(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
			return &engine.JobResult{ExitCode: 0}, nil
		},
	}

	// Create pipeline controller
	controller := NewPipelineController(nil, mockSourceExtractor, mockEngine)

	// Create test handler
	handler := NewTestPipelineOutputHandler()

	// Create a pipeline with one work item
	pipeline := &Pipeline{
		Name:        "test-pipeline",
		WorkContext: make(WorkContext),
		FinalWork: []*Work{
			{
				ID:          "work-1",
				Description: "Test work",
				WorkContext: make(WorkContext),
				WorkDefinition: &ContainerisedWorkDefinition{
					Image: "alpine:latest",
					Cmd:   []string{"echo", "hello"},
				},
			},
		},
		Metadata: PipelineMetadata{
			PipelineName: "test-pipeline",
		},
	}

	// Execute the pipeline
	result, err := controller.Execute(ctx, pipeline, handler)

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Get all events
	events := handler.GetEvents()
	require.Len(t, events, 4) // PipelineStarted, WorkStarted, WorkFinished, PipelineFinished

	// Verify event order
	_, ok1 := events[0].(PipelineStarted)
	assert.True(t, ok1, "First event should be PipelineStarted")

	_, ok2 := events[1].(WorkStarted)
	assert.True(t, ok2, "Second event should be WorkStarted")

	_, ok3 := events[2].(WorkFinished)
	assert.True(t, ok3, "Third event should be WorkFinished")

	_, ok4 := events[3].(PipelineFinished)
	assert.True(t, ok4, "Fourth event should be PipelineFinished")
}
