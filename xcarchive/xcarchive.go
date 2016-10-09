package xcarchive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-xcode/exportoptions"
	"github.com/bitrise-tools/go-xcode/provisioningprofile"
)

func embeddedMobileProvisionPth(archivePth string) (string, error) {
	applicationPth := filepath.Join(archivePth, "/Products/Applications")
	mobileProvisionPthPattern := filepath.Join(applicationPth, "*.app/embedded.mobileprovision")
	mobileProvisionPths, err := filepath.Glob(mobileProvisionPthPattern)
	if err != nil {
		return "", fmt.Errorf("failed to find embedded.mobileprovision with pattern: %s, error: %s", mobileProvisionPthPattern, err)
	}
	if len(mobileProvisionPths) == 0 {
		return "", fmt.Errorf("no embedded.mobileprovision with pattern: %s", mobileProvisionPthPattern)
	}
	return mobileProvisionPths[0], nil
}

// DefaultExportOptions ...
func DefaultExportOptions(archivePth string) (exportoptions.ExportOptions, error) {
	embeddedProfilePth, err := embeddedMobileProvisionPth(archivePth)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedded mobileprovision path, error: %s", err)
	}

	provProfile, err := provisioningprofile.NewFromFile(embeddedProfilePth)
	if err != nil {
		return nil, fmt.Errorf("failed to collect embedded mobile provision, error: %s", err)
	}

	method := provProfile.GetExportMethod()
	developerTeamID := provProfile.GetDeveloperTeam()

	if method == exportoptions.MethodAppStore {
		options := exportoptions.NewAppStoreOptions()
		options.TeamID = developerTeamID
		return options, nil
	}

	options := exportoptions.NewNonAppStoreOptions(method)
	options.TeamID = developerTeamID
	return options, nil
}

// ExportDSYMs ...
func ExportDSYMs(archivePth string) (string, []string, error) {
	dsymsPattern := filepath.Join(archivePth, "dSYMs", "*.dSYM")
	dsyms, err := filepath.Glob(dsymsPattern)
	if err != nil {
		return "", []string{}, fmt.Errorf("failed to find dSYM with pattern: %s, error: %s", dsymsPattern, err)
	}
	appDSYM := ""
	frameworkDSYMs := []string{}
	for _, dsym := range dsyms {
		if strings.HasSuffix(dsym, "*.app.dSYM") {
			appDSYM = dsym
		} else {
			frameworkDSYMs = append(frameworkDSYMs, dsym)
		}
	}
	return appDSYM, frameworkDSYMs, nil
}

// CommandCallback ...
type CommandCallback func(printableCommand string)

// ExportIPA ...
func ExportIPA(archivePth, exportOptionsPth string, callback CommandCallback) (string, error) {
	return export(archivePth, exportOptionsPth, ExportFormatIPA, callback)
}

// ExportAPP ...
func ExportAPP(archivePth, exportOptionsPth string, callback CommandCallback) (string, error) {
	return export(archivePth, exportOptionsPth, ExportFormatAPP, callback)
}

func export(archivePth, exportOptionsPth string, exportFormat ExportFormat, callback CommandCallback) (string, error) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("output")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir, error: %s", err)
	}

	cmdSlice := []string{
		"xcodebuild", "-exportArchive",
		"-archivePath", archivePth,
		"-exportOptionsPlist", exportOptionsPth,
		"-exportPath", tmpDir,
	}

	if callback != nil {
		callback(cmdex.PrintableCommandArgs(false, cmdSlice))
	}

	cmd, err := cmdex.NewCommandFromSlice(cmdSlice)
	if err != nil {
		return "", fmt.Errorf("failed to create command from (%s)", strings.Join(cmdSlice, " "))
	}

	cmd.SetStdin(os.Stdin)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("export command failed, error: %s", err)
	}

	pattern := filepath.Join(tmpDir, exportFormat.Ext())
	matches, err := filepath.Glob(pattern)
	if len(matches) == 0 {
		return "", fmt.Errorf("no %s found with pattern: %s", exportFormat.String(), pattern)
	}

	return matches[0], nil
}

// EmbeddedProfileName ...
func EmbeddedProfileName(archivePth string) (string, error) {
	embeddedProfilePth, err := embeddedMobileProvisionPth(archivePth)
	if err != nil {
		return "", fmt.Errorf("failed to get embedded mobileprovision path, error: %s", err)
	}

	provProfile, err := provisioningprofile.NewFromFile(embeddedProfilePth)
	if err != nil {
		return "", fmt.Errorf("failed to collect embedded mobile provision, error: %s", err)
	}
	if provProfile.Name == nil {
		return "", fmt.Errorf("Name not found in prov profile")
	}

	return *provProfile.Name, nil
}

// ExportFormat ...
type ExportFormat string

const (
	// ExportFormatIPA ...
	ExportFormatIPA ExportFormat = "ipa"
	// ExportFormatAPP ...
	ExportFormatAPP ExportFormat = "app"
)

// Ext ...
func (exportFormat ExportFormat) Ext() string {
	switch exportFormat {
	case ExportFormatIPA:
		return ".ipa"
	case ExportFormatAPP:
		return ".app"
	default:
		return ""
	}
}

// String ...
func (exportFormat ExportFormat) String() string {
	switch exportFormat {
	case ExportFormatIPA:
		return "ipa"
	case ExportFormatAPP:
		return "app"
	default:
		return ""
	}
}

// LegacyExport ...
func LegacyExport(archivePth, provisioningProfileName string, exportFormat ExportFormat, callback CommandCallback) (string, error) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("output")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir, error: %s", err)
	}

	outputName := strings.TrimSuffix(filepath.Base(archivePth), filepath.Ext(archivePth))
	outputExt := exportFormat.Ext()
	outputPth := filepath.Join(tmpDir, outputName+outputExt)

	cmdSlice := []string{
		"xcodebuild", "-exportArchive",
		"-archivePath", archivePth,
		"-exportProvisioningProfile", provisioningProfileName,
		"-exportFormat", exportFormat.String(),
		"-exportPath", outputPth,
	}

	if callback != nil {
		callback(cmdex.PrintableCommandArgs(false, cmdSlice))
	}

	cmd, err := cmdex.NewCommandFromSlice(cmdSlice)
	if err != nil {
		return "", fmt.Errorf("failed to create command from (%s)", strings.Join(cmdSlice, " "))
	}

	cmd.SetStdin(os.Stdin)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("export command failed, error: %s", err)
	}

	return outputPth, nil
}
