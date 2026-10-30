package xcodeproject_tests

import (
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/_integration_tests"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcworkspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The sample workspace holds a scheme of its own and references two projects, one through groups.
const xcodeProjectTestRepoURL = "https://github.com/bitrise-samples/xcode-project-test.git"

func TestContainer_workspace(t *testing.T) {
	dir := _integration_tests.GetRepository(t, xcodeProjectTestRepoURL, "master")
	workspacePath := filepath.Join(dir, "XcodeProj.xcworkspace")
	projectPath := filepath.Join(dir, "XcodeProj.xcodeproj")
	subProjectPath := filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj")

	container, err := newFactory().Open(workspacePath)
	require.NoError(t, err)
	assert.Equal(t, workspacePath, container.Path())
	assert.Equal(t, "XcodeProj", container.Name())

	projects, err := container.Projects()
	require.NoError(t, err)
	require.Len(t, projects, 2)
	assert.Equal(t, projectPath, projects[0].Path)
	assert.Equal(t, subProjectPath, projects[1].Path)

	schemes, err := container.Schemes()
	require.NoError(t, err)
	require.Len(t, schemes, 8)
	assert.Equal(t, "WorkspaceScheme", schemes[0].Name)
	assert.Equal(t, workspacePath, schemes[0].ContainerPath)

	scheme, err := container.Scheme("SubProjectScheme")
	require.NoError(t, err)
	assert.Equal(t, subProjectPath, scheme.ContainerPath)

	subProject, err := container.Project(scheme.ContainerPath)
	require.NoError(t, err)
	assert.Same(t, projects[1], subProject, "the scheme's project is the one Projects opened")
}

func TestContainer_project(t *testing.T) {
	dir := _integration_tests.GetRepository(t, xcodeProjectTestRepoURL, "master")
	projectPath := filepath.Join(dir, "XcodeProj.xcodeproj")

	container, err := newFactory().Open(projectPath)
	require.NoError(t, err)
	assert.Equal(t, "XcodeProj", container.Name())

	projects, err := container.Projects()
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, projectPath, projects[0].Path)

	schemes, err := container.Schemes()
	require.NoError(t, err)
	assert.Len(t, schemes, 3)
}

func newFactory() xcodeproject.Factory {
	logger := log.NewLogger()
	fileManager := fileutil.NewFileManager()
	plistFileHandler := plistutil.NewFileHandler(fileManager)
	pathModifier := pathutil.NewPathModifier()
	userProvider := xcodeproj.NewUserProvider()
	projectFactory := xcodeproj.NewFactory(logger, nil, fileManager, plistFileHandler, pathModifier, pathutil.NewPathProvider(), userProvider, xcodeproj.NewDirCreator())
	workspaceFactory := xcworkspace.NewFactory(logger, nil, fileManager, plistFileHandler, pathModifier, pathutil.NewPathChecker(), userProvider, projectFactory)

	return xcodeproject.NewFactory(projectFactory, workspaceFactory)
}
