package loggingtools

import (
	"bytes"
	"io"
	"os"
	"regexp"

	"github.com/bitrise-io/go-utils/v2/log"
)

// XCBuildWithLoggingToolIO is a helper struct to define the setup and binding of tools and
// xcbuild with a filter and stdout. It is purely boilerplate reduction and it is the
// users responsibility to choose between this and manual hooking of the in/outputs.
// It also provides a convenient Close() method that only closes things that can/should be closed.
type XCBuildWithLoggingToolIO struct {
	XcbuildRawout bytes.Buffer
	XcbuildStdout io.Writer
	XcbuildStderr io.Writer
	ToolStdin     io.ReadCloser
	ToolStdout    io.WriteCloser
	ToolStderr    io.WriteCloser
	Filter        PrefixFilter
}

// Close closes the IO instances that needs to be closing as part of this instance.
//
// In reality it can only clos the filter and the tool input as everything else is
// managed by a command or the os.
func (x *XCBuildWithLoggingToolIO) Close(logger log.Logger) {
	// XcbuildRawout - no need to close
	// XcbuildStdout - Multiwriter, meaning we need to close the subwriters
	// XcbuildStderr - Multiwriter, meaning we need to close the subwriters
	if err := x.ToolStdin.Close(); err != nil {
		logger.Warnf("Failed to close xcodebuild-xcpretty pipe, error: %s", err)
	}
	// ToolStdout - We are not closing stdout
	// ToolSterr - We are not closing stderr
	if err := x.Filter.Close(); err != nil {
		logger.Warnf("Failed to close log interceptor, error: %s", err)
	}
}

// SetupLoggingIO creates a new XCBuildWithLoggingToolIO instance that contains the usual
// input/outputs that an xcodebuild command and a logging tool needs when we are also
// using a logging filter.
func SetupLoggingIO() *XCBuildWithLoggingToolIO {
	// Create a buffer to store raw xcbuild output
	var rawXcbuild bytes.Buffer
	// Pipe filtered logs to tool
	toolPipeR, toolPipeW := io.Pipe()

	// Add a buffer before stdout
	bufferedStdout := NewSink(os.Stdout)
	// Add a buffer before tool input
	xcbuildLogs := NewSink(toolPipeW)
	// Create a filter for [Bitrise ...] prefixes
	bitrisePrefixFilter := NewPrefixFilter(
		regexp.MustCompile(`^\[Bitrise.*\].*`),
		bufferedStdout,
		xcbuildLogs,
	)

	// Send raw xcbuild out to raw out and filter
	rawInputDuplication := io.MultiWriter(&rawXcbuild, bitrisePrefixFilter)

	return &XCBuildWithLoggingToolIO{
		XcbuildRawout: rawXcbuild,
		XcbuildStdout: rawInputDuplication,
		XcbuildStderr: rawInputDuplication,
		ToolStdin:     toolPipeR,
		ToolStdout:    bufferedStdout,
		ToolStderr:    os.Stderr,
		Filter:        bitrisePrefixFilter,
	}
}
