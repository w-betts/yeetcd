package pipeline

import "github.com/yeetcd/yeetcd/internal/core/types"

// PipelineEvent is a marker interface for pipeline events
type PipelineEvent interface {
	IsPipelineEvent()
}

// WorkStarted event
type WorkStarted struct {
	Work       Work
	JobStreams interface{}
}

func (e WorkStarted) IsPipelineEvent() {}

// WorkFinished event
type WorkFinished struct {
	Work       Work
	WorkStatus types.WorkStatus
}

func (e WorkFinished) IsPipelineEvent() {}

// PipelineStarted event
type PipelineStarted struct {
	Pipeline Pipeline
}

func (e PipelineStarted) IsPipelineEvent() {}

// PipelineFinished event
type PipelineFinished struct {
	PipelineStatus types.PipelineStatus
}

func (e PipelineFinished) IsPipelineEvent() {}
