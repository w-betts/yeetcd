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
// If external writers are passed, internal buffers are also created to capture
// output for GetStdOut/GetStdErr methods - the external writer still receives
// the output via the writer interface.
func NewJobStreams(stdout, stderr io.Writer) *JobStreams {
	// Always create internal buffers to ensure GetStdOut/GetStdErr work correctly
	// even when external writers (like *os.File) are provided
	stdoutBuf := bytes.NewBuffer(nil)
	stderrBuf := bytes.NewBuffer(nil)

	// If nil was passed, use internal buffer as the writer
	// Otherwise, wrap with a tee writer so both internal buffer and external writer receive output
	if stdout == nil {
		stdout = stdoutBuf
	} else {
		// Tee: write to both internal buffer and external writer
		stdout = io.MultiWriter(stdoutBuf, stdout)
	}

	if stderr == nil {
		stderr = stderrBuf
	} else {
		// Tee: write to both internal buffer and external writer
		stderr = io.MultiWriter(stderrBuf, stderr)
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
