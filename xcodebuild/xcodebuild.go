package xcodebuild

import (
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
)

const (
	xcodebuildCmdName = "xcodebuild"
)

// Factory creates xcodebuild commands.
type Factory struct {
	cmdFactory     command.Factory
	additionalArgs []string
}

// NewFactory creates a new Factory.
func NewFactory(envRepository env.Repository) Factory {
	cmdFactory := command.NewFactory(envRepository)
	return Factory{
		cmdFactory: cmdFactory,
	}
}

// NewFactoryWithAdditionalArgs creates a new Factory with additional arguments to be carried over for all the xcodebuild commands created.
func NewFactoryWithAdditionalArgs(envRepository env.Repository, additionalArgs []string) Factory {
	cmdFactory := command.NewFactory(envRepository)
	return Factory{
		cmdFactory:     cmdFactory,
		additionalArgs: additionalArgs,
	}
}

// CreateWithoutDefaultAdditionalArgs creates a new xcodebuild command without the default additional arguments.
func (factory Factory) CreateWithoutDefaultAdditionalArgs(options *CommandOptions, actions []string, buildSettings *CommandBuildSettings, additionalArgs []string, cmdOpts *command.Opts) command.Command {
	return factory.create(options, actions, buildSettings, additionalArgs, cmdOpts)
}

// Create creates a new xcodebuild command.
func (factory Factory) Create(options *CommandOptions, actions []string, buildSettings *CommandBuildSettings, additionalArgs []string, cmdOpts *command.Opts) command.Command {
	defaultAdditionalArgsResult := ParseAdditionalArgs(factory.additionalArgs)
	additionalArgsResult := ParseAdditionalArgs(additionalArgs)
	mergedAdditionalArgsResult := MergeAdditionalArgs(defaultAdditionalArgsResult, additionalArgsResult)
	mergedAdditionalArgs := mergedAdditionalArgsResult.ToArgs()

	return factory.create(options, actions, buildSettings, mergedAdditionalArgs, cmdOpts)
}

func (factory Factory) create(options *CommandOptions, actions []string, buildSettings *CommandBuildSettings, additionalArgs []string, cmdOpts *command.Opts) command.Command {
	var args []string
	if options != nil {
		args = append(args, options.cmdArgs()...)
	}
	args = append(args, actions...)
	if buildSettings != nil {
		args = append(args, buildSettings.cmdArgs()...)
	}
	args = append(args, additionalArgs...)

	return factory.cmdFactory.Create(xcodebuildCmdName, args, cmdOpts)
}
