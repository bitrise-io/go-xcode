package xcodecommand

import (
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	version "github.com/hashicorp/go-version"
)

type FallbackRunner struct {
	runner         Runner
	fallbackRunner Runner
	logger         log.Logger
}

func NewFallbackRunner(runner Runner, logger log.Logger, commandFactory command.Factory) *FallbackRunner {
	return &FallbackRunner{
		runner:         runner,
		fallbackRunner: NewRawCommandRunner(logger, commandFactory),
		logger:         logger,
	}
}

func (f *FallbackRunner) CheckInstall() (*version.Version, error) {
	ver, err := f.runner.CheckInstall()
	if err == nil {
		return ver, nil
	}

	f.logger.Errorf("Checking log formatter failed: %s", err)
	f.logger.Infof("Falling back to xcodebuild log formatter")
	f.runner = f.fallbackRunner

	return f.runner.CheckInstall()
}

func (f *FallbackRunner) Run(workDir string, xcodebuildArgs []string, xcbeautifyArgs []string) (Output, error) {
	return f.runner.Run(workDir, xcodebuildArgs, xcbeautifyArgs)
}
