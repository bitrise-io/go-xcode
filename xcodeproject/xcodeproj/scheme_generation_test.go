package xcodeproj

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/xcodeproject/xcscheme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Expected value ported from v1 TestXcodeProj_ReCreateSchemes.
func TestXcodeProj_RecreateSchemes(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "ios-simple-objc.pbxproj"))
	require.NoError(t, err)

	// No file manager or user provider: generating schemes must not touch the filesystem.
	project, err := NewFactory(log.NewLogger(), nil, nil, nil, nil, nil).Parse(content, "test_path/test.xcodeproj")
	require.NoError(t, err)

	want := []xcscheme.Scheme{
		{
			LastUpgradeVersion: "1240",
			Version:            "1.3",
			BuildAction: xcscheme.BuildAction{
				ParallelizeBuildables:     "YES",
				BuildImplicitDependencies: "YES",
				BuildActionEntries: []xcscheme.BuildActionEntry{
					{
						BuildForTesting:   "YES",
						BuildForRunning:   "YES",
						BuildForProfiling: "YES",
						BuildForArchiving: "YES",
						BuildForAnalyzing: "YES",
						BuildableReference: xcscheme.BuildableReference{
							BuildableIdentifier: "primary",
							BlueprintIdentifier: "BA3CBE7419F7A93800CED4D5",
							BuildableName:       "ios-simple-objc.app",
							BlueprintName:       "ios-simple-objc",
							ReferencedContainer: "container:test.xcodeproj",
						},
					},
				},
			},
			TestAction: xcscheme.TestAction{
				BuildConfiguration:           "Debug",
				SelectedDebuggerIdentifier:   "Xcode.DebuggerFoundation.Debugger.LLDB",
				SelectedLauncherIdentifier:   "Xcode.DebuggerFoundation.Launcher.LLDB",
				ShouldUseLaunchSchemeArgsEnv: "YES",
				Testables: []xcscheme.TestableReference{
					{
						Skipped: "NO",
						BuildableReference: xcscheme.BuildableReference{
							BuildableIdentifier: "primary",
							BlueprintIdentifier: "BA3CBE9019F7A93900CED4D5",
							BuildableName:       "ios-simple-objcTests.xctest",
							BlueprintName:       "ios-simple-objcTests",
							ReferencedContainer: "container:test.xcodeproj",
						},
					},
				},
				MacroExpansion: xcscheme.MacroExpansion{
					BuildableReference: xcscheme.BuildableReference{
						BuildableIdentifier: "primary",
						BlueprintIdentifier: "BA3CBE7419F7A93800CED4D5",
						BuildableName:       "ios-simple-objc.app",
						BlueprintName:       "ios-simple-objc",
						ReferencedContainer: "container:test.xcodeproj",
					},
				},
				AdditionalOptions: xcscheme.AdditionalOptions{},
			},
			LaunchAction: xcscheme.LaunchAction{
				BuildConfiguration:             "Debug",
				SelectedDebuggerIdentifier:     "Xcode.DebuggerFoundation.Debugger.LLDB",
				SelectedLauncherIdentifier:     "Xcode.DebuggerFoundation.Launcher.LLDB",
				LaunchStyle:                    "0",
				UseCustomWorkingDirectory:      "NO",
				IgnoresPersistentStateOnLaunch: "NO",
				DebugDocumentVersioning:        "YES",
				DebugServiceExtension:          "internal",
				AllowLocationSimulation:        "YES",
				BuildableProductRunnable: xcscheme.BuildableProductRunnable{
					RunnableDebuggingMode: "0",
					BuildableReference: xcscheme.BuildableReference{
						BuildableIdentifier: "primary",
						BlueprintIdentifier: "BA3CBE7419F7A93800CED4D5",
						BuildableName:       "ios-simple-objc.app",
						BlueprintName:       "ios-simple-objc",
						ReferencedContainer: "container:test.xcodeproj",
					},
				},
				AdditionalOptions: xcscheme.AdditionalOptions{},
			},
			ProfileAction: xcscheme.ProfileAction{
				BuildConfiguration:           "Release",
				ShouldUseLaunchSchemeArgsEnv: "YES",
				SavedToolIdentifier:          "",
				UseCustomWorkingDirectory:    "NO",
				DebugDocumentVersioning:      "YES",
				BuildableProductRunnable: xcscheme.BuildableProductRunnable{
					RunnableDebuggingMode: "0",
					BuildableReference: xcscheme.BuildableReference{
						BuildableIdentifier: "primary",
						BlueprintIdentifier: "BA3CBE7419F7A93800CED4D5",
						BuildableName:       "ios-simple-objc.app",
						BlueprintName:       "ios-simple-objc",
						ReferencedContainer: "container:test.xcodeproj",
					},
				},
			},
			AnalyzeAction: xcscheme.AnalyzeAction{
				BuildConfiguration: "Debug",
			},
			ArchiveAction: xcscheme.ArchiveAction{
				BuildConfiguration:       "Release",
				RevealArchiveInOrganizer: "YES",
			},
			Name:     "ios-simple-objc",
			Path:     "",
			IsShared: false,
		},
	}

	require.Equal(t, want, project.RecreateSchemes())
}

func TestXcodeProj_SaveSharedScheme(t *testing.T) {
	project := schemesProject(t, "ios-sample.pbxproj", fakeUserProvider{name: testUserName})

	schemes := project.RecreateSchemes()
	require.NotEmpty(t, schemes)
	scheme := schemes[0]

	require.NoError(t, project.SaveSharedScheme(scheme))
	require.NoError(t, project.SaveSharedScheme(scheme), "saving again overwrites the file")

	pth := sharedSchemePath(project, scheme.Name)

	fileInfo, err := os.Stat(pth)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm(), "scheme file mode, as in v1")

	dirInfo, err := os.Stat(filepath.Dir(pth))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755)&^currentUmask(), dirInfo.Mode().Perm(), "xcschemes directory is not owner-only")

	saved, err := project.Schemes()
	require.NoError(t, err)
	require.Len(t, saved, 1)
	assert.Equal(t, scheme.Name, saved[0].Name)
	assert.True(t, saved[0].IsShared)
	assert.Equal(t, scheme.BuildAction, saved[0].BuildAction)
}

func currentUmask() os.FileMode {
	mask := syscall.Umask(0)
	syscall.Umask(mask)
	return os.FileMode(mask)
}
