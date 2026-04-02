package pipeline

import (
	"context"
	"fmt"

	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// Execute implements WorkDefinition for CompoundWorkDefinition
// Records WorkStarted event with empty JobStreams, executes all final work items recursively,
// determines compound result (SUCCESS if all succeeded, FAILURE otherwise),
// returns WorkResult with compound status
func (c *CompoundWorkDefinition) Execute(ctx context.Context, work Work, mergedContext types.WorkContext, eng engine.ExecutionEngine, metadata PipelineMetadata, tracker *WorkResultTracker, handler PipelineOutputHandler) (*types.WorkResult, error) {
	// Record WorkStarted event with empty streams
	handler.RecordEvent(WorkStarted{
		Work:       work,
		JobStreams: nil,
	})

	// Execute all final work items recursively
	allSucceeded := true
	for _, finalWork := range c.FinalWork {
		result, err := finalWork.Execute(ctx, mergedContext, eng, metadata, tracker, handler)
		if err != nil {
			return nil, fmt.Errorf("failed to execute final work %s: %w", finalWork.ID, err)
		}

		// Track if any work failed
		if result.WorkStatus != types.SUCCESS && result.WorkStatus != types.SKIPPED {
			allSucceeded = false
		}
	}

	// Determine compound result
	var workStatus types.WorkStatus
	if allSucceeded {
		workStatus = types.SUCCESS
	} else {
		workStatus = types.FAILURE
	}

	return &types.WorkResult{
		WorkStatus: workStatus,
	}, nil
}
