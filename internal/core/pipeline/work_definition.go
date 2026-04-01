package pipeline

import (
	"context"
	"errors"
)

// WorkDefinition is the interface for all work definition types
type WorkDefinition interface {
	Execute(ctx context.Context, work Work, engine interface{}, metadata PipelineMetadata, tracker *WorkResultTracker, handler interface{}) (*WorkResult, error)
}

// ContainerisedWorkDefinition runs a command in an existing container image
type ContainerisedWorkDefinition struct {
	Image      string
	Cmd        []string
	WorkingDir string
}

// Execute implements WorkDefinition
func (c *ContainerisedWorkDefinition) Execute(ctx context.Context, work Work, engine interface{}, metadata PipelineMetadata, tracker *WorkResultTracker, handler interface{}) (*WorkResult, error) {
	return nil, errors.New("not implemented")
}

// CustomWorkDefinition executes user-defined code
type CustomWorkDefinition struct {
	ExecutionID string
}

// Execute implements WorkDefinition
func (c *CustomWorkDefinition) Execute(ctx context.Context, work Work, engine interface{}, metadata PipelineMetadata, tracker *WorkResultTracker, handler interface{}) (*WorkResult, error) {
	return nil, errors.New("not implemented")
}

// CompoundWorkDefinition groups multiple work items
type CompoundWorkDefinition struct {
	FinalWork []Work
}

// Execute implements WorkDefinition
func (c *CompoundWorkDefinition) Execute(ctx context.Context, work Work, engine interface{}, metadata PipelineMetadata, tracker *WorkResultTracker, handler interface{}) (*WorkResult, error) {
	return nil, errors.New("not implemented")
}

// DynamicWorkGeneratingWorkDefinition generates work at runtime
type DynamicWorkGeneratingWorkDefinition struct {
	ExecutionID string
}

// Execute implements WorkDefinition
func (d *DynamicWorkGeneratingWorkDefinition) Execute(ctx context.Context, work Work, engine interface{}, metadata PipelineMetadata, tracker *WorkResultTracker, handler interface{}) (*WorkResult, error) {
	return nil, errors.New("not implemented")
}
