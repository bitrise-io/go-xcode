package logio_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/logio"
)

func BenchmarkPrefixFilterWithMultiWriter(b *testing.B) {
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)

	var wg sync.WaitGroup
	wg.Add(1)
	var buildOutBuffer bytes.Buffer
	stdout := logio.NewSink(os.Stdout)
	var buildOutWriter = logio.NewSink(io.MultiWriter(&buildOutBuffer, stdout))
	sut := logio.NewPrefixFilter(
		re,
		stdout,
		buildOutWriter,
	)

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, _ = sut.Write(fmt.Append(nil, "Log message without prefix: ", i, "\n"))
		_, _ = sut.Write(fmt.Append(nil, "[Bitrise Analytics] Log message with prefix: ", i, "\n"))
		b.StopTimer()

		select {
		case err := <-sut.MessageLost():
			fmt.Printf("Failed on %d: %s", i, err.Error())
			b.FailNow()
		default:
		}
	}
	wg.Done()
}
