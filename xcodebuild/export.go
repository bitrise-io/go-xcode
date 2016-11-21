package xcodebuild

import (
	"os"
	"os/exec"

	"github.com/bitrise-io/go-utils/cmdex"
)

// ExportCommandModel ...
type ExportCommandModel struct {
	archivePath        string
	exportDir          string
	exportOptionsPlist string
}

// NewExportCommand ...
func NewExportCommand() *ExportCommandModel {
	return &ExportCommandModel{}
}

// SetArchivePath ...
func (c *ExportCommandModel) SetArchivePath(archivePath string) *ExportCommandModel {
	c.archivePath = archivePath
	return c
}

// SetExportDir ...
func (c *ExportCommandModel) SetExportDir(exportDir string) *ExportCommandModel {
	c.exportDir = exportDir
	return c
}

// SetExportOptionsPlist ...
func (c *ExportCommandModel) SetExportOptionsPlist(exportOptionsPlist string) *ExportCommandModel {
	c.exportOptionsPlist = exportOptionsPlist
	return c
}

func (c ExportCommandModel) cmdSlice() []string {
	slice := []string{xcodeBuildToolName}
	slice = append(slice, "-exportArchive")
	if c.archivePath != "" {
		slice = append(slice, "-archivePath", c.archivePath)
	}
	if c.exportDir != "" {
		slice = append(slice, "-exportPath", c.exportDir)
	}
	if c.exportOptionsPlist != "" {
		slice = append(slice, "-exportOptionsPlist", c.exportOptionsPlist)
	}
	return slice
}

// PrintableCmd ...
func (c ExportCommandModel) PrintableCmd() string {
	cmdSlice := c.cmdSlice()
	return cmdex.PrintableCommandArgs(false, cmdSlice)
}

// Command ...
func (c ExportCommandModel) Command() *cmdex.CommandModel {
	cmdSlice := c.cmdSlice()
	return cmdex.NewCommand(cmdSlice[0], cmdSlice[1:]...)
}

// Cmd ...
func (c ExportCommandModel) Cmd() *exec.Cmd {
	command := c.Command()
	return command.GetCmd()
}

// Run ...
func (c ExportCommandModel) Run() error {
	command := c.Command()

	command.SetStdout(os.Stdout)
	command.SetStderr(os.Stderr)

	return command.Run()
}
