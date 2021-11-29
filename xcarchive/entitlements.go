package xcarchive

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-xcode/plistutil"
)

func executableNameFromInfoPlist(infoPlist plistutil.PlistData) string {
	if name, ok := infoPlist.GetString("CFBundleExecutable"); ok {
		return name
	}
	return ""
}

func getEntitlements(basePath, entitlementsRelativePath, executableRelativePath string) (plistutil.PlistData, error) {
	entitlements, err := entitlementsFromXcentFile(basePath, entitlementsRelativePath)
	if err != nil {
		return plistutil.PlistData{}, err
	}

	if entitlements != nil {
		return *entitlements, nil
	}

	entitlements, err = entitlementsFromExecutable(basePath, executableRelativePath)
	if err != nil {
		return plistutil.PlistData{}, err
	}

	if entitlements != nil {
		return *entitlements, nil
	}

	return plistutil.PlistData{}, nil
}

func entitlementsFromXcentFile(basePath, entitlementsRelativePath string) (*plistutil.PlistData, error) {
	entitlementsPath := filepath.Join(basePath, entitlementsRelativePath)
	exist, err := pathutil.IsPathExists(entitlementsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check if entitlements exists at: %s, error: %s", entitlementsPath, err)
	}

	if exist == false {
		return nil, nil
	}

	plist, err := plistutil.NewPlistDataFromFile(entitlementsPath)
	if err != nil {
		return nil, err
	}

	return &plist, nil
}

func entitlementsFromExecutable(basePath, executableRelativePath string) (*plistutil.PlistData, error) {
	factory := command.NewFactory(env.NewRepository())
	cmd := factory.Create("codesign", []string{"--display", "--entitlements", ":-", filepath.Join(basePath, executableRelativePath)}, nil)
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
