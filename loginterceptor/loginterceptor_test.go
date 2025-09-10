package loginterceptor_test

import (
	"bytes"
	"io"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/loginterceptor"
	"github.com/stretchr/testify/assert"
)

func TestPrefixInterceptor(t *testing.T) {
	interceptReader, interceptWriter := io.Pipe()
	targetReader, targetWriter := io.Pipe()
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)

	sut := loginterceptor.NewPrefixInterceptor(re, interceptWriter, targetWriter, log.NewLogger())

	msg1 := "Log message without prefix\n"
	msg2 := "[Bitrise Analytics] Log message with prefix\n"
	msg3 := "[Bitrise Build Cache] Log message with prefix\n"
	msg4 := "Stuff [Bitrise Build Cache] Log message without prefix\n"

	go func() {
		//nolint:errCheck
		defer sut.Close()

		_, _ = sut.Write([]byte(msg1))
		_, _ = sut.Write([]byte(msg2))
		_, _ = sut.Write([]byte(msg3))
		_, _ = sut.Write([]byte(msg4))
	}()

	intercepted, target, err := readTwo(interceptReader, targetReader)
	assert.NoError(t, err)
	assert.Equal(t, msg2+msg3, string(intercepted))
	assert.Equal(t, msg1+msg2+msg3+msg4, string(target))
}

func TestPrefixInterceptorWithPrematureClose(t *testing.T) {
	interceptReader, interceptWriter := io.Pipe()
	targetReader, targetWriter := io.Pipe()
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)

	sut := loginterceptor.NewPrefixInterceptor(re, interceptWriter, targetWriter, log.NewLogger())

	msg1 := "Log message without prefix\n"
	msg2 := "[Bitrise Analytics] Log message with prefix\n"
	msg3 := "[Bitrise Build Cache] Log message with prefix\n"
	msg4 := "Stuff [Bitrise Build Cache] Log message without prefix\n"

	go func() {

		_, _ = sut.Write([]byte(msg1))
		_, _ = sut.Write([]byte(msg2))
		_, _ = sut.Write([]byte(msg3))
		sut.Close()
		_, _ = sut.Write([]byte(msg4))
	}()

	intercepted, target, err := readTwo(interceptReader, targetReader)
	assert.NoError(t, err)
	assert.Equal(t, msg2+msg3, string(intercepted))
	assert.Equal(t, msg1+msg2+msg3, string(target))
}

func TestPrefixInterceptorWhenWritersAreBlocking(t *testing.T) {
	blockingWriter := newBlockingWriter()
	targetReader, targetWriter := io.Pipe()
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)

	sut := loginterceptor.NewPrefixInterceptorWithTimeout(re, blockingWriter, targetWriter, log.NewLogger(), 50*time.Millisecond)

	msg1 := "Log message without prefix\n"
	msg2 := "[Bitrise Analytics] Log message with prefix\n"
	msg3 := "[Bitrise Build Cache] Log message with prefix\n"
	msg4 := "Stuff [Bitrise Build Cache] Log message without prefix\n"

	go func() {
		//nolint:errCheck
		defer sut.Close()

		_, _ = sut.Write([]byte(msg1))
		time.Sleep(10 * time.Millisecond)
		blockingWriter.Allow()
		_, _ = sut.Write([]byte(msg2))
		time.Sleep(10 * time.Millisecond)
		blockingWriter.Allow()
		_, _ = sut.Write([]byte(msg3))
		time.Sleep(10 * time.Millisecond)
		blockingWriter.Allow()
		_, _ = sut.Write([]byte(msg4))
		time.Sleep(10 * time.Millisecond)
		blockingWriter.Allow()
	}()

	target, err := io.ReadAll(targetReader)
	assert.NoError(t, err)
	assert.Equal(t, msg2+msg3, string(blockingWriter.buf))
	assert.Equal(t, msg1+msg2+msg3+msg4, string(target))
}

func TestPrefixInterceptorWhenWriteIsSlowerThenTimeout(t *testing.T) {
	blockingWriter := newBlockingWriter()
	targetReader, targetWriter := io.Pipe()
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)

	sut := loginterceptor.NewPrefixInterceptorWithTimeout(re, blockingWriter, targetWriter, log.NewLogger(), 10*time.Millisecond)

	msg1 := "Log message without prefix\n"
	msg2 := "[Bitrise Analytics] Log message with prefix\n"
	msg3 := "[Bitrise Build Cache] Log message with prefix\n"
	msg4 := "Stuff [Bitrise Build Cache] Log message without prefix\n"

	go func() {
		//nolint:errCheck
		defer sut.Close()

		_, _ = sut.Write([]byte(msg1))
		time.Sleep(30 * time.Millisecond)
		blockingWriter.Allow()
		_, _ = sut.Write([]byte(msg2))
		time.Sleep(30 * time.Millisecond)
		blockingWriter.Allow()
		_, _ = sut.Write([]byte(msg3))
		time.Sleep(30 * time.Millisecond)
		blockingWriter.Allow()
		_, _ = sut.Write([]byte(msg4))
		time.Sleep(30 * time.Millisecond)
		blockingWriter.Allow()
	}()

	target, err := io.ReadAll(targetReader)
	assert.NoError(t, err)
	assert.Equal(t, msg2+msg3, string(blockingWriter.buf))
	assert.Equal(t, msg1+msg2+msg3+msg4, string(target))
}

func readTwo(r1, r2 io.Reader) (out1, out2 []byte, err error) {
	var (
		wg     sync.WaitGroup
		e1, e2 error
	)
	wg.Add(2)

	var b1, b2 bytes.Buffer

	go func() {
		defer wg.Done()
		_, e1 = io.Copy(&b1, r1)
	}()

	go func() {
		defer wg.Done()
		_, e2 = io.Copy(&b2, r2)
	}()

	wg.Wait()

	// prefer to return the first non-nil error
	if e1 != nil {
		return b1.Bytes(), b2.Bytes(), e1
	}
	if e2 != nil {
		return b1.Bytes(), b2.Bytes(), e2
	}
	return b1.Bytes(), b2.Bytes(), nil
}

// blockingWriter blocks on each Write until an allow signal is received.
type blockingWriter struct {
	allow  chan struct{}
	buf    []byte
	mu     sync.Mutex
	closed bool
}

func newBlockingWriter() *blockingWriter {
	return &blockingWriter{allow: make(chan struct{})}
}

func (w *blockingWriter) Write(p []byte) (int, error) {
	// block until allowed
	<-w.allow
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return 0, io.ErrClosedPipe
	}
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *blockingWriter) Allow() {
	// Non-blocking safe: close or send depending on usage.
	// For repeated allows, use a buffered channel or recreate it.
	select {
	case w.allow <- struct{}{}:
	default:
	}
}

func (w *blockingWriter) Close() {
	w.mu.Lock()
	w.closed = true
	w.mu.Unlock()
}
