package logio

import (
	"io"
	"time"

	"github.com/globocom/go-buffer/v2"
)

// Sink is an io.WriteCloser that uses a bufio.Writer to wrap the downstream and
// default buffer sizes for convenience.
type Sink interface {
	io.WriteCloser
}

type sink struct {
	buffer *buffer.Buffer
}

// NewSink creates a new Sink instance
func NewSink(downstream io.Writer) Sink {
	return &sink{
		buffer: buffer.New(
			// Flush after five writes
			buffer.WithSize(5),
			// Flushed every second if not full
			buffer.WithFlushInterval(time.Second),
			// Flush writes to downstream
			buffer.WithFlusher(buffer.FlusherFunc(func(items []interface{}) {
				for _, item := range items {
					downstream.Write(item.([]byte))
				}
			})),
		),
	}
}

// Write conformance
func (s *sink) Write(p []byte) (int, error) {
	return len(p), s.buffer.Push(p)
}

// Close conformance
func (s *sink) Close() error {
	return s.buffer.Close()
}
