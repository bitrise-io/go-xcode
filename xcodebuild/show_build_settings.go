package xcodebuild

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
)

const (
	toolName = "xcodebuild"

	showBuildSettingsFlag = "-showBuildSettings"
	projectFlag           = "-project"
	workspaceFlag         = "-workspace"
	targetFlag            = "-target"
	schemeFlag            = "-scheme"
	configurationFlag     = "-configuration"

	xcworkspaceExtension = ".xcworkspace"
)

// BuildSettingsProvider returns effective build settings using xcodebuild -showBuildSettings.
type BuildSettingsProvider interface {
	TargetBuildSettings(projectPath, target, configuration string, extraArgs ...string) (serialized.Object, error)
	SchemeBuildSettings(projectPath, scheme, configuration string, extraArgs ...string) (serialized.Object, error)
}

type showBuildSettingsProvider struct {
	commandFactory command.Factory
	logger         log.Logger
}

// NewShowBuildSettingsProvider returns a BuildSettingsProvider backed by xcodebuild.
func NewShowBuildSettingsProvider(commandFactory command.Factory, logger log.Logger) BuildSettingsProvider {
	return showBuildSettingsProvider{
		commandFactory: commandFactory,
		logger:         logger,
	}
}

// TargetBuildSettings returns the effective build settings of one target.
func (p showBuildSettingsProvider) TargetBuildSettings(projectPath, target, configuration string, extraArgs ...string) (serialized.Object, error) {
	return p.run(showBuildSettingsArgs(projectPath, targetFlag, target, configuration, extraArgs))
}

// SchemeBuildSettings returns the effective build settings of one scheme.
func (p showBuildSettingsProvider) SchemeBuildSettings(projectPath, scheme, configuration string, extraArgs ...string) (serialized.Object, error) {
	return p.run(showBuildSettingsArgs(projectPath, schemeFlag, scheme, configuration, extraArgs))
}

func (p showBuildSettingsProvider) run(args []string) (serialized.Object, error) {
	cmd := p.commandFactory.Create(toolName, args, nil)

	// Logged at normal level, as in v1, so the command shows up in step logs.
	p.logger.TPrintf("Reading build settings...")
	p.logger.TDonef("$ %s", cmd.PrintableCommandArgs())

	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// The output explains the failure better than the exit status.
			return nil, fmt.Errorf("%s failed: %s: %w", cmd.PrintableCommandArgs(), out, err)
		}
		return nil, fmt.Errorf("failed to run %s: %w", cmd.PrintableCommandArgs(), err)
	}

	p.logger.TPrintf("Read target settings.")

	return parseShowBuildSettingsOutput(out)
}

func showBuildSettingsArgs(projectPath, nameFlag, name, configuration string, extraArgs []string) []string {
	var args []string

	if projectPath != "" {
		containerFlag := projectFlag
		if filepath.Ext(projectPath) == xcworkspaceExtension {
			containerFlag = workspaceFlag
		}
		args = append(args, containerFlag, projectPath)
	}

	if name != "" {
		args = append(args, nameFlag, name)
	}

	if configuration != "" {
		args = append(args, configurationFlag, configuration)
	}

	args = append(args, showBuildSettingsFlag)

	return append(args, extraArgs...)
}

// parseShowBuildSettingsOutput keeps the first occurrence of a repeated key, which for a
// multi-target scheme is the main target (v1 fix 0c84f25). ReadLine is used because values can
// exceed bufio.Scanner's line limit.
func parseShowBuildSettingsOutput(out string) (serialized.Object, error) {
	settings := serialized.Object{}

	reader := bufio.NewReader(strings.NewReader(out))
	var buffer bytes.Buffer

	for {
		fragment, isPrefix, err := reader.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		buffer.Write(fragment)

		if isPrefix {
			continue
		}

		line := strings.TrimSpace(buffer.String())
		buffer.Reset()

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"`)

		if _, exists := settings[key]; !exists {
			settings[key] = value
		}
	}

	return settings, nil
}
