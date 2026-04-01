package engine

import (
	"bytes"
	"io"
)

// JobStreams captures stdout and stderr
type JobStreams struct {
	stdout *bytes.Buffer
	stderr *bytes.Buffer
	// stdoutWriter and stderrWriter are the actual writers used
	// (may be the buffers above, or external writers passed to NewJobStreams)
	stdoutWriter io.Writer
	stderrWriter io.Writer
}

// NewJobStreams creates new JobStreams with the provided writers.
// If nil writers are passed, internal buffers are used.
func NewJobStreams(stdout, stderr io.Writer) *JobStreams {
	var stdoutBuf, stderrBuf *bytes.Buffer

	if stdout == nil {
		stdoutBuf = bytes.NewBuffer(nil)
		stdout = stdoutBuf
	} else if buf, ok := stdout.(*bytes.Buffer); ok {
		stdoutBuf = buf
	}

	if stderr == nil {
		stderrBuf = bytes.NewBuffer(nil)
		stderr = stderrBuf
	} else if buf, ok := stderr.(*bytes.Buffer); ok {
		stderrBuf = buf
	}

	return &JobStreams{
		stdout:       stdoutBuf,
		stderr:       stderrBuf,
		stdoutWriter: stdout,
		stderrWriter: stderr,
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
	if j.stdoutWriter != nil {
		return j.stdoutWriter
	}
	return j.stdout
}

// StderrWriter returns the stderr writer
func (j *JobStreams) StderrWriter() io.Writer {
	if j.stderrWriter != nil {
		return j.stderrWriter
	}
	return j.stderr
}
