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

// BuildCommandModel ...
type BuildCommandModel struct {
	projectPath   string
	isWorkspace   bool
	scheme        string
	configuration string

	// buildsetting
	forceDevelopmentTeam              string
	forceProvisioningProfileSpecifier string
	forceProvisioningProfile          string
	forceCodeSignIdentity             string

	// buildaction
	customBuildActions []string

	// Options
	archivePath   string
	customOptions []string
	sdk           string

	// Archive
	isArchive bool
}

// NewBuildCommand ...
func NewBuildCommand(projectPath string, isWorkspace bool) *BuildCommandModel {
	return &BuildCommandModel{
		projectPath: projectPath,
		isWorkspace: isWorkspace,
	}
}

// SetScheme ...
func (c *BuildCommandModel) SetScheme(scheme string) *BuildCommandModel {
	c.scheme = scheme
	return c
}

// SetConfiguration ...
func (c *BuildCommandModel) SetConfiguration(configuration string) *BuildCommandModel {
	c.configuration = configuration
	return c
}

// SetForceDevelopmentTeam ...
func (c *BuildCommandModel) SetForceDevelopmentTeam(forceDevelopmentTeam string) *BuildCommandModel {
	c.forceDevelopmentTeam = forceDevelopmentTeam
	return c
}

// SetForceProvisioningProfileSpecifier ...
func (c *BuildCommandModel) SetForceProvisioningProfileSpecifier(forceProvisioningProfileSpecifier string) *BuildCommandModel {
	c.forceProvisioningProfileSpecifier = forceProvisioningProfileSpecifier
	return c
}

// SetForceProvisioningProfile ...
func (c *BuildCommandModel) SetForceProvisioningProfile(forceProvisioningProfile string) *BuildCommandModel {
	c.forceProvisioningProfile = forceProvisioningProfile
	return c
}

// SetForceCodeSignIdentity ...
func (c *BuildCommandModel) SetForceCodeSignIdentity(forceCodeSignIdentity string) *BuildCommandModel {
	c.forceCodeSignIdentity = forceCodeSignIdentity
	return c
}

// SetCustomBuildAction ...
func (c *BuildCommandModel) SetCustomBuildAction(buildAction ...string) *BuildCommandModel {
	c.customBuildActions = buildAction
	return c
}

// SetArchivePath ...
func (c *BuildCommandModel) SetArchivePath(archivePath string) *BuildCommandModel {
	c.archivePath = archivePath
	return c
}

// SetCustomOptions ...
func (c *BuildCommandModel) SetCustomOptions(customOptions []string) *BuildCommandModel {
	c.customOptions = customOptions
	return c
}

// SetSDK ...
func (c *BuildCommandModel) SetSDK(sdk string) *BuildCommandModel {
	c.sdk = sdk
	return c
}

func (c *BuildCommandModel) cmdSlice() []string {
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
	}

	slice = append(slice, c.customBuildActions...)

	if c.isArchive {
		slice = append(slice, "archive")

		if c.archivePath != "" {
			slice = append(slice, "-archivePath", c.archivePath)
		}
	}

	if c.sdk != "" {
		slice = append(slice, "-sdk", c.sdk)
	}

	slice = append(slice, c.customOptions...)

	return slice
}

// PrintableCmd ...
func (c BuildCommandModel) PrintableCmd() string {
	cmdSlice := c.cmdSlice()
	return command.PrintableCommandArgs(false, cmdSlice)
}

// Command ...
func (c BuildCommandModel) Command() *command.Model {
	cmdSlice := c.cmdSlice()
	return command.New(cmdSlice[0], cmdSlice[1:]...)
}

// Cmd ...
func (c BuildCommandModel) Cmd() *exec.Cmd {
	command := c.Command()
	return command.GetCmd()
}

// Run ...
func (c BuildCommandModel) Run() error {
	command := c.Command()

	command.SetStdout(os.Stdout)
	command.SetStderr(os.Stderr)

	return command.Run()
}
