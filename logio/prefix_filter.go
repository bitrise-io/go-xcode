package logio

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sync"
)

// readBufferSize is the size of the line fragments a longer line is forwarded in.
const readBufferSize = 64 * 1024

// PrefixFilter intercept writes: when the message has a prefix that matches a
// regexp it writes into the `Matching` sink, otherwise to the `Filtered` sink.
//
// PrefixFilter is not safe for concurrent use: Write buffers into a bufio.Writer, and
// it expects a single writer. Handing the same PrefixFilter to both exec.Cmd.Stdout and
// exec.Cmd.Stderr satisfies that, os/exec copies both streams with one goroutine then.
//
// Note: Callers are responsible for closing `Matching` and `Filtered` Sinks
type PrefixFilter struct {
	prefixRegexp *regexp.Regexp

	// internal buffered middleman between xcbuild and scan
	filterInput *bufio.Writer
	pipeW       *io.PipeWriter
	pipeR       *io.PipeReader

	Matching *Sink
	Filtered io.Writer

	// closing
	closeOnce sync.Once

	done         chan struct{}
	messageLost  chan error
	scannerError chan error
}

// Done returns a channel on which the user can observe when the last messages are
// written to the outputs. The channel has a buffer of one to prevent early unreceived
// messages or late subscriptions to the channel.
func (p *PrefixFilter) Done() <-chan struct{} { return p.done }

// MessageLost returns a channel on which the user can observe if there were
// messages lost. The channel has a buffer of one to prevent early unreceived
// messages or late subscriptions to the channel. Reports never block: a lost
// message is reported if the previous report has been received already, and
// dropped otherwise.
func (p *PrefixFilter) MessageLost() <-chan error { return p.messageLost }

// ScannerError returns a channel on which the user can observe if there were
// any scanner errors. The channel has a buffer of one to prevent early unreceived
// messages or late subscriptions to the channel.
func (p *PrefixFilter) ScannerError() <-chan error { return p.scannerError }

// NewPrefixFilter returns a new PrefixFilter. Writes are based on line prefix.
//
// Note: Callers are responsible for closing intercepted and target writers that implement io.Closer
func NewPrefixFilter(prefixRegexp *regexp.Regexp, matching *Sink, filtered io.Writer) *PrefixFilter {
	pipeR, pipeW := io.Pipe()
	messageLost := make(chan error, 1)
	done := make(chan struct{}, 1)
	scannerError := make(chan error, 1)

	filter := &PrefixFilter{
		prefixRegexp: prefixRegexp,
		filterInput:  bufio.NewWriter(pipeW),
		pipeW:        pipeW,
		pipeR:        pipeR,
		closeOnce:    sync.Once{},
		messageLost:  messageLost,
		done:         done,
		scannerError: scannerError,

		Matching: matching,
		Filtered: filtered,
	}
	go filter.run()
	return filter
}

// Write implements io.Writer. It writes into an internal pipe which the interceptor goroutine consumes.
// It must not be called concurrently, see PrefixFilter.
func (p *PrefixFilter) Write(data []byte) (int, error) {
	return p.filterInput.Write(data)
}

// Close stops the interceptor and closes the pipe.
func (p *PrefixFilter) Close() error {
	var errString string
	p.closeOnce.Do(func() {
		// Flush and close scanner input
		if err := p.filterInput.Flush(); err != nil {
			errString += fmt.Sprintf("failed to flush xcbuildoutput (%v)", err.Error())
		}
		if err := p.pipeW.Close(); err != nil {
			if len(errString) > 0 {
				errString += ", "
			}
			errString += fmt.Sprintf("failed to close scanner input (%v)", err.Error())
		}
	})

	if len(errString) > 0 {
		return fmt.Errorf("failed to close prefixFilter: %s", errString)
	}

	return nil
}

// run reads lines (and partial final chunk) and writes them.
func (p *PrefixFilter) run() {
	defer func() {
		// With the reader gone a writer parked in Write would block forever, so unblock it.
		_ = p.pipeR.Close()

		// Signal done and close signaling channels
		p.done <- struct{}{}
		close(p.done)
		close(p.messageLost)
		close(p.scannerError)
	}()

	// A line longer than the buffer is forwarded in fragments rather than buffered whole (a
	// Scanner would give up on it and stop). The destination is picked on the first fragment,
	// the prefix regexp is anchored so that is enough, and kept until the newline.
	reader := bufio.NewReaderSize(p.pipeR, readBufferSize)
	var dst io.Writer
	for {
		fragment, err := reader.ReadSlice('\n')
		if len(fragment) > 0 {
			if dst == nil {
				dst = p.Filtered
				if p.prefixRegexp.Match(bytes.TrimSuffix(fragment, []byte{'\n'})) {
					dst = p.Matching
				}
			}
			// Copy: fragment is a view into the reader's buffer, and Sink keeps what it is given.
			logLine := bytes.Clone(fragment)
			if err == io.EOF {
				// re-append newline to the partial final chunk to preserve same output format
				logLine = append(logLine, '\n')
			}
			if _, werr := dst.Write(logLine); werr != nil {
				p.reportMessageLost(werr)
			}
			if logLine[len(logLine)-1] == '\n' {
				dst = nil
			}
		}

		switch err {
		case nil, bufio.ErrBufferFull:
			continue
		case io.EOF:
			return
		default:
			select {
			case p.scannerError <- err:
			default:
			}
			return
		}
	}
}

// reportMessageLost reports err on the MessageLost channel without blocking. Nothing is required to drain
// the channel, and a blocking send here would stop the scanner, block every Write and hang the command.
func (p *PrefixFilter) reportMessageLost(err error) {
	select {
	case p.messageLost <- fmt.Errorf("intercepting message: %w", err):
	default:
	}
}
