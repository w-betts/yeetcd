package events

import (
	"github.com/yeetcd/yeetcd/internal/core/pipeline"
)

// WorkFinished event
type WorkFinished struct {
	Work       pipeline.Work
	WorkStatus pipeline.WorkStatus
}

func (e WorkFinished) IsPipelineEvent() {}
