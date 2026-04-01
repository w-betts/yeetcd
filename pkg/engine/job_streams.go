package engine

import (
	"bytes"
	"io"
)

// JobStreams captures stdout and stderr
type JobStreams struct {
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

// NewJobStreams creates new JobStreams
func NewJobStreams(stdout, stderr io.Writer) *JobStreams {
	return &JobStreams{
		stdout: bytes.NewBuffer(nil),
		stderr: bytes.NewBuffer(nil),
	}
}

// GetStdOut returns captured stdout
func (j *JobStreams) GetStdOut() []byte {
	return j.stdout.Bytes()
}

// GetStdErr returns captured stderr
func (j *JobStreams) GetStdErr() []byte {
	return j.stderr.Bytes()
}

// StdoutWriter returns the stdout writer
func (j *JobStreams) StdoutWriter() io.Writer {
	return j.stdout
}

// StderrWriter returns the stderr writer
func (j *JobStreams) StderrWriter() io.Writer {
	return j.stderr
}
