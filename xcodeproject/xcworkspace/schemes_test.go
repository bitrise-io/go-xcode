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

func namesByContainer(schemes map[string][]xcscheme.Scheme) map[string][]string {
	names := map[string][]string{}
	for container, containerSchemes := range schemes {
		names[container] = schemeNames(containerSchemes)
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

	require.Equal(t, map[string][]string{
		workspace.Path: {"Shared", "Mine"},
		projectPath:    {"App"},
	}, namesByContainer(schemes), "shared schemes come before user schemes, non-.xcscheme files are ignored, a project without schemes gets the autocreated ones, the missing SubProject.xcodeproj is left out")

	assert.Equal(t, sharedSchemePath(workspace.Path, "Shared"), schemes[workspace.Path][0].Path)
	assert.True(t, schemes[workspace.Path][0].IsShared)
	assert.False(t, schemes[workspace.Path][1].IsShared)
	assert.True(t, schemes[projectPath][0].IsShared, "autocreated schemes count as shared")
}

func TestWorkspace_Schemes_otherUsersSchemesAreNotVisible(t *testing.T) {
	workspacePath := writeWorkspace(t, t.TempDir(), "XcodeProj", emptyContents)
	workspace, err := testFactory(nil, userProvider(t, "someone-else")).Open(workspacePath)
	require.NoError(t, err)
	writeFile(t, sharedSchemePath(workspace.Path, "Shared"), minimalSchemeXML)
	writeFile(t, userSchemesPath(workspace, "Mine.xcscheme"), minimalSchemeXML)

	schemes, err := workspace.Schemes()
	require.NoError(t, err)
	assert.Equal(t, map[string][]string{workspace.Path: {"Shared"}}, namesByContainer(schemes))
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

// xcodebuild warns about a settings file it cannot read and still lists the schemes, autocreating
// them as by default (checked with xcodebuild 26.5 -list).
func TestWorkspace_Schemes_unusableAutocreateSettingFallsBackToTheDefault(t *testing.T) {
	settingsPath := func(workspace *Workspace) string {
		return filepath.Join(workspace.Path, "xcshareddata", "WorkspaceSettings.xcsettings")
	}

	tests := []struct {
		name  string
		write func(t *testing.T, workspace *Workspace)
	}{
		{name: "not a plist", write: func(t *testing.T, workspace *Workspace) {
			writeFile(t, settingsPath(workspace), "not a plist")
		}},
		{name: "not a boolean", write: func(t *testing.T, workspace *Workspace) {
			writeAutocreateSetting(t, workspace, "<string>NO</string>")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
			projectPath := filepath.Join(dir, "XcodeProj.xcodeproj")
			writeProject(t, projectPath)
			subProjectPath := filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj")
			writeProject(t, subProjectPath)
			writeFile(t, sharedSchemePath(subProjectPath, "SubProjectScheme"), minimalSchemeXML)
			tt.write(t, workspace)

			schemes, err := workspace.Schemes()
			require.NoError(t, err)
			assert.Equal(t, map[string][]string{
				projectPath:    {"App"},
				subProjectPath: {"SubProjectScheme"},
			}, namesByContainer(schemes), "a project without schemes gets the autocreated ones, a project's own schemes are listed")
		})
	}
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

	scheme, container, err := workspace.Scheme("Shared")
	require.NoError(t, err)
	assert.Equal(t, workspace.Path, container)
	assert.Equal(t, "Shared", scheme.Name)

	scheme, container, err = workspace.Scheme("App")
	require.NoError(t, err)
	assert.Equal(t, projectPath, container, "an autocreated scheme belongs to its project")
	assert.Equal(t, "App", scheme.Name)

	_, _, err = workspace.Scheme("Missing")
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

	scheme, container, err := workspace.Scheme("Gdańsk")
	require.NoError(t, err)
	assert.Equal(t, projectPath, container)
	assert.Equal(t, "Gdańsk", norm.NFC.String(scheme.Name), "the precomposed name finds the decomposed file")

	scheme, container, err = workspace.Scheme("Kraków")
	require.NoError(t, err)
	assert.Equal(t, workspace.Path, container)
	assert.Equal(t, "Kraków", norm.NFC.String(scheme.Name), "the decomposed name finds the precomposed file")
}

func TestWorkspace_Scheme_stopsAtTheFirstMatch(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	writeFile(t, sharedSchemePath(workspace.Path, "App"), minimalSchemeXML)
	writeFile(t, filepath.Join(dir, "XcodeProj.xcodeproj", "project.pbxproj"), "not a project")

	_, container, err := workspace.Scheme("App")
	require.NoError(t, err, "the project after the match is never opened")
	assert.Equal(t, workspace.Path, container)

	_, err = workspace.Schemes()
	assert.ErrorContains(t, err, "failed to open project", "listing everything still reaches it")
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
