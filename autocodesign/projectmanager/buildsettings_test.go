package projectmanager

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/v2/fileutil"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
	"github.com/bitrise-io/go-xcode/v2/plistutil"
	plistmocks "github.com/bitrise-io/go-xcode/v2/plistutil/mocks"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj"
	xcodeprojmocks "github.com/bitrise-io/go-xcode/v2/xcodeproject/xcodeproj/mocks"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcworkspace"
	xcworkspacemocks "github.com/bitrise-io/go-xcode/v2/xcodeproject/xcworkspace/mocks"
	"github.com/stretchr/testify/require"
)

// These tests pin down the xcodebuild calls a settings read makes, so a change to the helper
// that would alter them is visible without running xcodebuild. The plan, for a workspace: every
// target is read from the workspace, by a scheme of the target's name; when that fails, the
// target is read from the project; in compat mode both are read. A project is only ever read
// from the project.

const testProjectPath = "/repo/App.xcodeproj"

var testOptions = []string{"-skipMacroValidation", "-onlyUsePackageVersionsFromResolvedFile"}

// settingsHelper builds a ProjectHelper over xcodeproj's minimal fixture, whose target settings
// come from projectSettings. Given workspaceSettings, the helper also holds a workspace, written
// to a temp dir, whose scheme settings come from workspaceSettings; its path is XcWorkspace.Path.
func settingsHelper(t *testing.T, workspaceSettings xcworkspace.BuildSettingsProvider, compat bool, projectSettings xcodeproj.BuildSettingsProvider) *ProjectHelper {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", "xcodeproject", "xcodeproj", "testdata", "minimal.pbxproj"))
	require.NoError(t, err)

	logger := log.NewLogger()
	fileManager := fileutil.NewFileManager()
	plistHandler := plistutil.NewFileHandler(fileManager)
	pathModifier := pathutil.NewPathModifier()
	userProvider := xcodeproj.NewUserProvider()
	projectFactory := xcodeproj.NewFactory(logger, projectSettings, fileManager, plistHandler, pathModifier, pathutil.NewPathProvider(), userProvider, xcodeproj.NewDirCreator())
	project, err := projectFactory.Parse(content, testProjectPath)
	require.NoError(t, err)

	var workspace *xcworkspace.Workspace
	if workspaceSettings != nil {
		workspacePath := filepath.Join(t.TempDir(), "App.xcworkspace")
		require.NoError(t, os.MkdirAll(workspacePath, 0o755))
		contents := `<?xml version="1.0" encoding="UTF-8"?>
<Workspace version = "1.0">
   <FileRef location = "group:App.xcodeproj"></FileRef>
</Workspace>
`
		require.NoError(t, os.WriteFile(filepath.Join(workspacePath, "contents.xcworkspacedata"), []byte(contents), 0o644))
		workspaceFactory := xcworkspace.NewFactory(logger, workspaceSettings, fileManager, plistHandler, pathModifier, pathutil.NewPathChecker(), userProvider, projectFactory)
		workspace, err = workspaceFactory.Open(workspacePath)
		require.NoError(t, err)
	}

	return &ProjectHelper{
		logger:                      logger,
		MainTarget:                  xcodeproj.Target{Name: "App"},
		XcWorkspace:                 workspace,
		XcProj:                      project,
		Configuration:               "Release",
		additionalXcodebuildOptions: testOptions,
		isCompatMode:                compat,
	}
}

func TestFetchBuildSettings_workspace(t *testing.T) {
	appSettings := serialized.Object{"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.app"}
	extensionSettings := serialized.Object{"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.app.extension"}

	workspaceSettings := xcworkspacemocks.NewBuildSettingsProvider(t)
	projectSettings := xcodeprojmocks.NewBuildSettingsProvider(t) // never asked
	helper := settingsHelper(t, workspaceSettings, false, projectSettings)
	workspaceSettings.On("SchemeBuildSettings", helper.XcWorkspace.Path, "App", "Release", testOptions).Return(appSettings, nil).Once()
	workspaceSettings.On("SchemeBuildSettings", helper.XcWorkspace.Path, "Extension", "Release", testOptions).Return(extensionSettings, nil).Once()

	// Each target: one workspace read, by a scheme of the target's name, no project read.
	got, err := helper.cachedBuildSettings("App", "Release")
	require.NoError(t, err)
	require.Equal(t, []serialized.Object{appSettings}, got)
	got, err = helper.cachedBuildSettings("Extension", "Release")
	require.NoError(t, err)
	require.Equal(t, []serialized.Object{extensionSettings}, got)

	// A second read of a target is served from the cache.
	_, err = helper.cachedBuildSettings("App", "Release")
	require.NoError(t, err)
}

func TestFetchBuildSettings_workspaceFailureFallsBackToTheProject(t *testing.T) {
	projectOnly := serialized.Object{"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.app.extension"}
	workspaceSettings := xcworkspacemocks.NewBuildSettingsProvider(t)
	projectSettings := xcodeprojmocks.NewBuildSettingsProvider(t)
	projectSettings.On("TargetBuildSettings", testProjectPath, "Extension", "Release", testOptions).Return(projectOnly, nil).Once()
	helper := settingsHelper(t, workspaceSettings, false, projectSettings)
	workspaceSettings.On("SchemeBuildSettings", helper.XcWorkspace.Path, "Extension", "Release", testOptions).Return(nil, errors.New("xcodebuild: error: does not contain a scheme named Extension")).Once()

	got, err := helper.fetchBuildSettings("Extension", "Release")
	require.NoError(t, err)
	require.Equal(t, []serialized.Object{projectOnly}, got)
}

func TestFetchBuildSettings_compatModeReadsBoth(t *testing.T) {
	fromWorkspace := serialized.Object{"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.app"}
	projectOnly := serialized.Object{"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.app"}
	workspaceSettings := xcworkspacemocks.NewBuildSettingsProvider(t)
	projectSettings := xcodeprojmocks.NewBuildSettingsProvider(t)
	projectSettings.On("TargetBuildSettings", testProjectPath, "App", "Release", testOptions).Return(projectOnly, nil).Once()
	helper := settingsHelper(t, workspaceSettings, true, projectSettings)
	workspaceSettings.On("SchemeBuildSettings", helper.XcWorkspace.Path, "App", "Release", testOptions).Return(fromWorkspace, nil).Once()

	got, err := helper.fetchBuildSettings("App", "Release")
	require.NoError(t, err)
	require.Equal(t, []serialized.Object{fromWorkspace, projectOnly}, got, "workspace settings first, project settings second")
}

func TestReadSchemeBuildSettingString_missingKeyIsErrBuildSettingNotFound(t *testing.T) {
	projectSettings := xcodeprojmocks.NewBuildSettingsProvider(t)
	projectSettings.On("TargetBuildSettings", testProjectPath, "App", "Release", testOptions).Return(serialized.Object{"SDKROOT": "iphoneos"}, nil).Once()
	project := Project{logger: log.NewLogger(), projHelper: *settingsHelper(t, nil, false, projectSettings)} // no workspace

	_, err := project.ReadSchemeBuildSettingString("PRODUCT_NAME")
	require.ErrorIs(t, err, ErrBuildSettingNotFound, "a step tells a setting xcodebuild did not print from a failed read by this error")
}

func TestTargetBundleID_fallsBackToTheInfoPlist(t *testing.T) {
	// Without PRODUCT_BUNDLE_IDENTIFIER the bundle ID is CFBundleIdentifier of the Info.plist the
	// settings name, read through the injected plist handler; a variable in it expands from the
	// same settings, so xcodebuild runs once.
	settings := serialized.Object{"INFOPLIST_FILE": "App/Info.plist", "PRODUCT_NAME": "App"}
	for name, tt := range map[string]struct{ plist, want string }{
		"literal":  {plist: "io.bitrise.app", want: "io.bitrise.app"},
		"variable": {plist: "io.bitrise.$(PRODUCT_NAME)", want: "io.bitrise.App"},
	} {
		t.Run(name, func(t *testing.T) {
			projectSettings := xcodeprojmocks.NewBuildSettingsProvider(t)
			projectSettings.On("TargetBuildSettings", testProjectPath, "App", "Release", testOptions).Return(settings, nil).Once()
			plistHandler := plistmocks.NewFileHandler(t)
			plistHandler.On("Read", "/repo/App/Info.plist").Return(plistutil.PlistData{"CFBundleIdentifier": tt.plist}, 0, nil).Once()
			helper := settingsHelper(t, nil, false, projectSettings) // no workspace
			helper.plistFileHandler = plistHandler

			got, err := helper.TargetBundleID("App", "Release")
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFetchBuildSettings_project(t *testing.T) {
	projectOnly := serialized.Object{"PRODUCT_BUNDLE_IDENTIFIER": "io.bitrise.app"}
	projectSettings := xcodeprojmocks.NewBuildSettingsProvider(t)
	projectSettings.On("TargetBuildSettings", testProjectPath, "App", "Release", testOptions).Return(projectOnly, nil).Once()
	projectSettings.On("TargetBuildSettings", testProjectPath, "Extension", "Release", testOptions).Return(projectOnly, nil).Once()
	helper := settingsHelper(t, nil, false, projectSettings) // no workspace: nothing to resolve a scheme in

	for _, target := range []string{"App", "Extension"} {
		got, err := helper.fetchBuildSettings(target, "Release")
		require.NoError(t, err)
		require.Equal(t, []serialized.Object{projectOnly}, got)
	}
}
