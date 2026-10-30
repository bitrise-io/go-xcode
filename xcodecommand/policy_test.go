package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// One row per command: the flags its policy refuses and lets through, as verified against
// xcodebuild. A flag xcodebuild refuses by itself (-testPlan on archive, -xctestrun next to
// -scheme) passes: the build fails on it with xcodebuild's message. The merge rules are
// covered in merge_test.go.
func TestPolicies(t *testing.T) {
	tests := []struct {
		policy  actionPolicy
		rejects []string
		accepts []string
	}{
		{
			policy:  archivePolicy,
			rejects: []string{"-exportArchive", "-exportNotarizedApp", "-showBuildSettings", "-json", "-only-testing", "-test-iterations", "-enumerate-tests"},
			accepts: []string{"-allowProvisioningUpdates", "-skipPackagePluginValidation", "-someFutureFlag", "-testPlan", "-xctestrun", "-enableCodeCoverage"},
		},
		{
			policy:  buildPolicy,
			rejects: []string{"-exportArchive", "-json", "-only-testing"},
			accepts: []string{"-quiet", "-testPlan"},
		},
		{
			policy:  analyzePolicy,
			rejects: []string{"-exportArchive", "-json", "-skip-testing"},
			accepts: []string{"-quiet", "-testPlan"},
		},
		{
			// build-for-testing bakes the test selection into the xctestrun.
			policy:  buildForTestingPolicy,
			rejects: []string{"-exportArchive", "-showBuildSettings", "-json"},
			accepts: []string{"-testPlan", "-only-testing", "-skip-testing", "-only-test-configuration", "-enableCodeCoverage", "-testProductsPath", "-testLanguage", "-xctestrun"},
		},
		{
			policy:  testPolicy,
			rejects: []string{"-exportArchive", "-enumerate-tests", "-json"},
			accepts: []string{"-testPlan", "-only-testing", "-parallel-testing-enabled", "-enableCodeCoverage", "-xctestrun"},
		},
		{
			policy:  testWithoutBuildingPolicy,
			rejects: []string{"-exportArchive", "-showBuildSettings", "-json", "-enumerate-tests"},
			accepts: []string{"-xctestrun", "-testPlan", "-enableCodeCoverage", "-only-testing", "-skip-testing", "-test-iterations", "-scheme", "-testProductsPath"},
		},
		{
			policy:  exportArchivePolicy,
			rejects: []string{"-resolvePackageDependencies", "-showBuildSettings", "-only-testing"},
			accepts: []string{"-exportArchive", "-allowProvisioningUpdates", "-testPlan"},
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
