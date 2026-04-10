package types

import (
	"sync"
)

// SimpleWorkResultTracker is a simple implementation of WorkResultTracker
// for use in tests and other places where a full tracker is not needed
type SimpleWorkResultTracker struct {
	results map[string]*WorkResult
	order   []string // Track the order of work IDs for GetLastResult
	mu      sync.RWMutex
}

// NewSimpleWorkResultTracker creates a new SimpleWorkResultTracker
func NewSimpleWorkResultTracker() *SimpleWorkResultTracker {
	return &SimpleWorkResultTracker{
		results: make(map[string]*WorkResult),
		order:   make([]string, 0),
	}
}

// GetLastResult returns the most recent work result for a given work ID
// If workID is empty, returns the last executed work result
func (t *SimpleWorkResultTracker) GetLastResult(workID string) *WorkResult {
	t.mu.RLock()
	defer t.mu.RUnlock()

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

// GetWorkResultMap returns the map of work results
func (t *SimpleWorkResultTracker) GetWorkResultMap() map[string]*WorkResult {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]*WorkResult)
	for k, v := range t.results {
		result[k] = v
	}
	return result
}

// RecordResult records a work result
func (t *SimpleWorkResultTracker) RecordResult(workID string, result *WorkResult) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.results[workID] = result
	t.order = append(t.order, workID)
}
