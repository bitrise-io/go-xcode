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
// tees also receive xcodebuild's complete output, both streams, as it arrives: in the
// chunks os/exec copies, before the filter, from a single goroutine. A tee that needs
// lines has to assemble them itself.
func SetupPipeWiring(filter *regexp.Regexp, tees ...io.Writer) *PipeWiring {
	// Create a buffer to store raw xcbuild output
	rawXcbuild := bytes.NewBuffer(nil)
	// Pipe filtered logs to tool
	toolPipeR, toolPipeW := io.Pipe()

	// Add a buffer before stdout
	bufferedStdout := NewSink(os.Stdout)
	// Add a buffer before tool input
	toolInSink := NewSink(toolPipeW)
	xcbuildLogs := io.MultiWriter(rawXcbuild, toolInSink)
	// Create a filter for [Bitrise ...] prefixes
	bitrisePrefixFilter := NewPrefixFilter(
		filter,
		bufferedStdout,
		xcbuildLogs,
	)

	// One value for both of xcodebuild's streams (the `2>&1` of the shell pipeline). os/exec
	// merges the two streams onto one pipe, copied by one goroutine, only while cmd.Stdout and
	// cmd.Stderr are the same value; that is what keeps the single-writer filter safe. Do not
	// wrap the two fields separately, and do not split them into two filters.
	xcbuildOut := io.Writer(bitrisePrefixFilter)
	if len(tees) > 0 {
		xcbuildOut = io.MultiWriter(append([]io.Writer{bitrisePrefixFilter}, tees...)...)
	}

	return &PipeWiring{
		XcbuildRawout: rawXcbuild,
		XcbuildStdout: xcbuildOut,
		XcbuildStderr: xcbuildOut,
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
