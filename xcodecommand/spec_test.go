package xcodecommand

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// One row per command: what its spec replaces, keeps both of, refuses and lets through.
// The constructors are wired to these specs; the merge rules themselves are covered in
// merge_test.go.
func TestSpecs(t *testing.T) {
	tests := []struct {
		spec       actionSpec
		defaults   []string
		appendable []string
		rejects    []string
		accepts    []string
	}{
		{
			spec:       archiveSpec,
			defaults:   []string{"-destination"},
			appendable: []string{"-arch"},
			rejects:    []string{"-exportArchive", "-showBuildSettings", "-testPlan", "-only-testing", "-test-iterations", "-xctestrun"},
			accepts:    []string{"-allowProvisioningUpdates", "-skipPackagePluginValidation", "-someFutureFlag"},
		},
		{
			spec:       buildSpec,
			defaults:   []string{"-destination"},
			appendable: []string{"-arch"},
			rejects:    []string{"-exportArchive", "-testPlan", "-only-testing"},
			accepts:    []string{"-quiet"},
		},
		{
			spec:       analyzeSpec,
			defaults:   []string{"-destination", "-resultBundlePath"},
			appendable: []string{"-arch"},
			rejects:    []string{"-exportArchive", "-testPlan", "-skip-testing"},
			accepts:    []string{"-quiet"},
		},
		{
			// build-for-testing bakes the test selection into the xctestrun.
			spec:       buildForTestingSpec,
			defaults:   []string{"-destination"},
			appendable: []string{"-arch"},
			rejects:    []string{"-exportArchive", "-showBuildSettings"},
			accepts:    []string{"-testPlan", "-only-testing", "-skip-testing", "-only-test-configuration"},
		},
		{
			spec:       testSpec,
			defaults:   []string{"-collect-test-diagnostics"},
			appendable: []string{"-arch", "-destination", "-only-test-configuration", "-only-testing", "-skip-test-configuration", "-skip-testing"},
			rejects:    []string{"-exportArchive", "-xctestrun"},
			accepts:    []string{"-testPlan", "-only-testing", "-parallel-testing-enabled", "-enableCodeCoverage"},
		},
		{
			spec:       testWithoutBuildingSpec,
			defaults:   []string{"-collect-test-diagnostics"},
			appendable: []string{"-arch", "-destination", "-only-test-configuration", "-only-testing", "-skip-test-configuration", "-skip-testing"},
			rejects:    []string{"-exportArchive", "-showBuildSettings"},
			accepts:    []string{"-xctestrun", "-only-testing", "-skip-testing", "-test-iterations"},
		},
		{
			spec:    exportArchiveSpec,
			rejects: []string{"-resolvePackageDependencies", "-showBuildSettings", "-testPlan"},
			accepts: []string{"-exportArchive", "-allowProvisioningUpdates"},
		},
		{
			spec:    resolvePackagesSpec,
			rejects: []string{"-exportArchive", "-showBuildSettings", "-only-testing"},
			accepts: []string{"-resolvePackageDependencies", "-skipPackagePluginValidation"},
		},
		{
			spec:    showBuildSettingsSpec,
			rejects: []string{"-exportArchive", "-resolvePackageDependencies", "-testPlan"},
			accepts: []string{"-showBuildSettings", "-skipMacroValidation"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.spec.name, func(t *testing.T) {
			require.Equal(t, tt.defaults, slices.Sorted(slices.Values(tt.spec.defaults)))
			require.Equal(t, tt.appendable, slices.Sorted(slices.Values(tt.spec.appendable)))
			for _, flag := range tt.rejects {
				require.Equal(t, []DiagnosticKind{RejectedOption}, kinds(tt.spec.check(Options{{Kind: Switch, Name: flag}})), flag)
			}
			for _, flag := range tt.accepts {
				require.Empty(t, tt.spec.check(Options{{Kind: Switch, Name: flag}}), flag)
			}
			require.Equal(t, []DiagnosticKind{ActionInOptions}, kinds(tt.spec.check(Options{{Kind: Action, Name: "clean"}})), "every command owns its action list")
		})
	}
}

func TestActionSpec_checkMessages(t *testing.T) {
	opts, _ := ParseAdditionalOptions([]string{"-exportArchive", "-test-iterations", "2", "clean"})
	diags := archiveSpec.check(opts)
	require.Equal(t, []string{
		`"-exportArchive" is not valid for archive: switches xcodebuild into export mode`,
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
