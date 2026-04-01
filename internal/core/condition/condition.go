package condition

import (
	pipelinepb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
	"github.com/yeetcd/yeetcd/internal/core/pipeline"
)

// Condition is the interface for all condition types
type Condition interface {
	Evaluate(workContext pipeline.WorkContext, workResultTracker *pipeline.WorkResultTracker) (bool, error)
	ToProtobuf() (*pipelinepb.Condition, error)
}
