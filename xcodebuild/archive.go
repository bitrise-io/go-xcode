package xcodebuild

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bitrise-io/go-utils/cmdex"
)

// ArchiveCommandModel ...
type ArchiveCommandModel struct {
	projectPath   string
	workspacePath string
	scheme        string
	configuration string

	isCleanBuild bool

	archivePath string

	forceDevelopmentTeam              string
	forceProvisioningProfileSpecifier string
	forceProvisioningProfile          string
	forceCodeSignIdentity             string

	customOptions []string
}

// NewArchiveCommandModel ...
func NewArchiveCommandModel() *ArchiveCommandModel {
	return &ArchiveCommandModel{}
}

// SetProjectPath ...
func (c *ArchiveCommandModel) SetProjectPath(projectPath string) *ArchiveCommandModel {
	c.projectPath = projectPath
	return c
}

// SetWorkspacePath ...
func (c *ArchiveCommandModel) SetWorkspacePath(workspacePath string) *ArchiveCommandModel {
	c.workspacePath = workspacePath
	return c
}

// SetScheme ...
func (c *ArchiveCommandModel) SetScheme(scheme string) *ArchiveCommandModel {
	c.scheme = scheme
	return c
}

// SetConfiguration ...
func (c *ArchiveCommandModel) SetConfiguration(configuration string) *ArchiveCommandModel {
	c.configuration = configuration
	return c
}

// SetIsCleanBuild ...
func (c *ArchiveCommandModel) SetIsCleanBuild(isCleanBuild bool) *ArchiveCommandModel {
	c.isCleanBuild = isCleanBuild
	return c
}

// SetArchivePath ...
func (c *ArchiveCommandModel) SetArchivePath(archivePath string) *ArchiveCommandModel {
	c.archivePath = archivePath
	return c
}

// SetForceDevelopmentTeam ...
func (c *ArchiveCommandModel) SetForceDevelopmentTeam(forceDevelopmentTeam string) *ArchiveCommandModel {
	c.forceDevelopmentTeam = forceDevelopmentTeam
	return c
}

// SetForceProvisioningProfileSpecifier ...
func (c *ArchiveCommandModel) SetForceProvisioningProfileSpecifier(forceProvisioningProfileSpecifier string) *ArchiveCommandModel {
	c.forceProvisioningProfileSpecifier = forceProvisioningProfileSpecifier
	return c
}

// SetForceProvisioningProfile ...
func (c *ArchiveCommandModel) SetForceProvisioningProfile(forceProvisioningProfile string) *ArchiveCommandModel {
	c.forceProvisioningProfile = forceProvisioningProfile
	return c
}

// SetForceCodeSignIdentity ...
func (c *ArchiveCommandModel) SetForceCodeSignIdentity(forceCodeSignIdentity string) *ArchiveCommandModel {
	c.forceCodeSignIdentity = forceCodeSignIdentity
	return c
}

// SetCustomOptions ...
func (c *ArchiveCommandModel) SetCustomOptions(customOptions []string) *ArchiveCommandModel {
	c.customOptions = customOptions
	return c
}

func (c *ArchiveCommandModel) cmdSlice() []string {
	slice := []string{xcodeBuildToolName}

	if c.projectPath != "" {
		slice = append(slice, "-project", c.projectPath)
	} else if c.workspacePath != "" {
		slice = append(slice, "-workspace", c.workspacePath)
	}

	if c.scheme != "" {
		slice = append(slice, "-scheme", c.scheme)
	}

	if c.configuration != "" {
		slice = append(slice, "-configuration", c.configuration)
	}

	if c.isCleanBuild {
		slice = append(slice, "clean")
	}

	slice = append(slice, "archive")

	if c.archivePath != "" {
		slice = append(slice, "-archivePath", c.archivePath)
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

	slice = append(slice, c.customOptions...)

	return slice
}

// PrintableCmd ...
func (c ArchiveCommandModel) PrintableCmd() string {
	cmdSlice := c.cmdSlice()
	return cmdex.PrintableCommandArgs(false, cmdSlice)
}

// Command ...
func (c ArchiveCommandModel) Command() *cmdex.CommandModel {
	cmdSlice := c.cmdSlice()
	return cmdex.NewCommand(cmdSlice[0], cmdSlice[1:]...)
}

// Cmd ...
func (c ArchiveCommandModel) Cmd() *exec.Cmd {
	command := c.Command()
	return command.GetCmd()
}

// Run ...
func (c ArchiveCommandModel) Run() error {
	command := c.Command()

	command.SetStdout(os.Stdout)
	command.SetStderr(os.Stderr)

	return command.Run()
}
