package xcodecommand

import (
	"bytes"
	"time"

	"github.com/bitrise-io/go-utils/progress"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	version "github.com/hashicorp/go-version"
)

var unbufferedIOEnv = []string{"NSUnbufferedIO=YES"}

// RawXcodeCommandRunner is an xcodebuild runner that uses no additional log formatter
type RawXcodeCommandRunner struct {
	logger         log.Logger
	commandFactory command.Factory
}

// NewRawCommandRunner creates a new RawXcodeCommandRunner
func NewRawCommandRunner(logger log.Logger, commandFactory command.Factory) Runner {
	return &RawXcodeCommandRunner{
		logger:         logger,
		commandFactory: commandFactory,
	}
}

// Run runs xcodebuild using no additional log formatter
func (c *RawXcodeCommandRunner) Run(workDir string, args []string, _ []string) (Output, error) {
	var (
		outBuffer bytes.Buffer
		err       error
		exitCode  int
	)

	// Stdout and Stderr share one buffer on purpose: os/exec then uses a single pipe and copy goroutine, so the
	// (not goroutine-safe) bytes.Buffer has one writer. No ErrorFinder for the same reason, see attachXcodebuildErrors.
	command := c.commandFactory.Create("xcodebuild", args, &command.Opts{
		Stdout: &outBuffer,
		Stderr: &outBuffer,
		Env:    unbufferedIOEnv,
		Dir:    workDir,
	})

	c.logger.TPrintf("$ %s", command.PrintableCommandArgs())

	progress.SimpleProgress(".", time.Minute, func() {
		exitCode, err = command.RunAndReturnExitCode()
	})

	if err != nil {
		err = attachXcodebuildErrors(err, command.PrintableCommandArgs(), outBuffer.Bytes())
	}

	return Output{
		RawOut:   outBuffer.Bytes(),
		ExitCode: exitCode,
	}, err
}

// CheckInstall does nothing as no additional log formatter is used
func (c *RawXcodeCommandRunner) CheckInstall() (*version.Version, error) {
	return nil, nil
}
