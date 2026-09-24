package xcodeproj

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUserName = "bitrise"

type fakeUserProvider struct {
	name string
	err  error
}

func (f fakeUserProvider) CurrentUserName() (string, error) { return f.name, f.err }

const minimalSchemeXML = `<?xml version="1.0" encoding="UTF-8"?>
<Scheme LastUpgradeVersion = "1240" version = "1.3">
</Scheme>
`

func schemesProject(t *testing.T, fixture string, user UserProvider) *XcodeProj {
	t.Helper()

	projectPath := filepath.Join(t.TempDir(), "XcodeProj.xcodeproj")
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	content, err := os.ReadFile(filepath.Join("testdata", fixture))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "project.pbxproj"), content, 0644))

	factory := NewFactory(log.NewLogger(), nil, fileutil.NewFileManager(), pathutil.NewPathModifier(), nil, user)
	project, err := factory.Open(projectPath)
	require.NoError(t, err)

	return project
}

func writeFile(t *testing.T, pth, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(pth), 0755))
	require.NoError(t, os.WriteFile(pth, []byte(content), 0644))
}

func sharedSchemePath(project *XcodeProj, name string) string {
	return filepath.Join(project.Path, "xcshareddata", "xcschemes", name+".xcscheme")
}

func userSchemesPath(project *XcodeProj, file string) string {
	return filepath.Join(project.Path, "xcuserdata", testUserName+".xcuserdatad", "xcschemes", file)
}

func writeAutocreateSetting(t *testing.T, project *XcodeProj, value string) {
	t.Helper()
	writeFile(t, filepath.Join(project.Path, "project.xcworkspace", "xcshareddata", "WorkspaceSettings.xcsettings"),
		`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>IDEWorkspaceSharedSettings_AutocreateContextsIfNeeded</key>
	`+value+`
</dict>
</plist>`)
}

func schemeNames(schemes []xcscheme.Scheme) []string {
	var names []string
	for _, scheme := range schemes {
		names = append(names, scheme.Name)
	}
	return names
}

func TestXcodeProj_Schemes_sharedAndUserSchemes(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: testUserName})
	writeFile(t, sharedSchemePath(project, "Shared"), minimalSchemeXML)
	writeFile(t, userSchemesPath(project, "Mine.xcscheme"), minimalSchemeXML)
	writeFile(t, userSchemesPath(project, "notes.txt"), "not a scheme")

	schemes, err := project.Schemes()
	require.NoError(t, err)

	require.Equal(t, []string{"Shared", "Mine"}, schemeNames(schemes), "shared schemes come first; non-.xcscheme files are ignored")
	assert.True(t, schemes[0].IsShared)
	assert.False(t, schemes[1].IsShared)
	assert.Equal(t, sharedSchemePath(project, "Shared"), schemes[0].Path)
}

func TestXcodeProj_Schemes_otherUsersSchemesAreNotVisible(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: "someone-else"})
	writeFile(t, sharedSchemePath(project, "Shared"), minimalSchemeXML)
	writeFile(t, userSchemesPath(project, "Mine.xcscheme"), minimalSchemeXML)

	schemes, err := project.Schemes()
	require.NoError(t, err)
	assert.Equal(t, []string{"Shared"}, schemeNames(schemes))
}

func TestXcodeProj_Schemes_autocreatedByDefault(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: testUserName})

	schemes, err := project.Schemes()
	require.NoError(t, err)

	require.Equal(t, []string{"XcodeProj", "TodayExtension"}, schemeNames(schemes))
	for _, scheme := range schemes {
		assert.True(t, scheme.IsShared, "generated schemes are marked shared, as in v1")
	}

	app := schemes[0]
	require.Len(t, app.TestAction.Testables, 1)
	assert.Equal(t, uiTestTargetID, app.TestAction.Testables[0].BuildableReference.BlueprintIdentifier)
	assert.Equal(t, appTargetID, app.BuildAction.BuildActionEntries[0].BuildableReference.BlueprintIdentifier)
	assert.Equal(t, "XcodeProj.app", app.BuildAction.BuildActionEntries[0].BuildableReference.BuildableName)
	assert.Equal(t, "container:XcodeProj.xcodeproj", app.BuildAction.BuildActionEntries[0].BuildableReference.ReferencedContainer)
	assert.Equal(t, "Release", app.ArchiveAction.BuildConfiguration)
	assert.Equal(t, "Debug", app.LaunchAction.BuildConfiguration)
}

func TestXcodeProj_Schemes_schemeManagementFileMeansDefaults(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: testUserName})
	writeAutocreateSetting(t, project, "<false/>")
	writeFile(t, userSchemesPath(project, "xcschememanagement.plist"), "")

	schemes, err := project.Schemes()
	require.NoError(t, err)
	assert.Equal(t, []string{"XcodeProj", "TodayExtension"}, schemeNames(schemes),
		"an existing scheme-management file yields the defaults even with autocreate off")
}

func TestXcodeProj_Schemes_autocreateOff(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: testUserName})
	writeAutocreateSetting(t, project, "<false/>")

	_, err := project.Schemes()
	require.ErrorContains(t, err, "'Autocreate schemes' option is disabled")
}

func TestXcodeProj_SchemesWithAutocreateOverride(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: testUserName})
	// The override must win over the project's own setting.
	writeAutocreateSetting(t, project, "<true/>")

	schemes, err := project.SchemesWithAutocreateOverride(false)
	require.NoError(t, err)
	assert.Nil(t, schemes)

	schemes, err = project.SchemesWithAutocreateOverride(true)
	require.NoError(t, err)
	assert.Equal(t, []string{"XcodeProj", "TodayExtension"}, schemeNames(schemes))
}

func TestXcodeProj_Schemes_settingsNotReadWhenSchemesExist(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: testUserName})
	writeFile(t, sharedSchemePath(project, "Shared"), minimalSchemeXML)
	writeFile(t, filepath.Join(project.Path, "project.xcworkspace", "xcshareddata", "WorkspaceSettings.xcsettings"), "not a plist")

	schemes, err := project.Schemes()
	require.NoError(t, err)
	assert.Equal(t, []string{"Shared"}, schemeNames(schemes))
}

func TestXcodeProj_Schemes_autocreateSettingThatIsNotABoolean(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: testUserName})
	writeAutocreateSetting(t, project, "<string>YES</string>")

	_, err := project.Schemes()
	require.ErrorContains(t, err, "not a boolean")
}

func TestXcodeProj_Schemes_autocreateWithNothingToGenerate(t *testing.T) {
	project := &XcodeProj{
		Path:         filepath.Join(t.TempDir(), "Tests.xcodeproj"),
		logger:       log.NewLogger(),
		fileManager:  fileutil.NewFileManager(),
		userProvider: fakeUserProvider{name: testUserName},
		targets: []Target{
			{ID: "T", Name: "Tests", isa: nativeTargetISA, productType: "com.apple.product-type.bundle.unit-test"},
		},
	}

	schemes, err := project.Schemes()
	require.NoError(t, err)
	assert.Empty(t, schemes)
}

func TestXcodeProj_Schemes_malformedSchemeFile(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: testUserName})
	writeFile(t, sharedSchemePath(project, "Broken"), "<Scheme")

	_, err := project.Schemes()
	require.ErrorContains(t, err, "Broken.xcscheme")
}

func TestXcodeProj_Schemes_userProviderFailure(t *testing.T) {
	failure := errors.New("no such user")
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{err: failure})

	_, err := project.Schemes()
	assert.ErrorIs(t, err, failure)
}

func TestXcodeProj_Scheme(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: testUserName})
	// "Café" with a decomposed é (e + combining acute accent).
	writeFile(t, sharedSchemePath(project, "Café"), minimalSchemeXML)

	t.Run("found, matching regardless of Unicode normalisation", func(t *testing.T) {
		// Looked up with a precomposed é.
		scheme, container, err := project.Scheme("Café")
		require.NoError(t, err)
		assert.Equal(t, "Café", scheme.Name)
		assert.Equal(t, project.Path, container)
	})

	t.Run("unknown name", func(t *testing.T) {
		_, _, err := project.Scheme("Nope")
		var notFound xcscheme.NotFoundError
		require.ErrorAs(t, err, &notFound)
		assert.Equal(t, "Nope", notFound.Scheme)
	})
}
