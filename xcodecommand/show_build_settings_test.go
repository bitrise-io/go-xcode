package xcodecommand

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShowBuildSettings(t *testing.T) {
	tests := []struct {
		name   string
		params showBuildSettingsParams
		want   []string
	}{
		{
			name:   "target in a project",
			params: showBuildSettingsParams{projectPath: "App.xcodeproj", target: "App", configuration: "Release"},
			want:   []string{"-project", "App.xcodeproj", "-target", "App", "-configuration", "Release", "-showBuildSettings"},
		},
		{
			name:   "scheme in a workspace",
			params: showBuildSettingsParams{projectPath: "App.xcworkspace", scheme: "App"},
			want:   []string{"-workspace", "App.xcworkspace", "-scheme", "App", "-showBuildSettings"},
		},
		{
			name:   "SPM flags and build settings pass through after the flag",
			params: showBuildSettingsParams{projectPath: "App.xcodeproj", scheme: "App", additionalOptions: []string{"-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm", "BUNDLE_IDENTIFIER=io.bitrise.sample"}},
			want:   []string{"-project", "App.xcodeproj", "-scheme", "App", "-showBuildSettings", "-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm", "BUNDLE_IDENTIFIER=io.bitrise.sample"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := showBuildSettings(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
			require.Empty(t, cmd.Diagnostics())
		})
	}
}

// The xcode-archive flow: narrow the user's archive options to what a settings query
// accepts, then assemble the query from them.
func TestShowBuildSettings_narrowedArchiveOptions(t *testing.T) {
	spmFlags := []string{"-skipPackagePluginValidation", "-skipMacroValidation", "-skipPackageUpdates",
		"-disableAutomaticPackageResolution", "-onlyUsePackageVersionsFromResolvedFile", "-clonedSourcePackagesDirPath"}
	user, diags := ParseAdditionalOptions([]string{"-destination", "generic/platform=iOS", "-skipMacroValidation", "COMPILER_INDEX_STORE_ENABLE=NO", "-quiet"})
	require.Empty(t, diags)
	narrowed := user.Filter(func(o Option) bool { return o.Kind == BuildSetting || slices.Contains(spmFlags, o.Name) })

	cmd, err := showBuildSettings(showBuildSettingsParams{projectPath: "App.xcodeproj", scheme: "App", additionalOptions: narrowed.Args()})
	require.NoError(t, err)
	require.Equal(t, []string{"-project", "App.xcodeproj", "-scheme", "App", "-showBuildSettings", "-skipMacroValidation", "COMPILER_INDEX_STORE_ENABLE=NO"}, cmd.Args())
}
