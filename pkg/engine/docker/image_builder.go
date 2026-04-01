package docker

import (
	"context"
	"errors"

	"github.com/yeetcd/yeetcd/pkg/engine"
)

// DockerDaemonImageBuilder builds Docker images using the Docker daemon
type DockerDaemonImageBuilder struct{}

// NewDockerDaemonImageBuilder creates a new image builder
func NewDockerDaemonImageBuilder() *DockerDaemonImageBuilder {
	return &DockerDaemonImageBuilder{}
}

// BuildImage builds a Docker image using the Docker daemon API
func (b *DockerDaemonImageBuilder) BuildImage(ctx context.Context, dockerClient interface{}, def engine.BuildImageDefinition) (*engine.BuildImageResult, error) {
	return nil, errors.New("not implemented")
}
