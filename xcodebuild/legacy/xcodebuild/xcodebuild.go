package xcodebuild

import "github.com/bitrise-io/go-utils/command"

const (
	toolName = "xcodebuild"
)

// CommandModel ...
type CommandModel interface {
	PrintableCmd() string
	Command() *command.Model
}

// AuthenticationParams are used to authenticate to App Store Connect API and let xcodebuild download missing provisioning profiles.
type AuthenticationParams struct {
	KeyID     string
	IsssuerID string
	KeyPath   string
}

func (a *AuthenticationParams) args() []string {
	return []string{
		"-allowProvisioningUpdates",
		"-authenticationKeyPath", a.KeyPath,
		"-authenticationKeyID", a.KeyID,
		"-authenticationKeyIssuerID", a.IsssuerID,
	}
}

/*
Used functions:
- xcodebuild.CommandBuilder
	- .Command()
	- SetScheme
	- SetConfiguration
	- .SetExportFormat
	- .SetExportPath
	- SetArchivePath
	- SetExportDir
	- SetExportOptionsPlist
	- SetCustomOptions
	- SetExportDir
	- SetExportOptionsPlist
	- SetAuthentication

- xcodebuild.NewLegacyExportCommand()
- xcodebuild.NewExportCommand()
- xcodebuild.AuthenticationParams
- xcodebuild.NewResolvePackagesCommandModel
- xcodebuild.NewShowBuildSettingsCommand
- xcodebuild.NewCommandBuilder (with archive and clean archive actions)
- xcodebuild.NewExportCommand()
- xcodebuild.NewTestCommand
- xcodebuild.ExportCommandModel
*/
