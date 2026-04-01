package build

import (
	"context"
	"errors"

	"github.com/yeetcd/yeetcd/pkg/engine"
)

// DockerBuildService implements BuildService using Docker
type DockerBuildService struct {
	engine engine.ExecutionEngine
}

// NewDockerBuildService creates a new Docker build service
func NewDockerBuildService(eng engine.ExecutionEngine) *DockerBuildService {
	return &DockerBuildService{engine: eng}
}

// Build builds the source code and returns the result
func (d *DockerBuildService) Build(ctx context.Context, source Source) (*BuildResult, error) {
	return nil, errors.New("not implemented")
}
