package xcodebuild

import (
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
)

const (
	xcodebuildCmdName = "xcodebuild"
)

const (
	ArchiveAction            = "archive"
	TestAction               = "test"
	CleanAction              = "clean"
	ExportArchiveAction      = "-exportArchive"
	ResolvePackageDepsAction = "-resolvePackageDependencies"
	ShowBuildSettingsAction  = "-showBuildSettings"
)

type Factory struct {
	cmdFactory command.Factory
}

// NewFactory ...
func NewFactory(envRepository env.Repository) Factory {
	cmdFactory := command.NewFactory(envRepository)
	return Factory{cmdFactory: cmdFactory}
}

func (factory Factory) Create(action string, options CommandOptions, settings CommandBuildSettings, cmdOpts *command.Opts) command.Command {
	args := []string{action}
	args = append(args, options.cmdArgs()...)
	args = append(args, settings.cmdArgs()...)
	return factory.cmdFactory.Create(xcodebuildCmdName, args, cmdOpts)
}
