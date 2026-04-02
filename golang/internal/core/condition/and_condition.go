package condition

import (
	"errors"
	"fmt"

	pipelinepb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
	"github.com/yeetcd/yeetcd/internal/core/types"
)

// AndCondition combines two conditions with AND
type AndCondition struct {
	Left  Condition
	Right Condition
}

// NewAndCondition creates a new AndCondition
func NewAndCondition(left, right Condition) *AndCondition {
	return &AndCondition{
		Left:  left,
		Right: right,
	}
}

// Evaluate returns left && right
func (a *AndCondition) Evaluate(workContext types.WorkContext, workResultTracker types.WorkResultTracker) (bool, error) {
	if a.Left == nil || a.Right == nil {
		return false, errors.New("and condition has nil sub-conditions")
	}

	leftResult, err := a.Left.Evaluate(workContext, workResultTracker)
	if err != nil {
		return false, fmt.Errorf("left condition evaluation failed: %w", err)
	}

	if !leftResult {
		return false, nil
	}

	rightResult, err := a.Right.Evaluate(workContext, workResultTracker)
	if err != nil {
		return false, fmt.Errorf("right condition evaluation failed: %w", err)
	}

	return rightResult, nil
}

// ToProtobuf serializes the condition to protobuf
func (a *AndCondition) ToProtobuf() (*pipelinepb.Condition, error) {
	if a.Left == nil || a.Right == nil {
		return nil, errors.New("and condition has nil sub-conditions")
	}

	leftProto, err := a.Left.ToProtobuf()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize left condition: %w", err)
	}

	rightProto, err := a.Right.ToProtobuf()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize right condition: %w", err)
	}

	return &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_AndCondition{
			AndCondition: &pipelinepb.AndCondition{
				Left:  leftProto,
				Right: rightProto,
			},
		},
	}, nil
}

// AndConditionFromProtobuf converts protobuf AndCondition to Go struct
func AndConditionFromProtobuf(proto *pipelinepb.AndCondition) (*AndCondition, error) {
	if proto == nil {
		return nil, errors.New("proto AndCondition is nil")
	}

	if proto.Left == nil || proto.Right == nil {
		return nil, errors.New("proto AndCondition has nil sub-conditions")
	}

	left, err := FromProtobuf(proto.Left)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize left condition: %w", err)
	}

	right, err := FromProtobuf(proto.Right)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize right condition: %w", err)
	}

	return &AndCondition{
		Left:  left,
		Right: right,
	}, nil
}
