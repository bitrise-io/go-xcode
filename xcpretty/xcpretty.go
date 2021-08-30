package xcpretty

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/bitrise-io/go-utils/env"

	"github.com/bitrise-io/go-steputils/command/rubycommand"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/xcodebuild"
	version "github.com/hashicorp/go-version"
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

	// Configure cmd in- and outputs
	pipeReader, pipeWriter := io.Pipe()

	var outBuffer bytes.Buffer
	outWriter := io.MultiWriter(&outBuffer, pipeWriter)

	xcodebuildCmd := c.xcodebuildCommand.Command(&command.Opts{
		Stdin:  nil,
		Stdout: outWriter,
		Stderr: outWriter,
	})

	prettyCmd := c.Command(&command.Opts{
		Stdin:  pipeReader,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	// Run
	if err := xcodebuildCmd.GetCmd().Start(); err != nil {
		out := outBuffer.String()
		return out, err
	}
	if err := prettyCmd.GetCmd().Start(); err != nil {
		out := outBuffer.String()
		return out, err
	}

	// Always close xcpretty outputs
	defer func() {
		if err := pipeWriter.Close(); err != nil {
			log.Warnf("Failed to close xcodebuild-xcpretty pipe, error: %s", err)
		}

		if err := prettyCmd.GetCmd().Wait(); err != nil {
			log.Warnf("xcpretty command failed, error: %s", err)
		}
	}()

	if err := xcodebuildCmd.GetCmd().Wait(); err != nil {
		out := outBuffer.String()
		return out, err
	}

	return outBuffer.String(), nil
}

// IsInstalled ...
func IsInstalled() (bool, error) {
	return rubycommand.IsGemInstalled("xcpretty", "")
}

// Install ...
func Install() ([]command.Command, error) {
	cmds, err := rubycommand.GemInstall("xcpretty", "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to create command model, error: %s", err)
	}
	return cmds, nil
}

// Version ...
func Version() (*version.Version, error) {
	cmd := command.NewFactory(env.NewRepository()).Create("xcpretty", []string{"--version"}, nil)
	versionOut, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, err
	}

	return version.NewVersion(versionOut)
}
