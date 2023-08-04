package xcodecommand

import (
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	version "github.com/hashicorp/go-version"
)

// FallbackRunner falls back to the raw xcodebuild runner in case xcbeautify (or other) is unavaialble
type FallbackRunner struct {
	runner         Runner
	fallbackRunner Runner
	logger         log.Logger
}

// NewFallbackRunner wraps an xcbeatufy (or other) runner
func NewFallbackRunner(runner Runner, logger log.Logger, commandFactory command.Factory) *FallbackRunner {
	return &FallbackRunner{
		runner:         runner,
		fallbackRunner: NewRawCommandRunner(logger, commandFactory),
		logger:         logger,
	}
}

// CheckInstall checks if the wrapped Runner is available,
// if not changes the current runner to the raw xcodebuild runner
func (f *FallbackRunner) CheckInstall() (*version.Version, error) {
	if f.runner == nil || f.fallbackRunner == nil {
		panic("runner or fallback runner is nil")
	}

	ver, err := f.runner.CheckInstall()
	if err == nil {
		return ver, nil
	}

	f.logger.Errorf("Selected log formatter is unavailable: %s", err)
	f.logger.Infof("Switching back to xcodebuild log formatter.")
	f.runner = f.fallbackRunner

	return f.runner.CheckInstall()
}

// Run runs the current runner.
// An earlier call to CheckInstall may have changed the current runner.
func (f *FallbackRunner) Run(workDir string, xcodebuildArgs []string, xcbeautifyArgs []string) (Output, error) {
	return f.runner.Run(workDir, xcodebuildArgs, xcbeautifyArgs)
}
