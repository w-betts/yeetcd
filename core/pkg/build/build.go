package build

import (
	"context"
)

// BuildService interface for building source
type BuildService interface {
	Build(ctx context.Context, source Source) (*BuildResult, error)
}
