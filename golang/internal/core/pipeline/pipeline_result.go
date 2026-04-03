package pipeline

// PipelineResult represents the result of pipeline execution
type PipelineResult struct {
	WorkResults map[string]*WorkResult
}

// PipelineStatus returns the overall pipeline status
// Returns SUCCESS only if ALL work items succeeded (or were skipped).
// Returns FAILURE if any work item failed.
func (pr *PipelineResult) PipelineStatus() PipelineStatus {
	// If no work results, consider it a success (nothing to fail)
	if len(pr.WorkResults) == 0 {
		return PipelineSuccess
	}

	// Check all work results - any failure means pipeline failed
	for _, result := range pr.WorkResults {
		if result.WorkStatus == WorkStatusFailed {
			return PipelineFailure
		}
	}

	// All work succeeded (or were skipped)
	return PipelineSuccess
}
