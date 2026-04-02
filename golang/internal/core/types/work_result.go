package types

// WorkResult represents the result of work execution
type WorkResult struct {
	WorkStatus              WorkStatus
	OutputDirectoriesParent string
	JobStreams              interface{} // JobStreams
}
