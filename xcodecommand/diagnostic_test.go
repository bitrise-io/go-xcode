package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Warn passes every option through and reports; Fail refuses the same input before
// xcodebuild runs. The rows are production xcodebuild_options seen on xcode-archive.
func TestValidation(t *testing.T) {
	tests := []struct {
		name     string
		params   ArchiveParams
		wantArgs []string
		wantKind DiagnosticKind
		wantErr  string
	}{
		{
			name:     "actions in the options",
			params:   ArchiveParams{AdditionalOptions: []string{"clean", "archive"}},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "clean", "archive"},
			wantKind: ActionInOptions,
			wantErr:  `invalid additional option: "clean" is a build action. The archive command sets its own actions. Remove it.`,
		},
		{
			name:     "test-only flags",
			params:   ArchiveParams{AdditionalOptions: []string{"-test-iterations", "2", "-retry-tests-on-failure"}},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-test-iterations", "2", "-retry-tests-on-failure"},
			wantKind: RejectedOption,
			wantErr:  `invalid additional option: "-test-iterations 2" applies to test actions only and is not valid for archive. Remove it.`,
		},
		{
			name:     "free-form value flag left without its value",
			params:   ArchiveParams{AdditionalOptions: []string{"-skipMacroValidation", "-destination"}},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-skipMacroValidation", "-destination"},
			wantKind: MalformedOption,
			wantErr:  `invalid additional option: "-destination" has no value. Add one or remove the flag.`,
		},
		{
			name:     "flag quoted together with its value",
			params:   ArchiveParams{AdditionalOptions: []string{"-destination generic/platform=iOS"}},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-destination generic/platform=iOS"},
			wantKind: SuspiciousUserDefault,
			wantErr:  `invalid additional option: "-destination generic/platform=iOS" is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it. Use -destination generic/platform=iOS instead.`,
		},
		{
			name:     "build setting written with a leading dash",
			params:   ArchiveParams{AdditionalOptions: []string{"-ENABLE_BITCODE=NO"}},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-ENABLE_BITCODE=NO"},
			wantKind: SuspiciousUserDefault,
			wantErr:  `invalid additional option: "-ENABLE_BITCODE=NO" looks like the build setting ENABLE_BITCODE=NO with a leading dash. xcodebuild reads it as a user default and the setting never applies. Use ENABLE_BITCODE=NO instead.`,
		},
		{
			name:     "repeated value option",
			params:   ArchiveParams{XCConfigPath: "/tmp/temp.xcconfig", AdditionalOptions: []string{"-xcconfig", "mine.xcconfig"}},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-xcconfig", "/tmp/temp.xcconfig", "-xcconfig", "mine.xcconfig"},
			wantKind: RepeatedOption,
			wantErr:  `invalid additional option: "-xcconfig mine.xcconfig" repeats "-xcconfig /tmp/temp.xcconfig", which the Step sets. xcodebuild refuses a repeated option. Remove it, or change the Step input instead.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.params
			params.ProjectPath = "App.xcodeproj"

			cmd, err := Archive(params)
			require.NoError(t, err)
			require.Equal(t, tt.wantArgs, cmd.Args())
			require.NotEmpty(t, cmd.Diagnostics())
			require.Equal(t, tt.wantKind, cmd.Diagnostics()[0].Kind)

			params.Validation = Fail
			_, err = Archive(params)
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestFail_ignoresInformationalDiagnostics(t *testing.T) {
	// A redundant switch and a replaced default are informational: Fail still succeeds.
	cmd, err := Archive(ArchiveParams{
		ProjectPath:       "App.xcodeproj",
		Destination:       "generic/platform=iOS",
		Authentication:    &testAuth,
		AdditionalOptions: []string{"-allowProvisioningUpdates", "-destination", "id=SIM"},
		Validation:        Fail,
	})
	require.NoError(t, err)
	require.Equal(t, []DiagnosticKind{Override, RedundantOption}, kinds(cmd.Diagnostics()), "reported in derived-flag order")
	require.Equal(t, []string{
		"archive", "-project", "App.xcodeproj",
		"-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID",
		"-allowProvisioningUpdates", "-destination", "id=SIM",
	}, cmd.Args())
}

func kinds(diagnostics []Diagnostic) []DiagnosticKind {
	var out []DiagnosticKind
	for _, d := range diagnostics {
		out = append(out, d.Kind)
	}
	return out
}

// -allowProvisioningUpdates only works with the API key flags the step adds when its
// automatic code signing is on; passed alone it points the user at the step input.
func TestPreferStepInput_allowProvisioningUpdates(t *testing.T) {
	cmd, err := Archive(ArchiveParams{ProjectPath: "App.xcodeproj", AdditionalOptions: []string{"-allowProvisioningUpdates"}})
	require.NoError(t, err)
	require.Equal(t, []string{"archive", "-project", "App.xcodeproj", "-allowProvisioningUpdates"}, cmd.Args(), "passed through under Warn")
	require.Equal(t, []Diagnostic{{Kind: PreferStepInput, Message: `"-allowProvisioningUpdates" cannot update provisioning profiles on its own. xcodebuild needs the App Store Connect API key flags, which the Step adds when its automatic code signing is enabled. Remove it and enable automatic code signing instead.`}}, cmd.Diagnostics())

	_, err = Archive(ArchiveParams{ProjectPath: "App.xcodeproj", AdditionalOptions: []string{"-allowProvisioningUpdates"}, Validation: Fail})
	require.Error(t, err)

	// With the step's API key on, the same option is merely redundant.
	cmd, err = Archive(ArchiveParams{ProjectPath: "App.xcodeproj", Authentication: &testAuth, AdditionalOptions: []string{"-allowProvisioningUpdates"}})
	require.NoError(t, err)
	require.Equal(t, []DiagnosticKind{RedundantOption}, kinds(cmd.Diagnostics()))
}
