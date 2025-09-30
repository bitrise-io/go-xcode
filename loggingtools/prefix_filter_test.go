//nolint:errcheck
package loggingtools_test

import (
	"io"
	"regexp"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/loggingtools"
	"github.com/stretchr/testify/assert"
)

const (
	msg1 = "Log message without prefix\n"
	msg2 = "[Bitrise Analytics] Log message with prefixs\n"
	msg3 = "[Bitrise Build Cache] Log message with prefix\n"
	msg4 = "Stuff [Bitrise Build Cache] Log message without prefix\n"
)

func TestPrefixInterceptor(t *testing.T) {
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)
	matching := NewChanWriterCloser()
	matchingSink := loggingtools.NewSink(matching)
	rest := NewChanWriterCloser()
	restSink := loggingtools.NewSink(rest)

	sut := loggingtools.NewPrefixFilter(
		re,
		matchingSink,
		restSink,
	)

	_, _ = sut.Write([]byte(msg1))
	_, _ = sut.Write([]byte(msg2))
	_, _ = sut.Write([]byte(msg3))
	_, _ = sut.Write([]byte(msg4))

	_ = sut.Close()
	<-sut.Done()
	_ = matchingSink.Close()
	_ = restSink.Close()
	matching.Close()
	rest.Close()

	assert.Equal(t, msg2+msg3, matching.Messages())
	assert.Equal(t, msg1+msg4, rest.Messages())
}

func TestPrefixInterceptorWithPrematureClose(t *testing.T) {
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)
	matching := NewChanWriterCloser()
	matchingSink := loggingtools.NewSink(matching)
	rest := NewChanWriterCloser()
	restSink := loggingtools.NewSink(rest)

	sut := loggingtools.NewPrefixFilter(
		re,
		matchingSink,
		restSink,
	)

	_, _ = sut.Write([]byte(msg1))
	_, _ = sut.Write([]byte(msg2))
	_, _ = sut.Write([]byte(msg3))
	_ = sut.Close()
	_, _ = sut.Write([]byte(msg4))

	<-sut.Done()

	_ = restSink.Close()
	_ = matchingSink.Close()
	matching.Close()
	rest.Close()

	assert.Equal(t, msg1, rest.Messages())
	assert.Equal(t, msg2+msg3, matching.Messages())
}

func TestPrefixInterceptorWithBlockedPipe(t *testing.T) {
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)
	_, matching := io.Pipe()
	matchingSink := loggingtools.NewSink(matching)
	rest := NewChanWriterCloser()
	restSink := loggingtools.NewSink(rest)

	sut := loggingtools.NewPrefixFilter(
		re,
		matchingSink,
		restSink,
	)

	_, _ = sut.Write([]byte(msg1))
	_, _ = sut.Write([]byte(msg2))
	_, _ = sut.Write([]byte(msg3))
	_ = sut.Close()
	_, _ = sut.Write([]byte(msg4))

	<-sut.Done()

	_ = restSink.Close()
	matching.Close()
	rest.Close()

	assert.Equal(t, msg1, rest.Messages())
}

// --------------------------------
// Helpers
// --------------------------------
type ChanWriterCloser struct {
	channel chan string
}

func NewChanWriterCloser() *ChanWriterCloser {
	return &ChanWriterCloser{
		channel: make(chan string, 1000),
	}
}

func (ch *ChanWriterCloser) Write(p []byte) (int, error) {
	ch.channel <- string(p)
	return len(p), nil
}

// Close stops the interceptor and closes the pipe.
func (ch *ChanWriterCloser) Close() error {
	close(ch.channel)
	return nil
}

func (ch *ChanWriterCloser) Messages() string {
	var result string
	for msg := range ch.channel {
		result += msg
	}
	return result
}
