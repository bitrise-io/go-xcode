package xcworkspace

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/norm"
)

const minimalSchemeXML = `<?xml version="1.0" encoding="UTF-8"?>
<Scheme LastUpgradeVersion = "1240" version = "1.3">
</Scheme>
`

func sharedSchemePath(containerPath, name string) string {
	return filepath.Join(containerPath, "xcshareddata", "xcschemes", name+".xcscheme")
}

func userSchemesPath(workspace *Workspace, file string) string {
	return filepath.Join(workspace.Path, "xcuserdata", testUserName+".xcuserdatad", "xcschemes", file)
}

func writeAutocreateSetting(t *testing.T, workspace *Workspace, value string) {
	t.Helper()
	writeFile(t, filepath.Join(workspace.Path, "xcshareddata", "WorkspaceSettings.xcsettings"),
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

func TestWorkspace_Schemes(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	projectPath := filepath.Join(dir, "XcodeProj.xcodeproj")
	writeProject(t, projectPath)
	writeFile(t, sharedSchemePath(workspace.Path, "Shared"), minimalSchemeXML)
	writeFile(t, userSchemesPath(workspace, "Mine.xcscheme"), minimalSchemeXML)
	writeFile(t, userSchemesPath(workspace, "notes.txt"), "not a scheme")

	schemes, err := workspace.Schemes()
	require.NoError(t, err)

	require.Len(t, schemes, 2, "the missing SubProject.xcodeproj holds no schemes")

	require.Equal(t, []string{"Shared", "Mine"}, schemeNames(schemes[workspace.Path]), "shared schemes come first; non-.xcscheme files are ignored")
	assert.True(t, schemes[workspace.Path][0].IsShared)
	assert.False(t, schemes[workspace.Path][1].IsShared)
	assert.Equal(t, sharedSchemePath(workspace.Path, "Shared"), schemes[workspace.Path][0].Path)

	require.Equal(t, []string{"App"}, schemeNames(schemes[projectPath]), "a project without schemes gets the autocreated ones")
	assert.True(t, schemes[projectPath][0].IsShared)
}

func TestWorkspace_Schemes_otherUsersSchemesAreNotVisible(t *testing.T) {
	workspacePath := writeWorkspace(t, t.TempDir(), "XcodeProj", emptyContents)
	workspace, err := testFactory(nil, staticUser("someone-else")).Open(workspacePath)
	require.NoError(t, err)
	writeFile(t, sharedSchemePath(workspace.Path, "Shared"), minimalSchemeXML)
	writeFile(t, userSchemesPath(workspace, "Mine.xcscheme"), minimalSchemeXML)

	schemes, err := workspace.Schemes()
	require.NoError(t, err)
	assert.Equal(t, []string{"Shared"}, schemeNames(schemes[workspace.Path]))
}

func TestWorkspace_Schemes_autocreateOff(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	writeProject(t, filepath.Join(dir, "XcodeProj.xcodeproj"))
	writeAutocreateSetting(t, workspace, "<false/>")

	schemes, err := workspace.Schemes()
	require.NoError(t, err)
	assert.Empty(t, schemes, "a project without schemes yields none when autocreate is off")
}

func TestWorkspace_Schemes_autocreateOffKeepsExistingProjectSchemes(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	projectPath := filepath.Join(dir, "XcodeProj.xcodeproj")
	writeProject(t, projectPath)
	writeFile(t, sharedSchemePath(projectPath, "ProjectScheme"), minimalSchemeXML)
	writeAutocreateSetting(t, workspace, "<false/>")

	schemes, err := workspace.Schemes()
	require.NoError(t, err)
	assert.Equal(t, map[string][]string{projectPath: {"ProjectScheme"}}, namesByContainer(schemes))
}

func TestWorkspace_Schemes_autocreateSettingThatIsNotABoolean(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	writeProject(t, filepath.Join(dir, "XcodeProj.xcodeproj"))
	writeAutocreateSetting(t, workspace, "<string>YES</string>")

	_, err := workspace.Schemes()
	assert.ErrorContains(t, err, "not a boolean")
}

func TestWorkspace_Schemes_settingsNotReadWithoutProjects(t *testing.T) {
	workspace := openWorkspace(t, writeWorkspace(t, t.TempDir(), "XcodeProj", emptyContents))
	writeFile(t, filepath.Join(workspace.Path, "xcshareddata", "WorkspaceSettings.xcsettings"), "not a plist")

	schemes, err := workspace.Schemes()
	require.NoError(t, err)
	assert.Empty(t, schemes)
}

func TestWorkspace_Schemes_malformedSchemeFile(t *testing.T) {
	workspace := openWorkspace(t, writeWorkspace(t, t.TempDir(), "XcodeProj", emptyContents))
	writeFile(t, sharedSchemePath(workspace.Path, "Broken"), "<Scheme")

	_, err := workspace.Schemes()
	assert.ErrorContains(t, err, "failed to read shared schemes")
}

func TestWorkspace_Scheme(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	projectPath := filepath.Join(dir, "XcodeProj.xcodeproj")
	writeProject(t, projectPath)
	subProjectPath := filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj")
	writeProject(t, subProjectPath)
	writeFile(t, sharedSchemePath(workspace.Path, "Shared"), minimalSchemeXML)
	// Gdańsk with a decomposed ń, as older macOS versions stored file names.
	writeFile(t, sharedSchemePath(subProjectPath, "Gdańsk"), minimalSchemeXML)

	scheme, container, err := workspace.Scheme("Shared")
	require.NoError(t, err)
	assert.Equal(t, workspace.Path, container)
	assert.Equal(t, "Shared", scheme.Name)

	scheme, container, err = workspace.Scheme("App")
	require.NoError(t, err)
	assert.Equal(t, projectPath, container, "an autocreated scheme belongs to its project")
	assert.Equal(t, "App", scheme.Name)

	scheme, container, err = workspace.Scheme("Gdańsk")
	require.NoError(t, err)
	assert.Equal(t, subProjectPath, container)
	assert.Equal(t, "Gdańsk", norm.NFC.String(scheme.Name), "the precomposed name finds the decomposed file")

	_, _, err = workspace.Scheme("Missing")
	assert.True(t, xcscheme.IsNotFoundError(err))
	assert.ErrorContains(t, err, "Missing")
}

func TestWorkspace_Scheme_workspaceSchemesComeFirst(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	writeProject(t, filepath.Join(dir, "XcodeProj.xcodeproj"))
	writeFile(t, sharedSchemePath(workspace.Path, "App"), minimalSchemeXML)

	_, container, err := workspace.Scheme("App")
	require.NoError(t, err)
	assert.Equal(t, workspace.Path, container, "the workspace's App wins over the project's autocreated App")
}

func namesByContainer(schemes map[string][]xcscheme.Scheme) map[string][]string {
	names := map[string][]string{}
	for container, containerSchemes := range schemes {
		names[container] = schemeNames(containerSchemes)
	}
	return names
}
