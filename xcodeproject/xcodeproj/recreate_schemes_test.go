package xcodeproj

import (
	"testing"

	plist "github.com/bitrise-io/go-plist"
	"github.com/bitrise-io/go-xcode/xcodeproject/serialized"
	"github.com/bitrise-io/go-xcode/xcodeproject/xcscheme"
	"github.com/stretchr/testify/require"
)

func TestXcodeProj_ReCreateSchemes(t *testing.T) {
	var raw serialized.Object
	_, err := plist.Unmarshal([]byte(rawProj), &raw)
	require.NoError(t, err)

	proj, err := parseProj("BA3CBE6D19F7A93800CED4D5", raw, nil)
	require.NoError(t, err)

	tests := []struct {
		name     string
		proj     Proj
		projPath string
		want     []xcscheme.Scheme
	}{
		{
			name:     "simple",
			proj:     proj,
			projPath: "test_path/test.xcodeproj",
			want: []xcscheme.Scheme{
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := XcodeProj{
				Proj: tt.proj,
				Path: tt.projPath,
			}
			got := p.ReCreateSchemes()

			require.Equal(t, tt.want, got)
		})
	}
}
