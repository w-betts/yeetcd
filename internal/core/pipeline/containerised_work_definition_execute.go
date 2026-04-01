package pipeline

import (
	"context"
	"fmt"

	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// Execute implements WorkDefinition for ContainerisedWorkDefinition
// Creates JobStreams via PipelineOutputHandler, records WorkStarted event,
// builds JobDefinition with image/cmd/workingDir/env/mounts/outputs,
// calls ExecutionEngine.RunJob(), maps exit code to WorkStatus (0=SUCCESS, non-zero=FAILURE)
func (c *ContainerisedWorkDefinition) Execute(ctx context.Context, work Work, eng engine.ExecutionEngine, metadata PipelineMetadata, tracker *WorkResultTracker, handler PipelineOutputHandler) (*types.WorkResult, error) {
	// Create JobStreams via PipelineOutputHandler
	jobStreams := handler.NewJobStreams()

	// Record WorkStarted event
	handler.RecordEvent(WorkStarted{
		Work:       work,
		JobStreams: jobStreams,
	})

	// Build merged context
	mergedContext := work.WorkContext

	// Add previous work stdout as environment variables
	prevWorkStdOutContext := work.PreviousWorkStdOutAsWorkContext(tracker)
	for k, v := range prevWorkStdOutContext {
		mergedContext[k] = v
	}

	// Get mount inputs from previous work
	mountInputs := work.PreviousWorkMountInputs(tracker)

	// Convert mount inputs to engine.MountInput
	inputFilePaths := make(map[string]engine.MountInput)
	for path, mountInput := range mountInputs {
		if onDisk, ok := mountInput.(engine.OnDiskMountInput); ok {
			inputFilePaths[path] = onDisk
		}
	}

	// Get output directory paths
	outputPaths := work.OutputDirectoryPaths()

	// Build JobDefinition
	jobDef := engine.JobDefinition{
		Image:                c.Image,
		Cmd:                  c.Cmd,
		WorkingDir:           c.WorkingDir,
		Environment:          mergedContext,
		InputFilePaths:       inputFilePaths,
		OutputDirectoryPaths: outputPaths,
		JobStreams:           jobStreams.(*engine.JobStreams),
	}

	// Call ExecutionEngine.RunJob()
	jobResult, err := eng.RunJob(ctx, jobDef)
	if err != nil {
		return nil, fmt.Errorf("failed to run job: %w", err)
	}

	// Map exit code to WorkStatus
	var workStatus types.WorkStatus
	if jobResult.ExitCode == 0 {
		workStatus = types.SUCCESS
	} else {
		workStatus = types.FAILURE
	}

	return &types.WorkResult{
		WorkStatus:              workStatus,
		OutputDirectoriesParent: jobResult.OutputDirectoriesParent,
		JobStreams:              jobStreams,
	}, nil
}
