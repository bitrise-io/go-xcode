package xcodebuild

import (
	"bufio"
	"strings"

	"github.com/bitrise-io/go-utils/command"
)

// ShowBuildSettingsCommandModel ...
type ShowBuildSettingsCommandModel struct {
	commandFactory command.Factory

	projectPath string
	isWorkspace bool
}

// NewShowBuildSettingsCommand ...
func NewShowBuildSettingsCommand(projectPath string, isWorkspace bool, commandFactory command.Factory) *ShowBuildSettingsCommandModel {
	return &ShowBuildSettingsCommandModel{
		commandFactory: commandFactory,
		projectPath:    projectPath,
		isWorkspace:    isWorkspace,
	}
}

func (c *ShowBuildSettingsCommandModel) args() []string {
	var slice []string

	if c.projectPath != "" {
		if c.isWorkspace {
			slice = append(slice, "-workspace", c.projectPath)
		} else {
			slice = append(slice, "-project", c.projectPath)
		}
	}

	return slice
}

// Command ...
func (c ShowBuildSettingsCommandModel) Command(opts *command.Opts) command.Command {
	return c.commandFactory.Create(toolName, c.args(), opts)
}

// PrintableCmd ...
func (c ShowBuildSettingsCommandModel) PrintableCmd() string {
	return c.Command(nil).PrintableCommandArgs()
}

func parseBuildSettings(out string) (map[string]string, error) {
	settings := map[string]string{}

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if split := strings.Split(line, "="); len(split) == 2 {
			key := strings.TrimSpace(split[0])
			value := strings.TrimSpace(split[1])
			value = strings.Trim(value, `"`)

			settings[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return map[string]string{}, err
	}

	return settings, nil
}

// RunAndReturnSettings ...
func (c ShowBuildSettingsCommandModel) RunAndReturnSettings() (map[string]string, error) {
	command := c.Command(nil)
	out, err := command.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return map[string]string{}, err
	}

	return parseBuildSettings(out)
}
