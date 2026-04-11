package condition

import (
	"errors"
	"fmt"

	pipelinepb "github.com/yeetcd/yeetcd/pkg/proto/pipeline"
	"github.com/yeetcd/yeetcd/internal/core/types"
)

// Operator represents comparison operators
type Operator int

const (
	OperatorEquals Operator = iota
)

// WorkContextCondition checks work context values
type WorkContextCondition struct {
	Key           string
	ExpectedValue string
	Operator      Operator
}

// NewWorkContextCondition creates a new WorkContextCondition
func NewWorkContextCondition(key, expectedValue string, operator Operator) *WorkContextCondition {
	return &WorkContextCondition{
		Key:           key,
		ExpectedValue: expectedValue,
		Operator:      operator,
	}
}

// Evaluate checks if work context matches condition
func (w *WorkContextCondition) Evaluate(workContext types.WorkContext, workResultTracker types.WorkResultTracker) (bool, error) {
	if workContext == nil {
		return false, errors.New("work context is nil")
	}

	value, exists := workContext[w.Key]
	if !exists {
		return false, fmt.Errorf("key '%s' not found in work context", w.Key)
	}

	switch w.Operator {
	case OperatorEquals:
		return value == w.ExpectedValue, nil
	default:
		return false, fmt.Errorf("unknown operator: %v", w.Operator)
	}
}

// ToProtobuf serializes the condition to protobuf
func (w *WorkContextCondition) ToProtobuf() (*pipelinepb.Condition, error) {
	var protoOperator pipelinepb.WorkContextCondition_Operand
	switch w.Operator {
	case OperatorEquals:
		protoOperator = pipelinepb.WorkContextCondition_EQUALS
	default:
		return nil, fmt.Errorf("unknown operator: %v", w.Operator)
	}

	return &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_WorkContextCondition{
			WorkContextCondition: &pipelinepb.WorkContextCondition{
				Key:     w.Key,
				Value:   w.ExpectedValue,
				Operand: protoOperator,
			},
		},
	}, nil
}

// WorkContextConditionFromProtobuf converts protobuf WorkContextCondition to Go struct
func WorkContextConditionFromProtobuf(proto *pipelinepb.WorkContextCondition) (*WorkContextCondition, error) {
	if proto == nil {
		return nil, errors.New("proto WorkContextCondition is nil")
	}

	var operator Operator
	switch proto.Operand {
	case pipelinepb.WorkContextCondition_EQUALS:
		operator = OperatorEquals
	default:
		return nil, fmt.Errorf("unknown proto operator: %v", proto.Operand)
	}

	return &WorkContextCondition{
		Key:           proto.Key,
		ExpectedValue: proto.Value,
		Operator:      operator,
	}, nil
}
