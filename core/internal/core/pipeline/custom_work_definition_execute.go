package pipeline

import (
	"context"

	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// Execute implements WorkDefinition for CustomWorkDefinition
// Delegates to NativeWorkDefinition.Execute() with executionId from CustomWorkDefinition struct
func (c *CustomWorkDefinition) Execute(ctx context.Context, work Work, mergedContext types.WorkContext, eng engine.ExecutionEngine, metadata PipelineMetadata, tracker *WorkResultTracker, handler PipelineOutputHandler) (*types.WorkResult, error) {
	native := &NativeWorkDefinition{
		ExecutionID: c.ExecutionID,
	}
	return native.Execute(ctx, work, mergedContext, eng, metadata, tracker, handler)
}
