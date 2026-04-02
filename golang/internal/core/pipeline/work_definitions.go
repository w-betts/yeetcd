package pipeline

import (
	"fmt"

	pb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
)

// FromWorkProtobuf converts protobuf Work to appropriate WorkDefinition type
func FromWorkProtobuf(protoWork *pb.Work) (WorkDefinition, error) {
	if protoWork == nil {
		return nil, fmt.Errorf("protoWork is nil")
	}

	// Check which work definition type is set using getter methods
	if customDef := protoWork.GetCustomWorkDefinition(); customDef != nil {
		return &CustomWorkDefinition{
			ExecutionID: customDef.GetExecutionId(),
		}, nil
	}

	if containerisedDef := protoWork.GetContainerisedWorkDefinition(); containerisedDef != nil {
		return &ContainerisedWorkDefinition{
			Image: containerisedDef.GetImage(),
			Cmd:   containerisedDef.GetCmd(),
		}, nil
	}

	if compoundDef := protoWork.GetCompoundWorkDefinition(); compoundDef != nil {
		finalWork := make([]Work, 0, len(compoundDef.FinalWork))
		for _, protoW := range compoundDef.FinalWork {
			if protoW != nil {
				w, err := WorkFromProtobuf(protoW)
				if err != nil {
					return nil, fmt.Errorf("failed to convert compound work: %w", err)
				}
				finalWork = append(finalWork, *w)
			}
		}
		return &CompoundWorkDefinition{
			FinalWork: finalWork,
		}, nil
	}

	if dynamicDef := protoWork.GetDynamicWorkGeneratingWorkDefinition(); dynamicDef != nil {
		return &DynamicWorkGeneratingWorkDefinition{
			ExecutionID: dynamicDef.GetExecutionId(),
		}, nil
	}

	// No work definition set
	return nil, fmt.Errorf("no work definition set in protobuf Work")
}
