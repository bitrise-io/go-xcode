package xcodeproj

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-xcode/xcodeproject/testhelper"
	"github.com/stretchr/testify/require"
)

func Test_GivenNewlyGeneratedXcodeProject_WhenListingSchemes_ThenReturnsTheDefaultScheme(t *testing.T) {
	xcodeProjectPath := testhelper.NewlyGeneratedXcodeProjectPath(t)
	proj, err := Open(xcodeProjectPath)
	require.NoError(t, err)

	schemes, err := proj.Schemes()
	require.NoError(t, err)

	expectedSchemeName := "ios-sample"
	require.Equal(t, 1, len(schemes))
	require.Equal(t, expectedSchemeName, schemes[0].Name)
}

func Test_GivenNewlyGeneratedXcodeProjectWithUserDataGitignored_WhenListingSchemes_ThenReturnsTheDefaultScheme(t *testing.T) {
	xcodeProjectPath := testhelper.NewlyGeneratedXcodeProjectPath(t)

	userDataDir := filepath.Join(xcodeProjectPath, "xcuserdata")
	require.NoError(t, os.RemoveAll(userDataDir))

	proj, err := Open(xcodeProjectPath)
	require.NoError(t, err)

	schemes, err := proj.Schemes()
	require.NoError(t, err)

	expectedSchemeName := "ios-sample"
	require.Equal(t, 1, len(schemes))
	require.Equal(t, expectedSchemeName, schemes[0].Name)
}

func Test_GivenNewlyGeneratedXcodeProjectWithAutocreateSchemesDisabled_WhenListingSchemes_ThenReturnsError(t *testing.T) {
	xcodeProjectPath := testhelper.NewlyGeneratedXcodeProjectPath(t)

	worksaceSettingsPth := filepath.Join(xcodeProjectPath, "project.xcworkspace/xcshareddata/WorkspaceSettings.xcsettings")
	require.NoError(t, fileutil.WriteStringToFile(worksaceSettingsPth, workspaceSettingsWithAutocreateSchemesDisabledContent))

	proj, err := Open(xcodeProjectPath)
	require.NoError(t, err)

	schemes, err := proj.Schemes()
	require.EqualError(t, err, `no schemes found and the Xcode project's 'Autocreate schemes' option is disabled`)
	require.Equal(t, 0, len(schemes))
}

func Test_GivenNewlyGeneratedXcodeProjectWithASharedAndAUserScheme_WhenListingSchemes_ThenReturnsTheSharedScheme(t *testing.T) {
	xcodeProjectPath := testhelper.NewlyGeneratedXcodeProjectPath(t)

	userSchemePth := filepath.Join(xcodeProjectPath, "xcuserdata/bitrise-test-user.xcuserdatad/xcschemes/ios-sample.xcscheme")
	require.NoError(t, os.MkdirAll(filepath.Dir(userSchemePth), os.ModePerm))
	require.NoError(t, fileutil.WriteStringToFile(userSchemePth, defaultSchemeContent))

	sharedSchemeName := "custom-scheme"
	sharedSchemePth := filepath.Join(xcodeProjectPath, "xcshareddata/xcschemes", sharedSchemeName+".xcscheme")
	require.NoError(t, os.MkdirAll(filepath.Dir(sharedSchemePth), os.ModePerm))
	require.NoError(t, fileutil.WriteStringToFile(sharedSchemePth, defaultSchemeContent))

	proj, err := Open(xcodeProjectPath)
	require.NoError(t, err)

	schemes, err := proj.Schemes()
	require.NoError(t, err)

	expectedSchemeName := sharedSchemeName
	require.Equal(t, 1, len(schemes))
	require.Equal(t, expectedSchemeName, schemes[0].Name)
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
