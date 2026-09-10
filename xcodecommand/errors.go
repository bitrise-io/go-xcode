package xcodecommand

import (
	"errors"
	"os/exec"

	"github.com/bitrise-io/go-utils/v2/command"
)

// attachXcodebuildErrors re-wraps a failed xcodebuild command's error with the error lines an errorfinder.Finder found
// in its output, as go-utils/command does when given an ErrorFinder. The lines come in the finder's order (regular
// "error:" lines first, "xcodebuild: error:" blocks last) rather than in output order.
//
// The runners must not use command.Opts.ErrorFinder: with it set, go-utils wraps Stdout and Stderr in two separate
// io.MultiWriters. os/exec merges the two streams onto one pipe and one copy goroutine only while
// cmd.Stdout == cmd.Stderr, so the wrapping hands the shared, single-writer output (PrefixFilter, bytes.Buffer) to
// two goroutines and corrupts it. A Finder teed off that single shared writer sees the complete output of both
// streams, like the go-utils collector did, and assembles the lines itself.
func attachXcodebuildErrors(err error, printableCmdArgs string, errorLines []string) error {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return err
	}

	return command.NewExitStatusError(printableCmdArgs, exitErr, errorLines)
}
