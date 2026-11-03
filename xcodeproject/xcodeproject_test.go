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
	xcodeprojmocks "github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj/mocks"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcworkspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minimalSchemeXML = `<?xml version="1.0" encoding="UTF-8"?>
<Scheme LastUpgradeVersion = "1240" version = "1.3">
</Scheme>
`

func testFactory(t *testing.T) Factory {
	logger := log.NewLogger()
	fileManager := fileutil.NewFileManager()
	plistFileHandler := plistutil.NewFileHandler(fileManager)
	pathModifier := pathutil.NewPathModifier()
	user := xcodeprojmocks.NewUserProvider(t)
	user.On("CurrentUserName").Return("bitrise", nil).Maybe()

	projectFactory := xcodeproj.NewFactory(logger, nil, fileManager, plistFileHandler, pathModifier, nil, user, xcodeproj.NewDirCreator())
	workspaceFactory := xcworkspace.NewFactory(logger, nil, fileManager, plistFileHandler, pathModifier, pathutil.NewPathChecker(), user, projectFactory)

	return NewFactory(projectFactory, workspaceFactory)
}

func writeFile(t *testing.T, pth, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(pth), 0755))
	require.NoError(t, os.WriteFile(pth, []byte(content), 0644))
}

// writeProject creates <dir>/App.xcodeproj from xcodeproj's minimal fixture, which has one app
// target, App, and returns its path.
func writeProject(t *testing.T, dir string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("xcodeproj", "testdata", "minimal.pbxproj"))
	require.NoError(t, err)

	pth := filepath.Join(dir, "App"+xcodeproj.XcodeProjExtension)
	writeFile(t, filepath.Join(pth, "project.pbxproj"), string(content))
	return pth
}

// writeWorkspace creates <dir>/Workspace.xcworkspace referencing App.xcodeproj next to it, with a shared
// scheme named Workspace, and returns its path.
func writeWorkspace(t *testing.T, dir string) string {
	t.Helper()
	pth := filepath.Join(dir, "Workspace"+xcworkspace.XCWorkspaceExtension)
	writeFile(t, filepath.Join(pth, "contents.xcworkspacedata"), `<?xml version="1.0" encoding="UTF-8"?>
<Workspace version = "1.0">
   <FileRef location = "group:App.xcodeproj"></FileRef>
</Workspace>
`)
	writeFile(t, filepath.Join(pth, "xcshareddata", "xcschemes", "Workspace.xcscheme"), minimalSchemeXML)
	return pth
}

func TestFactory_Open(t *testing.T) {
	dir := t.TempDir()
	projectPath := writeProject(t, dir)
	workspacePath := writeWorkspace(t, dir)

	container, err := testFactory(t).Open(projectPath)
	require.NoError(t, err)
	assert.Equal(t, projectPath, container.Path())
	assert.Equal(t, "App", container.Name())

	container, err = testFactory(t).Open(workspacePath)
	require.NoError(t, err)
	assert.Equal(t, workspacePath, container.Path())
	assert.Equal(t, "Workspace", container.Name(), "the workspace's own name, not its project's")

	_, err = testFactory(t).Open(filepath.Join(dir, "Package.swift"))
	assert.ErrorContains(t, err, "neither an Xcode project")

	_, err = testFactory(t).Open(filepath.Join(dir, "Missing.xcodeproj"))
	assert.ErrorIs(t, err, os.ErrNotExist)

	_, err = testFactory(t).Open(filepath.Join(dir, "Missing.xcworkspace"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestFactory_Scheme(t *testing.T) {
	dir := t.TempDir()
	projectPath := writeProject(t, dir)
	workspacePath := writeWorkspace(t, dir)

	tests := []struct {
		name              string
		pth               string
		schemeName        string
		wantContainerPath string
	}{
		{name: "project scheme", pth: projectPath, schemeName: "App", wantContainerPath: projectPath},
		{name: "workspace scheme", pth: workspacePath, schemeName: "Workspace", wantContainerPath: workspacePath},
		{name: "scheme of a referenced project", pth: workspacePath, schemeName: "App", wantContainerPath: projectPath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, containerPath, err := testFactory(t).Scheme(tt.pth, tt.schemeName)
			require.NoError(t, err)
			assert.Equal(t, tt.schemeName, scheme.Name)
			assert.Equal(t, tt.wantContainerPath, containerPath)
		})
	}
}

func TestFactory_Scheme_notFound(t *testing.T) {
	dir := t.TempDir()
	projectPath := writeProject(t, dir)
	workspacePath := writeWorkspace(t, dir)

	for _, pth := range []string{projectPath, workspacePath} {
		t.Run(filepath.Base(pth), func(t *testing.T) {
			_, _, err := testFactory(t).Scheme(pth, "Missing")
			assert.True(t, xcscheme.IsNotFoundError(err))
		})
	}
}

func TestFactory_Scheme_pathForms(t *testing.T) {
	dir := t.TempDir()
	projectPath := writeProject(t, dir)
	workspacePath := writeWorkspace(t, dir)

	t.Run("trailing separator", func(t *testing.T) {
		for pth, wantName := range map[string]string{projectPath + "/": "App", workspacePath + "/": "Workspace"} {
			container, err := testFactory(t).Open(pth)
			require.NoError(t, err)
			assert.Equal(t, filepath.Clean(pth), container.Path())
			assert.Equal(t, wantName, container.Name())

			scheme, containerPath, err := testFactory(t).Scheme(pth, "App")
			require.NoError(t, err)
			assert.Equal(t, "App", scheme.Name)
			assert.Equal(t, projectPath, containerPath)
		}
	})

	t.Run("neither a project nor a workspace", func(t *testing.T) {
		_, _, err := testFactory(t).Scheme(filepath.Join(dir, "Package.swift"), "App")
		assert.ErrorContains(t, err, "neither an Xcode project")
	})

	t.Run("missing from disk", func(t *testing.T) {
		_, _, err := testFactory(t).Scheme(filepath.Join(dir, "Missing.xcodeproj"), "App")
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}
