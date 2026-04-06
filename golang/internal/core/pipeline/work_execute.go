package pipeline

import (
	"context"
	"fmt"

	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// Execute runs the work with the given context and engine
// Logic: (1) recursively execute all previous work dependencies, (2) merge work context with containing context,
// (3) execute work definition, (4) handle dynamic work generation if applicable, (5) record WorkFinished event
func (w *Work) Execute(ctx context.Context, containingContext types.WorkContext, engine engine.ExecutionEngine, metadata PipelineMetadata, tracker *WorkResultTracker, handler PipelineOutputHandler) (*types.WorkResult, error) {
	return tracker.GetOrExecute(*w, func() (*types.WorkResult, error) {
		// Step 1: Recursively execute all previous work dependencies
		for _, prevWork := range w.PreviousWork {
			_, err := prevWork.Work.Execute(ctx, containingContext, engine, metadata, tracker, handler)
			if err != nil {
				return nil, fmt.Errorf("failed to execute previous work %s: %w", prevWork.Work.ID, err)
			}
		}

		// Step 2: Merge work context with containing context (work context overrides)
		mergedContext := w.WorkContext.MergeInto(containingContext)

		// Step 3: Execute work definition
		result, err := w.WorkDefinition.Execute(ctx, *w, mergedContext, engine, metadata, tracker, handler)
		if err != nil {
			return nil, fmt.Errorf("work definition execution failed: %w", err)
		}

		// Step 4: Handle dynamic work generation if applicable
		if dynamicDef, ok := w.WorkDefinition.(*DynamicWorkGeneratingWorkDefinition); ok {
			if result.WorkStatus == types.SUCCESS {
				// Parse stdout as protobuf Work message and execute it
				generatedWork, err := dynamicDef.ParseAndCreateWork(result.JobStreams)
				if err != nil {
					return nil, fmt.Errorf("failed to parse generated work: %w", err)
				}

				// Execute the generated work recursively
				generatedResult, err := generatedWork.Execute(ctx, mergedContext, engine, metadata, tracker, handler)
				if err != nil {
					return nil, fmt.Errorf("failed to execute generated work: %w", err)
				}

				// Return the generated work's result
				result = generatedResult
			}
		}

		// Step 5: Record WorkFinished event
		handler.RecordEvent(WorkFinished{
			Work:       *w,
			WorkStatus: result.WorkStatus,
		})

		return result, nil
	})
}
