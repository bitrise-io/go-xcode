package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActionSpec_check(t *testing.T) {
	tests := []struct {
		name string
		spec actionSpec
		args []string
		want []DiagnosticKind
		msg  string
	}{
		{name: "archive accepts ordinary flags", spec: archiveSpec, args: []string{"-destination", "generic/platform=iOS", "-quiet", "CODE_SIGNING_ALLOWED=NO", "-UseModernBuildSystem=YES", "-allowProvisioningUpdates", "-someFutureFlag", "value", "-skipPackagePluginValidation", "-clonedSourcePackagesDirPath", "/tmp/spm"}},
		{name: "archive refuses a mode-switching flag", spec: archiveSpec, args: []string{"-exportArchive"}, want: []DiagnosticKind{RejectedOption}, msg: `"-exportArchive" is not valid for archive: switches xcodebuild into export mode`},
		{name: "archive refuses a test-only flag", spec: archiveSpec, args: []string{"-test-iterations", "2"}, want: []DiagnosticKind{RejectedOption}, msg: `"-test-iterations 2" is not valid for archive: applies to test actions only`},
		{name: "archive refuses a test plan", spec: archiveSpec, args: []string{"-testPlan", "Full"}, want: []DiagnosticKind{RejectedOption}},
		{name: "archive refuses a colon test selection", spec: archiveSpec, args: []string{"-only-testing:AppTests"}, want: []DiagnosticKind{RejectedOption}},
		{name: "archive refuses actions", spec: archiveSpec, args: []string{"clean", "archive"}, want: []DiagnosticKind{ActionInOptions, ActionInOptions}, msg: `"clean" is a build action, and archive sets its own actions`},
		{name: "build-for-testing accepts test selection and a test plan", spec: buildForTestingSpec, args: []string{"-only-testing:AppTests", "-skip-testing:AppTests/Slow", "-only-test-configuration", "Debug", "-testPlan", "Full"}},
		{name: "build-for-testing refuses a mode-switching flag", spec: buildForTestingSpec, args: []string{"-showBuildSettings"}, want: []DiagnosticKind{RejectedOption}},
		{name: "export accepts -exportArchive itself", spec: exportArchiveSpec, args: []string{"-exportArchive"}},
		{name: "export refuses another mode", spec: exportArchiveSpec, args: []string{"-resolvePackageDependencies"}, want: []DiagnosticKind{RejectedOption}},
		{name: "resolve accepts its own mode flag", spec: resolvePackagesSpec, args: []string{"-resolvePackageDependencies", "-skipPackagePluginValidation"}},
		{name: "resolve refuses export mode", spec: resolvePackagesSpec, args: []string{"-exportArchive"}, want: []DiagnosticKind{RejectedOption}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, parseDiags := ParseAdditionalOptions(tt.args)
			require.Empty(t, parseDiags)

			diags := tt.spec.check(opts)

			var kinds []DiagnosticKind
			for _, d := range diags {
				kinds = append(kinds, d.Kind)
			}
			require.Equal(t, tt.want, kinds)
			if tt.msg != "" {
				require.Equal(t, tt.msg, diags[0].Message)
			}
		})
	}
}
