package pipeline

// PipelineMetadata contains metadata about a pipeline
type PipelineMetadata struct {
	PipelineName     string
	BuiltSourceImage string
	SourceLanguage   interface{} // SourceLanguage
}
