package xcodeproj_tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/_integration_tests"
	"github.com/bitrise-io/go-xcode/v2/xcodebuild"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These run the real xcodebuild, which the unit tests only mock.

const (
	simpleObjcRepoURL = "https://github.com/bitrise-io/sample-apps-ios-simple-objc.git"
	iosSampleRepoURL  = "https://github.com/bitrise-io/newly-generated-ios-sample-project.git"
)

func buildSettingsProvider() xcodebuild.BuildSettingsProvider {
	logger := log.NewLogger()
	return xcodebuild.NewShowBuildSettingsProvider(command.NewFactory(env.NewRepository()), logger)
}

func newFactory() xcodeproj.Factory {
	return xcodeproj.NewFactory(
		log.NewLogger(),
		buildSettingsProvider(),
		fileutil.NewFileManager(),
		pathutil.NewPathModifier(),
		pathutil.NewPathProvider(),
		xcodeproj.NewUserProvider(),
	)
}

func simpleObjcProjectPath(t *testing.T) string {
	return filepath.Join(_integration_tests.GetRepository(t, simpleObjcRepoURL, "master"), "ios-simple-objc", "ios-simple-objc.xcodeproj")
}

func TestXcodeProj_simpleObjc(t *testing.T) {
	projectPath := simpleObjcProjectPath(t)
	project, err := newFactory().Open(projectPath)
	require.NoError(t, err)

	target, ok := project.TargetByName("ios-simple-objc")
	require.True(t, ok)

	bundleID, err := project.TargetBundleID(target.Name, "Release")
	require.NoError(t, err)
	assert.Equal(t, "Bitrise.ios-simple-objc", bundleID)

	infoPlistPath, err := project.TargetInfoplistPath(target.Name, "Release")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(filepath.Dir(projectPath), "ios-simple-objc", "Info.plist"), infoPlistPath)

	_, err = project.TargetCodeSignEntitlements(target.Name, "Release")
	assert.ErrorIs(t, err, xcodeproj.ErrEntitlementsNotFound)

	scheme, _, err := project.Scheme("ios-simple-objc")
	require.NoError(t, err)
	assert.True(t, scheme.IsShared)

	icons, err := project.AppIconSetPaths()
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(filepath.Dir(projectPath), "ios-simple-objc", "Images.xcassets", "AppIcon.appiconset")}, icons[target.ID])
}

func TestXcodeProj_iosSample(t *testing.T) {
	projectPath := filepath.Join(_integration_tests.GetRepository(t, iosSampleRepoURL, "main"), "ios-sample.xcodeproj")
	project, err := newFactory().Open(projectPath)
	require.NoError(t, err)

	entitlements, err := project.TargetCodeSignEntitlements("ios-sample", "Release")
	require.NoError(t, err)
	sandbox, ok := entitlements.Bool("com.apple.security.app-sandbox")
	require.True(t, ok)
	assert.True(t, sandbox)

	// The target generates its Info.plist, so there is no INFOPLIST_FILE to resolve.
	_, err = project.TargetInfoplistPath("ios-sample", "Release")
	assert.ErrorIs(t, err, xcodeproj.ErrInfoPlistNotFound)
}

// -scheme needs a destination for the scheme's platform. The multiplatform ios-sample can resolve
// to macOS, which every Xcode install has; an iOS-only project would need the iOS platform installed.
func TestShowBuildSettingsProvider_SchemeBuildSettings(t *testing.T) {
	projectPath := filepath.Join(_integration_tests.GetRepository(t, iosSampleRepoURL, "main"), "ios-sample.xcodeproj")

	settings, err := buildSettingsProvider().SchemeBuildSettings(projectPath, "ios-sample", "Release")
	require.NoError(t, err)

	bundleID, ok := settings.String("PRODUCT_BUNDLE_IDENTIFIER")
	require.True(t, ok)
	assert.Equal(t, "io.bitrise.ios-sample", bundleID)
}

// Edits a copy, so the shared clone stays untouched, and reads the result back through xcodebuild.
func TestXcodeProj_ForceCodeSignRoundTrip(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(simpleObjcProjectPath(t), "project.pbxproj"))
	require.NoError(t, err)

	projectPath := filepath.Join(t.TempDir(), "ios-simple-objc.xcodeproj")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	project, err := newFactory().Parse(content, projectPath)
	require.NoError(t, err)
	require.NoError(t, project.ForceCodeSign(xcodeproj.ForceCodeSignOptions{
		TargetName:              "ios-simple-objc",
		Configuration:           "Release",
		DevelopmentTeam:         "TEAM123456",
		CodesignIdentity:        "Apple Distribution: Bitrise (TEAM123456)",
		ProvisioningProfileUUID: "11111111-2222-3333-4444-555555555555",
	}))
	require.NoError(t, project.Save())

	reopened, err := newFactory().Open(projectPath)
	require.NoError(t, err)

	settings, err := reopened.TargetBuildSettings("ios-simple-objc", "Release")
	require.NoError(t, err)
	for key, want := range map[string]string{
		"CODE_SIGN_STYLE":      "Manual",
		"DEVELOPMENT_TEAM":     "TEAM123456",
		"CODE_SIGN_IDENTITY":   "Apple Distribution: Bitrise (TEAM123456)",
		"PROVISIONING_PROFILE": "11111111-2222-3333-4444-555555555555",
	} {
		got, _ := settings.String(key)
		assert.Equal(t, want, got, key)
	}
}
