package xcodeproj

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The signing values the golden fixtures were produced with.
var testForceCodeSignOptions = ForceCodeSignOptions{
	DevelopmentTeam:         "ABCD1234",
	CodesignIdentity:        "Apple Development: John Doe (ASDF1234)",
	ProvisioningProfileUUID: "asdf56b6-e75a-4f86-bf25-101bfc2fasdf",
}

func forceCodeSignOptions(target, configuration string) ForceCodeSignOptions {
	opts := testForceCodeSignOptions
	opts.TargetName = target
	opts.Configuration = configuration
	return opts
}

func requireBuildSetting(t *testing.T, project XcodeProj, target, configuration, key, want string) {
	t.Helper()

	got, ok := requireBuildSettings(t, requireTarget(t, project, target), configuration).String(key)
	require.True(t, ok, "build setting %s is missing", key)
	assert.Equal(t, want, got, key)
}

func requireBuildSettings(t *testing.T, target Target, configuration string) serialized.Object {
	t.Helper()

	for _, buildConfiguration := range target.BuildConfigurations {
		if buildConfiguration.Name == configuration {
			return buildConfiguration.buildSettings
		}
	}

	require.Failf(t, "build configuration not found", "%s of target %s", configuration, target.Name)
	return nil
}

func requireTarget(t *testing.T, project XcodeProj, name string) Target {
	t.Helper()
	target, ok := project.TargetByName(name)
	require.True(t, ok, "target %s not found", name)
	return target
}

func TestXcodeProj_ForceCodeSign(t *testing.T) {
	project := parseFixture(t, "without-target-attributes.pbxproj")

	require.NoError(t, project.ForceCodeSign(forceCodeSignOptions("Target", "Debug")))

	target := requireTarget(t, project, "Target")
	attributes, ok := pbxFile(t, project).targetAttributes.Object(target.ID)
	require.True(t, ok)
	assert.Equal(t, "Manual", attributes["ProvisioningStyle"])
	assert.Equal(t, "ABCD1234", attributes["DevelopmentTeam"])
	assert.Equal(t, "", attributes["DevelopmentTeamName"])

	requireBuildSetting(t, project, "Target", "Debug", "CODE_SIGN_STYLE", "Manual")
	requireBuildSetting(t, project, "Target", "Debug", "DEVELOPMENT_TEAM", "ABCD1234")
	requireBuildSetting(t, project, "Target", "Debug", "CODE_SIGN_IDENTITY", "Apple Development: John Doe (ASDF1234)")
	requireBuildSetting(t, project, "Target", "Debug", "CODE_SIGN_IDENTITY[sdk=iphoneos*]", "Apple Development: John Doe (ASDF1234)")
	requireBuildSetting(t, project, "Target", "Debug", "PROVISIONING_PROFILE_SPECIFIER", "")
	requireBuildSetting(t, project, "Target", "Debug", "PROVISIONING_PROFILE", "asdf56b6-e75a-4f86-bf25-101bfc2fasdf")
}

func TestXcodeProj_ForceCodeSign_targetWithoutTargetAttributes(t *testing.T) {
	project := parseFixture(t, "without-target-attributes.pbxproj")

	require.NoError(t, project.ForceCodeSign(forceCodeSignOptions("TargetWithouthTargetAttributes", "Debug")))

	requireBuildSetting(t, project, "TargetWithouthTargetAttributes", "Debug", "CODE_SIGN_STYLE", "Manual")
	requireBuildSetting(t, project, "TargetWithouthTargetAttributes", "Debug", "DEVELOPMENT_TEAM", "ABCD1234")
	requireBuildSetting(t, project, "TargetWithouthTargetAttributes", "Debug", "CODE_SIGN_IDENTITY", "Apple Development: John Doe (ASDF1234)")
	requireBuildSetting(t, project, "TargetWithouthTargetAttributes", "Debug", "PROVISIONING_PROFILE_SPECIFIER", "")
	requireBuildSetting(t, project, "TargetWithouthTargetAttributes", "Debug", "PROVISIONING_PROFILE", "asdf56b6-e75a-4f86-bf25-101bfc2fasdf")

	target := requireTarget(t, project, "TargetWithouthTargetAttributes")
	_, ok := pbxFile(t, project).targetAttributes.Object(target.ID)
	assert.False(t, ok, "no TargetAttributes entry may be created for a target that had none")
}

func TestXcodeProj_ForceCodeSign_leavesOtherBuildSettingsAlone(t *testing.T) {
	project := parseFixture(t, "without-target-attributes.pbxproj")

	require.NoError(t, project.ForceCodeSign(forceCodeSignOptions("TargetWithouthTargetAttributes", "Debug")))

	requireBuildSetting(t, project, "TargetWithouthTargetAttributes", "Debug", "INFOPLIST_FILE", "Target copy-Info.plist")
}

func TestXcodeProj_ForceCodeSign_unknownTargetOrConfiguration(t *testing.T) {
	project := parseFixture(t, "without-target-attributes.pbxproj")

	require.ErrorContains(t, project.ForceCodeSign(forceCodeSignOptions("Nope", "Debug")), "Nope")
	require.ErrorContains(t, project.ForceCodeSign(forceCodeSignOptions("Target", "Nope")), "Nope")
}

func TestXcodeProj_SetBuildSetting(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	// Fetched before the write: targets share their build settings with the project.
	earlier := requireTarget(t, project, "XcodeProj")

	require.NoError(t, project.SetBuildSetting("XcodeProj", "Release", "MARKETING_VERSION", "2.0"))

	requireBuildSetting(t, project, "XcodeProj", "Release", "MARKETING_VERSION", "2.0")

	value, ok := earlier.BuildConfigurations[1].BuildSetting("MARKETING_VERSION")
	require.True(t, ok)
	assert.Equal(t, "2.0", value, "a Target fetched before the write must see it")

	assert.NotEqual(t, "2.0", requireBuildSettings(t, requireTarget(t, project, "XcodeProj"), "Debug")["MARKETING_VERSION"])
}

func TestXcodeProj_SetBuildSetting_errors(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	require.ErrorContains(t, project.SetBuildSetting("Nope", "Release", "K", "V"), "Nope")
	require.ErrorContains(t, project.SetBuildSetting("XcodeProj", "Nope", "K", "V"), "Nope")
}

func TestKeysForAllSDKs(t *testing.T) {
	buildSettings := serialized.Object{
		"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "old",
		"CODE_SIGN_IDENTITY_OTHER":          "untouched",
	}

	keys := keysForAllSDKs(buildSettings, "CODE_SIGN_IDENTITY")

	assert.ElementsMatch(t, []string{"CODE_SIGN_IDENTITY", "CODE_SIGN_IDENTITY[sdk=iphoneos*]"}, keys)
}

func TestXcodeProj_Save(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "XcodeProj.xcodeproj")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	content, err := os.ReadFile(filepath.Join("testdata", "ios-sample.pbxproj"))
	require.NoError(t, err)
	pbxProjPath := filepath.Join(projectPath, "project.pbxproj")
	require.NoError(t, os.WriteFile(pbxProjPath, content, 0600))
	require.NoError(t, os.Chmod(pbxProjPath, 0664))

	project, err := testFactory().Open(projectPath)
	require.NoError(t, err)

	require.NoError(t, project.SetBuildSetting("XcodeProj", "Release", "MARKETING_VERSION", "2.0"))
	require.NoError(t, project.Save())

	reopened, err := testFactory().Open(projectPath)
	require.NoError(t, err)
	requireBuildSetting(t, reopened, "XcodeProj", "Release", "MARKETING_VERSION", "2.0")

	info, err := os.Stat(pbxProjPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0664), info.Mode().Perm(), "an existing file keeps its mode")
}

func TestXcodeProj_Save_newFileGetsDefaultMode(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "minimal.pbxproj"))
	require.NoError(t, err)

	projectPath := filepath.Join(t.TempDir(), "App.xcodeproj")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	project, err := testFactory().newPBXProj(content, projectPath)
	require.NoError(t, err)
	require.NoError(t, project.Save())

	info, err := os.Stat(filepath.Join(projectPath, "project.pbxproj"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}
