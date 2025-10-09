package logio_test

import (
	"io"
	"regexp"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/logio"
	"github.com/stretchr/testify/assert"
)

func TestPipeWiring(t *testing.T) {
	sut := logio.SetupPipeWiring(regexp.MustCompile(`^\[Bitrise.*\].*`))

	out := NewChanWriterCloser()
	go func() {
		_, _ = io.Copy(out, sut.ToolStdin)
		_ = out.Close()
	}()

	_, _ = sut.XcbuildStdout.Write([]byte(msg1))
	_, _ = sut.XcbuildStdout.Write([]byte(msg2))
	_, _ = sut.XcbuildStdout.Write([]byte(msg3))
	_, _ = sut.XcbuildStdout.Write([]byte(msg4))

	_ = sut.Close()

	assert.Equal(t, msg1+msg4, sut.XcbuildRawout.String())
	assert.Equal(t, msg1+msg4, out.Messages())
}
