package build

// BuildResult is the result of building source
type BuildResult struct {
	ImageID            string
	Pipelines          []interface{} // protobuf Pipeline messages
	SourceBuildResults []SourceBuildResult
}
