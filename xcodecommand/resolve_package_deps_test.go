package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePackages(t *testing.T) {
	tests := []struct {
		name     string
		params   ResolvePackagesParams
		want     []string
		wantNote string // the one informational line about what was left out
	}{
		{
			name:   "workspace",
			params: ResolvePackagesParams{ProjectPath: "test.xcworkspace", Scheme: "Test"},
			want:   []string{"-workspace", "test.xcworkspace", "-scheme", "Test", "-resolvePackageDependencies"},
		},
		{
			name:   "project",
			params: ResolvePackagesParams{ProjectPath: "test.xcodeproj", Scheme: "Test", Configuration: "Debug"},
			want:   []string{"-project", "test.xcodeproj", "-scheme", "Test", "-configuration", "Debug", "-resolvePackageDependencies"},
		},
		{
			name:   "user options pass through, including the SPM flags xcode-archive forwards",
			params: ResolvePackagesParams{ProjectPath: "test.xcodeproj", AdditionalOptions: additional("-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm")},
			want:   []string{"-project", "test.xcodeproj", "-resolvePackageDependencies", "-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm"},
		},
		{
			name:   "options that do not apply to resolution are left out, and one note says so",
			params: ResolvePackagesParams{ProjectPath: "test.xcodeproj", AdditionalOptions: additional("clean", "archive", "-test-iterations", "2", "-exportArchive", "Distribution", "-skipMacroValidation", "FOO=1", "-allowProvisioningUpdates")},
			want:   []string{"-project", "test.xcodeproj", "-resolvePackageDependencies", "-skipMacroValidation", "FOO=1", "-allowProvisioningUpdates"},
			// "-exportArchive Distribution": an unknown flag takes the bare word after it as its value
			wantNote: `"clean", "archive", "-test-iterations 2", "-exportArchive Distribution" left out of package resolution; the Step's main command gets the full xcodebuild_options.`,
		},
		{
			name:     "one option left out",
			params:   ResolvePackagesParams{ProjectPath: "test.xcodeproj", AdditionalOptions: additional("-quiet", "clean")},
			want:     []string{"-project", "test.xcodeproj", "-resolvePackageDependencies", "-quiet"},
			wantNote: `"clean" left out of package resolution; the Step's main command gets the full xcodebuild_options.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ResolvePackages(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
			if tt.wantNote == "" {
				require.Empty(t, cmd.Diagnostics())
				return
			}
			require.Equal(t, []string{tt.wantNote}, messages(cmd.Diagnostics()))
			require.True(t, cmd.Diagnostics()[0].Informational, "the main command reports what is wrong; this is a note")
			require.Equal(t, LeftOut, cmd.Diagnostics()[0].Kind)
		})
	}
}

func TestResolvePackages_WorkspaceNeedsScheme(t *testing.T) {
	_, err := ResolvePackages(ResolvePackagesParams{ProjectPath: "test.xcworkspace"})
	require.EqualError(t, err, "resolving the packages of workspace test.xcworkspace needs a scheme")
}
