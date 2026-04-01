package events

import "github.com/yeetcd/yeetcd/internal/core/pipeline"

// PipelineStarted event
type PipelineStarted struct {
	Pipeline pipeline.Pipeline
}

func (e PipelineStarted) IsPipelineEvent() {}
