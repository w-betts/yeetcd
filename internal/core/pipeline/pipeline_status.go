package pipeline

import "github.com/yeetcd/yeetcd/internal/core/types"

// PipelineStatus represents the status of a pipeline execution
// This is an alias for types.PipelineStatus for backward compatibility
type PipelineStatus = types.PipelineStatus

const (
	PipelineSuccess = types.PipelineSuccess
	PipelineFailure = types.PipelineFailure
)
