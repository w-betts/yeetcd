package pipeline

import (
	"context"
	"fmt"

	"github.com/yeetcd/yeetcd/pkg/build"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// PipelineController orchestrates pipeline assembly and execution
type PipelineController struct {
	buildService     build.BuildService
	sourceExtractor  build.SourceExtractor
	executionEngine  engine.ExecutionEngine
}

// NewPipelineController creates a new PipelineController
func NewPipelineController(buildService build.BuildService, sourceExtractor build.SourceExtractor, engine engine.ExecutionEngine) *PipelineController {
	return &PipelineController{
		buildService:     buildService,
		sourceExtractor:  sourceExtractor,
		executionEngine:  engine,
	}
}

// Assemble extracts source, builds it, builds source image, runs container to generate protobuf definitions, parses into Pipeline structs
func (pc *PipelineController) Assemble(ctx context.Context, source build.Source) ([]*Pipeline, error) {
	return nil, fmt.Errorf("not implemented")
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
