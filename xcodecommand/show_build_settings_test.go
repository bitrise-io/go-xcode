package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShowBuildSettings(t *testing.T) {
	tests := []struct {
		name   string
		params ShowBuildSettingsParams
		want   []string
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
			params: ShowBuildSettingsParams{ProjectPath: "App.xcodeproj", Scheme: "App", AdditionalOptions: []string{"-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm", "BUNDLE_IDENTIFIER=io.bitrise.sample"}},
			want:   []string{"-project", "App.xcodeproj", "-scheme", "App", "-showBuildSettings", "-json", "-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm", "BUNDLE_IDENTIFIER=io.bitrise.sample"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ShowBuildSettings(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
		})
	}
}
