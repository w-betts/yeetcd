package condition

import (
	"errors"
	"fmt"

	pipelinepb "github.com/yeetcd/yeetcd/pkg/proto/pipeline"
)

// PREVIOUS_WORK_SUCCESS is a constant condition for previous work success
var PREVIOUS_WORK_SUCCESS Condition

// FromProtobuf converts protobuf Condition to Go Condition
func FromProtobuf(proto *pipelinepb.Condition) (Condition, error) {
	if proto == nil {
		return nil, errors.New("proto condition is nil")
	}
	
	// Use getter methods to check which condition type is set
	if proto.GetWorkContextCondition() != nil {
		return WorkContextConditionFromProtobuf(proto.GetWorkContextCondition())
	}
	if proto.GetPreviousWorkStatusCondition() != nil {
		return PreviousWorkStatusConditionFromProtobuf(proto.GetPreviousWorkStatusCondition())
	}
	if proto.GetAndCondition() != nil {
		return AndConditionFromProtobuf(proto.GetAndCondition())
	}
	if proto.GetOrCondition() != nil {
		return OrConditionFromProtobuf(proto.GetOrCondition())
	}
	if proto.GetNotCondition() != nil {
		return NotConditionFromProtobuf(proto.GetNotCondition())
	}
	
	return nil, fmt.Errorf("no condition type set in protobuf Condition")
}

// WorkContextCondition creates a work context condition
func NewWorkContextConditionFunc(key, expectedValue string, operator Operator) *WorkContextCondition {
	return NewWorkContextCondition(key, expectedValue, operator)
}

// PreviousWorkStatusCondition creates a previous work status condition
func NewPreviousWorkStatusConditionFunc(status WorkStatus) *PreviousWorkStatusCondition {
	return NewPreviousWorkStatusCondition(status)
}

// And creates an AND condition
func And(left, right Condition) *AndCondition {
	return NewAndCondition(left, right)
}

// Or creates an OR condition
func Or(left, right Condition) *OrCondition {
	return NewOrCondition(left, right)
}

// Not creates a NOT condition
func Not(condition Condition) *NotCondition {
	return NewNotCondition(condition)
}

func init() {
	// Initialize PREVIOUS_WORK_SUCCESS
	PREVIOUS_WORK_SUCCESS = &PreviousWorkStatusCondition{Status: WorkStatusSuccess}
}
