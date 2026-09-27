package xcodecommand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Warn passes every option through and reports; Fail refuses the first non-informational
// finding before xcodebuild runs. The rows are production xcodebuild_options seen on
// xcode-archive; the message texts are pinned in TestDiagnosticMessages.
func TestValidation(t *testing.T) {
	tests := []struct {
		name      string
		params    ArchiveParams
		wantArgs  []string
		wantKinds []DiagnosticKind
		failPass  bool // only informational findings: Fail still builds the command
	}{
		{
			name:      "actions in the options",
			params:    ArchiveParams{AdditionalOptions: []string{"clean", "archive"}},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "clean", "archive"},
			wantKinds: []DiagnosticKind{ActionInOptions, ActionInOptions},
		},
		{
			name:      "test-only flags",
			params:    ArchiveParams{AdditionalOptions: []string{"-test-iterations", "2", "-retry-tests-on-failure"}},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "-test-iterations", "2", "-retry-tests-on-failure"},
			wantKinds: []DiagnosticKind{RejectedOption, RejectedOption},
		},
		{
			name:      "free-form value flag left without its value",
			params:    ArchiveParams{AdditionalOptions: []string{"-skipMacroValidation", "-destination"}},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "-skipMacroValidation", "-destination"},
			wantKinds: []DiagnosticKind{MalformedOption},
		},
		{
			name:      "flag quoted together with its value, build setting with a dash",
			params:    ArchiveParams{AdditionalOptions: []string{"-destination generic/platform=iOS", "-ENABLE_BITCODE=NO"}},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "-destination generic/platform=iOS", "-ENABLE_BITCODE=NO"},
			wantKinds: []DiagnosticKind{SuspiciousUserDefault, SuspiciousUserDefault},
		},
		{
			name:      "repeated value option",
			params:    ArchiveParams{XCConfigPath: "/tmp/temp.xcconfig", AdditionalOptions: []string{"-xcconfig", "mine.xcconfig"}},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "-xcconfig", "/tmp/temp.xcconfig", "-xcconfig", "mine.xcconfig"},
			wantKinds: []DiagnosticKind{RepeatedOption},
		},
		{
			// -allowProvisioningUpdates only works with the API key flags the step adds.
			name:      "lone -allowProvisioningUpdates points at the step input",
			params:    ArchiveParams{AdditionalOptions: []string{"-allowProvisioningUpdates"}},
			wantArgs:  []string{"archive", "-project", "App.xcodeproj", "-allowProvisioningUpdates"},
			wantKinds: []DiagnosticKind{PreferStepInput},
		},
		{
			name:     "a redundant switch and a replaced default are informational",
			params:   ArchiveParams{Destination: "generic/platform=iOS", Authentication: &testAuth, AdditionalOptions: []string{"-allowProvisioningUpdates", "-destination", "id=SIM"}},
			wantArgs: []string{"archive", "-project", "App.xcodeproj", "-authenticationKeyPath", "/key/path", "-authenticationKeyID", "keyID", "-authenticationKeyIssuerID", "issuerID", "-allowProvisioningUpdates", "-destination", "id=SIM"},
			// reported in derived-flag order
			wantKinds: []DiagnosticKind{Override, RedundantOption},
			failPass:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.params
			params.ProjectPath = "App.xcodeproj"

			cmd, err := Archive(params)
			require.NoError(t, err)
			require.Equal(t, tt.wantArgs, cmd.Args(), "Warn passes everything through")
			require.Equal(t, tt.wantKinds, kinds(cmd.Diagnostics()))

			params.Validation = Fail
			failed, err := Archive(params)
			if tt.failPass {
				require.NoError(t, err)
				require.Equal(t, cmd, failed)
				return
			}
			require.EqualError(t, err, "invalid additional option: "+cmd.Diagnostics()[0].Message)
		})
	}
}

// Every message is pinned here, one per finding. Each row lints user options against
// derived ones the way assemble does.
func TestDiagnosticMessages(t *testing.T) {
	tests := []struct {
		name    string
		user    []string
		derived Options
		policy  actionPolicy
		want    []string
	}{
		{
			name:   "parser",
			user:   []string{"", "-destination 'platform=iOS Simulator,name=iPhone 15'", "-sdk macosx", "-=x", "-only-testing:", "Distribution", "-ENABLE_BITCODE=NO", "-destination"},
			policy: archivePolicy,
			want: []string{
				`"" is empty. xcodebuild treats it as an unknown build action. Remove it.`,
				`"-destination 'platform=iOS Simulator,name=iPhone 15'" is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it, so today's build runs without it. Remove it to keep that, or use -destination 'platform=iOS Simulator,name=iPhone 15' to apply it.`,
				`"-sdk macosx" is a flag quoted together with its value. xcodebuild refuses it. Use -sdk macosx instead.`,
				`"-=x" is not a valid flag. xcodebuild refuses it. Remove it.`,
				`"-only-testing:" has no value after the colon. Add the value or remove the flag.`,
				`"Distribution" is not a flag, a NAME=value build setting or a build action. xcodebuild treats it as an unknown build action. Quote a value with spaces, for example CODE_SIGN_IDENTITY="Apple Distribution", or remove it.`,
				`"-ENABLE_BITCODE=NO" looks like the build setting ENABLE_BITCODE=NO with a leading dash. xcodebuild reads it as a user default and ignores it, so today's build runs without it. Remove it to keep that, or use ENABLE_BITCODE=NO to apply it.`,
				`"-destination" has no value. Add one or remove the flag.`,
			},
		},
		{
			name:   "policy and step input",
			user:   []string{"-exportArchive", "-test-iterations", "2", "clean", "-allowProvisioningUpdates"},
			policy: archivePolicy,
			want: []string{
				`"-exportArchive" switches xcodebuild into another mode and is not valid for archive. Remove it.`,
				`"-test-iterations 2" applies to test actions only and is not valid for archive. Remove it.`,
				`"clean" is a build action. The archive command sets its own actions. Remove it.`,
				`"-allowProvisioningUpdates" cannot update provisioning profiles on its own. xcodebuild needs the App Store Connect API key flags, which the Step adds when its automatic code signing is enabled. Remove it and enable automatic code signing instead.`,
			},
		},
		{
			name: "merge",
			user: []string{"-destination", "generic/platform=tvOS", "-allowProvisioningUpdates", "-xcconfig", "mine.xcconfig", "-collect-test-diagnostics=on-failure"},
			derived: Options{
				{Kind: ValueOption, Name: "-destination", Value: "generic/platform=iOS"},
				{Kind: Switch, Name: "-allowProvisioningUpdates"},
				{Kind: ValueOption, Name: "-xcconfig", Value: "/tmp/temp.xcconfig"},
				{Kind: ValueOption, Name: "-collect-test-diagnostics", Value: "never"},
			},
			policy: actionPolicy{name: "archive", defaults: []string{"-destination"}},
			want: []string{
				`"-collect-test-diagnostics=on-failure" is written with "=". xcodebuild reads it as a user default and ignores it, so today's build uses the Step's "-collect-test-diagnostics never". Remove it to keep that, or use -collect-test-diagnostics on-failure to apply it.`,
				`"-destination generic/platform=tvOS" replaces the Step's default "-destination generic/platform=iOS".`,
				`"-allowProvisioningUpdates" is already set by the Step. Remove it.`,
				`"-xcconfig mine.xcconfig" repeats "-xcconfig /tmp/temp.xcconfig", which the Step sets. xcodebuild refuses a repeated option. Remove it, or change the Step input instead.`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := ParseAdditionalOptions(tt.user)
			_, collisions := merge(tt.derived, user, tt.policy)
			require.Equal(t, tt.want, messages(lint(user, tt.derived, collisions, tt.policy)))
		})
	}
}

func kinds(diagnostics []Diagnostic) []DiagnosticKind {
	var out []DiagnosticKind
	for _, d := range diagnostics {
		out = append(out, d.Kind)
	}
	return out
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
