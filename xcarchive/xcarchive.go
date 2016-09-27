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

// ExportIpa ...
func ExportIpa(archivePth, exportOptionsPth string) (string, error) {
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

	fmt.Printf("=> %s\n", cmdex.PrintableCommandArgs(false, cmdSlice))

	cmd, err := cmdex.NewCommandFromSlice(cmdSlice)
	if err != nil {
		return "", fmt.Errorf("failed to create command from (%s)", strings.Join(cmdSlice, " "))
	}

	cmd.SetStdin(os.Stdin)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("export command failed, error: %s", err)
	}

	ipaPthPattern := filepath.Join(tmpDir, "*.ipa")
	matches, err := filepath.Glob(ipaPthPattern)
	if len(matches) == 0 {
		return "", fmt.Errorf("no ipa found with pattern: %s", ipaPthPattern)
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

// LegacyExportIpa ...
func LegacyExportIpa(archivePth, provisioningProfileName string) (string, error) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("output")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir, error: %s", err)
	}

	archiveName := strings.TrimSuffix(filepath.Base(archivePth), filepath.Ext(archivePth))
	ipaPth := filepath.Join(tmpDir, archiveName+".ipa")

	cmdSlice := []string{
		"xcodebuild", "-exportArchive",
		"-archivePath", archivePth,
		"-exportFormat", "ipa",
		"-exportProvisioningProfile", provisioningProfileName,
		"-exportPath", ipaPth,
	}

	fmt.Printf("=> %s\n", cmdex.PrintableCommandArgs(false, cmdSlice))

	cmd, err := cmdex.NewCommandFromSlice(cmdSlice)
	if err != nil {
		return "", fmt.Errorf("failed to create command from (%s)", strings.Join(cmdSlice, " "))
	}

	cmd.SetStdin(os.Stdin)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("export command failed, error: %s", err)
	}

	return ipaPth, nil
}
