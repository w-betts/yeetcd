package pipeline

import (
	"context"
	"fmt"

	pb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
)

// Work represents a unit of work in a pipeline
type Work struct {
	ID             string
	Description    string
	WorkContext    WorkContext
	WorkDefinition WorkDefinition
	Condition      interface{} // Condition interface
	OutputPaths    []WorkOutputPath
	PreviousWork   []PreviousWork
}

// WorkFromProtobuf converts a protobuf Work message to Go Work struct
func WorkFromProtobuf(protoWork *pb.Work) (*Work, error) {
	if protoWork == nil {
		return nil, fmt.Errorf("protoWork is nil")
	}

	work := &Work{
		ID:          protoWork.Id,
		Description: protoWork.Description,
		WorkContext: make(WorkContext),
		OutputPaths: make([]WorkOutputPath, 0, len(protoWork.OutputPaths)),
		PreviousWork: make([]PreviousWork, 0, len(protoWork.PreviousWork)),
	}

	// Convert work context
	for k, v := range protoWork.WorkContext {
		work.WorkContext[k] = v
	}

	// Convert work definition
	workDef, err := FromWorkProtobuf(protoWork)
	if err != nil {
		return nil, fmt.Errorf("failed to convert work definition: %w", err)
	}
	work.WorkDefinition = workDef

	// Convert output paths
	for _, protoPath := range protoWork.OutputPaths {
		if protoPath != nil {
			work.OutputPaths = append(work.OutputPaths, WorkOutputPath{
				Name: protoPath.Name,
				Path: protoPath.Path,
			})
		}
	}

	// Convert previous work
	for _, protoPrevWork := range protoWork.PreviousWork {
		if protoPrevWork != nil && protoPrevWork.Work != nil {
			prevWork, err := WorkFromProtobuf(protoPrevWork.Work)
			if err != nil {
				return nil, fmt.Errorf("failed to convert previous work: %w", err)
			}
			
			work.PreviousWork = append(work.PreviousWork, PreviousWork{
				Work:             *prevWork,
				OutputPathsMount: protoPrevWork.OutputPathsMount,
				StdOutEnvVar:     protoPrevWork.StdOutEnvVar,
			})
		}
	}

	return work, nil
}

// Execute runs the work with the given context and engine
func (w *Work) Execute(ctx context.Context, containingContext WorkContext, engine interface{}, metadata PipelineMetadata, tracker *WorkResultTracker, handler interface{}) (*WorkResult, error) {
	return nil, fmt.Errorf("not implemented")
}

// PreviousWorkStdOutAsWorkContext returns previous work stdout as work context
func (w *Work) PreviousWorkStdOutAsWorkContext() WorkContext {
	return WorkContext{}
}

// PreviousWorkMountInputs returns mount inputs from previous work
func (w *Work) PreviousWorkMountInputs() map[string]interface{} {
	return nil
}

// OutputDirectoryPaths returns output directory paths
func (w *Work) OutputDirectoryPaths() map[string]string {
	return nil
}
