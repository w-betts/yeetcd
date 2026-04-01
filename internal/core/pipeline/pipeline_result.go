package pipeline

// PipelineResult represents the result of pipeline execution
type PipelineResult struct {
	WorkResults map[string]*WorkResult
}

// PipelineStatus returns the overall pipeline status
func (pr *PipelineResult) PipelineStatus() PipelineStatus {
	// Stub - should check all work results
	return PipelineSuccess
}
