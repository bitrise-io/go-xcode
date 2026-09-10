package logio_test

import (
	"fmt"
	"io"
	"regexp"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/logio"
)

// BenchmarkPrefixFilterSmallWrites is the NSUnbufferedIO=YES shape: many ~100 byte lines, one Write each, then a
// full drain. Close and Done are inside the timed region so the whole write-and-drain cost is measured.
func BenchmarkPrefixFilterSmallWrites(b *testing.B) {
	re := regexp.MustCompile(`^\[Bitrise.*\].*`)
	lines := make([][]byte, 0, 2*b.N)
	for i := 0; i < b.N; i++ {
		lines = append(lines,
			fmt.Appendf(nil, "CompileSwift normal arm64 /Users/vagrant/git/App/Sources/File%d.swift (in target 'App')\n", i),
			fmt.Appendf(nil, "[Bitrise Build Cache] hit for compilation unit %d\n", i))
	}

	b.ResetTimer()
	matching := logio.NewSink(io.Discard)
	sut := logio.NewPrefixFilter(re, matching, logio.NewSink(io.Discard))
	for _, line := range lines {
		_, _ = sut.Write(line)
	}
	_ = sut.Close()
	<-sut.Done()
	_ = matching.Close()
}
