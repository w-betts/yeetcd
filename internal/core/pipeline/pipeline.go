package pipeline

import (
	"fmt"

	pb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
)

// Pipeline represents a pipeline definition
type Pipeline struct {
	Name        string
	Parameters  Parameters
	WorkContext WorkContext
	FinalWork   []*Work
	Metadata    PipelineMetadata
}

// FromProtobuf converts a protobuf Pipeline message to Go Pipeline struct
func FromProtobuf(protoPipeline *pb.Pipeline) (*Pipeline, error) {
	if protoPipeline == nil {
		return nil, fmt.Errorf("protoPipeline is nil")
	}

	pipeline := &Pipeline{
		Name:        protoPipeline.Name,
		Parameters:  make(Parameters),
		WorkContext: make(WorkContext),
		FinalWork:   make([]*Work, 0, len(protoPipeline.FinalWork)),
	}

	// Convert parameters
	for name, protoParam := range protoPipeline.Parameters {
		if protoParam == nil {
			continue
		}
		
		var typeCheck TypeCheck
		switch protoParam.TypeCheck {
		case pb.Parameter_STRING:
			typeCheck = STRING
		case pb.Parameter_NUMBER:
			typeCheck = NUMBER
		case pb.Parameter_BOOLEAN:
			typeCheck = BOOLEAN
		default:
			typeCheck = STRING
		}

		pipeline.Parameters[name] = Parameter{
			TypeCheck:    typeCheck,
			Required:     protoParam.Required,
			DefaultValue: protoParam.GetDefaultValue(),
			Choices:      protoParam.Choices,
		}
	}

	// Convert work context
	for k, v := range protoPipeline.WorkContext {
		pipeline.WorkContext[k] = v
	}

	// Convert final work
	for _, protoWork := range protoPipeline.FinalWork {
		work, err := WorkFromProtobuf(protoWork)
		if err != nil {
			return nil, fmt.Errorf("failed to convert work: %w", err)
		}
		pipeline.FinalWork = append(pipeline.FinalWork, work)
	}

	return pipeline, nil
}

// WithArguments merges arguments into pipeline work context
func (p *Pipeline) WithArguments(args Arguments) (*Pipeline, error) {
	// Validate arguments against parameters and existing context
	validatedContext, err := args.AsValidatedWorkContext(p.Parameters, p.WorkContext)
	if err != nil {
		return nil, fmt.Errorf("argument validation failed: %w", err)
	}

	// Create new pipeline with merged context
	newPipeline := &Pipeline{
		Name:        p.Name,
		Parameters:  p.Parameters,
		FinalWork:   p.FinalWork,
		Metadata:    p.Metadata,
		WorkContext: validatedContext.MergeInto(p.WorkContext),
	}

	return newPipeline, nil
}
