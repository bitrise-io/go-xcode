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

func (sel *FallbackRunner) CheckInstall() (*version.Version, error) {
	ver, err := sel.runner.CheckInstall()
	if err == nil {
		return ver, nil
	}

	sel.logger.Errorf("Checking log formatter failed: %s", err)
	sel.logger.Infof("Falling back to xcodebuild log formatter")
	sel.runner = sel.fallbackRunner

	return sel.runner.CheckInstall()
}

func (sel *FallbackRunner) Run(workDir string, xcodebuildArgs []string, xcbeautifyArgs []string) (Output, error) {
	return sel.runner.Run(workDir, xcodebuildArgs, xcbeautifyArgs)
}
