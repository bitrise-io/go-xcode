package xcodebuild

import (
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
)

const (
	xcodebuildCmdName = "xcodebuild"
)

const (
	BuildAction               = "build"
	BuildForTest              = "build-for-testing"
	AnalyzeAction             = "analyze"
	ArchiveAction             = "archive"
	TestAction                = "test"
	TestWithoutBuildingAction = "test-without-building"
	DocBuildAction            = "docbuild"
	InstallSrcAction          = "installsrc"
	InstallAction             = "install"
	CleanAction               = "clean"
)

type Factory struct {
	cmdFactory     command.Factory
	additionalArgs []string
}

// NewFactory ...
func NewFactory(envRepository env.Repository, additionalArgs []string) Factory {
	cmdFactory := command.NewFactory(envRepository)
	return Factory{cmdFactory: cmdFactory}
}

func (factory Factory) CreateWithoutDefaultAdditionalArgs(options *CommandOptions, actions []string, buildSettings *CommandBuildSettings, additionalArgs []string, cmdOpts *command.Opts) command.Command {
	return factory.Create(options, actions, buildSettings, additionalArgs, cmdOpts)
}

func (factory Factory) Create(options *CommandOptions, actions []string, buildSettings *CommandBuildSettings, additionalArgs []string, cmdOpts *command.Opts) command.Command {
	defaultAdditionalArgsResult := parseAdditionalArgs(factory.additionalArgs)
	additionalArgsResult := parseAdditionalArgs(additionalArgs)
	mergedAdditionalArgsResult := mergeAdditionalArgs(defaultAdditionalArgsResult, additionalArgsResult)
	mergedAdditionalArgs := mergedAdditionalArgsResult.toArgs()

	return factory.Create(options, actions, buildSettings, mergedAdditionalArgs, cmdOpts)
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
