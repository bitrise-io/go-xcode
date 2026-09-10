package logio

import (
	"bytes"
	"errors"
	"io"
	"os"
	"regexp"
	"sync"
)

// PipeWiring is a helper struct to define the setup and binding of tools and
// xcbuild with a filter and stdout. It is purely boilerplate reduction and it is the
// users responsibility to choose between this and manual hooking of the in/outputs.
// It also provides a convenient Close() method that only closes things that can/should be closed.
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

// Close ...
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
//
// filteredTees also receive the filtered (non-matching) output, that is xcodebuild's own
// lines without the ones matching filter. They are written from a single goroutine, in
// line fragments: a line longer than the filter's read buffer arrives in several writes,
// the last of which ends with the newline.
func SetupPipeWiring(filter *regexp.Regexp, filteredTees ...io.Writer) *PipeWiring {
	// Create a buffer to store raw xcbuild output
	rawXcbuild := bytes.NewBuffer(nil)
	// Pipe filtered logs to tool
	toolPipeR, toolPipeW := io.Pipe()

	// Add a buffer before stdout
	bufferedStdout := NewSink(os.Stdout)
	// Add a buffer before tool input
	toolInSink := NewSink(toolPipeW)
	xcbuildLogs := io.MultiWriter(append([]io.Writer{rawXcbuild, toolInSink}, filteredTees...)...)
	// Create a filter for [Bitrise ...] prefixes
	bitrisePrefixFilter := NewPrefixFilter(
		filter,
		bufferedStdout,
		xcbuildLogs,
	)

	return &PipeWiring{
		XcbuildRawout: rawXcbuild,
		// Deliberately the same writer for both streams (the `2>&1` of the shell pipeline). os/exec
		// merges the two streams onto one pipe, copied by one goroutine, only while cmd.Stdout and
		// cmd.Stderr are the same value; that is what keeps the single-writer filter safe. Do not
		// wrap them separately, and do not split them into two filters.
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
