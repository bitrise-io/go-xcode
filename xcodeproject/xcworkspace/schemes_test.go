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

// namedContainer is a run of schemes of one container, reduced to what the assertions compare.
type namedContainer struct {
	path  string
	names []string
}

// namedContainers groups consecutive schemes by ContainerPath.
func namedContainers(schemes []xcscheme.Scheme) []namedContainer {
	var named []namedContainer
	for _, scheme := range schemes {
		if n := len(named); n == 0 || named[n-1].path != scheme.ContainerPath {
			named = append(named, namedContainer{path: scheme.ContainerPath})
		}
		named[len(named)-1].names = append(named[len(named)-1].names, scheme.Name)
	}
	return named
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

	require.Equal(t, []namedContainer{
		{path: workspace.Path, names: []string{"Shared", "Mine"}},
		{path: projectPath, names: []string{"App"}},
	}, namedContainers(schemes), "the workspace comes first with shared before user schemes, non-.xcscheme files are ignored, a project without schemes gets the autocreated ones, the missing SubProject.xcodeproj is left out")

	assert.Equal(t, sharedSchemePath(workspace.Path, "Shared"), schemes[0].Path)
	assert.True(t, schemes[0].IsShared)
	assert.False(t, schemes[1].IsShared)
	assert.True(t, schemes[2].IsShared, "autocreated schemes count as shared")
}

func TestWorkspace_Schemes_otherUsersSchemesAreNotVisible(t *testing.T) {
	workspacePath := writeWorkspace(t, t.TempDir(), "XcodeProj", emptyContents)
	workspace, err := testFactory(nil, userProvider(t, "someone-else")).Open(workspacePath)
	require.NoError(t, err)
	writeFile(t, sharedSchemePath(workspace.Path, "Shared"), minimalSchemeXML)
	writeFile(t, userSchemesPath(workspace, "Mine.xcscheme"), minimalSchemeXML)

	schemes, err := workspace.Schemes()
	require.NoError(t, err)
	assert.Equal(t, []namedContainer{{path: workspace.Path, names: []string{"Shared"}}}, namedContainers(schemes))
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
	assert.Equal(t, []namedContainer{{path: projectPath, names: []string{"ProjectScheme"}}}, namedContainers(schemes))
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
	writeFile(t, sharedSchemePath(workspace.Path, "Shared"), minimalSchemeXML)

	scheme, err := workspace.Scheme("Shared")
	require.NoError(t, err)
	assert.Equal(t, workspace.Path, scheme.ContainerPath)
	assert.Equal(t, "Shared", scheme.Name)

	scheme, err = workspace.Scheme("App")
	require.NoError(t, err)
	assert.Equal(t, projectPath, scheme.ContainerPath, "an autocreated scheme belongs to its project")
	assert.Equal(t, "App", scheme.Name)

	_, err = workspace.Scheme("Missing")
	assert.True(t, xcscheme.IsNotFoundError(err))
	assert.ErrorContains(t, err, "Missing")
}

// Scheme names are compared after Unicode normalisation: HFS+ stored file names decomposed
// (Gdańsk) while a user types the precomposed form (Gdańsk), and the two must match
// either way round.
func TestWorkspace_Scheme_unicodeNormalisation(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	projectPath := filepath.Join(dir, "XcodeProj.xcodeproj")
	writeProject(t, projectPath)
	writeFile(t, sharedSchemePath(projectPath, "Gdańsk"), minimalSchemeXML)
	writeFile(t, sharedSchemePath(workspace.Path, "Kraków"), minimalSchemeXML)

	scheme, err := workspace.Scheme("Gdańsk")
	require.NoError(t, err)
	assert.Equal(t, projectPath, scheme.ContainerPath)
	assert.Equal(t, "Gdańsk", norm.NFC.String(scheme.Name), "the precomposed name finds the decomposed file")

	scheme, err = workspace.Scheme("Kraków")
	require.NoError(t, err)
	assert.Equal(t, workspace.Path, scheme.ContainerPath)
	assert.Equal(t, "Kraków", norm.NFC.String(scheme.Name), "the decomposed name finds the precomposed file")
}

func TestWorkspace_Scheme_stopsAtTheFirstMatch(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	writeFile(t, sharedSchemePath(workspace.Path, "App"), minimalSchemeXML)
	writeFile(t, filepath.Join(dir, "XcodeProj.xcodeproj", "project.pbxproj"), "not a project")

	scheme, err := workspace.Scheme("App")
	require.NoError(t, err, "the project after the match is never opened")
	assert.Equal(t, workspace.Path, scheme.ContainerPath)

	_, err = workspace.Schemes()
	assert.ErrorContains(t, err, "failed to open project", "listing everything still reaches it")
}

func TestWorkspace_Scheme_workspaceSchemesComeFirst(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	writeProject(t, filepath.Join(dir, "XcodeProj.xcodeproj"))
	writeFile(t, sharedSchemePath(workspace.Path, "App"), minimalSchemeXML)

	scheme, err := workspace.Scheme("App")
	require.NoError(t, err)
	assert.Equal(t, workspace.Path, scheme.ContainerPath, "the workspace's App wins over the project's autocreated App")
}
