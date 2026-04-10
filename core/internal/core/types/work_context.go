package types

// WorkContext is a map of key-value pairs for work context
type WorkContext map[string]string

// EmptyWorkContext creates an empty WorkContext
func EmptyWorkContext() WorkContext {
	return make(WorkContext)
}

// WorkContextFromMap creates a WorkContext from a map
func WorkContextFromMap(m map[string]string) WorkContext {
	return WorkContext(m)
}

// NewWorkContext creates a new WorkContext from a map (alias for WorkContextFromMap)
func NewWorkContext(m map[string]string) WorkContext {
	return WorkContext(m)
}

// MergeInto merges this context into another, with source overriding destination
func (wc WorkContext) MergeInto(dest WorkContext) WorkContext {
	result := make(WorkContext)

	// First copy all from destination
	for k, v := range dest {
		result[k] = v
	}

	// Then override with source values
	for k, v := range wc {
		result[k] = v
	}

	return result
}
