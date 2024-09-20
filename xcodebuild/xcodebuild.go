package xcodebuild

import (
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
)

const (
	xcodebuildCmdName = "xcodebuild"
)

type Factory struct {
	cmdFactory command.Factory
}

// NewFactory ...
func NewFactory(envRepository env.Repository) Factory {
	cmdFactory := command.NewFactory(envRepository)
	return Factory{cmdFactory: cmdFactory}
}

func (factory Factory) Create(action string, options CommandOptions, settings CommandBuildSettings, opts *command.Opts) command.Command {
	args := []string{action}
	args = append(args, options.toCmdArgs()...)
	args = append(args, settings.toCmdArgs()...)
	return factory.cmdFactory.Create(xcodebuildCmdName, args, opts)
}
