package pipeline

// WorkStatus represents the status of work execution
type WorkStatus int

const (
	WorkStatusPending WorkStatus = iota
	WorkStatusRunning
	WorkStatusSucceeded
	WorkStatusFailed
	WorkStatusSkipped
)

// Deprecated constants for backward compatibility
const (
	SUCCESS = WorkStatusSucceeded
	SKIPPED = WorkStatusSkipped
	FAILURE = WorkStatusFailed
)
