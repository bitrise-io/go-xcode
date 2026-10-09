package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// One row per command: the flags its policy refuses and lets through, as verified against
// xcodebuild. The merge rules are covered in merge_test.go.
func TestPolicies(t *testing.T) {
	tests := []struct {
		policy  actionPolicy
		rejects []string
		accepts []string
	}{
		{
			policy:  archivePolicy,
			rejects: []string{"-exportArchive", "-exportNotarizedApp", "-showBuildSettings", "-json", "-testPlan", "-only-testing", "-test-iterations", "-xctestrun", "-enableCodeCoverage", "-enumerate-tests"},
			accepts: []string{"-allowProvisioningUpdates", "-skipPackagePluginValidation", "-someFutureFlag"},
		},
		{
			policy:  buildPolicy,
			rejects: []string{"-exportArchive", "-testPlan", "-only-testing"},
			accepts: []string{"-quiet"},
		},
		{
			policy:  analyzePolicy,
			rejects: []string{"-exportArchive", "-testPlan", "-skip-testing"},
			accepts: []string{"-quiet"},
		},
		{
			// build-for-testing bakes the test selection into the xctestrun.
			policy:  buildForTestingPolicy,
			rejects: []string{"-exportArchive", "-showBuildSettings", "-json", "-xctestrun"},
			accepts: []string{"-testPlan", "-only-testing", "-skip-testing", "-only-test-configuration", "-enableCodeCoverage", "-testProductsPath", "-testLanguage"},
		},
		{
			policy:  testPolicy,
			rejects: []string{"-exportArchive", "-xctestrun", "-enumerate-tests", "-json"},
			accepts: []string{"-testPlan", "-only-testing", "-parallel-testing-enabled", "-enableCodeCoverage"},
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
