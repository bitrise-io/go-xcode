package xcarchive

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/stretchr/testify/assert"
)

func TestGiveniOS_WhenAskingForEntitlements_ThenReadsItFromTheExecutable(t *testing.T) {
	// Given
	cmdFactory := command.NewFactory(env.NewRepository())
	appPath := filepath.Join(sampleRepoPath(t), "archives/Fruta.xcarchive/Products/Applications/Fruta.app")
	executable := executableRelativePath(appPath, "Info.plist", "")

	// When
	entitlements, err := getEntitlements(cmdFactory, appPath, executable)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, iosEntitlements(), entitlements)
}

func TestGivenMacos_WhenAskingForEntitlements_ThenReadsItFromTheExecutable(t *testing.T) {
	// Given
	cmdFactory := command.NewFactory(env.NewRepository())
	appPath := filepath.Join(sampleRepoPath(t), "archives/macos.xcarchive/Products/Applications/Test.app")
	executable := executableRelativePath(appPath, "Contents/Info.plist", "Contents/MacOS/")

	// When
	entitlements, err := getEntitlements(cmdFactory, appPath, executable)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, macosEntitlements(), entitlements)
}

func executableRelativePath(basePath, infoPlistRelativePath, executableFolderRelativePath string) string {
	infoPlistPath := filepath.Join(basePath, infoPlistRelativePath)
	exist, err := pathutil.NewPathChecker().IsPathExists(infoPlistPath)
	if err != nil {
		return ""
	}

	if exist == false {
		return ""
	}

	plist, err := plistutil.NewPlistDataFromFile(infoPlistPath)
	if err != nil {
		return ""
	}

	return filepath.Join(executableFolderRelativePath, executableNameFromInfoPlist(plist))
}

func iosEntitlements() plistutil.PlistData {
	return map[string]interface{}{
		"application-identifier":                           "72SA8V3WYL.io.bitrise.appcliptest",
		"com.apple.developer.applesignin":                  []interface{}{"Default"},
		"com.apple.developer.icloud-container-identifiers": []interface{}{},
		"com.apple.developer.team-identifier":              "72SA8V3WYL",
		"com.apple.security.application-groups":            []interface{}{"group.io.bitrise.appcliptest"},
		"get-task-allow":                                   false,
	}
}

func macosEntitlements() plistutil.PlistData {
	return map[string]interface{}{
		"com.apple.security.app-sandbox":                   true,
		"com.apple.security.files.user-selected.read-only": true,
	}
}

func Test_GivenArchiveWithMultipleAppAndFrameworkDSYMs_WhenFindDSYMsCalled_ThenExpectAllDSYMsToBeReturned(t *testing.T) {
	testCases := []struct {
		name                   string
		numberOfAppDSYMs       int
		numberOfFrameworkDSYMs int
	}{
		{
			name:                   "1. Given archive with multiple app and framework dSYMs when FindDSYMs called then expect all dSYMs to be returned",
			numberOfAppDSYMs:       2,
			numberOfFrameworkDSYMs: 2,
		},
		{
			name:                   "2. Given archive with singe app and framework dSYMs when FindDSYMs called then expect both dSYMs to be returned",
			numberOfAppDSYMs:       1,
			numberOfFrameworkDSYMs: 1,
		},
		{
			name:                   "3. Given archive with multiple app dSYMs when FindDSYMs called then expect all app dSYMs to be returned",
			numberOfAppDSYMs:       2,
			numberOfFrameworkDSYMs: 0,
		},
		{
			name:                   "4. Given archive with multiple framework dSYMs when FindDSYMs called then expect all framework dSYMs to be returned",
			numberOfAppDSYMs:       0,
			numberOfFrameworkDSYMs: 2,
		},
		{
			name:                   "5. Given archive with single app dSYM when FindDSYMs called then expect the app dSYM to be returned",
			numberOfAppDSYMs:       1,
			numberOfFrameworkDSYMs: 0,
		},
		{
			name:                   "6. Given archive with single framework dSYM when FindDSYMs called then expect the framework dSYM to be returned",
			numberOfAppDSYMs:       0,
			numberOfFrameworkDSYMs: 1,
		},
		{
			name:                   "7. Given archive without any dSYM when FindDSYMs called then expect no dSYM to be returned",
			numberOfAppDSYMs:       0,
			numberOfFrameworkDSYMs: 0,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			archivePath, err := createArchiveWithAppAndFrameworkDSYMs(
				"archives/ios.dsyms.xcarchive",
				testCase.numberOfAppDSYMs,
				testCase.numberOfFrameworkDSYMs,
			)
			assert.NoError(t, err)

			appDSYMs, frameworkDSYMs, err := findDSYMs(archivePath)
			assert.NoError(t, err)
			assert.Equal(t, testCase.numberOfAppDSYMs, len(appDSYMs))
			assert.Equal(t, testCase.numberOfFrameworkDSYMs, len(frameworkDSYMs))
		})
	}
}

func createArchiveWithAppAndFrameworkDSYMs(archivePath string, numberOfAppDSYMs, numberOfFrameworkDSYMs int) (string, error) {
	archivePath, err := createArchive(archivePath)
	if err != nil {
		return "", err
	}

	err = createAppDSYMs(archivePath, numberOfAppDSYMs)
	if err != nil {
		return "", err
	}

	err = createFrameworkDSYMs(archivePath, numberOfFrameworkDSYMs)
	if err != nil {
		return "", err
	}

	return archivePath, nil
}

func createAppDSYMs(archivePath string, numberOfDSYMs int) error {
	return createDSYMs(archivePath, "app", numberOfDSYMs)
}

func createFrameworkDSYMs(archivePath string, numberOfDSYMs int) error {
	return createDSYMs(archivePath, "framework", numberOfDSYMs)
}

func createDSYMs(archivePath, dSYMType string, numberOfDSYMs int) error {
	for i := 0; i < numberOfDSYMs; i++ {
		err := os.WriteFile(createDSYMFilePath(archivePath, dSYMType, i), nil, 0777)
		if err != nil {
			return err
		}
	}

	return nil
}

func createDSYMFilePath(archivePath, dSYMType string, index int) string {
	return filepath.Join(archivePath, DSYMSDirName, fmt.Sprintf("ios-%d.%s.dSYM", index, dSYMType))
}

func createArchive(archivePath string) (string, error) {
	tempDirPath, err := pathutil.NewPathProvider().CreateTempDir(tempDirName)
	if err != nil {
		return "", err
	}

	archivePath = filepath.Join(tempDirPath, archivePath)
	err = os.MkdirAll(filepath.Join(archivePath, DSYMSDirName), 0755)
	if err != nil {
		return "", err
	}

	return archivePath, nil
}
