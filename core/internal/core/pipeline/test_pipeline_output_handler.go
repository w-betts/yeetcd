package pipeline

import (
	"sync"

	"github.com/yeetcd/yeetcd/pkg/engine"
)

// TestPipelineOutputHandler is a test helper that implements PipelineOutputHandler
// and records all events for verification
type TestPipelineOutputHandler struct {
	events     []interface{}
	mu         sync.RWMutex
	jobStreams []*engine.JobStreams
}

// NewTestPipelineOutputHandler creates a new TestPipelineOutputHandler
func NewTestPipelineOutputHandler() *TestPipelineOutputHandler {
	return &TestPipelineOutputHandler{
		events:     make([]interface{}, 0),
		jobStreams: make([]*engine.JobStreams, 0),
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
	streams := engine.NewJobStreams(nil, nil)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.jobStreams = append(t.jobStreams, streams)
	return streams
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
	t.jobStreams = make([]*engine.JobStreams, 0)
}

// GetJobStreams returns all captured JobStreams
func (t *TestPipelineOutputHandler) GetJobStreams() []*engine.JobStreams {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]*engine.JobStreams, len(t.jobStreams))
	copy(result, t.jobStreams)
	return result
}

// GetStdOutByWorkDescription returns the stdout for a specific work by its description
// This is useful for verifying custom work execution output
func (t *TestPipelineOutputHandler) GetStdOutByWorkDescription(description string) []byte {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Find the WorkStarted event for the given work description
	for _, event := range t.events {
		if ws, ok := event.(WorkStarted); ok {
			if ws.Work.Description == description {
				if js, ok := ws.JobStreams.(*engine.JobStreams); ok {
					return js.GetStdOut()
				}
			}
		}
	}

	return nil
}
