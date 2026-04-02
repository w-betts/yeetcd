package condition

import (
	"errors"
	"fmt"

	pipelinepb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
	"github.com/yeetcd/yeetcd/internal/core/types"
)

// NotCondition negates a condition
type NotCondition struct {
	Condition Condition
}

// NewNotCondition creates a new NotCondition
func NewNotCondition(condition Condition) *NotCondition {
	return &NotCondition{
		Condition: condition,
	}
}

// Evaluate returns !condition
func (n *NotCondition) Evaluate(workContext types.WorkContext, workResultTracker types.WorkResultTracker) (bool, error) {
	if n.Condition == nil {
		return false, errors.New("not condition has nil wrapped condition")
	}

	result, err := n.Condition.Evaluate(workContext, workResultTracker)
	if err != nil {
		return false, fmt.Errorf("wrapped condition evaluation failed: %w", err)
	}

	return !result, nil
}

// ToProtobuf serializes the condition to protobuf
func (n *NotCondition) ToProtobuf() (*pipelinepb.Condition, error) {
	if n.Condition == nil {
		return nil, errors.New("not condition has nil wrapped condition")
	}

	wrappedProto, err := n.Condition.ToProtobuf()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize wrapped condition: %w", err)
	}

	return &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_NotCondition{
			NotCondition: &pipelinepb.NotCondition{
				Condition: wrappedProto,
			},
		},
	}, nil
}

// NotConditionFromProtobuf converts protobuf NotCondition to Go struct
func NotConditionFromProtobuf(proto *pipelinepb.NotCondition) (*NotCondition, error) {
	if proto == nil {
		return nil, errors.New("proto NotCondition is nil")
	}

	if proto.Condition == nil {
		return nil, errors.New("proto NotCondition has nil wrapped condition")
	}

	wrapped, err := FromProtobuf(proto.Condition)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize wrapped condition: %w", err)
	}

	return &NotCondition{
		Condition: wrapped,
	}, nil
}
