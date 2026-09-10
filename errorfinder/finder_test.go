package errorfinder

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	regularError = `error: exportArchive: "watchkit-app.app" requires a provisioning profile.`
	nsErrorLine  = `Error Domain=IDEProvisioningErrorDomain Code=9 ""watchkit-app.app" requires a provisioning profile." UserInfo={IDEDistributionIssueSeverity=3, NSLocalizedDescription="watchkit-app.app" requires a provisioning profile., NSLocalizedRecoverySuggestion=Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.}`
	// with the leading whitespace xcodebuild indents the continuation lines with
	xcodebuildError = "xcodebuild: error: Failed to build project code-sign-test with scheme code-sign-test.\n" +
		"        Reason: This scheme builds an embedded Apple Watch app. watchOS 9.0 must be installed in order to archive the scheme\n" +
		"        Recovery suggestion: watchOS 9.0 is not installed. To use with Xcode, first download and install the platform."
)

// streamOutput exercises every branch of the state machine, with a multiline error in the middle.
var streamOutput = strings.Join([]string{
	"Build settings from command line:",
	regularError,
	"",
	nsErrorLine,
	"CompileSwift normal arm64 /path/to/File.swift",
	xcodebuildError,
	"A bunch of nonsense",
	" error: store upload failed",
}, "\n")

func writeInChunks(t *testing.T, f *Finder, in string, chunk int) {
	for len(in) > 0 {
		n := min(chunk, len(in))
		written, err := f.Write([]byte(in[:n]))
		require.NoError(t, err)
		require.Equal(t, n, written)
		in = in[n:]
	}
}

func TestFinder_ChunkingDoesNotChangeTheResult(t *testing.T) {
	want := FindXcodebuildErrors(streamOutput)
	require.Len(t, want, 3, "fixture should yield the NSError-merged regular error, the store upload error and the xcodebuild error")

	for _, chunk := range []int{1, 2, 7, 64, len(streamOutput)} {
		t.Run(fmt.Sprintf("%d byte chunks", chunk), func(t *testing.T) {
			f := NewFinder()
			writeInChunks(t, f, streamOutput, chunk)

			assert.Equal(t, want, f.Errors())
		})
	}
}

func TestFinder_ErrorsIsRepeatable(t *testing.T) {
	f := NewFinder()
	writeInChunks(t, f, streamOutput, len(streamOutput))

	first := f.Errors()
	assert.Equal(t, first, f.Errors())
}

func TestFinder_SkipsOverlongLines(t *testing.T) {
	longLine := strings.Repeat("x", maxLineLength+1)
	out := "error: first\n" + longLine + "\nerror: second\n"
	want := []string{"error: first", "error: second"}

	assert.Equal(t, want, FindXcodebuildErrors(out))

	// in chunks as well, so the cap is hit in the middle of the line
	f := NewFinder()
	writeInChunks(t, f, out, 4096)
	assert.Equal(t, want, f.Errors())
}

func TestFinder_OverlongLineEndsMultilineError(t *testing.T) {
	out := "xcodebuild: error: testing failed\nReason: test again\n" +
		strings.Repeat("x", maxLineLength+1) + "\n" +
		"Recovery suggestion: not part of it\n"

	assert.Equal(t, []string{"xcodebuild: error: testing failed\nReason: test again"}, FindXcodebuildErrors(out))
}

func TestFinder_TrailingLineWithoutNewline(t *testing.T) {
	f := NewFinder()
	writeInChunks(t, f, "nonsense\nerror: at the very end", 5)

	assert.Equal(t, []string{"error: at the very end"}, f.Errors())
}
