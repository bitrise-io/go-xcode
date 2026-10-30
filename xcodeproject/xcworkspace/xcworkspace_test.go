package xcworkspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	xcodeprojmocks "github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj/mocks"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcworkspace/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUserName = "bitrise"

// userProvider is a UserProvider that always names the given user.
func userProvider(t *testing.T, name string) *xcodeprojmocks.UserProvider {
	provider := xcodeprojmocks.NewUserProvider(t)
	provider.On("CurrentUserName").Return(name, nil).Maybe()
	return provider
}

func testFactory(buildSettings BuildSettingsProvider, user xcodeproj.UserProvider) Factory {
	logger := log.NewLogger()
	fileManager := fileutil.NewFileManager()
	plistFileHandler := plistutil.NewFileHandler(fileManager)
	pathModifier := pathutil.NewPathModifier()
	projectFactory := xcodeproj.NewFactory(logger, nil, fileManager, plistFileHandler, pathModifier, nil, user, xcodeproj.NewDirCreator())

	return NewFactory(logger, buildSettings, fileManager, plistFileHandler, pathModifier, pathutil.NewPathChecker(), user, projectFactory)
}

func writeFile(t *testing.T, pth, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(pth), 0755))
	require.NoError(t, os.WriteFile(pth, []byte(content), 0644))
}

// writeWorkspace creates <dir>/<name>.xcworkspace with the given contents and returns its path.
func writeWorkspace(t *testing.T, dir, name, contents string) string {
	t.Helper()
	pth := filepath.Join(dir, name+XCWorkspaceExtension)
	writeFile(t, filepath.Join(pth, contentsFileName), contents)
	return pth
}

// writeProject creates the project at pth from xcodeproj's minimal fixture, which has one app
// target, App.
func writeProject(t *testing.T, pth string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "xcodeproj", "testdata", "minimal.pbxproj"))
	require.NoError(t, err)
	writeFile(t, filepath.Join(pth, "project.pbxproj"), string(content))
}

func openWorkspace(t *testing.T, pth string) *Workspace {
	t.Helper()
	workspace, err := testFactory(nil, userProvider(t, testUserName)).Open(pth)
	require.NoError(t, err)
	return workspace
}

func TestIsWorkspace(t *testing.T) {
	assert.True(t, IsWorkspace("/a/b/App.xcworkspace"))
	assert.False(t, IsWorkspace("/a/b/App.xcodeproj"))
	assert.False(t, IsWorkspace("/a/b/App"))
}

func TestFactory_Open(t *testing.T) {
	workspacePath := writeWorkspace(t, t.TempDir(), "XcodeProj", nestedContents)

	workspace, err := testFactory(nil, userProvider(t, testUserName)).Open(workspacePath)
	require.NoError(t, err)

	assert.Equal(t, "XcodeProj", workspace.Name)
	assert.Equal(t, workspacePath, workspace.Path)
	require.Len(t, workspace.FileRefs, 1)
	assert.Equal(t, "group:XcodeProj.xcodeproj", workspace.FileRefs[0].Location)
	require.Len(t, workspace.Groups, 1)
	assert.Equal(t, "group:Group", workspace.Groups[0].Location)
}

func TestFactory_Open_missingWorkspace(t *testing.T) {
	_, err := testFactory(nil, userProvider(t, testUserName)).Open(filepath.Join(t.TempDir(), "Missing.xcworkspace"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestFactory_Open_malformedContents(t *testing.T) {
	workspacePath := writeWorkspace(t, t.TempDir(), "Broken", "<Workspace")

	_, err := testFactory(nil, userProvider(t, testUserName)).Open(workspacePath)
	assert.ErrorContains(t, err, "failed to unmarshal")
}

func TestWorkspace_FileLocations(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))

	locations, err := workspace.FileLocations()
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(dir, "XcodeProj.xcodeproj"),
		filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj"),
		filepath.Join(dir, "Group/SubProject/SubProjectTests/Info.plist"),
		filepath.Join(dir, "Group/SubProject/SubProject/ViewController.swift"),
		filepath.Join(dir, "Group/SubProject/SubProject/AppDelegate.swift"),
		filepath.Join(dir, "Group/SubProject/SubProject/Info.plist"),
		filepath.Join(dir, "Group/SubProject/SubProject/Assets.xcassets/Contents.json"),
		filepath.Join(dir, "Group/SubProject/SubProject/Assets.xcassets/AppIcon.appiconset/Contents.json"),
		filepath.Join(dir, "Group/SubProject/SubProject/Base.lproj/LaunchScreen.storyboard"),
		filepath.Join(dir, "Group/SubProject/SubProject/Base.lproj/Main.storyboard"),
	}, locations, "top-level files first, then each group's own files before its subgroups'")
}

func TestWorkspace_FileLocations_referenceTypes(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "Refs", `<?xml version="1.0" encoding="UTF-8"?>
<Workspace version = "1.0">
   <Group location = "group:Modules" name = "Modules">
      <FileRef location = "group:Feature/Feature.xcodeproj"></FileRef>
      <FileRef location = "container:Shared/Shared.xcodeproj"></FileRef>
      <FileRef location = "absolute:/opt/Pods/Pods.xcodeproj"></FileRef>
      <FileRef location = "absolute:Demo/Demo.xcodeproj"></FileRef>
   </Group>
</Workspace>`))

	locations, err := workspace.FileLocations()
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(dir, "Modules/Feature/Feature.xcodeproj"),
		filepath.Join(dir, "Shared/Shared.xcodeproj"),
		"/opt/Pods/Pods.xcodeproj",
		filepath.Join(dir, "Demo/Demo.xcodeproj"),
	}, locations, "container: paths and relative absolute: paths are relative to the workspace's directory even inside a group")
}

// Modelled on tuist/XcodeProj's Fixtures/Workspace.xcworkspace, which references the Simulator app
// from the Developer directory next to its projects.
func TestWorkspace_FileLocations_locationsOutsideTheWorkspaceAreLeftOut(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "Refs", `<?xml version="1.0" encoding="UTF-8"?>
<Workspace version = "1.0">
   <FileRef location = "developer:Applications/Simulator.app"></FileRef>
   <FileRef location = "self:"></FileRef>
   <Group location = "developer:Library" name = "Library">
      <FileRef location = "group:Frameworks/Frameworks.xcodeproj"></FileRef>
   </Group>
   <FileRef location = "group:App.xcodeproj"></FileRef>
</Workspace>`))

	locations, err := workspace.FileLocations()
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(dir, "App.xcodeproj")}, locations)
}

// Xcode writes <name>.xcodeproj/project.xcworkspace with a single self: reference to the project,
// and xcodebuild -list on that workspace reports the project's schemes.
func TestWorkspace_embeddedInProject(t *testing.T) {
	dir := t.TempDir()
	projectPath := filepath.Join(dir, "App.xcodeproj")
	writeProject(t, projectPath)
	workspace := openWorkspace(t, writeWorkspace(t, projectPath, "project", `<?xml version="1.0" encoding="UTF-8"?>
<Workspace version = "1.0">
   <FileRef location = "self:App.xcodeproj"></FileRef>
</Workspace>`))

	locations, err := workspace.ProjectFileLocations()
	require.NoError(t, err)
	assert.Equal(t, []string{projectPath}, locations)

	scheme, err := workspace.Scheme("App")
	require.NoError(t, err)
	assert.Equal(t, projectPath, scheme.ContainerPath, "the project's autocreated scheme")
}

func TestWorkspace_FileLocations_unknownReferenceType(t *testing.T) {
	workspace := openWorkspace(t, writeWorkspace(t, t.TempDir(), "Refs", `<?xml version="1.0" encoding="UTF-8"?>
<Workspace version = "1.0">
   <FileRef location = "bogus:Something"></FileRef>
</Workspace>`))

	_, err := workspace.FileLocations()
	assert.ErrorContains(t, err, "unknown file reference type")
}

func TestWorkspace_ProjectFileLocations(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))

	locations, err := workspace.ProjectFileLocations()
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(dir, "XcodeProj.xcodeproj"),
		filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj"),
	}, locations)
}

func TestWorkspace_Projects(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	subProjectPath := filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj")
	writeProject(t, subProjectPath)

	projects, err := workspace.Projects()
	require.NoError(t, err)

	require.Len(t, projects, 1, "the missing XcodeProj.xcodeproj is skipped")
	assert.Equal(t, subProjectPath, projects[0].Path)
}

func TestWorkspace_Projects_malformedProject(t *testing.T) {
	dir := t.TempDir()
	workspace := openWorkspace(t, writeWorkspace(t, dir, "XcodeProj", nestedContents))
	writeFile(t, filepath.Join(dir, "XcodeProj.xcodeproj", "project.pbxproj"), "not a project")

	_, err := workspace.Projects()
	assert.ErrorContains(t, err, "failed to open project")
}

func TestWorkspace_SchemeBuildSettings(t *testing.T) {
	workspacePath := writeWorkspace(t, t.TempDir(), "App", emptyContents)
	provider := mocks.NewBuildSettingsProvider(t)
	provider.On("SchemeBuildSettings", workspacePath, "App", "Release", []string{"-sdk", "iphoneos"}).
		Return(serialized.Object{"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.app"}, nil)

	workspace, err := testFactory(provider, userProvider(t, testUserName)).Open(workspacePath)
	require.NoError(t, err)

	settings, err := workspace.SchemeBuildSettings("App", "Release", "-sdk", "iphoneos")
	require.NoError(t, err)

	bundleID, ok := settings.String("PRODUCT_BUNDLE_IDENTIFIER")
	require.True(t, ok)
	assert.Equal(t, "io.bitrise.app", bundleID)
}

const emptyContents = `<?xml version="1.0" encoding="UTF-8"?>
<Workspace version = "1.0">
</Workspace>
`

// nestedContents references XcodeProj.xcodeproj next to the workspace and SubProject.xcodeproj
// through nested groups.
const nestedContents = `<?xml version="1.0" encoding="UTF-8"?>
<Workspace
   version = "1.0">
   <Group
      location = "group:Group"
      name = "Group">
      <Group
         location = "group:SubProject/SubProject"
         name = "SubProject">
         <FileRef
            location = "group:../SubProjectTests/Info.plist">
         </FileRef>
         <FileRef
            location = "group:ViewController.swift">
         </FileRef>
         <Group
            location = "group:Assets.xcassets"
            name = "Assets.xcassets">
            <Group
               location = "group:AppIcon.appiconset"
               name = "AppIcon.appiconset">
               <FileRef
                  location = "group:Contents.json">
               </FileRef>
            </Group>
            <FileRef
               location = "group:Contents.json">
            </FileRef>
         </Group>
         <Group
            location = "group:Base.lproj"
            name = "Base.lproj">
            <FileRef
               location = "group:LaunchScreen.storyboard">
            </FileRef>
            <FileRef
               location = "group:Main.storyboard">
            </FileRef>
         </Group>
         <FileRef
            location = "group:AppDelegate.swift">
         </FileRef>
         <FileRef
            location = "group:Info.plist">
         </FileRef>
      </Group>
      <FileRef
         location = "group:SubProject/SubProject.xcodeproj">
      </FileRef>
   </Group>
   <FileRef
      location = "group:XcodeProj.xcodeproj">
   </FileRef>
</Workspace>
`
