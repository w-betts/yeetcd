package engine

// JobResult is the result of running a job
type JobResult struct {
	ExitCode               int
	OutputDirectoriesParent string
}
