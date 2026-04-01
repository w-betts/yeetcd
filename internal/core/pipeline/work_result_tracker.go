package pipeline

import (
	"errors"
)

// WorkResultTracker tracks work execution results
type WorkResultTracker struct {
	results map[string]*WorkResult
	order   []string // Track the order of work IDs for GetLastResult
}

// NewWorkResultTracker creates a new WorkResultTracker
func NewWorkResultTracker() *WorkResultTracker {
	return &WorkResultTracker{
		results: make(map[string]*WorkResult),
		order:   make([]string, 0),
	}
}

// GetOrExecute returns existing result or executes the work
func (t *WorkResultTracker) GetOrExecute(work Work, executeFunc func() (*WorkResult, error)) (*WorkResult, error) {
	return nil, errors.New("not implemented")
}

// GetWorkResultMap returns the map of work results
func (t *WorkResultTracker) GetWorkResultMap() map[string]*WorkResult {
	return t.results
}

// OutputDirectoriesMountInput returns mount input for output directories
func (t *WorkResultTracker) OutputDirectoriesMountInput(work Work) map[string]interface{} {
	return nil
}

// StdOut returns stdout from a previous work result
func (t *WorkResultTracker) StdOut(work Work) string {
	return ""
}

// GetLastResult returns the most recent work result for a given work ID
// If workID is empty, returns the last executed work result
func (t *WorkResultTracker) GetLastResult(workID string) *WorkResult {
	if workID != "" {
		if result, exists := t.results[workID]; exists {
			return result
		}
		return nil
	}
	
	// If no workID specified, return the last result in order
	if len(t.order) == 0 {
		return nil
	}
	
	lastWorkID := t.order[len(t.order)-1]
	return t.results[lastWorkID]
}

// RecordResult records a work result
func (t *WorkResultTracker) RecordResult(workID string, result *WorkResult) {
	t.results[workID] = result
	t.order = append(t.order, workID)
}
