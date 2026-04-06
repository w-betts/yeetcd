package cli

import (
	"fmt"
	"os"

	"github.com/yeetcd/yeetcd/internal/core/pipeline"
	"github.com/yeetcd/yeetcd/pkg/progress"
)

// ProgressReporter is a CLI-specific progress reporter that provides
// human-readable progress output
type ProgressReporter struct {
	useColor bool
}

// NewProgressReporter creates a new CLI progress reporter
func NewProgressReporter() *ProgressReporter {
	// Check if terminal supports color
	useColor := os.Getenv("NO_COLOR") == "" && isTerminal()
	return &ProgressReporter{useColor: useColor}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	// Simple check - in production, use a proper terminal detection library
	return true
}

// colorize adds ANSI color codes if color is enabled
func (p *ProgressReporter) colorize(text string, color string) string {
	if !p.useColor {
		return text
	}
	// ANSI color codes
	colors := map[string]string{
		"green":  "\033[32m",
		"red":    "\033[31m",
		"yellow": "\033[33m",
		"blue":   "\033[34m",
		"reset":  "\033[0m",
	}
	if code, ok := colors[color]; ok {
		return code + text + colors["reset"]
	}
	return text
}

// PipelineStarted prints that a pipeline has started
func (p *ProgressReporter) PipelineStarted(pl interface{}) {
	if pl, ok := pl.(*pipeline.Pipeline); ok {
		fmt.Printf("Starting pipeline: %s\n", p.colorize(pl.Name, "blue"))
	} else {
		fmt.Println("Starting pipeline...")
	}
}

// WorkStarted prints that a work item is running
func (p *ProgressReporter) WorkStarted(work interface{}, streams interface{}) {
	if w, ok := work.(*pipeline.Work); ok {
		fmt.Printf("  Running: %s\n", w.Description)
	} else {
		fmt.Println("  Running work...")
	}
}

// WorkFinished prints that a work item has completed
func (p *ProgressReporter) WorkFinished(work interface{}, status interface{}) {
	if w, ok := work.(*pipeline.Work); ok {
		statusStr := fmt.Sprintf("%v", status)
		color := "green"
		if statusStr == "FAILURE" {
			color = "red"
		}
		fmt.Printf("  Done: %s (%s)\n", w.Description, p.colorize(statusStr, color))
	} else {
		fmt.Printf("  Done: %v\n", status)
	}
}

// PipelineFinished prints that the pipeline has completed
func (p *ProgressReporter) PipelineFinished(status interface{}) {
	statusStr := fmt.Sprintf("%v", status)
	color := "green"
	if statusStr == "FAILURE" {
		color = "red"
	}
	fmt.Printf("Pipeline completed: %s\n", p.colorize(statusStr, color))
}

// Ensure ProgressReporter implements the progress.ProgressReporter interface
var _ progress.ProgressReporter = (*ProgressReporter)(nil)
