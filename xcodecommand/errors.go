package xcodecommand

import (
	"errors"
	"os/exec"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-xcode/v2/errorfinder"
)

// attachXcodebuildErrors re-wraps a failed xcodebuild command's error with the error lines found in its output,
// as go-utils/command does when given an ErrorFinder. The lines come in FindXcodebuildErrors' order (regular
// "error:" lines first, "xcodebuild: error:" blocks last) rather than in output order.
//
// The runners must not use command.Opts.ErrorFinder: with it set, go-utils wraps Stdout and Stderr in two separate
// io.MultiWriters. os/exec merges the two streams onto one pipe and one copy goroutine only while
// cmd.Stdout == cmd.Stderr, so the wrapping hands the shared, single-writer output (PrefixFilter, bytes.Buffer) to
// two goroutines and corrupts it. Collecting the errors from the output afterwards keeps the single writer, and
// sees whole lines instead of arbitrary chunks. With xcbeautify and xcpretty that output is the filtered raw log,
// so a `[Bitrise ...]` prefixed line is not scanned.
func attachXcodebuildErrors(err error, printableCmdArgs string, rawOut []byte) error {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return err
	}

	return command.NewExitStatusError(printableCmdArgs, exitErr, errorfinder.FindXcodebuildErrors(string(rawOut)))
}
