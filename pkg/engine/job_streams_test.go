package engine

import (
	"bytes"
	"io"
	"testing"
)

func TestNewJobStreams_UsesProvidedWriters(t *testing.T) {
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}

	js := NewJobStreams(stdoutBuf, stderrBuf)

	// Write to the writers
	js.StdoutWriter().Write([]byte("stdout test"))
	js.StderrWriter().Write([]byte("stderr test"))

	// Verify the buffers received the data
	if stdoutBuf.String() != "stdout test" {
		t.Errorf("stdout buffer = %q, want %q", stdoutBuf.String(), "stdout test")
	}
	if stderrBuf.String() != "stderr test" {
		t.Errorf("stderr buffer = %q, want %q", stderrBuf.String(), "stderr test")
	}
}

func TestNewJobStreams_NilWritersCreatesBuffers(t *testing.T) {
	js := NewJobStreams(nil, nil)

	// Write to the writers
	js.StdoutWriter().Write([]byte("stdout test"))
	js.StderrWriter().Write([]byte("stderr test"))

	// Verify GetStdOut and GetStdErr return the data
	if string(js.GetStdOut()) != "stdout test" {
		t.Errorf("GetStdOut() = %q, want %q", string(js.GetStdOut()), "stdout test")
	}
	if string(js.GetStdErr()) != "stderr test" {
		t.Errorf("GetStdErr() = %q, want %q", string(js.GetStdErr()), "stderr test")
	}
}

func TestNewJobStreams_MixedWritersAndBuffers(t *testing.T) {
	stdoutBuf := &bytes.Buffer{}

	js := NewJobStreams(stdoutBuf, nil)

	// Write to both
	js.StdoutWriter().Write([]byte("stdout test"))
	js.StderrWriter().Write([]byte("stderr test"))

	// stdout should go to provided buffer
	if stdoutBuf.String() != "stdout test" {
		t.Errorf("stdout buffer = %q, want %q", stdoutBuf.String(), "stdout test")
	}

	// stderr should go to internal buffer
	if string(js.GetStdErr()) != "stderr test" {
		t.Errorf("GetStdErr() = %q, want %q", string(js.GetStdErr()), "stderr test")
	}
}

func TestJobStreams_WriterInterfaces(t *testing.T) {
	js := NewJobStreams(nil, nil)

	// Verify the writers implement io.Writer
	var _ io.Writer = js.StdoutWriter()
	var _ io.Writer = js.StderrWriter()
}
