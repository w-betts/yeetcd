package sdk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	pb "github.com/yeetcd/yeetcd/pkg/proto/pipeline"
	"google.golang.org/protobuf/proto"
)

// ToProto converts a Pipeline to its protobuf representation.
func (p Pipeline) ToProto() *pb.Pipeline {
	pbPipeline := &pb.Pipeline{
		Name:        p.Name,
		WorkContext: p.WorkContext,
		FinalWork:   make([]*pb.Work, len(p.FinalWork)),
		Parameters:  make(map[string]*pb.Parameter),
	}

	// Convert parameters
	for name, param := range p.Parameters {
		pbPipeline.Parameters[name] = param.toProto()
	}

	// Convert final work
	for i, work := range p.FinalWork {
		pbPipeline.FinalWork[i] = work.toProto(p.WorkContext)
	}

	return pbPipeline
}

// ToProto converts Pipelines to its protobuf representation.
func (p Pipelines) ToProto() *pb.Pipelines {
	pbPipelines := &pb.Pipelines{
		Pipelines: make([]*pb.Pipeline, len(p)),
	}

	for i, pipeline := range p {
		pbPipelines.Pipelines[i] = pipeline.ToProto()
	}

	return pbPipelines
}

// toProto converts a Work to its protobuf representation.
func (w Work) toProto(containingContext WorkContext) *pb.Work {
	mergedContext := w.WorkContext.Merge(containingContext)

	pbWork := &pb.Work{
		Id:           w.id(containingContext),
		Description:  w.Description,
		WorkContext:  mergedContext,
		OutputPaths:  make([]*pb.WorkOutputPath, len(w.OutputPaths)),
		PreviousWork: make([]*pb.PreviousWork, len(w.PreviousWork)),
	}

	// Convert output paths
	for i, path := range w.OutputPaths {
		pbWork.OutputPaths[i] = path.toProto()
	}

	// Convert previous work
	for i, prev := range w.PreviousWork {
		pbWork.PreviousWork[i] = prev.toProto(containingContext)
	}

	// Convert condition
	if w.Condition != nil {
		pbWork.Condition = w.Condition.toProtoCondition()
	}

	// Apply work definition to the oneof field
	w.WorkDefinition.applyToProto(pbWork)

	return pbWork
}

// id generates a unique ID for a work item based on its content.
func (w Work) id(containingContext WorkContext) string {
	// Generate a deterministic ID based on work content
	data := fmt.Sprintf("%s:%v:%v", w.Description, w.WorkContext, containingContext)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// toProto converts a WorkOutputPath to its protobuf representation.
func (w WorkOutputPath) toProto() *pb.WorkOutputPath {
	return &pb.WorkOutputPath{
		Name: w.Name,
		Path: w.Path,
	}
}

// toProto converts a PreviousWork to its protobuf representation.
func (p PreviousWork) toProto(containingContext WorkContext) *pb.PreviousWork {
	return &pb.PreviousWork{
		Work:             p.Work.toProto(containingContext),
		OutputPathsMount: p.OutputPathsMount,
		StdOutEnvVar:     p.StdOutEnvVar,
	}
}

// toProto converts a Parameter to its protobuf representation.
func (p Parameter) toProto() *pb.Parameter {
	pbParam := &pb.Parameter{
		TypeCheck: p.TypeCheck.toProto(),
		Required:  p.Required,
		Choices:   p.Choices,
	}

	if p.DefaultValue != "" {
		pbParam.DefaultValue = proto.String(p.DefaultValue)
	}

	return pbParam
}

// toProto converts a TypeCheck to its protobuf representation.
func (t TypeCheck) toProto() pb.Parameter_TYPE_CHECK {
	switch t {
	case TypeCheckString:
		return pb.Parameter_STRING
	case TypeCheckNumber:
		return pb.Parameter_NUMBER
	case TypeCheckBoolean:
		return pb.Parameter_BOOLEAN
	default:
		return pb.Parameter_STRING
	}
}

// applyToProto applies the work definition to the protobuf Work message.
func (c *ContainerisedWorkDefinition) applyToProto(pbWork *pb.Work) {
	pbWork.OneofTaskActions = &pb.Work_ContainerisedWorkDefinition{
		ContainerisedWorkDefinition: &pb.ContainerisedWorkDefinition{
			Image: c.Image,
			Cmd:   c.Cmd,
		},
	}
}

// applyToProto applies the work definition to the protobuf Work message.
func (c *CustomWorkDefinition) applyToProto(pbWork *pb.Work) {
	pbWork.OneofTaskActions = &pb.Work_CustomWorkDefinition{
		CustomWorkDefinition: &pb.CustomWorkDefinition{
			ExecutionId: c.executionID(),
		},
	}
}

// applyToProto applies the work definition to the protobuf Work message.
func (c *CompoundWorkDefinition) applyToProto(pbWork *pb.Work) {
	pbCompound := &pb.CompoundWorkDefinition{
		FinalWork: make([]*pb.Work, len(c.FinalWork)),
	}

	// Note: CompoundWork doesn't have a containing context, use empty
	for i, work := range c.FinalWork {
		pbCompound.FinalWork[i] = work.toProto(WorkContext{})
	}

	pbWork.OneofTaskActions = &pb.Work_CompoundWorkDefinition{
		CompoundWorkDefinition: pbCompound,
	}
}

// applyToProto applies the work definition to the protobuf Work message.
func (d *DynamicWorkGeneratingWorkDefinition) applyToProto(pbWork *pb.Work) {
	pbWork.OneofTaskActions = &pb.Work_DynamicWorkGeneratingWorkDefinition{
		DynamicWorkGeneratingWorkDefinition: &pb.DynamicWorkGeneratingWorkDefinition{
			ExecutionId: generateExecutionID(),
		},
	}
}

func generateExecutionID() string {
	data := fmt.Sprintf("%d", workIDCounter.Add(1))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// toProtoCondition converts a Condition to its protobuf representation.
func (c *WorkContextCondition) toProtoCondition() *pb.Condition {
	return &pb.Condition{
		Conditions: &pb.Condition_WorkContextCondition{
			WorkContextCondition: &pb.WorkContextCondition{
				Key:     c.Key,
				Operand: c.Operand.toProto(),
				Value:   c.Value,
			},
		},
	}
}

// toProto converts an Operand to its protobuf representation.
func (o Operand) toProto() pb.WorkContextCondition_Operand {
	switch o {
	case OperandEquals:
		return pb.WorkContextCondition_EQUALS
	default:
		return pb.WorkContextCondition_EQUALS
	}
}

// toProtoCondition converts a Condition to its protobuf representation.
func (n *NotCondition) toProtoCondition() *pb.Condition {
	return &pb.Condition{
		Conditions: &pb.Condition_NotCondition{
			NotCondition: &pb.NotCondition{
				Condition: n.Condition.toProtoCondition(),
			},
		},
	}
}

// toProtoCondition converts a Condition to its protobuf representation.
func (a *AndCondition) toProtoCondition() *pb.Condition {
	return &pb.Condition{
		Conditions: &pb.Condition_AndCondition{
			AndCondition: &pb.AndCondition{
				Left:  a.Left.toProtoCondition(),
				Right: a.Right.toProtoCondition(),
			},
		},
	}
}

// toProtoCondition converts a Condition to its protobuf representation.
func (o *OrCondition) toProtoCondition() *pb.Condition {
	return &pb.Condition{
		Conditions: &pb.Condition_OrCondition{
			OrCondition: &pb.OrCondition{
				Left:  o.Left.toProtoCondition(),
				Right: o.Right.toProtoCondition(),
			},
		},
	}
}

// toProtoCondition converts a Condition to its protobuf representation.
func (p *PreviousWorkStatusCondition) toProtoCondition() *pb.Condition {
	return &pb.Condition{
		Conditions: &pb.Condition_PreviousWorkStatusCondition{
			PreviousWorkStatusCondition: &pb.PreviousWorkStatusCondition{
				Status: p.Status.toProto(),
			},
		},
	}
}

// toProto converts a Status to its protobuf representation.
func (s Status) toProto() pb.PreviousWorkStatusCondition_Status {
	switch s {
	case StatusSuccess:
		return pb.PreviousWorkStatusCondition_SUCCESS
	case StatusFailure:
		return pb.PreviousWorkStatusCondition_FAILURE
	case StatusAny:
		return pb.PreviousWorkStatusCondition_ANY
	default:
		return pb.PreviousWorkStatusCondition_SUCCESS
	}
}
