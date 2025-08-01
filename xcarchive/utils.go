package xcarchive

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-xcode/plistutil"
)

func executableNameFromInfoPlist(infoPlist plistutil.PlistData) string {
	if name, ok := infoPlist.GetString("CFBundleExecutable"); ok {
		return name
	}
	return ""
}

func getEntitlements(basePath, executableRelativePath string) (plistutil.PlistData, error) {
	entitlements, err := entitlementsFromExecutable(basePath, executableRelativePath)
	if err != nil {
		return plistutil.PlistData{}, err
	}

	if entitlements != nil {
		return *entitlements, nil
	}

	return plistutil.PlistData{}, nil
}

func entitlementsFromExecutable(basePath, executableRelativePath string) (*plistutil.PlistData, error) {
	fmt.Printf("Fetching entitlements from executable")

	cmd := command.New("codesign", "--display", "--entitlements", ":-", filepath.Join(basePath, executableRelativePath))
	entitlementsString, err := cmd.RunAndReturnTrimmedOutput()
	if err != nil {
		return nil, err
	}

	plist, err := plistutil.NewPlistDataFromContent(entitlementsString)
	if err != nil {
		return nil, err
	}

	return &plist, nil
}

func findDSYMs(archivePath string) ([]string, []string, error) {
	dsymsDirPth := filepath.Join(archivePath, "dSYMs")
	dsyms, err := pathutil.ListEntries(dsymsDirPth, pathutil.ExtensionFilter(".dsym", true))
	if err != nil {
		return []string{}, []string{}, err
	}

	appDSYMs := []string{}
	frameworkDSYMs := []string{}
	for _, dsym := range dsyms {
		if strings.HasSuffix(dsym, ".app.dSYM") {
			appDSYMs = append(appDSYMs, dsym)
		} else {
			frameworkDSYMs = append(frameworkDSYMs, dsym)
		}
	}

	return appDSYMs, frameworkDSYMs, nil
}
