package loginterceptor_test

import (
	"io"
	"regexp"
	"sync"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/loginterceptor"
	"github.com/stretchr/testify/assert"
)

const (
	msg1 = "Log message without prefix\n"
	msg2 = "[Bitrise Analytics] Log message with prefixs\n"
	msg3 = "[Bitrise Build Cache] Log message with prefix\n"
	msg4 = "Stuff [Bitrise Build Cache] Log message without prefix\n"
)

func TestPrefixInterceptor(t *testing.T) {
	interceptedMsgs := NewChanWriterCloser()
	targetMsgs := NewChanWriterCloser()
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)

	sut := loginterceptor.NewPrefixInterceptor(re, &interceptedMsgs, &targetMsgs, log.NewLogger())

	go func() {
		defer func() { _ = sut.Close() }()
		_, _ = sut.Write([]byte(msg1))
		_, _ = sut.Write([]byte(msg2))
		_, _ = sut.Write([]byte(msg3))
		_, _ = sut.Write([]byte(msg4))
	}()

	waitForBoth(sut)
	_ = interceptedMsgs.Close()
	_ = targetMsgs.Close()

	assert.Equal(t, msg2+msg3, interceptedMsgs.Messages())
	assert.Equal(t, msg1+msg2+msg3+msg4, targetMsgs.Messages())
}

func TestPrefixInterceptorWithPrematureClose(t *testing.T) {
	interceptedMsgs := NewChanWriterCloser()
	targetMsgs := NewChanWriterCloser()
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)

	sut := loginterceptor.NewPrefixInterceptor(re, &interceptedMsgs, &targetMsgs, log.NewLogger())

	go func() {
		_, _ = sut.Write([]byte(msg1))
		_, _ = sut.Write([]byte(msg2))
		_, _ = sut.Write([]byte(msg3))
		_ = sut.Close()
		_, _ = sut.Write([]byte(msg4))
	}()

	waitForBoth(sut)
	_ = interceptedMsgs.Close()
	_ = targetMsgs.Close()

	assert.Equal(t, msg2+msg3, interceptedMsgs.Messages())
	assert.Equal(t, msg1+msg2+msg3, targetMsgs.Messages())
}

func TestPrefixInterceptorWithBlockedPipe(t *testing.T) {
	_, interceptWriter := io.Pipe()
	targetMsgs := NewChanWriterCloser()
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)

	sut := loginterceptor.NewPrefixInterceptor(re, interceptWriter, &targetMsgs, log.NewLogger())

	go func() {
		_, _ = sut.Write([]byte(msg1))
		_, _ = sut.Write([]byte(msg2))
		_, _ = sut.Write([]byte(msg3))
		_ = sut.Close()
		_, _ = sut.Write([]byte(msg4))
	}()

	<-sut.TargetDelivered
	_ = targetMsgs.Close()

	assert.Equal(t, msg1+msg2+msg3, targetMsgs.Messages())
}

// --------------------------------
// Helpers
// --------------------------------
func waitForBoth(sut *loginterceptor.PrefixInterceptor) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func(wg *sync.WaitGroup) {
		<-sut.TargetDelivered
		wg.Done()
	}(&wg)

	go func(wg *sync.WaitGroup) {
		<-sut.InterceptedDelivered
		wg.Done()
	}(&wg)

	wg.Wait()
}

type ChanWriterCloser struct {
	channel chan string
}

func NewChanWriterCloser() ChanWriterCloser {
	return ChanWriterCloser{
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
