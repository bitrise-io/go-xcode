package logio

import (
	"bytes"
	"errors"
	"io"
	"os"
	"regexp"
	"sync"
)

// PipeWiring connects xcodebuild's output to a log formatter and stdout through a prefix filter.
type PipeWiring struct {
	XcbuildRawout *bytes.Buffer
	XcbuildStdout io.Writer
	XcbuildStderr io.Writer
	ToolStdin     io.ReadCloser
	ToolStdout    io.WriteCloser
	ToolStderr    io.WriteCloser

	toolPipeW      *io.PipeWriter
	bufferedStdout *Sink
	toolInSink     *Sink
	filter         *PrefixFilter

	closeFilterOnce sync.Once
}

// CloseFilter closes the filter and waits for it to finish
func (p *PipeWiring) CloseFilter() error {
	err := error(nil)
	p.closeFilterOnce.Do(func() {
		err = p.filter.Close()
		<-p.filter.Done()

	})
	return err
}

func (p *PipeWiring) Close() error {
	filterErr := p.CloseFilter()
	toolSinkErr := p.toolInSink.Close()
	pipeWErr := p.toolPipeW.Close()
	bufferedStdoutErr := p.bufferedStdout.Close()

	return errors.Join(filterErr, toolSinkErr, pipeWErr, bufferedStdoutErr)
}

// SetupPipeWiring creates a new PipeWiring instance that contains the usual
// input/outputs that an xcodebuild command and a logging tool needs when we are also
// using a logging filter.
func SetupPipeWiring(filter *regexp.Regexp) *PipeWiring {
	rawXcbuild := bytes.NewBuffer(nil)
	// Pipe filtered logs to tool
	toolPipeR, toolPipeW := io.Pipe()

	bufferedStdout := NewSink(os.Stdout)
	toolInSink := NewSink(toolPipeW)
	xcbuildLogs := io.MultiWriter(rawXcbuild, toolInSink)
	// Create a filter for [Bitrise ...] prefixes
	bitrisePrefixFilter := NewPrefixFilter(
		filter,
		bufferedStdout,
		xcbuildLogs,
	)

	return &PipeWiring{
		XcbuildRawout: rawXcbuild,
		XcbuildStdout: bitrisePrefixFilter,
		XcbuildStderr: bitrisePrefixFilter,
		ToolStdin:     toolPipeR,
		ToolStdout:    os.Stdout,
		ToolStderr:    os.Stderr,

		toolPipeW:      toolPipeW,
		bufferedStdout: bufferedStdout,
		toolInSink:     toolInSink,
		filter:         bitrisePrefixFilter,

		closeFilterOnce: sync.Once{},
	}
}
