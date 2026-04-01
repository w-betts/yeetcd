package events

import (
	"github.com/yeetcd/yeetcd/internal/core/pipeline"
)

// WorkStarted event
type WorkStarted struct {
	Work       pipeline.Work
	JobStreams interface{}
}

func (e WorkStarted) IsPipelineEvent() {}
