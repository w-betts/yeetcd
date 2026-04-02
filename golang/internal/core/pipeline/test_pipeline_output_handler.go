package pipeline

import (
	"sync"

	"github.com/yeetcd/yeetcd/pkg/engine"
)

// TestPipelineOutputHandler is a test helper that implements PipelineOutputHandler
// and records all events for verification
type TestPipelineOutputHandler struct {
	events []interface{}
	mu     sync.RWMutex
}

// NewTestPipelineOutputHandler creates a new TestPipelineOutputHandler
func NewTestPipelineOutputHandler() *TestPipelineOutputHandler {
	return &TestPipelineOutputHandler{
		events: make([]interface{}, 0),
	}
}

// RecordEvent records a pipeline event
func (t *TestPipelineOutputHandler) RecordEvent(event interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
}

// NewJobStreams creates new JobStreams for job output capture
func (t *TestPipelineOutputHandler) NewJobStreams() interface{} {
	return engine.NewJobStreams(nil, nil)
}

// GetEvents returns all recorded events
func (t *TestPipelineOutputHandler) GetEvents() []interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]interface{}, len(t.events))
	copy(result, t.events)
	return result
}

// GetEventsOfType returns events of a specific type
func GetEventsOfType[T any](handler *TestPipelineOutputHandler) []T {
	var result []T
	for _, event := range handler.GetEvents() {
		if typedEvent, ok := event.(T); ok {
			result = append(result, typedEvent)
		}
	}
	return result
}

// GetEventCount returns the total number of recorded events
func (t *TestPipelineOutputHandler) GetEventCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.events)
}

// Clear clears all recorded events
func (t *TestPipelineOutputHandler) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = make([]interface{}, 0)
}
