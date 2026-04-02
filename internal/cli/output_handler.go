package cli

import (
	"io"
	"log/slog"
	"os"

	"github.com/yeetcd/yeetcd/internal/core/pipeline"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// OutputHandler is a CLI-specific PipelineOutputHandler implementation
// that logs events using slog with appropriate log levels
type OutputHandler struct {
	stdout io.Writer
	stderr io.Writer
}

// NewOutputHandler creates a new CLI output handler
func NewOutputHandler() *OutputHandler {
	return &OutputHandler{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// RecordEvent logs pipeline events using slog with appropriate log levels
func (h *OutputHandler) RecordEvent(event interface{}) {
	switch e := event.(type) {
	case pipeline.PipelineStarted:
		slog.Info("pipeline started", "name", e.Pipeline.Name)
	case pipeline.WorkStarted:
		slog.Info("work started", "description", e.Work.Description)
	case pipeline.WorkFinished:
		slog.Info("work finished",
			"description", e.Work.Description,
			"status", e.WorkStatus)
	case pipeline.PipelineFinished:
		slog.Info("pipeline finished", "status", e.PipelineStatus)
	default:
		slog.Debug("unknown event", "type", event)
	}
}

// NewJobStreams returns JobStreams with os.Stdout/os.Stderr for live output
func (h *OutputHandler) NewJobStreams() interface{} {
	return engine.NewJobStreams(h.stdout, h.stderr)
}
