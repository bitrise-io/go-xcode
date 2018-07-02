package xcodebuild

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bitrise-io/go-utils/command"
)

/*
xcodebuild [-project <projectname>] \
	-scheme <schemeName> \
	[-destination <destinationspecifier>]... \
	[-configuration <configurationname>] \
	[-arch <architecture>]... \
	[-sdk [<sdkname>|<sdkpath>]] \
	[-showBuildSettings] \
	[<buildsetting>=<value>]... \
	[<buildaction>]...
xcodebuild -workspace <workspacename> \
	-scheme <schemeName> \
	[-destination <destinationspecifier>]... \
	[-configuration <configurationname>] \
	[-arch <architecture>]... \
	[-sdk [<sdkname>|<sdkpath>]] \
	[-showBuildSettings] \
	[<buildsetting>=<value>]... \
	[<buildaction>]...
*/

// const ...
const (
	ArchiveAction Action = "archiveAction"
	BuildAction   Action = "buildAction"
	AnalyzeAction Action = "analyzeAction"
)

// Action ...
type Action string

// Command ...
type Command struct {
	projectPath   string
	isWorkspace   bool
	scheme        string
	configuration string

	// buildsetting
	forceDevelopmentTeam              string
	forceProvisioningProfileSpecifier string
	forceProvisioningProfile          string
	forceCodeSignIdentity             string
	disableCodesign                   bool

	// buildaction
	customBuildActions []string

	// Options
	archivePath   string
	customOptions []string
	sdk           string

	// Archive
	action Action
}

// NewCommand ...
func NewCommand(projectPath string, isWorkspace bool, action Action) *Command {
	return &Command{
		projectPath: projectPath,
		isWorkspace: isWorkspace,
		action:      action,
	}
}

// SetScheme ...
func (c *Command) SetScheme(scheme string) *Command {
	c.scheme = scheme
	return c
}

// SetConfiguration ...
func (c *Command) SetConfiguration(configuration string) *Command {
	c.configuration = configuration
	return c
}

// SetForceDevelopmentTeam ...
func (c *Command) SetForceDevelopmentTeam(forceDevelopmentTeam string) *Command {
	c.forceDevelopmentTeam = forceDevelopmentTeam
	return c
}

// SetForceProvisioningProfileSpecifier ...
func (c *Command) SetForceProvisioningProfileSpecifier(forceProvisioningProfileSpecifier string) *Command {
	c.forceProvisioningProfileSpecifier = forceProvisioningProfileSpecifier
	return c
}

// SetForceProvisioningProfile ...
func (c *Command) SetForceProvisioningProfile(forceProvisioningProfile string) *Command {
	c.forceProvisioningProfile = forceProvisioningProfile
	return c
}

// SetForceCodeSignIdentity ...
func (c *Command) SetForceCodeSignIdentity(forceCodeSignIdentity string) *Command {
	c.forceCodeSignIdentity = forceCodeSignIdentity
	return c
}

// SetCustomBuildAction ...
func (c *Command) SetCustomBuildAction(buildAction ...string) *Command {
	c.customBuildActions = buildAction
	return c
}

// SetArchivePath ...
func (c *Command) SetArchivePath(archivePath string) *Command {
	c.archivePath = archivePath
	return c
}

// SetCustomOptions ...
func (c *Command) SetCustomOptions(customOptions []string) *Command {
	c.customOptions = customOptions
	return c
}

// SetSDK ...
func (c *Command) SetSDK(sdk string) *Command {
	c.sdk = sdk
	return c
}

// SetDisableCodesign ...
func (c *Command) SetDisableCodesign(disable bool) *Command {
	c.disableCodesign = disable
	return c
}

func (c *Command) cmdSlice() []string {
	slice := []string{toolName}

	if c.projectPath != "" {
		if c.isWorkspace {
			slice = append(slice, "-workspace", c.projectPath)
		} else {
			slice = append(slice, "-project", c.projectPath)
		}
	}

	if c.scheme != "" {
		slice = append(slice, "-scheme", c.scheme)
	}
	if c.configuration != "" {
		slice = append(slice, "-configuration", c.configuration)
	}

	if c.forceDevelopmentTeam != "" {
		slice = append(slice, fmt.Sprintf("DEVELOPMENT_TEAM=%s", c.forceDevelopmentTeam))
	}
	if c.forceProvisioningProfileSpecifier != "" {
		slice = append(slice, fmt.Sprintf("PROVISIONING_PROFILE_SPECIFIER=%s", c.forceProvisioningProfileSpecifier))
	}
	if c.forceProvisioningProfile != "" {
		slice = append(slice, fmt.Sprintf("PROVISIONING_PROFILE=%s", c.forceProvisioningProfile))
	}
	if c.forceCodeSignIdentity != "" {
		slice = append(slice, fmt.Sprintf("CODE_SIGN_IDENTITY=%s", c.forceCodeSignIdentity))
	} else if c.disableCodesign {
		slice = append(slice, "CODE_SIGN_IDENTITY=")
		slice = append(slice, "CODE_SIGNING_REQUIRED=NO")
	}

	slice = append(slice, c.customBuildActions...)

	switch c.action {
	case ArchiveAction:
		slice = append(slice, "archive")

		if c.archivePath != "" {
			slice = append(slice, "-archivePath", c.archivePath)
		}
	case BuildAction:
		slice = append(slice, "build")
	case AnalyzeAction:
		slice = append(slice, "analyze")
	}

	if c.sdk != "" {
		slice = append(slice, "-sdk", c.sdk)
	}

	slice = append(slice, c.customOptions...)

	return slice
}

// PrintableCmd ...
func (c Command) PrintableCmd() string {
	cmdSlice := c.cmdSlice()
	return command.PrintableCommandArgs(false, cmdSlice)
}

// ExecCommand ...
func (c Command) ExecCommand() *command.Model {
	cmdSlice := c.cmdSlice()
	return command.New(cmdSlice[0], cmdSlice[1:]...)
}

// Cmd ...
func (c Command) Cmd() *exec.Cmd {
	command := c.ExecCommand()
	return command.GetCmd()
}

// Run ...
func (c Command) Run() error {
	command := c.ExecCommand()

	command.SetStdout(os.Stdout)
	command.SetStderr(os.Stderr)

	return command.Run()
}
