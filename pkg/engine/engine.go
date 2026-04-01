package engine

import (
	"context"
)

// ExecutionEngine interface for running containers
type ExecutionEngine interface {
	BuildImage(ctx context.Context, def BuildImageDefinition) (*BuildImageResult, error)
	RemoveImage(ctx context.Context, imageID string) error
	RunJob(ctx context.Context, def JobDefinition) (*JobResult, error)
}
