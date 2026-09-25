package xcodeproj

import (
	"os"
	"path/filepath"
	"testing"

	rootmocks "github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// The signing values v1's tests used; the golden fixtures were produced with them.
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

func requireBuildSetting(t *testing.T, project *XcodeProj, target, configuration, key, want string) {
	t.Helper()

	buildSettings, err := targetBuildSettings(requireTarget(t, project, target), configuration)
	require.NoError(t, err)

	got, ok := buildSettings.String(key)
	require.True(t, ok, "build setting %s is missing", key)
	assert.Equal(t, want, got, key)
}

func requireTarget(t *testing.T, project *XcodeProj, name string) Target {
	t.Helper()
	target, ok := project.TargetByName(name)
	require.True(t, ok, "target %s not found", name)
	return target
}

func TestXcodeProj_ForceCodeSign(t *testing.T) {
	project := parseFixture(t, "without-target-attributes.pbxproj")

	require.NoError(t, project.ForceCodeSign(forceCodeSignOptions("Target", "Debug")))

	target := requireTarget(t, project, "Target")
	attributes, ok := project.targetAttributes.Object(target.ID)
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
	_, ok := project.targetAttributes.Object(target.ID)
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

func TestXcodeProj_perObjectModify(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		target        string
		configuration string
		want          string
	}{
		{
			name:          "no target attributes",
			fixture:       "without-target-attributes.pbxproj",
			target:        "TargetWithouthTargetAttributes",
			configuration: "Debug",
			want:          "without-target-attributes-modified.pbxproj",
		},
		{
			name:          "changes 2 objects, as target attributes are included in the project",
			fixture:       "ios-sample.pbxproj",
			target:        "XcodeProj",
			configuration: "Debug",
			want:          "ios-sample-changed.pbxproj",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := parseFixture(t, tt.fixture)
			require.NoError(t, project.ForceCodeSign(forceCodeSignOptions(tt.target, tt.configuration)))

			got, err := project.perObjectModify()
			require.NoError(t, err)

			want, err := os.ReadFile(filepath.Join("testdata", tt.want))
			require.NoError(t, err)
			require.Equal(t, string(want), string(got))
		})
	}
}

func TestXcodeProj_perObjectModify_noChangesReturnsOriginalBytes(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	got, err := project.perObjectModify()
	require.NoError(t, err)

	want, err := os.ReadFile(filepath.Join("testdata", "ios-sample.pbxproj"))
	require.NoError(t, err)
	require.Equal(t, string(want), string(got))
}

func TestXcodeProj_SetBuildSetting(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	// Fetched before the write: build settings are shared, as in v1.
	earlier := requireTarget(t, project, "XcodeProj")

	require.NoError(t, project.SetBuildSetting("XcodeProj", "Release", "MARKETING_VERSION", "2.0"))

	requireBuildSetting(t, project, "XcodeProj", "Release", "MARKETING_VERSION", "2.0")

	value, ok := earlier.BuildConfigurations[1].BuildSetting("MARKETING_VERSION")
	require.True(t, ok)
	assert.Equal(t, "2.0", value, "a Target fetched before the write must see it")

	buildSettings, err := targetBuildSettings(requireTarget(t, project, "XcodeProj"), "Debug")
	require.NoError(t, err)
	assert.NotEqual(t, "2.0", buildSettings["MARKETING_VERSION"])
}

func TestXcodeProj_SetBuildSetting_errors(t *testing.T) {
	project := parseFixture(t, "ios-sample.pbxproj")

	require.ErrorContains(t, project.SetBuildSetting("Nope", "Release", "K", "V"), "Nope")
	require.ErrorContains(t, project.SetBuildSetting("XcodeProj", "Nope", "K", "V"), "Nope")
}

func TestXcodeProj_SetBuildSetting_configurationWithoutBuildSettings(t *testing.T) {
	project := &XcodeProj{
		targets: []Target{{
			Name:                "App",
			BuildConfigurations: []BuildConfiguration{{Name: "Debug"}},
		}},
	}

	require.ErrorContains(t, project.SetBuildSetting("App", "Debug", "K", "V"), "no buildSettings")
}

func TestWriteBuildSettingForAllSDKs(t *testing.T) {
	buildSettings := serialized.Object{
		"CODE_SIGN_IDENTITY":                "old",
		"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "old",
		"CODE_SIGN_IDENTITY_OTHER":          "untouched",
	}

	writeBuildSettingForAllSDKs(buildSettings, "CODE_SIGN_IDENTITY", "new")

	assert.Equal(t, serialized.Object{
		"CODE_SIGN_IDENTITY":                "new",
		"CODE_SIGN_IDENTITY[sdk=iphoneos*]": "new",
		"CODE_SIGN_IDENTITY_OTHER":          "untouched",
	}, buildSettings)
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
	assert.Equal(t, os.FileMode(0664), info.Mode().Perm(), "an existing file keeps its mode, as in v1")
}

func TestXcodeProj_Save_newFileGetsV1Mode(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "minimal.pbxproj"))
	require.NoError(t, err)

	projectPath := filepath.Join(t.TempDir(), "App.xcodeproj")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	project, err := testFactory().Parse(content, projectPath)
	require.NoError(t, err)
	require.NoError(t, project.Save())

	info, err := os.Stat(filepath.Join(projectPath, "project.pbxproj"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

func TestXcodeProj_Save_fallsBackToFullRewrite(t *testing.T) {
	projectPath := filepath.Join(t.TempDir(), "App.xcodeproj")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	content, err := os.ReadFile(filepath.Join("testdata", "minimal.pbxproj"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "project.pbxproj"), content, 0644))

	project, err := testFactory().Open(projectPath)
	require.NoError(t, err)

	objects, ok := project.rawProj.Object("objects")
	require.True(t, ok)
	objects["BB00000000000000000000A1"] = map[string]any{"isa": "PBXGroup", "children": []any{}, "sourceTree": "<group>"}

	_, err = project.perObjectModify()
	require.ErrorContains(t, err, "new object added")

	require.NoError(t, project.Save())

	reopened, err := testFactory().Open(projectPath)
	require.NoError(t, err)
	require.Len(t, reopened.Targets(), 1)

	reopenedObjects, ok := reopened.rawProj.Object("objects")
	require.True(t, ok)
	assert.True(t, reopenedObjects.Has("BB00000000000000000000A1"))
}

// Overwriting must not chmod: FileManager.Write chmods, which fails for a file the process can write
// but doesn't own, so an existing file is written with WriteBytes.
func TestXcodeProj_Save_existingFileIsNotChmodded(t *testing.T) {
	project := parseFixture(t, "minimal.pbxproj")
	pth := filepath.Join(project.Path, "project.pbxproj")

	fileManager := rootmocks.NewFileManager(t)
	fileManager.On("Lstat", pth).Return(nil, nil)
	fileManager.On("WriteBytes", pth, mock.Anything).Return(nil)
	project.fileManager = fileManager

	require.NoError(t, project.Save())
}

// Save rebuilds the file from the bytes given to Parse, so Parse must not keep the caller's buffer.
func TestFactory_Parse_copiesContent(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "without-target-attributes.pbxproj"))
	require.NoError(t, err)
	want, err := os.ReadFile(filepath.Join("testdata", "without-target-attributes-modified.pbxproj"))
	require.NoError(t, err)

	project, err := testFactory().Parse(content, "/projects/App.xcodeproj")
	require.NoError(t, err)

	for i := range content {
		content[i] = 'x'
	}

	require.NoError(t, project.ForceCodeSign(forceCodeSignOptions("TargetWithouthTargetAttributes", "Debug")))
	got, err := project.perObjectModify()
	require.NoError(t, err)
	assert.Equal(t, string(want), string(got))
}
