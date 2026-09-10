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
	_, _ = sut.XcbuildStderr.Write([]byte(msg3))
	_, _ = sut.XcbuildStderr.Write([]byte(msg4))

	_ = sut.Close()

	assert.Equal(t, msg1+msg4, sut.XcbuildRawout.String())
	assert.Equal(t, msg1+msg4, out.Messages())
}

func TestSetupPipeWiring_XcbuildStreamsShareOneWriter(t *testing.T) {
	sut := logio.SetupPipeWiring(regexp.MustCompile(`^\[Bitrise.*\].*`))
	defer func() { _ = sut.Close() }()

	// == on purpose: os/exec merges the two streams onto one pipe and one copy goroutine only while the two values
	// are equal, which is what keeps the single-writer filter safe. assert.Equal falls back to reflect.DeepEqual,
	// which would also accept two distinct wrappers around the same filter.
	if sut.XcbuildStdout != sut.XcbuildStderr {
		t.Fatal("XcbuildStdout and XcbuildStderr must be the same writer")
	}
}

func TestSetupPipeWiring_TeesReceiveTheCompleteOutput(t *testing.T) {
	tee := &safeBuffer{}
	sut := logio.SetupPipeWiring(regexp.MustCompile(`^\[Bitrise.*\].*`), tee)
	go func() { _, _ = io.Copy(io.Discard, sut.ToolStdin) }()

	_, _ = sut.XcbuildStdout.Write([]byte(msg1))
	_, _ = sut.XcbuildStdout.Write([]byte(msg2))
	_, _ = sut.XcbuildStderr.Write([]byte(msg3))
	_, _ = sut.XcbuildStderr.Write([]byte(msg4))

	_ = sut.Close()

	assert.Equal(t, msg1+msg2+msg3+msg4, tee.String(), "the tee gets everything xcodebuild wrote, on either stream, before filtering")
	assert.Equal(t, msg1+msg4, sut.XcbuildRawout.String())
	if sut.XcbuildStdout != sut.XcbuildStderr {
		t.Fatal("with a tee the two streams must still be the same writer")
	}
}
