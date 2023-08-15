package xcodeproj

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-xcode/xcodeproject/testhelper"
	"github.com/stretchr/testify/require"
)

func Test_GivenNewlyGeneratedXcodeProject_WhenListingSchemes_ThenReturnsTheDefaultScheme(t *testing.T) {
	xcodeProjectPath := newlyGeneratedXcodeProjectPath(t)
	proj, err := Open(xcodeProjectPath)
	require.NoError(t, err)

	schemes, err := proj.Schemes()
	require.NoError(t, err)

	expectedSchemeNames := []string{"ios-sample"}
	require.Equal(t, len(expectedSchemeNames), len(schemes))
	for _, expectedSchemeName := range expectedSchemeNames {
		schemeFound := false
		for _, scheme := range schemes {
			if scheme.Name == expectedSchemeName {
				schemeFound = true
				break
			}
		}
		require.True(t, schemeFound)
	}
}

func Test_GivenNewlyGeneratedXcodeProjectWithUserDataGitignored_WhenListingSchemes_ThenReturnsTheDefaultScheme(t *testing.T) {
	xcodeProjectPath := newlyGeneratedXcodeProjectPath(t)

	userDataDir := filepath.Join(xcodeProjectPath, "xcuserdata")
	require.NoError(t, os.RemoveAll(userDataDir))

	proj, err := Open(xcodeProjectPath)
	require.NoError(t, err)

	schemes, err := proj.Schemes()
	require.NoError(t, err)

	expectedSchemeNames := []string{"ios-sample"}
	require.Equal(t, len(expectedSchemeNames), len(schemes))
	for _, expectedSchemeName := range expectedSchemeNames {
		schemeFound := false
		for _, scheme := range schemes {
			if scheme.Name == expectedSchemeName {
				schemeFound = true
				break
			}
		}
		require.True(t, schemeFound)
	}
}

func Test_GivenNewlyGeneratedXcodeProjectWithAutocreateSchemesDisabled_WhenListingSchemes_ThenReturnsError(t *testing.T) {
	xcodeProjectPath := newlyGeneratedXcodeProjectPath(t)

	worksaceSettingsPth := filepath.Join(xcodeProjectPath, "project.xcworkspace/xcshareddata/WorkspaceSettings.xcsettings")
	require.NoError(t, fileutil.WriteStringToFile(worksaceSettingsPth, workspaceSettingsWithAutocreateSchemesDisabledContent))

	proj, err := Open(xcodeProjectPath)
	require.NoError(t, err)

	schemes, err := proj.Schemes()
	require.EqualError(t, err, `no schemes found and the Xcode project's 'Autocreate schemes' option is disabled`)
	require.Equal(t, 0, len(schemes))
}

func Test_GivenNewlyGeneratedXcodeProjectWithASharedAndAUserScheme_WhenListingSchemes_ThenReturnsTheSharedScheme(t *testing.T) {
	xcodeProjectPath := newlyGeneratedXcodeProjectPath(t)

	vagrantUserSchemePth := filepath.Join(xcodeProjectPath, "xcuserdata/vagrant.xcuserdatad/xcschemes/ios-sample.xcscheme")
	require.NoError(t, os.MkdirAll(filepath.Dir(vagrantUserSchemePth), os.ModePerm))
	require.NoError(t, fileutil.WriteStringToFile(vagrantUserSchemePth, defaultSchemeContent))

	sharedSchemeName := "custom-scheme"
	sharedSchemePth := filepath.Join(xcodeProjectPath, "xcshareddata/xcschemes", sharedSchemeName+".xcscheme")
	require.NoError(t, os.MkdirAll(filepath.Dir(sharedSchemePth), os.ModePerm))
	require.NoError(t, fileutil.WriteStringToFile(sharedSchemePth, defaultSchemeContent))

	proj, err := Open(xcodeProjectPath)
	require.NoError(t, err)

	schemes, err := proj.Schemes()
	require.NoError(t, err)

	expectedSchemeNames := []string{sharedSchemeName}
	require.Equal(t, len(expectedSchemeNames), len(schemes))
	for _, expectedSchemeName := range expectedSchemeNames {
		schemeFound := false
		for _, scheme := range schemes {
			if scheme.Name == expectedSchemeName {
				schemeFound = true
				break
			}
		}
		require.True(t, schemeFound)
	}
}

func ensureTmpTestdataDir(t *testing.T) string {
	_, callerFilename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	callerDir := filepath.Dir(callerFilename)
	callerPackageDir := filepath.Dir(callerDir)
	packageTmpTestdataDir := filepath.Join(callerPackageDir, "_testdata")
	if _, err := os.Stat(packageTmpTestdataDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(packageTmpTestdataDir, os.ModePerm)
		require.NoError(t, err)
	}
	return packageTmpTestdataDir
}

func newlyGeneratedXcodeProjectPath(t *testing.T) string {
	testdataDir := ensureTmpTestdataDir(t)
	newlyGeneratedXcodeProjectDir := filepath.Join(testdataDir, "newly_generated_xcode_project")
	_, err := os.Stat(newlyGeneratedXcodeProjectDir)
	newlyGeneratedXcodeProjectDirExist := !errors.Is(err, os.ErrNotExist)
	if newlyGeneratedXcodeProjectDirExist {
		cmd := command.New("git", "clean", "-f", "-x", "-d")
		cmd.SetDir(newlyGeneratedXcodeProjectDir)
		require.NoError(t, cmd.Run())
	} else {
		repo := "https://github.com/godrei/ios-sample.git"
		branch := "main"
		testhelper.GitCloneBranch(t, repo, branch, newlyGeneratedXcodeProjectDir)
	}
	return filepath.Join(newlyGeneratedXcodeProjectDir, "ios-sample.xcodeproj")
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

const defaultSchemeContent = `<?xml version="1.0" encoding="UTF-8"?>
<Scheme
   LastUpgradeVersion = "1430"
   version = "1.7">
   <BuildAction
      parallelizeBuildables = "YES"
      buildImplicitDependencies = "YES">
      <BuildActionEntries>
         <BuildActionEntry
            buildForTesting = "YES"
            buildForRunning = "YES"
            buildForProfiling = "YES"
            buildForArchiving = "YES"
            buildForAnalyzing = "YES">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "133616C22A7945DD00B69017"
               BuildableName = "ios-sample.app"
               BlueprintName = "ios-sample"
               ReferencedContainer = "container:ios-sample.xcodeproj">
            </BuildableReference>
         </BuildActionEntry>
      </BuildActionEntries>
   </BuildAction>
   <TestAction
      buildConfiguration = "Debug"
      selectedDebuggerIdentifier = "Xcode.DebuggerFoundation.Debugger.LLDB"
      selectedLauncherIdentifier = "Xcode.DebuggerFoundation.Launcher.LLDB"
      shouldUseLaunchSchemeArgsEnv = "YES"
      shouldAutocreateTestPlan = "YES">
      <Testables>
         <TestableReference
            skipped = "NO"
            parallelizable = "YES">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "133616D32A7945DE00B69017"
               BuildableName = "ios-sampleTests.xctest"
               BlueprintName = "ios-sampleTests"
               ReferencedContainer = "container:ios-sample.xcodeproj">
            </BuildableReference>
         </TestableReference>
         <TestableReference
            skipped = "NO"
            parallelizable = "YES">
            <BuildableReference
               BuildableIdentifier = "primary"
               BlueprintIdentifier = "133616DD2A7945DE00B69017"
               BuildableName = "ios-sampleUITests.xctest"
               BlueprintName = "ios-sampleUITests"
               ReferencedContainer = "container:ios-sample.xcodeproj">
            </BuildableReference>
         </TestableReference>
      </Testables>
   </TestAction>
   <LaunchAction
      buildConfiguration = "Debug"
      selectedDebuggerIdentifier = "Xcode.DebuggerFoundation.Debugger.LLDB"
      selectedLauncherIdentifier = "Xcode.DebuggerFoundation.Launcher.LLDB"
      launchStyle = "0"
      useCustomWorkingDirectory = "NO"
      ignoresPersistentStateOnLaunch = "NO"
      debugDocumentVersioning = "YES"
      debugServiceExtension = "internal"
      allowLocationSimulation = "YES">
      <BuildableProductRunnable
         runnableDebuggingMode = "0">
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "133616C22A7945DD00B69017"
            BuildableName = "ios-sample.app"
            BlueprintName = "ios-sample"
            ReferencedContainer = "container:ios-sample.xcodeproj">
         </BuildableReference>
      </BuildableProductRunnable>
   </LaunchAction>
   <ProfileAction
      buildConfiguration = "Release"
      shouldUseLaunchSchemeArgsEnv = "YES"
      savedToolIdentifier = ""
      useCustomWorkingDirectory = "NO"
      debugDocumentVersioning = "YES">
      <BuildableProductRunnable
         runnableDebuggingMode = "0">
         <BuildableReference
            BuildableIdentifier = "primary"
            BlueprintIdentifier = "133616C22A7945DD00B69017"
            BuildableName = "ios-sample.app"
            BlueprintName = "ios-sample"
            ReferencedContainer = "container:ios-sample.xcodeproj">
         </BuildableReference>
      </BuildableProductRunnable>
   </ProfileAction>
   <AnalyzeAction
      buildConfiguration = "Debug">
   </AnalyzeAction>
   <ArchiveAction
      buildConfiguration = "Release"
      revealArchiveInOrganizer = "YES">
   </ArchiveAction>
</Scheme>
`
