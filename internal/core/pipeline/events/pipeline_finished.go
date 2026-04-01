package events

import "github.com/yeetcd/yeetcd/internal/core/pipeline"

// PipelineFinished event
type PipelineFinished struct {
	PipelineStatus pipeline.PipelineStatus
}

func (e PipelineFinished) IsPipelineEvent() {}
