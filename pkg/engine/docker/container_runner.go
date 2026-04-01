package docker

import (
	"context"
	"errors"

	"github.com/yeetcd/yeetcd/pkg/engine"
)

// ContainerRunner handles container lifecycle operations
type ContainerRunner struct{}

// NewContainerRunner creates a new container runner
func NewContainerRunner() *ContainerRunner {
	return &ContainerRunner{}
}

// PullImage pulls a Docker image if not present locally
func (r *ContainerRunner) PullImage(ctx context.Context, dockerClient interface{}, imageTag string) error {
	return errors.New("not implemented")
}

// CreateContainer creates a new container with the given configuration
func (r *ContainerRunner) CreateContainer(ctx context.Context, dockerClient interface{}, def engine.JobDefinition) (string, error) {
	return "", errors.New("not implemented")
}

// RunContainer runs a container and captures its output
func (r *ContainerRunner) RunContainer(ctx context.Context, dockerClient interface{}, containerID string, streams *engine.JobStreams) (int, error) {
	return -1, errors.New("not implemented")
}

// ExtractArchive extracts files from a container path
func (r *ContainerRunner) ExtractArchive(ctx context.Context, dockerClient interface{}, containerID string, path string) ([]byte, error) {
	return nil, errors.New("not implemented")
}
