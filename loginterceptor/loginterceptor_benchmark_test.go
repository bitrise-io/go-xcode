package loginterceptor_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/loginterceptor"
)

// This fails with some messages being dropped
func BenchmarkPrefixFilterWithMultiWriter(b *testing.B) {
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)

	var wg sync.WaitGroup
	wg.Add(1)
	var buildOutBuffer bytes.Buffer
	stdout := os.Stdout
	var buildOutWriter = io.MultiWriter(&buildOutBuffer, stdout)
	sut := loginterceptor.NewPrefixInterceptor(
		re,
		stdout,
		buildOutWriter,
		log.NewLogger(),
	)

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, _ = sut.Write(fmt.Append(nil, "Log message without prefix: ", i, "\n"))
		_, _ = sut.Write(fmt.Append(nil, "[Bitrise Analytics] Log message with prefix: ", i, "\n"))
		b.StopTimer()

		select {
		case err := <-sut.MessageLost:
			fmt.Printf("Failed on %d: %s", i, err.Error())
			b.FailNow()
		default:
		}
	}
	wg.Done()
}
