package pipeline

import (
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



// PreviousWorkStdOutAsWorkContext returns previous work stdout as work context
func (w *Work) PreviousWorkStdOutAsWorkContext(tracker *WorkResultTracker) WorkContext {
	ctx := make(WorkContext)
	for _, prevWork := range w.PreviousWork {
		if prevWork.StdOutEnvVar != "" {
			stdout := tracker.StdOut(prevWork.Work)
			ctx[prevWork.StdOutEnvVar] = stdout
		}
	}
	return ctx
}

// PreviousWorkMountInputs returns mount inputs from previous work
func (w *Work) PreviousWorkMountInputs(tracker *WorkResultTracker) map[string]interface{} {
	mounts := make(map[string]interface{})
	for _, prevWork := range w.PreviousWork {
		if prevWork.OutputPathsMount != "" {
			mountInput := tracker.OutputDirectoriesMountInput(prevWork.Work)
			if mountInput != nil {
				mounts[prevWork.OutputPathsMount] = mountInput
			}
		}
	}
	return mounts
}

// OutputDirectoryPaths returns output directory paths
func (w *Work) OutputDirectoryPaths() map[string]string {
	paths := make(map[string]string, len(w.OutputPaths))
	for _, outputPath := range w.OutputPaths {
		paths[outputPath.Name] = outputPath.Path
	}
	return paths
}
