package pipeline

import (
	"context"
	"fmt"

	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/pkg/engine"
)

// NativeWorkDefinition provides shared execution logic for CustomWorkDefinition and DynamicWorkGeneratingWorkDefinition
type NativeWorkDefinition struct {
	ExecutionID string
}

// Execute runs the native work definition
// Creates JobStreams, records WorkStarted event, builds JobDefinition using
// pipelineMetadata.builtSourceImage and sourceLanguage.GetCustomTaskRunnerCmd(),
// calls ExecutionEngine.RunJob(), maps exit code to WorkStatus
func (n *NativeWorkDefinition) Execute(ctx context.Context, work Work, eng engine.ExecutionEngine, metadata PipelineMetadata, tracker *WorkResultTracker, handler PipelineOutputHandler) (*types.WorkResult, error) {
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
	prevWorkStdOutContext := work.PreviousWorkStdOutAsWorkContext()
	for k, v := range prevWorkStdOutContext {
		mergedContext[k] = v
	}

	// Get mount inputs from previous work
	mountInputs := work.PreviousWorkMountInputs()

	// Convert mount inputs to engine.MountInput
	inputFilePaths := make(map[string]engine.MountInput)
	for path, mountInput := range mountInputs {
		if onDisk, ok := mountInput.(engine.OnDiskMountInput); ok {
			inputFilePaths[path] = onDisk
		}
	}

	// Get output directory paths
	outputPaths := work.OutputDirectoryPaths()

	// Get the command from source language
	var cmd []string
	if metadata.SourceLanguage != nil {
		if sourceLang, ok := metadata.SourceLanguage.(interface {
			GetCustomTaskRunnerCmd(pipelineName, taskName string) []string
		}); ok {
			cmd = sourceLang.GetCustomTaskRunnerCmd(metadata.PipelineName, n.ExecutionID)
		}
	}

	// Build JobDefinition using builtSourceImage
	jobDef := engine.JobDefinition{
		Image:                metadata.BuiltSourceImage,
		Cmd:                  cmd,
		WorkingDir:           "/",
		Environment:          mergedContext,
		InputFilePaths:       inputFilePaths,
		OutputDirectoryPaths: outputPaths,
		JobStreams:           jobStreams.(*engine.JobStreams),
	}

	// Call ExecutionEngine.RunJob()
	jobResult, err := eng.RunJob(ctx, jobDef)
	if err != nil {
		return nil, fmt.Errorf("failed to run native job: %w", err)
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
