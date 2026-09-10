package xcodecommand

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const xcodebuildFixture = `Build settings from command line:
[Bitrise Analytics] a Bitrise line that happens to contain error: nope
CompileSwift normal arm64 /path/to/File.swift
error: exportArchive: "code-sign-test.app" requires a provisioning profile.
xcodebuild: error: Failed to build project code-sign-test with scheme code-sign-test.
        Reason: This scheme builds an embedded Apple Watch app. watchOS 9.0 must be installed in order to archive the scheme
        Recovery suggestion: watchOS 9.0 is not installed. To use with Xcode, first download and install the platform
** TEST FAILED **
`

var wantErrorLines = []string{
	`error: exportArchive: "code-sign-test.app" requires a provisioning profile.`,
	"xcodebuild: error: Failed to build project code-sign-test with scheme code-sign-test.\n" +
		"Reason: This scheme builds an embedded Apple Watch app. watchOS 9.0 must be installed in order to archive the scheme\n" +
		"Recovery suggestion: watchOS 9.0 is not installed. To use with Xcode, first download and install the platform",
}

func TestXcbeautifyRunner_Run(t *testing.T) {
	var opts *command.Opts
	failure := command.NewExitStatusError("xcodebuild test", exitError(t, 65), nil)

	build := new(mocks.Command)
	build.On("PrintableCommandArgs").Return("xcodebuild test")
	build.On("Start").Return(nil)
	build.On("Wait").Run(func(mock.Arguments) {
		// os/exec copies the child's output between Start and Wait
		_, _ = io.WriteString(opts.Stdout, xcodebuildFixture)
	}).Return(failure)

	tool := new(mocks.Command)
	tool.On("PrintableCommandArgs").Return(xcbeautify)
	tool.On("Start").Return(nil)
	tool.On("Wait").Return(nil)

	factory := new(mocks.CommandFactory)
	factory.On("Create", "xcodebuild", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		opts = args.Get(2).(*command.Opts)
	}).Return(build)
	factory.On("Create", xcbeautify, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// xcbeautify would read its stdin; drain it so the pipeline does not stall
		go func() { _, _ = io.Copy(io.Discard, args.Get(2).(*command.Opts).Stdin) }()
	}).Return(tool)

	out, err := NewXcbeautifyRunner(log.NewLogger(), factory).Run("", []string{"test"}, nil)

	assertSingleWriter(t, opts)
	assert.Equal(t, 65, out.ExitCode)
	assert.Contains(t, string(out.RawOut), "CompileSwift")
	assert.NotContains(t, string(out.RawOut), "[Bitrise Analytics]", "Bitrise's own lines go to the console, not the raw output")
	assertReportsErrors(t, err, wantErrorLines)
	assert.NotContains(t, err.Error(), "nope", "only xcodebuild's own output is scanned for errors")
}

func TestRawXcodeCommandRunner_Run(t *testing.T) {
	var opts *command.Opts
	failure := command.NewExitStatusError("xcodebuild test", exitError(t, 65), nil)

	build := new(mocks.Command)
	build.On("PrintableCommandArgs").Return("xcodebuild test")
	build.On("RunAndReturnExitCode").Run(func(mock.Arguments) {
		_, _ = io.WriteString(opts.Stdout, xcodebuildFixture)
	}).Return(65, failure)

	factory := new(mocks.CommandFactory)
	factory.On("Create", "xcodebuild", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		opts = args.Get(2).(*command.Opts)
	}).Return(build)

	out, err := NewRawCommandRunner(log.NewLogger(), factory).Run("", []string{"test"}, nil)

	assertSingleWriter(t, opts)
	assert.Equal(t, 65, out.ExitCode)
	assert.Equal(t, xcodebuildFixture, string(out.RawOut))
	assertReportsErrors(t, err, wantErrorLines)
}

// assertSingleWriter pins the topology the runners rely on: os/exec merges xcodebuild's stdout and stderr onto one
// pipe with one copy goroutine only while cmd.Stdout == cmd.Stderr, and an ErrorFinder would make go-utils wrap the
// two separately (see attachXcodebuildErrors).
func assertSingleWriter(t *testing.T, opts *command.Opts) {
	t.Helper()
	require.NotNil(t, opts, "xcodebuild should have been created")
	assert.Nil(t, opts.ErrorFinder, "an ErrorFinder splits the streams into two writers")
	// == on purpose, assert.Equal (reflect.DeepEqual) would also accept two distinct wrappers around the same writer
	if opts.Stdout != opts.Stderr {
		t.Fatal("xcodebuild's Stdout and Stderr must be the same writer")
	}
}

func assertReportsErrors(t *testing.T, err error, lines []string) {
	t.Helper()
	require.Error(t, err)
	var exitStatusErr *command.ExitStatusError
	require.True(t, errors.As(err, &exitStatusErr), "the error should still be an ExitStatusError: %v", err)
	for _, line := range lines {
		assert.Contains(t, err.Error(), line)
	}
}

func exitError(t *testing.T, code int) *exec.ExitError {
	t.Helper()
	err := exec.Command("sh", "-c", fmt.Sprintf("exit %d", code)).Run()
	var exitErr *exec.ExitError
	require.True(t, errors.As(err, &exitErr))
	return exitErr
}
