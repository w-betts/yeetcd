package build

import "github.com/yeetcd/yeetcd/pkg/config"

// SourceBuildResult is the result of building a single source project
type SourceBuildResult struct {
	YeetcdConfig            config.YeetcdConfig
	OutputDirectoriesParent string
	ImageID                 string
}
