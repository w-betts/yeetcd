package condition

import (
	"errors"
	"fmt"

	pipelinepb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
	"github.com/yeetcd/yeetcd/internal/core/pipeline"
)

// OrCondition combines two conditions with OR
type OrCondition struct {
	Left  Condition
	Right Condition
}

// NewOrCondition creates a new OrCondition
func NewOrCondition(left, right Condition) *OrCondition {
	return &OrCondition{
		Left:  left,
		Right: right,
	}
}

// Evaluate returns left || right
func (o *OrCondition) Evaluate(workContext pipeline.WorkContext, workResultTracker *pipeline.WorkResultTracker) (bool, error) {
	if o.Left == nil || o.Right == nil {
		return false, errors.New("or condition has nil sub-conditions")
	}
	
	leftResult, err := o.Left.Evaluate(workContext, workResultTracker)
	if err != nil {
		return false, fmt.Errorf("left condition evaluation failed: %w", err)
	}
	
	if leftResult {
		return true, nil
	}
	
	rightResult, err := o.Right.Evaluate(workContext, workResultTracker)
	if err != nil {
		return false, fmt.Errorf("right condition evaluation failed: %w", err)
	}
	
	return rightResult, nil
}

// ToProtobuf serializes the condition to protobuf
func (o *OrCondition) ToProtobuf() (*pipelinepb.Condition, error) {
	if o.Left == nil || o.Right == nil {
		return nil, errors.New("or condition has nil sub-conditions")
	}
	
	leftProto, err := o.Left.ToProtobuf()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize left condition: %w", err)
	}
	
	rightProto, err := o.Right.ToProtobuf()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize right condition: %w", err)
	}
	
	return &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_OrCondition{
			OrCondition: &pipelinepb.OrCondition{
				Left:  leftProto,
				Right: rightProto,
			},
		},
	}, nil
}

// FromProtobuf converts protobuf OrCondition to Go struct
func OrConditionFromProtobuf(proto *pipelinepb.OrCondition) (*OrCondition, error) {
	if proto == nil {
		return nil, errors.New("proto OrCondition is nil")
	}
	
	if proto.Left == nil || proto.Right == nil {
		return nil, errors.New("proto OrCondition has nil sub-conditions")
	}
	
	left, err := FromProtobuf(proto.Left)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize left condition: %w", err)
	}
	
	right, err := FromProtobuf(proto.Right)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize right condition: %w", err)
	}
	
	return &OrCondition{
		Left:  left,
		Right: right,
	}, nil
}
