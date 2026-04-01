package condition

import (
	"errors"
	"fmt"

	pipelinepb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
	"github.com/yeetcd/yeetcd/internal/core/pipeline"
)

// WorkStatus represents previous work status
type WorkStatus int

const (
	WorkStatusSuccess WorkStatus = iota
	WorkStatusFailure
	WorkStatusAny
)

// SUCCESS is a backward-compatible alias for WorkStatusSuccess
const SUCCESS = WorkStatusSuccess

// PreviousWorkStatusCondition checks previous work status
type PreviousWorkStatusCondition struct {
	Status WorkStatus
}

// NewPreviousWorkStatusCondition creates a new PreviousWorkStatusCondition
func NewPreviousWorkStatusCondition(status WorkStatus) *PreviousWorkStatusCondition {
	return &PreviousWorkStatusCondition{
		Status: status,
	}
}

// Evaluate checks if previous work status matches condition
func (p *PreviousWorkStatusCondition) Evaluate(workContext pipeline.WorkContext, workResultTracker *pipeline.WorkResultTracker) (bool, error) {
	if workResultTracker == nil {
		return false, errors.New("previous work result is nil")
	}
	
	// Get the last work result from the tracker
	lastResult := workResultTracker.GetLastResult("")
	if lastResult == nil {
		return false, errors.New("no previous work result available")
	}
	
	switch p.Status {
	case WorkStatusSuccess:
		return lastResult.WorkStatus == pipeline.WorkStatusSucceeded, nil
	case WorkStatusFailure:
		return lastResult.WorkStatus == pipeline.WorkStatusFailed, nil
	case WorkStatusAny:
		return true, nil
	default:
		return false, fmt.Errorf("unknown work status: %v", p.Status)
	}
}

// ToProtobuf serializes the condition to protobuf
func (p *PreviousWorkStatusCondition) ToProtobuf() (*pipelinepb.Condition, error) {
	var protoStatus pipelinepb.PreviousWorkStatusCondition_Status
	switch p.Status {
	case WorkStatusSuccess:
		protoStatus = pipelinepb.PreviousWorkStatusCondition_SUCCESS
	case WorkStatusFailure:
		protoStatus = pipelinepb.PreviousWorkStatusCondition_FAILURE
	case WorkStatusAny:
		protoStatus = pipelinepb.PreviousWorkStatusCondition_ANY
	default:
		return nil, fmt.Errorf("unknown work status: %v", p.Status)
	}
	
	return &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_PreviousWorkStatusCondition{
			PreviousWorkStatusCondition: &pipelinepb.PreviousWorkStatusCondition{
				Status: protoStatus,
			},
		},
	}, nil
}

// FromProtobuf converts protobuf PreviousWorkStatusCondition to Go struct
func PreviousWorkStatusConditionFromProtobuf(proto *pipelinepb.PreviousWorkStatusCondition) (*PreviousWorkStatusCondition, error) {
	if proto == nil {
		return nil, errors.New("proto PreviousWorkStatusCondition is nil")
	}
	
	var status WorkStatus
	switch proto.Status {
	case pipelinepb.PreviousWorkStatusCondition_SUCCESS:
		status = WorkStatusSuccess
	case pipelinepb.PreviousWorkStatusCondition_FAILURE:
		status = WorkStatusFailure
	case pipelinepb.PreviousWorkStatusCondition_ANY:
		status = WorkStatusAny
	default:
		return nil, fmt.Errorf("unknown proto status: %v", proto.Status)
	}
	
	return &PreviousWorkStatusCondition{
		Status: status,
	}, nil
}
