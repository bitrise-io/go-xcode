// Package xcpretty installs and inspects the xcpretty log formatter.
package xcpretty

import (
	"github.com/bitrise-io/go-steputils/v2/ruby"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/hashicorp/go-version"
)

const (
	toolName = "xcpretty"
)

// Xcpretty installs and inspects the xcpretty gem. To run xcodebuild through xcpretty use
// xcodecommand.NewXcprettyCommandRunner.
type Xcpretty interface {
	IsInstalled() (bool, error)
	Install() ([]command.Command, error)
	Version() (*version.Version, error)
}

type xcpretty struct {
	logger log.Logger
}

// NewXcpretty ...
func NewXcpretty(logger log.Logger) Xcpretty {
	return &xcpretty{
		logger: logger,
	}
}

// IsInstalled ...
func (x xcpretty) IsInstalled() (bool, error) {
	locator := env.NewCommandLocator()
	factory, err := ruby.NewCommandFactory(command.NewFactory(env.NewRepository()), locator, x.logger)
	if err != nil {
		return false, err
	}

	return ruby.NewEnvironment(factory, locator, x.logger).IsGemInstalled(toolName, "")
}

// Install ...
func (x xcpretty) Install() ([]command.Command, error) {
	locator := env.NewCommandLocator()
	factory, err := ruby.NewCommandFactory(command.NewFactory(env.NewRepository()), locator, x.logger)
	if err != nil {
		return nil, err
	}

	cmds := factory.CreateGemInstall(toolName, "", false, false, nil)

	return cmds, nil
}

// Version ...
func (x xcpretty) Version() (*version.Version, error) {
	cmd := command.NewFactory(env.NewRepository()).Create(toolName, []string{"--version"}, nil)
	versionOut, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, err
	}

	return version.NewVersion(versionOut)
}
