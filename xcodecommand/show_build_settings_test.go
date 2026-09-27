package xcodecommand

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShowBuildSettings(t *testing.T) {
	tests := []struct {
		name   string
		params ShowBuildSettingsParams
		want   []string
		kinds  []DiagnosticKind
	}{
		{
			name:   "target in a project",
			params: ShowBuildSettingsParams{ProjectPath: "App.xcodeproj", Target: "App", Configuration: "Release"},
			want:   []string{"-project", "App.xcodeproj", "-target", "App", "-configuration", "Release", "-showBuildSettings", "-json"},
		},
		{
			name:   "scheme in a workspace",
			params: ShowBuildSettingsParams{ProjectPath: "App.xcworkspace", Scheme: "App"},
			want:   []string{"-workspace", "App.xcworkspace", "-scheme", "App", "-showBuildSettings", "-json"},
		},
		{
			// -json is derived; a user's own copy is a redundant switch, reported and sent once.
			name:   "a user -json is reported, not repeated",
			params: ShowBuildSettingsParams{ProjectPath: "App.xcodeproj", Target: "App", AdditionalOptions: []string{"-json"}},
			want:   []string{"-project", "App.xcodeproj", "-target", "App", "-showBuildSettings", "-json"},
			kinds:  []DiagnosticKind{RejectedOption, RedundantOption},
		},
		{
			name:   "SPM flags and build settings pass through after the flag",
			params: ShowBuildSettingsParams{ProjectPath: "App.xcodeproj", Scheme: "App", AdditionalOptions: []string{"-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm", "BUNDLE_IDENTIFIER=io.bitrise.sample"}},
			want:   []string{"-project", "App.xcodeproj", "-scheme", "App", "-showBuildSettings", "-json", "-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm", "BUNDLE_IDENTIFIER=io.bitrise.sample"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ShowBuildSettings(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
			require.Equal(t, tt.kinds, kinds(cmd.Diagnostics()))
		})
	}
}

// The xcode-archive flow: narrow the user's archive options to what a settings query
// accepts, then assemble the query from them.
func TestShowBuildSettings_narrowedArchiveOptions(t *testing.T) {
	spmFlags := []string{"-skipPackagePluginValidation", "-skipMacroValidation", "-skipPackageUpdates",
		"-disableAutomaticPackageResolution", "-onlyUsePackageVersionsFromResolvedFile", "-clonedSourcePackagesDirPath"}
	user := ParseAdditionalOptions([]string{"-destination", "generic/platform=iOS", "-skipMacroValidation", "COMPILER_INDEX_STORE_ENABLE=NO", "-quiet"})
	require.Empty(t, user.Diagnostics())
	narrowed := user.Filter(func(o Option) bool { return o.Kind == BuildSetting || slices.Contains(spmFlags, o.Name) })

	cmd, err := ShowBuildSettings(ShowBuildSettingsParams{ProjectPath: "App.xcodeproj", Scheme: "App", AdditionalOptions: narrowed.Args()})
	require.NoError(t, err)
	require.Equal(t, []string{"-project", "App.xcodeproj", "-scheme", "App", "-showBuildSettings", "-json", "-skipMacroValidation", "COMPILER_INDEX_STORE_ENABLE=NO"}, cmd.Args())
}
