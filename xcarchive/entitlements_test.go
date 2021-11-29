package xcarchive

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"

	"github.com/bitrise-io/go-xcode/plistutil"
	"github.com/stretchr/testify/assert"
)

func TestGiveniOS_WhenXcentFileIsMissing_ThenReadsEntitlementsFromTheExecutable(t *testing.T) {
	// Given
	appPath := filepath.Join(sampleRepoPath(t), "archives/Fruta.xcarchive/Products/Applications/Fruta.app")
	executable := executableRelativePath(appPath, "Info.plist", "")

	// When
	entitlements, err := getEntitlements(appPath, "non-existing-entitlements-file.xcent", executable)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, iosEntitlements(), entitlements)
}

func TestGivenMacos_WhenAskingForEntitlements_ThenReadsItFromTheXcentFile(t *testing.T) {
	// Given
	appPath := filepath.Join(sampleRepoPath(t), "archives/macos.xcarchive/Products/Applications/Test.app")
	executable := executableRelativePath(appPath, "Contents/Info.plist", "Contents/MacOS/")

	// When
	entitlements, err := getEntitlements(appPath, "Contents/Resources/archived-expanded-entitlements.xcent", executable)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, macosEntitlements(), entitlements)
}

func TestGivenMacos_WhenXcentFileIsMissing_ThenReadsEntitlementsFromTheExecutable(t *testing.T) {
	// Given
	appPath := filepath.Join(sampleRepoPath(t), "archives/macos.xcarchive/Products/Applications/Test.app")
	executable := executableRelativePath(appPath, "Contents/Info.plist", "Contents/MacOS/")

	// When
	entitlements, err := getEntitlements(appPath, "Contents/Resources/non-existing-entitlements-file.xcent", executable)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, macosEntitlements(), entitlements)
}

func executableRelativePath(basePath, infoPlistRelativePath, executableFolderRelativePath string) string {
	infoPlistPath := filepath.Join(basePath, infoPlistRelativePath)
	exist, err := pathutil.IsPathExists(infoPlistPath)
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
