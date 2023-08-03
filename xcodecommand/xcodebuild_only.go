package xcodecommand

import (
	"bytes"
	"time"

	"github.com/bitrise-io/go-utils/progress"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/errorfinder"
	version "github.com/hashicorp/go-version"
)

var xcodeCommandEnvs = []string{"NSUnbufferedIO=YES"}

type rawXcodeCommand struct {
	logger         log.Logger
	commandFactory command.Factory
}

func NewRawCommandRunner(logger log.Logger, commandFactory command.Factory) Runner {
	return &rawXcodeCommand{
		logger:         logger,
		commandFactory: commandFactory,
	}
}

func (c *rawXcodeCommand) Run(workDir string, args []string, _ []string) (Output, error) {
	var (
		outBuffer bytes.Buffer
		err       error
		exitCode  int
	)

	command := c.commandFactory.Create("xcodebuild", args, &command.Opts{
		Stdout:      &outBuffer,
		Stderr:      &outBuffer,
		Env:         xcodeCommandEnvs,
		Dir:         workDir,
		ErrorFinder: errorfinder.FindXcodebuildErrors,
	})

	c.logger.TPrintf("$ %s", command.PrintableCommandArgs())

	progress.SimpleProgress(".", time.Minute, func() {
		exitCode, err = command.RunAndReturnExitCode()
	})

	return Output{
		RawOut:   outBuffer.Bytes(),
		ExitCode: exitCode,
	}, err
}

func (c *rawXcodeCommand) CheckInstall() (*version.Version, error) {
	return nil, nil
}
