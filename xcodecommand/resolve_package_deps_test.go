package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePackages(t *testing.T) {
	tests := []struct {
		name   string
		params ResolvePackagesParams
		want   []string
	}{
		{
			name:   "workspace",
			params: ResolvePackagesParams{ProjectPath: "test.xcworkspace"},
			want:   []string{"-workspace", "test.xcworkspace", "-resolvePackageDependencies"},
		},
		{
			name:   "project",
			params: ResolvePackagesParams{ProjectPath: "test.xcodeproj", Scheme: "Test", Configuration: "Debug"},
			want:   []string{"-project", "test.xcodeproj", "-scheme", "Test", "-configuration", "Debug", "-resolvePackageDependencies"},
		},
		{
			name:   "user options pass through, including the SPM flags xcode-archive forwards",
			params: ResolvePackagesParams{ProjectPath: "test.xcodeproj", AdditionalOptions: []string{"-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm"}},
			want:   []string{"-project", "test.xcodeproj", "-resolvePackageDependencies", "-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm"},
		},
		{
			name:   "options that do not apply to resolution are left out, for the main command to report",
			params: ResolvePackagesParams{ProjectPath: "test.xcodeproj", AdditionalOptions: []string{"clean", "archive", "-test-iterations", "2", "-exportArchive", "Distribution", "-skipMacroValidation", "FOO=1", "-allowProvisioningUpdates"}},
			want:   []string{"-project", "test.xcodeproj", "-resolvePackageDependencies", "-skipMacroValidation", "FOO=1", "-allowProvisioningUpdates"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ResolvePackages(tt.params)
			require.NoError(t, err)
			require.Equal(t, tt.want, cmd.Args())
			require.Empty(t, cmd.Diagnostics(), "the main command reports the options, once")
		})
	}
}
