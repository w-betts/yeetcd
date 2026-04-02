package pipeline

// PipelineOutputHandler handles pipeline output events
type PipelineOutputHandler interface {
	RecordEvent(event interface{})
	NewJobStreams() interface{} // returns JobStreams
}
