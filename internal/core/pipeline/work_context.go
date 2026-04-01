package pipeline

import "github.com/yeetcd/yeetcd/internal/core/types"

// WorkContext is a map of key-value pairs for work context
// This is an alias for types.WorkContext for backward compatibility
type WorkContext = types.WorkContext

// EmptyWorkContext creates an empty WorkContext
var EmptyWorkContext = types.EmptyWorkContext

// WorkContextFromMap creates a WorkContext from a map
var WorkContextFromMap = types.WorkContextFromMap

// NewWorkContext creates a new WorkContext from a map
var NewWorkContext = types.NewWorkContext
