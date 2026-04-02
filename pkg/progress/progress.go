package progress

// ProgressReporter interface for progress reporting
type ProgressReporter interface {
	PipelineStarted(pipeline interface{})
	WorkStarted(work interface{}, streams interface{})
	WorkFinished(work interface{}, status interface{})
	PipelineFinished(status interface{})
}
