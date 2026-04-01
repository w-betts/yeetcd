package pipeline

import (
	"context"

	"github.com/yeetcd/yeetcd/pkg/engine"
)

// Execute implements WorkDefinition for CustomWorkDefinition
// Delegates to NativeWorkDefinition.Execute() with executionId from CustomWorkDefinition struct
func (c *CustomWorkDefinition) Execute(ctx context.Context, work Work, eng engine.ExecutionEngine, metadata PipelineMetadata, tracker *WorkResultTracker, handler PipelineOutputHandler) (*WorkResult, error) {
	native := &NativeWorkDefinition{
		ExecutionID: c.ExecutionID,
	}
	return native.Execute(ctx, work, eng, metadata, tracker, handler)
}
