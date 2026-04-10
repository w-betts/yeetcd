package types

// ConditionEvaluator is the interface for evaluating conditions
// This interface is separated from the protobuf serialization to avoid import cycles
type ConditionEvaluator interface {
	// Evaluate returns true if the condition is met
	Evaluate(workContext WorkContext, workResultTracker WorkResultTracker) (bool, error)
}

// WorkResultTracker is the interface for tracking work results
// This interface is used by conditions to check previous work results
type WorkResultTracker interface {
	// GetLastResult returns the most recent work result for a given work ID
	// If workID is empty, returns the last executed work result
	GetLastResult(workID string) *WorkResult

	// GetWorkResultMap returns the map of work results
	GetWorkResultMap() map[string]*WorkResult

	// RecordResult records a work result
	RecordResult(workID string, result *WorkResult)
}
