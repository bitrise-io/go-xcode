package loggingtools

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sync"
)

// PrefixFilter intercept writes: when the message has a prefix that matches a
// regexp it writes into the `Matching` sink, otherwise to the `Filtered` sink.
//
// Note: Callers are responsible for closing `Matching` and `Filtered` Sinks
type PrefixFilter interface {
	io.WriteCloser

	// A Sink (buffered io.Writer) that will emit only matching lines
	Matching() Sink
	// A Sink (buffered io.Writer) that will emit only non-matching lines
	Filtered() Sink
	// A channel that will signal when the filter is done processing messages
	// (ie safe to close downstream io)
	Done() <-chan struct{}
	// A channel that will emit errors on each message lost. Intended for tests.
	MessageLost() <-chan error
	// An error field that should contain the scanner error if any after done.
	ScannerError() error
}

type prefixFilter struct {
	prefixRegexp *regexp.Regexp

	// internal buffered middleman between xcbuild and scan
	xcBuildOutput bufio.ReadWriter
	pipeR         *io.PipeReader
	pipeW         *io.PipeWriter

	matching Sink
	filtered Sink

	// closing
	closeOnce sync.Once

	done         chan struct{}
	messageLost  chan error
	scannerError error
}

// Matching conformance
func (p *prefixFilter) Matching() Sink { return p.matching }

// Filtered conformance
func (p *prefixFilter) Filtered() Sink { return p.filtered }

// Done conformance
func (p *prefixFilter) Done() <-chan struct{} { return p.done }

// MessageLost conformance
func (p *prefixFilter) MessageLost() <-chan error { return p.messageLost }

// ScannerError conformance
func (p *prefixFilter) ScannerError() error { return p.scannerError }

// NewPrefixFilter returns a new PrefixFilter. Writes are based on line prefix.
//
// Note: Callers are responsible for closing intercepted and target writers that implement io.Closer
func NewPrefixFilter(prefixRegexp *regexp.Regexp, matching, filtered Sink) PrefixFilter {
	pipeR, pipeW := io.Pipe()
	xcbuildOut := bufio.NewReader(pipeR)
	scanIn := bufio.NewWriter(pipeW)

	filter := &prefixFilter{
		prefixRegexp:  prefixRegexp,
		xcBuildOutput: *bufio.NewReadWriter(xcbuildOut, scanIn),
		pipeR:         pipeR,
		pipeW:         pipeW,
		matching:      matching,
		filtered:      filtered,
		closeOnce:     sync.Once{},
		messageLost:   make(chan error, 1),
		done:          make(chan struct{}, 1),
	}
	go filter.run()
	return filter
}

// Write implements io.Writer. It writes into an internal pipe which the interceptor goroutine consumes.
func (i *prefixFilter) Write(p []byte) (int, error) {
	return i.xcBuildOutput.Write(p)
}

// Close stops the interceptor and closes the pipe.
func (i *prefixFilter) Close() error {
	var errString string
	i.closeOnce.Do(func() {
		// Flush and close scanner input
		if err := i.xcBuildOutput.Flush(); err != nil {
			errString += fmt.Sprintf("failed to flush xcbuildoutput (%v)", err.Error())
		}
		if err := i.pipeW.Close(); err != nil {
			if len(errString) != 0 {
				errString += ", "
			}
			errString += fmt.Sprintf("failed to close scanner input (%v)", err.Error())
		}
	})
	return fmt.Errorf("failed to close prefixFilter: %s", errString)
}

// run reads lines (and partial final chunk) and writes them.
func (i *prefixFilter) run() {
	defer func() {
		// Signal done and close signaling channels
		i.done <- struct{}{}
		close(i.done)
		close(i.messageLost)
	}()

	// Use a scanner but with a large buffer to handle long lines.
	scanner := bufio.NewScanner(i.xcBuildOutput)
	const maxTokenSize = 10 * 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxTokenSize)

	for scanner.Scan() {
		line := scanner.Text() // note: newline removed
		// re-append newline to preserve same output format
		logLine := line + "\n"

		if i.prefixRegexp.MatchString(line) {
			if _, err := i.matching.WriteString(logLine); err != nil {
				i.messageLost <- fmt.Errorf("intercepting message: %w", err)
			}
		} else {
			if _, err := i.filtered.Write([]byte(logLine)); err != nil {
				i.messageLost <- fmt.Errorf("intercepting message: %w", err)
			}
		}
	}

	// handle any scanner error
	i.scannerError = scanner.Err()
}
