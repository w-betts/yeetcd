package events

// PipelineEvent is a marker interface for pipeline events
type PipelineEvent interface {
	IsPipelineEvent()
}
