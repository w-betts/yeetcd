package build

// SourceBuildResult is the result of building a single source project
type SourceBuildResult struct {
	YeetcdConfig           interface{} // Will be pkg/config.YeetcdConfig
	OutputDirectoriesParent string
}
