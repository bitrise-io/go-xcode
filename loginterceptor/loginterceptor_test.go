package loginterceptor_test

import (
	"bytes"
	"io"
	"regexp"
	"sync"
	"testing"

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
	msg4 := "Last message that won't be sent\n"

	go func() {
		_, _ = sut.Write([]byte(msg1))
		_, _ = sut.Write([]byte(msg2))
		_, _ = sut.Write([]byte(msg3))
		_ = sut.Close()
		_, _ = sut.Write([]byte(msg4))
	}()

	intercepted, target, err := readTwo(interceptReader, targetReader)
	assert.NoError(t, err)
	assert.Equal(t, msg2+msg3, string(intercepted))
	assert.Equal(t, msg1+msg2+msg3, string(target))
}

func TestPrefixInterceptorWithBlockedPipe(t *testing.T) {
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

	target, err := io.ReadAll(targetReader)
	assert.NoError(t, err)
	assert.Equal(t, msg1+msg2+msg3+msg4, string(target))

	// Reading from interceptReader should block until targetWriter is read
	intercepted, err := io.ReadAll(interceptReader)
	assert.NoError(t, err)
	assert.Equal(t, msg2+msg3, string(intercepted))
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
