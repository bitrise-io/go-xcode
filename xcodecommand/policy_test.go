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
			rejects:    []string{"-exportArchive", "-showBuildSettings", "-testPlan", "-only-testing", "-test-iterations", "-xctestrun"},
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
			rejects:    []string{"-exportArchive", "-showBuildSettings"},
			accepts:    []string{"-testPlan", "-only-testing", "-skip-testing", "-only-test-configuration"},
		},
		{
			policy:     testPolicy,
			defaults:   []string{"-collect-test-diagnostics"},
			appendable: []string{"-arch", "-destination", "-only-test-configuration", "-only-testing", "-skip-test-configuration", "-skip-testing"},
			rejects:    []string{"-exportArchive", "-xctestrun"},
			accepts:    []string{"-testPlan", "-only-testing", "-parallel-testing-enabled", "-enableCodeCoverage"},
		},
		{
			policy:     testWithoutBuildingPolicy,
			defaults:   []string{"-collect-test-diagnostics"},
			appendable: []string{"-arch", "-destination", "-only-test-configuration", "-only-testing", "-skip-test-configuration", "-skip-testing"},
			rejects:    []string{"-exportArchive", "-showBuildSettings"},
			accepts:    []string{"-xctestrun", "-only-testing", "-skip-testing", "-test-iterations"},
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
		{
			policy:  showBuildSettingsPolicy,
			rejects: []string{"-exportArchive", "-resolvePackageDependencies", "-testPlan"},
			accepts: []string{"-showBuildSettings", "-skipMacroValidation"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.policy.name, func(t *testing.T) {
			require.Equal(t, tt.defaults, slices.Sorted(slices.Values(tt.policy.defaults)))
			require.Equal(t, tt.appendable, slices.Sorted(slices.Values(tt.policy.appendable)))
			for _, flag := range tt.rejects {
				require.Equal(t, []DiagnosticKind{RejectedOption}, kinds(tt.policy.check(Options{{Kind: Switch, Name: flag}})), flag)
			}
			for _, flag := range tt.accepts {
				require.Empty(t, tt.policy.check(Options{{Kind: Switch, Name: flag}}), flag)
			}
			require.Equal(t, []DiagnosticKind{ActionInOptions}, kinds(tt.policy.check(Options{{Kind: Action, Name: "clean"}})), "every command owns its action list")
		})
	}
}

func TestActionPolicy_checkMessages(t *testing.T) {
	opts, _ := ParseAdditionalOptions([]string{"-exportArchive", "-test-iterations", "2", "clean"})
	diags := archivePolicy.check(opts)
	require.Equal(t, []string{
		`"-exportArchive" is not valid for archive: switches xcodebuild into another mode`,
		`"-test-iterations 2" is not valid for archive: applies to test actions only`,
		`"clean" is a build action, and archive sets its own actions`,
	}, messages(diags))
}

func messages(diagnostics []Diagnostic) []string {
	var out []string
	for _, d := range diagnostics {
		out = append(out, d.Message)
	}
	return out
}
