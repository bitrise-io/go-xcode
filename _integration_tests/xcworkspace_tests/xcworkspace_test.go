package xcworkspace_tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/_integration_tests"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	"github.com/bitrise-io/go-xcode/v2/xcodebuild"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcworkspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/norm"
)

// The sample workspace references one project next to it and one through nested groups.
const xcodeProjectTestRepoURL = "https://github.com/bitrise-samples/xcode-project-test.git"

func TestWorkspace_Schemes(t *testing.T) {
	dir := _integration_tests.GetRepository(t, xcodeProjectTestRepoURL, "master")
	workspace, err := newFactory().Open(filepath.Join(dir, "XcodeProj.xcworkspace"))
	require.NoError(t, err)

	projectPaths, err := workspace.ProjectFileLocations()
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(dir, "XcodeProj.xcodeproj"),
		filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj"),
	}, projectPaths)

	schemes, err := workspace.Schemes()
	require.NoError(t, err)
	require.Equal(t, []string{
		"WorkspaceScheme",
		"Gdańsk",
		"ProjectScheme",
		"ProjectTodayExtensionScheme",
		"SubProjectScheme",
		"SubProjectTestScheme",
		"SubProjectWatchKitApp (Notification) Scheme",
		"SubProjectWatchKitAppScheme",
	}, normalised(schemeNames(schemes)), "the workspace's own scheme first, then its projects' in file order")

	assert.Equal(t, workspace.Path, schemes[0].ContainerPath)
	assert.Equal(t, filepath.Join(dir, "XcodeProj.xcodeproj"), schemes[1].ContainerPath)
	assert.Equal(t, filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj"), schemes[4].ContainerPath)
}

// Checked against xcodebuild 26.5: with a referenced project removed, -list exits 0 and lists only the
// other projects' schemes and the workspace's own, and -showBuildSettings for one of the missing
// project's schemes fails with "does not contain a scheme named".
func TestWorkspace_Schemes_missingProject(t *testing.T) {
	dir := _integration_tests.GetRepository(t, xcodeProjectTestRepoURL, "master")

	// The clone is shared with the other tests, which run sequentially; put the project back afterwards.
	subProjectPath := filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj")
	hiddenPath := subProjectPath + ".missing"
	require.NoError(t, os.Rename(subProjectPath, hiddenPath))
	t.Cleanup(func() {
		require.NoError(t, os.Rename(hiddenPath, subProjectPath))
	})

	workspace, err := newFactory().Open(filepath.Join(dir, "XcodeProj.xcworkspace"))
	require.NoError(t, err)

	schemes, err := workspace.Schemes()
	require.NoError(t, err)
	assert.Equal(t, []string{"WorkspaceScheme", "Gda\u0144sk", "ProjectScheme", "ProjectTodayExtensionScheme"}, normalised(schemeNames(schemes)))
	assert.Equal(t, workspace.Path, schemes[0].ContainerPath)
	assert.Equal(t, filepath.Join(dir, "XcodeProj.xcodeproj"), schemes[1].ContainerPath)

	_, err = workspace.Scheme("SubProjectScheme")
	assert.True(t, xcscheme.IsNotFoundError(err))
}

func TestWorkspace_Scheme(t *testing.T) {
	dir := _integration_tests.GetRepository(t, xcodeProjectTestRepoURL, "master")
	workspace, err := newFactory().Open(filepath.Join(dir, "XcodeProj.xcworkspace"))
	require.NoError(t, err)

	scheme, err := workspace.Scheme("SubProjectWatchKitApp (Notification) Scheme")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "Group/SubProject/SubProject.xcodeproj"), scheme.ContainerPath)
	assert.Equal(t, "SubProjectWatchKitApp (Notification) Scheme", scheme.Name)

	// Gdańsk as High Sierra (precomposed ń) and Mojave (decomposed ń) represent it.
	for _, name := range []string{"Gdańsk", "Gdańsk"} {
		scheme, err := workspace.Scheme(name)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, "XcodeProj.xcodeproj"), scheme.ContainerPath)
		assert.Equal(t, norm.NFC.String(name), norm.NFC.String(scheme.Name))
	}

	_, err = workspace.Scheme("Not Exist Scheme")
	assert.EqualError(t, err, "scheme Not Exist Scheme not found in XcodeProj")
}

func TestWorkspace_SchemeBuildSettings(t *testing.T) {
	dir := _integration_tests.GetRepository(t, xcodeProjectTestRepoURL, "master")
	workspace, err := newFactory().Open(filepath.Join(dir, "XcodeProj.xcworkspace"))
	require.NoError(t, err)

	settings, err := workspace.SchemeBuildSettings("ProjectScheme", "Release")
	require.NoError(t, err)

	bundleID, ok := settings.String("PRODUCT_BUNDLE_IDENTIFIER")
	require.True(t, ok)
	assert.NotEmpty(t, bundleID)
}

const (
	// A CocoaPods layout: the app project and Pods/Pods.xcodeproj, both group: references.
	cocoaPodsSampleRepoURL = "https://github.com/bitrise-io/sample-apps-ios-workspace-swift.git"
	// A workspace holding a single top-level container: reference.
	containerSampleRepoURL = "https://github.com/bitrise-io/sample-swift-project-with-parallel-ui-test.git"
)

func TestWorkspace_cocoaPodsLayout(t *testing.T) {
	dir := _integration_tests.GetRepository(t, cocoaPodsSampleRepoURL, "master")
	workspace, err := newFactory().Open(filepath.Join(dir, "sample-apps-ios-workspace-swift.xcworkspace"))
	require.NoError(t, err)

	appProjectPath := filepath.Join(dir, "sample-apps-ios-workspace-swift.xcodeproj")
	projectPaths, err := workspace.ProjectFileLocations()
	require.NoError(t, err)
	assert.Equal(t, []string{appProjectPath, filepath.Join(dir, "Pods/Pods.xcodeproj")}, projectPaths)

	scheme, err := workspace.Scheme("sample-apps-ios-workspace-swift")
	require.NoError(t, err)
	assert.Equal(t, appProjectPath, scheme.ContainerPath)
	assert.True(t, scheme.IsShared)

	assert.ElementsMatch(t, xcodebuildSchemes(t, workspace.Path), allSchemeNames(t, workspace),
		"the Pods project has no scheme files, so its schemes are autocreated the way xcodebuild does")
}

func TestWorkspace_containerReference(t *testing.T) {
	dir := _integration_tests.GetRepository(t, containerSampleRepoURL, "master")
	workspace, err := newFactory().Open(filepath.Join(dir, "BullsEye.xcworkspace"))
	require.NoError(t, err)

	projectPath := filepath.Join(dir, "BullsEye.xcodeproj")
	projectPaths, err := workspace.ProjectFileLocations()
	require.NoError(t, err)
	assert.Equal(t, []string{projectPath}, projectPaths)

	scheme, err := workspace.Scheme("BullsEye")
	require.NoError(t, err)
	assert.Equal(t, projectPath, scheme.ContainerPath)

	assert.ElementsMatch(t, xcodebuildSchemes(t, workspace.Path), allSchemeNames(t, workspace))
}

// xcodebuildSchemes returns the schemes xcodebuild -list reports for the workspace.
func xcodebuildSchemes(t *testing.T, workspacePath string) []string {
	t.Helper()

	cmd := command.NewFactory(env.NewRepository()).Create("xcodebuild", []string{"-workspace", workspacePath, "-list", "-json"}, nil)
	out, err := cmd.RunAndReturnTrimmedOutput()
	require.NoError(t, err, out)

	var list struct {
		Workspace struct {
			Schemes []string `json:"schemes"`
		} `json:"workspace"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &list), out)

	return normalised(list.Workspace.Schemes)
}

// allSchemeNames flattens Schemes() over its containers.
func allSchemeNames(t *testing.T, workspace *xcworkspace.Workspace) []string {
	t.Helper()

	schemes, err := workspace.Schemes()
	require.NoError(t, err)

	return normalised(schemeNames(schemes))
}

func newFactory() xcworkspace.Factory {
	logger := log.NewLogger()
	buildSettings := xcodebuild.NewShowBuildSettingsProvider(command.NewFactory(env.NewRepository()), logger)
	fileManager := fileutil.NewFileManager()
	plistFileHandler := plistutil.NewFileHandler(fileManager)
	pathModifier := pathutil.NewPathModifier()
	userProvider := xcodeproj.NewUserProvider()
	projectFactory := xcodeproj.NewFactory(logger, buildSettings, fileManager, plistFileHandler, pathModifier, pathutil.NewPathProvider(), userProvider, xcodeproj.NewDirCreator())

	return xcworkspace.NewFactory(logger, buildSettings, fileManager, plistFileHandler, pathModifier, pathutil.NewPathChecker(), userProvider, projectFactory)
}

func schemeNames(schemes []xcscheme.Scheme) []string {
	var names []string
	for _, scheme := range schemes {
		names = append(names, scheme.Name)
	}
	return names
}

func normalised(names []string) []string {
	var out []string
	for _, name := range names {
		out = append(out, norm.NFC.String(name))
	}
	return out
}
