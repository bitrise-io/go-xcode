package xcodeproject

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcworkspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFactory() Factory {
	logger := log.NewLogger()
	fileManager := fileutil.NewFileManager()
	plistFileHandler := plistutil.NewFileHandler(fileManager)
	pathModifier := pathutil.NewPathModifier()
	userProvider := xcodeproj.NewUserProvider()
	projectFactory := xcodeproj.NewFactory(logger, nil, fileManager, plistFileHandler, pathModifier, nil, userProvider, xcodeproj.NewDirCreator())
	workspaceFactory := xcworkspace.NewFactory(logger, nil, fileManager, plistFileHandler, pathModifier, pathutil.NewPathChecker(), userProvider, projectFactory)

	return NewFactory(projectFactory, workspaceFactory)
}

func writeFile(t *testing.T, pth, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(pth), 0755))
	require.NoError(t, os.WriteFile(pth, []byte(content), 0644))
}

// writeProject creates <dir>/<name>.xcodeproj from xcodeproj's minimal fixture, which has one app
// target, App, and returns its path.
func writeProject(t *testing.T, dir, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("xcodeproj", "testdata", "minimal.pbxproj"))
	require.NoError(t, err)

	pth := filepath.Join(dir, name+xcodeproj.XcodeProjExtension)
	writeFile(t, filepath.Join(pth, "project.pbxproj"), string(content))
	return pth
}

// writeWorkspace creates <dir>/App.xcworkspace referencing App.xcodeproj and returns its path.
func writeWorkspace(t *testing.T, dir string) string {
	t.Helper()
	pth := filepath.Join(dir, "App.xcworkspace")
	writeFile(t, filepath.Join(pth, "contents.xcworkspacedata"), `<?xml version="1.0" encoding="UTF-8"?>
<Workspace version = "1.0">
   <FileRef location = "group:App.xcodeproj"></FileRef>
</Workspace>`)
	return pth
}

const minimalSchemeXML = `<?xml version="1.0" encoding="UTF-8"?>
<Scheme LastUpgradeVersion = "1240" version = "1.3">
</Scheme>
`

func TestFactory_Open(t *testing.T) {
	dir := t.TempDir()
	projectPath := writeProject(t, dir, "App")
	workspacePath := writeWorkspace(t, dir)

	container, err := testFactory().Open(projectPath)
	require.NoError(t, err)
	assert.Equal(t, projectPath, container.Path())
	assert.Equal(t, "App", container.Name())

	container, err = testFactory().Open(workspacePath)
	require.NoError(t, err)
	assert.Equal(t, workspacePath, container.Path())
	assert.Equal(t, "App", container.Name())

	_, err = testFactory().Open(filepath.Join(dir, "Package.swift"))
	assert.ErrorContains(t, err, "neither an Xcode project")

	_, err = testFactory().Open(filepath.Join(dir, "Missing.xcodeproj"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestContainer_Schemes(t *testing.T) {
	dir := t.TempDir()
	projectPath := writeProject(t, dir, "App")
	workspacePath := writeWorkspace(t, dir)
	writeFile(t, filepath.Join(workspacePath, "xcshareddata", "xcschemes", "Shared.xcscheme"), minimalSchemeXML)

	container, err := testFactory().Open(projectPath)
	require.NoError(t, err)
	schemes, err := container.Schemes()
	require.NoError(t, err)
	require.Len(t, schemes, 1, "the project's autocreated scheme")
	assert.Equal(t, "App", schemes[0].Name)
	assert.Equal(t, projectPath, schemes[0].ContainerPath)

	container, err = testFactory().Open(workspacePath)
	require.NoError(t, err)
	schemes, err = container.Schemes()
	require.NoError(t, err)
	require.Len(t, schemes, 2, "the workspace's own scheme, then the project's")
	assert.Equal(t, "Shared", schemes[0].Name)
	assert.Equal(t, workspacePath, schemes[0].ContainerPath)
	assert.Equal(t, "App", schemes[1].Name)
	assert.Equal(t, projectPath, schemes[1].ContainerPath)
}

func TestFactory_Scheme(t *testing.T) {
	dir := t.TempDir()
	projectPath := writeProject(t, dir, "App")
	workspacePath := writeWorkspace(t, dir)

	scheme, err := testFactory().Scheme(projectPath, "App")
	require.NoError(t, err)
	assert.Equal(t, "App", scheme.Name)
	assert.Equal(t, projectPath, scheme.ContainerPath)

	scheme, err = testFactory().Scheme(workspacePath, "App")
	require.NoError(t, err)
	assert.Equal(t, "App", scheme.Name)
	assert.Equal(t, projectPath, scheme.ContainerPath, "a scheme found through a workspace still names the project holding it")

	_, err = testFactory().Scheme(workspacePath, "Missing")
	assert.True(t, xcscheme.IsNotFoundError(err))
}

func TestContainer_Projects(t *testing.T) {
	dir := t.TempDir()
	projectPath := writeProject(t, dir, "App")
	workspacePath := writeWorkspace(t, dir)

	container, err := testFactory().Open(projectPath)
	require.NoError(t, err)
	projects, err := container.Projects()
	require.NoError(t, err)
	require.Len(t, projects, 1, "a project is its own only project")
	assert.Equal(t, projectPath, projects[0].Path)

	container, err = testFactory().Open(workspacePath)
	require.NoError(t, err)
	projects, err = container.Projects()
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, projectPath, projects[0].Path)

	again, err := container.Projects()
	require.NoError(t, err)
	assert.Same(t, projects[0], again[0], "listed once, kept")
}

func TestContainer_Project(t *testing.T) {
	dir := t.TempDir()
	projectPath := writeProject(t, dir, "App")
	libraryPath := writeProject(t, dir, "Library")
	container, err := testFactory().Open(writeWorkspace(t, dir))
	require.NoError(t, err)

	projects, err := container.Projects()
	require.NoError(t, err)

	project, err := container.Project(projectPath)
	require.NoError(t, err)
	assert.Same(t, projects[0], project, "a referenced project is the one Projects opened")

	library, err := container.Project(libraryPath)
	require.NoError(t, err)
	assert.Equal(t, libraryPath, library.Path, "a project the workspace does not reference is opened on demand")
	libraryAgain, err := container.Project(libraryPath)
	require.NoError(t, err)
	assert.Same(t, library, libraryAgain, "and kept")

	projects, err = container.Projects()
	require.NoError(t, err)
	assert.Len(t, projects, 1, "Projects stays what the bundle references")

	_, err = container.Project(filepath.Join(dir, "Missing.xcodeproj"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}
