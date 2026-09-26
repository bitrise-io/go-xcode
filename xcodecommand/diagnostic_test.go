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
			wantErr:  `invalid additional option: "-destination generic/platform=iOS" is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it, so today's build runs without it. Remove it to keep that, or use -destination generic/platform=iOS to apply it.`,
		},
		{
			name:     "build setting written with a leading dash",
			params:   ArchiveParams{AdditionalOptions: []string{"-ENABLE_BITCODE=NO"}},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-ENABLE_BITCODE=NO"},
			wantKind: SuspiciousUserDefault,
			wantErr:  `invalid additional option: "-ENABLE_BITCODE=NO" looks like the build setting ENABLE_BITCODE=NO with a leading dash. xcodebuild reads it as a user default and ignores it, so today's build runs without it. Remove it to keep that, or use ENABLE_BITCODE=NO to apply it.`,
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

// Every message string is pinned here, one per finding; the rules that produce the
// findings are covered in options_test.go, policy_test.go and merge_test.go.

func TestOptions_Diagnostics_messages(t *testing.T) {
	opts := ParseAdditionalOptions([]string{"", "-destination 'platform=iOS Simulator,name=iPhone 15'", "-sdk macosx", "-=x", "-only-testing:", "Distribution", "-ENABLE_BITCODE=NO", "-destination"})
	require.Equal(t, []string{
		`"" is empty. xcodebuild treats it as an unknown build action. Remove it.`,
		`"-destination 'platform=iOS Simulator,name=iPhone 15'" is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it, so today's build runs without it. Remove it to keep that, or use -destination 'platform=iOS Simulator,name=iPhone 15' to apply it.`,
		`"-sdk macosx" is a flag quoted together with its value. xcodebuild refuses it. Use -sdk macosx instead.`,
		`"-=x" is not a valid flag. xcodebuild refuses it. Remove it.`,
		`"-only-testing:" has no value after the colon. Add the value or remove the flag.`,
		`"Distribution" is not a flag, a NAME=value build setting or a build action. xcodebuild treats it as an unknown build action. Quote a value with spaces, for example CODE_SIGN_IDENTITY="Apple Distribution", or remove it.`,
		`"-ENABLE_BITCODE=NO" looks like the build setting ENABLE_BITCODE=NO with a leading dash. xcodebuild reads it as a user default and ignores it, so today's build runs without it. Remove it to keep that, or use ENABLE_BITCODE=NO to apply it.`,
		`"-destination" has no value. Add one or remove the flag.`,
	}, messages(opts.Diagnostics()))
}

func TestLintPolicy_messages(t *testing.T) {
	opts := ParseAdditionalOptions([]string{"-exportArchive", "-test-iterations", "2", "clean"})
	require.Equal(t, []string{
		`"-exportArchive" switches xcodebuild into another mode and is not valid for archive. Remove it.`,
		`"-test-iterations 2" applies to test actions only and is not valid for archive. Remove it.`,
		`"clean" is a build action. The archive command sets its own actions. Remove it.`,
	}, messages(lintPolicy(opts, archivePolicy)))
}

func TestLint_mergeMessages(t *testing.T) {
	derived := Options{
		{Kind: ValueOption, Name: "-destination", Value: "generic/platform=iOS"},
		{Kind: Switch, Name: "-allowProvisioningUpdates"},
		{Kind: ValueOption, Name: "-xcconfig", Value: "/tmp/temp.xcconfig"},
		{Kind: ValueOption, Name: "-collect-test-diagnostics", Value: "never"},
	}
	user := Options{
		{Kind: ValueOption, Name: "-destination", Value: "generic/platform=tvOS"},
		{Kind: Switch, Name: "-allowProvisioningUpdates"},
		{Kind: ValueOption, Name: "-xcconfig", Value: "mine.xcconfig"},
		{Kind: UserDefault, Name: "-collect-test-diagnostics", Value: "on-failure"},
	}
	policy := actionPolicy{name: "archive", defaults: []string{"-destination"}}
	_, collisions := merge(derived, user, policy)

	require.Equal(t, []string{
		`"-collect-test-diagnostics=on-failure" is written with "=". xcodebuild reads it as a user default and ignores it, so today's build uses the Step's "-collect-test-diagnostics never". Remove it to keep that, or use -collect-test-diagnostics on-failure to apply it.`,
		`"-destination generic/platform=tvOS" replaces the Step's default "-destination generic/platform=iOS".`,
		`"-allowProvisioningUpdates" is already set by the Step. Remove it.`,
		`"-xcconfig mine.xcconfig" repeats "-xcconfig /tmp/temp.xcconfig", which the Step sets. xcodebuild refuses a repeated option. Remove it, or change the Step input instead.`,
	}, messages(lint(user, derived, collisions, policy)))
}

func TestLintStepInputs_message(t *testing.T) {
	user := Options{{Kind: Switch, Name: "-allowProvisioningUpdates"}}
	require.Equal(t, []string{
		`"-allowProvisioningUpdates" cannot update provisioning profiles on its own. xcodebuild needs the App Store Connect API key flags, which the Step adds when its automatic code signing is enabled. Remove it and enable automatic code signing instead.`,
	}, messages(lintStepInputs(nil, user)))
	require.Empty(t, lintStepInputs(user, user), "derived by the step: the merge reports it as redundant instead")
}

func messages(diagnostics []Diagnostic) []string {
	var out []string
	for _, d := range diagnostics {
		out = append(out, d.Message)
	}
	return out
}

// The corrected forms in the messages are pasted back into xcodebuild_options, which
// SplitAdditionalOptions reads with POSIX shell rules; each rendering must read back as
// the same value.
func TestShellQuoted(t *testing.T) {
	for value, want := range map[string]string{
		"":                                      `''`,
		"iphoneos":                              "iphoneos",
		"generic/platform=iOS":                  "generic/platform=iOS",
		"platform=iOS Simulator,name=iPhone 15": `'platform=iOS Simulator,name=iPhone 15'`,
		"Apple Development: Bot":                `'Apple Development: Bot'`,
		`say "hi"`:                              `'say "hi"'`,
		"it's here":                             `'it'\''s here'`,
		`a\b`:                                   `a\\b`,
		"$HOME/dd":                              `\$HOME/dd`,
		"don't $shout":                          `'don'\''t $shout'`,
	} {
		require.Equal(t, want, shellQuoted(value), value)
		args, err := SplitAdditionalOptions("-flag " + shellQuoted(value))
		require.NoError(t, err)
		require.Equal(t, []string{"-flag", value}, args, "must read back as the same value")
	}
	require.Equal(t, `-destination 'platform=iOS Simulator,name=iPhone 15'`, unquoteFlag("-destination 'platform=iOS Simulator,name=iPhone 15'"))
	require.Equal(t, "-sdk macosx", unquoteFlag("-sdk  macosx"))
}

func TestSplitAdditionalOptions(t *testing.T) {
	args, err := SplitAdditionalOptions(`-destination "platform=iOS Simulator,name=iPhone 15" -only-testing:'App Tests/Login' CODE_SIGN_IDENTITY=Apple\ Distribution`)
	require.NoError(t, err)
	require.Equal(t, []string{"-destination", "platform=iOS Simulator,name=iPhone 15", "-only-testing:App Tests/Login", "CODE_SIGN_IDENTITY=Apple Distribution"}, args)

	args, err = SplitAdditionalOptions("")
	require.NoError(t, err)
	require.Empty(t, args)

	_, err = SplitAdditionalOptions(`-destination "platform=iOS`)
	require.EqualError(t, err, `xcodebuild_options "-destination \"platform=iOS" cannot be split like a shell command line: Unterminated double-quoted string`)
}
