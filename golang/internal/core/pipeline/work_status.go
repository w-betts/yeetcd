package pipeline

import "github.com/yeetcd/yeetcd/internal/core/types"

// WorkStatus represents the status of work execution
// This is an alias for types.WorkStatus for backward compatibility
type WorkStatus = types.WorkStatus

// Work status constants - aliases for types package constants
const (
	WorkStatusPending   = types.WorkStatusPending
	WorkStatusRunning    = types.WorkStatusRunning
	WorkStatusSucceeded = types.WorkStatusSucceeded
	WorkStatusFailed     = types.WorkStatusFailed
	WorkStatusSkipped    = types.WorkStatusSkipped
)

// Deprecated constants for backward compatibility
const (
	SUCCESS = types.SUCCESS
	SKIPPED = types.SKIPPED
	FAILURE = types.FAILURE
)
