package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShowBuildSettings(t *testing.T) {
	tests := []struct {
		name     string
		params   ShowBuildSettingsParams
		want     []string
		wantNote string // the one informational line about what was left out
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
			name:   "SPM flags and build settings pass through after the flag",
			params: ShowBuildSettingsParams{ProjectPath: "App.xcodeproj", Scheme: "App", AdditionalOptions: additional("-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm", "BUNDLE_IDENTIFIER=io.bitrise.sample")},
			want:   []string{"-project", "App.xcodeproj", "-scheme", "App", "-showBuildSettings", "-json", "-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm", "BUNDLE_IDENTIFIER=io.bitrise.sample"},
		},
		{
			name:     "options that do not apply to the query are left out, and one note says so",
			params:   ShowBuildSettingsParams{ProjectPath: "App.xcodeproj", Scheme: "App", AdditionalOptions: additional("clean", "archive", "-test-iterations", "2", "-resolvePackageDependencies", "-json", "FOO=1")},
			want:     []string{"-project", "App.xcodeproj", "-scheme", "App", "-showBuildSettings", "-json", "FOO=1"},
			wantNote: `"clean", "archive", "-test-iterations 2", "-resolvePackageDependencies" left out of the build settings query; the Step's main command gets the full xcodebuild_options.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ShowBuildSettings(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
			if tt.wantNote == "" {
				require.Empty(t, cmd.Diagnostics())
				return
			}
			require.Equal(t, []string{tt.wantNote}, messages(cmd.Diagnostics()))
			require.True(t, cmd.Diagnostics()[0].Informational)
			require.Equal(t, LeftOut, cmd.Diagnostics()[0].Kind)
		})
	}
}
