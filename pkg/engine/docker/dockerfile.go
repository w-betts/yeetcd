package docker

import (
	"context"
	"errors"

	"github.com/yeetcd/yeetcd/pkg/engine"
)

// CreateDockerfile generates a Dockerfile for building images
func CreateDockerfile(ctx context.Context, def engine.BuildImageDefinition, contextDir string) (string, func(), error) {
	return "", nil, errors.New("not implemented")
}
