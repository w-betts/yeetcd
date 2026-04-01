package condition

import (
	pipelinepb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
	"github.com/yeetcd/yeetcd/internal/core/types"
)

// Condition is the interface for all condition types
// It extends types.ConditionEvaluator with protobuf serialization
type Condition interface {
	types.ConditionEvaluator
	ToProtobuf() (*pipelinepb.Condition, error)
}
