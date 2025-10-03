package xcpretty

import (
	"fmt"
	"regexp"

	"github.com/bitrise-io/go-steputils/v2/ruby"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/logio"
	"github.com/bitrise-io/go-xcode/v2/xcodebuild"
	"github.com/hashicorp/go-version"
)

const (
	toolName = "xcpretty"
)

// CommandModel ...
type CommandModel struct {
	xcodebuildCommand xcodebuild.CommandModel
	customOptions     []string
}

// New ...
func New(xcodebuildCommand xcodebuild.CommandModel) *CommandModel {
	return &CommandModel{
		xcodebuildCommand: xcodebuildCommand,
	}
}

// SetCustomOptions ...
func (c *CommandModel) SetCustomOptions(customOptions []string) *CommandModel {
	c.customOptions = customOptions
	return c
}

// Command ...
func (c CommandModel) Command(opts *command.Opts) command.Command {
	return command.NewFactory(env.NewRepository()).Create(toolName, c.customOptions, opts)
}

// PrintableCmd ...
func (c CommandModel) PrintableCmd() string {
	prettyCmdStr := c.Command(nil).PrintableCommandArgs()
	xcodebuildCmdStr := c.xcodebuildCommand.PrintableCmd()

	return fmt.Sprintf("set -o pipefail && %s | %s", xcodebuildCmdStr, prettyCmdStr)
}

// Run ...
func (c CommandModel) Run() (string, error) {
	loggingIO := logio.SetupPipeWiring(regexp.MustCompile(`^\[Bitrise.*\].*`))

	xcodebuildCmd := c.xcodebuildCommand.Command(&command.Opts{
		Stdin:  nil,
		Stdout: loggingIO.XcbuildStdout,
		Stderr: loggingIO.XcbuildStderr,
	})

	prettyCmd := c.Command(&command.Opts{
		Stdin:  loggingIO.ToolStdin,
		Stdout: loggingIO.ToolStdout,
		Stderr: loggingIO.ToolStderr,
	})

	// Always close xcpretty outputs
	defer func() {
		if err := loggingIO.CloseFilter(); err != nil {
			fmt.Printf("logging IO failure, error: %s", err)
		}

		if err := loggingIO.CloseToolInput(); err != nil {
			fmt.Printf("logging IO failure, error: %s", err)
		}

		if err := prettyCmd.Wait(); err != nil {
			fmt.Printf("xcpretty command failed, error: %s", err)
		}
	}()

	// Run
	if err := xcodebuildCmd.Start(); err != nil {
		out := loggingIO.XcbuildRawout.String()
		return out, err
	}
	if err := prettyCmd.Start(); err != nil {
		out := loggingIO.XcbuildRawout.String()
		return out, err
	}

	if err := xcodebuildCmd.Wait(); err != nil {
		out := loggingIO.XcbuildRawout.String()
		return out, err
	}

	return loggingIO.XcbuildRawout.String(), nil
}

// Xcpretty ...
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
	factory, err := ruby.NewCommandFactory(command.NewFactory(env.NewRepository()), locator)
	if err != nil {
		return false, err
	}

	return ruby.NewEnvironment(factory, locator, x.logger).IsGemInstalled("xcpretty", "")
}

// Install ...
func (x xcpretty) Install() ([]command.Command, error) {
	locator := env.NewCommandLocator()
	factory, err := ruby.NewCommandFactory(command.NewFactory(env.NewRepository()), locator)
	if err != nil {
		return nil, err
	}

	cmds := factory.CreateGemInstall("xcpretty", "", false, false, nil)

	return cmds, nil
}

// Version ...
func (x xcpretty) Version() (*version.Version, error) {
	cmd := command.NewFactory(env.NewRepository()).Create("xcpretty", []string{"--version"}, nil)
	versionOut, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, err
	}

	return version.NewVersion(versionOut)
}
