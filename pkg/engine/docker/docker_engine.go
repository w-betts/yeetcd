package docker

import (
	"context"
	"errors"

	"github.com/yeetcd/yeetcd/pkg/engine"
)

// DockerExecutionEngine implements the ExecutionEngine interface using Docker
type DockerExecutionEngine struct {
	// dockerClient is the Docker client (will be *client.Client when implemented)
	dockerClient interface{}
}

// NewDockerExecutionEngine creates a new Docker execution engine
func NewDockerExecutionEngine() (*DockerExecutionEngine, error) {
	return nil, errors.New("not implemented")
}

// BuildImage builds a Docker image from the given definition
func (d *DockerExecutionEngine) BuildImage(ctx context.Context, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
	return nil, errors.New("not implemented")
}

// RemoveImage removes a Docker image
func (d *DockerExecutionEngine) RemoveImage(ctx context.Context, imageID string) error {
	return errors.New("not implemented")
}

// RunJob runs a job in a Docker container
func (d *DockerExecutionEngine) RunJob(ctx context.Context, def engine.JobDefinition) (*engine.JobResult, error) {
	return nil, errors.New("not implemented")
}
