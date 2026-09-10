//nolint:errcheck
package logio_test

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bitrise-io/go-xcode/v2/logio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A line twice the size of the buffer the old bufio.Scanner gave up at; giving up meant everything after it was lost
// and, with nothing reading the pipe any more, the next Write blocked forever.
func TestPrefixFilter_LongLineIsForwardedNotDropped(t *testing.T) {
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)
	matching := &safeBuffer{}
	matchingSink := logio.NewSink(matching)
	rest := &safeBuffer{}

	sut := logio.NewPrefixFilter(re, matchingSink, rest)

	longLine := strings.Repeat("x", 20*1024*1024) + "\n"
	_, _ = sut.Write([]byte(longLine))
	_, _ = sut.Write([]byte(msg2))
	_, _ = sut.Write([]byte(msg1))

	require.NoError(t, sut.Close())
	<-sut.Done()
	require.NoError(t, matchingSink.Close())

	assert.Equal(t, longLine+msg1, rest.String())
	assert.Equal(t, msg2, matching.String())
	select {
	case err := <-sut.ScannerError():
		assert.NoError(t, err)
	default:
	}
}

// Nothing in production drains MessageLost, so a blocking send there stalled the scanner on the second lost message.
func TestPrefixFilter_LostMessagesDoNotStallTheFilter(t *testing.T) {
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)
	matchingSink := logio.NewSink(&safeBuffer{})
	failing := &failingWriter{}

	sut := logio.NewPrefixFilter(re, matchingSink, failing)

	const lines = 100
	for i := 0; i < lines; i++ {
		_, _ = sut.Write([]byte(msg1))
	}

	done := make(chan struct{})
	go func() {
		_ = sut.Close()
		<-sut.Done()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("filter stalled on lost messages")
	}

	assert.EqualValues(t, lines, failing.calls.Load(), "every line should still have been offered to the destination")
	assert.Error(t, <-sut.MessageLost(), "the first lost message is reported")
	_, more := <-sut.MessageLost()
	assert.False(t, more, "the rest are dropped and the channel is closed")
}

// --------------------------------
// Helpers
// --------------------------------

// safeBuffer is a bytes.Buffer that can be written by a Sink's flusher goroutine and read by the test.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

type failingWriter struct {
	calls atomic.Int32
}

func (w *failingWriter) Write(p []byte) (int, error) {
	w.calls.Add(1)
	return 0, errors.New("destination gone")
}
