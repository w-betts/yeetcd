package pipeline

import (
	"context"
	"fmt"

	pb "github.com/yeetcd/yeetcd/pkg/proto/pipeline"
	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/pkg/engine"
	"google.golang.org/protobuf/proto"
)

// Execute implements WorkDefinition for DynamicWorkGeneratingWorkDefinition
// Delegates to NativeWorkDefinition.Execute() with executionId.
// After successful execution, parses stdout as protobuf Work message and recursively executes the generated work.
func (d *DynamicWorkGeneratingWorkDefinition) Execute(ctx context.Context, work Work, mergedContext types.WorkContext, eng engine.ExecutionEngine, metadata PipelineMetadata, tracker *WorkResultTracker, handler PipelineOutputHandler) (*types.WorkResult, error) {
	native := &NativeWorkDefinition{
		ExecutionID: d.ExecutionID,
	}
	
	// Execute the native work definition
	result, err := native.Execute(ctx, work, mergedContext, eng, metadata, tracker, handler)
	if err != nil {
		return nil, err
	}

	// If successful, parse stdout as protobuf Work message and execute generated work
	if result.WorkStatus == types.SUCCESS {
		generatedWork, err := d.ParseAndCreateWork(result.JobStreams)
		if err != nil {
			return nil, fmt.Errorf("failed to parse generated work: %w", err)
		}

		// Execute the generated work recursively
		generatedResult, err := generatedWork.Execute(ctx, mergedContext, eng, metadata, tracker, handler)
		if err != nil {
			return nil, fmt.Errorf("failed to execute generated work: %w", err)
		}

		// Return the generated work's result
		return generatedResult, nil
	}

	return result, nil
}

// ParseAndCreateWork parses stdout as protobuf Work message and creates a Work struct
func (d *DynamicWorkGeneratingWorkDefinition) ParseAndCreateWork(jobStreams interface{}) (*Work, error) {
	if jobStreams == nil {
		return nil, fmt.Errorf("job streams is nil")
	}

	streams, ok := jobStreams.(*engine.JobStreams)
	if !ok {
		return nil, fmt.Errorf("invalid job streams type")
	}

	stdout := streams.GetStdOut()
	if len(stdout) == 0 {
		return nil, fmt.Errorf("no stdout to parse")
	}

	// Parse stdout as protobuf Work message
	var protoWork pb.Work
	if err := proto.Unmarshal(stdout, &protoWork); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf work: %w", err)
	}

	// Convert to Go Work struct
	return WorkFromProtobuf(&protoWork)
}
