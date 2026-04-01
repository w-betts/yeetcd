package pipeline

// PipelineStatus represents the status of a pipeline execution
type PipelineStatus int

const (
	PipelineSuccess PipelineStatus = iota
	PipelineFailure
)
