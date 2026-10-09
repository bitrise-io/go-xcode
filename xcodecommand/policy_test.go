package xcodecommand

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// One row per command: what its policy replaces, keeps both of, refuses and lets through.
// The constructors are wired to these policies; the merge rules themselves are covered in
// merge_test.go.
func TestPolicies(t *testing.T) {
	tests := []struct {
		policy     actionPolicy
		defaults   []string
		appendable []string
		rejects    []string
		accepts    []string
	}{
		{
			policy:     archivePolicy,
			defaults:   []string{"-destination"},
			appendable: []string{"-arch"},
			rejects:    []string{"-exportArchive", "-exportNotarizedApp", "-showBuildSettings", "-json", "-testPlan", "-only-testing", "-test-iterations", "-xctestrun", "-enableCodeCoverage", "-enumerate-tests"},
			accepts:    []string{"-allowProvisioningUpdates", "-skipPackagePluginValidation", "-someFutureFlag"},
		},
		{
			policy:     buildPolicy,
			defaults:   []string{"-destination"},
			appendable: []string{"-arch"},
			rejects:    []string{"-exportArchive", "-testPlan", "-only-testing"},
			accepts:    []string{"-quiet"},
		},
		{
			policy:     analyzePolicy,
			defaults:   []string{"-destination", "-resultBundlePath"},
			appendable: []string{"-arch"},
			rejects:    []string{"-exportArchive", "-testPlan", "-skip-testing"},
			accepts:    []string{"-quiet"},
		},
		{
			// build-for-testing bakes the test selection into the xctestrun.
			policy:     buildForTestingPolicy,
			defaults:   []string{"-destination"},
			appendable: []string{"-arch"},
			rejects:    []string{"-exportArchive", "-showBuildSettings", "-json"},
			accepts:    []string{"-testPlan", "-only-testing", "-skip-testing", "-only-test-configuration", "-enableCodeCoverage", "-testProductsPath", "-testLanguage"},
		},
		{
			policy:  exportArchivePolicy,
			rejects: []string{"-resolvePackageDependencies", "-showBuildSettings", "-testPlan"},
			accepts: []string{"-exportArchive", "-allowProvisioningUpdates"},
		},
		{
			policy:  resolvePackagesPolicy,
			rejects: []string{"-exportArchive", "-showBuildSettings", "-only-testing"},
			accepts: []string{"-resolvePackageDependencies", "-skipPackagePluginValidation"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.policy.name, func(t *testing.T) {
			require.Equal(t, tt.defaults, slices.Sorted(slices.Values(tt.policy.defaults)))
			require.Equal(t, tt.appendable, slices.Sorted(slices.Values(tt.policy.appendable)))
			for _, flag := range tt.rejects {
				_, rejected := tt.policy.rejects(Option{Kind: Switch, Name: flag})
				require.True(t, rejected, flag)
			}
			for _, flag := range tt.accepts {
				_, rejected := tt.policy.rejects(Option{Kind: Switch, Name: flag})
				require.False(t, rejected, flag)
			}
		})
	}
}
