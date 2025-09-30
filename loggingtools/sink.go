package loggingtools

import (
	"bufio"
	"io"
)

// Sink is an io.WriteCloser that uses a bufio.Writer to wrap the downstream and
// default buffer sizes for convenience.
type Sink interface {
	io.WriteCloser
	io.StringWriter
}

type sink struct {
	buffer     *bufio.Writer
	downstream io.Writer
}

// NewSink creates a new Sink instance
func NewSink(downstream io.Writer) Sink {
	return &sink{
		buffer:     bufio.NewWriterSize(downstream, 10*1024*1024),
		downstream: downstream,
	}
}

// Write conformance
func (s *sink) Write(p []byte) (int, error) {
	return s.buffer.Write(p)
}

// WriteString conformance
func (s *sink) WriteString(data string) (int, error) {
	return s.buffer.WriteString(data)
}

// Close conformance
func (s *sink) Close() error {
	if err := s.buffer.Flush(); err != nil {
		return err
	}

	return nil
}
