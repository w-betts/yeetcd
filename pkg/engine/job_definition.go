package engine

// JobDefinition defines a job to run
type JobDefinition struct {
	Image                string
	Cmd                  []string
	WorkingDir           string
	Environment          map[string]string
	InputFilePaths       map[string]MountInput
	OutputDirectoryPaths map[string]string
	JobStreams           *JobStreams
}

// MountInput interface for mount inputs
type MountInput interface {
	Directory() string
}

// OnDiskMountInput implements MountInput
type OnDiskMountInput struct {
	Dir string
}

// Directory returns the directory path
func (o OnDiskMountInput) Directory() string {
	return o.Dir
}
