package pipeline

import (
	"sync"

	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// WorkResultTracker tracks work execution results
type WorkResultTracker struct {
	results map[string]*types.WorkResult
	order   []string // Track the order of work IDs for GetLastResult
	mu      sync.RWMutex
}

// NewWorkResultTracker creates a new WorkResultTracker
func NewWorkResultTracker() *WorkResultTracker {
	return &WorkResultTracker{
		results: make(map[string]*types.WorkResult),
		order:   make([]string, 0),
	}
}

// GetOrExecute returns existing result or executes the work
func (t *WorkResultTracker) GetOrExecute(work Work, executeFunc func() (*types.WorkResult, error)) (*types.WorkResult, error) {
	// Check if result already exists (read lock)
	t.mu.RLock()
	if result, exists := t.results[work.ID]; exists {
		t.mu.RUnlock()
		return result, nil
	}
	t.mu.RUnlock()

	// Execute the work
	result, err := executeFunc()
	if err != nil {
		return nil, err
	}

	// Store the result (write lock)
	t.mu.Lock()
	t.results[work.ID] = result
	t.order = append(t.order, work.ID)
	t.mu.Unlock()

	return result, nil
}

// GetWorkResultMap returns the map of work results
func (t *WorkResultTracker) GetWorkResultMap() map[string]*types.WorkResult {
	return t.results
}

// OutputDirectoriesMountInput returns mount input for output directories
func (t *WorkResultTracker) OutputDirectoriesMountInput(work Work) interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result, exists := t.results[work.ID]
	if !exists || result == nil {
		return nil
	}

	if result.OutputDirectoriesParent == "" {
		return nil
	}

	return engine.OnDiskMountInput{Dir: result.OutputDirectoriesParent}
}

// StdOut returns stdout from a previous work result
func (t *WorkResultTracker) StdOut(work Work) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result, exists := t.results[work.ID]
	if !exists || result == nil || result.JobStreams == nil {
		return ""
	}

	if streams, ok := result.JobStreams.(*engine.JobStreams); ok {
		return string(streams.GetStdOut())
	}

	return ""
}

// GetLastResult returns the most recent work result for a given work ID
// If workID is empty, returns the last executed work result
func (t *WorkResultTracker) GetLastResult(workID string) *types.WorkResult {
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

// RecordResult records a work result
func (t *WorkResultTracker) RecordResult(workID string, result *types.WorkResult) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.results[workID] = result
	t.order = append(t.order, workID)
}
