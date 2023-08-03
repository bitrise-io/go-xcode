package xcodecommand

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/errorfinder"
	version "github.com/hashicorp/go-version"
)

const xcbeautify = "xcbeautify"

type xcbeautifyRunner struct {
	logger         log.Logger
	commandFactory command.Factory
}

func NewXcbeautifyRunner(logger log.Logger, commandFactory command.Factory) Runner {
	return &xcbeautifyRunner{
		logger:         logger,
		commandFactory: commandFactory,
	}
}

func (c *xcbeautifyRunner) Run(workDir string, xcodebuildArgs []string, xcbeautifyArgs []string) (Output, error) {
	var (
		buildOutBuffer         bytes.Buffer
		pipeReader, pipeWriter = io.Pipe()
		buildOutWriter         = io.MultiWriter(&buildOutBuffer, pipeWriter)
	)

	// For parallel and concurrent destination testing, it helps to use unbuffered I/O for stdout and to redirect stderr to stdout.
	// NSUnbufferedIO=YES xcodebuild [args] 2>&1 | xcbeautify
	buildCmd := c.commandFactory.Create("xcodebuild", xcodebuildArgs, &command.Opts{
		Stdout:      buildOutWriter,
		Stderr:      buildOutWriter,
		Env:         xcodeCommandEnvs,
		Dir:         workDir,
		ErrorFinder: errorfinder.FindXcodebuildErrors,
	})

	beautifyCmd := c.commandFactory.Create(xcbeautify, xcbeautifyArgs, &command.Opts{
		Stdin:  pipeReader,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	defer func() {
		if err := pipeWriter.Close(); err != nil {
			c.logger.Warnf("Failed to close xcodebuild-xcbeautify pipe: %s", err)
		}

		if err := beautifyCmd.Wait(); err != nil {
			c.logger.Warnf("xcbeautify command failed: %s", err)
		}
	}()

	c.logger.TPrintf("$ set -o pipefail && %s | %s", buildCmd.PrintableCommandArgs(), beautifyCmd.PrintableCommandArgs())

	err := buildCmd.Start()
	if err == nil {
		err = beautifyCmd.Start()
	}
	if err == nil {
		err = buildCmd.Wait()
	}

	exitCode := 0
	if err != nil {
		exitCode = -1

		var exerr *exec.ExitError
		if errors.As(err, &exerr) {
			exitCode = exerr.ExitCode()
		}
	}

	return Output{
		RawOut:   buildOutBuffer.Bytes(),
		ExitCode: exitCode,
	}, err
}

func (c *xcbeautifyRunner) CheckInstall() (*version.Version, error) {
	c.logger.Println()
	c.logger.Infof("Checking log formatter (xcbeautify) version")

	versionCmd := c.commandFactory.Create(xcbeautify, []string{"--version"}, nil)

	out, err := versionCmd.RunAndReturnTrimmedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return nil, fmt.Errorf("xcbeautify version command failed: %w", err)
		}

		return nil, fmt.Errorf("failed to run xcbeautify command: %w", err)
	}

	return version.NewVersion(out)
}
