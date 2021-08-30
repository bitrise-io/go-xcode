package xcodebuild

import (
	"os"
	"os/exec"

	"github.com/bitrise-io/go-utils/env"

	"github.com/bitrise-io/go-utils/command"
)

/*
xcodebuild -exportArchive \
	-exportFormat format \
	-archivePath xcarchivepath \
    -exportPath destinationpath \
    [-exportProvisioningProfile profilename] \
	[-exportSigningIdentity identityname] \
	[-exportInstallerIdentity identityname]
*/

// LegacyExportCommandModel ...
type LegacyExportCommandModel struct {
	exportFormat                  string
	archivePath                   string
	exportPath                    string
	exportProvisioningProfileName string
}

// NewLegacyExportCommand ...
func NewLegacyExportCommand() *LegacyExportCommandModel {
	return &LegacyExportCommandModel{}
}

// SetExportFormat ...
func (c *LegacyExportCommandModel) SetExportFormat(exportFormat string) *LegacyExportCommandModel {
	c.exportFormat = exportFormat
	return c
}

// SetArchivePath ...
func (c *LegacyExportCommandModel) SetArchivePath(archivePath string) *LegacyExportCommandModel {
	c.archivePath = archivePath
	return c
}

// SetExportPath ...
func (c *LegacyExportCommandModel) SetExportPath(exportPath string) *LegacyExportCommandModel {
	c.exportPath = exportPath
	return c
}

// SetExportProvisioningProfileName ...
func (c *LegacyExportCommandModel) SetExportProvisioningProfileName(exportProvisioningProfileName string) *LegacyExportCommandModel {
	c.exportProvisioningProfileName = exportProvisioningProfileName
	return c
}

func (c LegacyExportCommandModel) args() []string {
	slice := []string{"-exportArchive"}
	if c.exportFormat != "" {
		slice = append(slice, "-exportFormat", c.exportFormat)
	}
	if c.archivePath != "" {
		slice = append(slice, "-archivePath", c.archivePath)
	}
	if c.exportPath != "" {
		slice = append(slice, "-exportPath", c.exportPath)
	}
	if c.exportProvisioningProfileName != "" {
		slice = append(slice, "-exportProvisioningProfile", c.exportProvisioningProfileName)
	}
	return slice
}

// Command ...
func (c LegacyExportCommandModel) Command(opts *command.Opts) command.Command {
	f := command.NewFactory(env.NewRepository())
	return f.Create(toolName, c.args(), opts)
}

// PrintableCmd ...
func (c LegacyExportCommandModel) PrintableCmd() string {
	return c.Command(nil).PrintableCommandArgs()
}

// Cmd ...
func (c LegacyExportCommandModel) Cmd(opts *command.Opts) *exec.Cmd {
	command := c.Command(opts)
	return command.GetCmd()
}

// Run ...
func (c LegacyExportCommandModel) Run() error {
	command := c.Command(&command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	return command.Run()
}
