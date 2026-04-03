package pipeline

import (
	"context"
	"fmt"

	"github.com/yeetcd/yeetcd/pkg/build"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// PipelineController orchestrates pipeline assembly and execution
type PipelineController struct {
	buildService    build.BuildService
	sourceExtractor *build.SourceExtractor
	executionEngine engine.ExecutionEngine
}

// NewPipelineController creates a new PipelineController
func NewPipelineController(buildService build.BuildService, sourceExtractor *build.SourceExtractor, engine engine.ExecutionEngine) *PipelineController {
	return &PipelineController{
		buildService:    buildService,
		sourceExtractor: sourceExtractor,
		executionEngine: engine,
	}
}

// Assemble extracts source, builds it, builds source image, runs container to generate protobuf definitions, parses into Pipeline structs
func (pc *PipelineController) Assemble(ctx context.Context, source build.Source) ([]*Pipeline, error) {
	// Build the source using the build service
	buildResult, err := pc.buildService.Build(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to build source: %w", err)
	}

	// Parse protobuf Pipeline messages into Go Pipeline structs
	// and populate PipelineMetadata with BuiltSourceImage
	pipelines := make([]*Pipeline, 0, len(buildResult.Pipelines))
	for i, pbPipeline := range buildResult.Pipelines {
		// Convert protobuf to Go struct
		pipeline, err := FromProtobuf(pbPipeline)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pipeline: %w", err)
		}

		// Populate PipelineMetadata with pipeline name
		pipeline.Metadata.PipelineName = pipeline.Name

		// Populate PipelineMetadata with BuiltSourceImage and SourceLanguage from source build results
		// The image ID for this pipeline should be in the corresponding SourceBuildResult
		if i < len(buildResult.SourceBuildResults) {
			sourceBuildResult := buildResult.SourceBuildResults[i]
			pipeline.Metadata.BuiltSourceImage = sourceBuildResult.ImageID
			pipeline.Metadata.SourceLanguage = sourceBuildResult.YeetcdConfig.Language
		}

		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}

// Execute creates WorkResultTracker, records PipelineStarted event, executes all final work items,
// collects results, records PipelineFinished event, returns PipelineResult
func (pc *PipelineController) Execute(ctx context.Context, pipeline *Pipeline, handler PipelineOutputHandler) (*PipelineResult, error) {
	// Create WorkResultTracker
	tracker := NewWorkResultTracker()

	// Record PipelineStarted event
	handler.RecordEvent(PipelineStarted{
		Pipeline: *pipeline,
	})

	// Execute all final work items
	for _, work := range pipeline.FinalWork {
		_, err := work.Execute(ctx, pipeline.WorkContext, pc.executionEngine, pipeline.Metadata, tracker, handler)
		if err != nil {
			// Continue execution even if one work fails
			// The work result will be recorded in the tracker
			_ = err
		}
	}

	// Collect results
	workResults := tracker.GetWorkResultMap()

	// Create PipelineResult
	result := &PipelineResult{
		WorkResults: workResults,
	}

	// Record PipelineFinished event
	handler.RecordEvent(PipelineFinished{
		PipelineStatus: result.PipelineStatus(),
	})

	return result, nil
}
