package xcworkspace

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/fileutil"

	"github.com/bitrise-io/go-xcode/xcodeproject/testhelper"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcscheme"
	"github.com/stretchr/testify/require"
)

func Test_GivenNewlyGeneratedWorkspace_WhenListingSchemes_ThenReturnsTheDefaultScheme(t *testing.T) {
	xcodeWorkspacePath := testhelper.NewlyGeneratedXcodeWorkspacePath(t)
	workspace, err := Open(xcodeWorkspacePath)
	require.NoError(t, err)

	schemesByContainer, err := workspace.Schemes()
	require.NoError(t, err)

	expectedSchemeName := "ios-sample"
	var actualSchemes []xcscheme.Scheme
	for _, schemes := range schemesByContainer {
		actualSchemes = append(actualSchemes, schemes...)
	}

	require.Equal(t, 1, len(actualSchemes))
	require.Equal(t, expectedSchemeName, actualSchemes[0].Name)
}

func Test_GivenNewlyGeneratedWorkspaceWithAutocreateSchemesDisabled_WhenListingSchemes_ThenReturnsError(t *testing.T) {
	xcodeWorkspacePath := testhelper.NewlyGeneratedXcodeWorkspacePath(t)

	xcodeProjectPath := filepath.Join(filepath.Dir(xcodeWorkspacePath), "ios-sample.xcodeproj")
	projectEmbeddedWorksaceSettingsPth := filepath.Join(xcodeProjectPath, "project.xcworkspace/xcshareddata/WorkspaceSettings.xcsettings")
	require.NoError(t, fileutil.WriteStringToFile(projectEmbeddedWorksaceSettingsPth, workspaceSettingsWithAutocreateSchemesEnabledContent))

	worksaceSettingsPth := filepath.Join(xcodeWorkspacePath, "xcshareddata/WorkspaceSettings.xcsettings")
	require.NoError(t, fileutil.WriteStringToFile(worksaceSettingsPth, workspaceSettingsWithAutocreateSchemesDisabledContent))

	workspace, err := Open(xcodeWorkspacePath)
	require.NoError(t, err)

	schemesByContainer, err := workspace.Schemes()
	require.EqualError(t, err, `no schemes found and the Xcode project's 'Autocreate schemes' option is disabled`)
	require.Equal(t, 0, len(schemesByContainer))
}

const workspaceSettingsWithAutocreateSchemesDisabledContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>IDEWorkspaceSharedSettings_AutocreateContextsIfNeeded</key>
	<false/>
</dict>
</plist>
`

const workspaceSettingsWithAutocreateSchemesEnabledContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>IDEWorkspaceSharedSettings_AutocreateContextsIfNeeded</key>
	<true/>
</dict>
</plist>
`
