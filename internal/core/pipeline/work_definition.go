package pipeline

import (
	"context"

	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// WorkDefinition is the interface for all work definition types
type WorkDefinition interface {
	Execute(ctx context.Context, work Work, eng engine.ExecutionEngine, metadata PipelineMetadata, tracker *WorkResultTracker, handler PipelineOutputHandler) (*types.WorkResult, error)
}

// ContainerisedWorkDefinition runs a command in an existing container image
type ContainerisedWorkDefinition struct {
	Image      string
	Cmd        []string
	WorkingDir string
}

// CustomWorkDefinition executes user-defined code
type CustomWorkDefinition struct {
	ExecutionID string
}

// CompoundWorkDefinition groups multiple work items
type CompoundWorkDefinition struct {
	FinalWork []Work
}

// DynamicWorkGeneratingWorkDefinition generates work at runtime
type DynamicWorkGeneratingWorkDefinition struct {
	ExecutionID string
}
